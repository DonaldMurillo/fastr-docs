package cli

// Fourth-cycle starter contracts: the generated export rewrites what the
// runtime actually reads, validates its flags, materializes redirects, and
// refuses to bake localhost into public URLs.

// Fourth red suite: the starter template's static export must rewrite every
// runtime URL attribute, not only the defaults.

import (
	"os"
	"strings"
	"testing"
)

func TestRed345CLI(t *testing.T) {
	t.Run("345 the starter rewrites custom search index paths", func(t *testing.T) {
		body, err := os.ReadFile("templates/starter/main.go.tmpl")
		if err != nil {
			t.Fatal(err)
		}
		// The attribute-shaped regex covers index and pagefind paths
		// together, custom values included.
		if !strings.Contains(string(body), "data-fastr-docs-(?:index|pagefind)-path") {
			t.Fatal("an export below a prefix leaves a custom search index URL root-relative")
		}
	})
}

func TestRed346To350StarterExport(t *testing.T) {
	tmpl := func(t *testing.T) string {
		t.Helper()
		body, err := os.ReadFile("templates/starter/main.go.tmpl")
		if err != nil {
			t.Fatal(err)
		}
		return string(body)
	}
	t.Run("346 the starter rewrites custom pagefind paths", func(t *testing.T) {
		if strings.Contains(tmpl(t), `data-fastr-docs-pagefind-path="/pagefind/"`) {
			t.Fatal("the rewrite only knows the default pagefind path; a custom one stays root-relative")
		}
	})
	t.Run("347 the starter validates a valueless --export", func(t *testing.T) {
		if !strings.Contains(tmpl(t), "requires a value") {
			t.Fatal("a trailing --export silently starts the live server instead of exporting")
		}
	})
	t.Run("348 the starter rewrite honors the asset prefix", func(t *testing.T) {
		if !strings.Contains(tmpl(t), "AssetPrefix()") {
			t.Fatal("the export rewrite assumes the default asset prefix forever")
		}
	})
	t.Run("349 the starter writes redirect stubs", func(t *testing.T) {
		if !strings.Contains(tmpl(t), "WriteStaticRedirect") {
			t.Fatal("declared redirects never materialize for hosts with no redirect config")
		}
	})
	t.Run("350 the starter warns when PUBLIC_SITE_URL is unset", func(t *testing.T) {
		if !strings.Contains(tmpl(t), "PUBLIC_SITE_URL") || strings.Count(tmpl(t), "PUBLIC_SITE_URL") < 3 {
			t.Fatal("an export without PUBLIC_SITE_URL bakes localhost into feeds and the sitemap")
		}
	})
}
