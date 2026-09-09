package docs

// Strict metadata validation: dates, redirect graphs, slugs,
// tags, versions, bylines, and badges.
import (
	"strings"
	"testing"
)

func TestRouteMetadataValidation(t *testing.T) {
	t.Run("a malformed publish date is flagged", func(t *testing.T) {
		r := NewRouter()
		r.MustPage("/g", PageConfig{Title: "G", Description: "d", Order: 1, Source: "# G",
			Metadata: ContentMetadata{DatePublished: "yesterday"}})
		if err := r.Validate(); err == nil || !strings.Contains(err.Error(), "date") {
			t.Fatalf("Validate() = %v, want a malformed-date complaint", err)
		}
	})

	t.Run("a modified date before publication is flagged", func(t *testing.T) {
		r := NewRouter()
		r.MustPage("/g", PageConfig{Title: "G", Description: "d", Order: 1, Source: "# G",
			Metadata: ContentMetadata{DatePublished: "2026-03-04T00:00:00Z", DateModified: "2026-01-02T00:00:00Z"}})
		if err := r.Validate(); err == nil || !strings.Contains(err.Error(), "modified") {
			t.Fatalf("Validate() = %v, want a date-order complaint", err)
		}
	})

	t.Run("a three-hop redirect cycle is flagged", func(t *testing.T) {
		r := NewRouter()
		r.MustPage("/a", PageConfig{Title: "A", Description: "d", Order: 1, Source: "# A", Metadata: ContentMetadata{Redirects: []string{"/b"}}})
		r.MustPage("/b", PageConfig{Title: "B", Description: "d", Order: 2, Source: "# B", Metadata: ContentMetadata{Redirects: []string{"/c"}}})
		r.MustPage("/c", PageConfig{Title: "C", Description: "d", Order: 3, Source: "# C", Metadata: ContentMetadata{Redirects: []string{"/a"}}})
		if err := r.Validate(); err == nil || !strings.Contains(err.Error(), "redirect") {
			t.Fatalf("Validate() = %v, want a redirect-cycle complaint", err)
		}
	})

	t.Run("a future publish date is a warning", func(t *testing.T) {
		r := NewRouter()
		r.MustPage("/g", PageConfig{Title: "G", Description: "d", Order: 1, Source: "# G",
			Metadata: ContentMetadata{DatePublished: "2030-01-02T00:00:00Z"}})
		for _, warning := range r.Warnings() {
			if strings.Contains(warning, "future") {
				return
			}
		}
		t.Fatalf("Warnings() = %v, want a future-date note", r.Warnings())
	})

	t.Run("a slug containing a slash is flagged", func(t *testing.T) {
		r := NewRouter()
		r.MustPage("/g", PageConfig{Title: "G", Description: "d", Order: 1, Source: "# G",
			Metadata: ContentMetadata{Slug: "a/b"}})
		if err := r.Validate(); err == nil || !strings.Contains(err.Error(), "slug") {
			t.Fatalf("Validate() = %v, want a multi-segment slug complaint", err)
		}
	})

	t.Run("empty tag entries are dropped", func(t *testing.T) {
		r := NewRouter()
		r.MustPage("/g", PageConfig{Title: "G", Description: "d", Order: 1, Source: "# G",
			Metadata: ContentMetadata{Tags: []string{"", "go"}}})
		route := r.routeAtPath("/g")
		for _, tag := range route.Tags {
			if strings.TrimSpace(tag) == "" {
				t.Fatal("an empty tag reached the route")
			}
		}
	})

	t.Run("a redirect without a leading slash is flagged", func(t *testing.T) {
		r := NewRouter()
		r.MustPage("/g", PageConfig{Title: "G", Description: "d", Order: 1, Source: "# G",
			Metadata: ContentMetadata{Redirects: []string{"blog"}}})
		if err := r.Validate(); err == nil || !strings.Contains(err.Error(), "redirect") {
			t.Fatalf("Validate() = %v, want a redirect-shape complaint", err)
		}
	})

	t.Run("a redirect source claimed twice is flagged", func(t *testing.T) {
		r := NewRouter()
		r.MustPage("/a", PageConfig{Title: "A", Description: "d", Order: 1, Source: "# A", Metadata: ContentMetadata{Redirects: []string{"/old"}}})
		r.MustPage("/b", PageConfig{Title: "B", Description: "d", Order: 2, Source: "# B", Metadata: ContentMetadata{Redirects: []string{"/old"}}})
		if err := r.Validate(); err == nil || !strings.Contains(err.Error(), "redirect") {
			t.Fatalf("Validate() = %v, want a duplicate-redirect complaint", err)
		}
	})

	t.Run("an absolute redirect target is flagged", func(t *testing.T) {
		r := NewRouter()
		r.MustPage("/g", PageConfig{Title: "G", Description: "d", Order: 1, Source: "# G",
			Metadata: ContentMetadata{Redirects: []string{"https://example.com/x"}}})
		if err := r.Validate(); err == nil || !strings.Contains(err.Error(), "redirect") {
			t.Fatalf("Validate() = %v, want an absolute redirect complaint", err)
		}
	})

	t.Run("a malformed page version is flagged", func(t *testing.T) {
		r := NewRouter()
		r.MustPage("/g", PageConfig{Title: "G", Description: "d", Order: 1, Source: "# G",
			Metadata: ContentMetadata{Version: "v 1"}})
		if err := r.Validate(); err == nil || !strings.Contains(err.Error(), "version") {
			t.Fatalf("Validate() = %v, want a version-shape complaint", err)
		}
	})

	t.Run("excerpt length is bounded", func(t *testing.T) {
		long := strings.Repeat("word ", 200)
		r := NewRouter()
		r.MustPage("/g", PageConfig{Title: "G", Description: "d", Order: 1, Source: "# G",
			Metadata: ContentMetadata{Excerpt: long}})
		if err := r.Validate(); err == nil || !strings.Contains(err.Error(), "excerpt") {
			t.Fatalf("Validate() = %v, want an overlong-excerpt note", err)
		}
	})

	t.Run("authors with commas are flagged", func(t *testing.T) {
		r := NewRouter()
		r.MustPage("/g", PageConfig{Title: "G", Description: "d", Order: 1, Source: "# G",
			Metadata: ContentMetadata{Authors: []string{"Doe, Jane"}}})
		if err := r.Validate(); err == nil || !strings.Contains(err.Error(), "author") {
			t.Fatalf("Validate() = %v, want a comma-in-author complaint", err)
		}
	})

	t.Run("badges keep their labels short", func(t *testing.T) {
		r := NewRouter()
		r.MustPage("/g", PageConfig{Title: "G", Description: "d", Order: 1, Source: "# G",
			Badge: NavBadge{Label: strings.Repeat("n", 40), Tone: NavBadgeToneInfo}})
		if err := r.Validate(); err == nil || !strings.Contains(err.Error(), "badge") {
			t.Fatalf("Validate() = %v, want an overlong-badge complaint", err)
		}
	})

	t.Run("source and source path are mutually exclusive", func(t *testing.T) {
		r := NewRouter()
		if err := r.Page("/x", PageConfig{Title: "X", Description: "x", Source: "# a", SourcePath: "content/build-blog.md"}); err == nil {
			t.Fatal("Page accepted both Source and SourcePath")
		}
	})
	t.Run("redirect loops are flagged", func(t *testing.T) {
		r := NewRouter()
		r.MustPage("/a", PageConfig{Title: "A", Description: "x", Source: "# a", Order: 1, Metadata: ContentMetadata{Redirects: []string{"/a"}}})
		if err := r.Validate(); err == nil {
			t.Fatal("Validate accepted a redirect onto itself")
		}
	})
	t.Run("redirect cycles across pages are flagged", func(t *testing.T) {
		r := NewRouter()
		r.MustPage("/a", PageConfig{Title: "A", Description: "x", Source: "# a", Order: 1, Metadata: ContentMetadata{Redirects: []string{"/b"}}})
		r.MustPage("/b", PageConfig{Title: "B", Description: "x", Source: "# b", Order: 2, Metadata: ContentMetadata{Redirects: []string{"/a"}}})
		if err := r.Validate(); err == nil {
			t.Fatal("Validate accepted a two-hop redirect cycle")
		}
	})
	t.Run("duplicate redirects within one route are flagged", func(t *testing.T) {
		r := NewRouter()
		r.MustPage("/g", PageConfig{Title: "G", Description: "d", Order: 1, Source: "# G", Metadata: ContentMetadata{Redirects: []string{"/old", "/old"}}})
		if err := r.Validate(); err == nil || !strings.Contains(err.Error(), "redirect") {
			t.Fatalf("Validate() = %v, want a duplicate-redirect complaint", err)
		}
	})
	t.Run("redirects onto drafts are flagged", func(t *testing.T) {
		r := NewRouter()
		r.MustPage("/g", PageConfig{Title: "G", Description: "d", Order: 1, Source: "# G", Metadata: ContentMetadata{Redirects: []string{"/draft"}}})
		r.MustPage("/draft", PageConfig{Title: "D", Description: "d", Order: 2, Source: "# D", Metadata: ContentMetadata{Draft: true}})
		if err := r.Validate(); err == nil || !strings.Contains(err.Error(), "draft") {
			t.Fatalf("Validate() = %v, want a redirect-to-draft complaint", err)
		}
	})
	t.Run("a redirect with inner whitespace is flagged", func(t *testing.T) {
		r := NewRouter()
		r.MustPage("/g", PageConfig{Title: "G", Description: "d", Order: 1, Source: "# G",
			Metadata: ContentMetadata{Redirects: []string{"/a b"}}})
		if err := r.Validate(); err == nil || !strings.Contains(err.Error(), "redirect") {
			t.Fatalf("Validate() = %v, want a redirect-whitespace complaint", err)
		}
	})
	t.Run("duplicate slugs are flagged", func(t *testing.T) {
		r := NewRouter()
		if err := r.Page("/one", PageConfig{Title: "One", Description: "x", Source: "# a", Order: 1, Metadata: ContentMetadata{Slug: "same"}}); err != nil {
			t.Fatal(err)
		}
		err := r.Page("/two", PageConfig{Title: "Two", Description: "x", Source: "# b", Order: 2, Metadata: ContentMetadata{Slug: "same"}})
		if err == nil || !strings.Contains(err.Error(), "/same") {
			t.Fatalf("two pages claiming one slug must collide loudly: %v", err)
		}
	})
	t.Run("alternates naming unserved languages are flagged", func(t *testing.T) {
		r := NewRouter()
		r.MustPage("/x", PageConfig{Title: "X", Description: "x", Source: "# a", Order: 1,
			Metadata: ContentMetadata{Alternates: map[string]string{"fr": "/nowhere"}}})
		if err := r.Validate(); err == nil {
			t.Fatal("Validate accepted a fr alternate no route serves")
		}
	})
	t.Run("duplicate sibling titles are flagged", func(t *testing.T) {
		r := docsSiteRouter(t)
		r.MustPage("/docs/twin", PageConfig{Title: "Start", Description: "x", Source: "# a", Order: 9})
		if err := r.Validate(); err == nil {
			t.Fatal("Validate accepted two siblings titled Start")
		}
	})
	t.Run("canonical urls must be absolute", func(t *testing.T) {
		r := NewRouter()
		r.MustPage("/x", PageConfig{Title: "X", Description: "x", Source: "# a", Order: 1,
			Metadata: ContentMetadata{CanonicalURL: "/relative"}})
		if err := r.Validate(); err == nil {
			t.Fatal("Validate accepted a relative canonical URL")
		}
	})
	t.Run("a canonical with a fragment is flagged", func(t *testing.T) {
		r := NewRouter()
		r.MustPage("/g", PageConfig{Title: "G", Description: "d", Order: 1, Source: "# G",
			Metadata: ContentMetadata{CanonicalURL: "https://example.com/docs/g#section"}})
		if err := r.Validate(); err == nil || !strings.Contains(err.Error(), "canonical") {
			t.Fatalf("Validate() = %v, want a canonical-fragment complaint", err)
		}
	})
	t.Run("stored locales are normalized at registration", func(t *testing.T) {
		r := NewRouter()
		r.MustPage("/x", PageConfig{Title: "X", Description: "x", Source: "# a", Order: 1, Metadata: ContentMetadata{Locale: "ES"}})
		if got := r.routeAtPath("/x").Metadata.Locale; got != "es" {
			t.Fatalf("route keeps raw locale %q", got)
		}
	})
	t.Run("translation_of to a draft warns", func(t *testing.T) {
		r := NewRouter()
		r.MustPage("/draft", PageConfig{Title: "D", Description: "d", Order: 1, Source: "# D", Metadata: ContentMetadata{Draft: true}})
		r.MustPage("/es/draft", PageConfig{Title: "D", Description: "d", Order: 2, Source: "# D", Metadata: ContentMetadata{Locale: "es", TranslationOf: "/draft"}})
		if err := r.Validate(); err != nil {
			t.Fatalf("Validate() = %v, translating a not-yet-published page should warn, not fail", err)
		}
		for _, warning := range r.Warnings() {
			if strings.Contains(warning, "/draft") {
				return
			}
		}
		t.Fatal("Warnings() never mentions the draft translation")
	})
	t.Run("an unknown page template is flagged", func(t *testing.T) {
		r := NewRouter()
		r.MustPage("/g", PageConfig{Title: "G", Description: "d", Order: 1, Source: "# G",
			Metadata: ContentMetadata{PageTemplate: "fancy"}})
		if err := r.Validate(); err == nil || !strings.Contains(err.Error(), "template") {
			t.Fatalf("Validate() = %v, want an unknown-template complaint", err)
		}
	})
	t.Run("a relative edit URL is flagged", func(t *testing.T) {
		r := NewRouter()
		r.MustPage("/g", PageConfig{Title: "G", Description: "d", Order: 1, Source: "# G",
			Metadata: ContentMetadata{EditURL: "edit/page.md"}})
		if err := r.Validate(); err == nil || !strings.Contains(err.Error(), "edit") {
			t.Fatalf("Validate() = %v, want an edit-URL complaint", err)
		}
	})
	t.Run("order gaps are warned", func(t *testing.T) {
		r := NewRouter()
		g := r.MustGroup("/docs", GroupConfig{Title: "Docs", Description: "d", Order: 1})
		g.MustPage("a", PageConfig{Title: "A", Description: "d", Order: 1, Source: "# A"})
		g.MustPage("z", PageConfig{Title: "Z", Description: "d", Order: 9, Source: "# Z"})
		for _, warning := range r.Warnings() {
			if strings.Contains(warning, "order") {
				return
			}
		}
		t.Fatal("an order gap of eight between siblings passes silently")
	})
	t.Run("heading level jumps are named", func(t *testing.T) {
		r := NewRouter()
		r.MustPage("/g", PageConfig{Title: "G", Description: "d", Order: 1, Source: "## Top\n\n#### Skipped three\n\nBody."})
		if err := r.Validate(); err == nil || !strings.Contains(err.Error(), "heading") {
			t.Fatalf("Validate() = %v, want a heading-level jump complaint", err)
		}
	})
}
