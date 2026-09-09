package openapi

import (
	"context"
	"strconv"
	"strings"

	"github.com/DonaldMurillo/gofastr/core/render"
	"github.com/DonaldMurillo/gofastr/framework/ui"
)

// Strings are the labels the reference surface renders. A translated mount
// supplies its own; every empty field falls back to the English default, so
// a translation can land a label at a time the way UIStrings does.
type Strings struct {
	// Eyebrow is the small label above the title. Title case in English
	// ("OpenAPI reference"); translate in the style your language uses for
	// eyebrow labels.
	Eyebrow string
	// OperationsLabel is the unit word after the operation count in the meta
	// row, rendered as "3 operations". Lowercase in English because it
	// follows the number; follow your language's convention for counted
	// nouns.
	OperationsLabel string
	// SchemasLabel is the unit word after the schema count, "2 schemas".
	// Same casing rule as OperationsLabel.
	SchemasLabel string
	// NoServer fills the meta row when the spec declares no server and
	// none was overridden. Sentence case.
	NoServer string
	// FilterPlaceholder is the endpoint filter's placeholder. An ellipsis
	// is conventional; keep it if your language does too.
	FilterPlaceholder string
	// TryRequest heads the request console. Sentence case, imperative in
	// English.
	TryRequest string
	// ConsoleNoServer is the console paragraph shown instead of a server
	// URL when none resolved. Sentence case.
	ConsoleNoServer string
	// OperationLabel captions the console's operation dropdown.
	OperationLabel string
	// ServerLabel captions the server dropdown a multi-server spec
	// renders. One word.
	ServerLabel string
	// TokenPlaceholder is the placeholder of the bearer token field, for
	// specs the console can authenticate.
	TokenPlaceholder string
	// DeprecatedLabel marks an operation the spec retired. One word.
	DeprecatedLabel string
	// ResponseHeadersLabel captions the documented response headers.
	ResponseHeadersLabel string
	// SendRequest is the console button. Imperative.
	SendRequest string
	// CopyCurlLabel is the button that emits the prepared request as a
	// curl command. Sentence case.
	CopyCurlLabel string
	// ResponsePrompt is the response pane's initial text and describes the
	// action, not the pane. Full sentence, capitalized.
	ResponsePrompt string
	// NoOperations replaces the console controls when the spec has no
	// operations. Full sentence, capitalized.
	NoOperations string
	// CORSNote is the footnote under the console. Full sentence,
	// capitalized.
	CORSNote string
	// NoInputs is shown per operation that has no parameters and no request
	// body. Full sentence, capitalized.
	NoInputs string
	// OperationIDPrefix precedes the raw operation id and includes its own
	// trailing separator and space: "Operation ID: listProjects".
	OperationIDPrefix string
	// ParametersLabel, RequestBodyLabel, and ResponseLabel head the three
	// detail sections of an operation card. Cased as headings.
	ParametersLabel  string
	RequestBodyLabel string
	// RequestBodyNote is the sentence under the request body heading.
	RequestBodyNote string
	ResponseLabel   string
	// ValuePlaceholder is the fallback placeholder of a parameter input
	// when the spec offers no example and no type. Single word.
	ValuePlaceholder string
	// ParameterWord fills a parameter's label when the spec does not
	// say where it goes (path, query, ...). Single word, lowercase in
	// English, because it sits between name and type.
	ParameterWord string
	// RequiredWord marks a parameter or body as required, appended as
	// " · required". Lowercase in English; the separator is added by the
	// renderer.
	RequiredWord string
	// NoSummary fills an operation card whose operation declares neither
	// summary nor description. Full sentence, capitalized.
	NoSummary string
}

// DefaultStrings returns the English surface labels.
func DefaultStrings() Strings {
	return Strings{
		Eyebrow:              "OpenAPI reference",
		OperationsLabel:      "operations",
		SchemasLabel:         "schemas",
		NoServer:             "No server configured",
		FilterPlaceholder:    "Filter endpoints…",
		TryRequest:           "Try a request",
		ConsoleNoServer:      "No OpenAPI server URL configured.",
		OperationLabel:       "Operation",
		ServerLabel:          "Server",
		TokenPlaceholder:     "Bearer token",
		DeprecatedLabel:      "Deprecated",
		ResponseHeadersLabel: "Response headers",
		SendRequest:          "Send request",
		ResponsePrompt:       "Select an operation and send a request.",
		NoOperations:         "No operations were found in this spec.",
		CORSNote:             "Requests run from the browser and require the API server to allow CORS.",
		NoInputs:             "This operation has no request inputs.",
		OperationIDPrefix:    "Operation ID: ",
		ParametersLabel:      "Parameters",
		RequestBodyLabel:     "Request body",
		RequestBodyNote:      "Request body required by the spec.",
		ResponseLabel:        "Response",
		ValuePlaceholder:     "Value",
		ParameterWord:        "parameter",
		RequiredWord:         "required",
		NoSummary:            "No summary provided.",
	}
}

func (s Strings) withDefaults() Strings {
	def := DefaultStrings()
	s.Eyebrow = firstNonEmpty(s.Eyebrow, def.Eyebrow)
	s.OperationsLabel = firstNonEmpty(s.OperationsLabel, def.OperationsLabel)
	s.SchemasLabel = firstNonEmpty(s.SchemasLabel, def.SchemasLabel)
	s.NoServer = firstNonEmpty(s.NoServer, def.NoServer)
	s.FilterPlaceholder = firstNonEmpty(s.FilterPlaceholder, def.FilterPlaceholder)
	s.TryRequest = firstNonEmpty(s.TryRequest, def.TryRequest)
	s.ConsoleNoServer = firstNonEmpty(s.ConsoleNoServer, def.ConsoleNoServer)
	s.OperationLabel = firstNonEmpty(s.OperationLabel, def.OperationLabel)
	s.SendRequest = firstNonEmpty(s.SendRequest, def.SendRequest)
	s.ResponsePrompt = firstNonEmpty(s.ResponsePrompt, def.ResponsePrompt)
	s.NoOperations = firstNonEmpty(s.NoOperations, def.NoOperations)
	s.CORSNote = firstNonEmpty(s.CORSNote, def.CORSNote)
	s.NoInputs = firstNonEmpty(s.NoInputs, def.NoInputs)
	s.OperationIDPrefix = firstNonEmpty(s.OperationIDPrefix, def.OperationIDPrefix)
	s.ParametersLabel = firstNonEmpty(s.ParametersLabel, def.ParametersLabel)
	s.RequestBodyLabel = firstNonEmpty(s.RequestBodyLabel, def.RequestBodyLabel)
	s.RequestBodyNote = firstNonEmpty(s.RequestBodyNote, def.RequestBodyNote)
	s.ResponseLabel = firstNonEmpty(s.ResponseLabel, def.ResponseLabel)
	s.ValuePlaceholder = firstNonEmpty(s.ValuePlaceholder, def.ValuePlaceholder)
	s.ParameterWord = firstNonEmpty(s.ParameterWord, def.ParameterWord)
	s.RequiredWord = firstNonEmpty(s.RequiredWord, def.RequiredWord)
	s.NoSummary = firstNonEmpty(s.NoSummary, def.NoSummary)
	return s
}

// Reference is the server-rendered OpenAPI reference screen contributed by
// Plugin. It is intentionally a normal GoFastr component, so projects can
// replace it, wrap it, or add actions without changing the Router API.
type Reference struct {
	// IDPrefix discriminates this mount's element ids (operation cards,
	// console option values) so two mounts never collide on one host.
	IDPrefix string

	// Title heads the reference page.
	Title string
	// Description is the page's intro line.
	Description string
	// Version is the document's OpenAPI version, shown in the meta row.
	Version string
	// ServerURL is where the request console sends; empty shows the
	// no-server note.
	ServerURL string
	// Servers lists every declared server. More than one renders a
	// selector, so the second declared server is reachable.
	Servers []string
	// BearerScheme names an HTTP bearer security scheme the document
	// requires; the console offers a token field for it.
	BearerScheme string
	// APIKeyHeader names an apiKey security header the console offers as a
	// plain field; APIKeyName is its scheme name for the label.
	APIKeyHeader, APIKeyName string
	// Operations are the document's operations, in document order.
	Operations []Operation
	// Schemas are the document's component schemas.
	Schemas map[string]Schema
	// Strings labels the surface; the zero value falls back to
	// DefaultStrings field by field.
	Strings Strings
}

// Render draws the whole reference surface: header and meta, the endpoint
// filter, the operation index and cards, and the request console.
// RenderCtx renders per request. The surface carries no request-dependent
// chrome yet, so it delegates to Render; implementing the method keeps the
// component usable wherever GoFastr prefers context-aware rendering.
func (r *Reference) RenderCtx(ctx context.Context) render.HTML {
	return r.Render()
}

func (r *Reference) Render() render.HTML {
	attrs := map[string]string{
		"class":                  "fastr-openapi-reference",
		"data-openapi-reference": "true",
		// The console flips this while a request is in flight; markup
		// carries the state even before scripts run.
		"aria-busy": "false",
	}
	if r.ServerURL != "" {
		attrs["data-openapi-server-url"] = r.ServerURL
	}
	if r.BearerScheme != "" {
		attrs["data-openapi-bearer"] = r.BearerScheme
	}
	if r.APIKeyHeader != "" {
		attrs["data-openapi-api-key"] = r.APIKeyHeader
	}
	return render.Tag("div", attrs,
		render.Tag("header", map[string]string{"class": "fastr-openapi-reference__header"},
			render.Tag("span", map[string]string{"class": "fastr-openapi-reference__eyebrow"}, render.Text(r.Strings.Eyebrow)),
			render.Tag("h1", nil, render.Text(r.Title)),
			render.Tag("p", nil, render.Text(r.Description)),
			render.Tag("div", map[string]string{"class": "fastr-openapi-reference__meta"},
				metaTag("OpenAPI "+r.Version),
				metaTag(strconv.Itoa(len(r.Operations))+" "+r.Strings.OperationsLabel),
				metaTag(strconv.Itoa(len(r.Schemas))+" "+r.Strings.SchemasLabel),
				metaTag(firstNonEmpty(r.ServerURL, r.Strings.NoServer)),
			),
		),
		render.Tag("div", map[string]string{"class": "fastr-openapi-reference__toolbar"},
			ui.SearchInput(ui.SearchInputConfig{
				Name:        "endpoint",
				ID:          "fastr-openapi-filter",
				Placeholder: r.Strings.FilterPlaceholder,
				ExtraAttrs:  map[string]string{"data-openapi-filter": "true", "autocomplete": "off"},
			}),
		),
		render.Tag("div", map[string]string{"class": "fastr-openapi-reference__layout"},
			r.operationIndex(),
			render.Tag("section", map[string]string{"class": "fastr-openapi-reference__operations"}, r.operationCards()...),
			r.requestConsole(),
		),
	)
}

func (r *Reference) requestConsole() render.HTML {
	options := make([]render.HTML, 0, len(r.Operations))
	for i, op := range r.Operations {
		options = append(options, render.Tag("option", map[string]string{
			"value":               r.operationID(i),
			"role":                "option",
			"data-openapi-method": strings.ToUpper(op.Method),
			"data-openapi-path":   op.Path,
		}, render.Text(strings.ToUpper(op.Method)+" · "+op.Path+optionSummary(op))))
	}
	children := []render.HTML{
		render.Tag("strong", nil, render.Text(r.Strings.TryRequest)),
	}
	// A multi-server spec gets a selector; a single one stays prose.
	if len(r.Servers) > 1 {
		options := make([]render.HTML, 0, len(r.Servers))
		for _, server := range r.Servers {
			options = append(options, render.Tag("option", map[string]string{"value": server}, render.Text(server)))
		}
		children = append(children,
			render.Tag("label", nil, render.Text(r.Strings.ServerLabel), render.Tag("select", map[string]string{"data-openapi-servers": "true"}, options...)))
	}
	children = append(children, render.Tag("p", map[string]string{"class": "fastr-openapi-reference__server", "data-openapi-server-note": "true"}, render.Text(firstNonEmpty(r.ServerURL, r.Strings.ConsoleNoServer))))
	if r.BearerScheme != "" {
		children = append(children, render.Tag("label", map[string]string{"class": "fastr-openapi-reference__field"},
			render.Text(r.BearerScheme+" token"),
			render.Tag("input", map[string]string{"type": "password", "data-openapi-token": "true", "autocomplete": "off", "placeholder": r.Strings.TokenPlaceholder})))
	}
	if r.APIKeyHeader != "" {
		children = append(children, render.Tag("label", map[string]string{"class": "fastr-openapi-reference__field"},
			render.Text(firstNonEmpty(r.APIKeyName, r.APIKeyHeader)+" header ("+r.APIKeyHeader+")"),
			render.Tag("input", map[string]string{"type": "text", "data-openapi-api-key-value": "true", "autocomplete": "off", "placeholder": r.APIKeyHeader})))
	}
	if len(options) > 0 {
		children = append(children,
			render.Tag("label", nil, render.Text(r.Strings.OperationLabel), render.Tag("select", map[string]string{"data-openapi-operation-select": "true"}, options...)),
			r.requestInputs(),
			render.Join(
				render.Tag("button", map[string]string{"type": "button", "data-openapi-try": "true"}, render.Text(r.Strings.SendRequest)),
				render.Tag("button", map[string]string{"type": "button", "data-openapi-curl": "true"}, render.Text(r.Strings.CopyCurlLabel)),
				render.Tag("pre", map[string]string{"class": "fastr-openapi-reference__response", "role": "status", "aria-live": "polite", "data-openapi-response": "true"}, render.Text(r.Strings.ResponsePrompt)),
			))
	} else {
		children = append(children, render.Tag("p", nil, render.Text(r.Strings.NoOperations)))
	}
	children = append(children, render.Tag("p", map[string]string{"class": "fastr-openapi-reference__note"}, render.Text(r.Strings.CORSNote)))
	return render.Tag("aside", map[string]string{"class": "fastr-openapi-reference__console"}, children...)
}

func (r *Reference) requestInputs() render.HTML {
	groups := make([]render.HTML, 0, len(r.Operations))
	for index, op := range r.Operations {
		operation := r.operationID(index)
		attrs := map[string]string{
			"class":                   "fastr-openapi-reference__inputs",
			"data-openapi-inputs-for": operation,
		}
		if index > 0 {
			attrs["hidden"] = "hidden"
		}
		fields := make([]render.HTML, 0, len(op.ParameterSpecs)+1)
		for _, parameter := range op.ParameterSpecs {
			label := parameter.Name + " · " + firstNonEmpty(parameter.In, r.Strings.ParameterWord)
			if parameter.Type != "" {
				label += " · " + parameter.Type
			}
			if parameter.Required {
				label += " · " + r.Strings.RequiredWord
			}
			if parameter.Description != "" {
				label += " — " + parameter.Description
			}
			// An enum is a closed set: offer it as choices rather than
			// trusting free text the spec would reject.
			if len(parameter.Enum) > 0 {
				options := make([]render.HTML, 0, len(parameter.Enum))
				for _, value := range parameter.Enum {
					optionAttrs := map[string]string{"value": value}
					if value == firstNonEmpty(parameter.Example, parameter.Default, parameter.Enum[0]) {
						optionAttrs["selected"] = "selected"
					}
					options = append(options, render.Tag("option", optionAttrs, render.Text(value)))
				}
				fields = append(fields, render.Tag("label", map[string]string{"class": "fastr-openapi-reference__field"}, render.Text(label),
					render.Tag("select", map[string]string{
						"data-openapi-param-name":     parameter.Name,
						"data-openapi-param-in":       strings.ToLower(parameter.In),
						"data-openapi-param-required": boolText(parameter.Required),
					}, options...)))
				continue
			}
			fieldAttrs := map[string]string{
				"type":                        parameterInputType(parameter),
				"data-openapi-param-name":     parameter.Name,
				"data-openapi-param-in":       strings.ToLower(parameter.In),
				"data-openapi-param-required": boolText(parameter.Required),
				"value":                       firstNonEmpty(parameter.Example, parameter.Default),
				"placeholder":                 firstNonEmpty(parameter.Example, parameter.Type, r.Strings.ValuePlaceholder),
			}
			fields = append(fields, render.Tag("label", map[string]string{"class": "fastr-openapi-reference__field"}, render.Text(label), render.Tag("input", fieldAttrs)))
		}
		if op.RequestBody {
			bodyAttrs := map[string]string{
				"data-openapi-body":          "true",
				"data-openapi-content-type":  firstNonEmpty(op.RequestBodyContentType, "application/json"),
				"data-openapi-body-required": boolText(op.RequestBodyRequired),
				"rows":                       "7",
				"placeholder":                firstNonEmpty(op.RequestBodyExample, "{\n  \"key\": \"value\"\n}"),
			}
			body := render.Tag("textarea", bodyAttrs, render.Text(op.RequestBodyExample))
			label := r.Strings.RequestBodyLabel + " · " + firstNonEmpty(op.RequestBodyContentType, "application/json")
			if op.RequestBodyRequired {
				label += " · " + r.Strings.RequiredWord
			}
			fields = append(fields, render.Tag("label", map[string]string{"class": "fastr-openapi-reference__field"}, render.Text(label), body))
		}
		if len(fields) == 0 {
			fields = append(fields, render.Tag("p", map[string]string{"class": "fastr-openapi-reference__no-inputs"}, render.Text(r.Strings.NoInputs)))
		}
		groups = append(groups, render.Tag("div", attrs, fields...))
	}
	return render.Tag("div", map[string]string{"class": "fastr-openapi-reference__input-groups"}, groups...)
}

func (r *Reference) operationIndex() render.HTML {
	items := make([]ui.RailItem, 0, len(r.Operations))
	for i, op := range r.Operations {
		id := r.operationID(i)
		items = append(items, ui.RailItem{Anchor: id, Text: op.Path, Eyebrow: strings.ToUpper(op.Method)})
	}
	if len(items) == 0 {
		return render.Text("")
	}
	return ui.AnchoredRail(ui.AnchoredRailConfig{
		Label:           "API operations",
		Items:           items,
		ObserveSelector: ".fastr-openapi-reference__operations",
		TargetSelector:  ".fastr-openapi-operation[id]",
		Class:           "fastr-openapi-reference__index",
	})
}

func (r *Reference) operationCards() []render.HTML {
	items := make([]render.HTML, 0, len(r.Operations)+len(r.Schemas))
	// Operations render grouped: the document's own tag when it carries
	// one, else the first path segment. The group wrapper is what the
	// filter hides once every operation inside it is filtered out.
	groups := make([]render.HTML, 0, len(r.Operations))
	var currentGroup string
	var currentChildren []render.HTML
	flush := func() {
		if len(currentChildren) == 0 {
			return
		}
		groups = append(groups, render.Tag("section",
			map[string]string{"class": "fastr-openapi-group", "data-openapi-group": currentGroup},
			render.Join(
				render.Tag("h2", map[string]string{"class": "fastr-openapi-group__title"}, render.Text(currentGroup)),
				render.Join(currentChildren...))))
		currentChildren = nil
	}
	for i, op := range r.Operations {
		group := operationGroup(op)
		if group != currentGroup {
			flush()
			currentGroup = group
		}
		cardAttrs := map[string]string{"id": r.operationID(i), "class": "fastr-openapi-operation", "data-openapi-operation": "true", "data-openapi-search": operationSearchText(op)}
		if op.Deprecated {
			cardAttrs["data-deprecated"] = "true"
		}
		children := []render.HTML{
			render.Tag("div", map[string]string{"class": "fastr-openapi-operation__route"},
				methodBadge(op.Method),
				render.Tag("code", nil, render.Text(op.Path)),
			),
			render.Tag("p", map[string]string{"class": "fastr-openapi-operation__summary"}, render.Text(firstNonEmpty(op.Summary, op.Description, op.OperationID, r.Strings.NoSummary))),
		}
		if op.Deprecated {
			children = append(children, render.Tag("p", map[string]string{"class": "fastr-openapi-operation__deprecated"}, render.Text(r.Strings.DeprecatedLabel)))
		}
		if op.OperationID != "" {
			children = append(children, render.Tag("p", map[string]string{"class": "fastr-openapi-operation__id"}, render.Text(r.Strings.OperationIDPrefix+op.OperationID)))
		}
		if len(op.Parameters) > 0 || op.RequestBody || op.Response != "" {
			var details []render.HTML
			if len(op.Parameters) > 0 {
				params := make([]render.HTML, 0, len(op.Parameters))
				for _, param := range op.Parameters {
					params = append(params, render.Tag("li", nil, render.Text(param)))
				}
				details = append(details, render.Tag("div", nil, render.Tag("h3", nil, render.Text(r.Strings.ParametersLabel)), render.Tag("ul", nil, params...)))
			}
			if op.RequestBody {
				details = append(details, render.Tag("div", nil, render.Tag("h3", nil, render.Text(r.Strings.RequestBodyLabel)), render.Tag("p", nil, render.Text(r.Strings.RequestBodyNote))))
			}
			if op.Response != "" {
				details = append(details, render.Tag("div", nil, render.Tag("h3", nil, render.Text(r.Strings.ResponseLabel)), render.Tag("code", nil, render.Text(op.Response))))
			}
			if len(op.ResponseHeaders) > 0 {
				headers := make([]render.HTML, 0, len(op.ResponseHeaders))
				for _, name := range op.ResponseHeaders {
					headers = append(headers, render.Tag("li", nil, render.Tag("code", nil, render.Text(name))))
				}
				details = append(details, render.Tag("div", nil, render.Tag("h3", nil, render.Text(r.Strings.ResponseHeadersLabel)), render.Tag("ul", nil, headers...)))
			}
			children = append(children, render.Tag("div", map[string]string{"class": "fastr-openapi-operation__details"}, details...))
		}
		currentChildren = append(currentChildren, render.Tag("article", cardAttrs, children...))
	}
	flush()
	items = append(items, groups...)
	if len(r.Schemas) > 0 {
		for _, name := range sortedKeys(r.Schemas) {
			schema := r.Schemas[name]
			properties := make([]render.HTML, 0, len(schema.Properties))
			for _, property := range sortedKeys(schema.Properties) {
				item := schema.Properties[property]
				properties = append(properties, render.Tag("li", nil, render.Tag("code", nil, render.Text(property)), render.Text(" · "+item.Type)))
			}
			items = append(items, render.Tag("section", map[string]string{"class": "fastr-openapi-schema", "data-openapi-schema": name}, render.Tag("h2", nil, render.Text(name)), render.Tag("p", nil, render.Text(firstNonEmpty(schema.Description, "Schema model"))), render.Tag("ul", nil, properties...)))
		}
	}
	return items
}

// operationGroup names the bucket an operation renders under: the
// document's own tag when present, else the first path segment.
// optionSummary names an operation in the console select so two routes
// with the same method and path stay tellable apart.
func optionSummary(op Operation) string {
	if summary := strings.TrimSpace(op.Summary); summary != "" {
		return " — " + summary
	}
	if id := strings.TrimSpace(op.OperationID); id != "" {
		return " — " + id
	}
	return ""
}

func operationGroup(op Operation) string {
	if len(op.Tags) > 0 && strings.TrimSpace(op.Tags[0]) != "" {
		return strings.TrimSpace(op.Tags[0])
	}
	trimmed := strings.Trim(op.Path, "/")
	if trimmed == "" {
		return "api"
	}
	if slash := strings.IndexByte(trimmed, '/'); slash > 0 {
		return trimmed[:slash]
	}
	return trimmed
}

func inputType(parameterType string) string {
	switch strings.ToLower(parameterType) {
	case "integer", "number":
		return "number"
	case "email":
		return "email"
	default:
		return "text"
	}
}

// parameterInputType lets a format refine the input a parameter renders:
// a date format becomes a date picker instead of free text.
func parameterInputType(parameter Parameter) string {
	if strings.EqualFold(parameter.Format, "date") {
		return "date"
	}
	if strings.EqualFold(parameter.Format, "date-time") {
		return "datetime-local"
	}
	return inputType(parameter.Type)
}

func boolText(value bool) string {
	if value {
		return "true"
	}
	return "false"
}

func metaTag(value string) render.HTML {
	return ui.StatusBadge(ui.StatusBadgeConfig{Label: value, Variant: ui.StatusNeutral, Class: "fastr-openapi-reference__tag"})
}

func methodBadge(method string) render.HTML {
	variant := ui.StatusInfo
	switch strings.ToUpper(method) {
	case "GET", "HEAD":
		variant = ui.StatusSuccess
	case "POST", "PUT", "PATCH":
		variant = ui.StatusWarning
	case "DELETE":
		variant = ui.StatusDanger
	}
	return ui.StatusBadge(ui.StatusBadgeConfig{
		Label:   strings.ToUpper(method),
		Variant: variant,
		Class:   "fastr-openapi-method fastr-openapi-method--" + strings.ToLower(method),
	})
}

func (r *Reference) operationID(index int) string {
	prefix := "fastr-openapi-operation"
	if r.IDPrefix != "" {
		prefix += "-" + r.IDPrefix
	}
	return prefix + "-" + strconv.Itoa(index+1)
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return value
		}
	}
	return ""
}

// CSS is the default structural style for Reference. It intentionally uses
// CSS variables so the generated project can completely rebrand it.
func CSS() string {
	return strings.Join([]string{
		".fastr-openapi-reference { max-width: 1180px; margin: 0 auto; padding: 32px clamp(20px, 4vw, 56px) 72px; color: var(--color-text, #18181b); }",
		".fastr-openapi-reference select:focus-visible, .fastr-openapi-reference input:focus-visible, .fastr-openapi-reference button:focus-visible { outline: 2px solid var(--color-accent, #0891b2); outline-offset: 2px; }",
		".fastr-openapi-reference__response[data-state='loading'] { opacity: .6; }",
		".fastr-openapi-reference__header h1 { margin: 8px 0; font-size: clamp(2rem, 4vw, 3.5rem); letter-spacing: -.04em; }",
		".fastr-openapi-reference__header p { max-width: 70ch; color: var(--color-text-muted, #52525b); line-height: 1.7; }",
		".fastr-openapi-reference__eyebrow, .fastr-openapi-reference__tag { color: var(--color-primary, #4f46e5); font-size: .72rem; font-weight: 700; letter-spacing: .08em; text-transform: uppercase; }",
		".fastr-openapi-reference__meta { display: flex; flex-wrap: wrap; gap: 8px; margin-top: 18px; }",
		".fastr-openapi-reference__tag { padding: 5px 8px; border: 1px solid var(--color-border, #e4e4e7); border-radius: 999px; color: var(--color-text-muted, #52525b); letter-spacing: .02em; text-transform: none; }",
		".fastr-openapi-reference__toolbar { margin: 28px 0 14px; } .fastr-openapi-reference__toolbar .ui-search-input { width: min(420px, 100%); }",
		".fastr-openapi-reference__layout { display: grid; grid-template-columns: 190px minmax(0, 1fr) 220px; gap: 16px; align-items: start; }",
		".fastr-openapi-reference__layout > .scrollspy { position: sticky; top: 20px; align-self: start; min-width: 0; } .fastr-openapi-reference__index, .fastr-openapi-reference__console { padding: 14px; border: 1px solid var(--color-border, #e4e4e7); border-radius: 8px; background: var(--color-surface, #fff); }",
		".fastr-openapi-reference__index { display: grid; gap: 5px; }",
		".fastr-openapi-reference__index a { display: flex; gap: 7px; padding: 7px; border-radius: 5px; color: var(--color-text-muted, #52525b); text-decoration: none; font-size: .78rem; } .fastr-openapi-reference__index code { overflow: hidden; text-overflow: ellipsis; white-space: nowrap; }",
		".fastr-openapi-reference__index a:hover, .fastr-openapi-reference__index a.is-active { background: var(--color-surface-soft, #f4f4f5); color: var(--color-primary, #4f46e5); }",
		".fastr-openapi-reference__index .ui-anchored-rail__eyebrow { min-width: 34px; padding: 3px 4px; border-radius: 4px; background: var(--color-surface-soft, #f4f4f5); font-size: .62rem; font-weight: 700; letter-spacing: .04em; text-align: center; }",
		".fastr-openapi-method { display: inline-flex; min-width: 38px; justify-content: center; padding: 3px 4px; border-radius: 4px; font-size: .65rem; font-weight: 700; letter-spacing: .04em; }",
		".fastr-openapi-method--get { color: #287c54; background: #e7f5ed; }.fastr-openapi-method--post { color: #a65b1c; background: #fff0df; }.fastr-openapi-method--delete { color: #ad3b3b; background: #fde8e8; }",
		".fastr-openapi-operation, .fastr-openapi-schema { margin-bottom: 12px; padding: 18px; border: 1px solid var(--color-border, #e4e4e7); border-radius: 8px; background: var(--color-surface, #fff); }",
		".fastr-openapi-operation__route { display: flex; align-items: center; gap: 9px; font-size: 1rem; }",
		".fastr-openapi-operation__summary { margin: 12px 0 0; color: var(--color-text-muted, #52525b); line-height: 1.6; }",
		".fastr-openapi-operation__id { color: var(--color-text-muted, #71717a); font-family: ui-monospace, monospace; font-size: .75rem; }",
		".fastr-openapi-operation__details { display: grid; grid-template-columns: repeat(auto-fit, minmax(150px, 1fr)); gap: 16px; margin-top: 18px; padding-top: 14px; border-top: 1px solid var(--color-border, #e4e4e7); color: var(--color-text-muted, #52525b); font-size: .8rem; }",
		".fastr-openapi-operation__details h3 { margin: 0 0 7px; font-size: .7rem; text-transform: uppercase; letter-spacing: .08em; }",
		".fastr-openapi-operation__details ul, .fastr-openapi-schema ul { margin: 0; padding-left: 18px; }",
		".fastr-openapi-schema h2 { margin: 0 0 6px; } .fastr-openapi-schema p { color: var(--color-text-muted, #52525b); }",
		".fastr-openapi-reference__console { display: grid; gap: 10px; } .fastr-openapi-reference__console label { display: grid; gap: 5px; color: var(--color-text-muted, #71717a); font-size: .75rem; } .fastr-openapi-reference__console select, .fastr-openapi-reference__console input, .fastr-openapi-reference__console textarea { width: 100%; box-sizing: border-box; padding: 8px; border: 1px solid var(--color-border, #e4e4e7); border-radius: 6px; background: var(--color-surface, #fff); color: inherit; font: inherit; } .fastr-openapi-reference__console textarea { min-height: 120px; resize: vertical; font-family: ui-monospace, monospace; font-size: .75rem; line-height: 1.5; } .fastr-openapi-reference__console button { padding: 9px 11px; border: 0; border-radius: 6px; background: var(--color-primary, #4f46e5); color: var(--color-primary-fg, #fff); cursor: pointer; }",
		".fastr-openapi-reference__input-groups { display: grid; gap: 9px; } .fastr-openapi-reference__inputs { display: grid; gap: 9px; } .fastr-openapi-reference__no-inputs { margin: 0; color: var(--color-text-muted, #71717a); font-size: .8rem; }",
		".fastr-openapi-reference__console p { margin: 0; color: var(--color-text-muted, #71717a); font-size: .8rem; line-height: 1.6; } .fastr-openapi-reference__server { overflow-wrap: anywhere; font-family: ui-monospace, monospace; } .fastr-openapi-reference__response { max-height: 220px; overflow: auto; padding: 10px; border: 1px solid var(--color-border, #e4e4e7); border-radius: 6px; background: var(--color-surface-soft, #f4f4f5); color: var(--color-text, #18181b); font-size: .72rem; white-space: pre-wrap; }",
		"@media (max-width: 1050px) { .fastr-openapi-reference__layout { grid-template-columns: 190px minmax(0, 1fr); }.fastr-openapi-reference__layout > .scrollspy { position: static; }.fastr-openapi-reference__console { grid-column: 1 / -1; position: static; } }",
		"@media (max-width: 680px) { .fastr-openapi-reference__layout { display: block; }.fastr-openapi-reference__index { margin-bottom: 12px; }.fastr-openapi-reference__index a:nth-child(n+7) { display: none; } }",
		// Grouped operations: the section heading names its slice of the API.
		".fastr-openapi-group { margin-top: 34px; }",
		".fastr-openapi-group__title { margin: 0 0 10px; font-size: 15px; font-weight: 650; letter-spacing: .02em; text-transform: uppercase; color: var(--color-text-muted, #52525b); }",
		".fastr-openapi-group[hidden] { display: none; }",
		// A retired operation says so in more than an attribute.
		".fastr-openapi-operation[data-deprecated] { opacity: .72; }",
		".fastr-openapi-operation__deprecated { margin: 4px 0 0; font-size: 12px; font-weight: 650; letter-spacing: .04em; text-transform: uppercase; color: #b3261e; }",
		// Typed inputs keep the console column's shape.
		".fastr-openapi-reference__field input[type=\"date\"], .fastr-openapi-reference__field input[type=\"datetime-local\"] { min-width: 0; width: 100%; }",
		// The filtered-to-nothing and in-flight states have a look.
		".fastr-openapi-reference__operationsEmpty, .fastr-openapi-empty { padding: 18px 4px; color: var(--color-text-muted, #52525b); font-style: italic; }",
		".fastr-openapi-reference__response[data-state=\"loading\"] { opacity: .7; }",
	}, "\n")
}

// operationSearchText is the filter's haystack: method, path, summary,
// operation id, and the parameter names readers remember endpoints by.
func operationSearchText(op Operation) string {
	parts := []string{op.Method, op.Path, op.Summary, op.OperationID}
	for _, parameter := range op.ParameterSpecs {
		parts = append(parts, parameter.Name)
	}
	return strings.ToLower(strings.Join(parts, " "))
}
