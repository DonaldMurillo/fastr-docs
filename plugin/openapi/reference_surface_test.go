package openapi

// Server-rendered reference contracts: servers, deprecation, typed inputs,
// security, response headers, and the locale guard.

import (
	"context"
	"strings"
	"testing"

	docs "github.com/DonaldMurillo/fastr-docs"
	uiapp "github.com/DonaldMurillo/gofastr/core-ui/app"
)

func specPage(t *testing.T, spec, path string) string {
	t.Helper()
	r := docs.NewRouter()
	if err := r.Use(Plugin{Spec: []byte(spec), Path: path}); err != nil {
		t.Fatal(err)
	}
	site := uiapp.NewApp("T")
	if err := r.Mount(site, r.Layout()); err != nil {
		t.Fatal(err)
	}
	html, err := site.RenderPage(context.Background(), path)
	if err != nil {
		t.Fatal(err)
	}
	return string(html)
}

func TestReferenceSurface(t *testing.T) {
	t.Run("a spec without servers implies same origin", func(t *testing.T) {
		html := specPage(t, redSpec, "/ref")
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
	t.Run("documented response headers render", func(t *testing.T) {
		spec := `{"openapi":"3.1.0","info":{"title":"T"},"paths":{"/x":{"get":{"responses":{"200":{"description":"ok","headers":{"X-Rate-Limit":{"description":"calls left"}}}}}}}}`
		html := specPage(t, spec, "/ref")
		if !strings.Contains(html, "X-Rate-Limit") {
			t.Fatal("documented response headers never appear")
		}
	})
	t.Run("a bad locale shape is refused at mount", func(t *testing.T) {
		r := docs.NewRouter()
		if err := r.Use(Plugin{Spec: []byte(redSpec), Path: "/ref", Locale: "e"}); err == nil {
			t.Fatal("a one-letter locale mounts; the language pairing forks silently")
		}
	})
}
