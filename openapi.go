package docs

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"sort"
	"strings"
)

// OpenAPIPlugin adds a readable API reference page to the same Router as the
// rest of the site. It intentionally starts from the portable OpenAPI JSON
// contract; a richer renderer can later replace the generated Markdown while
// keeping this registration surface stable.
type OpenAPIPlugin struct {
	SpecPath    string
	Spec        []byte
	Path        string
	Title       string
	Description string
	Order       int
}

func (p OpenAPIPlugin) Name() string { return "openapi" }

// Apply validates the OpenAPI document, creates a generated reference page,
// and registers it as a plugin route. Spec is preferred in tests or embedded
// builds; SpecPath is resolved from the process working directory.
func (p OpenAPIPlugin) Apply(r *Router) error {
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

	var spec openAPIDocument
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
	source := renderOpenAPIMarkdown(title, description, spec)
	order := p.Order
	if order < 1 {
		order = r.seq + 1
	}
	route, err := r.addRoute(path, Route{
		Title:       title,
		Description: description,
		Kind:        KindPlugin,
		Order:       order,
		Offline:     true,
		Plugin:      p.Name(),
		page:        &PageConfig{Title: title, Description: description, Source: source, Offline: true},
	})
	if err != nil {
		return err
	}
	// Keep the route pointer useful to callers inspecting the tree without
	// exposing the generated source as a second public model.
	if route != nil {
		route.Plugin = p.Name()
	}
	return nil
}

type openAPIDocument struct {
	OpenAPI string `json:"openapi"`
	Swagger string `json:"swagger"`
	Info    struct {
		Title       string `json:"title"`
		Description string `json:"description"`
	} `json:"info"`
	Paths map[string]map[string]json.RawMessage `json:"paths"`
}

func renderOpenAPIMarkdown(title, description string, spec openAPIDocument) string {
	var b strings.Builder
	fmt.Fprintf(&b, "# %s\n\n%s\n\n", title, description)
	b.WriteString("## Endpoints\n\n")
	paths := make([]string, 0, len(spec.Paths))
	for path := range spec.Paths {
		paths = append(paths, path)
	}
	sort.Strings(paths)
	for _, path := range paths {
		methods := spec.Paths[path]
		methodNames := make([]string, 0, len(methods))
		for method := range methods {
			if !isOpenAPIMethod(method) {
				continue
			}
			methodNames = append(methodNames, method)
		}
		sort.Strings(methodNames)
		for _, method := range methodNames {
			var op struct {
				Summary     string `json:"summary"`
				Description string `json:"description"`
				OperationID string `json:"operationId"`
			}
			if err := json.Unmarshal(methods[method], &op); err != nil {
				continue
			}
			label := strings.ToUpper(method) + " " + path
			b.WriteString("### `" + label + "`\n\n")
			if op.Summary != "" {
				b.WriteString(op.Summary + "\n\n")
			} else if op.Description != "" {
				b.WriteString(op.Description + "\n\n")
			}
			if op.OperationID != "" {
				b.WriteString("Operation ID: `" + op.OperationID + "`\n\n")
			}
		}
	}
	if len(spec.Paths) == 0 {
		b.WriteString("No paths were found in the contract.\n")
	}
	return b.String()
}

func isOpenAPIMethod(method string) bool {
	switch strings.ToLower(method) {
	case "get", "put", "post", "delete", "options", "head", "patch", "trace":
		return true
	default:
		return false
	}
}
