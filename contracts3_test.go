package docs

// Contracts from the third audit cycle: options and assets, content rules,
// locales and runtime behavior, links and presentation, determinism, and
// the blog feed surface. Each case failed before its fix.

import (
	"context"
	"encoding/json"
	"reflect"
	"strings"
	"testing"
	"testing/fstest"

	uiapp "github.com/DonaldMurillo/gofastr/core-ui/app"
)

func r3Render(t *testing.T, r *Router, path string) string {
	t.Helper()
	site := uiapp.NewApp("T")
	if err := r.Mount(site, r.Layout()); err != nil {
		t.Fatal(err)
	}
	html, err := site.RenderPage(context.Background(), path)
	if err != nil {
		t.Fatal(err)
	}
	return string(html)
}

func r3BlogYear(t *testing.T) *Router {
	t.Helper()
	mapFS := fstest.MapFS{
		"index.md": &fstest.MapFile{Data: []byte("---\ntitle: Notas\ndescription: d\n---\n\nWelcome.")},
		"newer.md": &fstest.MapFile{Data: []byte("---\ntitle: Nuevo\ndescription: d\ndate: 2026-05-01\ntags: [go]\n---\n\nBody.")},
		"older.md": &fstest.MapFile{Data: []byte("---\ntitle: Viejo\ndescription: d\ndate: 2025-05-01\n---\n\nBody.")},
	}
	r := NewRouter()
	if err := r.MarkdownBlogFS("/blog", mapFS, ".", BlogConfig{}); err != nil {
		t.Fatal(err)
	}
	return r
}

func TestAssetAndOptionContracts(t *testing.T) {
	t.Run("toc titles drop math delimiters", func(t *testing.T) {
		for _, heading := range markdownHeadings("## The $x^2$ rule\n\nbody") {
			if strings.Contains(heading.Title, "$") {
				t.Fatalf("toc title carries math delimiters: %q", heading.Title)
			}
		}
	})
	t.Run("colliding plugin asset names error", func(t *testing.T) {
		r := NewRouter()
		r.MustPage("/g", PageConfig{Title: "G", Description: "d", Order: 1, Source: "# G"})
		dup := stringAssetPlugin{path: "/one", asset: "clash", body: []byte("one")}
		other := stringAssetPlugin{path: "/two", asset: "clash", body: []byte("two")}
		if err := r.Use(dup); err != nil {
			t.Fatal(err)
		}
		if err := r.Use(other); err != nil {
			t.Fatal(err)
		}
		if _, err := r.RuntimeAssets(""); err == nil || !strings.Contains(err.Error(), "collides") {
			t.Fatal("two plugins writing different bytes under one asset name silently overwrite each other")
		}
	})
	t.Run("duplicate page script names error", func(t *testing.T) {
		r := NewRouter(WithPageScript("poll", "window.p=1;"))
		r.MustPage("/g", PageConfig{Title: "G", Description: "d", Order: 1, Source: "# G"})
		if err := r.Page("/h", PageConfig{Title: "H", Description: "d", Order: 2, Source: "# H"}); err != nil {
			t.Fatal(err)
		}
		r2 := NewRouter(WithPageScript("poll", "a"), WithPageScript("poll", "b"))
		r2.MustPage("/g", PageConfig{Title: "G", Description: "d", Order: 1, Source: "# G"})
		if err := r2.Validate(); err == nil || !strings.Contains(err.Error(), "poll") {
			t.Fatalf("Validate() = %v, want the duplicate script name flagged", err)
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
	t.Run("a redirect with inner whitespace is flagged", func(t *testing.T) {
		r := NewRouter()
		r.MustPage("/g", PageConfig{Title: "G", Description: "d", Order: 1, Source: "# G",
			Metadata: ContentMetadata{Redirects: []string{"/a b"}}})
		if err := r.Validate(); err == nil || !strings.Contains(err.Error(), "redirect") {
			t.Fatalf("Validate() = %v, want a redirect-whitespace complaint", err)
		}
	})
	t.Run("an unknown page template is flagged", func(t *testing.T) {
		r := NewRouter()
		r.MustPage("/g", PageConfig{Title: "G", Description: "d", Order: 1, Source: "# G",
			Metadata: ContentMetadata{PageTemplate: "fancy"}})
		if err := r.Validate(); err == nil || !strings.Contains(err.Error(), "template") {
			t.Fatalf("Validate() = %v, want an unknown-template complaint", err)
		}
	})
}

type stringAssetPlugin struct {
	path, asset string
	body        []byte
}

func (p stringAssetPlugin) Name() string { return p.asset + "-plugin" }
func (p stringAssetPlugin) Apply(r *Router) error {
	r.MustPage(p.path, PageConfig{Title: strings.Trim(p.path, "/") + " page", Description: "d", Order: r.nextChildOrder("/", 9), Source: "# x"})
	return nil
}
func (p stringAssetPlugin) RuntimeAssets() (map[string][]byte, error) {
	return map[string][]byte{p.asset: p.body}, nil
}

func TestContentRuleContracts(t *testing.T) {
	t.Run("a splash hero without a title is flagged", func(t *testing.T) {
		r := NewRouter()
		r.MustPage("/g", PageConfig{Title: "G", Description: "d", Order: 1, Source: "# G",
			Metadata: ContentMetadata{PageTemplate: PageTemplateSplash, Hero: &PageHero{Tagline: "t"}}})
		if err := r.Validate(); err == nil || !strings.Contains(err.Error(), "hero") {
			t.Fatalf("Validate() = %v, want a hero-title complaint", err)
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
	t.Run("route titles are trimmed", func(t *testing.T) {
		r := NewRouter()
		r.MustPage("/g", PageConfig{Title: "  Guide  ", Description: "d", Order: 1, Source: "# G"})
		if got := r.routeAtPath("/g").Title; got != "Guide" {
			t.Fatalf("title = %q, want trimmed", got)
		}
	})
	t.Run("builtin shortcode prop names are checked", func(t *testing.T) {
		r := NewRouter()
		r.MustPage("/g", PageConfig{Title: "G", Description: "d", Order: 1,
			Source: "{{< tabs >}}{{< tab label=\"A\" colour=\"red\" >}}x{{< /tab >}}{{< /tabs >}}"})
		if err := r.Validate(); err == nil || !strings.Contains(err.Error(), "colour") {
			t.Fatalf("Validate() = %v, want the unknown tab prop named", err)
		}
	})
	t.Run("excerpts collapse to one line", func(t *testing.T) {
		r := NewRouter()
		r.MustPage("/g", PageConfig{Title: "G", Description: "d", Order: 1, Source: "# G",
			Metadata: ContentMetadata{Excerpt: "line one\nline two"}})
		if got := r.routeAtPath("/g").Metadata.Excerpt; strings.Contains(got, "\n") {
			t.Fatalf("excerpt keeps a newline: %q", got)
		}
	})
	t.Run("repeated authors collapse", func(t *testing.T) {
		r := NewRouter()
		r.MustPage("/g", PageConfig{Title: "G", Description: "d", Order: 1, Source: "# G",
			Metadata: ContentMetadata{Authors: []string{"Ada", "Ada"}}})
		if got := r.routeAtPath("/g").Metadata.Authors; len(got) != 1 {
			t.Fatalf("authors = %v, want deduplicated", got)
		}
	})
	t.Run("tags differing only by case collapse", func(t *testing.T) {
		r := NewRouter()
		r.MustPage("/g", PageConfig{Title: "G", Description: "d", Order: 1, Source: "# G",
			Metadata: ContentMetadata{Tags: []string{"Go", "go"}}})
		if got := r.routeAtPath("/g").Tags; len(got) != 1 {
			t.Fatalf("tags = %v, want folded to one", got)
		}
	})
	t.Run("the feed description uses the excerpt", func(t *testing.T) {
		r := r3BlogYear(t)
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
	t.Run("a year archive route exists", func(t *testing.T) {
		r := r3BlogYear(t)
		if r.routeAtPath("/blog/archive/2025") == nil {
			t.Fatal("no route serves the year archive")
		}
	})
	t.Run("a year archive lists only that year", func(t *testing.T) {
		r := r3BlogYear(t)
		html := r3Render(t, r, "/blog/archive/2025")
		cards := html[strings.Index(html, "fastr-docs-blog-card"):]
		if !strings.Contains(cards, "Viejo") || strings.Contains(cards, "Nuevo") {
			t.Fatal("the year archive mixes posts from other years")
		}
	})
}

func TestLocaleAndRuntimeContracts(t *testing.T) {
	t.Run("locale names fall back to the primary language", func(t *testing.T) {
		r := NewRouter(WithLocaleNames(map[string]string{"es": "Español"}))
		if got := r.LocaleName("es-MX"); got != "Español" {
			t.Fatalf("LocaleName(es-MX) = %q, want the Spanish name", got)
		}
	})
	t.Run("region label sets fall back to the primary locale", func(t *testing.T) {
		r := NewRouter(WithLocaleUIStrings("es", UIStrings{OnThisPage: "En esta página"}))
		r.MustPage("/g", PageConfig{Title: "G", Description: "d", Order: 1, Source: "# G"})
		r.MustPage("/es-MX/g", PageConfig{Title: "G", Description: "d", Order: 2, Source: "# G", Metadata: ContentMetadata{Locale: "es-MX"}})
		if got := r.UIStringsForLocale("es-mx").OnThisPage; got != "En esta página" {
			t.Fatalf("region labels = %q, want the primary locale's set", got)
		}
	})
	t.Run("region variants count toward coverage", func(t *testing.T) {
		r := NewRouter()
		r.MustPage("/g", PageConfig{Title: "G", Description: "d", Order: 1, Source: "# G"})
		r.MustPage("/es-MX/g", PageConfig{Title: "G", Description: "d", Order: 2, Source: "# G", Metadata: ContentMetadata{Locale: "es-MX"}})
		if missing := r.UntranslatedFamilies("es"); len(missing) != 0 {
			t.Fatalf("an es-MX variant does not count as Spanish coverage: %v", missing)
		}
	})
	t.Run("the static 404 answers in the site's language", func(t *testing.T) {
		r := NewRouter(WithLanguage("es"),
			WithLocaleUIStrings("es", UIStrings{NotFound: NotFoundStrings{Heading: "Página no encontrada", SiteFallback: "la documentación"}}))
		r.MustPage("/g", PageConfig{Title: "G", Description: "d", Order: 1, Source: "# G"})
		screen := r.NotFoundScreen()
		if screen.Strings.Heading != "Página no encontrada" {
			t.Fatalf("a Spanish-language site ships an English 404: %+v", screen.Strings)
		}
	})
	t.Run("the palette empty state is localized", func(t *testing.T) {
		js := RuntimeJS()
		if !strings.Contains(js, "data-fastr-docs-search-empty") {
			t.Fatal("the empty result row has no label hook to localize through")
		}
	})
	t.Run("the search trigger exposes its expanded state", func(t *testing.T) {
		r := NewRouter()
		r.MustPage("/g", PageConfig{Title: "G", Description: "d", Order: 1, Source: "# G"})
		html := r3Render(t, r, "/g")
		if !strings.Contains(html, "fastr-docs-command-trigger") {
			t.Fatal("no trigger rendered")
		}
		index := strings.Index(html, "fastr-docs-command-trigger")
		chunk := html[index-100 : index+400]
		if !strings.Contains(chunk, "aria-expanded") {
			t.Fatal("the palette trigger hides its expanded state from assistive tech")
		}
	})
}

func isSpanishish(s string) bool {
	return strings.ContainsAny(s, "áéíóúñ¿¡") || strings.Contains(s, " no ")
}

func TestLinkAndPresentationContracts(t *testing.T) {
	t.Run("internal links may point at a redirect source", func(t *testing.T) {
		r := NewRouter()
		r.MustPage("/g", PageConfig{Title: "G", Description: "d", Order: 1, Source: "# G",
			Metadata: ContentMetadata{Redirects: []string{"/old"}}})
		r.MustPage("/h", PageConfig{Title: "H", Description: "d", Order: 2, Source: "[old](/old)"})
		if err := r.Validate(); err != nil {
			t.Fatalf("a link to a redirect source is treated as broken: %v", err)
		}
	})
	t.Run("Markdown convenience pages get an order", func(t *testing.T) {
		r := NewRouter()
		if err := r.Markdown("/g", "Guide", "# Guide"); err != nil {
			t.Fatal(err)
		}
		if err := r.Validate(); err != nil {
			t.Fatalf("the one-liner helper cannot satisfy strict validation: %v", err)
		}
	})
	t.Run("the palette loading row is styled", func(t *testing.T) {
		if !strings.Contains(NewRouter().CSS(), ".fastr-docs-search-loading") {
			t.Fatal("the loading row renders as an unstyled list entry")
		}
	})
	t.Run("print shows where links lead", func(t *testing.T) {
		css := NewRouter().CSS()
		start := strings.Index(css, "@media print")
		end := start + strings.Index(css[start:], "\n}")
		if !strings.Contains(css[start:end], "attr(href)") {
			t.Fatal("printed pages keep external destinations invisible")
		}
	})
	t.Run("the sticky offset token is consumed", func(t *testing.T) {
		css := NewRouter().CSS()
		if !strings.Contains(css, "--docs-sticky-offset") {
			t.Fatal("token never defined")
		}
		if !strings.Contains(css[strings.Index(css, "scroll-margin-top"):strings.Index(css, "scroll-margin-top")+80], "var(--docs-sticky-offset") {
			t.Fatal("scroll margins hardcode their offsets instead of the token")
		}
	})
	t.Run("share buttons show a focus ring", func(t *testing.T) {
		if !strings.Contains(NewRouter().CSS(), "fastr-docs-blog-share:focus-visible") && !strings.Contains(NewRouter().CSS(), "[data-fastr-docs-share]:focus-visible") {
			t.Fatal("share controls are invisible to keyboard focus")
		}
	})
	t.Run("theme variables name their template", func(t *testing.T) {
		r := NewRouter(WithTemplate(TemplateEditorial))
		vars := r.ThemeConfig().Variables()
		if vars["template"] != "editorial" {
			t.Fatalf("Variables() = %v, want the template named", vars)
		}
	})
	t.Run("low-contrast theme overrides are warned", func(t *testing.T) {
		r := NewRouter(WithTheme(ThemeConfig{Overrides: ThemeOverrides{Background: "#ffffff", Text: "#ffffff"}}))
		r.MustPage("/g", PageConfig{Title: "G", Description: "d", Order: 1, Source: "# G"})
		for _, warning := range r.Warnings() {
			if strings.Contains(warning, "contrast") {
				return
			}
		}
		t.Fatal("white-on-white text ships without a word")
	})
}

func TestDeterminismContracts(t *testing.T) {
	t.Run("an unknown template is warned", func(t *testing.T) {
		r := NewRouter(WithTemplate(Template("fancy")))
		r.MustPage("/g", PageConfig{Title: "G", Description: "d", Order: 1, Source: "# G"})
		for _, warning := range r.Warnings() {
			if strings.Contains(warning, "fancy") {
				return
			}
		}
		t.Fatal("a typo'd template silently becomes the default")
	})
	t.Run("links inside code spans are not checked", func(t *testing.T) {
		r := NewRouter()
		r.MustPage("/g", PageConfig{Title: "G", Description: "d", Order: 1, Source: "Run `[check](/nope)` for detail."})
		if err := r.Validate(); err != nil {
			t.Fatalf("a documented dead link inside backticks fails the build: %v", err)
		}
	})
	t.Run("images without alt text are named", func(t *testing.T) {
		r := NewRouter()
		r.MustPage("/g", PageConfig{Title: "G", Description: "d", Order: 1, Source: "![]( /img.png )"})
		if err := r.Validate(); err == nil || !strings.Contains(err.Error(), "alt") {
			t.Fatalf("Validate() = %v, want an alt-text complaint", err)
		}
	})
	t.Run("manifest routes are sorted by path", func(t *testing.T) {
		r := NewRouter()
		r.MustPage("/z", PageConfig{Title: "Z", Description: "d", Order: 1, Source: "# Z"})
		r.MustPage("/a", PageConfig{Title: "A", Description: "d", Order: 2, Source: "# A"})
		body, err := r.ExportManifestJSON("")
		if err != nil {
			t.Fatal(err)
		}
		var manifest ExportManifest
		if err := json.Unmarshal(body, &manifest); err != nil {
			t.Fatal(err)
		}
		zi, ai := -1, -1
		for i, route := range manifest.Routes {
			if route.Path == "/z" {
				zi = i
			}
			if route.Path == "/a" {
				ai = i
			}
		}
		if zi < 0 || ai < 0 || ai > zi {
			t.Fatalf("manifest routes are not path-sorted: %+v", manifest.Routes)
		}
	})
	t.Run("the year archive carries the year in its title", func(t *testing.T) {
		r := r3BlogYear(t)
		route := r.routeAtPath("/blog/archive/2025")
		if route == nil || !strings.Contains(route.Title, "2025") {
			t.Fatalf("year route title = %+v", route)
		}
	})
}

func TestThirdSuiteTail(t *testing.T) {
	t.Run("manifest locales are sorted", func(t *testing.T) {
		body, err := r2Bilingual().ExportManifestJSON("")
		if err != nil {
			t.Fatal(err)
		}
		var manifest ExportManifest
		if err := json.Unmarshal(body, &manifest); err != nil {
			t.Fatal(err)
		}
		for i := 1; i < len(manifest.Locales); i++ {
			if manifest.Locales[i-1] > manifest.Locales[i] {
				t.Fatalf("manifest.Locales = %v, not sorted", manifest.Locales)
			}
		}
	})
	t.Run("manifest redirects are sorted", func(t *testing.T) {
		r := NewRouter()
		r.MustPage("/z", PageConfig{Title: "Z", Description: "d", Order: 1, Source: "# Z",
			Metadata: ContentMetadata{Redirects: []string{"/zz"}}})
		r.MustPage("/a", PageConfig{Title: "A", Description: "d", Order: 2, Source: "# A",
			Metadata: ContentMetadata{Redirects: []string{"/aa"}}})
		body, err := r.ExportManifestJSON("")
		if err != nil {
			t.Fatal(err)
		}
		var manifest ExportManifest
		if err := json.Unmarshal(body, &manifest); err != nil {
			t.Fatal(err)
		}
		for i := 1; i < len(manifest.Redirects); i++ {
			if manifest.Redirects[i-1].From > manifest.Redirects[i].From {
				t.Fatalf("manifest.Redirects = %+v, not sorted", manifest.Redirects)
			}
		}
	})
	t.Run("toc titles drop strikethrough markers", func(t *testing.T) {
		for _, heading := range markdownHeadings("## ~~Removed~~ feature\n\nbody") {
			if strings.Contains(heading.Title, "~~") {
				t.Fatalf("toc title carries strikethrough: %q", heading.Title)
			}
		}
	})
	t.Run("toc titles drop code backticks", func(t *testing.T) {
		for _, heading := range markdownHeadings("## The `router` type\n\nbody") {
			if strings.Contains(heading.Title, "`") {
				t.Fatalf("toc title carries code markers: %q", heading.Title)
			}
		}
	})
	t.Run("image headings keep only the alt text", func(t *testing.T) {
		for _, heading := range markdownHeadings("## Deploying ![build pipeline](/img.png) everywhere\n\nbody") {
			if strings.Contains(heading.Title, "img.png") || !strings.Contains(heading.Title, "build pipeline") {
				t.Fatalf("toc title mangles an imaged heading: %q", heading.Title)
			}
		}
	})
	t.Run("broken image sources are named", func(t *testing.T) {
		r := NewRouter()
		r.MustPage("/g", PageConfig{Title: "G", Description: "d", Order: 1, Source: "![diagram](/nope.png)"})
		if err := r.Validate(); err == nil || !strings.Contains(err.Error(), "/nope.png") {
			t.Fatalf("Validate() = %v, want the missing image named", err)
		}
	})
	t.Run("fragment targets fold case", func(t *testing.T) {
		r := NewRouter()
		r.MustPage("/g", PageConfig{Title: "G", Description: "d", Order: 1, Source: "## Deep section\n\nSee [deep](#Deep-Section)."})
		if err := r.Validate(); err != nil {
			t.Fatalf("a case-mismatched fragment is treated as broken: %v", err)
		}
	})
	t.Run("future posts stay out of search", func(t *testing.T) {
		r := r3BlogYear(t)
		route := r.routeAtPath("/blog/newer")
		route.Metadata.DatePublished = "2030-01-01"
		for _, entry := range r.SearchIndex() {
			if entry.Path == "/blog/newer" {
				t.Fatal("a future-dated post is searchable")
			}
		}
	})
	t.Run("future posts stay out of the sitemap", func(t *testing.T) {
		r := r3BlogYear(t)
		route := r.routeAtPath("/blog/newer")
		route.Metadata.DatePublished = "2030-01-01"
		if strings.Contains(string(r.Sitemap()), "/blog/newer") {
			t.Fatal("a future-dated post is offered to crawlers")
		}
	})
	t.Run("translation_of cycles are flagged", func(t *testing.T) {
		r := NewRouter()
		r.MustPage("/a", PageConfig{Title: "A", Description: "d", Order: 1, Source: "# A",
			Metadata: ContentMetadata{Locale: "es", TranslationOf: "/b"}})
		r.MustPage("/b", PageConfig{Title: "B", Description: "d", Order: 2, Source: "# B",
			Metadata: ContentMetadata{TranslationOf: "/a"}})
		if err := r.Validate(); err == nil || !strings.Contains(err.Error(), "translation") {
			t.Fatalf("Validate() = %v, want a translation cycle complaint", err)
		}
	})
	t.Run("sidebar group disclosures show focus", func(t *testing.T) {
		if !strings.Contains(NewRouter().CSS(), "summary.ui-sidebar__link:focus-visible") && !strings.Contains(NewRouter().CSS(), "summary:focus-visible") {
			t.Fatal("sidebar group summaries are invisible to keyboard focus")
		}
	})
	t.Run("heading anchor labels are translatable", func(t *testing.T) {
		labels := UIStrings{}
		field := reflect.ValueOf(&labels).Elem().FieldByName("AnchorLabel")
		if !field.IsValid() || !field.CanSet() {
			t.Fatal("UIStrings has no settable AnchorLabel; the anchor button is English on every page")
		}
		field.SetString("Copiar enlace a esta sección")
		r := NewRouter(WithLocaleUIStrings("es", labels))
		r.MustPage("/es/g", PageConfig{Title: "G", Description: "d", Order: 1, Source: "## Sección\n\nTexto.", Metadata: ContentMetadata{Locale: "es"}})
		html := r3Render(t, r, "/es/g")
		if !strings.Contains(html, "Copiar enlace") {
			t.Fatal("the anchor button ignores the locale's label")
		}
	})
	t.Run("docs css declares a print font size", func(t *testing.T) {
		css := NewRouter().CSS()
		start := strings.Index(css, "@media print")
		end := start + strings.Index(css[start:], "\n}")
		if !strings.Contains(css[start:end], "font-size") {
			t.Fatal("printed pages keep screen type sizes")
		}
	})
}

func TestThirdSuiteTopUp(t *testing.T) {
	t.Run("the empty palette row is styled", func(t *testing.T) {
		if !strings.Contains(NewRouter().CSS(), ".fastr-docs-search-empty") {
			t.Fatal("the empty result row renders unstyled")
		}
	})
	t.Run("headings balance their line breaks", func(t *testing.T) {
		if !strings.Contains(NewRouter().CSS(), "text-wrap") {
			t.Fatal("long headings never balance")
		}
	})
	t.Run("the palette modal contains scroll chaining", func(t *testing.T) {
		css := NewRouter().CSS()
		i := strings.Index(css, "fastr-docs-command-palette")
		if i < 0 || !strings.Contains(css[i:i+700], "overscroll-behavior") {
			t.Fatal("scrolling past the palette results scrolls the page behind")
		}
	})
	t.Run("reference-style links are checked", func(t *testing.T) {
		r := NewRouter()
		r.MustPage("/g", PageConfig{Title: "G", Description: "d", Order: 1, Source: "See [guide][g].\n\n[g]: /nope"})
		if err := r.Validate(); err == nil || !strings.Contains(err.Error(), "/nope") {
			t.Fatalf("Validate() = %v, want the reference target checked", err)
		}
	})
	t.Run("Keys is sorted", func(t *testing.T) {
		keys := UIStrings{}.Keys()
		for i := 1; i < len(keys); i++ {
			if keys[i-1] > keys[i] {
				t.Fatalf("Keys() is not sorted around %q", keys[i])
			}
		}
	})
	t.Run("the manifest honors the asset prefix", func(t *testing.T) {
		r := NewRouter()
		r.MustPage("/g", PageConfig{Title: "G", Description: "d", Order: 1, Source: "# G"})
		body, err := r.ExportManifestJSON("")
		if err != nil {
			t.Fatal(err)
		}
		var manifest ExportManifest
		if err := json.Unmarshal(body, &manifest); err != nil {
			t.Fatal(err)
		}
		if !strings.HasPrefix(manifest.AssetsPrefix, r.AssetPrefix()) {
			t.Fatalf("AssetsPrefix = %q, want the configured %q", manifest.AssetsPrefix, r.AssetPrefix())
		}
	})
	t.Run("generated blog views stay out of search", func(t *testing.T) {
		r := r3BlogYear(t)
		for _, entry := range r.SearchIndex() {
			if strings.HasPrefix(entry.Path, "/blog/tags") || strings.HasPrefix(entry.Path, "/blog/authors") || strings.HasPrefix(entry.Path, "/blog/search") {
				t.Fatalf("generated view %q pollutes the search index", entry.Path)
			}
		}
	})
	t.Run("generated blog views stay out of the manifest", func(t *testing.T) {
		r := r3BlogYear(t)
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
	t.Run("a versioned collection rejects stray root files", func(t *testing.T) {
		mapFS := fstest.MapFS{
			"readme.md":   &fstest.MapFile{Data: []byte("# stray")},
			"v1/index.md": &fstest.MapFile{Data: []byte("---\ntitle: V1\ndescription: d\n---\n\nBody.")},
		}
		r := NewRouter()
		err := r.MarkdownVersionedCollectionFS("/docs", mapFS, ".", VersionedCollectionConfig{Current: "v1"})
		if err == nil {
			t.Fatal("a stray file at the collection root becomes the index of every version")
		}
	})
	t.Run("the toc select hides without enough headings", func(t *testing.T) {
		r := NewRouter()
		r.MustPage("/g", PageConfig{Title: "G", Description: "d", Order: 1, Source: "## Only\n\nOne heading."})
		html := r3Render(t, r, "/g")
		if strings.Contains(html, "fastr-docs-toc-select") {
			t.Fatal("a one-heading page still renders a toc select")
		}
	})
	t.Run("print hides the share row", func(t *testing.T) {
		css := NewRouter().CSS()
		start := strings.Index(css, "@media print")
		end := start + strings.Index(css[start:], "\n}")
		if !strings.Contains(css[start:end], "share") {
			t.Fatal("share buttons print beside the post")
		}
	})
	t.Run("the runtime exposes its base helpers", func(t *testing.T) {
		js := RuntimeJS()
		if !strings.Contains(js, "fastrDocs.withBase") || !strings.Contains(js, "fastrDocs.base") {
			t.Fatal("plugins cannot share the runtime's base-path helpers")
		}
	})
	t.Run("the runtime exposes path normalization", func(t *testing.T) {
		if !strings.Contains(RuntimeJS(), "fastrDocs.normalizePath") {
			t.Fatal("plugins cannot compare route paths the way the runtime does")
		}
	})
	t.Run("the feed names its generator", func(t *testing.T) {
		r := r3BlogYear(t)
		feed, err := r.RSSXML(RSSConfig{Prefix: "/blog"})
		if err != nil {
			t.Fatal(err)
		}
		if !strings.Contains(string(feed), "<generator>") {
			t.Fatal("the feed does not say what generated it")
		}
	})
}

func TestThirdSuiteFeed(t *testing.T) {
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
	t.Run("fence languages fold case", func(t *testing.T) {
		language, options := splitFenceInfo("GO {1}")
		fence := parseFenceOptions(language, options)
		if fence.language != "go" || len(fence.highlight) != 1 {
			t.Fatalf("fence = %+v, want go with its highlight", fence)
		}
	})
	t.Run("fence languages tolerate trailing space", func(t *testing.T) {
		language, _ := splitFenceInfo("go ")
		if language != "go" {
			t.Fatalf("splitFenceInfo = %q, want go", language)
		}
	})
	t.Run("shortcodes inside headings are named", func(t *testing.T) {
		r := NewRouter()
		r.MustPage("/g", PageConfig{Title: "G", Description: "d", Order: 1, Source: "## Notes {{< callout >}}x{{< /callout >}}\n\nbody"})
		if err := r.Validate(); err == nil || !strings.Contains(err.Error(), "heading") {
			t.Fatalf("Validate() = %v, want a block-shortcode-in-heading complaint", err)
		}
	})
}

func TestRed268To278Root(t *testing.T) {
	t.Run("kbd chips are styled", func(t *testing.T) {
		css := NewRouter().CSS()
		i := strings.Index(css, "\nkbd {")
		if i < 0 || !strings.Contains(css[i:i+220], "border") {
			t.Fatal("keyboard keys render as plain text")
		}
	})
	t.Run("theme changes re-sync the chrome color", func(t *testing.T) {
		if !strings.Contains(RuntimeJS(), ":theme") {
			t.Fatal("a theme change never reaches the browser chrome color")
		}
	})
	t.Run("scroll listeners are passive", func(t *testing.T) {
		if !strings.Contains(RuntimeJS(), "passive:") {
			t.Fatal("scroll and touch listeners block the main thread")
		}
	})
	t.Run("term pages hide future posts", func(t *testing.T) {
		r := r3BlogYear(t)
		route := r.routeAtPath("/blog/newer")
		route.Metadata.DatePublished = "2030-01-01"
		html := r3Render(t, r, "/blog/tags/go")
		cards := html[strings.Index(html, "fastr-docs-blog"):]
		if strings.Contains(cards, "Nuevo") {
			t.Fatal("a future-dated post appears on a tag page")
		}
	})
	t.Run("partial chrome translations warn", func(t *testing.T) {
		r := NewRouter(WithLocaleUIStrings("es", UIStrings{Home: "Inicio"}))
		r.MustPage("/g", PageConfig{Title: "G", Description: "d", Order: 1, Source: "# G"})
		for _, warning := range r.Warnings() {
			if strings.Contains(warning, "chrome") || strings.Contains(warning, "labels") {
				return
			}
		}
		t.Fatal("a locale translating one chrome label of seventy passes silently")
	})
	t.Run("generated views stay out of the sitemap", func(t *testing.T) {
		r := r3BlogYear(t)
		if strings.Contains(string(r.Sitemap()), "/blog/tags") || strings.Contains(string(r.Sitemap()), "/blog/search") {
			t.Fatal("generated listing pages are offered to crawlers as content")
		}
	})
	t.Run("the blog search input is debounced", func(t *testing.T) {
		js := RuntimeJS()
		start := strings.Index(js, "function initBlogSearch")
		end := start + strings.Index(js[start:], "function blogShareURL")
		if start < 0 || !strings.Contains(js[start:end], "setTimeout") {
			t.Fatal("every keystroke re-filters and re-announces the whole list")
		}
	})
	t.Run("server-side blog search folds accents", func(t *testing.T) {
		post := &Route{Title: "Nuevo", Metadata: ContentMetadata{Tags: []string{"Diseño"}}}
		if !blogQueryMatch(post, "diseno") {
			t.Fatal("a folded query misses an accented tag before JavaScript runs")
		}
	})
	t.Run("excerpts strip link syntax", func(t *testing.T) {
		if got := markdownText("Read [the guide](/guide) now."); strings.Contains(got, "](/guide)") {
			t.Fatalf("excerpt keeps markdown link syntax: %q", got)
		}
	})
	t.Run("pages with empty bodies are flagged", func(t *testing.T) {
		r := NewRouter()
		r.MustPage("/g", PageConfig{Title: "G", Description: "d", Order: 1, Source: "---\ntitle: G\ndescription: d\n---\n"})
		if err := r.Validate(); err == nil || !strings.Contains(err.Error(), "empty") {
			t.Fatalf("Validate() = %v, want an empty-page complaint", err)
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
}

func TestRed279To281Root(t *testing.T) {
	t.Run("openapi console marks busy in markup", func(t *testing.T) {
		if !strings.Contains(RuntimeJS(), "aria-busy") {
			t.Fatal("the docs runtime never marks its surfaces busy")
		}
	})
	t.Run("versioned collections pair their versions", func(t *testing.T) {
		mapFS := fstest.MapFS{
			"v1/index.md": &fstest.MapFile{Data: []byte("---\ntitle: One\ndescription: d\n---\n\nBody.")},
			"v2/index.md": &fstest.MapFile{Data: []byte("---\ntitle: Two\ndescription: d\n---\n\nBody.")},
		}
		r := NewRouter()
		if err := r.MarkdownVersionedCollectionFS("/docs", mapFS, ".", VersionedCollectionConfig{Current: "v2", Versions: []string{"v1", "v2"}}); err != nil {
			t.Fatal(err)
		}
		r.MustPage("/", PageConfig{Title: "Home", Description: "d", Order: 1, Source: "# H"})
		for _, name := range r.NavigationDrawerNames() {
			if strings.Contains(name, "v1") {
				return
			}
		}
		t.Fatalf("version drawers missing: %v", r.NavigationDrawerNames())
	})
	t.Run("blog cards expose their dates to machines", func(t *testing.T) {
		r := r3BlogYear(t)
		html := r3Render(t, r, "/blog")
		if !strings.Contains(html, "<time") {
			t.Fatal("card dates render as bare text")
		}
	})
}

func TestRed282To282Root(t *testing.T) {
	t.Run("blog feeds carry their collection title", func(t *testing.T) {
		r := NewRouter()
		mapFS := fstest.MapFS{
			"index.md": &fstest.MapFile{Data: []byte("---\ntitle: Notas\ndescription: d\n---\n\nWelcome.")},
		}
		if err := r.MarkdownBlogFS("/blog", mapFS, ".", BlogConfig{Title: "Engineering Notes"}); err != nil {
			t.Fatal(err)
		}
		feed, err := r.RSSXML(RSSConfig{Prefix: "/blog"})
		if err != nil {
			t.Fatal(err)
		}
		if !strings.Contains(string(feed), "<title>Engineering Notes</title>") {
			t.Fatal("a titled collection feeds under a synthesized name")
		}
	})
}
