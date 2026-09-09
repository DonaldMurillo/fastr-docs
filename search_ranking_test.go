package docs

// Search behavior: folded and punctuation-tolerant terms, URL
// slugs, exact titles, configured weights, and bounded result counts.
import (
	"strings"
	"testing"
)

func TestSearchRanking(t *testing.T) {
	t.Run("search terms survive punctuation", func(t *testing.T) {
		r := NewRouter()
		r.MustPage("/g", PageConfig{Title: "G", Description: "d", Order: 1, Source: "# G\n\nThe route tree navigation page."})
		if len(r.Search("route-tree")) == 0 {
			t.Fatal(`query "route-tree" finds nothing; the hyphen is not a word boundary`)
		}
	})

	t.Run("a zero limit means zero results", func(t *testing.T) {
		r := NewRouter()
		r.MustPage("/g", PageConfig{Title: "Guide", Description: "d", Order: 1, Source: "# Guide"})
		if got := r.SearchLimited("guide", 0); len(got) != 0 {
			t.Fatalf("SearchLimited(q, 0) returned %d results", len(got))
		}
	})

	t.Run("negative search weights are rejected", func(t *testing.T) {
		r := NewRouter(WithSearchWeights(map[string]int{"title": -5}))
		r.MustPage("/g", PageConfig{Title: "Guide", Description: "d", Order: 1, Source: "# Guide"})
		if err := r.Validate(); err == nil || !strings.Contains(err.Error(), "weight") {
			t.Fatalf("Validate() = %v, want a negative-weight complaint", err)
		}
	})

	t.Run("search text strips shortcode syntax", func(t *testing.T) {
		r := NewRouter()
		r.MustPage("/g", PageConfig{Title: "G", Description: "d", Order: 1, Source: "{{< callout >}}note{{< /callout >}}\n\nprose"})
		for _, entry := range r.SearchIndex() {
			if strings.Contains(entry.Text, "{{<") {
				t.Fatalf("search text carries shortcode markers: %q", entry.Text)
			}
		}
	})

	t.Run("an exact title outranks a substring title", func(t *testing.T) {
		r := NewRouter()
		r.MustPage("/advanced", PageConfig{Title: "Guide advanced", Description: "d", Source: "# A", Order: 1})
		r.MustPage("/plain", PageConfig{Title: "Guide", Description: "d", Source: "# P", Order: 2})
		results := r.Search("guide")
		if len(results) == 0 || results[0].Entry.Path != "/plain" {
			t.Fatalf("top result = %+v, want the exact-title page", results)
		}
	})

	t.Run("a negative limit also means none", func(t *testing.T) {
		r := NewRouter()
		r.MustPage("/g", PageConfig{Title: "Guide", Description: "d", Order: 1, Source: "# Guide"})
		if got := r.SearchLimited("guide", -3); len(got) != 0 {
			t.Fatalf("SearchLimited(q, -3) returned %d results", len(got))
		}
	})

	t.Run("coverage names the default locale", func(t *testing.T) {
		coverage := bilingualPair().LocaleCoverage()
		if _, ok := coverage["en"]; !ok {
			t.Fatalf("LocaleCoverage() = %v, the default locale is invisible to translation tooling", coverage)
		}
	})

	t.Run("search matches the url slug", func(t *testing.T) {
		r := NewRouter()
		r.MustPage("/docs/deep-dive", PageConfig{Title: "Deep dive", Description: "d", Order: 1, Source: "# Deep"})
		r.MustPage("/docs/other", PageConfig{Title: "Other", Description: "unrelated words", Order: 2, Source: "# O"})
		for _, result := range r.Search("deep-dive") {
			if result.Entry.Path == "/docs/deep-dive" {
				return
			}
		}
		t.Fatal("a reader who types a url fragment finds nothing; paths are not searched")
	})

	t.Run("unknown search weight fields are named", func(t *testing.T) {
		r := NewRouter(WithSearchWeights(map[string]int{"titel": 5}))
		r.MustPage("/g", PageConfig{Title: "G", Description: "d", Order: 1, Source: "# G"})
		for _, warning := range r.Warnings() {
			if strings.Contains(warning, "titel") {
				return
			}
		}
		t.Fatalf("Warnings() = %v, a typo'd field name does nothing silently", r.Warnings())
	})

	t.Run("search folds diacritics", func(t *testing.T) {
		r := NewRouter()
		r.MustPage("/es/p", PageConfig{Title: "Traducción", Description: "d", Source: "# Traducción", Order: 1})
		if got := r.Search("traduccion"); len(got) == 0 {
			t.Fatal("traduccion does not match traducción")
		}
	})
	t.Run("search lowercases unicode", func(t *testing.T) {
		r := NewRouter()
		r.MustPage("/p", PageConfig{Title: "ÜBER", Description: "d", Source: "# über", Order: 1})
		if got := r.Search("UBER"); len(got) == 0 {
			t.Fatal("uppercase query missed an umlaut title")
		}
	})
	t.Run("search text folds diacritics", func(t *testing.T) {
		if foldSearchText("Diseño") != "diseno" || foldSearchText("TRADUCCIÓN") != "traduccion" {
			t.Fatalf("fold = %q / %q", foldSearchText("Diseño"), foldSearchText("TRADUCCIÓN"))
		}
	})
	t.Run("tag matching folds diacritics", func(t *testing.T) {
		r := NewRouter()
		r.MustPage("/p", PageConfig{Title: "T", Description: "d", Source: "# T", Order: 1, Tags: []string{"diseño"}})
		if got := r.Search("diseno"); len(got) == 0 {
			t.Fatal("a tag written without its diacritic does not match")
		}
	})
	t.Run("search entries carry heading ids", func(t *testing.T) {
		r := NewRouter()
		r.MustPage("/p", PageConfig{Title: "T", Description: "d", Source: "# T\n\n## Section One", Order: 1})
		for _, e := range r.SearchIndex() {
			if e.Path == "/p" && len(e.Headings) == 0 {
				t.Fatal("page with an h2 carries no headings into its search entry")
			}
		}
	})
	t.Run("search text strips html comments", func(t *testing.T) {
		r := NewRouter()
		r.MustPage("/g", PageConfig{Title: "G", Description: "d", Order: 1, Source: "<!-- secret note -->\n\nbody"})
		for _, entry := range r.SearchIndex() {
			if strings.Contains(entry.Text, "secret note") {
				t.Fatal("an html comment is searchable content")
			}
		}
	})
	t.Run("repeated terms rank higher", func(t *testing.T) {
		r := NewRouter()
		r.MustPage("/once", PageConfig{Title: "A", Description: "d", Order: 1, Source: "# zebra once"})
		r.MustPage("/many", PageConfig{Title: "B", Description: "d", Order: 2, Source: "# zebra zebra zebra"})
		results := r.Search("zebra")
		if len(results) != 2 || results[0].Entry.Path != "/many" {
			t.Fatalf("term frequency does not rank: %+v", results)
		}
	})
	t.Run("body weighting is configurable", func(t *testing.T) {
		r := NewRouter(WithSearchWeights(map[string]int{"title": 1}))
		r.MustPage("/t", PageConfig{Title: "zebra", Description: "d", Order: 1, Source: "# z"})
		r.MustPage("/b", PageConfig{Title: "plain", Description: "d", Order: 2, Source: "# zebra zebra body"})
		results := r.Search("zebra")
		if len(results) != 2 || results[0].Entry.Path != "/b" {
			t.Fatalf("a term the body repeats cannot outrank a title tuned down to its weight: %+v", results)
		}
	})
	t.Run("phrase queries keep their words together", func(t *testing.T) {
		r := NewRouter()
		r.MustPage("/split", PageConfig{Title: "A", Description: "d", Order: 1, Source: "# route and tree apart"})
		r.MustPage("/joined", PageConfig{Title: "B", Description: "d", Order: 2, Source: "# the route tree page"})
		results := r.Search(`"route tree"`)
		for _, result := range results {
			if result.Entry.Path == "/split" {
				t.Fatal("a phrase matches pages missing the phrase")
			}
		}
	})
}
