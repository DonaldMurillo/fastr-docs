package docs

// Shared builders for the coverage suites: a rendered page, a virtual
// blog, and a minimally paired bilingual site.

import (
	"context"
	"testing"
	"testing/fstest"

	uiapp "github.com/DonaldMurillo/gofastr/core-ui/app"
)

func r2Render(t *testing.T, r *Router, path string) string {
	t.Helper()
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

func r2Blog(t *testing.T, files map[string]string) *Router {
	t.Helper()
	mapFS := fstest.MapFS{}
	for name, body := range files {
		mapFS[name] = &fstest.MapFile{Data: []byte(body)}
	}
	r := NewRouter()
	if err := r.MarkdownBlogFS("/blog", mapFS, ".", BlogConfig{}); err != nil {
		t.Fatal(err)
	}
	return r
}

func r2Bilingual() *Router {
	r := NewRouter(WithLocaleFallback("en"))
	r.MustPage("/docs/guide", PageConfig{Title: "Guide", Description: "d", Source: "# Guide", Order: 1})
	r.MustPage("/es/docs/guide", PageConfig{Title: "Guía", Description: "d", Source: "# Guía", Order: 2,
		Metadata: ContentMetadata{Locale: "es"}})
	return r
}
