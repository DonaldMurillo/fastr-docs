package docs

import (
	"strings"
	"testing"
)

// A site that translates slugs as well as prose. /es/docs/guia does not mirror
// /docs/guide, so path-shaped pairing finds nothing; translation_of says which
// page it is.
func translatedSlugSite(t *testing.T) *Router {
	t.Helper()
	r := NewRouter(
		WithSiteName("Docs"),
		WithLocaleFallback("en"),
		WithLocaleNames(map[string]string{"en": "English", "es": "Español", "fr": "Français"}),
	)
	r.MustPage("/", PageConfig{Title: "Docs", Description: "Home", Source: "# Docs\n"})
	docs := r.MustGroup("/docs", GroupConfig{Title: "Documentation", Description: "en", Order: 1})
	docs.MustPage("guide", PageConfig{Title: "Guide", Description: "en", Source: "# Guide\n", Order: 1})
	docs.MustPage("setup", PageConfig{Title: "Setup", Description: "en", Source: "# Setup\n", Order: 2})

	r.MustPage("/es", PageConfig{Title: "Inicio", Description: "es", Source: "# Inicio\n", Metadata: ContentMetadata{Locale: "es"}})
	// The section is translated too, under a translated segment.
	documentacion := r.MustGroup("/es/documentacion", GroupConfig{Title: "Documentación", Description: "es", Order: 2, Locale: "es", TranslationOf: "/docs"})
	documentacion.MustPage("guia", PageConfig{Title: "Guía", Description: "es", Order: 1,
		Source: "---\nlocale: es\ntranslation_of: /docs/guide\n---\n# Guía\n"})
	return r
}

func optionHrefs(options []docsVariantOption) map[string]string {
	out := map[string]string{}
	for _, option := range options {
		out[option.value] = option.href
	}
	return out
}

func TestATranslatedSlugPairsThroughTranslationOf(t *testing.T) {
	r := translatedSlugSite(t)
	guia := r.routeAtPath("/es/documentacion/guia")
	if guia == nil {
		t.Fatal("no route at /es/documentacion/guia")
	}
	if guia.Metadata.TranslationOf != "/docs/guide" {
		t.Fatalf("front matter translation_of was not read: %+v", guia.Metadata)
	}
	if got := r.familyOf(guia); got != "docs/guide" {
		t.Fatalf("familyOf(guia) = %q, want the original's family", got)
	}

	// The selector on the Spanish page offers the English original, and the
	// selector on the original offers the Spanish page, whatever its slug.
	if got := optionHrefs(r.variantOptions(guia, "locale")); got["en"] != "/docs/guide" || got["es"] != "/es/documentacion/guia" {
		t.Fatalf("Spanish page selector = %v", got)
	}
	guide := r.routeAtPath("/docs/guide")
	if got := optionHrefs(r.variantOptions(guide, "locale")); got["es"] != "/es/documentacion/guia" {
		t.Fatalf("English page selector = %v", got)
	}

	// A page with no translation is unchanged by any of this.
	if options := r.variantOptions(r.routeAtPath("/docs/setup"), "locale"); len(options) > 1 {
		t.Fatalf("untranslated page grew a selector: %+v", options)
	}
}

// A group pairs the same way, so the header tab and the sidebar section follow
// the reader into the translated section rather than staying English above a
// Spanish article.
func TestATranslatedSectionPairsThroughTranslationOf(t *testing.T) {
	r := translatedSlugSite(t)
	items := r.headerItems("/es/documentacion/guia")
	if len(items) != 1 || items[0].Label != "Documentación" {
		t.Fatalf("Spanish nav = %+v, want the translated section only", items)
	}
	items = r.headerItems("/docs/guide")
	if len(items) != 1 || items[0].Label != "Documentation" {
		t.Fatalf("English nav = %+v, want the original section only", items)
	}
}

// Pointing at another translation lands in the same family as pointing at
// the original, so a third language can reference whichever page its
// translator worked from.
func TestTranslationOfFollowsThroughAnotherTranslation(t *testing.T) {
	r := translatedSlugSite(t)
	r.MustPage("/fr/docs/le-guide", PageConfig{Title: "Le guide", Description: "fr", Source: "# Le guide\n",
		Metadata: ContentMetadata{Locale: "fr", TranslationOf: "/es/documentacion/guia"}})
	got := optionHrefs(r.variantOptions(r.routeAtPath("/docs/guide"), "locale"))
	if got["fr"] != "/fr/docs/le-guide" || got["es"] != "/es/documentacion/guia" {
		t.Fatalf("selector on the original = %v", got)
	}
	// Coverage sees the family as translated into both languages.
	if missing := r.UntranslatedFamilies("fr"); contains(missing, "docs/guide") {
		t.Fatalf("docs/guide reported untranslated into fr: %v", missing)
	}
}

// The path-shaped pairing still works beside it, so a site can translate one
// section folder for folder and another page by reference.
func TestPathShapedPairingStillWorksBesideTranslationOf(t *testing.T) {
	r := translatedSlugSite(t)
	r.MustPage("/es/docs/setup", PageConfig{Title: "Instalación", Description: "es", Source: "# Instalación\n",
		Metadata: ContentMetadata{Locale: "es"}})
	if got := optionHrefs(r.variantOptions(r.routeAtPath("/docs/setup"), "locale")); got["es"] != "/es/docs/setup" {
		t.Fatalf("mirrored path stopped pairing: %v", got)
	}
}

// Every pair emits hreflang both ways without anyone writing an alternates
// map, and a declared alternate still wins over the derived one.
func TestPairedPagesEmitHreflangBothWays(t *testing.T) {
	r := translatedSlugSite(t)
	guide, guia := r.routeAtPath("/docs/guide"), r.routeAtPath("/es/documentacion/guia")
	if got := r.alternatesFor(guide); got["es"] != "/es/documentacion/guia" {
		t.Fatalf("alternates of the original = %v", got)
	}
	if got := r.alternatesFor(guia); got["en"] != "/docs/guide" {
		t.Fatalf("alternates of the translation = %v", got)
	}
	head := r.metadataHeadHTML(guide)
	if !strings.Contains(head, `<link rel="alternate" hreflang="es" href="/es/documentacion/guia">`) {
		t.Fatalf("original head lacks the Spanish alternate: %s", head)
	}
	head = r.metadataHeadHTML(guia)
	if !strings.Contains(head, `<link rel="alternate" hreflang="en" href="/docs/guide">`) {
		t.Fatalf("translation head lacks the English alternate: %s", head)
	}
	// The page with no metadata of its own still gets a head, or the pairing
	// would be invisible to search engines from that side.
	if !r.routeHasMetadata(guide) {
		t.Fatal("the original is not wrapped for head metadata")
	}

	guia.Metadata.Alternates = map[string]string{"en": "https://docs.example.com/guide"}
	if got := r.alternatesFor(guia); got["en"] != "https://docs.example.com/guide" {
		t.Fatalf("declared alternate lost: %v", got)
	}

	for _, entry := range r.SearchIndex() {
		if entry.Path == "/docs/guide" && entry.Alternates["es"] != "/es/documentacion/guia" {
			t.Fatalf("search entry alternates = %v", entry.Alternates)
		}
	}
}

// A reference that cannot pair fails loudly. Silently it is a page missing
// from the selector, which nobody notices until a reader does.
func TestTranslationOfThatCannotPairIsAnIssue(t *testing.T) {
	r := translatedSlugSite(t)
	if issues := r.ContentIssues(); len(issues) != 0 {
		t.Fatalf("a valid site reported issues: %v", issues)
	}
	r.MustPage("/es/nowhere", PageConfig{Title: "Nada", Description: "es", Source: "# Nada\n",
		Metadata: ContentMetadata{Locale: "es", TranslationOf: "/docs/missing"}})
	r.MustPage("/es/self", PageConfig{Title: "Yo", Description: "es", Source: "# Yo\n",
		Metadata: ContentMetadata{Locale: "es", TranslationOf: "/es/self"}})
	r.MustPage("/docs/copy", PageConfig{Title: "Copy", Description: "en", Source: "# Copy\n",
		Metadata: ContentMetadata{TranslationOf: "/docs/guide"}})

	issues := r.ContentIssues()
	if len(issues) != 3 {
		t.Fatalf("want three issues, got %d: %v", len(issues), issues)
	}
	messages := make([]string, 0, len(issues))
	for _, issue := range issues {
		messages = append(messages, issue.RoutePath+": "+issue.Message)
	}
	for _, want := range []string{
		`/docs/copy: translation_of points at "/docs/guide", which is in the same language ("en")`,
		`/es/nowhere: translation_of points at "/docs/missing", which no route serves`,
		`/es/self: translation_of points at the page itself`,
	} {
		if !containsPrefix(messages, want) {
			t.Fatalf("missing issue %q in %v", want, messages)
		}
	}
	// The broken reference pairs with nothing rather than with the wrong page.
	if options := r.variantOptions(r.routeAtPath("/es/nowhere"), "locale"); len(options) > 1 {
		t.Fatalf("a dangling reference produced a selector: %+v", options)
	}
}

func contains(values []string, want string) bool {
	for _, value := range values {
		if value == want {
			return true
		}
	}
	return false
}

func containsPrefix(values []string, prefix string) bool {
	for _, value := range values {
		if strings.HasPrefix(value, prefix) {
			return true
		}
	}
	return false
}
