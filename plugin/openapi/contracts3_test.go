package openapi

// Reference surface contracts from the third audit cycle: styling for
// the new surfaces, apiKey security, response headers, option summaries,
// and spec parsing edges.

import (
	"context"
	"os"
	"strings"
	"testing"

	docs "github.com/DonaldMurillo/fastr-docs"
	uiapp "github.com/DonaldMurillo/gofastr/core-ui/app"
)

func r3Page(t *testing.T, spec, path string) string {
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

func r3JS(t *testing.T) string {
	t.Helper()
	body, err := os.ReadFile("openapi.js")
	if err != nil {
		t.Fatal(err)
	}
	return string(body)
}

const r3Spec = `{"openapi":"3.1.0","info":{"title":"T"},"paths":{"/x":{"get":{"summary":"List things","responses":{"200":{"description":"ok"}}}}}}`

func TestRed289To297(t *testing.T) {
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
	t.Run("an apiKey scheme renders its header field", func(t *testing.T) {
		spec := `{"openapi":"3.1.0","info":{"title":"T"},"security":[{"key":[]}],"components":{"securitySchemes":{"key":{"type":"apiKey","in":"header","name":"X-Api-Key"}}},"paths":{"/x":{"get":{"responses":{"200":{"description":"ok"}}}}}}`
		html := r3Page(t, spec, "/ref")
		if !strings.Contains(html, "X-Api-Key") {
			t.Fatal("an apiKey requirement is invisible to the console")
		}
	})
	t.Run("received response headers are listed", func(t *testing.T) {
		js := r3JS(t)
		if !strings.Contains(js, "result.headers") && !strings.Contains(js, "response.headers") {
			t.Fatal("the pane shows the body only; headers the server sent are dropped")
		}
	})
	t.Run("console options include their summary", func(t *testing.T) {
		html := r3Page(t, r3Spec, "/ref")
		if !strings.Contains(html, "GET · /x — List things") && !strings.Contains(html, "List things</option>") {
			t.Fatal("two operations with the same method and path are indistinguishable in the select")
		}
	})
	t.Run("paths carrying query strings are refused", func(t *testing.T) {
		spec := `{"openapi":"3.1.0","info":{"title":"T"},"paths":{"/x?y":{"get":{"responses":{"200":{"description":"ok"}}}}}}`
		r := docs.NewRouter()
		if err := r.Use(Plugin{Spec: []byte(spec), Path: "/ref"}); err == nil {
			t.Fatal("a query string inside a path key mounts a route that cannot be served")
		}
	})
	t.Run("the console exposes its busy state in markup", func(t *testing.T) {
		html := r3Page(t, r3Spec, "/ref")
		if !strings.Contains(html, "data-openapi-busy") && !strings.Contains(html, `aria-busy="false"`) {
			t.Fatal("busy state lives only in JavaScript")
		}
	})
}
