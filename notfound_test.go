package docs

import (
	"strings"
	"testing"
)

func TestNotFoundScreenEscapesRequestedPathAndProvidesRecoveryLink(t *testing.T) {
	html := string(NotFoundScreen{SiteName: "Manual Docs"}.RenderNotFound(`/missing"><script>alert(1)</script>`))
	for _, marker := range []string{"404", "Page not found", "Back to Manual Docs", "&lt;script&gt;"} {
		if !strings.Contains(html, marker) {
			t.Fatalf("NotFoundScreen missing %q: %s", marker, html)
		}
	}
	if strings.Contains(html, "<script>alert") {
		t.Fatalf("NotFoundScreen reflected unsafe markup: %s", html)
	}
}

func TestTheNotFoundScreenOffersSearchAndLocaleSiblings(t *testing.T) {
	t.Run("the 404 offers a way back to search", func(t *testing.T) {
		r := NewRouter()
		r.MustPage("/g", PageConfig{Title: "G", Description: "d", Order: 1, Source: "# G"})
		html := string(r.NotFoundScreen().RenderNotFound("/deep/miss"))
		if !strings.Contains(html, "search") {
			t.Fatal("the 404 dead-ends with no route back to search")
		}
	})
	t.Run("the 404 carries hreflang for its locales", func(t *testing.T) {
		r := bilingualDocsSite(t)
		html := string(r.NotFoundScreen().Render())
		if !strings.Contains(html, "hreflang") {
			t.Fatal("the localized 404 never tells search engines about its siblings")
		}
	})
	t.Run("the 404 answers in the site's language", func(t *testing.T) {
		r := NewRouter(WithLanguage("es"),
			WithLocaleUIStrings("es", UIStrings{NotFound: NotFoundStrings{Heading: "Página no encontrada", SiteFallback: "la documentación"}}))
		r.MustPage("/g", PageConfig{Title: "G", Description: "d", Order: 1, Source: "# G"})
		screen := r.NotFoundScreen()
		if screen.Strings.Heading != "Página no encontrada" {
			t.Fatalf("a Spanish-language site ships an English 404: %+v", screen.Strings)
		}
	})
}
