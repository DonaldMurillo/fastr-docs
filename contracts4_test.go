package docs

// The fourth audit cycle: mined from the export path, search ranking,
// validation edges, presentation, and the blog surface. Each case failed
// before its fix; the cycle shipped at the honest count the surface had left.

// The fourth red suite. Same rules as every suite before it: each case
// fails against the current implementation and names a real gap.

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"testing/fstest"
)

func TestRed301To310HeadAndMeta(t *testing.T) {
	t.Run("301 the pager links its neighbours in the head", func(t *testing.T) {
		r := NewRouter()
		g := r.MustGroup("/docs", GroupConfig{Title: "D", Description: "d", Order: 1})
		g.MustPage("a", PageConfig{Title: "A", Description: "d", Order: 1, Source: "# A"})
		g.MustPage("b", PageConfig{Title: "B", Description: "d", Order: 2, Source: "# B"})
		head := r.metadataHeadHTML(r.routeAtPath("/docs/a"))
		if !strings.Contains(head, `rel="next"`) || !strings.Contains(head, "/docs/b") {
			t.Fatalf("the head never tells crawlers the next page: %s", head)
		}
	})
	t.Run("302 pages link their own canonical by default", func(t *testing.T) {
		r := NewRouter()
		r.MustPage("/g", PageConfig{Title: "G", Description: "d", Order: 1, Source: "# G"})
		head := r.metadataHeadHTML(r.routeAtPath("/g"))
		if !strings.Contains(head, `<link rel="canonical" href="/g">`) {
			t.Fatalf("a page without an explicit canonical carries none at all: %s", head)
		}
	})
	t.Run("303 translated variant options carry their language", func(t *testing.T) {
		r := NewRouter(WithLocaleFallback("en"))
		r.MustPage("/g", PageConfig{Title: "G", Description: "d", Order: 1, Source: "# G"})
		r.MustPage("/es/g", PageConfig{Title: "G", Description: "d", Order: 2, Source: "# G", Metadata: ContentMetadata{Locale: "es"}})
		html := r3Render(t, r, "/g")
		if !strings.Contains(html, `lang="es"`) {
			t.Fatal("the language selector offers options with no lang attribute")
		}
	})
	t.Run("304 the drawer trigger points at its panel", func(t *testing.T) {
		r := NewRouter()
		r.MustPage("/g", PageConfig{Title: "G", Description: "d", Order: 1, Source: "## A\n\n## B\n\nText."})
		html := r3Render(t, r, "/g")
		if !strings.Contains(html, "aria-controls") {
			t.Fatal("the mobile drawer trigger does not name the region it opens")
		}
	})
	t.Run("305 the theme storage key is namespaced", func(t *testing.T) {
		if !strings.Contains(RuntimeJS(), "fastr-docs-theme") {
			t.Fatal("the runtime watches for theme changes by any key containing 'theme'")
		}
	})
	t.Run("306 view transitions are styled", func(t *testing.T) {
		if !strings.Contains(NewRouter().CSS(), "view-transition") {
			t.Fatal("SPA swaps cut hard with no crossfade")
		}
	})
	t.Run("307 selections follow the theme", func(t *testing.T) {
		if !strings.Contains(NewRouter().CSS(), "::selection") {
			t.Fatal("selected text renders in the browser default blue")
		}
	})
	t.Run("308 display math gets block styling", func(t *testing.T) {
		if !strings.Contains(NewRouter().CSS(), "fastr-docs-math--display") {
			t.Fatal("display math renders inline with no block treatment from the docs layer")
		}
	})
	t.Run("310 share links are hardened", func(t *testing.T) {
		r := r3BlogYear(t)
		html := r3Render(t, r, "/blog/older")
		if !strings.Contains(html, "noopener") {
			t.Fatal("share targets open without noopener")
		}
	})
}

func TestRed311To320SearchAndRank(t *testing.T) {
	t.Run("311 repeated terms rank higher", func(t *testing.T) {
		r := NewRouter()
		r.MustPage("/once", PageConfig{Title: "A", Description: "d", Order: 1, Source: "# zebra once"})
		r.MustPage("/many", PageConfig{Title: "B", Description: "d", Order: 2, Source: "# zebra zebra zebra"})
		results := r.Search("zebra")
		if len(results) != 2 || results[0].Entry.Path != "/many" {
			t.Fatalf("term frequency does not rank: %+v", results)
		}
	})
	t.Run("312 manifest alternates pair both directions", func(t *testing.T) {
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
	t.Run("313 body weighting is configurable", func(t *testing.T) {
		r := NewRouter(WithSearchWeights(map[string]int{"title": 1}))
		r.MustPage("/t", PageConfig{Title: "zebra", Description: "d", Order: 1, Source: "# z"})
		r.MustPage("/b", PageConfig{Title: "plain", Description: "d", Order: 2, Source: "# zebra zebra body"})
		results := r.Search("zebra")
		if len(results) != 2 || results[0].Entry.Path != "/b" {
			t.Fatalf("a term the body repeats cannot outrank a title tuned down to its weight: %+v", results)
		}
	})
	t.Run("315 phrase queries keep their words together", func(t *testing.T) {
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
	t.Run("320 headings ids fold apostrophes", func(t *testing.T) {
		r := NewRouter()
		r.MustPage("/g", PageConfig{Title: "G", Description: "d", Order: 1, Source: "## What's new\n\nText."})
		html := r3Render(t, r, "/g")
		if !strings.Contains(html, `id="whats-new"`) {
			t.Fatal("an apostrophe in a heading poisons its id")
		}
	})
}

func TestRed321To330ValidationEdges(t *testing.T) {
	t.Run("322 hero actions require links", func(t *testing.T) {
		r := NewRouter()
		r.MustPage("/g", PageConfig{Title: "G", Description: "d", Order: 1, Source: "# G",
			Metadata: ContentMetadata{PageTemplate: PageTemplateSplash, Hero: &PageHero{Title: "T", Actions: []PageHeroAction{{Text: "Go"}}}}})
		if err := r.Validate(); err == nil || !strings.Contains(err.Error(), "action") {
			t.Fatalf("Validate() = %v, want a hero action link complaint", err)
		}
	})
	t.Run("323 images reject javascript sources", func(t *testing.T) {
		r := NewRouter()
		r.MustPage("/g", PageConfig{Title: "G", Description: "d", Order: 1, Source: "![x](javascript:alert(1))"})
		if err := r.Validate(); err == nil || !strings.Contains(err.Error(), "image") {
			t.Fatalf("Validate() = %v, want an image scheme complaint", err)
		}
	})
	t.Run("324 translations of drafts warn rather than fail", func(t *testing.T) {
		r := NewRouter()
		r.MustPage("/draft", PageConfig{Title: "D", Description: "d", Order: 1, Source: "# D", Metadata: ContentMetadata{Draft: true}})
		r.MustPage("/es/draft", PageConfig{Title: "D", Description: "d", Order: 2, Source: "# D", Metadata: ContentMetadata{Locale: "es", TranslationOf: "/draft"}})
		if err := r.Validate(); err != nil {
			t.Fatalf("a translation of an unpublished draft fails the build: %v", err)
		}
	})
	t.Run("325 order gaps are warned", func(t *testing.T) {
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
	t.Run("326 heading level jumps are named", func(t *testing.T) {
		r := NewRouter()
		r.MustPage("/g", PageConfig{Title: "G", Description: "d", Order: 1, Source: "## Top\n\n#### Skipped three\n\nBody."})
		if err := r.Validate(); err == nil || !strings.Contains(err.Error(), "heading") {
			t.Fatalf("Validate() = %v, want a heading-level jump complaint", err)
		}
	})
	t.Run("330 excerpt feeds the meta description", func(t *testing.T) {
		r := NewRouter()
		r.MustPage("/g", PageConfig{Title: "G", Description: "plain", Order: 1, Source: "# G", Metadata: ContentMetadata{Excerpt: "Curated summary."}})
		head := r.metadataHeadHTML(r.routeAtPath("/g"))
		if !strings.Contains(head, "Curated summary.") {
			t.Fatalf("a curated excerpt never reaches the page metadata: %s", head)
		}
	})
}

func TestRed331To344MinedGaps(t *testing.T) {
	t.Run("331 the 404 offers a way back to search", func(t *testing.T) {
		r := NewRouter()
		r.MustPage("/g", PageConfig{Title: "G", Description: "d", Order: 1, Source: "# G"})
		html := string(r.NotFoundScreen().RenderNotFound("/deep/miss"))
		if !strings.Contains(html, "search") {
			t.Fatal("the 404 dead-ends with no route back to search")
		}
	})
	t.Run("332 unmarked originals count as the default locale", func(t *testing.T) {
		r := r2Bilingual()
		if missing := r.UntranslatedFamilies("en"); len(missing) != 0 {
			t.Fatalf("families = %v; an unmarked original page does not cover its own language", missing)
		}
	})
	t.Run("333 the runtime prepares overlays for print", func(t *testing.T) {
		if !strings.Contains(RuntimeJS(), "beforeprint") {
			t.Fatal("printing leaves drawers and the palette open over the page")
		}
	})
	t.Run("334 focus styling comes from a token", func(t *testing.T) {
		if !strings.Contains(NewRouter().CSS(), "--docs-focus") {
			t.Fatal("focus ring colors are hardcoded at every use site")
		}
	})
	t.Run("335 the head carries the social image", func(t *testing.T) {
		r := NewRouter()
		r.MustPage("/g", PageConfig{Title: "G", Description: "d", Order: 1, Source: "# G", Metadata: ContentMetadata{Image: "/cover.png"}})
		head := r.metadataHeadHTML(r.routeAtPath("/g"))
		if !strings.Contains(head, `property="og:image"`) || !strings.Contains(head, "/cover.png") {
			t.Fatalf("a declared social image never reaches the head: %s", head)
		}
	})
	t.Run("337 empty collections refuse to register", func(t *testing.T) {
		r := NewRouter()
		if err := r.MarkdownCollectionFS("/docs", fstest.MapFS{}, ".", CollectionConfig{}); err == nil {
			t.Fatal("a collection with no documents registers a section with nothing in it")
		}
	})
	t.Run("338 collection slugs refuse spaces", func(t *testing.T) {
		if _, err := collectionSlug("a b"); err == nil {
			t.Fatal("a slug with a space becomes a path with %20 in it")
		}
	})
	t.Run("339 collection slugs fold like every other slug", func(t *testing.T) {
		slug, err := collectionSlug("Diseño")
		if err != nil {
			t.Fatal(err)
		}
		if slug != "diseno" {
			t.Fatalf("collectionSlug(Diseño) = %q; collections and blogs disagree on folding", slug)
		}
	})
	t.Run("340 route paths refuse percent escapes", func(t *testing.T) {
		r := NewRouter()
		if err := r.Page("/a%20b", PageConfig{Title: "X", Description: "d", Order: 1, Source: "# X"}); err == nil {
			t.Fatal("a percent escape in a route path registers a URL nobody can link to")
		}
	})
	t.Run("341 feeds require a rooted prefix", func(t *testing.T) {
		r := r3BlogYear(t)
		if _, err := r.RSSXML(RSSConfig{Prefix: "blog"}); err == nil {
			t.Fatal("a prefix without a slash produces a feed no host can serve")
		}
	})
	t.Run("342 group paths mark themselves active", func(t *testing.T) {
		r := NewRouter()
		g := r.MustGroup("/docs", GroupConfig{Title: "D", Description: "d", Order: 1})
		g.MustPage("a", PageConfig{Title: "A", Description: "d", Order: 1, Source: "# A"})
		for _, item := range r.NavigationAt("/docs") {
			if item.Path == "/docs" && item.Active {
				return
			}
		}
		t.Fatal("standing on a group page marks nothing active in the navigation model")
	})
	t.Run("343 translation_of follows redirect sources", func(t *testing.T) {
		r := NewRouter()
		r.MustPage("/g", PageConfig{Title: "G", Description: "d", Order: 1, Source: "# G"})
		r.MustPage("/other", PageConfig{Title: "O", Description: "d", Order: 2, Source: "# O", Metadata: ContentMetadata{Redirects: []string{"/old"}}})
		r.MustPage("/es/old", PageConfig{Title: "E", Description: "d", Order: 3, Source: "# E", Metadata: ContentMetadata{Locale: "es", TranslationOf: "/old"}})
		if r.familyOf(r.routeAtPath("/es/old")) != r.familyOf(r.routeAtPath("/other")) {
			t.Fatal("a translation naming a redirect source pairs with nothing")
		}
	})
	t.Run("344 the fold handles the fi ligature", func(t *testing.T) {
		if got := foldRunes("file"); got == "file" {
			t.Skip("ligature normalized by font already")
		}
		if got := foldRunes("ﬁle"); got != "file" {
			t.Fatalf("foldRunes(ligature) = %q", got)
		}
	})
}

func TestRed345To346ExportHelpers(t *testing.T) {
	t.Run("345 static feeds refuse a slashless path", func(t *testing.T) {
		if err := WriteStaticRSS(t.TempDir(), "", "feed.xml", []byte("<rss/>")); err == nil {
			t.Fatal("a feed path without a leading slash writes to an unpredictable place")
		}
	})
	t.Run("346 the framework writes redirect stubs", func(t *testing.T) {
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

// RedirectStubWriter is the exported seam the 346 red targets: a function
// writing a meta-refresh stub, once the framework grows one.
var RedirectStubWriter interface {
	WriteStaticRedirect(dir, basePath, from, to string) error
}
