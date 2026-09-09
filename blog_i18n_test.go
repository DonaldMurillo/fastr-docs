package docs

import (
	"context"
	"path/filepath"
	"strings"
	"testing"
)

// A site with a blog per language. The Spanish collection is under the
// Spanish home, which is where a translated tree puts it, and one of its
// posts keeps its file name while the other translates its slug.
func bilingualBlogSite(t *testing.T) *Router {
	t.Helper()
	root := t.TempDir()
	writeFiles(t, root, map[string]string{
		"blog/index.md":               "---\ntitle: Blog\ndescription: Updates.\n---\n# Blog\n",
		"blog/2026-01-02-first.md":    "---\ntitle: First\ndescription: One.\ndate: 2026-01-02\ntags: [news]\n---\n# First\n\nBody.\n",
		"blog/second.md":              "---\ntitle: Second\ndescription: Two.\ndate: 2026-01-03\n---\n# Second\n\nBody.\n",
		"es/blog/index.md":            "---\ntitle: Blog\ndescription: Novedades.\n---\n# Blog\n",
		"es/blog/2026-01-02-first.md": "---\ntitle: Primera\ndescription: Uno.\ndate: 2026-01-02\ntags: [news]\n---\n# Primera\n\nCuerpo.\n",
		"es/blog/segunda.md":          "---\ntitle: Segunda\ndescription: Dos.\ndate: 2026-01-03\ntranslation_of: /blog/second\n---\n# Segunda\n\nCuerpo.\n",
	})

	r := NewRouter(
		WithSiteName("Docs"),
		WithLocaleFallback("en"),
		WithLocaleNames(map[string]string{"en": "English", "es": "Español"}),
		WithLocaleUIStrings("es", UIStrings{
			DateFormat: "2 de January de 2006",
			Months:     []string{"enero", "febrero", "marzo", "abril", "mayo", "junio", "julio", "agosto", "septiembre", "octubre", "noviembre", "diciembre"},
			Published:  "Publicado",
			By:         "por",
			Previous:   "← Anterior",
			Next:       "Siguiente →",
			Blog: BlogStrings{
				AllPosts: "Todas las entradas", Archive: "Archivo", Tags: "Etiquetas",
				Authors: "Autores", Search: "Buscar", LatestPosts: "Últimas entradas",
				ReadingTime: "%s min de lectura", Feed: "RSS",
			},
		}),
	)
	r.MustPage("/", PageConfig{Title: "Docs", Description: "Home", Source: "# Docs\n"})
	r.MustPage("/docs", PageConfig{Title: "Guide", Description: "en", Source: "# Guide\n", Order: 1})
	r.MustPage("/es", PageConfig{Title: "Inicio", Description: "es", Source: "# Inicio\n", Order: 2, Metadata: ContentMetadata{Locale: "es"}})
	if err := r.MarkdownBlog("/blog", filepath.Join(root, "blog"), BlogConfig{Title: "Blog", Description: "Updates.", Order: 3}); err != nil {
		t.Fatal(err)
	}
	if err := r.MarkdownBlog("/es/blog", filepath.Join(root, "es", "blog"), BlogConfig{Title: "Blog", Description: "Novedades.", Order: 3, DefaultLocale: "es"}); err != nil {
		t.Fatal(err)
	}
	return r
}

// Each collection reads the strings of its own language. Reading the
// Router-wide set everywhere left a Spanish blog with "All posts" and
// "Archive" in its sidebar.
func TestBlogLabelsFollowTheCollection(t *testing.T) {
	r := bilingualBlogSite(t)
	if got := r.blogLabels("/es/blog").Archive; got != "Archivo" {
		t.Fatalf("Spanish collection Archive = %q", got)
	}
	if got := r.blogLabels("/blog").Archive; got != "Archive" {
		t.Fatalf("English collection Archive = %q", got)
	}
	// The generated views carry the collection's language, so their chrome
	// reads the right strings and they pair with the English views.
	archive := r.routeAtPath("/es/blog/archive")
	if archive == nil || archive.Metadata.Locale != "es" {
		t.Fatalf("Spanish archive view = %+v", archive)
	}
	if archive.Title != "Archivo" {
		t.Fatalf("Spanish archive title = %q", archive.Title)
	}
	if got := optionHrefs(r.variantOptions(archive, "locale")); got["en"] != "/blog/archive" {
		t.Fatalf("archive selector = %v", got)
	}

	sidebar := string((&blogSidebar{router: r, prefix: "/es/blog"}).Render())
	for _, want := range []string{"Todas las entradas", "Archivo", "Etiquetas", "Autores", "Buscar"} {
		if !strings.Contains(sidebar, want) {
			t.Fatalf("Spanish blog sidebar lacks %q: %s", want, sidebar)
		}
	}
	if strings.Contains(sidebar, "All posts") {
		t.Fatalf("Spanish blog sidebar carries the English label: %s", sidebar)
	}

	landing := string(r.renderBlogArchive("/es/blog", 1, ""))
	if !strings.Contains(landing, "Últimas entradas") || !strings.Contains(landing, "min de lectura") {
		t.Fatalf("Spanish landing is not in Spanish: %s", landing)
	}
	// Dates follow the language's own format, month names included: Go
	// prints them in English only.
	if !strings.Contains(landing, "2 de enero de 2026") {
		t.Fatalf("Spanish landing did not use the Spanish date format: %s", landing)
	}
	// The post's own meta line reads the collection's language too. It read
	// the Router-wide strings, so "Published" and "By" stayed English over
	// a Spanish post.
	post := string((&pageComponent{router: r, route: r.routeAtPath("/es/blog/segunda")}).Render())
	if !strings.Contains(post, `Publicado <time datetime="2026-01-03">3 de enero de 2026</time>`) || strings.Contains(post, "Published") {
		t.Fatalf("Spanish post meta is not in Spanish: %s", post)
	}
	english := string(r.renderBlogArchive("/blog", 1, ""))
	if !strings.Contains(english, "Latest posts") || strings.Contains(english, "Últimas") {
		t.Fatalf("English landing was affected: %s", english)
	}
}

// A post that keeps its file name pairs by path; one with a translated slug
// pairs through translation_of. Both are the same blog to the reader.
func TestBlogPostsPairAcrossCollections(t *testing.T) {
	r := bilingualBlogSite(t)
	first := r.routeAtPath("/blog/2026-01-02-first")
	if got := optionHrefs(r.variantOptions(first, "locale")); got["es"] != "/es/blog/2026-01-02-first" {
		t.Fatalf("same-name post selector = %v", got)
	}
	second := r.routeAtPath("/blog/second")
	if got := optionHrefs(r.variantOptions(second, "locale")); got["es"] != "/es/blog/segunda" {
		t.Fatalf("translated-slug post selector = %v", got)
	}
	if got := r.alternatesFor(r.routeAtPath("/es/blog/segunda")); got["en"] != "/blog/second" {
		t.Fatalf("translated-slug post alternates = %v", got)
	}
	// The pager on a Spanish post stays in the Spanish collection and reads
	// its labels from it: the neighbour, the "all posts" fallback on the
	// oldest post, and the direction lines.
	pager := r.blogPager(r.routeAtPath("/es/blog/segunda"))
	if pager == nil || pager.PrevHref != "/es/blog/2026-01-02-first" || pager.PrevDirLabel != "← Anterior" || pager.NextDirLabel != "Siguiente →" {
		t.Fatalf("Spanish post pager = %+v", pager)
	}
	oldest := r.blogPager(r.routeAtPath("/es/blog/2026-01-02-first"))
	if oldest == nil || oldest.PrevHref != "/es/blog" || oldest.PrevLabel != "Todas las entradas" {
		t.Fatalf("oldest Spanish post pager = %+v", oldest)
	}
}

// GoFastr re-renders a layout layer on client-side navigation only when its
// key changes, so a multilingual site keys its section layouts by language.
// A single-language site keeps its names, and the CSS classes derived from
// them.
func TestSectionLayoutsAreKeyedByLanguageOnAMultilingualSite(t *testing.T) {
	r := bilingualBlogSite(t)
	if got := r.sectionLayout(r.routeAtPath("/es/blog")).Name; got != "blog-es" {
		t.Fatalf("Spanish blog layout = %q", got)
	}
	if got := r.sectionLayout(r.routeAtPath("/blog")).Name; got != "blog-en" {
		t.Fatalf("English blog layout = %q", got)
	}
	if got := r.sectionLayout(r.routeAtPath("/docs")).Name; got != "docs-section-en" {
		t.Fatalf("docs section layout = %q", got)
	}

	single := NewRouter(WithSiteName("Docs"))
	single.MustPage("/docs", PageConfig{Title: "Guide", Description: "en", Source: "# Guide\n"})
	if got := single.sectionLayout(single.routeAtPath("/docs")).Name; got != "docs-section" {
		t.Fatalf("single-language layout = %q, want the plain name", got)
	}
}

// A blog under a locale home is still grouped as a blog, so it wears the
// blog sidebar rather than the docs one it would inherit from its tree root.
func TestABlogUnderTheLocaleHomeKeepsTheBlogLayout(t *testing.T) {
	r := bilingualBlogSite(t)
	post := r.routeAtPath("/es/blog/segunda")
	if root := r.blogRootFor(post); root == nil || root.Path != "/es/blog" {
		t.Fatalf("blogRootFor(segunda) = %+v", root)
	}
	if root := r.blogRootFor(r.routeAtPath("/es")); root != nil {
		t.Fatalf("the locale home is not a blog: %+v", root)
	}
}

// The header lives in the layout layer GoFastr keeps across client-side
// navigations, so every page carries its own header in a template inside
// the swapped region, and the runtime copies the parts that change.
func TestEveryPageCarriesItsChromeTemplate(t *testing.T) {
	r := bilingualBlogSite(t)
	page := &pageComponent{router: r, route: r.routeAtPath("/es/blog/segunda")}
	html := string(page.RenderCtx(context.Background()))
	if !strings.Contains(html, `<template data-fastr-docs-chrome="" data-fastr-docs-dir="ltr" data-fastr-docs-lang="es" data-fastr-docs-skip="Skip to main content">`) {
		t.Fatalf("Spanish page lacks its chrome template: %s", html)
	}
	// The template is the page's own header: its tabs point into the
	// Spanish tree and its selector at this page's translation.
	template := html[strings.Index(html, "<template data-fastr-docs-chrome"):]
	for _, want := range []string{`href="/es/blog"`, `value="/blog/second"`, `data-fastr-docs-locale="es"`} {
		if !strings.Contains(template, want) {
			t.Fatalf("chrome template lacks %q: %s", want, template)
		}
	}
	english := string((&pageComponent{router: r, route: r.routeAtPath("/docs")}).RenderCtx(context.Background()))
	if !strings.Contains(english, `data-fastr-docs-lang="en"`) {
		t.Fatalf("English page template lang: %s", english)
	}
	// The selector container is rendered even when empty, so the runtime has
	// somewhere to put a page's selectors after a navigation from a page
	// that had none.
	if !strings.Contains(english, `class="fastr-docs-variant-selectors"`) {
		t.Fatalf("English page template lacks the selector container: %s", english)
	}
}
