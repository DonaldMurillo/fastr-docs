package docs

import (
	"regexp"
	"strconv"
	"strings"
	"testing"

	gofastrRouter "github.com/DonaldMurillo/gofastr/core/router"
)

// Below md the inline sidebar is gone and the header tabs are hidden, so the
// drawer is the only navigation a phone has. A section select at its top is
// what lets a reader jump between top-level sections without scrolling the
// whole tree.
func TestDrawerCarriesASectionSelect(t *testing.T) {
	r := NewRouter()
	r.MustPage("/", PageConfig{Title: "Home", Description: "Home", Source: "# Home", Order: 1})
	docs := r.MustGroup("/docs", GroupConfig{Title: "Docs", Description: "Guides", Order: 2})
	docs.MustPage("start", PageConfig{Title: "Start", Description: "Start", Source: "# Start", Order: 1})
	r.MustPage("/api-reference", PageConfig{Title: "API reference", Description: "API", Source: "# API", Order: 3})

	home := r.localeHome("")
	body := string(r.docsDrawerBody(home, docsDrawerName("")))
	for _, want := range []string{
		"data-fastr-docs-section-select",
		// Home first, then every section in the order the header tabs use.
		`<option value="/">Home</option>`,
		// A group's option aims at its first visible page, matching the tab.
		`<option value="/docs/start">Docs</option>`,
		`<option value="/api-reference">API reference</option>`,
		// The nav tree still follows: the select replaces nothing.
		`ui-sidebar__nav`,
	} {
		if !strings.Contains(body, want) {
			t.Fatalf("drawer body missing %q: %s", want, body)
		}
	}
}

// One drawer per language, because the drawer is mounted once for the whole
// site: a single English drawer would hand a Spanish reader the English tree
// with an English home link leading out of the Spanish site.
func TestEachLocaleGetsItsOwnDrawer(t *testing.T) {
	r := NewRouter(
		WithLocaleFallback("en"),
		WithLocaleNames(map[string]string{"en": "English", "es": "Español"}),
		WithLocaleUIStrings("es", UIStrings{
			Contents: "Contenido",
			Home:     "Inicio",
			Sections: "Secciones",
		}),
	)
	r.MustPage("/", PageConfig{Title: "Home", Description: "Home", Source: "# Home", Order: 1, Metadata: ContentMetadata{Locale: "en"}})
	english := r.MustGroup("/docs", GroupConfig{Title: "Docs", Description: "en", Order: 2})
	english.MustPage("guide", PageConfig{Title: "Guide", Description: "en", Source: "# Guide", Order: 1, Metadata: ContentMetadata{Locale: "en"}})
	r.MustPage("/es", PageConfig{Title: "Español", Description: "es", Source: "# Inicio", Order: 2, Metadata: ContentMetadata{Locale: "es"}})
	spanish := r.MustGroup("/es/docs", GroupConfig{Title: "Documentación", Description: "es", Order: 3, Locale: "es"})
	spanish.MustPage("guide", PageConfig{Title: "Guía", Description: "es", Source: "# Guía", Order: 1, Metadata: ContentMetadata{Locale: "es"}})

	homes := r.localeDrawerHomes()
	if len(homes) != 2 {
		t.Fatalf("localeDrawerHomes() = %d homes, want the default and Spanish", len(homes))
	}
	if homes[0].Path != "/" || homes[1].Path != "/es" {
		t.Fatalf("localeDrawerHomes() = [%s, %s], want the default home first then /es", homes[0].Path, homes[1].Path)
	}

	es := string(r.docsDrawerBody(r.routes["/es"], docsDrawerName("es")))
	for _, want := range []string{
		"Secciones",
		`<option value="/es">Inicio</option>`,
		`<option value="/es/docs/guide">Documentación</option>`,
		`href="/es/docs/guide"`,
	} {
		if !strings.Contains(es, want) {
			t.Fatalf("Spanish drawer missing %q: %s", want, es)
		}
	}
	if strings.Contains(es, `href="/docs/guide"`) || strings.Contains(es, `<option value="/">`) {
		t.Fatalf("Spanish drawer leaks the default locale's routes: %s", es)
	}

	en := string(r.docsDrawerBody(r.localeHome(""), docsDrawerName("")))
	if !strings.Contains(en, `href="/docs/guide"`) || strings.Contains(en, "/es") {
		t.Fatalf("default drawer must stay in the default locale: %s", en)
	}

	httpRouter := gofastrRouter.New()
	if err := r.MountNavigation(httpRouter); err != nil {
		t.Fatalf("MountNavigation() error = %v", err)
	}
	for _, name := range []string{"fastr-docs-sections", "fastr-docs-sections-es"} {
		var found bool
		for _, route := range httpRouter.Routes() {
			if strings.Contains(route.Pattern, name) {
				found = true
				break
			}
		}
		if !found {
			t.Fatalf("MountNavigation did not register the %q drawer", name)
		}
	}

	// The trigger carries the home-path-to-drawer map the runtime re-aims
	// it with; the path is the key so the runtime can prefix-match a URL,
	// exactly like the blog drawer map.
	header := string((&docsHeader{router: r}).render("/es/docs/guide"))
	if !strings.Contains(header, `data-fastr-docs-locale-drawers="/es=fastr-docs-sections-es"`) {
		t.Fatalf("header trigger lacks the locale drawer map: %s", header)
	}
}

// A site with no route at "/" used to mount the drawer from the bare roots
// list; the per-locale rework must keep that working rather than dereference
// a home that does not exist.
func TestMountNavigationSurvivesASiteWithNoHomeRoute(t *testing.T) {
	r := NewRouter()
	docs := r.MustGroup("/docs", GroupConfig{Title: "Docs", Description: "Guides", Order: 1})
	docs.MustPage("start", PageConfig{Title: "Start", Description: "Start", Source: "# Start", Order: 1})

	httpRouter := gofastrRouter.New()
	if err := r.MountNavigation(httpRouter); err != nil {
		t.Fatalf("MountNavigation() error = %v", err)
	}
	var found bool
	for _, route := range httpRouter.Routes() {
		if strings.Contains(route.Pattern, "fastr-docs-sections") {
			found = true
			break
		}
	}
	if !found {
		t.Fatal("a home-less site still needs the default drawer under its pinned name")
	}
}

// The default drawer keeps the short name by position, not by comparing
// locale spellings: a site that tags its original pages "en" without
// declaring WithLocaleFallback would otherwise mount fastr-docs-sections-en
// and leave the header trigger aimed at a widget nobody mounted.
func TestTheDefaultDrawerKeepsItsNameWithoutADeclaredFallback(t *testing.T) {
	r := NewRouter()
	r.MustPage("/", PageConfig{Title: "Home", Description: "Home", Source: "# Home", Order: 1, Metadata: ContentMetadata{Locale: "en"}})
	r.MustPage("/es", PageConfig{Title: "Español", Description: "es", Source: "# Inicio", Order: 2, Metadata: ContentMetadata{Locale: "es"}})
	spanish := r.MustGroup("/es/docs", GroupConfig{Title: "Documentación", Description: "es", Order: 3, Locale: "es"})
	spanish.MustPage("guide", PageConfig{Title: "Guía", Description: "es", Source: "# Guía", Order: 1, Metadata: ContentMetadata{Locale: "es"}})

	httpRouter := gofastrRouter.New()
	if err := r.MountNavigation(httpRouter); err != nil {
		t.Fatalf("MountNavigation() error = %v", err)
	}
	patterns := make([]string, 0, 4)
	for _, route := range httpRouter.Routes() {
		patterns = append(patterns, route.Pattern)
	}
	joined := strings.Join(patterns, " ")
	if !strings.Contains(joined, "fastr-docs-sections") || strings.Contains(joined, "fastr-docs-sections-en") {
		t.Fatalf("default drawer lost its pinned name: %s", joined)
	}
	if !strings.Contains(joined, "fastr-docs-sections-es") {
		t.Fatalf("Spanish drawer missing: %s", joined)
	}
	header := string((&docsHeader{router: r}).render("/es/docs/guide"))
	if !strings.Contains(header, `data-fastr-docs-locale-drawers="/es=fastr-docs-sections-es"`) {
		t.Fatalf("header trigger lacks the locale drawer map: %s", header)
	}
}

// Phone width stacks the sticky header (62px) over the sticky TOC select
// (measured 85.5px), so a heading needs at least 148px of scroll margin to
// clear them. 144px dropped the top of every h2 under the select.
func TestMobileHeadingScrollOffsetClearsTheStickyBars(t *testing.T) {
	// Only the media blocks where the TOC select is visible bind at phone
	// width; the base 110px rule serves the desktop rail, where the sticky
	// select does not exist. The effective value is the last declaration
	// inside those blocks, which is what the cascade scrolls to.
	start := strings.Index(docsCSS, "@media (max-width: 1120px)")
	end := strings.Index(docsCSS, "@media (max-width: 540px)")
	if start < 0 || end < 0 || end <= start {
		t.Fatal("could not isolate the tablet and phone media blocks")
	}
	mobile := docsCSS[start:end]
	effective := func(rule string) float64 {
		re := regexp.MustCompile(regexp.QuoteMeta(rule) + `[^{]*\{[^}]*scroll-margin-top:\s*(\d+(?:\.\d+)?)px`)
		matches := re.FindAllStringSubmatch(mobile, -1)
		if len(matches) == 0 {
			t.Fatalf("no scroll-margin-top for %q in the mobile blocks", rule)
		}
		value, err := strconv.ParseFloat(matches[len(matches)-1][1], 64)
		if err != nil {
			t.Fatalf("scroll-margin-top for %q = %q: %v", rule, matches[len(matches)-1][1], err)
		}
		return value
	}
	// 62px sticky header + 85.5px sticky select = 147.5px of chrome; 160px
	// clears them with air. h2 declared 144px at the phone breakpoint, which
	// parked the top of every h2 under the select.
	if got := effective(".layout-docs .ui-markdown h2"); got < 160 {
		t.Fatalf("h2 scroll-margin-top = %vpx in the mobile blocks, want at least 160", got)
	}
	if got := effective(".layout-docs .ui-markdown h3"); got < 160 {
		t.Fatalf("h3 scroll-margin-top = %vpx in the mobile blocks, want at least 160", got)
	}
}
