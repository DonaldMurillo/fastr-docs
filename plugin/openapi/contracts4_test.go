package openapi

// Fourth-cycle console contracts: searchable parameter names and the api
// key in the exported command.

// Fourth red suite: console and search contracts.

import (
	"context"
	"os"
	"strings"
	"testing"

	docs "github.com/DonaldMurillo/fastr-docs"
	uiapp "github.com/DonaldMurillo/gofastr/core-ui/app"
)

const r4Spec = `{"openapi":"3.1.0","info":{"title":"T"},"paths":{"/x":{"get":{"summary":"List","operationId":"listX","parameters":[{"name":"limit","in":"query","schema":{"type":"integer"}}],"responses":{"200":{"description":"ok"}}}}}}`

func r4Page(t *testing.T, spec, path string) string {
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

func TestRed344To349OpenAPI(t *testing.T) {
	t.Run("344 search matches parameter names", func(t *testing.T) {
		html := r4Page(t, r4Spec, "/ref")
		if !strings.Contains(html, "limit") {
			t.Fatal("parameter names never render at all")
		}
		index := strings.LastIndex(html, "data-openapi-search=")
		if index < 0 || !strings.Contains(html[index:index+220], "limit") {
			t.Fatal(`filtering by "limit" finds nothing; parameter names are not searchable`)
		}
	})
	t.Run("346 the apiKey feeds the curl command", func(t *testing.T) {
		body, err := os.ReadFile("openapi.js")
		if err != nil {
			t.Fatal(err)
		}
		start := strings.Index(string(body), "data-openapi-curl")
		if start < 0 {
			t.Skip("no curl button")
		}
		if !strings.Contains(string(body)[start:], "api-key") {
			t.Fatal("the curl command drops the api key")
		}
	})
}
