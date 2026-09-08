package docs

// Blog feed and term contracts: self links, folded slugs, future
// posts, unknown prefixes, and index requirements.
import (
	"strings"
	"testing"
	"testing/fstest"
)

func TestBlogFeeds(t *testing.T) {
	t.Run("the feed declares its self link", func(t *testing.T) {
		r := r2Blog(t, map[string]string{
			"index.md": "---\ntitle: Notas\ndescription: d\n---\n\nWelcome.",
			"uno.md":   "---\ntitle: Uno\ndescription: d\ndate: 2026-01-02\n---\n\nBody.",
		})
		feed, err := r.RSSXML(RSSConfig{Prefix: "/blog"})
		if err != nil {
			t.Fatal(err)
		}
		if !strings.Contains(string(feed), "atom:link") || !strings.Contains(string(feed), `rel="self"`) {
			t.Fatal("the channel carries no atom:link rel=self; validators reject the feed")
		}
	})

	t.Run("tag slugs fold diacritics", func(t *testing.T) {
		if got := blogSlug("diseño"); got != "diseno" {
			t.Fatalf("blogSlug(diseño) = %q, want diseno", got)
		}
	})

	t.Run("tags differing only by accents merge", func(t *testing.T) {
		posts := []*Route{
			{Title: "A", Metadata: ContentMetadata{Tags: []string{"Diseño"}}},
			{Title: "B", Metadata: ContentMetadata{Tags: []string{"diseno"}}},
		}
		if got := blogTerms(posts, false); len(got) != 1 {
			t.Fatalf("blogTerms = %v, want one folded term", got)
		}
	})

	t.Run("an unknown feed prefix errors", func(t *testing.T) {
		r := r2Blog(t, map[string]string{
			"index.md": "---\ntitle: Notas\ndescription: d\n---\n\nWelcome.",
		})
		if _, err := r.RSSXML(RSSConfig{Prefix: "/missing"}); err == nil {
			t.Fatal("RSSXML for an unregistered prefix silently renders an empty channel")
		}
	})

	t.Run("future posts stay out of the feed", func(t *testing.T) {
		r := r2Blog(t, map[string]string{
			"index.md":  "---\ntitle: Notas\ndescription: d\n---\n\nWelcome.",
			"future.md": "---\ntitle: Later\ndescription: d\ndate: 2030-01-02\n---\n\nBody.",
		})
		feed, err := r.RSSXML(RSSConfig{Prefix: "/blog"})
		if err != nil {
			t.Fatal(err)
		}
		if strings.Contains(string(feed), "Later") {
			t.Fatal("a future-dated post is already syndicated")
		}
	})

	t.Run("tag routes fold their slugs", func(t *testing.T) {
		r := r2Blog(t, map[string]string{
			"index.md": "---\ntitle: Notas\ndescription: d\n---\n\nWelcome.",
			"uno.md":   "---\ntitle: Uno\ndescription: d\ndate: 2026-01-02\ntags: [\"Diseño\"]\n---\n\nBody.",
		})
		if r.routeAtPath("/blog/tags/diseno") == nil {
			t.Fatal("a reader who types the folded tag path gets a 404")
		}
	})

	t.Run("a blog without an index errors", func(t *testing.T) {
		mapFS := fstest.MapFS{"uno.md": &fstest.MapFile{Data: []byte("---\ntitle: Uno\ndescription: d\ndate: 2026-01-02\n---\n\nBody.")}}
		r := NewRouter()
		if err := r.MarkdownBlogFS("/blog", mapFS, ".", BlogConfig{}); err == nil {
			t.Fatal("a blog with no index.md registers anyway; the archive route is untitled")
		}
	})

	t.Run("future posts stay out of the archive listing", func(t *testing.T) {
		r := r2Blog(t, map[string]string{
			"index.md":  "---\ntitle: Notas\ndescription: d\n---\n\nWelcome.",
			"future.md": "---\ntitle: Later\ndescription: d\ndate: 2030-01-02\n---\n\nBody.",
		})
		for _, post := range r.BlogPosts("/blog") {
			if post.Title == "Later" {
				t.Fatal("a future-dated post is listed as published")
			}
		}
	})
}
