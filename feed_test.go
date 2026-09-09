package docs

// Blog feeds and term pages: self links, folded slugs, future posts,
// unknown prefixes, and index requirements.
import (
	"strings"
	"testing"
	"testing/fstest"
)

func TestBlogFeeds(t *testing.T) {
	t.Run("the feed declares its self link", func(t *testing.T) {
		r := blogRouter(t, map[string]string{
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
		r := blogRouter(t, map[string]string{
			"index.md": "---\ntitle: Notas\ndescription: d\n---\n\nWelcome.",
		})
		if _, err := r.RSSXML(RSSConfig{Prefix: "/missing"}); err == nil {
			t.Fatal("RSSXML for an unregistered prefix silently renders an empty channel")
		}
	})

	t.Run("future posts stay out of the feed", func(t *testing.T) {
		r := blogRouter(t, map[string]string{
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
		r := blogRouter(t, map[string]string{
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
		r := blogRouter(t, map[string]string{
			"index.md":  "---\ntitle: Notas\ndescription: d\n---\n\nWelcome.",
			"future.md": "---\ntitle: Later\ndescription: d\ndate: 2030-01-02\n---\n\nBody.",
		})
		for _, post := range r.BlogPosts("/blog") {
			if post.Title == "Later" {
				t.Fatal("a future-dated post is listed as published")
			}
		}
	})

	t.Run("noindex posts stay out of the feed", func(t *testing.T) {
		r := blogRouter(t, map[string]string{
			"index.md":  "---\ntitle: Blog\n---\n\nIntro.\n",
			"public.md": "---\ntitle: Public\ndate: 2026-08-01\n---\n\nBody.\n",
			"secret.md": "---\ntitle: Secret\ndate: 2026-08-02\nnoindex: true\n---\n\nHidden.\n",
		})
		xml, err := r.RSSXML(RSSConfig{Prefix: "/blog", Title: "T", Description: "D", SiteURL: "https://x.example"})
		if err != nil {
			t.Fatal(err)
		}
		if strings.Contains(string(xml), "/blog/secret") {
			t.Fatal("noindex post syndicated")
		}
		if !strings.Contains(string(xml), "lastBuildDate") {
			t.Fatal("channel carries no last build date")
		}
	})
	t.Run("the feed declares its collection language", func(t *testing.T) {
		r := blogRouter(t, map[string]string{
			"index.md": "---\ntitle: Blog\n---\n\nIntro.\n",
			"post.md":  "---\ntitle: Post\ndate: 2026-08-01\n---\n\nBody.\n",
		}, BlogConfig{Title: "Blog", Order: 2, DefaultLocale: "es"})
		xml, err := r.RSSXML(RSSConfig{Prefix: "/blog", Title: "T", Description: "D", SiteURL: "https://x.example"})
		if err != nil {
			t.Fatal(err)
		}
		if !strings.Contains(string(xml), "<language>es</language>") {
			t.Fatal("channel declares no language")
		}
	})
	t.Run("the feed carries a last build date", func(t *testing.T) {
		r := blogRouter(t, map[string]string{
			"index.md": "---\ntitle: Blog\n---\n\nIntro.\n",
			"post.md":  "---\ntitle: Post\ndate: 2026-08-01\n---\n\nBody.\n",
		})
		xml, err := r.RSSXML(RSSConfig{Prefix: "/blog", Title: "T", Description: "D", SiteURL: "https://x.example"})
		if err != nil {
			t.Fatal(err)
		}
		if !strings.Contains(string(xml), "lastBuildDate") {
			t.Fatal("channel carries no last build date")
		}
	})

	t.Run("the feed description uses the excerpt", func(t *testing.T) {
		r := blogAcrossTwoYears(t)
		route := r.routeAtPath("/blog/older")
		route.Metadata.Excerpt = "Hand-written summary."
		feed, err := r.RSSXML(RSSConfig{Prefix: "/blog"})
		if err != nil {
			t.Fatal(err)
		}
		if !strings.Contains(string(feed), "Hand-written summary.") {
			t.Fatal("a curated excerpt never reaches the feed description")
		}
	})
	t.Run("the feed names its generator", func(t *testing.T) {
		r := blogAcrossTwoYears(t)
		feed, err := r.RSSXML(RSSConfig{Prefix: "/blog"})
		if err != nil {
			t.Fatal(err)
		}
		if !strings.Contains(string(feed), "<generator>") {
			t.Fatal("the feed does not say what generated it")
		}
	})
	t.Run("feeds carry their collection title", func(t *testing.T) {
		r := blogRouter(t, map[string]string{
			"index.md": "---\ntitle: Notas\ndescription: d\n---\n\nWelcome.",
		}, BlogConfig{Title: "Engineering Notes"})
		feed, err := r.RSSXML(RSSConfig{Prefix: "/blog"})
		if err != nil {
			t.Fatal(err)
		}
		if !strings.Contains(string(feed), "<title>Engineering Notes</title>") {
			t.Fatal("a titled collection feeds under a synthesized name")
		}
	})
	t.Run("feeds require a rooted prefix", func(t *testing.T) {
		r := blogAcrossTwoYears(t)
		if _, err := r.RSSXML(RSSConfig{Prefix: "blog"}); err == nil {
			t.Fatal("a prefix without a slash produces a feed no host can serve")
		}
	})
}
