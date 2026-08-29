// Package openapi is the first-party OpenAPI route plugin. Keeping it in a
// subpackage demonstrates the extension model: the plugin lives in the same
// project, but it contributes to the generic docs.Router like any external
// plugin would.
package openapi

import (
	"embed"
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"sort"
	"strings"

	docs "github.com/DonaldMurillo/fastr-docs"
)

//go:embed openapi.js
var runtimeFS embed.FS

// Assets exposes the plugin's CSP-safe browser runtime for mounting through
// gofastr/core/static at a project-owned URL.
func Assets() fs.FS { return runtimeFS }

// Plugin is a small forwarding adapter with a stable extension-package API.
type Plugin struct {
	SpecPath    string
	Spec        []byte
	Path        string
	Title       string
	Description string
	// ServerURL overrides the first OpenAPI servers entry. This is useful for
	// staging/production deployments where the same contract is rendered from
	// different docs hosts. Empty uses the contract's resolved default server.
	ServerURL string
	Order     int
}

func (Plugin) Name() string { return "openapi" }

func (p Plugin) Apply(r *docs.Router) error {
	if r == nil {
		return errors.New("router is nil")
	}
	data := p.Spec
	if len(data) == 0 {
		if strings.TrimSpace(p.SpecPath) == "" {
			return errors.New("SpecPath or Spec is required")
		}
		var err error
		data, err = os.ReadFile(p.SpecPath)
		if err != nil {
			return fmt.Errorf("read %q: %w", p.SpecPath, err)
		}
	}
	var spec document
	if err := json.Unmarshal(data, &spec); err != nil {
		return fmt.Errorf("parse spec: %w", err)
	}
	if spec.OpenAPI == "" && spec.Swagger == "" {
		return errors.New("document must declare openapi or swagger")
	}
	path := p.Path
	if path == "" {
		path = "/api-reference"
	}
	title := p.Title
	if title == "" {
		title = spec.Info.Title
	}
	if title == "" {
		title = "API reference"
	}
	description := p.Description
	if description == "" {
		description = spec.Info.Description
	}
	if description == "" {
		description = "Generated from the project OpenAPI contract."
	}
	operations, err := spec.operations()
	if err != nil {
		return fmt.Errorf("parse operations: %w", err)
	}
	order := p.Order
	if order < 1 {
		order = len(r.Routes()) + 1
	}
	serverURL := spec.serverURL(p.ServerURL)
	r.AllowConnectOrigin(serverURL)
	return r.Screen(path, docs.ScreenConfig{
		Title:       title,
		Description: description,
		Component:   &Reference{Title: title, Description: description, Version: spec.version(), ServerURL: serverURL, Operations: operations, Schemas: spec.Components.Schemas},
		SearchText:  spec.searchText(operations),
		Plugin:      "openapi",
		Order:       order,
		Offline:     true,
	})
}

type document struct {
	OpenAPI string `json:"openapi"`
	Swagger string `json:"swagger"`
	Info    struct {
		Title       string `json:"title"`
		Description string `json:"description"`
	} `json:"info"`
	Servers    []server                              `json:"servers"`
	Host       string                                `json:"host"`
	BasePath   string                                `json:"basePath"`
	Schemes    []string                              `json:"schemes"`
	Paths      map[string]map[string]json.RawMessage `json:"paths"`
	Components struct {
		Schemas map[string]Schema `json:"schemas"`
	} `json:"components"`
}

type server struct {
	URL       string                    `json:"url"`
	Variables map[string]serverVariable `json:"variables"`
}

type serverVariable struct {
	Default string `json:"default"`
}

type operation struct {
	Summary     string `json:"summary"`
	Description string `json:"description"`
	OperationID string `json:"operationId"`
	Parameters  []struct {
		Name     string `json:"name"`
		In       string `json:"in"`
		Required bool   `json:"required"`
		Schema   struct {
			Type string `json:"type"`
		} `json:"schema"`
	} `json:"parameters"`
	RequestBody *struct {
		Required bool `json:"required"`
	} `json:"requestBody"`
	Responses map[string]struct {
		Description string `json:"description"`
	} `json:"responses"`
}

// Schema is a normalized component schema used by Reference.
type Schema struct {
	Type        string `json:"type"`
	Description string `json:"description"`
	Properties  map[string]struct {
		Type        string `json:"type"`
		Description string `json:"description"`
	} `json:"properties"`
}

// Operation is the normalized operation model used by Reference.
type Operation struct {
	Method, Path, Summary, Description, OperationID string
	Parameters                                      []string
	RequestBody                                     bool
	Response                                        string
}

func (d document) version() string {
	if d.OpenAPI != "" {
		return d.OpenAPI
	}
	return d.Swagger
}

func (d document) serverURL(override string) string {
	if value := strings.TrimSpace(override); value != "" {
		return strings.TrimRight(value, "/")
	}
	if len(d.Servers) == 0 {
		if strings.TrimSpace(d.Host) == "" {
			return ""
		}
		scheme := "https"
		if len(d.Schemes) > 0 && strings.TrimSpace(d.Schemes[0]) != "" {
			scheme = strings.TrimSpace(d.Schemes[0])
		}
		basePath := strings.Trim(strings.TrimSpace(d.BasePath), "/")
		if basePath == "" {
			return strings.TrimRight(scheme+"://"+strings.TrimSpace(d.Host), "/")
		}
		return strings.TrimRight(scheme+"://"+strings.TrimSpace(d.Host)+"/"+basePath, "/")
	}
	value := strings.TrimSpace(d.Servers[0].URL)
	for name, variable := range d.Servers[0].Variables {
		value = strings.ReplaceAll(value, "{"+name+"}", variable.Default)
	}
	return strings.TrimRight(value, "/")
}

func (d document) operations() ([]Operation, error) {
	var out []Operation
	paths := make([]string, 0, len(d.Paths))
	for path := range d.Paths {
		paths = append(paths, path)
	}
	sort.Strings(paths)
	for _, path := range paths {
		methods := d.Paths[path]
		methodNames := make([]string, 0, len(methods))
		for method := range methods {
			if !isHTTPMethod(method) {
				continue
			}
			methodNames = append(methodNames, method)
		}
		sort.Strings(methodNames)
		for _, method := range methodNames {
			var op operation
			if err := json.Unmarshal(methods[method], &op); err != nil {
				return nil, fmt.Errorf("%s %s: %w", strings.ToUpper(method), path, err)
			}
			params := make([]string, 0, len(op.Parameters))
			for _, param := range op.Parameters {
				label := param.Name + " · " + param.In
				if param.Schema.Type != "" {
					label += " · " + param.Schema.Type
				}
				if param.Required {
					label += " · required"
				}
				params = append(params, label)
			}
			response := ""
			codes := make([]string, 0, len(op.Responses))
			for code := range op.Responses {
				codes = append(codes, code)
			}
			sort.Strings(codes)
			if len(codes) > 0 {
				response = codes[0]
			}
			out = append(out, Operation{Method: strings.ToUpper(method), Path: path, Summary: op.Summary, Description: op.Description, OperationID: op.OperationID, Parameters: params, RequestBody: op.RequestBody != nil, Response: response})
		}
	}
	return out, nil
}

func isHTTPMethod(method string) bool {
	switch strings.ToLower(method) {
	case "get", "put", "post", "delete", "options", "head", "patch", "trace":
		return true
	default:
		return false
	}
}

func (d document) searchText(operations []Operation) string {
	var b strings.Builder
	for _, op := range operations {
		fmt.Fprintf(&b, "%s %s %s %s %s\n", op.Method, op.Path, op.Summary, op.Description, op.OperationID)
	}
	for name, item := range d.Components.Schemas {
		fmt.Fprintf(&b, "%s %s %s\n", name, item.Type, item.Description)
	}
	return b.String()
}
