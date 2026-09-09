package docs

import (
	"context"
	"strings"
	"testing"

	fastrdocs "github.com/DonaldMurillo/fastr-docs"
	uiapp "github.com/DonaldMurillo/gofastr/core-ui/app"
)

func renderSitePage(t *testing.T, path string) string {
	t.Helper()
	router := NewRouter()
	site := uiapp.NewApp("fastr-docs")
	if err := router.Mount(site, router.Layout()); err != nil {
		t.Fatal(err)
	}
	html, err := site.RenderPage(context.Background(), path)
	if err != nil {
		t.Fatal(err)
	}
	return string(html)
}

func TestSpanishReferenceCanSendRequests(t *testing.T) {
	t.Setenv("API_SERVER_URL", "https://api.example.com/v1")
	html := renderSitePage(t, "/es/api-reference")
	if !strings.Contains(html, "data-openapi-server-url") {
		t.Fatal("the Spanish mount has no server; its console is decorative")
	}
}

func TestConsoleSelectsHaveFocusStyles(t *testing.T) {
	if !strings.Contains(OpenAPICSS(), ":focus") {
		t.Fatal("the reference console is invisible to keyboard users")
	}
}

// The reference console loads its interactivity from the framework's JS
// runtime; this keeps that runtime linked into the site.
var _ = fastrdocs.RuntimeJS
