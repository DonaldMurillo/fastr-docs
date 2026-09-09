package openapi

// Server-rendered reference surface: servers, deprecation, typed inputs,
// security, response headers, styling, and the locale guard.

import (
	"context"
	"strings"
	"testing"

	docs "github.com/DonaldMurillo/fastr-docs"
	uiapp "github.com/DonaldMurillo/gofastr/core-ui/app"
	"github.com/DonaldMurillo/gofastr/core/render"
)

const minimalSpec = `{"openapi":"3.1.0","info":{"title":"T","description":"D"},"paths":{"/x":{"get":{"summary":"S","operationId":"x","responses":{"200":{"description":"ok"}}}}}}`

// summarySpec names an operation summary so option markup can be checked.
const summarySpec = `{"openapi":"3.1.0","info":{"title":"T"},"paths":{"/x":{"get":{"summary":"List things","responses":{"200":{"description":"ok"}}}}}}`

// mountSpec applies the plugin to a fresh router, mounts it, and renders
// its page, so a test can assert on both the router and the markup.
func mountSpec(t *testing.T, p Plugin) (*docs.Router, string) {
	t.Helper()
	r := docs.NewRouter()
	if err := r.Use(p); err != nil {
		t.Fatal(err)
	}
	site := uiapp.NewApp("Docs")
	if err := r.Mount(site, r.Layout()); err != nil {
		t.Fatal(err)
	}
	path := p.Path
	if path == "" {
		path = "/api-reference"
	}
	html, err := site.RenderPage(context.Background(), path)
	if err != nil {
		t.Fatal(err)
	}
	return r, string(html)
}

func specPage(t *testing.T, spec, path string) string {
	_, html := mountSpec(t, Plugin{Spec: []byte(spec), Path: path})
	return html
}

func TestReferenceSurface(t *testing.T) {
	t.Run("a spec without servers implies same origin", func(t *testing.T) {
		html := specPage(t, minimalSpec, "/ref")
		if !strings.Contains(html, `data-openapi-server-url="/"`) {
			t.Fatal("no servers means the console is dead instead of same-origin")
		}
	})
	t.Run("every declared server is offered", func(t *testing.T) {
		spec := `{"openapi":"3.1.0","info":{"title":"T"},"servers":[{"url":"https://a.test"},{"url":"https://b.test"}],"paths":{"/x":{"get":{"responses":{"200":{"description":"ok"}}}}}}`
		html := specPage(t, spec, "/ref")
		if !strings.Contains(html, "https://a.test") || !strings.Contains(html, "https://b.test") {
			t.Fatal("only the first server is reachable; the second is silent")
		}
	})
	t.Run("deprecated operations are marked", func(t *testing.T) {
		spec := `{"openapi":"3.1.0","info":{"title":"T"},"paths":{"/x":{"get":{"deprecated":true,"summary":"Old","responses":{"200":{"description":"ok"}}}}}}`
		html := specPage(t, spec, "/ref")
		if !strings.Contains(html, "deprecated") {
			t.Fatal("a deprecated operation renders indistinguishably from a current one")
		}
	})
	t.Run("enum parameters render a select", func(t *testing.T) {
		spec := `{"openapi":"3.1.0","info":{"title":"T"},"paths":{"/x":{"get":{"parameters":[{"name":"size","in":"query","schema":{"type":"string","enum":["small","large"]}}],"responses":{"200":{"description":"ok"}}}}}}`
		html := specPage(t, spec, "/ref")
		if !strings.Contains(html, "small") || !strings.Contains(html, "<select") {
			t.Fatal("an enum parameter renders as free text")
		}
	})
	t.Run("date parameters render a date input", func(t *testing.T) {
		spec := `{"openapi":"3.1.0","info":{"title":"T"},"paths":{"/x":{"get":{"parameters":[{"name":"since","in":"query","schema":{"type":"string","format":"date"}}],"responses":{"200":{"description":"ok"}}}}}}`
		html := specPage(t, spec, "/ref")
		if !strings.Contains(html, `type="date"`) {
			t.Fatal("a date parameter renders as plain text")
		}
	})
	t.Run("a bearer scheme renders a token field", func(t *testing.T) {
		spec := `{"openapi":"3.1.0","info":{"title":"T"},"security":[{"bearer":[]}],"components":{"securitySchemes":{"bearer":{"type":"http","scheme":"bearer"}}},"paths":{"/x":{"get":{"responses":{"200":{"description":"ok"}}}}}}`
		html := specPage(t, spec, "/ref")
		if !strings.Contains(html, "bearer") || !strings.Contains(html, "token") {
			t.Fatal("security requirements are invisible; no credentials can be sent")
		}
	})
	t.Run("an apiKey scheme renders its header field", func(t *testing.T) {
		spec := `{"openapi":"3.1.0","info":{"title":"T"},"security":[{"key":[]}],"components":{"securitySchemes":{"key":{"type":"apiKey","in":"header","name":"X-Api-Key"}}},"paths":{"/x":{"get":{"responses":{"200":{"description":"ok"}}}}}}`
		html := specPage(t, spec, "/ref")
		if !strings.Contains(html, "X-Api-Key") {
			t.Fatal("an apiKey requirement is invisible to the console")
		}
	})
	t.Run("documented response headers render", func(t *testing.T) {
		spec := `{"openapi":"3.1.0","info":{"title":"T"},"paths":{"/x":{"get":{"responses":{"200":{"description":"ok","headers":{"X-Rate-Limit":{"description":"calls left"}}}}}}}}`
		html := specPage(t, spec, "/ref")
		if !strings.Contains(html, "X-Rate-Limit") {
			t.Fatal("documented response headers never appear")
		}
	})
	t.Run("a bad locale shape is refused at mount", func(t *testing.T) {
		r := docs.NewRouter()
		if err := r.Use(Plugin{Spec: []byte(minimalSpec), Path: "/ref", Locale: "e"}); err == nil {
			t.Fatal("a one-letter locale mounts; the language pairing forks silently")
		}
	})
	t.Run("the reference renders context aware", func(t *testing.T) {
		html := specPage(t, minimalSpec, "/api")
		if !strings.Contains(html, "data-openapi-reference") {
			t.Fatal("reference did not render")
		}
		var ref *Reference
		var ctxAware interface {
			RenderCtx(context.Context) render.HTML
		}
		_ = ref
		if _, ok := any(&Reference{}).(interface {
			RenderCtx(context.Context) render.HTML
		}); !ok {
			_ = ctxAware
			t.Fatal("Reference has no RenderCtx path for request-aware chrome")
		}
	})
	t.Run("operation options carry option semantics", func(t *testing.T) {
		html := specPage(t, minimalSpec, "/api")
		if !strings.Contains(html, `role="option"`) {
			t.Fatal("console options are bare text")
		}
	})
	t.Run("console options include their summary", func(t *testing.T) {
		html := specPage(t, summarySpec, "/ref")
		if !strings.Contains(html, "GET · /x — List things") && !strings.Contains(html, "List things</option>") {
			t.Fatal("two operations with the same method and path are indistinguishable in the select")
		}
	})
}

func TestReferenceMountSemantics(t *testing.T) {
	t.Run("a mixed-case locale is normalized", func(t *testing.T) {
		r, _ := mountSpec(t, Plugin{Spec: []byte(minimalSpec), Path: "/es/api", Locale: "ES"})
		if got := r.Routes()[0].Metadata.Locale; got != "es" {
			t.Fatalf("metadata locale = %q, want es", got)
		}
	})
	t.Run("the reference screen opts into preload", func(t *testing.T) {
		r, _ := mountSpec(t, Plugin{Spec: []byte(minimalSpec), Path: "/api"})
		if r.Routes()[0].Preload == "" {
			t.Fatal("plugin screen sets no Preload")
		}
	})
	t.Run("operation ids are unique per mount", func(t *testing.T) {
		_, html := mountSpec(t, Plugin{Spec: []byte(minimalSpec), Path: "/es/api", Locale: "es"})
		if !strings.Contains(html, `id="fastr-openapi-operation-es-api-1"`) {
			t.Fatal("operation ids carry no mount discriminator; two mounts collide")
		}
	})
	t.Run("duplicate ids name both operations", func(t *testing.T) {
		spec := `{"openapi":"3.1.0","info":{"title":"T"},"paths":{"/a":{"get":{"operationId":"dupe","responses":{"200":{"description":"ok"}}}},"/b":{"post":{"operationId":"dupe","responses":{"200":{"description":"ok"}}}}}}`
		r := docs.NewRouter()
		err := r.Use(Plugin{Spec: []byte(spec), Path: "/api"})
		if err == nil || !strings.Contains(err.Error(), "/a") || !strings.Contains(err.Error(), "/b") {
			t.Fatalf("error does not name both operations: %v", err)
		}
	})
	t.Run("duplicate operation ids are rejected", func(t *testing.T) {
		spec := `{"openapi":"3.1.0","info":{"title":"T"},"paths":{"/a":{"get":{"operationId":"dupe","responses":{"200":{"description":"ok"}}}},"/b":{"get":{"operationId":"dupe","responses":{"200":{"description":"ok"}}}}}}`
		r := docs.NewRouter()
		if err := r.Use(Plugin{Spec: []byte(spec), Path: "/api"}); err == nil {
			t.Fatal("Apply accepted two operations claiming one id")
		}
	})
	t.Run("paths carrying query strings are refused", func(t *testing.T) {
		spec := `{"openapi":"3.1.0","info":{"title":"T"},"paths":{"/x?y":{"get":{"responses":{"200":{"description":"ok"}}}}}}`
		r := docs.NewRouter()
		if err := r.Use(Plugin{Spec: []byte(spec), Path: "/ref"}); err == nil {
			t.Fatal("a query string inside a path key mounts a route that cannot be served")
		}
	})
}

func TestReferenceStyling(t *testing.T) {
	t.Run("group headings are styled", func(t *testing.T) {
		if !strings.Contains(CSS(), "fastr-openapi-group") {
			t.Fatal("group sections render with default heading styles")
		}
	})
	t.Run("deprecated operations are styled", func(t *testing.T) {
		if !strings.Contains(CSS(), "deprecated") {
			t.Fatal("a retired operation looks identical to a current one")
		}
	})
	t.Run("date inputs are sized", func(t *testing.T) {
		if !strings.Contains(CSS(), "type=\"date\"") && !strings.Contains(CSS(), "[type=date]") {
			t.Fatal("the date picker overflows the console column")
		}
	})
	t.Run("the empty filter state is styled", func(t *testing.T) {
		if !strings.Contains(CSS(), "empty") {
			t.Fatal("a filtered-to-nothing console has no empty state styling")
		}
	})
}
