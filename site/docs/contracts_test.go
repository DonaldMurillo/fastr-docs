package docs

// Site-level contracts from the red-suite audit: each case failed before
// the fix and passes now.

import (
	"context"
	"strings"
	"testing"

	fastrdocs "github.com/DonaldMurillo/fastr-docs"
	uiapp "github.com/DonaldMurillo/gofastr/core-ui/app"
)

func redSiteRender(t *testing.T, path string) string {
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

func TestRed098To100Site(t *testing.T) {
	t.Run("098 the Spanish reference can send requests", func(t *testing.T) {
		t.Setenv("API_SERVER_URL", "https://api.example.com/v1")
		html := redSiteRender(t, "/es/api-reference")
		if !strings.Contains(html, "data-openapi-server-url") {
			t.Fatal("the Spanish mount has no server; its console is decorative")
		}
	})
	t.Run("099 the console selects have focus styles", func(t *testing.T) {
		if !strings.Contains(OpenAPICSS(), ":focus") {
			t.Fatal("the reference console is invisible to keyboard users")
		}
	})
	t.Run("100 skip links are localized", func(t *testing.T) {
		html := redSiteRender(t, "/es/docs/getting-started")
		if !strings.Contains(html, "Saltar") {
			t.Fatal("the skip link stays English on Spanish pages")
		}
	})
}

var _ = fastrdocs.RuntimeJS
