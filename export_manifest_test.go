package docs

// Export artifacts: the sitemap namespaces and dates, the
// manifest's locales, drawers, and search index, and agent assets.
import (
	"encoding/json"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestExportArtifacts(t *testing.T) {
	t.Run("the sitemap declares the xhtml namespace", func(t *testing.T) {
		r := r2Bilingual()
		if !strings.Contains(string(r.Sitemap()), "xmlns:xhtml=") {
			t.Fatal("xhtml:link alternates are emitted with an unbound prefix")
		}
	})

	t.Run("the manifest honors a custom search index path", func(t *testing.T) {
		r := NewRouter(WithSearchIndexPath("/idx.json"))
		r.MustPage("/g", PageConfig{Title: "G", Description: "d", Order: 1, Source: "# G"})
		body, err := r.ExportManifestJSON("")
		if err != nil {
			t.Fatal(err)
		}
		var manifest ExportManifest
		if err := json.Unmarshal(body, &manifest); err != nil {
			t.Fatal(err)
		}
		if manifest.SearchIndex != "/idx.json" {
			t.Fatalf("manifest.SearchIndex = %q, want /idx.json", manifest.SearchIndex)
		}
	})

	t.Run("agent card links are prefixed under a base", func(t *testing.T) {
		dir := t.TempDir()
		handler := http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			if req.URL.Path == "/.well-known/agent-card.json" {
				_, _ = w.Write([]byte(`{"url":"/","documentation":"/docs"}`))
				return
			}
			http.NotFound(w, req)
		})
		if err := WriteAgentAssets(dir, "/site", handler); err != nil {
			t.Fatal(err)
		}
		body, err := os.ReadFile(filepath.Join(dir, "site", ".well-known", "agent-card.json"))
		if err != nil {
			t.Fatal(err)
		}
		if !strings.Contains(string(body), "/site/docs") {
			t.Fatalf("agent card links were not prefixed: %s", body)
		}
	})

	t.Run("the manifest reports the effective locale", func(t *testing.T) {
		r := r2Bilingual()
		body, err := r.ExportManifestJSON("")
		if err != nil {
			t.Fatal(err)
		}
		var manifest ExportManifest
		if err := json.Unmarshal(body, &manifest); err != nil {
			t.Fatal(err)
		}
		for _, route := range manifest.Routes {
			if route.Path == "/docs/guide" && route.Locale != "en" {
				t.Fatalf("manifest locale for an unmarked page = %q, want the default en", route.Locale)
			}
		}
	})

	t.Run("the sitemap emits only parseable lastmod values", func(t *testing.T) {
		r := NewRouter()
		r.MustPage("/g", PageConfig{Title: "G", Description: "d", Order: 1, Source: "# G",
			Metadata: ContentMetadata{DateModified: "sometime"}})
		if strings.Contains(string(r.Sitemap()), "sometime") {
			t.Fatal("a garbage date reached lastmod verbatim")
		}
	})

	t.Run("the manifest lists the default locale", func(t *testing.T) {
		body, err := r2Bilingual().ExportManifestJSON("")
		if err != nil {
			t.Fatal(err)
		}
		var manifest ExportManifest
		if err := json.Unmarshal(body, &manifest); err != nil {
			t.Fatal(err)
		}
		if !strings.Contains(strings.Join(manifest.Locales, ","), "en") {
			t.Fatalf("manifest.Locales = %v, want the default named", manifest.Locales)
		}
	})

	t.Run("the manifest precache lists no duplicate drawer", func(t *testing.T) {
		r := NewRouter()
		r.MustPage("/", PageConfig{Title: "Home", Description: "d", Order: 1, Source: "# H"})
		r.MustPage("/guide", PageConfig{Title: "Guide", Description: "d", Order: 2, Source: "# G",
			Metadata: ContentMetadata{Version: "v2"}})
		r.MustPage("/v1/guide", PageConfig{Title: "Guide v1", Description: "d", Order: 3, Source: "# G1",
			Metadata: ContentMetadata{Version: "v1"}})
		body, err := r.ExportManifestJSON("")
		if err != nil {
			t.Fatal(err)
		}
		var manifest ExportManifest
		if err := json.Unmarshal(body, &manifest); err != nil {
			t.Fatal(err)
		}
		seen := map[string]int{}
		for _, drawer := range manifest.Drawers {
			seen[drawer]++
			if seen[drawer] > 1 {
				t.Fatalf("manifest precaches %q twice: %v", drawer, manifest.Drawers)
			}
		}
	})

	t.Run("coverage names the default locale", func(t *testing.T) {
		coverage := r2Bilingual().LocaleCoverage()
		if _, ok := coverage["en"]; !ok {
			t.Fatalf("LocaleCoverage() = %v, the default locale is invisible to translation tooling", coverage)
		}
	})
}
