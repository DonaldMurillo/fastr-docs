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

// The tab labels stay in the language their sections were registered in.
// Translating a route title is the project's call; the framework only makes
// sure the same section is not listed twice.
func TestNavTabLabelsAreNotInvented(t *testing.T) {
	r := chromeLocaleSite(t)
	for _, path := range []string{"/guide", "/es/guide"} {
		items := r.headerItems(path)
		if len(items) != 1 || items[0].Label != "Guide" {
			t.Fatalf("%s nav = %+v, want the one registered section", path, items)
		}
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
