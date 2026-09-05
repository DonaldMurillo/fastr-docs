package openapi

import (
	"context"
	"strings"
	"testing"

	docs "github.com/DonaldMurillo/fastr-docs"
	uiapp "github.com/DonaldMurillo/gofastr/core-ui/app"
)

func TestPluginRegistersRichReferenceScreen(t *testing.T) {
	spec := []byte(`{"openapi":"3.1.0","info":{"title":"Public API","description":"The contract."},"servers":[{"url":"https://api.example.com/{version}","variables":{"version":{"default":"v1"}}}],"paths":{"/projects":{"get":{"summary":"List projects","operationId":"listProjects","responses":{"200":{"description":"ok"}}}}},"components":{"schemas":{"Project":{"type":"object","description":"A project","properties":{"id":{"type":"string"}}}}}}`)
	router := docs.NewRouter()
	if err := router.Use(Plugin{
		Spec:  spec,
		Order: 1,
		Badge: docs.NavBadge{Label: "Demo", Tone: docs.NavBadgeToneNeutral},
	}); err != nil {
		t.Fatalf("Use() error = %v", err)
	}
	if got := router.ConnectOrigins(); len(got) != 1 || got[0] != "https://api.example.com" {
		t.Fatalf("OpenAPI connect origins = %#v, want contract server origin", got)
	}
	if err := router.Validate(); err != nil {
		t.Fatalf("Validate() error = %v", err)
	}
	if routes := router.Routes(); len(routes) != 1 || routes[0].Plugin != "openapi" {
		t.Fatalf("plugin provenance = %#v", router.Routes())
	}
	if route := router.Routes()[0]; route.Badge.Label != "Demo" || route.Badge.Tone != docs.NavBadgeToneNeutral {
		t.Fatalf("plugin badge = %#v, want Demo/neutral", route.Badge)
	}
	entries := router.SearchIndex()
	if len(entries) != 1 || !strings.Contains(entries[0].Text, "listProjects") {
		t.Fatalf("search entry = %#v", entries)
	}
	site := uiapp.NewApp("Docs")
	if err := router.Mount(site, router.Layout()); err != nil {
		t.Fatalf("Mount() error = %v", err)
	}
	html, err := site.RenderPage(context.Background(), "/api-reference")
	if err != nil {
		t.Fatalf("RenderPage() error = %v", err)
	}
	for _, marker := range []string{"data-openapi-reference", "data-openapi-server-url=\"https://api.example.com/v1\"", "data-fui-scrollspy", "ui-anchored-rail", "data-openapi-try", "data-openapi-inputs-for=\"fastr-openapi-operation-1\"", "#fastr-openapi-operation-1", "GET", "/projects", "Project"} {
		if !strings.Contains(string(html), marker) {
			t.Fatalf("rendered reference missing %q: %s", marker, html)
		}
	}
}

func TestPluginServerURLOverrideWins(t *testing.T) {
	spec := []byte(`{"openapi":"3.1.0","servers":[{"url":"https://api.example.com"}],"paths":{}}`)
	router := docs.NewRouter()
	if err := router.Use(Plugin{Spec: spec, ServerURL: "https://staging.example.com/api/"}); err != nil {
		t.Fatalf("Use() error = %v", err)
	}
	site := uiapp.NewApp("Docs")
	if err := router.Mount(site, router.Layout()); err != nil {
		t.Fatalf("Mount() error = %v", err)
	}
	html, err := site.RenderPage(context.Background(), "/api-reference")
	if err != nil {
		t.Fatalf("RenderPage() error = %v", err)
	}
	if !strings.Contains(string(html), `data-openapi-server-url="https://staging.example.com/api"`) {
		t.Fatalf("server URL override missing: %s", html)
	}
}

func TestSwaggerServerURLFallback(t *testing.T) {
	spec := []byte(`{"swagger":"2.0","host":"api.example.com","basePath":"/v1","schemes":["https"],"paths":{}}`)
	router := docs.NewRouter()
	if err := router.Use(Plugin{Spec: spec}); err != nil {
		t.Fatalf("Use() error = %v", err)
	}
	site := uiapp.NewApp("Docs")
	if err := router.Mount(site, router.Layout()); err != nil {
		t.Fatalf("Mount() error = %v", err)
	}
	html, err := site.RenderPage(context.Background(), "/api-reference")
	if err != nil {
		t.Fatalf("RenderPage() error = %v", err)
	}
	if !strings.Contains(string(html), `data-openapi-server-url="https://api.example.com/v1"`) {
		t.Fatalf("Swagger server URL fallback missing: %s", html)
	}
}

func TestPluginRejectsInvalidSpec(t *testing.T) {
	router := docs.NewRouter()
	if err := router.Use(Plugin{Spec: []byte(`{"info":{}}`), Order: 1}); err == nil {
		t.Fatal("invalid spec accepted")
	}
}

func TestPluginMergesPathItemParameters(t *testing.T) {
	spec := []byte(`{"openapi":"3.1.0","info":{"title":"API"},"paths":{"/v1/projects":{"parameters":[{"name":"trace","in":"header"}],"get":{"operationId":"listProjects"}}}}`)
	router := docs.NewRouter()
	if err := router.Use(Plugin{Spec: spec, Order: 1}); err != nil {
		t.Fatalf("Use() error = %v", err)
	}
	site := uiapp.NewApp("Docs")
	if err := router.Mount(site, router.Layout()); err != nil {
		t.Fatalf("Mount() error = %v", err)
	}
	html, err := site.RenderPage(context.Background(), "/api-reference")
	if err != nil {
		t.Fatalf("RenderPage() error = %v", err)
	}
	if !strings.Contains(string(html), "listProjects") || !strings.Contains(string(html), "data-openapi-param-name=\"trace\"") {
		t.Fatalf("rendered operations omitted path-level parameter: %s", html)
	}
}

func TestPluginAcceptsYAMLSpec(t *testing.T) {
	spec := []byte(`openapi: 3.1.0
info:
  title: YAML API
  description: A YAML contract.
servers:
  - url: https://api.example.com/v1
paths:
  /projects:
    get:
      operationId: listProjects
      responses:
        '200':
          description: ok
`)
	router := docs.NewRouter()
	if err := router.Use(Plugin{Spec: spec, Order: 1}); err != nil {
		t.Fatalf("Use(YAML) error = %v", err)
	}
	if err := router.Validate(); err != nil {
		t.Fatalf("Validate() error = %v", err)
	}
	routes := router.Routes()
	if len(routes) != 1 || routes[0].Title != "YAML API" {
		t.Fatalf("YAML route = %#v", routes)
	}
}
