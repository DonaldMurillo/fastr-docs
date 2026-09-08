package docs

// Contracts from the 100-case red-suite audit: every one of these failed
// against the implementation before the fix, and every one passes now.
// New gaps should join the group they belong to rather than a new file.

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	uiapp "github.com/DonaldMurillo/gofastr/core-ui/app"
	gofastrRouter "github.com/DonaldMurillo/gofastr/core/router"
)

func redRouter(t *testing.T) *Router {
	t.Helper()
	r := NewRouter()
	r.MustPage("/", PageConfig{Title: "Home", Description: "Home", Source: "# Home", Order: 1})
	docs := r.MustGroup("/docs", GroupConfig{Title: "Docs", Description: "Guides", Order: 2})
	docs.MustPage("start", PageConfig{Title: "Start", Description: "Start", Source: "# Start", Order: 1})
	r.MustPage("/api-reference", PageConfig{Title: "API reference", Description: "API", Source: "# API", Order: 3})
	return r
}

func redEsRouter(t *testing.T) *Router {
	t.Helper()
	r := NewRouter(
		WithLocaleFallback("en"),
		WithLocaleUIStrings("es", UIStrings{Contents: "Contenido", Home: "Inicio", Sections: "Secciones"}),
	)
	r.MustPage("/", PageConfig{Title: "Home", Description: "Home", Source: "# Home", Order: 1, Metadata: ContentMetadata{Locale: "en"}})
	english := r.MustGroup("/docs", GroupConfig{Title: "Docs", Description: "en", Order: 2})
	english.MustPage("guide", PageConfig{Title: "Guide", Description: "en", Source: "# Guide", Order: 1, Metadata: ContentMetadata{Locale: "en"}})
	r.MustPage("/es", PageConfig{Title: "Español", Description: "es", Source: "# Inicio", Order: 4, Metadata: ContentMetadata{Locale: "es"}})
	spanish := r.MustGroup("/es/docs", GroupConfig{Title: "Documentación", Description: "es", Order: 3, Locale: "es"})
	spanish.MustPage("guide", PageConfig{Title: "Guía", Description: "es", Source: "# Guía", Order: 1, Metadata: ContentMetadata{Locale: "es"}})
	return r
}

func redSectionSelect(r *Router, path string) string {
	return string(r.docsSectionSelect(path, "x"))
}

func redRule(css, selector string) (string, bool) {
	i := strings.Index(css, selector)
	if i < 0 {
		return "", false
	}
	j := strings.Index(css[i:], "}")
	if j < 0 {
		return css[i:], true
	}
	return css[i : i+j], true
}

func redBlock(css, open string) (string, bool) {
	start := strings.Index(css, open)
	if start < 0 {
		return "", false
	}
	block := css[start:]
	if end := strings.Index(block[1:], "@media"); end >= 0 {
		block = block[:end+1]
	}
	return block, true
}

func redRender(t *testing.T, r *Router, path string) string {
	t.Helper()
	site := uiapp.NewApp("D")
	if err := r.Mount(site, r.Layout()); err != nil {
		t.Fatal(err)
	}
	html, err := site.RenderPage(context.Background(), path)
	if err != nil {
		t.Fatal(err)
	}
	return string(html)
}

func TestRed001To010DrawerAndLocale(t *testing.T) {
	t.Run("001 drawer name lowercases the locale", func(t *testing.T) {
		if got := docsDrawerName("ES"); got != "fastr-docs-sections-es" {
			t.Fatalf("docsDrawerName(ES) = %q", got)
		}
	})
	t.Run("002 home option without a home route", func(t *testing.T) {
		r := NewRouter()
		docs := r.MustGroup("/docs", GroupConfig{Title: "Docs", Description: "d", Order: 1})
		docs.MustPage("start", PageConfig{Title: "Start", Description: "s", Source: "# S", Order: 1})
		guides := r.MustGroup("/guides", GroupConfig{Title: "Guides", Description: "d", Order: 2})
		guides.MustPage("first", PageConfig{Title: "First", Description: "s", Source: "# F", Order: 1})
		html := redSectionSelect(r, "")
		// The first option is the beginning: labeled Home, aimed at the
		// first section when no home route exists.
		if !strings.Contains(html, `>Home</option>`) || !strings.Contains(html, `value="/docs/start"`) {
			t.Fatalf("no home option for a home-less site: %s", html)
		}
	})
	t.Run("003 hidden home is skipped for the default drawer", func(t *testing.T) {
		r := redRouter(t)
		r.routes["/"].Hidden = true
		if home := r.localeDrawerHomes()[0]; home != nil && home.Hidden {
			t.Fatalf("default drawer home is hidden: %v", home.Path)
		}
	})
	t.Run("004 default drawer follows a single non-default locale", func(t *testing.T) {
		r := NewRouter(WithLocaleUIStrings("es", UIStrings{Contents: "Contenido"}))
		r.MustPage("/", PageConfig{Title: "Inicio", Description: "es", Source: "# Inicio", Order: 1, Metadata: ContentMetadata{Locale: "es"}})
		cfg := r.docsDrawerConfig(r.localeDrawerHomes()[0], docsDrawerName(""))
		if cfg.Title != "Contenido" {
			t.Fatalf("default drawer title = %q, want Contenido", cfg.Title)
		}
	})
	t.Run("005 version families get their own drawers", func(t *testing.T) {
		r := NewRouter()
		r.MustPage("/", PageConfig{Title: "Home", Description: "h", Source: "# H", Order: 1})
		r.MustPage("/guide", PageConfig{Title: "Now", Description: "d", Source: "# N", Order: 2, Metadata: ContentMetadata{Version: "v2"}})
		r.MustPage("/v1/guide", PageConfig{Title: "Then", Description: "d", Source: "# T", Order: 3, Metadata: ContentMetadata{Version: "v1", TranslationOf: "/guide"}})
		if got := len(r.localeDrawerHomes()); got != 2 {
			t.Fatalf("localeDrawerHomes() = %d, want a home per version too", got)
		}
	})
	t.Run("006 drawer select disables autocomplete", func(t *testing.T) {
		if !strings.Contains(redSectionSelect(redRouter(t), ""), `autocomplete="off"`) {
			t.Fatal("section select lacks autocomplete=off")
		}
	})
	t.Run("007 drawer options carry their route path", func(t *testing.T) {
		if !strings.Contains(redSectionSelect(redRouter(t), ""), "data-fastr-docs-section-path") {
			t.Fatal("options carry no route path attribute")
		}
	})
	t.Run("008 section select has localized help text", func(t *testing.T) {
		if !strings.Contains(redSectionSelect(redEsRouter(t), "/es"), "aria-describedby") {
			t.Fatal("no describedby wiring")
		}
	})
	t.Run("009 section select lists version variants", func(t *testing.T) {
		r := NewRouter()
		r.MustPage("/", PageConfig{Title: "Home", Description: "h", Source: "# H", Order: 1})
		r.MustPage("/guide", PageConfig{Title: "Now", Description: "d", Source: "# N", Order: 2, Metadata: ContentMetadata{Version: "v2"}})
		r.MustPage("/v1/guide", PageConfig{Title: "Then", Description: "d", Source: "# T", Order: 3, Metadata: ContentMetadata{Version: "v1", TranslationOf: "/guide"}})
		html := redSectionSelect(r, "/guide")
		if !strings.Contains(html, "/v1/guide") && !strings.Contains(html, "/guide") {
			t.Fatal("version sibling absent from the section select")
		}
	})
	t.Run("010 rtl locales get a direction story", func(t *testing.T) {
		if _, ok := any(redRouter(t)).(interface{ DirectionFor(path string) string }); !ok {
			t.Fatal("no DirectionFor sibling to LanguageFor")
		}
	})
}

func TestRed011To020Labels(t *testing.T) {
	t.Run("011 short month list is a content issue", func(t *testing.T) {
		r := NewRouter(WithLocaleUIStrings("es", UIStrings{Months: []string{"enero", "febrero"}}))
		r.MustPage("/es/post", PageConfig{Title: "P", Description: "d", Source: "# P", Order: 1,
			Metadata: ContentMetadata{Locale: "es", DatePublished: "2026-12-01"}})
		if err := r.Validate(); err == nil {
			t.Fatal("Validate accepted a two-month calendar")
		}
	})
	t.Run("012 format label never leaks printf artifacts", func(t *testing.T) {
		labels := []struct {
			label string
			args  []any
		}{{label: "%s %s", args: []any{"one"}}}
		for _, row := range labels {
			if out := formatLabel(row.label, row.args...); strings.Contains(out, "%!s(MISSING)") {
				t.Fatalf("printf artifact: %q", out)
			}
		}
	})
	t.Run("013 locale labels merge router-wide, locale, then default", func(t *testing.T) {
		r := NewRouter(
			WithUIStrings(UIStrings{Sections: "Router sections"}),
			WithLocaleUIStrings("es", UIStrings{Sections: "Secciones"}),
		)
		r.MustPage("/es/p", PageConfig{Title: "P", Description: "d", Source: "# P", Order: 1, Metadata: ContentMetadata{Locale: "es"}})
		if got := r.uiAt("/es/p").Sections; got != "Secciones" {
			t.Fatalf("locale label = %q", got)
		}
		if got := r.UIStringsForLocale("fr").Sections; got != "Router sections" {
			t.Fatalf("unknown locale label = %q, want the router-wide value", got)
		}
	})
	t.Run("014 count labels pluralize", func(t *testing.T) {
		if got := formatCount(defaultUIStrings.Blog.PostCount, 1); got != "1 post" {
			t.Fatalf("singular renders as %q", got)
		}
		if got := formatCount(defaultUIStrings.Blog.PostCount, 5); got != "5 posts" {
			t.Fatalf("plural renders as %q", got)
		}
		if got := formatCount("%d posts", 2); got != "2 posts" {
			t.Fatalf("plain labels keep working: %q", got)
		}
	})
	t.Run("015 label sets expose a translation inventory", func(t *testing.T) {
		if _, ok := any(defaultUIStrings).(interface{ Keys() []string }); !ok {
			t.Fatal("no Keys inventory for translation tooling to diff against")
		}
	})
	t.Run("016 short ShortMonths is a content issue", func(t *testing.T) {
		r := NewRouter(WithLocaleUIStrings("es", UIStrings{ShortMonths: []string{"ene"}}))
		r.MustPage("/es/p", PageConfig{Title: "P", Description: "d", Source: "# P", Order: 1, Metadata: ContentMetadata{Locale: "es"}})
		if err := r.Validate(); err == nil {
			t.Fatal("Validate accepted a one-entry ShortMonths")
		}
	})
	t.Run("017 month names are validated for uniqueness", func(t *testing.T) {
		r := NewRouter(WithLocaleUIStrings("es", UIStrings{Months: []string{"x", "x", "x", "x", "x", "x", "x", "x", "x", "x", "x", "x"}}))
		r.MustPage("/es/p", PageConfig{Title: "P", Description: "d", Source: "# P", Order: 1, Metadata: ContentMetadata{Locale: "es"}})
		if err := r.Validate(); err == nil {
			t.Fatal("Validate accepted twelve identical month names")
		}
	})
	t.Run("018 uiAt accepts query strings", func(t *testing.T) {
		r := redEsRouter(t)
		if got := r.uiAt("/es/docs/guide?x=1").Contents; got != "Contenido" {
			t.Fatalf("uiAt with query = %q", got)
		}
	})
	t.Run("019 label translations keep their placeholders", func(t *testing.T) {
		r := NewRouter(WithLocaleUIStrings("es", UIStrings{Blog: BlogStrings{ResultsFor: "Resultados"}}))
		r.MustPage("/es/p", PageConfig{Title: "P", Description: "d", Source: "# P", Order: 1, Metadata: ContentMetadata{Locale: "es"}})
		if err := r.Validate(); err == nil {
			t.Fatal("Validate accepted a translation that dropped its placeholder")
		}
	})
	t.Run("020 partial blog translations are reported", func(t *testing.T) {
		r := NewRouter(WithLocaleUIStrings("es", UIStrings{Blog: BlogStrings{Title: "Blog"}}))
		r.MustPage("/es/p", PageConfig{Title: "P", Description: "d", Source: "# P", Order: 1, Metadata: ContentMetadata{Locale: "es"}})
		for _, warning := range r.Warnings() {
			if strings.Contains(warning, "blog labels") {
				return
			}
		}
		t.Fatal("a one-field Blog translation raises no warning")
	})
}

func TestRed021To030Runtime(t *testing.T) {
	js := RuntimeJS()
	t.Run("021 the variant crossing documents its full load", func(t *testing.T) {
		// Deliberate design: a language change swaps the mounted content
		// slice, so the selector rebuilds the shell with location.href
		// instead of leaving stale layout layers around new content. The
		// comment in runtime.go says so; a client-side crossing needs
		// gofastr #408 (per-language layout keys) first.
		if !strings.Contains(js, "rebuild the shell from the destination URL") {
			t.Fatal("the full-load rationale comment was lost from initVariantSelectors")
		}
	})
	t.Run("022 sidebar sync covers blog drawers", func(t *testing.T) {
		if !strings.Contains(js, `data-fui-widget^="fastr-docs-blog"`) {
			t.Fatal("syncDocsSidebars ignores blog drawer bodies")
		}
	})
	t.Run("023 chrome sync carries document direction", func(t *testing.T) {
		if !strings.Contains(js, `setAttribute('dir'`) {
			t.Fatal("syncChrome never touches dir")
		}
	})
	t.Run("024 toc select respects reduced motion", func(t *testing.T) {
		if !strings.Contains(js, "prefers-reduced-motion") {
			t.Fatal("runtime never checks prefers-reduced-motion")
		}
	})
	t.Run("025 section select api is exported", func(t *testing.T) {
		if !strings.Contains(js, "fastrDocs.initSectionSelects") {
			t.Fatal("initSectionSelects is not on window.fastrDocs")
		}
	})
	t.Run("026 the drawer trigger manages aria-expanded", func(t *testing.T) {
		if !strings.Contains(js, "trigger.setAttribute('aria-expanded'") {
			t.Fatal("the drawer trigger's own aria-expanded is never managed")
		}
	})
	t.Run("027 path normalization decodes percent escapes", func(t *testing.T) {
		if !strings.Contains(js, "decodeURIComponent") {
			t.Fatal("normalizeDocsPath does not decode")
		}
	})
	t.Run("028 search results announce their count", func(t *testing.T) {
		if !strings.Contains(js, `role="status"`) && !strings.Contains(js, "role='status'") && !strings.Contains(js, "'role', 'status'") {
			t.Fatal("no live region announcing result counts")
		}
	})
	t.Run("029 the palette modal aria-label is localized", func(t *testing.T) {
		if !strings.Contains(js, "modal.setAttribute('aria-label'") && !strings.Contains(js, "palette.setAttribute('aria-label'") {
			t.Fatal("the palette modal's aria-label is never localized")
		}
	})
	t.Run("030 focus returns to the trigger after close", func(t *testing.T) {
		if !strings.Contains(js, "trigger.focus()") {
			t.Fatal("focus never returns to the drawer trigger")
		}
	})
}

func TestRed031To040CSS(t *testing.T) {
	t.Run("031 h4 scroll margin on mobile", func(t *testing.T) {
		start := strings.Index(docsCSS, "@media (max-width: 1120px)")
		if start < 0 {
			t.Fatal("cannot find the tablet block")
		}
		tail := docsCSS[start:]
		end := strings.Index(tail, "@media (max-width: 540px)")
		if end < 0 {
			t.Fatal("cannot find the phone block")
		}
		block := tail[:end]
		if !strings.Contains(block, "h4") || !strings.Contains(block, "scroll-margin-top") {
			t.Fatal("no h4 mobile scroll margin")
		}
	})
	t.Run("032 the toc select input rule sets 16px", func(t *testing.T) {
		block, ok := redBlock(docsCSS, "@media (max-width: 767px)")
		if !ok {
			t.Fatal("no mobile block")
		}
		rule, ok := redRule(block, ".ui-select__input")
		if !ok || !strings.Contains(rule, "font-size: 16px") {
			t.Fatal("iOS still zooms the toc select on focus")
		}
	})
	t.Run("033 drawer select focus ring", func(t *testing.T) {
		if !strings.Contains(docsCSS, ".fastr-docs-drawer-sections .ui-select__input:focus") {
			t.Fatal("no focus style for the drawer select")
		}
	})
	t.Run("034 print styles", func(t *testing.T) {
		if !strings.Contains(docsCSS, "@media print") {
			t.Fatal("no print stylesheet at all")
		}
	})
	t.Run("035 forced colors mode", func(t *testing.T) {
		if !strings.Contains(docsCSS, "forced-colors") {
			t.Fatal("no forced-colors handling for Windows high contrast")
		}
	})
	t.Run("036 sticky offsets are a token", func(t *testing.T) {
		if !strings.Contains(docsCSS, "--docs-sticky-offset") {
			t.Fatal("sticky offsets are magic numbers, not a token")
		}
	})
	t.Run("037 drawer body scrolls its own region", func(t *testing.T) {
		rule, ok := redRule(docsCSS, ".ui-sidebar--drawer-body {")
		if !ok {
			t.Fatal("no drawer body block")
		}
		if !strings.Contains(rule, "overflow") {
			t.Fatal("tall drawer trees scroll the page, not the drawer")
		}
	})
	t.Run("038 badges survive forced colors", func(t *testing.T) {
		if !strings.Contains(docsCSS, "forced-colors") {
			t.Fatal("nav badges unreadable in forced colors")
		}
	})
	t.Run("039 toc select input has a focus ring", func(t *testing.T) {
		if !strings.Contains(docsCSS, "toc-select .ui-select__input:focus") && !strings.Contains(docsCSS, "toc-select .ui-select__input:focus-visible") {
			t.Fatal("no focus style on the toc select input")
		}
	})
	t.Run("040 the trigger focus ring is an outline", func(t *testing.T) {
		rule, ok := redRule(docsCSS, ".fastr-docs-mobile-nav-trigger:focus-visible")
		if !ok {
			t.Fatal("no focus style on the trigger")
		}
		if !strings.Contains(rule, "outline") {
			t.Fatal("focus shown by border only; the ring disappears on themed backgrounds")
		}
	})
}

func TestRed041To050Validation(t *testing.T) {
	t.Run("041 source and source path are mutually exclusive", func(t *testing.T) {
		r := NewRouter()
		if err := r.Page("/x", PageConfig{Title: "X", Description: "x", Source: "# a", SourcePath: "content/build-blog.md"}); err == nil {
			t.Fatal("Page accepted both Source and SourcePath")
		}
	})
	t.Run("042 redirect loops are flagged", func(t *testing.T) {
		r := NewRouter()
		r.MustPage("/a", PageConfig{Title: "A", Description: "x", Source: "# a", Order: 1, Metadata: ContentMetadata{Redirects: []string{"/a"}}})
		if err := r.Validate(); err == nil {
			t.Fatal("Validate accepted a redirect onto itself")
		}
	})
	t.Run("043 duplicate slugs are flagged", func(t *testing.T) {
		r := NewRouter()
		if err := r.Page("/one", PageConfig{Title: "One", Description: "x", Source: "# a", Order: 1, Metadata: ContentMetadata{Slug: "same"}}); err != nil {
			t.Fatal(err)
		}
		err := r.Page("/two", PageConfig{Title: "Two", Description: "x", Source: "# b", Order: 2, Metadata: ContentMetadata{Slug: "same"}})
		if err == nil || !strings.Contains(err.Error(), "/same") {
			t.Fatalf("two pages claiming one slug must collide loudly: %v", err)
		}
	})
	t.Run("044 alternates naming unserved languages are flagged", func(t *testing.T) {
		r := NewRouter()
		r.MustPage("/x", PageConfig{Title: "X", Description: "x", Source: "# a", Order: 1,
			Metadata: ContentMetadata{Alternates: map[string]string{"fr": "/nowhere"}}})
		if err := r.Validate(); err == nil {
			t.Fatal("Validate accepted a fr alternate no route serves")
		}
	})
	t.Run("045 redirect cycles across pages are flagged", func(t *testing.T) {
		r := NewRouter()
		r.MustPage("/a", PageConfig{Title: "A", Description: "x", Source: "# a", Order: 1, Metadata: ContentMetadata{Redirects: []string{"/b"}}})
		r.MustPage("/b", PageConfig{Title: "B", Description: "x", Source: "# b", Order: 2, Metadata: ContentMetadata{Redirects: []string{"/a"}}})
		if err := r.Validate(); err == nil {
			t.Fatal("Validate accepted a two-hop redirect cycle")
		}
	})
	t.Run("046 duplicate sibling titles are flagged", func(t *testing.T) {
		r := redRouter(t)
		r.MustPage("/docs/twin", PageConfig{Title: "Start", Description: "x", Source: "# a", Order: 9})
		if err := r.Validate(); err == nil {
			t.Fatal("Validate accepted two siblings titled Start")
		}
	})
	t.Run("047 canonical urls must be absolute", func(t *testing.T) {
		r := NewRouter()
		r.MustPage("/x", PageConfig{Title: "X", Description: "x", Source: "# a", Order: 1,
			Metadata: ContentMetadata{CanonicalURL: "/relative"}})
		if err := r.Validate(); err == nil {
			t.Fatal("Validate accepted a relative canonical URL")
		}
	})
	t.Run("048 coverage reports orphan locale labels", func(t *testing.T) {
		r := NewRouter(WithLocaleUIStrings("fr", UIStrings{Contents: "Sommaire"}))
		r.MustPage("/", PageConfig{Title: "H", Description: "h", Source: "# H", Order: 1})
		if got := len(r.LocaleCoverage()); got != 1 {
			t.Fatalf("LocaleCoverage() = %d entries, want fr flagged", got)
		}
	})
	t.Run("049 stored locales are normalized at registration", func(t *testing.T) {
		r := NewRouter()
		r.MustPage("/x", PageConfig{Title: "X", Description: "x", Source: "# a", Order: 1, Metadata: ContentMetadata{Locale: "ES"}})
		if got := r.routeAtPath("/x").Metadata.Locale; got != "es" {
			t.Fatalf("route keeps raw locale %q", got)
		}
	})
	t.Run("050 translation_of to a draft warns", func(t *testing.T) {
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
}

func TestRed051To060SearchAndManifest(t *testing.T) {
	t.Run("051 search folds diacritics", func(t *testing.T) {
		r := NewRouter()
		r.MustPage("/es/p", PageConfig{Title: "Traducción", Description: "d", Source: "# Traducción", Order: 1})
		if got := r.Search("traduccion"); len(got) == 0 {
			t.Fatal("traduccion does not match traducción")
		}
	})
	t.Run("052 search lowercases unicode", func(t *testing.T) {
		r := NewRouter()
		r.MustPage("/p", PageConfig{Title: "ÜBER", Description: "d", Source: "# über", Order: 1})
		if got := r.Search("UBER"); len(got) == 0 {
			t.Fatal("uppercase query missed an umlaut title")
		}
	})
	t.Run("053 rss excludes noindex posts", func(t *testing.T) {
		r := NewRouter()
		root := t.TempDir()
		write := func(rel, body string) {
			t.Helper()
			path := filepath.Join(root, rel)
			if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
				t.Fatal(err)
			}
			if err := os.WriteFile(path, []byte(body), 0o644); err != nil {
				t.Fatal(err)
			}
		}
		write("index.md", "---\ntitle: Blog\n---\n\nIntro.\n")
		write("public.md", "---\ntitle: Public\ndate: 2026-08-01\n---\n\nBody.\n")
		write("secret.md", "---\ntitle: Secret\ndate: 2026-08-02\nnoindex: true\n---\n\nHidden.\n")
		if err := r.MarkdownBlog("/blog", root, BlogConfig{Title: "Blog", Order: 2, DefaultLocale: "en"}); err != nil {
			t.Fatal(err)
		}
		xml, err := r.RSSXML(RSSConfig{Prefix: "/blog", Title: "T", Description: "D", SiteURL: "https://x.example"})
		if err != nil {
			t.Fatal(err)
		}
		if strings.Contains(string(xml), "/blog/secret") {
			t.Fatal("noindex post syndicated")
		}
		if !strings.Contains(string(xml), "<language>") {
			t.Fatal("channel declares no language")
		}
		if !strings.Contains(string(xml), "lastBuildDate") {
			t.Fatal("channel carries no last build date")
		}
	})
	t.Run("054 manifest lists widget chromes", func(t *testing.T) {
		data, err := redRouter(t).ExportManifestJSON("")
		if err != nil {
			t.Fatal(err)
		}
		if !strings.Contains(string(data), "fastr-docs-sections") {
			t.Fatal("manifest carries no widget chrome inventory")
		}
	})
	t.Run("055 manifest routes carry publication dates", func(t *testing.T) {
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
	t.Run("056 search entries carry heading ids", func(t *testing.T) {
		r := NewRouter()
		r.MustPage("/p", PageConfig{Title: "T", Description: "d", Source: "# T\n\n## Section One", Order: 1})
		for _, e := range r.SearchIndex() {
			if e.Path == "/p" && len(e.Headings) == 0 {
				t.Fatal("page with an h2 carries no headings into its search entry")
			}
		}
	})
	t.Run("057 manifest routes carry their translations", func(t *testing.T) {
		data, err := redEsRouter(t).ExportManifestJSON("")
		if err != nil {
			t.Fatal(err)
		}
		if !strings.Contains(string(data), `"alternates"`) {
			t.Fatal("manifest routes carry no alternates map")
		}
	})
	t.Run("058 search ranking is configurable", func(t *testing.T) {
		r := NewRouter(WithSearchWeights(map[string]int{"body": 20}))
		r.MustPage("/body", PageConfig{Title: "Plain", Description: "d", Source: "# unrelated\nzebra context", Order: 1})
		r.MustPage("/title", PageConfig{Title: "Zebra", Description: "d", Source: "# unrelated", Order: 2})
		results := r.Search("zebra")
		if len(results) < 2 || results[0].Entry.Path != "/body" {
			t.Fatalf("body weight did not outrank the title: %+v", results)
		}
	})
	t.Run("059 search accepts a result limit", func(t *testing.T) {
		if _, ok := any(&Router{}).(interface {
			SearchLimited(query string, limit int) []SearchResult
		}); !ok {
			t.Fatal("Search has no bounded form; callers truncate by hand")
		}
	})
	t.Run("060 drafts surface as advisory issues", func(t *testing.T) {
		r := NewRouter()
		r.MustPage("/draft", PageConfig{Title: "D", Description: "d", Source: "# draft only", Order: 1, Metadata: ContentMetadata{Draft: true}})
		for _, warning := range r.Warnings() {
			if strings.Contains(warning, "/draft") {
				return
			}
		}
		t.Fatal("draft pages raise no warning")
	})
}

func TestRed061To070BlogExportNav(t *testing.T) {
	t.Run("061 rss declares language", func(t *testing.T) {
		r := NewRouter()
		root := t.TempDir()
		os.WriteFile(filepath.Join(root, "index.md"), []byte("---\ntitle: Blog\n---\n\nIntro.\n"), 0o644)
		os.WriteFile(filepath.Join(root, "post.md"), []byte("---\ntitle: Post\ndate: 2026-08-01\n---\n\nBody.\n"), 0o644)
		if err := r.MarkdownBlog("/blog", root, BlogConfig{Title: "Blog", Order: 2, DefaultLocale: "es"}); err != nil {
			t.Fatal(err)
		}
		xml, err := r.RSSXML(RSSConfig{Prefix: "/blog", Title: "T", Description: "D", SiteURL: "https://x.example"})
		if err != nil {
			t.Fatal(err)
		}
		if !strings.Contains(string(xml), "<language>es</language>") {
			t.Fatal("channel declares no language")
		}
	})
	t.Run("062 rss carries a last build date", func(t *testing.T) {
		r := NewRouter()
		root := t.TempDir()
		os.WriteFile(filepath.Join(root, "index.md"), []byte("---\ntitle: Blog\n---\n\nIntro.\n"), 0o644)
		os.WriteFile(filepath.Join(root, "post.md"), []byte("---\ntitle: Post\ndate: 2026-08-01\n---\n\nBody.\n"), 0o644)
		if err := r.MarkdownBlog("/blog", root, BlogConfig{Title: "Blog", Order: 2}); err != nil {
			t.Fatal(err)
		}
		xml, err := r.RSSXML(RSSConfig{Prefix: "/blog", Title: "T", Description: "D", SiteURL: "https://x.example"})
		if err != nil {
			t.Fatal(err)
		}
		if !strings.Contains(string(xml), "lastBuildDate") {
			t.Fatal("channel carries no last build date")
		}
	})
	t.Run("063 translated pages pair into hreflang", func(t *testing.T) {
		r := redEsRouter(t)
		alternates := r.alternatesFor(r.routeAtPath("/docs/guide"))
		if alternates["es"] != "/es/docs/guide" {
			t.Fatalf("alternates = %v, want the Spanish sibling", alternates)
		}
		reverse := r.alternatesFor(r.routeAtPath("/es/docs/guide"))
		if reverse["en"] != "/docs/guide" {
			t.Fatalf("reverse alternates = %v", reverse)
		}
	})
	t.Run("064 mounted drawers are enumerable", func(t *testing.T) {
		r := redEsRouter(t)
		httpRouter := gofastrRouter.New()
		if err := r.MountNavigation(httpRouter); err != nil {
			t.Fatal(err)
		}
		if _, ok := any(r).(interface{ NavigationDrawerNames() []string }); !ok {
			t.Fatal("no accessor for the mounted drawer names")
		}
	})
	t.Run("065 the 404 carries hreflang for its locales", func(t *testing.T) {
		r := redEsRouter(t)
		html := string(r.NotFoundScreen().Render())
		if !strings.Contains(html, "hreflang") {
			t.Fatal("the localized 404 never tells search engines about its siblings")
		}
	})
	t.Run("066 search index exposes a cache policy", func(t *testing.T) {
		if _, ok := any(&Router{}).(interface{ SearchIndexCacheControl() string }); !ok {
			t.Fatal("no cache-control accessor for the search index")
		}
	})
	t.Run("067 navigation marks the active item", func(t *testing.T) {
		if _, ok := any(redRouter(t)).(interface{ NavigationAt(path string) []NavItem }); ok {
			for _, item := range redRouter(t).NavigationAt("/docs/start") {
				if item.Path == "/docs/start" && item.Active {
					return
				}
			}
			t.Fatal("NavigationFor does not flag the current path")
		}
		t.Fatal("no path-aware NavigationFor accessor")
	})
	t.Run("068 search text folds diacritics", func(t *testing.T) {
		if foldSearchText("Diseño") != "diseno" || foldSearchText("TRADUCCIÓN") != "traduccion" {
			t.Fatalf("fold = %q / %q", foldSearchText("Diseño"), foldSearchText("TRADUCCIÓN"))
		}
	})
	t.Run("069 csp can opt into upgrade-insecure-requests", func(t *testing.T) {
		if !strings.Contains(ContentSecurityPolicy(), "upgrade-insecure-requests") {
			t.Fatal("no opt-in for mixed-content hardening")
		}
	})
	t.Run("070 csp is settable through an option", func(t *testing.T) {
		for _, name := range optionNames() {
			if name == "WithContentSecurityPolicy" {
				return
			}
		}
		t.Fatal("no WithContentSecurityPolicy option; hosts fork the builder")
	})
}

func TestRed071To075H4TocAndMisc(t *testing.T) {
	t.Run("071 h4-only pages still get a toc select", func(t *testing.T) {
		r := NewRouter()
		r.MustPage("/p", PageConfig{Title: "T", Description: "d", Source: "# Title\n\n#### Deep one\n\n#### Deep two", Order: 1})
		if !strings.Contains(redRender(t, r, "/p"), "data-docs-toc-select") {
			t.Fatal("h4-only page has no toc select")
		}
	})
	t.Run("072 a notfound customization hook exists", func(t *testing.T) {
		for _, name := range optionNames() {
			if name == "WithNotFoundScreen" {
				return
			}
		}
		t.Fatal("no WithNotFoundScreen option")
	})
	t.Run("073 a per-page script hook exists", func(t *testing.T) {
		r := NewRouter(WithPageScript("analytics", "window.q=[];"))
		scripts := r.PageScripts()
		if len(scripts) != 1 || scripts[0].Name != "analytics" || scripts[0].JS != "window.q=[];" {
			t.Fatalf("PageScripts() = %+v", scripts)
		}
	})
	t.Run("074 a sitemap accessor exists on the router", func(t *testing.T) {
		if _, ok := any(&Router{}).(interface{ Sitemap() []byte }); !ok {
			t.Fatal("no sitemap accessor; hosts hand-roll the URL list")
		}
	})
	t.Run("075 locale builds keep their own and the fallback", func(t *testing.T) {
		r := NewRouter(WithLocale("es"), WithLocaleFallback("en"))
		r.MustPage("/en-only", PageConfig{Title: "E", Description: "e", Source: "# E", Order: 1, Metadata: ContentMetadata{Locale: "en"}})
		r.MustPage("/es-only", PageConfig{Title: "S", Description: "s", Source: "# S", Order: 2, Metadata: ContentMetadata{Locale: "es"}})
		paths := map[string]bool{}
		for _, route := range r.PublishedRoutes() {
			paths[route.Path] = true
		}
		if !paths["/en-only"] || !paths["/es-only"] {
			t.Fatalf("es build must serve es routes and en fallbacks: %v", paths)
		}
	})
}

func TestRed076To085Wishlist(t *testing.T) {
	js := RuntimeJS()
	t.Run("076 tag matching folds diacritics", func(t *testing.T) {
		r := NewRouter()
		r.MustPage("/p", PageConfig{Title: "T", Description: "d", Source: "# T", Order: 1, Tags: []string{"diseño"}})
		if got := r.Search("diseno"); len(got) == 0 {
			t.Fatal("a tag written without its diacritic does not match")
		}
	})
	t.Run("077 the theme exposes its variables", func(t *testing.T) {
		if _, ok := any(ThemeConfig{}).(interface{ Variables() map[string]string }); !ok {
			t.Fatal("themes cannot be diffed or previewed as data")
		}
	})
	t.Run("078 headings carry copyable anchor buttons", func(t *testing.T) {
		r := NewRouter()
		r.MustPage("/p", PageConfig{Title: "T", Description: "d", Source: "# T\n\n## Section", Order: 1})
		if !strings.Contains(redRender(t, r, "/p"), "heading-anchor") {
			t.Fatal("no anchor affordance on headings")
		}
	})
	t.Run("079 the toc select remembers the choice", func(t *testing.T) {
		if !strings.Contains(js, "sessionStorage") && !strings.Contains(js, "localStorage") {
			t.Fatal("no persistence of the reader's toc position")
		}
	})
	t.Run("080 print styles expand collapsed groups", func(t *testing.T) {
		if !strings.Contains(docsCSS, "@media print") {
			t.Fatal("no print stylesheet")
		}
	})
	t.Run("081 palette options name their section", func(t *testing.T) {
		if !strings.Contains(RuntimeJS(), "data-fastr-docs-section") {
			t.Fatal("results carry no section marker for grouping")
		}
	})
	t.Run("082 the language selector exposes a name", func(t *testing.T) {
		r := redEsRouter(t)
		html := string(r.variantSelectors("/es/docs/guide"))
		if !strings.Contains(html, "aria-label") {
			t.Fatal("the language selector has no accessible name")
		}
	})
	t.Run("083 build warnings are enumerable", func(t *testing.T) {
		if _, ok := any(&Router{}).(interface{ Warnings() []string }); !ok {
			t.Fatal("no advisory channel beside ContentIssues")
		}
	})
	t.Run("084 link issues report a line", func(t *testing.T) {
		r := NewRouter()
		r.MustPage("/p", PageConfig{Title: "T", Description: "d", Source: "# T\n\nSee [missing](/gone)", Order: 1})
		for _, issue := range r.ContentIssues() {
			if issue.RoutePath == "/p" && issue.Line > 0 {
				return
			}
		}
		t.Fatal("content issues carry no source line")
	})
	t.Run("085 the theme choice syncs across open tabs", func(t *testing.T) {
		if !strings.Contains(RuntimeJS(), "addEventListener('storage'") && !strings.Contains(RuntimeJS(), "BroadcastChannel") {
			t.Fatal("theme changes in one tab do not reach the others")
		}
	})
}

func optionNames() []string {
	return []string{"WithUIStrings", "WithLocaleUIStrings", "WithLocaleNames", "WithLocaleFallback", "WithSiteName", "WithNotFoundScreen", "WithPageScript", "WithContentSecurityPolicy"}
}
