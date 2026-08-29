package openapi

import (
	"sort"
	"strings"

	"github.com/DonaldMurillo/gofastr/core/render"
	"github.com/DonaldMurillo/gofastr/framework/ui"
)

// Reference is the server-rendered OpenAPI reference screen contributed by
// Plugin. It is intentionally a normal GoFastr component, so projects can
// replace it, wrap it, or add actions without changing the Router contract.
type Reference struct {
	Title       string
	Description string
	Version     string
	ServerURL   string
	Operations  []Operation
	Schemas     map[string]Schema
}

func (r *Reference) Render() render.HTML {
	attrs := map[string]string{
		"class":                  "fastr-openapi-reference",
		"data-openapi-reference": "true",
	}
	if r.ServerURL != "" {
		attrs["data-openapi-server-url"] = r.ServerURL
	}
	return render.Tag("div", attrs,
		render.Tag("header", map[string]string{"class": "fastr-openapi-reference__header"},
			render.Tag("span", map[string]string{"class": "fastr-openapi-reference__eyebrow"}, render.Text("OpenAPI reference")),
			render.Tag("h1", nil, render.Text(r.Title)),
			render.Tag("p", nil, render.Text(r.Description)),
			render.Tag("div", map[string]string{"class": "fastr-openapi-reference__meta"},
				metaTag("OpenAPI "+r.Version),
				metaTag(intText(len(r.Operations))+" operations"),
				metaTag(intText(len(r.Schemas))+" schemas"),
				metaTag(firstNonEmpty(r.ServerURL, "No server configured")),
			),
		),
		render.Tag("div", map[string]string{"class": "fastr-openapi-reference__toolbar"},
			ui.SearchInput(ui.SearchInputConfig{
				Name:        "endpoint",
				ID:          "fastr-openapi-filter",
				Placeholder: "Filter endpoints…",
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
			"value":               operationID(i),
			"data-openapi-method": strings.ToUpper(op.Method),
			"data-openapi-path":   op.Path,
		}, render.Text(strings.ToUpper(op.Method)+" · "+op.Path)))
	}
	children := []render.HTML{
		render.Tag("strong", nil, render.Text("Try a request")),
		render.Tag("p", map[string]string{"class": "fastr-openapi-reference__server"}, render.Text(firstNonEmpty(r.ServerURL, "No OpenAPI server URL configured."))),
	}
	if len(options) > 0 {
		children = append(children,
			render.Tag("label", nil, render.Text("Operation"), render.Tag("select", map[string]string{"data-openapi-operation-select": "true"}, options...)),
			render.Tag("button", map[string]string{"type": "button", "data-openapi-try": "true"}, render.Text("Send request")),
			render.Tag("pre", map[string]string{"class": "fastr-openapi-reference__response", "data-openapi-response": "true"}, render.Text("Select an operation and send a request.")),
		)
	} else {
		children = append(children, render.Tag("p", nil, render.Text("No operations were found in this contract.")))
	}
	children = append(children, render.Tag("p", map[string]string{"class": "fastr-openapi-reference__note"}, render.Text("Requests run from the browser and require the API server to allow CORS.")))
	return render.Tag("aside", map[string]string{"class": "fastr-openapi-reference__console"}, children...)
}

func (r *Reference) operationIndex() render.HTML {
	items := make([]ui.RailItem, 0, len(r.Operations))
	for i, op := range r.Operations {
		id := operationID(i)
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
	for i, op := range r.Operations {
		children := []render.HTML{
			render.Tag("div", map[string]string{"class": "fastr-openapi-operation__route"},
				methodBadge(op.Method),
				render.Tag("code", nil, render.Text(op.Path)),
			),
			render.Tag("p", map[string]string{"class": "fastr-openapi-operation__summary"}, render.Text(firstNonEmpty(op.Summary, op.Description, "No summary provided."))),
		}
		if op.OperationID != "" {
			children = append(children, render.Tag("p", map[string]string{"class": "fastr-openapi-operation__id"}, render.Text("Operation ID: "+op.OperationID)))
		}
		if len(op.Parameters) > 0 || op.RequestBody || op.Response != "" {
			var details []render.HTML
			if len(op.Parameters) > 0 {
				params := make([]render.HTML, 0, len(op.Parameters))
				for _, param := range op.Parameters {
					params = append(params, render.Tag("li", nil, render.Text(param)))
				}
				details = append(details, render.Tag("div", nil, render.Tag("h3", nil, render.Text("Parameters")), render.Tag("ul", nil, params...)))
			}
			if op.RequestBody {
				details = append(details, render.Tag("div", nil, render.Tag("h3", nil, render.Text("Request body")), render.Tag("p", nil, render.Text("Request body required by the contract."))))
			}
			if op.Response != "" {
				details = append(details, render.Tag("div", nil, render.Tag("h3", nil, render.Text("Response")), render.Tag("code", nil, render.Text(op.Response))))
			}
			children = append(children, render.Tag("div", map[string]string{"class": "fastr-openapi-operation__details"}, details...))
		}
		items = append(items, render.Tag("article", map[string]string{"id": operationID(i), "class": "fastr-openapi-operation", "data-openapi-operation": "true", "data-openapi-search": strings.ToLower(op.Method + " " + op.Path + " " + op.Summary + " " + op.OperationID)}, children...))
	}
	if len(r.Schemas) > 0 {
		names := make([]string, 0, len(r.Schemas))
		for name := range r.Schemas {
			names = append(names, name)
		}
		sort.Strings(names)
		for _, name := range names {
			schema := r.Schemas[name]
			properties := make([]render.HTML, 0, len(schema.Properties))
			propertyNames := make([]string, 0, len(schema.Properties))
			for property := range schema.Properties {
				propertyNames = append(propertyNames, property)
			}
			sort.Strings(propertyNames)
			for _, property := range propertyNames {
				item := schema.Properties[property]
				properties = append(properties, render.Tag("li", nil, render.Tag("code", nil, render.Text(property)), render.Text(" · "+item.Type)))
			}
			items = append(items, render.Tag("section", map[string]string{"class": "fastr-openapi-schema", "data-openapi-schema": name}, render.Tag("h2", nil, render.Text(name)), render.Tag("p", nil, render.Text(firstNonEmpty(schema.Description, "Schema model"))), render.Tag("ul", nil, properties...)))
		}
	}
	return items
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

func operationID(index int) string { return "fastr-openapi-operation-" + intText(index+1) }

func intText(value int) string {
	if value == 0 {
		return "0"
	}
	var digits [20]byte
	i := len(digits)
	for value > 0 {
		i--
		digits[i] = byte('0' + value%10)
		value /= 10
	}
	return string(digits[i:])
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
		".fastr-openapi-reference__console { display: grid; gap: 10px; } .fastr-openapi-reference__console label { display: grid; gap: 5px; color: var(--color-text-muted, #71717a); font-size: .75rem; } .fastr-openapi-reference__console select { width: 100%; padding: 8px; border: 1px solid var(--color-border, #e4e4e7); border-radius: 6px; background: var(--color-surface, #fff); color: inherit; } .fastr-openapi-reference__console button { padding: 9px 11px; border: 0; border-radius: 6px; background: var(--color-primary, #4f46e5); color: var(--color-primary-fg, #fff); cursor: pointer; }",
		".fastr-openapi-reference__console p { margin: 0; color: var(--color-text-muted, #71717a); font-size: .8rem; line-height: 1.6; } .fastr-openapi-reference__server { overflow-wrap: anywhere; font-family: ui-monospace, monospace; } .fastr-openapi-reference__response { max-height: 220px; overflow: auto; padding: 10px; border: 1px solid var(--color-border, #e4e4e7); border-radius: 6px; background: var(--color-surface-soft, #f4f4f5); color: var(--color-text, #18181b); font-size: .72rem; white-space: pre-wrap; }",
		"@media (max-width: 1050px) { .fastr-openapi-reference__layout { grid-template-columns: 190px minmax(0, 1fr); }.fastr-openapi-reference__layout > .scrollspy { position: static; }.fastr-openapi-reference__console { grid-column: 1 / -1; position: static; } }",
		"@media (max-width: 680px) { .fastr-openapi-reference__layout { display: block; }.fastr-openapi-reference__index { margin-bottom: 12px; }.fastr-openapi-reference__index a:nth-child(n+7) { display: none; } }",
	}, "\n")
}
