// Package openapi is the first-party OpenAPI route plugin. Keeping it in a
// subpackage demonstrates the extension model: the plugin lives in the same
// project, but it contributes to the generic docs.Router like any external
// plugin would.
package openapi

import (
	"bytes"
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
	// SpecPath reads the OpenAPI document from a file; Spec wins when
	// both are set.
	SpecPath string
	// Spec is the OpenAPI document as bytes, for embedded specs.
	Spec []byte
	// Path mounts the reference; defaults to /api-reference.
	Path string
	// Title overrides the document's info.title in navigation.
	Title string
	// Description overrides the document's info.description.
	Description string
	// ServerURL overrides the first OpenAPI servers entry. This is useful for
	// staging/production deployments where the same spec is rendered from
	// different docs hosts. Empty uses the spec's resolved default server.
	ServerURL string
	// Locale declares the language of this mount's route, so a translated
	// reference at /es/api-reference pairs with the original section and the
	// header tab follows the reader's language. Empty leaves the route
	// unmarked, which the default locale build serves everywhere.
	Locale string
	// Strings carries the surface labels for this mount. Every empty field
	// falls back to DefaultStrings, so a translation can land a label at a
	// time.
	Strings Strings
	// Order positions the reference among the site sections.
	Order int
	// Badge pins a small badge beside the reference in sidebars and
	// drawers.
	Badge docs.NavBadge
}

func (Plugin) Name() string { return "openapi" }

// Apply validates the OpenAPI document, registers the reference as a screen
// with this mount's locale and surface labels, and allows the spec's
// server origin for the browser request console. Spec wins over SpecPath,
// and every unset field falls back to the document's own info, then to
// framework defaults.
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
		description = "Generated from the project OpenAPI spec."
	}
	operations, err := spec.operations()
	if err != nil {
		return fmt.Errorf("parse operations: %w", err)
	}
	if err := checkDuplicateOperationIDs(operations); err != nil {
		return fmt.Errorf("invalid spec %s: %w", path, err)
	}
	order := p.Order
	if order < 1 {
		order = len(r.Routes()) + 1
	}
	locale := strings.ToLower(strings.TrimSpace(p.Locale))
	if locale != "" && !docs.IsValidLocale(locale) {
		return fmt.Errorf("locale %q is not a BCP 47 shaped tag", locale)
	}
	apiKeyHeader, apiKeyName := spec.apiKeyHeader()
	serverURL := spec.serverURL(p.ServerURL)
	if serverURL == "" {
		// No declared server and no override means the API lives where
		// the docs live; a dead console is the wrong reading of silence.
		serverURL = "/"
	}
	if strings.Contains(serverURL, "://") {
		r.AllowConnectOrigin(serverURL)
	}
	idPrefix := mountIDPrefix(path)
	return r.Screen(path, docs.ScreenConfig{
		Title:       title,
		Description: description,
		Component: &Reference{Title: title, Description: description, Version: spec.version(), ServerURL: serverURL,
			Servers: spec.serverURLs(), BearerScheme: spec.bearerScheme(),
			APIKeyHeader: apiKeyHeader, APIKeyName: apiKeyName,
			Operations: operations, Schemas: spec.Components.Schemas, Strings: p.Strings.withDefaults(), IDPrefix: idPrefix},
		SearchText: spec.searchText(operations),
		Plugin:     "openapi",
		Order:      order,
		Offline:    true,
		Preload:    "hover",
		Badge:      p.Badge,
		Metadata:   docs.ContentMetadata{Locale: locale},
	})
}

// mountIDPrefix derives a per-mount discriminator from the mount path, so
// two mounts of the plugin (a reference and its translation) never produce
// colliding element ids on one host.
func mountIDPrefix(path string) string {
	value := strings.Trim(strings.ReplaceAll(strings.ReplaceAll(path, "/", "-"), "_", "-"), "-")
	if value == "" {
		value = "api"
	}
	return value
}

// checkDuplicateOperationIDs fails loudly on a spec that reuses an
// operation id: the console and anchors key off ids, and a silent collision
// makes one operation unreachable.
func checkDuplicateOperationIDs(operations []Operation) error {
	seen := map[string]string{}
	for _, op := range operations {
		if op.OperationID == "" {
			continue
		}
		if previous, ok := seen[op.OperationID]; ok {
			return fmt.Errorf("operations %s and %s share operationId %q", previous, op.Path, op.OperationID)
		}
		seen[op.OperationID] = op.Path
	}
	return nil
}

func decodeSpec(data []byte, target any) error {
	// Editors on Windows save a BOM; JSON does not allow one.
	data = bytes.TrimPrefix(data, []byte{0xEF, 0xBB, 0xBF})
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
	Security   []map[string][]string                 `json:"security"`
	Paths      map[string]map[string]json.RawMessage `json:"paths"`
	Components struct {
		Schemas         map[string]Schema         `json:"schemas"`
		SecuritySchemes map[string]securityScheme `json:"securitySchemes"`
	} `json:"components"`
}

// securityScheme is one declared way to authenticate. The console supports
// the HTTP bearer shape; other kinds still render their name so the reader
// knows credentials belong in the request.
type securityScheme struct {
	Type        string `json:"type"`
	Scheme      string `json:"scheme"`
	Name        string `json:"name"`
	In          string `json:"in"`
	Description string `json:"description"`
}

type server struct {
	URL       string                    `json:"url"`
	Variables map[string]serverVariable `json:"variables"`
}

type serverVariable struct {
	Default string `json:"default"`
}

// resolvedURL returns the server's URL with each declared variable
// substituted by its default, trailing slash trimmed, so a templated
// server becomes a URL the console can call.
func (s server) resolvedURL() string {
	value := strings.TrimSpace(s.URL)
	for name, variable := range s.Variables {
		value = strings.ReplaceAll(value, "{"+name+"}", variable.Default)
	}
	return strings.TrimRight(value, "/")
}

type operation struct {
	Summary     string          `json:"summary"`
	Description string          `json:"description"`
	OperationID string          `json:"operationId"`
	Deprecated  bool            `json:"deprecated"`
	Tags        []string        `json:"tags"`
	Parameters  []rawParameter  `json:"parameters"`
	RequestBody *rawRequestBody `json:"requestBody"`
	Responses   map[string]struct {
		Description string `json:"description"`
		Headers     map[string]struct {
			Description string `json:"description"`
		} `json:"headers"`
	} `json:"responses"`
}

type rawParameter struct {
	Name        string          `json:"name"`
	In          string          `json:"in"`
	Description string          `json:"description"`
	Required    bool            `json:"required"`
	Type        string          `json:"type"` // Swagger 2 parameter shape.
	Enum        []string        `json:"enum"`
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
	// Type is the schema's JSON type, such as "object".
	Type string `json:"type"`
	// Description is the schema's documentation string.
	Description string `json:"description"`
	// Properties maps each property name to its schema.
	Properties map[string]struct {
		Type        string `json:"type"`
		Description string `json:"description"`
	} `json:"properties"`
}

// Operation is the normalized operation model used by Reference.
type Operation struct {
	Method, Path, Summary, Description, OperationID string
	// Parameters are the input names shown in the operation card.
	Parameters []string
	// ParameterSpecs drive the request console's input fields.
	ParameterSpecs []Parameter
	// RequestBody says the operation expects a body.
	RequestBody bool
	// RequestBodyRequired says the spec demands the body.
	RequestBodyRequired bool
	// RequestBodyContentType is the body's media type.
	RequestBodyContentType string
	// RequestBodyExample seeds the console's body editor.
	RequestBodyExample string
	// Response is the declared success response description.
	Response string
	// ResponseHeaders names the documented response headers, so a reader
	// can find rate limits and pagination before the first call.
	ResponseHeaders []string
	// Deprecated marks an operation the spec retired.
	Deprecated bool
	// Tags carry the document's own grouping.
	Tags []string
}

// Parameter is the normalized request input model shown by Reference.
// In is one of path, query, header, or cookie; unsupported parameter kinds are
// retained in the reference but are not sent by the browser console.
type Parameter struct {
	// Name is the parameter as the spec spells it.
	Name string
	// In is where the parameter travels: path, query, header, or
	// cookie.
	In string
	// Description documents the parameter.
	Description string
	// Type is the parameter's JSON type.
	Type string
	// Default is the spec's default value.
	Default string
	// Example seeds the console's input.
	Example string
	// Required says the request fails without it.
	Required bool
	// Format refines Type, and picks the input: a date format renders a
	// date picker rather than free text.
	Format string
	// Enum lists the values the spec allows; the console offers them
	// as a select instead of trusting free text.
	Enum []string
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
	return d.Servers[0].resolvedURL()
}

// serverURLs lists every declared server, variables resolved, so the
// console can offer them all instead of silently using the first.
func (d document) serverURLs() []string {
	var urls []string
	for _, declared := range d.Servers {
		if value := declared.resolvedURL(); value != "" {
			urls = append(urls, value)
		}
	}
	return urls
}

// bearerScheme names the HTTP bearer security scheme the document's global
// security requires, or "". That is the one shape the browser console can
// send without a custom flow.
func (d document) bearerScheme() string {
	for _, requirement := range d.Security {
		for name := range requirement {
			scheme, ok := d.Components.SecuritySchemes[name]
			if ok && strings.EqualFold(scheme.Type, "http") && strings.EqualFold(scheme.Scheme, "bearer") {
				return name
			}
		}
	}
	return ""
}

// apiKeyHeader names an apiKey-in-header security scheme the console can
// offer as a plain header field, or "".
func (d document) apiKeyHeader() (name, in string) {
	for _, requirement := range d.Security {
		for schemeName := range requirement {
			scheme, ok := d.Components.SecuritySchemes[schemeName]
			if ok && strings.EqualFold(scheme.Type, "apikey") && strings.EqualFold(scheme.In, "header") {
				return scheme.Name, schemeName
			}
		}
	}
	return "", ""
}

func (d document) operations() ([]Operation, error) {
	var out []Operation
	paths := make([]string, 0, len(d.Paths))
	for path := range d.Paths {
		if strings.ContainsAny(path, "?#") {
			return nil, fmt.Errorf("path %q carries a query or fragment; paths are locations, not URLs", path)
		}
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
			codes := sortedKeys(op.Responses)
			response := ""
			responseHeaders := []string{}
			if len(codes) > 0 {
				response = codes[0]
				responseHeaders = sortedKeys(op.Responses[response].Headers)
			}
			hasBody, bodyRequired, contentType, bodyExample := normalizeRequestBody(op.RequestBody, parameterSpecs, d.Consumes)
			out = append(out, Operation{
				Method: strings.ToUpper(method), Path: path, Summary: op.Summary,
				Description: op.Description, OperationID: op.OperationID,
				Parameters: params, ParameterSpecs: parameterSpecs,
				RequestBody: hasBody, RequestBodyRequired: bodyRequired,
				RequestBodyContentType: contentType, RequestBodyExample: bodyExample,
				Response: response, ResponseHeaders: responseHeaders,
				Deprecated: op.Deprecated, Tags: op.Tags,
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
		Format  string          `json:"format"`
		Enum    []string        `json:"enum"`
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
	param.Format = schema.Format
	if len(raw.Enum) > 0 {
		param.Enum = raw.Enum
	} else if len(schema.Enum) > 0 {
		param.Enum = schema.Enum
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
		exampleNames := sortedKeys(media.Examples)
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
	names := sortedKeys(values)
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
		// Parameter names are how readers remember an endpoint ("the one
		// with the limit param"); they belong in the filter's haystack.
		for _, parameter := range op.ParameterSpecs {
			fmt.Fprintf(&b, "%s ", parameter.Name)
		}
	}
	for name, item := range d.Components.Schemas {
		fmt.Fprintf(&b, "%s %s %s\n", name, item.Type, item.Description)
	}
	return b.String()
}

// sortedKeys lists a map's keys alphabetically, because operations,
// response headers, media types, and schemas all render in a stable
// order regardless of Go's map iteration.
func sortedKeys[V any](m map[string]V) []string {
	keys := make([]string, 0, len(m))
	for key := range m {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	return keys
}
