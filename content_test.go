package docs

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"testing/fstest"

	"github.com/DonaldMurillo/gofastr/core/render"
)

func TestParseMarkdownFrontMatterSupportsAliasesAndStripsHeader(t *testing.T) {
	document, err := ParseMarkdown("---\n" +
		"title: Routing\n" +
		"slug: guide\n" +
		"description: Register pages\n" +
		"draft: true\n" +
		"no_index: true\n" +
		"tags: [router, content]\n" +
		"redirects: [/guide, /old-guide]\n" +
		"---\n\n## Start\n\nBody")
	if err != nil {
		t.Fatalf("ParseMarkdown() error = %v", err)
	}
	if document.Metadata.Title != "Routing" || document.Metadata.Description != "Register pages" {
		t.Fatalf("metadata = %#v", document.Metadata)
	}
	if document.Metadata.Slug != "guide" {
		t.Fatalf("slug metadata = %q", document.Metadata.Slug)
	}
	if !document.Metadata.Draft || !document.Metadata.NoIndex {
		t.Fatalf("boolean metadata = %#v", document.Metadata)
	}
	if strings.Contains(document.Body, "title: Routing") || !strings.Contains(document.Body, "## Start") {
		t.Fatalf("front matter was not stripped from body: %q", document.Body)
	}
	if len(document.Metadata.Tags) != 2 || len(document.Metadata.Redirects) != 2 {
		t.Fatalf("list metadata = %#v", document.Metadata)
	}
}

func TestMarkdownCollectionRegistersMetadataAndFiltersDraftsLocaleAndVersion(t *testing.T) {
	dir := t.TempDir()
	write := func(name, body string) {
		t.Helper()
		path := filepath.Join(dir, name)
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(path, []byte(body), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	write("index.md", "---\ntitle: Docs\norder: 1\nlocale: en\nversion: v2\n---\n# Docs\n\nOverview")
	write("guide.md", "---\ntitle: Guide\ntags: [guide]\nedit_url: https://github.com/acme/docs/edit/main/guide.md\nredirects: /getting-started\n---\n# Guide\n\nRead this.")
	write("draft.md", "---\ntitle: Draft\ndraft: true\n---\n# Draft\n\nNot public.")
	write("fr/guide.md", "---\ntitle: French\nlocale: fr\n---\n# French\n\nBonjour.")

	r := NewRouter(WithLocale("en"), WithVersion("v2"))
	if err := r.MarkdownCollection("/docs", dir, CollectionConfig{DefaultLocale: "en", DefaultVersion: "v2"}); err != nil {
		t.Fatalf("MarkdownCollection() error = %v", err)
	}
	if err := r.Validate(); err != nil {
		t.Fatalf("Validate() error = %v", err)
	}
	if got := len(r.Routes()); got != 4 {
		t.Fatalf("all collection routes = %d, want 4", got)
	}
	published := r.PublishedRoutes()
	if len(published) != 2 || published[0].Path != "/docs" || published[1].Path != "/docs/guide" {
		t.Fatalf("published routes = %#v", published)
	}
	if got := published[1].Metadata.EditURL; got == "" {
		t.Fatal("front matter edit URL did not reach route metadata")
	}
	if got := published[1].Metadata.Redirects[0]; got != "/getting-started" {
		t.Fatalf("redirect metadata = %q", got)
	}
	if strings.Contains(r.SearchIndex()[1].Text, "title: Guide") {
		t.Fatal("search text still contains front matter")
	}

	preview := NewRouter(WithIncludeDrafts(true), WithLocale("en"), WithVersion("v2"))
	if err := preview.MarkdownCollection("/docs", dir, CollectionConfig{IncludeDrafts: true, DefaultLocale: "en", DefaultVersion: "v2"}); err != nil {
		t.Fatalf("preview MarkdownCollection() error = %v", err)
	}
	if got := len(preview.PublishedRoutes()); got != 3 {
		t.Fatalf("preview published routes = %d, want 3", got)
	}
	prefixed := NewRouter(WithLocale("fr"), WithVersion("v2"))
	if err := prefixed.MarkdownCollection("/docs", dir, CollectionConfig{LocalePrefix: true, VersionPrefix: true, DefaultLocale: "en", DefaultVersion: "v2"}); err != nil {
		t.Fatalf("prefixed MarkdownCollection() error = %v", err)
	}
	var foundFrench bool
	for _, route := range prefixed.PublishedRoutes() {
		if route.Metadata.Locale == "fr" {
			foundFrench = route.Path == "/docs/fr/guide"
		}
	}
	if !foundFrench {
		t.Fatalf("locale-prefixed routes = %#v", prefixed.PublishedRoutes())
	}
}

func TestMarkdownCollectionFSSupportsVirtualContentAndCollectionDrafts(t *testing.T) {
	content := fstest.MapFS{
		"docs/index.md":    &fstest.MapFile{Data: []byte("# Embedded docs\n\nStart here.")},
		"docs/guide.md":    &fstest.MapFile{Data: []byte("---\ntitle: Embedded guide\norder: 2\n---\n# Guide\n\nRead this.")},
		"docs/draft.md":    &fstest.MapFile{Data: []byte("---\ntitle: Embedded draft\ndraft: true\n---\n# Draft\n\nPreview only.")},
		"docs/_partial.md": &fstest.MapFile{Data: []byte("# Hidden\n\nNot a route.")},
	}
	r := NewRouter()
	if err := r.MarkdownCollectionFS("/docs", content, "docs", CollectionConfig{IncludeDrafts: true}); err != nil {
		t.Fatalf("MarkdownCollectionFS() error = %v", err)
	}
	if err := r.Validate(); err != nil {
		t.Fatalf("Validate() error = %v", err)
	}
	if got := len(r.PublishedRoutes()); got != 3 {
		t.Fatalf("published virtual routes = %d, want 3", got)
	}
	var guide *Route
	for _, route := range r.Routes() {
		if route.Path == "/docs/guide" {
			guide = route
			break
		}
	}
	if guide == nil || guide.Title != "Embedded guide" {
		t.Fatalf("virtual guide route = %#v", guide)
	}
	if source := pageSource(guide.page); strings.Contains(source, "title: Embedded guide") {
		t.Fatal("virtual collection body retained front matter")
	}
}

func TestCollectionDraftVisibilityDoesNotLeakAcrossCollections(t *testing.T) {
	content := fstest.MapFS{
		"public.md": &fstest.MapFile{Data: []byte("---\ntitle: Public\n---\n# Public")},
		"draft.md":  &fstest.MapFile{Data: []byte("---\ntitle: Draft\ndraft: true\n---\n# Draft")},
	}
	r := NewRouter()
	if err := r.MarkdownCollectionFS("/preview", content, ".", CollectionConfig{IncludeDrafts: true}); err != nil {
		t.Fatal(err)
	}
	if err := r.MarkdownCollectionFS("/public", content, ".", CollectionConfig{}); err != nil {
		t.Fatal(err)
	}
	if r.routes["/preview/draft"] == nil || !r.routePublished(r.routes["/preview/draft"]) {
		t.Fatal("collection draft was not visible in its opted-in collection")
	}
	if r.routes["/public/draft"] == nil || r.routePublished(r.routes["/public/draft"]) {
		t.Fatal("collection draft visibility leaked into another collection")
	}
}

func TestMarkdownCollectionSupportsCustomSlugs(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "reference.md"), []byte("---\ntitle: Routing reference\nslug: guides/router\n---\n# Routing reference\n\nStart here."), 0o644); err != nil {
		t.Fatal(err)
	}
	r := NewRouter()
	if err := r.MarkdownCollection("/docs", dir, CollectionConfig{}); err != nil {
		t.Fatal(err)
	}
	routes := r.PublishedRoutes()
	if len(routes) != 1 || routes[0].Path != "/docs/guides/router" || routes[0].Metadata.Slug != "guides/router" {
		t.Fatalf("custom slug route = %#v", routes)
	}
}

func TestMarkdownCollectionRejectsUnsafeSlugs(t *testing.T) {
	for _, slug := range []string{"../outside", "guides/../router", "guide?tab=one", ".hidden"} {
		t.Run(slug, func(t *testing.T) {
			dir := t.TempDir()
			body := "---\ntitle: Invalid\nslug: " + slug + "\n---\n# Invalid"
			if err := os.WriteFile(filepath.Join(dir, "page.md"), []byte(body), 0o644); err != nil {
				t.Fatal(err)
			}
			if err := NewRouter().MarkdownCollection("/docs", dir, CollectionConfig{}); err == nil {
				t.Fatalf("slug %q was accepted", slug)
			}
		})
	}
}

func TestMarkdownCollectionPluginAcceptsVirtualContent(t *testing.T) {
	r := NewRouter()
	plugin := MarkdownCollectionPlugin{
		Path: "/handbook",
		FS: fstest.MapFS{
			"pages/index.md": &fstest.MapFile{Data: []byte("---\ntitle: Handbook\n---\n# Handbook\n\nStart here.")},
		},
		Root:   "pages",
		Config: CollectionConfig{OrderStart: 3},
	}
	if err := r.Use(plugin); err != nil {
		t.Fatalf("Use(MarkdownCollectionPlugin) error = %v", err)
	}
	if err := r.Validate(); err != nil {
		t.Fatalf("Validate() error = %v", err)
	}
	routes := r.PublishedRoutes()
	if len(routes) != 1 || routes[0].Path != "/handbook" || routes[0].Order != 3 {
		t.Fatalf("virtual plugin routes = %#v", routes)
	}
}

func TestPageMetadataRendersSEOAndEditLink(t *testing.T) {
	r := NewRouter()
	r.MustPage("/guide", PageConfig{
		Source: "---\ntitle: Metadata guide\ndescription: A metadata guide\nnoindex: true\ncanonical: https://docs.example/guide\nedit_url: https://github.com/acme/docs/edit/main/guide.md\n---\n# Guide\n\n## Setup",
		Order:  1,
	})
	if err := r.Validate(); err != nil {
		t.Fatalf("Validate() error = %v", err)
	}
	page := r.Routes()[0]
	if page.Title != "Metadata guide" || page.Description != "A metadata guide" {
		t.Fatalf("route metadata fallback = %#v", page)
	}
	component := &pageComponent{router: r, route: page}
	head := component.HeadHTML()
	for _, marker := range []string{"noindex,nofollow", "canonical", "https://docs.example/guide"} {
		if !strings.Contains(head, marker) {
			t.Fatalf("HeadHTML() missing %q: %s", marker, head)
		}
	}
	rendered := string(component.Render())
	if !strings.Contains(rendered, "Edit this page") {
		t.Fatalf("rendered page missing edit link: %s", rendered)
	}
}

func TestMarkdownComponentsRenderPropsAndNestedMarkdown(t *testing.T) {
	r := NewRouter()
	r.MustPage("/guide", PageConfig{
		Title:       "Guide",
		Description: "A component guide",
		Order:       1,
		Source:      "# Guide\n\n{{< callout variant='warning' title=\"Important\" >}}\n\n**Keep order explicit.**\n\n{{< /callout >}}",
		Components: map[string]MarkdownComponent{
			"callout": func(props map[string]string, body render.HTML) render.HTML {
				return render.Tag("aside", map[string]string{
					"class":            "docs-callout docs-callout--" + props["variant"],
					"data-callout":     props["title"],
					"data-test-marker": "shortcode",
				}, body)
			},
		},
	})
	if err := r.Validate(); err != nil {
		t.Fatalf("Validate() error = %v", err)
	}
	route := r.Routes()[0]
	html := string((&pageComponent{router: r, route: route}).Render())
	for _, marker := range []string{"data-test-marker=\"shortcode\"", "data-callout=\"Important\"", "docs-callout--warning", "<strong>Keep order explicit.</strong>"} {
		if !strings.Contains(html, marker) {
			t.Fatalf("rendered Markdown component missing %q: %s", marker, html)
		}
	}
	if !strings.Contains(html, `data-docs-route="/guide"`) {
		t.Fatalf("Markdown component render dropped page attributes: %s", html)
	}
	if strings.Contains(html, "<p><aside") || strings.Contains(html, "</aside></p>") {
		t.Fatalf("Markdown component produced invalid paragraph nesting: %s", html)
	}
}

func TestMarkdownComponentsRejectUnknownOrUnclosedShortcodes(t *testing.T) {
	for name, source := range map[string]string{
		"unknown":  "# Guide\n\n{{< missing />}}",
		"unclosed": "# Guide\n\n{{< callout >}}body",
	} {
		r := NewRouter()
		_, err := r.Group("/", GroupConfig{Title: "Home", Description: "Home", Order: 1})
		if err != nil {
			t.Fatalf("Group() error = %v", err)
		}
		err = r.Page("/guide", PageConfig{
			Title: "Guide", Description: "A guide", Order: 2, Source: source,
			Components: map[string]MarkdownComponent{
				"callout": func(_ map[string]string, body render.HTML) render.HTML { return body },
			},
		})
		if err == nil {
			t.Fatalf("%s shortcode should be rejected", name)
		}
	}
}

func TestMarkdownComponentsPluginRegistersGlobalVocabulary(t *testing.T) {
	r := NewRouter()
	if err := r.Use(MarkdownComponentsPlugin{Components: map[string]MarkdownComponent{
		"note": func(_ map[string]string, body render.HTML) render.HTML {
			return render.Tag("aside", map[string]string{"data-global-component": "note"}, body)
		},
	}}); err != nil {
		t.Fatalf("Use(MarkdownComponentsPlugin) error = %v", err)
	}
	r.MustPage("/guide", PageConfig{
		Title: "Guide", Description: "A guide", Order: 1,
		Source: "# Guide\n\n{{< note >}}Shared vocabulary{{< /note >}}",
	})
	page := r.Routes()[0]
	if !strings.Contains(string((&pageComponent{router: r, route: page}).Render()), `data-global-component="note"`) {
		t.Fatal("global Markdown component did not render")
	}
}

func TestContentIssuesFindBrokenRoutesAndAnchorsWithSourceLocations(t *testing.T) {
	r := NewRouter()
	r.MustPage("/docs", PageConfig{
		Title: "Docs", Description: "Docs", Order: 1,
		Source: "# Docs\n\n[Guide](/guide#setup)\n\n[Missing](/missing)\n\n[Bad anchor](/guide#nope)",
	})
	r.MustPage("/guide", PageConfig{
		Title: "Guide", Description: "Guide", Order: 2,
		Source: "# Guide\n\n## Setup",
	})

	issues := r.ContentIssues()
	if len(issues) != 2 {
		t.Fatalf("ContentIssues() = %#v, want missing route and anchor", issues)
	}
	if issues[0].Line != 5 || issues[1].Line != 7 {
		t.Fatalf("issue locations = %#v, want lines 5 and 7", issues)
	}
	if !strings.Contains(issues[0].Message, "unpublished route") || !strings.Contains(issues[1].Message, "does not exist") {
		t.Fatalf("issue messages = %#v", issues)
	}
	if err := r.Validate(); err == nil || !strings.Contains(err.Error(), "content /docs:5") {
		t.Fatalf("Validate() error = %v, want source-level content diagnostics", err)
	}
}

func TestContentIssuesResolveRelativeLinksAndIgnoreExternalLinksAndFences(t *testing.T) {
	r := NewRouter()
	r.MustPage("/docs/index", PageConfig{
		Title: "Docs", Description: "Docs", Order: 1,
		Source: "# Docs\n\n[Guide](guide)\n\n[External](https://example.com/docs)\n\n```md\n[Ignored](/missing)\n```",
	})
	r.MustPage("/docs/guide", PageConfig{
		Title: "Guide", Description: "Guide", Order: 2,
		Source: "# Guide\n\n## Setup",
	})
	if issues := r.ContentIssues(); len(issues) != 0 {
		t.Fatalf("ContentIssues() = %#v, want relative and fenced links to pass", issues)
	}
}

func TestContentIssuesRejectUnsafeLinkSchemes(t *testing.T) {
	r := NewRouter()
	r.MustPage("/docs", PageConfig{
		Title: "Docs", Description: "Docs", Order: 1,
		Source: "# Docs\n\n[Unsafe](javascript:alert(1))",
	})
	issues := r.ContentIssues()
	if len(issues) != 1 || !strings.Contains(issues[0].Message, "unsupported link scheme") {
		t.Fatalf("unsafe link issues = %#v", issues)
	}
}
