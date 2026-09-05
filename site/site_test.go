package main

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestSelfHostedSiteRoutes(t *testing.T) {
	built, err := buildSite()
	if err != nil {
		t.Fatalf("build site: %v", err)
	}
	checks := []struct {
		path   string
		marker string
	}{
		{path: "/", marker: "Build docs around one route tree."},
		{path: "/docs/getting-started", marker: "Install fastr-docs and create your first route tree."},
		{path: "/docs/concepts/content", marker: "MarkdownCollectionFS"},
		{path: "/docs/concepts/layouts", marker: "Layouts and navigation"},
		{path: "/docs/build/framework-ui", marker: "Native component catalog"},
		{path: "/docs/build/blog", marker: "MarkdownBlog"},
		{path: "/blog", marker: "One tree for docs and publishing"},
		{path: "/blog/archive", marker: "Browse every published post by year."},
		{path: "/blog/tags", marker: "architecture"},
		{path: "/blog/authors", marker: "fastr-docs"},
		{path: "/blog/search?q=router", marker: "Results for"},
		{path: "/blog/route-tree", marker: "One tree for docs and publishing"},
		{path: "/blog/feed.xml", marker: "<rss"},
		{path: "/examples/playground", marker: `data-testid="examples-playground"`},
		{path: "/api-reference", marker: "data-openapi-reference"},
		{path: "/__fastr-docs/docs.js", marker: "initTocSelect"},
		{path: "/__fastr-docs/openapi.js", marker: "data-openapi-filter"},
		{path: "/__fastr-docs/search.json", marker: "getting-started"},
		{path: "/__fastr-docs/manifest.json", marker: "fastr-docs/v1"},
		{path: "/llms.txt", marker: "## When to use"},
		{path: "/.well-known/agent-card.json", marker: "fastr-docs"},
		{path: "/manifest.webmanifest", marker: "fastr-docs"},
		{path: "/service-worker.js", marker: "__fastr-docs/search.json"},
		{path: "/__gofastr/pwa/register.js", marker: "gofastr:pwa-update"},
		{path: "/__gofastr/pwa/offline", marker: "offline"},
		{path: "/sitemap.xml", marker: "/docs/getting-started"},
		{path: "/robots.txt", marker: "Disallow: /__fastr-docs/"},
		{path: "/assets/favicon.svg", marker: "<svg"},
	}

	for _, check := range checks {
		request := httptest.NewRequest(http.MethodGet, check.path, nil)
		response := httptest.NewRecorder()
		built.server.Router().ServeHTTP(response, request)
		if response.Code != http.StatusOK {
			t.Fatalf("GET %s status = %d, want 200; body: %s", check.path, response.Code, response.Body.String())
		}
		if !strings.Contains(response.Body.String(), check.marker) {
			t.Fatalf("GET %s missing %q", check.path, check.marker)
		}
	}
}

func TestListenAddressAcceptsGoFastrHostPort(t *testing.T) {
	if got := listenAddress("localhost:8080", ":3079"); got != "localhost:8080" {
		t.Fatalf("listenAddress(host:port) = %q, want localhost:8080", got)
	}
	if got := listenAddress("3079", ":3079"); got != ":3079" {
		t.Fatalf("listenAddress(port) = %q, want :3079", got)
	}
	if got := listenURL("localhost:8080"); got != "http://localhost:8080" {
		t.Fatalf("listenURL(host:port) = %q, want http://localhost:8080", got)
	}
}
