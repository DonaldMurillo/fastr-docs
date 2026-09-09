package openapi

import (
	"strings"
	"testing"

	docs "github.com/DonaldMurillo/fastr-docs"
)

func TestPluginRegistersRichReferenceScreen(t *testing.T) {
	spec := []byte(`{"openapi":"3.1.0","info":{"title":"Public API","description":"The spec."},"servers":[{"url":"https://api.example.com/{version}","variables":{"version":{"default":"v1"}}}],"paths":{"/projects":{"get":{"summary":"List projects","operationId":"listProjects","responses":{"200":{"description":"ok"}}}}},"components":{"schemas":{"Project":{"type":"object","description":"A project","properties":{"id":{"type":"string"}}}}}}`)
	router, html := mountSpec(t, Plugin{
		Spec:  spec,
		Order: 1,
		Badge: docs.NavBadge{Label: "Demo", Tone: docs.NavBadgeToneNeutral},
	})
	if got := router.ConnectOrigins(); len(got) != 1 || got[0] != "https://api.example.com" {
		t.Fatalf("OpenAPI connect origins = %#v, want spec server origin", got)
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
	for _, marker := range []string{"data-openapi-reference", "data-openapi-server-url=\"https://api.example.com/v1\"", "data-fui-scrollspy", "ui-anchored-rail", "data-openapi-try", "data-openapi-inputs-for=\"fastr-openapi-operation-api-reference-1\"", "#fastr-openapi-operation-api-reference-1", "GET", "/projects", "Project"} {
		if !strings.Contains(html, marker) {
			t.Fatalf("rendered reference missing %q: %s", marker, html)
		}
	}
}

func TestPluginServerURLOverrideWins(t *testing.T) {
	spec := []byte(`{"openapi":"3.1.0","servers":[{"url":"https://api.example.com"}],"paths":{}}`)
	_, html := mountSpec(t, Plugin{Spec: spec, ServerURL: "https://staging.example.com/api/"})
	if !strings.Contains(html, `data-openapi-server-url="https://staging.example.com/api"`) {
		t.Fatalf("server URL override missing: %s", html)
	}
}

func TestSwaggerServerURLFallback(t *testing.T) {
	spec := []byte(`{"swagger":"2.0","host":"api.example.com","basePath":"/v1","schemes":["https"],"paths":{}}`)
	_, html := mountSpec(t, Plugin{Spec: spec})
	if !strings.Contains(html, `data-openapi-server-url="https://api.example.com/v1"`) {
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
	_, html := mountSpec(t, Plugin{Spec: spec, Order: 1})
	if !strings.Contains(html, "listProjects") || !strings.Contains(html, "data-openapi-param-name=\"trace\"") {
		t.Fatalf("rendered operations omitted path-level parameter: %s", html)
	}
}

func TestPluginAcceptsYAMLSpec(t *testing.T) {
	spec := []byte(`openapi: 3.1.0
info:
  title: YAML API
  description: A YAML spec.
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

// A translated mount declares its language so the route pairs with the
// original section, and carries its own surface labels: the plugin's chrome
// is otherwise English whatever the spec says.
func TestTranslatedMountCarriesLocaleAndStrings(t *testing.T) {
	spec := []byte(`{"openapi":"3.1.0","info":{"title":"API de contenido de ejemplo","description":"Una especificación pequeña."},"paths":{"/v1/projects":{"get":{"summary":"Listar proyectos","operationId":"listProjects","responses":{"200":{"description":"ok"}}}}}}`)
	router, html := mountSpec(t, Plugin{
		Spec:        spec,
		Path:        "/es/api-reference",
		Title:       "Referencia de la API de ejemplo",
		Description: "Una especificación pequeña.",
		Locale:      "es",
		// A label at a time: the fields left empty keep the English
		// defaults, the way a partly translated site behaves.
		Strings: Strings{
			Eyebrow:      "Referencia OpenAPI",
			TryRequest:   "Prueba una petición",
			SendRequest:  "Enviar la petición",
			NoOperations: "Esta especificación no tiene operaciones.",
		},
	})
	routes := router.Routes()
	if len(routes) != 1 || routes[0].Metadata.Locale != "es" {
		t.Fatalf("translated route metadata = %#v, want locale es", routes)
	}
	if got := router.LanguageFor("/es/api-reference"); got != "es" {
		t.Fatalf("LanguageFor = %q, want es", got)
	}
	for _, want := range []string{"Referencia OpenAPI", "Prueba una petición", "Enviar la petición"} {
		if !strings.Contains(html, want) {
			t.Fatalf("translated reference missing %q", want)
		}
	}
	// An empty field keeps the default rather than rendering blank.
	if !strings.Contains(html, "Filter endpoints…") {
		t.Fatal("untranslated field lost the English default")
	}
}
