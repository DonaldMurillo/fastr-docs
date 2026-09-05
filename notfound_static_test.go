package docs

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestWriteStaticNotFound(t *testing.T) {
	dir := t.TempDir()
	if err := WriteStaticNotFound(dir, "/handbook", NotFoundScreen{SiteName: "Manual Docs"}, ""); err != nil {
		t.Fatal(err)
	}
	body, err := os.ReadFile(filepath.Join(dir, "404.html"))
	if err != nil {
		t.Fatal(err)
	}
	for _, marker := range []string{"404: Manual Docs", "Page not found", `href="/handbook/"`, `href="/handbook/404.css"`, "fastr-docs-not-found"} {
		if !strings.Contains(string(body), marker) {
			t.Fatalf("static 404 missing %q", marker)
		}
	}
	if strings.Contains(string(body), "<style>") {
		t.Fatal("static 404 inlined CSS instead of using the exported stylesheet")
	}
	css, err := os.ReadFile(filepath.Join(dir, "404.css"))
	if err != nil || !strings.Contains(string(css), "fastr-docs-not-found") {
		t.Fatalf("static 404 stylesheet missing: %v", err)
	}
}

func TestRewriteStaticCSPMatchesMetadataSemantically(t *testing.T) {
	dir := t.TempDir()
	old := `<html><head><meta content="default-src 'self'" name="ignored"><meta content="old" http-equiv="Content-Security-Policy"></head></html>`
	if err := os.WriteFile(filepath.Join(dir, "index.html"), []byte(old), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "_headers"), []byte("/*\n  Content-Security-Policy: old\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := RewriteStaticCSP(dir, "default-src 'self'; connect-src https://api.example"); err != nil {
		t.Fatal(err)
	}
	body, err := os.ReadFile(filepath.Join(dir, "index.html"))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(body), "connect-src https://api.example") || strings.Contains(string(body), `content="old"`) {
		t.Fatalf("CSP metadata was not rewritten: %s", body)
	}
	header, err := os.ReadFile(filepath.Join(dir, "_headers"))
	if err != nil || !strings.Contains(string(header), "connect-src https://api.example") {
		t.Fatalf("CSP header was not rewritten: %v %s", err, header)
	}
}

func TestRewriteStaticCSPFailsWhenNoPolicyExists(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "index.html"), []byte("<html></html>"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := RewriteStaticCSP(dir, "default-src 'self'"); err == nil {
		t.Fatal("RewriteStaticCSP silently accepted an export without a CSP")
	}
}
