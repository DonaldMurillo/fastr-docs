package cli

import (
	"os"
	"strings"
	"testing"
)

func TestStarterExportTemplateRewritesAndValidates(t *testing.T) {
	tmpl := func(t *testing.T) string {
		t.Helper()
		body, err := os.ReadFile("templates/starter/main.go.tmpl")
		if err != nil {
			t.Fatal(err)
		}
		return string(body)
	}
	t.Run("the starter rewrites custom search index paths", func(t *testing.T) {
		// The attribute-shaped regex covers index and pagefind paths
		// together, custom values included.
		if !strings.Contains(tmpl(t), "data-fastr-docs-(?:index|pagefind)-path") {
			t.Fatal("an export below a prefix leaves a custom search index URL root-relative")
		}
	})
	t.Run("the starter rewrites custom pagefind paths", func(t *testing.T) {
		if strings.Contains(tmpl(t), `data-fastr-docs-pagefind-path="/pagefind/"`) {
			t.Fatal("the rewrite only knows the default pagefind path; a custom one stays root-relative")
		}
	})
	t.Run("the starter validates a valueless --export", func(t *testing.T) {
		if !strings.Contains(tmpl(t), "requires a value") {
			t.Fatal("a trailing --export silently starts the live server instead of exporting")
		}
	})
	t.Run("the starter rewrite honors the asset prefix", func(t *testing.T) {
		if !strings.Contains(tmpl(t), "AssetPrefix()") {
			t.Fatal("the export rewrite assumes the default asset prefix forever")
		}
	})
	t.Run("the starter writes redirect stubs", func(t *testing.T) {
		if !strings.Contains(tmpl(t), "WriteStaticRedirect") {
			t.Fatal("declared redirects never materialize for hosts with no redirect config")
		}
	})
	t.Run("the starter warns when PUBLIC_SITE_URL is unset", func(t *testing.T) {
		if !strings.Contains(tmpl(t), "PUBLIC_SITE_URL") || strings.Count(tmpl(t), "PUBLIC_SITE_URL") < 3 {
			t.Fatal("an export without PUBLIC_SITE_URL bakes localhost into feeds and the sitemap")
		}
	})
}
