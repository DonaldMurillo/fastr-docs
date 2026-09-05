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
	"gopkg.in/yaml.v3"
)

//go:embed openapi.js
var runtimeFS embed.FS

// Assets exposes the plugin's CSP-safe browser runtime for mounting through
// gofastr/core/static at a project-owned URL.
func Assets() fs.FS { return runtimeFS }

// RuntimeAssets lets the Router collect this plugin's browser runtime, so a
// project does not have to wire openapi.js into its own main.go for serving
// and again for export. It satisfies docs.RuntimeAssetPlugin.
func (Plugin) RuntimeAssets() (map[string][]byte, error) {
	body, err := fs.ReadFile(runtimeFS, "openapi.js")
	if err != nil {
		return nil, err
	}
	return map[string][]byte{"openapi.js": body}, nil
}

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
	Badge     docs.NavBadge
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
	if err := decodeSpec(data, &spec); err != nil {
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
		Badge:       p.Badge,
	})
}

func decodeSpec(data []byte, target any) error {
	if err := json.Unmarshal(data, target); err == nil {
		return nil
	}
	var yamlValue any
	if err := yaml.Unmarshal(data, &yamlValue); err != nil {
		return err
	}
	normalized, err := normalizeYAMLValue(yamlValue)
	if err != nil {
		return err
	}
	encoded, err := json.Marshal(normalized)
	if err != nil {
		return err
	}
	return json.Unmarshal(encoded, target)
}

func normalizeYAMLValue(value any) (any, error) {
	switch value := value.(type) {
	case map[string]any:
		out := make(map[string]any, len(value))
		for key, item := range value {
			normalized, err := normalizeYAMLValue(item)
			if err != nil {
				return nil, err
			}
			out[key] = normalized
		}
		return out, nil
	case map[any]any:
		out := make(map[string]any, len(value))
		for key, item := range value {
			name, ok := key.(string)
			if !ok {
				return nil, fmt.Errorf("YAML object key %v is not a string", key)
			}
			normalized, err := normalizeYAMLValue(item)
			if err != nil {
				return nil, err
			}
			out[name] = normalized
		}
		return out, nil
	case []any:
		out := make([]any, len(value))
		for index, item := range value {
			normalized, err := normalizeYAMLValue(item)
			if err != nil {
				return nil, err
			}
			out[index] = normalized
		}
		return out, nil
	default:
		return value, nil
	}
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
	Consumes   []string                              `json:"consumes"`
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
	Summary     string          `json:"summary"`
	Description string          `json:"description"`
	OperationID string          `json:"operationId"`
	Parameters  []rawParameter  `json:"parameters"`
	RequestBody *rawRequestBody `json:"requestBody"`
	Responses   map[string]struct {
		Description string `json:"description"`
	} `json:"responses"`
}

type rawParameter struct {
	Name        string          `json:"name"`
	In          string          `json:"in"`
	Description string          `json:"description"`
	Required    bool            `json:"required"`
	Type        string          `json:"type"` // Swagger 2 parameter shape.
	Schema      json.RawMessage `json:"schema"`
	Default     json.RawMessage `json:"default"`
	Example     json.RawMessage `json:"example"`
}

type rawRequestBody struct {
	Description string                  `json:"description"`
	Required    bool                    `json:"required"`
	Content     map[string]rawMediaType `json:"content"`
}

type rawMediaType struct {
	Schema   json.RawMessage              `json:"schema"`
	Example  json.RawMessage              `json:"example"`
	Examples map[string]rawMediaTypeValue `json:"examples"`
}

type rawMediaTypeValue struct {
	Value json.RawMessage `json:"value"`
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
	ParameterSpecs                                  []Parameter
	RequestBody                                     bool
	RequestBodyRequired                             bool
	RequestBodyContentType                          string
	RequestBodyExample                              string
	Response                                        string
}

// Parameter is the normalized request input contract shown by Reference.
// In is one of path, query, header, or cookie; unsupported parameter kinds are
// retained in the reference but are not sent by the browser console.
type Parameter struct {
	Name        string
	In          string
	Description string
	Type        string
	Default     string
	Example     string
	Required    bool
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
		pathParameters := []rawParameter{}
		if raw := methods["parameters"]; len(raw) > 0 {
			if err := json.Unmarshal(raw, &pathParameters); err != nil {
				return nil, fmt.Errorf("%s parameters: %w", path, err)
			}
		}
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
			parameters := mergeParameters(pathParameters, op.Parameters)
			params := make([]string, 0, len(op.Parameters))
			parameterSpecs := make([]Parameter, 0, len(parameters))
			for _, rawParameter := range parameters {
				param := normalizeParameter(rawParameter)
				parameterSpecs = append(parameterSpecs, param)
				label := param.Name + " · " + param.In
				if param.Type != "" {
					label += " · " + param.Type
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
			hasBody, bodyRequired, contentType, bodyExample := normalizeRequestBody(op.RequestBody, parameterSpecs, d.Consumes)
			out = append(out, Operation{
				Method: strings.ToUpper(method), Path: path, Summary: op.Summary,
				Description: op.Description, OperationID: op.OperationID,
				Parameters: params, ParameterSpecs: parameterSpecs,
				RequestBody: hasBody, RequestBodyRequired: bodyRequired,
				RequestBodyContentType: contentType, RequestBodyExample: bodyExample,
				Response: response,
			})
		}
	}
	return out, nil
}

func mergeParameters(pathParameters, operationParameters []rawParameter) []rawParameter {
	merged := append([]rawParameter(nil), pathParameters...)
	for _, parameter := range operationParameters {
		replaced := false
		for index, existing := range merged {
			if existing.Name == parameter.Name && existing.In == parameter.In {
				merged[index] = parameter
				replaced = true
				break
			}
		}
		if !replaced {
			merged = append(merged, parameter)
		}
	}
	return merged
}

func normalizeParameter(raw rawParameter) Parameter {
	param := Parameter{Name: raw.Name, In: raw.In, Description: raw.Description, Type: raw.Type, Required: raw.Required}
	var schema struct {
		Type    string          `json:"type"`
		Default json.RawMessage `json:"default"`
		Example json.RawMessage `json:"example"`
	}
	if len(raw.Schema) > 0 {
		_ = json.Unmarshal(raw.Schema, &schema)
	}
	if param.Type == "" {
		param.Type = schema.Type
	}
	param.Default = rawJSONText(raw.Default)
	if param.Default == "" {
		param.Default = rawJSONText(schema.Default)
	}
	param.Example = rawJSONText(raw.Example)
	if param.Example == "" {
		param.Example = rawJSONText(schema.Example)
	}
	return param
}

func normalizeRequestBody(raw *rawRequestBody, parameters []Parameter, consumes []string) (bool, bool, string, string) {
	for _, parameter := range parameters {
		if strings.EqualFold(parameter.In, "body") {
			contentType := firstContentType(consumes)
			return true, parameter.Required, firstNonEmpty(contentType, "application/json"), firstNonEmpty(parameter.Example, parameter.Default)
		}
	}
	if raw == nil {
		return false, false, "", ""
	}
	contentType := firstContentTypeMap(raw.Content)
	media := raw.Content[contentType]
	example := rawJSONText(media.Example)
	if example == "" {
		exampleNames := make([]string, 0, len(media.Examples))
		for name := range media.Examples {
			exampleNames = append(exampleNames, name)
		}
		sort.Strings(exampleNames)
		if len(exampleNames) > 0 {
			example = rawJSONText(media.Examples[exampleNames[0]].Value)
		}
	}
	return true, raw.Required, firstNonEmpty(contentType, "application/json"), example
}

func firstContentType(values []string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return strings.TrimSpace(value)
		}
	}
	return ""
}

func firstContentTypeMap(values map[string]rawMediaType) string {
	if len(values) == 0 {
		return ""
	}
	names := make([]string, 0, len(values))
	for name := range values {
		names = append(names, name)
	}
	sort.Slice(names, func(i, j int) bool {
		iJSON := strings.Contains(strings.ToLower(names[i]), "json")
		jJSON := strings.Contains(strings.ToLower(names[j]), "json")
		if iJSON != jJSON {
			return iJSON
		}
		return names[i] < names[j]
	})
	return names[0]
}

func rawJSONText(raw json.RawMessage) string {
	if len(raw) == 0 || string(raw) == "null" {
		return ""
	}
	var value any
	if err := json.Unmarshal(raw, &value); err != nil {
		return ""
	}
	if stringValue, ok := value.(string); ok {
		return stringValue
	}
	body, err := json.MarshalIndent(value, "", "  ")
	if err != nil {
		return ""
	}
	return string(body)
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
