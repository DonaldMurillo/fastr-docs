package docs

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func exportedSite(t *testing.T) (*Router, string) {
	t.Helper()
	r := NewRouter(WithSiteName("Docs"), WithLocaleFallback("en"))
	r.MustPage("/", PageConfig{Title: "Docs", Description: "Home", Source: "# Docs\n"})
	r.MustPage("/guide", PageConfig{Title: "Guide", Description: "en", Source: "# Guide\n", Order: 1})
	r.MustPage("/es/guide", PageConfig{Title: "Guía", Description: "es", Source: "# Guia\n", Order: 2,
		Metadata: ContentMetadata{Locale: "es"}})

	dir := t.TempDir()
	// Stand in for what GoFastr's export writes: one host-wide language on
	// every page, which is the whole problem.
	for _, path := range []string{"", "guide", "es/guide"} {
		file := filepath.Join(dir, filepath.FromSlash(path), "index.html")
		if err := os.MkdirAll(filepath.Dir(file), 0o755); err != nil {
			t.Fatal(err)
		}
		page := `<!DOCTYPE html><html data-fui-static lang="en"><head><title>x</title></head><body><p>&lt;html lang="zz"&gt; in prose</p></body></html>`
		if err := os.WriteFile(file, []byte(page), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	return r, dir
}

func langOf(t *testing.T, dir, path string) string {
	t.Helper()
	body, err := os.ReadFile(filepath.Join(dir, filepath.FromSlash(path), "index.html"))
	if err != nil {
		t.Fatal(err)
	}
	match := htmlLangAttribute.FindSubmatch(body)
	if match == nil {
		t.Fatalf("%s has no document language", path)
	}
	return string(match[2])
}

// Pagefind decides which language index a page belongs to by reading
// <html lang>. GoFastr writes one host-wide value, so without this every page
// is indexed as the same language and the per-language chunking never happens.
func TestWriteExportLocalesStampsEachPage(t *testing.T) {
	r, dir := exportedSite(t)
	if err := r.WriteExportLocales(dir); err != nil {
		t.Fatalf("WriteExportLocales() error = %v", err)
	}
	if got := langOf(t, dir, "es/guide"); got != "es" {
		t.Fatalf("Spanish page lang = %q", got)
	}
	for _, path := range []string{"", "guide"} {
		if got := langOf(t, dir, path); got != "en" {
			t.Fatalf("%s lang = %q, want en", path, got)
		}
	}
}

// Only the opening tag is rewritten. An <html> string inside the prose is
// content, and a search-and-replace would corrupt it.
func TestWriteExportLocalesLeavesProseAlone(t *testing.T) {
	r, dir := exportedSite(t)
	if err := r.WriteExportLocales(dir); err != nil {
		t.Fatalf("WriteExportLocales() error = %v", err)
	}
	body, err := os.ReadFile(filepath.Join(dir, "es", "guide", "index.html"))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(body), `&lt;html lang="zz"&gt; in prose`) {
		t.Fatalf("prose was rewritten: %s", body)
	}
}

// A route with no exported page is normal: screens and plugin routes may not
// produce one, and a missing file must not fail an export.
func TestWriteExportLocalesSkipsMissingPages(t *testing.T) {
	r, dir := exportedSite(t)
	r.MustPage("/es/only-registered", PageConfig{Title: "Solo", Description: "es", Source: "# Solo\n", Order: 3,
		Metadata: ContentMetadata{Locale: "es"}})
	if err := r.WriteExportLocales(dir); err != nil {
		t.Fatalf("WriteExportLocales() error = %v", err)
	}
}
