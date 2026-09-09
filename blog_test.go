package docs

import (
	"context"
	"encoding/json"
	"encoding/xml"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"testing/fstest"

	uiapp "github.com/DonaldMurillo/gofastr/core-ui/app"
)

func TestMarkdownBlogBuildsArchiveAndPublishedRSS(t *testing.T) {
	dir := t.TempDir()
	files := map[string]string{
		"index.md":   "# Updates\n\nNews about the project.",
		"older.md":   "---\ntitle: Older release\ndescription: The older release notes.\ndate: 2026-01-10\nauthors: [Ada]\ntags: [release, stable]\n---\n# Older release\n\nOlder details.",
		"newer.md":   "---\ntitle: Newer release\ndescription: The newer release notes.\ndate: 2026-02-15\nauthors: [Grace]\ntags: [release, new]\nslug: releases/newer\n---\n# Newer release\n\nNewer details.",
		"draft.md":   "---\ntitle: Work in progress\ndate: 2026-03-01\ndraft: true\n---\n# Work in progress",
		"private.md": "---\ntitle: Private note\ndate: 2026-04-01\nnoindex: true\n---\n# Private note",
	}
	for name, body := range files {
		path := filepath.Join(dir, name)
		if err := os.WriteFile(path, []byte(body), 0o644); err != nil {
			t.Fatal(err)
		}
	}

	r := NewRouter(WithSiteName("Release notes"))
	if err := r.MarkdownBlog("/blog", dir, BlogConfig{Title: "Release notes", Description: "Project updates.", Order: 2}); err != nil {
		t.Fatalf("MarkdownBlog() error = %v", err)
	}
	if got := r.routes["/blog"].BlogIndex; !got {
		t.Fatal("archive route was not marked as BlogIndex")
	}
	posts := r.BlogPosts("/blog")
	if len(posts) != 3 {
		t.Fatalf("published posts = %d, want 3", len(posts))
	}
	if posts[0].Path != "/blog/private" || posts[1].Path != "/blog/releases/newer" || posts[2].Path != "/blog/older" {
		t.Fatalf("post order = [%s, %s, %s]", posts[0].Path, posts[1].Path, posts[2].Path)
	}
	archive := r.routes["/blog"].page
	if archive == nil || !strings.Contains(pageSource(archive), "Newer release") || !strings.Contains(pageSource(archive), "/blog/releases/newer") {
		t.Fatalf("archive did not contain generated post links: %q", pageSource(archive))
	}

	body, err := r.RSSXML(RSSConfig{Prefix: "/blog", Title: "Release notes", SiteURL: "https://example.com/docs", Limit: 10})
	if err != nil {
		t.Fatalf("RSSXML() error = %v", err)
	}
	if !strings.HasPrefix(string(body), "<?xml version=") {
		t.Fatalf("RSS body has no XML header: %q", body[:minInt(len(body), 40)])
	}
	var feed rssDocument
	if err := xml.Unmarshal(body, &feed); err != nil {
		t.Fatalf("RSS XML invalid: %v\n%s", err, body)
	}
	if feed.Channel.Title != "Release notes" || len(feed.Channel.Items) != 2 {
		t.Fatalf("feed = %#v, want two public items", feed.Channel)
	}
	if feed.Channel.Items[0].Title != "Newer release" || feed.Channel.Items[0].Link != "https://example.com/docs/blog/releases/newer" {
		t.Fatalf("feed first item = %#v", feed.Channel.Items[0])
	}
	if strings.Contains(string(body), "Private note") || strings.Contains(string(body), "Work in progress") {
		t.Fatalf("RSS leaked no-index or draft content: %s", body)
	}
	if !strings.Contains(string(body), "release") || !strings.Contains(string(body), "Grace") {
		t.Fatalf("RSS omitted post metadata: %s", body)
	}
}

func TestMarkdownBlogFSIncludesDraftsAndStaticRSSUsesBasePath(t *testing.T) {
	content := fstest.MapFS{
		"blog/index.md": &fstest.MapFile{Data: []byte("# Blog")},
		"blog/post.md":  &fstest.MapFile{Data: []byte("---\ntitle: Post\ndate: 2026-05-01\ndraft: true\n---\n# Post")},
	}
	r := NewRouter()
	if err := r.MarkdownBlogFS("/blog", content, "blog", BlogConfig{IncludeDrafts: true}); err != nil {
		t.Fatalf("MarkdownBlogFS() error = %v", err)
	}
	if got := len(r.BlogPosts("/blog")); got != 1 {
		t.Fatalf("draft-inclusive posts = %d, want 1", got)
	}
	body, err := r.RSSXML(RSSConfig{Prefix: "/blog"})
	if err != nil {
		t.Fatalf("RSSXML() error = %v", err)
	}
	if strings.Contains(string(body), "Post") {
		t.Fatalf("RSS included a draft post: %s", body)
	}

	dir := t.TempDir()
	if err := WriteStaticRSS(dir, "/docs", "/blog/feed.xml", body); err != nil {
		t.Fatalf("WriteStaticRSS() error = %v", err)
	}
	path := filepath.Join(dir, "docs", "blog", "feed.xml")
	written, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read static feed: %v", err)
	}
	if string(written) != string(body) {
		t.Fatal("static feed differs from generated feed")
	}
}

func TestMarkdownBlogArchiveYearsHaveDistinctChildOrders(t *testing.T) {
	dir := t.TempDir()
	files := map[string]string{
		"index.md":       "# Updates\n\nRelease notes.",
		"2025-01-old.md": "---\ntitle: Old\ndate: 2025-01-01\n---\n# Old\n\nOld.",
		"2026-01-new.md": "---\ntitle: New\ndate: 2026-01-01\n---\n# New\n\nNew.",
	}
	for name, body := range files {
		if err := os.WriteFile(filepath.Join(dir, name), []byte(body), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	r := NewRouter()
	if err := r.MarkdownBlog("/blog", dir, BlogConfig{}); err != nil {
		t.Fatal(err)
	}
	archive := r.routes["/blog/archive"]
	if archive == nil || len(archive.Children) != 2 {
		t.Fatalf("archive children = %#v", archive)
	}
	if archive.Children[0].Order == archive.Children[1].Order {
		t.Fatalf("archive year routes share order %d: %#v", archive.Children[0].Order, archive.Children)
	}
	if err := r.Validate(); err != nil {
		t.Fatal(err)
	}
}

func TestRSSRejectsUnsafeSiteURLAndStaticPath(t *testing.T) {
	r := NewRouter()
	if err := r.MarkdownBlogFS("/blog", fstest.MapFS{"index.md": &fstest.MapFile{Data: []byte("# Blog")}}, ".", BlogConfig{}); err != nil {
		t.Fatal(err)
	}
	if _, err := r.RSSXML(RSSConfig{Prefix: "/blog", SiteURL: "javascript:alert(1)"}); err == nil {
		t.Fatal("RSSXML accepted an unsafe site URL")
	}
	if _, err := r.RSSXML(RSSConfig{Prefix: "/blog", SiteURL: "https://example.com/docs?preview=1"}); err == nil {
		t.Fatal("RSSXML accepted a site URL with a query")
	}
	if err := WriteStaticRSS(t.TempDir(), "", "/../feed.xml", []byte("feed")); err == nil {
		t.Fatal("WriteStaticRSS accepted a traversal path")
	}
	if err := WriteStaticRSS(t.TempDir(), "", `\..\feed.xml`, []byte("feed")); err == nil {
		t.Fatal("WriteStaticRSS accepted a Windows traversal path")
	}
}

func TestMarkdownBlogRegistersPublicationViewsAndUsesBlogTemplate(t *testing.T) {
	dir := t.TempDir()
	files := map[string]string{
		"index.md":             "# Updates\n\nProduct news.",
		"2026-08-01-router.md": "---\ntitle: Router notes\nauthors: [Core team]\ntags: [routing, release]\n---\n# Router notes\n\n## What changed\n\nA useful update.",
		"second.md":            "---\ntitle: Second note\ndate: 2026-07-01\nauthors: Core team\ntags: release\n---\n# Second note\n\n## Details\n\nMore context.",
	}
	for name, body := range files {
		if err := os.WriteFile(filepath.Join(dir, name), []byte(body), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	r := NewRouter(WithSiteName("Updates"))
	if err := r.MarkdownBlog("/blog", dir, BlogConfig{Title: "Updates", Description: "Product news."}); err != nil {
		t.Fatal(err)
	}
	for _, path := range []string{"/blog", "/blog/search", "/blog/archive", "/blog/archive/2026", "/blog/tags", "/blog/tags/release", "/blog/authors", "/blog/authors/core-team"} {
		route := r.routes[path]
		if route == nil || !route.Blog {
			t.Fatalf("publication route %q missing or not marked as blog", path)
		}
	}
	if got := r.routes["/blog/2026-08-01-router"].Metadata.DatePublished; got != "2026-08-01" {
		t.Fatalf("filename date fallback = %q", got)
	}
	post := r.routes["/blog/2026-08-01-router"]
	postHTML := string((&pageComponent{router: r, route: post}).Render())
	for _, marker := range []string{
		`<h1 id="fastr-docs-blog-share-blog-2026-08-01-router-title">Router notes</h1>`,
		`<article `,
		`aria-labelledby="fastr-docs-blog-share-blog-2026-08-01-router-title"`,
		`data-fastr-docs-share=""`,
		`data-fui-copy-text-from="#fastr-docs-blog-share-blog-2026-08-01-router"`,
		`class="fastr-docs-blog-post__related"`,
		`/blog/authors/core-team`,
		`What changed`,
	} {
		if !strings.Contains(postHTML, marker) {
			t.Fatalf("blog post template missing %q: %s", marker, postHTML)
		}
	}
	if strings.Count(postHTML, "<h1 ") != 1 || !strings.Contains(postHTML, "fastr-docs-blog-page--post") {
		t.Fatalf("blog post template is not reading-oriented: %s", postHTML)
	}
	search := r.routes["/blog/search"]
	request := &http.Request{URL: &url.URL{Path: "/blog/search", RawQuery: "q=router"}}
	searchHTML := string((&pageComponent{router: r, route: search}).RenderCtx(uiapp.WithRequest(context.Background(), request)))
	second := strings.Index(searchHTML, `data-fastr-docs-blog-search-item="Second note`)
	if !strings.Contains(searchHTML, "Results for") || !strings.Contains(searchHTML, "Router notes") || second < 0 || !strings.Contains(searchHTML[second:], `hidden=""`) {
		t.Fatalf("blog search did not filter server-side: %s", searchHTML)
	}
}

func TestYearArchivesListOnlyTheirYear(t *testing.T) {
	t.Run("a year archive route exists", func(t *testing.T) {
		r := blogAcrossTwoYears(t)
		if r.routeAtPath("/blog/archive/2025") == nil {
			t.Fatal("no route serves the year archive")
		}
	})
	t.Run("a year archive lists only that year", func(t *testing.T) {
		r := blogAcrossTwoYears(t)
		html := renderRouterPage(t, r, "/blog/archive/2025")
		cards := html[strings.Index(html, "fastr-docs-blog-card"):]
		if !strings.Contains(cards, "Viejo") || strings.Contains(cards, "Nuevo") {
			t.Fatal("the year archive mixes posts from other years")
		}
	})
	t.Run("the year archive carries the year in its title", func(t *testing.T) {
		r := blogAcrossTwoYears(t)
		route := r.routeAtPath("/blog/archive/2025")
		if route == nil || !strings.Contains(route.Title, "2025") {
			t.Fatalf("year route title = %+v", route)
		}
	})
}

func TestUnpublishedPostsStayHidden(t *testing.T) {
	t.Run("future posts stay out of search", func(t *testing.T) {
		r := blogAcrossTwoYears(t)
		route := r.routeAtPath("/blog/newer")
		route.Metadata.DatePublished = "2030-01-01"
		for _, entry := range r.SearchIndex() {
			if entry.Path == "/blog/newer" {
				t.Fatal("a future-dated post is searchable")
			}
		}
	})
	t.Run("future posts stay out of the sitemap", func(t *testing.T) {
		r := blogAcrossTwoYears(t)
		route := r.routeAtPath("/blog/newer")
		route.Metadata.DatePublished = "2030-01-01"
		if strings.Contains(string(r.Sitemap()), "/blog/newer") {
			t.Fatal("a future-dated post is offered to crawlers")
		}
	})
	t.Run("term pages hide future posts", func(t *testing.T) {
		r := blogAcrossTwoYears(t)
		route := r.routeAtPath("/blog/newer")
		route.Metadata.DatePublished = "2030-01-01"
		html := renderRouterPage(t, r, "/blog/tags/go")
		cards := html[strings.Index(html, "fastr-docs-blog"):]
		if strings.Contains(cards, "Nuevo") {
			t.Fatal("a future-dated post appears on a tag page")
		}
	})
}

func TestBlogSearchAndCards(t *testing.T) {
	t.Run("server-side blog search folds accents", func(t *testing.T) {
		post := &Route{Title: "Nuevo", Metadata: ContentMetadata{Tags: []string{"Diseño"}}}
		if !blogQueryMatch(post, "diseno") {
			t.Fatal("a folded query misses an accented tag before JavaScript runs")
		}
	})
	t.Run("blog cards expose their dates to machines", func(t *testing.T) {
		r := blogAcrossTwoYears(t)
		html := renderRouterPage(t, r, "/blog")
		if !strings.Contains(html, "<time") {
			t.Fatal("card dates render as bare text")
		}
	})
}

func TestGeneratedBlogViewsStayOutOfIndexes(t *testing.T) {
	t.Run("generated views stay out of search", func(t *testing.T) {
		r := blogAcrossTwoYears(t)
		for _, entry := range r.SearchIndex() {
			if strings.HasPrefix(entry.Path, "/blog/tags") || strings.HasPrefix(entry.Path, "/blog/authors") || strings.HasPrefix(entry.Path, "/blog/search") {
				t.Fatalf("generated view %q pollutes the search index", entry.Path)
			}
		}
	})
	t.Run("generated views stay out of the manifest", func(t *testing.T) {
		r := blogAcrossTwoYears(t)
		body, err := r.ExportManifestJSON("")
		if err != nil {
			t.Fatal(err)
		}
		var manifest ExportManifest
		if err := json.Unmarshal(body, &manifest); err != nil {
			t.Fatal(err)
		}
		for _, route := range manifest.Routes {
			if strings.HasPrefix(route.Path, "/blog/tags") || strings.HasPrefix(route.Path, "/blog/search") {
				t.Fatalf("generated view %q ships in the manifest", route.Path)
			}
		}
	})
	t.Run("generated views stay out of the sitemap", func(t *testing.T) {
		r := blogAcrossTwoYears(t)
		if strings.Contains(string(r.Sitemap()), "/blog/tags") || strings.Contains(string(r.Sitemap()), "/blog/search") {
			t.Fatal("generated listing pages are offered to crawlers as content")
		}
	})
}

func minInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}
