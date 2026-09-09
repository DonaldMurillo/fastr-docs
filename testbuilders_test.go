package docs

// Shared builders: a rendered page, a small docs site, virtual blogs, and
// two bilingual sites, one minimal and one full.

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"testing/fstest"

	uiapp "github.com/DonaldMurillo/gofastr/core-ui/app"
	gofastrRouter "github.com/DonaldMurillo/gofastr/core/router"
)

func renderRouterPage(t *testing.T, r *Router, path string) string {
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

// renderScratchPage renders one ungrouped page whose body is the test's
// whole subject; nothing about the surrounding chrome matters to it.
func renderScratchPage(t *testing.T, source string) string {
	t.Helper()
	r := NewRouter()
	r.MustPage("/g", PageConfig{Title: "G", Description: "d", Order: 1, Source: source})
	return renderRouterPage(t, r, "/g")
}

func docsSiteRouter(t *testing.T) *Router {
	t.Helper()
	r := NewRouter()
	r.MustPage("/", PageConfig{Title: "Home", Description: "Home", Source: "# Home", Order: 1})
	docs := r.MustGroup("/docs", GroupConfig{Title: "Docs", Description: "Guides", Order: 2})
	docs.MustPage("start", PageConfig{Title: "Start", Description: "Start", Source: "# Start", Order: 1})
	r.MustPage("/api-reference", PageConfig{Title: "API reference", Description: "API", Source: "# API", Order: 3})
	return r
}

// writeFiles lays a slash-separated virtual tree under dir, creating parent
// directories, so disk-backed collections can read it.
func writeFiles(t *testing.T, dir string, files map[string]string) {
	t.Helper()
	for name, body := range files {
		path := filepath.Join(dir, filepath.FromSlash(name))
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(path, []byte(body), 0o644); err != nil {
			t.Fatal(err)
		}
	}
}

// blogRouter mounts a virtual blog at /blog; pass a config to exercise a
// collection setting other than the default.
func blogRouter(t *testing.T, files map[string]string, config ...BlogConfig) *Router {
	t.Helper()
	blog := BlogConfig{}
	if len(config) > 0 {
		blog = config[0]
	}
	mapFS := fstest.MapFS{}
	for name, body := range files {
		mapFS[name] = &fstest.MapFile{Data: []byte(body)}
	}
	r := NewRouter()
	if err := r.MarkdownBlogFS("/blog", mapFS, ".", blog); err != nil {
		t.Fatal(err)
	}
	return r
}

// exportManifest decodes the manifest the way every artifact test reads it.
func exportManifest(t *testing.T, r *Router, base string) ExportManifest {
	t.Helper()
	body, err := r.ExportManifestJSON(base)
	if err != nil {
		t.Fatal(err)
	}
	var manifest ExportManifest
	if err := json.Unmarshal(body, &manifest); err != nil {
		t.Fatal(err)
	}
	return manifest
}

// hasRoutePattern reports whether any mounted route pattern contains want.
func hasRoutePattern(httpRouter *gofastrRouter.Router, want string) bool {
	for _, route := range httpRouter.Routes() {
		if strings.Contains(route.Pattern, want) {
			return true
		}
	}
	return false
}

// blogAcrossTwoYears is the archive fixture: one post in 2025, one in 2026,
// and a tag only the newer post carries.
func blogAcrossTwoYears(t *testing.T) *Router {
	t.Helper()
	return blogRouter(t, map[string]string{
		"index.md": "---\ntitle: Notas\ndescription: d\n---\n\nWelcome.",
		"newer.md": "---\ntitle: Nuevo\ndescription: d\ndate: 2026-05-01\ntags: [go]\n---\n\nBody.",
		"older.md": "---\ntitle: Viejo\ndescription: d\ndate: 2025-05-01\n---\n\nBody.",
	})
}

// bilingualPair is the minimal translation pair: two pages, one family.
func bilingualPair() *Router {
	r := NewRouter(WithLocaleFallback("en"))
	r.MustPage("/docs/guide", PageConfig{Title: "Guide", Description: "d", Source: "# Guide", Order: 1})
	r.MustPage("/es/docs/guide", PageConfig{Title: "Guía", Description: "d", Source: "# Guía", Order: 2,
		Metadata: ContentMetadata{Locale: "es"}})
	return r
}

// bilingualDocsSite is the full shape of a translated site: a home and a
// section per language, with the Spanish chrome labels set.
func bilingualDocsSite(t *testing.T) *Router {
	t.Helper()
	r := NewRouter(
		WithLocaleFallback("en"),
		WithLocaleUIStrings("es", UIStrings{Contents: "Contenido", Home: "Inicio", Sections: "Secciones"}),
	)
	r.MustPage("/", PageConfig{Title: "Home", Description: "Home", Source: "# Home", Order: 1, Metadata: ContentMetadata{Locale: "en"}})
	english := r.MustGroup("/docs", GroupConfig{Title: "Docs", Description: "en", Order: 2})
	english.MustPage("guide", PageConfig{Title: "Guide", Description: "en", Source: "# Guide", Order: 1, Metadata: ContentMetadata{Locale: "en"}})
	r.MustPage("/es", PageConfig{Title: "Español", Description: "es", Source: "# Inicio", Order: 4, Metadata: ContentMetadata{Locale: "es"}})
	spanish := r.MustGroup("/es/docs", GroupConfig{Title: "Documentación", Description: "es", Order: 3, Locale: "es"})
	spanish.MustPage("guide", PageConfig{Title: "Guía", Description: "es", Source: "# Guía", Order: 1, Metadata: ContentMetadata{Locale: "es"}})
	return r
}

func sectionSelectHTML(r *Router, path string) string {
	return string(r.docsSectionSelect(path, "x"))
}

func cssRule(css, selector string) (string, bool) {
	i := strings.Index(css, selector)
	if i < 0 {
		return "", false
	}
	j := strings.Index(css[i:], "}")
	if j < 0 {
		return css[i:], true
	}
	return css[i : i+j], true
}

func cssMediaBlock(css, open string) (string, bool) {
	start := strings.Index(css, open)
	if start < 0 {
		return "", false
	}
	block := css[start:]
	if end := strings.Index(block[1:], "@media"); end >= 0 {
		block = block[:end+1]
	}
	return block, true
}
