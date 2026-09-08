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
}
