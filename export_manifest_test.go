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
		r := bilingualPair()
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
		r := bilingualPair()
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
		body, err := bilingualPair().ExportManifestJSON("")
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
		coverage := bilingualPair().LocaleCoverage()
		if _, ok := coverage["en"]; !ok {
			t.Fatalf("LocaleCoverage() = %v, the default locale is invisible to translation tooling", coverage)
		}
	})

	t.Run("manifest lists widget chromes", func(t *testing.T) {
		data, err := docsSiteRouter(t).ExportManifestJSON("")
		if err != nil {
			t.Fatal(err)
		}
		if !strings.Contains(string(data), "fastr-docs-sections") {
			t.Fatal("manifest carries no widget chrome inventory")
		}
	})
	t.Run("manifest routes carry publication dates", func(t *testing.T) {
		r := NewRouter()
		r.MustPage("/p", PageConfig{Title: "P", Description: "d", Source: "# P", Order: 1,
			Metadata: ContentMetadata{DatePublished: "2026-08-30"}})
		data, err := r.ExportManifestJSON("")
		if err != nil {
			t.Fatal(err)
		}
		if !strings.Contains(string(data), `"datePublished": "2026-08-30"`) {
			t.Fatal("declared dates do not reach the manifest")
		}
	})
	t.Run("manifest routes carry their translations", func(t *testing.T) {
		data, err := bilingualDocsSite(t).ExportManifestJSON("")
		if err != nil {
			t.Fatal(err)
		}
		if !strings.Contains(string(data), `"alternates"`) {
			t.Fatal("manifest routes carry no alternates map")
		}
	})
	t.Run("manifest alternates pair both directions", func(t *testing.T) {
		r := NewRouter(WithLocaleFallback("en"))
		r.MustPage("/g", PageConfig{Title: "G", Description: "d", Order: 1, Source: "# G"})
		r.MustPage("/es/g", PageConfig{Title: "G", Description: "d", Order: 2, Source: "# G", Metadata: ContentMetadata{Locale: "es"}})
		body, err := r.ExportManifestJSON("")
		if err != nil {
			t.Fatal(err)
		}
		var manifest ExportManifest
		if err := json.Unmarshal(body, &manifest); err != nil {
			t.Fatal(err)
		}
		withAlternates := 0
		for _, route := range manifest.Routes {
			if len(route.Alternates) > 0 {
				withAlternates++
			}
		}
		if withAlternates != 2 {
			t.Fatalf("routes with alternates = %d, want both directions", withAlternates)
		}
	})
	t.Run("manifest routes are sorted by path", func(t *testing.T) {
		r := NewRouter()
		r.MustPage("/z", PageConfig{Title: "Z", Description: "d", Order: 1, Source: "# Z"})
		r.MustPage("/a", PageConfig{Title: "A", Description: "d", Order: 2, Source: "# A"})
		body, err := r.ExportManifestJSON("")
		if err != nil {
			t.Fatal(err)
		}
		var manifest ExportManifest
		if err := json.Unmarshal(body, &manifest); err != nil {
			t.Fatal(err)
		}
		zi, ai := -1, -1
		for i, route := range manifest.Routes {
			if route.Path == "/z" {
				zi = i
			}
			if route.Path == "/a" {
				ai = i
			}
		}
		if zi < 0 || ai < 0 || ai > zi {
			t.Fatalf("manifest routes are not path-sorted: %+v", manifest.Routes)
		}
	})
	t.Run("manifest locales are sorted", func(t *testing.T) {
		body, err := bilingualPair().ExportManifestJSON("")
		if err != nil {
			t.Fatal(err)
		}
		var manifest ExportManifest
		if err := json.Unmarshal(body, &manifest); err != nil {
			t.Fatal(err)
		}
		for i := 1; i < len(manifest.Locales); i++ {
			if manifest.Locales[i-1] > manifest.Locales[i] {
				t.Fatalf("manifest.Locales = %v, not sorted", manifest.Locales)
			}
		}
	})
	t.Run("manifest redirects are sorted", func(t *testing.T) {
		r := NewRouter()
		r.MustPage("/z", PageConfig{Title: "Z", Description: "d", Order: 1, Source: "# Z",
			Metadata: ContentMetadata{Redirects: []string{"/zz"}}})
		r.MustPage("/a", PageConfig{Title: "A", Description: "d", Order: 2, Source: "# A",
			Metadata: ContentMetadata{Redirects: []string{"/aa"}}})
		body, err := r.ExportManifestJSON("")
		if err != nil {
			t.Fatal(err)
		}
		var manifest ExportManifest
		if err := json.Unmarshal(body, &manifest); err != nil {
			t.Fatal(err)
		}
		for i := 1; i < len(manifest.Redirects); i++ {
			if manifest.Redirects[i-1].From > manifest.Redirects[i].From {
				t.Fatalf("manifest.Redirects = %+v, not sorted", manifest.Redirects)
			}
		}
	})
	t.Run("the manifest honors the asset prefix", func(t *testing.T) {
		r := NewRouter()
		r.MustPage("/g", PageConfig{Title: "G", Description: "d", Order: 1, Source: "# G"})
		body, err := r.ExportManifestJSON("")
		if err != nil {
			t.Fatal(err)
		}
		var manifest ExportManifest
		if err := json.Unmarshal(body, &manifest); err != nil {
			t.Fatal(err)
		}
		if !strings.HasPrefix(manifest.AssetsPrefix, r.AssetPrefix()) {
			t.Fatalf("AssetsPrefix = %q, want the configured %q", manifest.AssetsPrefix, r.AssetPrefix())
		}
	})
	t.Run("the search index exposes a cache policy", func(t *testing.T) {
		if _, ok := any(&Router{}).(interface{ SearchIndexCacheControl() string }); !ok {
			t.Fatal("no cache-control accessor for the search index")
		}
	})
	t.Run("static feeds refuse a slashless path", func(t *testing.T) {
		if err := WriteStaticRSS(t.TempDir(), "", "feed.xml", []byte("<rss/>")); err == nil {
			t.Fatal("a feed path without a leading slash writes to an unpredictable place")
		}
	})
	t.Run("redirect stubs aim at their target", func(t *testing.T) {
		dir := t.TempDir()
		if err := WriteStaticRedirect(dir, "/site", "/old", "/new"); err != nil {
			t.Fatalf("no helper materializes a redirect for a host with no server config: %v", err)
		}
		body, err := os.ReadFile(filepath.Join(dir, "site", "old", "index.html"))
		if err != nil {
			t.Fatal(err)
		}
		if !strings.Contains(string(body), "/site/new") {
			t.Fatalf("stub does not aim at the target: %s", body)
		}
	})
}
