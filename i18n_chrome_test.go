package docs

import (
	"strings"
	"testing"
)

// localeSite mirrors the shape a partly translated site has: an English tree
// and a Spanish subtree whose paths differ only by the locale segment, so
// variantFamily pairs them.
func chromeLocaleSite(t *testing.T) *Router {
	t.Helper()
	r := NewRouter(
		WithSiteName("Docs"),
		WithLocaleFallback("en"),
		WithLocaleNames(map[string]string{"en": "English", "es": "Español"}),
		WithLocaleUIStrings("es", UIStrings{
			Contents:   "Contenido",
			Home:       "Inicio",
			OnThisPage: "En esta página",
			Search:     "Buscar",
			Language:   "Idioma",
		}),
	)
	r.MustPage("/", PageConfig{Title: "Docs", Description: "Home", Source: "# Page\n", Metadata: ContentMetadata{Locale: "en"}})
	r.MustPage("/guide", PageConfig{Title: "Guide", Description: "Guide", Source: "# Page\n", Metadata: ContentMetadata{Locale: "en"}})
	r.MustPage("/es", PageConfig{Title: "Español", Description: "Inicio", Source: "# Pagina\n", Metadata: ContentMetadata{Locale: "es"}})
	r.MustPage("/es/guide", PageConfig{Title: "Guía", Description: "Guía", Source: "# Pagina\n", Metadata: ContentMetadata{Locale: "es"}})
	return r
}

// One build serves both languages, so the chrome has to follow the page rather
// than the Router. Otherwise a Spanish page arrives wrapped in English
// furniture: "Contents", "On this page", "Search".
func TestChromeFollowsThePagesLocale(t *testing.T) {
	r := chromeLocaleSite(t)
	if got := r.uiAt("/es/guide").Contents; got != "Contenido" {
		t.Fatalf("Spanish page got Contents = %q", got)
	}
	if got := r.uiAt("/guide").Contents; got != "Contents" {
		t.Fatalf("English page got Contents = %q", got)
	}
	// A label the translation omits falls back rather than rendering empty.
	if got := r.uiAt("/es/guide").EditPage; got != defaultUIStrings.EditPage {
		t.Fatalf("untranslated label = %q, want the English default", got)
	}
	// An unknown locale is not an error.
	if got := r.UIStringsForLocale("fr").Contents; got != "Contents" {
		t.Fatalf("unknown locale = %q", got)
	}
}

func TestLocaleUIStringsMergeOverTheRouterWideSet(t *testing.T) {
	r := NewRouter(
		WithUIStrings(UIStrings{Contents: "Sections", Home: "Start"}),
		WithLocaleUIStrings("es", UIStrings{Contents: "Secciones"}),
	)
	labels := r.UIStringsForLocale("es")
	if labels.Contents != "Secciones" {
		t.Fatalf("locale override lost: %q", labels.Contents)
	}
	// The Router-wide value shows through where the locale says nothing, which
	// is what lets a site translate one label at a time.
	if labels.Home != "Start" {
		t.Fatalf("Router-wide label lost: %q", labels.Home)
	}
}

// A reader who needs the language selector is the one least able to read a
// label in the wrong language, so it shows the language's own name.
func TestTheLanguageSelectorNamesLanguages(t *testing.T) {
	r := chromeLocaleSite(t)
	html := string(r.variantSelectors("/es/guide"))
	for _, want := range []string{"English", "Español", "Idioma"} {
		if !strings.Contains(html, want) {
			t.Fatalf("selector missing %q: %s", want, html)
		}
	}
	// The value stays the code; only the label is translated.
	if !strings.Contains(html, `data-docs-variant-value="es"`) {
		t.Fatalf("selector lost its locale value: %s", html)
	}
	// Without WithLocaleNames it falls back to the code rather than blanking.
	plain := NewRouter()
	if got := plain.LocaleName("es"); got != "es" {
		t.Fatalf("LocaleName fallback = %q, want the code", got)
	}
}

// A translated section is the same section. Left alone it appears beside the
// original as its own tab, where it reads as a topic rather than a language.
func TestATranslatedSectionIsNotASecondNavTab(t *testing.T) {
	r := chromeLocaleSite(t)
	for _, path := range []string{"/guide", "/es/guide"} {
		labels := make([]string, 0, 4)
		for _, item := range r.headerItems(path) {
			labels = append(labels, item.Label)
		}
		if len(labels) != 1 {
			t.Fatalf("%s produced tabs %v, want one section", path, labels)
		}
	}
}

// Where a translation of a section exists, the tab points at it, so a reader
// who switched language stays in it while moving around the site.
//
// A translated section is usually not a root: /es/docs sits under /es, so a
// search that only walked the roots found nothing and the whole nav stayed in
// the source language.
func TestNavTabsFollowTheReadersLanguage(t *testing.T) {
	r := chromeLocaleSite(t)
	spanish := r.headerItems("/es/guide")
	if len(spanish) != 1 || spanish[0].Label != "Guía" || spanish[0].Href != "/es/guide" {
		t.Fatalf("Spanish nav = %+v, want the Spanish section", spanish)
	}
	english := r.headerItems("/guide")
	if len(english) != 1 || english[0].Label != "Guide" || english[0].Href != "/guide" {
		t.Fatalf("English nav = %+v, want the English section", english)
	}
}

// A section with no translation keeps its original label rather than
// disappearing, which is the normal state of a partly translated site.
func TestAnUntranslatedSectionStaysInTheNav(t *testing.T) {
	r := chromeLocaleSite(t)
	r.MustPage("/api", PageConfig{Title: "API", Description: "API", Source: "# API\n", Metadata: ContentMetadata{Locale: "en"}})
	labels := make([]string, 0, 2)
	for _, item := range r.headerItems("/es/guide") {
		labels = append(labels, item.Label)
	}
	if len(labels) != 2 || labels[1] != "API" {
		t.Fatalf("nav = %v, want the untranslated section kept", labels)
	}
}

// The home route never appears as a tab, but its family still has to be
// claimed, or its translation lands in the nav as one.
func TestATranslatedHomeIsNotANavTab(t *testing.T) {
	r := chromeLocaleSite(t)
	for _, item := range r.headerItems("/es") {
		if item.Label == "Español" {
			t.Fatalf("the Spanish home became a nav tab: %+v", r.headerItems("/es"))
		}
	}
}

// A site with no translations must be untouched by any of this.
func TestASingleLocaleSiteIsUnaffected(t *testing.T) {
	r := NewRouter(WithSiteName("Docs"))
	r.MustPage("/", PageConfig{Title: "Docs", Description: "Home", Source: "# Page\n"})
	r.MustPage("/guide", PageConfig{Title: "Guide", Description: "Guide", Source: "# Page\n"})
	r.MustPage("/api", PageConfig{Title: "API", Description: "API", Source: "# Page\n"})

	if got := r.uiAt("/guide").Contents; got != "Contents" {
		t.Fatalf("Contents = %q", got)
	}
	if items := r.headerItems("/guide"); len(items) != 2 {
		t.Fatalf("nav = %+v, want both sections", items)
	}
	if html := string(r.variantSelectors("/guide")); strings.Contains(html, "variant-select") {
		t.Fatalf("a single-locale site got a selector: %s", html)
	}
}

// A section is usually a group, and a group has no front matter, so before
// GroupConfig.Locale existed a translated section could not be paired with its
// original at all: variantFamily had no locale segment to strip.
func TestAGroupCanDeclareItsLocale(t *testing.T) {
	r := NewRouter(WithSiteName("Docs"), WithLocaleFallback("en"))
	english := r.MustGroup("/guides", GroupConfig{Title: "Guides", Description: "en", Order: 1})
	english.MustPage("intro", PageConfig{Title: "Intro", Description: "en", Source: "# Intro\n", Order: 1})
	spanish := r.MustGroup("/es/guides", GroupConfig{Title: "Guías", Description: "es", Order: 2, Locale: "es"})
	spanish.MustPage("intro", PageConfig{Title: "Intro", Description: "es", Source: "# Intro\n", Order: 1,
		Metadata: ContentMetadata{Locale: "es"}})

	var translated *Route
	for _, root := range r.roots {
		if root.Path == "/es/guides" {
			translated = root
		}
	}
	if translated == nil {
		t.Fatal("the translated group was not registered")
	}
	if got := variantFamily(translated); got != "guides" {
		t.Fatalf("family of the translated group = %q, want %q", got, "guides")
	}
	// The nav swaps the section for its translation.
	items := r.headerItems("/es/guides/intro")
	if len(items) != 1 || items[0].Label != "Guías" {
		t.Fatalf("nav = %+v, want the Spanish section", items)
	}
}

// Marking `locale: en` on every original page to get a selector is busywork,
// and forgetting it fails silently. With a default locale named, an unmarked
// route counts as being in it.
func TestAnUnmarkedRoutePairsWithTheDefaultLocale(t *testing.T) {
	r := NewRouter(WithSiteName("Docs"), WithLocaleFallback("en"),
		WithLocaleNames(map[string]string{"en": "English", "es": "Español"}))
	// No locale in front matter, which is how a monolingual site is written.
	r.MustPage("/guide", PageConfig{Title: "Guide", Description: "en", Source: "# Guide\n", Order: 1})
	r.MustPage("/es/guide", PageConfig{Title: "Guía", Description: "es", Source: "# Guia\n", Order: 2,
		Metadata: ContentMetadata{Locale: "es"}})

	if got := r.effectiveLocale(r.routeAtPath("/guide")); got != "en" {
		t.Fatalf("unmarked route counts as %q, want the default locale", got)
	}
	html := string(r.variantSelectors("/guide"))
	for _, want := range []string{"English", "Español"} {
		if !strings.Contains(html, want) {
			t.Fatalf("selector missing %q: %s", want, html)
		}
	}
}

// Pairing is not publication. An unmarked route is still served in every locale
// build, so treating it as the default locale must not hide it.
func TestTheDefaultLocaleRuleDoesNotChangeWhatPublishes(t *testing.T) {
	r := NewRouter(WithSiteName("Docs"), WithLocale("es"), WithLocaleFallback("en"))
	r.MustPage("/shared", PageConfig{Title: "Shared", Description: "no locale", Source: "# Shared\n", Order: 1})
	for _, route := range r.PublishedRoutes() {
		if route.Path == "/shared" {
			return
		}
	}
	t.Fatal("a route with no declared locale stopped publishing in an es build")
}

// A locale home is the translation of "/", not a section of the site. Left as
// the active root it wraps the whole translated tree in one extra level the
// default locale does not have.
func TestALocaleHomeIsNotAnExtraSidebarLevel(t *testing.T) {
	r := chromeLocaleSite(t)
	spanish := r.sidebarItems(r.sidebarRoots("/es/guide"), "/es/guide")
	english := r.sidebarItems(r.sidebarRoots("/guide"), "/guide")
	if len(spanish) != len(english) {
		t.Fatalf("Spanish sidebar has %d top-level items, English has %d", len(spanish), len(english))
	}
	for _, item := range spanish {
		if item.Label == "Español" {
			t.Fatalf("the locale home appeared as a section: %+v", spanish)
		}
	}
}

// Translating the home's label while leaving its link alone is worse than not
// translating it: "Inicio" quietly took the reader out of the Spanish site.
func TestTheHomeLinkStaysInTheReadersLanguage(t *testing.T) {
	r := chromeLocaleSite(t)
	for path, want := range map[string]string{"/es/guide": "/es", "/guide": "/"} {
		items := r.sidebarItems(r.sidebarRoots(path), path)
		if len(items) == 0 {
			t.Fatalf("%s produced no sidebar", path)
		}
		// The home is always first, whatever explicit Order the routes carry.
		if items[0].Href != want {
			t.Fatalf("%s home link = %q, want %q", path, items[0].Href, want)
		}
		if len(items[0].Children) != 0 {
			t.Fatalf("%s home is a section, not a link: %+v", path, items[0])
		}
	}
	if got := r.sidebarItems(r.sidebarRoots("/es/guide"), "/es/guide")[0].Label; got != "Inicio" {
		t.Fatalf("Spanish home label = %q", got)
	}
}
