package docs

import (
	"strings"
	"testing"
)

const splashSource = `---
template: splash
hero:
  eyebrow: Docs for builders
  title: Ship documentation fast
  tagline: One route tree drives everything.
  actions:
    - text: Get started
      link: /docs/getting-started
      variant: primary
    - text: API
      link: https://example.com/api
      variant: secondary
---

## Body heading

Regular Markdown below the hero.
`

func splashRoute(t *testing.T, source string) *Route {
	t.Helper()
	r := NewRouter()
	r.MustPage("/", PageConfig{Title: "Home", Description: "Landing", Order: 1, Source: source})
	if err := r.Validate(); err != nil {
		t.Fatalf("Validate() error = %v", err)
	}
	return r.Routes()[0]
}

func TestSplashFrontMatterParsesTheHero(t *testing.T) {
	route := splashRoute(t, splashSource)
	if route.Metadata.PageTemplate != PageTemplateSplash {
		t.Fatalf("PageTemplate = %q, want %q", route.Metadata.PageTemplate, PageTemplateSplash)
	}
	hero := route.Metadata.Hero
	if hero == nil {
		t.Fatal("hero front matter was not parsed")
	}
	if hero.Eyebrow != "Docs for builders" || hero.Title != "Ship documentation fast" || hero.Tagline != "One route tree drives everything." {
		t.Fatalf("hero = %+v", hero)
	}
	if len(hero.Actions) != 2 {
		t.Fatalf("hero actions = %d, want 2", len(hero.Actions))
	}
	if hero.Actions[0].Text != "Get started" || hero.Actions[0].Link != "/docs/getting-started" {
		t.Fatalf("first action = %+v", hero.Actions[0])
	}
}

func TestSplashPageRendersTheHeroAndDropsTheTOC(t *testing.T) {
	r := NewRouter()
	r.MustPage("/", PageConfig{Title: "Home", Description: "Landing", Order: 1, Source: splashSource})
	if err := r.Validate(); err != nil {
		t.Fatalf("Validate() error = %v", err)
	}
	html := string((&pageComponent{router: r, route: r.Routes()[0]}).Render())

	for _, want := range []string{
		"Docs for builders",
		"Ship documentation fast",
		"One route tree drives everything.",
		`href="/docs/getting-started"`,
		"fastr-docs-splash",
		"Body heading",
	} {
		if !strings.Contains(html, want) {
			t.Fatalf("splash output missing %q: %s", want, html)
		}
	}
	// A landing page has no table of contents and no breadcrumb trail.
	if strings.Contains(html, "fastr-docs-toc") {
		t.Fatalf("splash page rendered a table of contents: %s", html)
	}
	if strings.Contains(html, "ui-doc-layout__crumbs") {
		t.Fatalf("splash page rendered breadcrumbs: %s", html)
	}
}

func TestOrdinaryPagesKeepTheirTOCAndCrumbs(t *testing.T) {
	r := NewRouter()
	r.MustPage("/guide", PageConfig{
		Title: "Guide", Description: "A guide", Order: 1,
		Source: "# Guide\n\n## First\n\nText.\n\n## Second\n\nMore.\n",
	})
	if err := r.Validate(); err != nil {
		t.Fatalf("Validate() error = %v", err)
	}
	html := string((&pageComponent{router: r, route: r.Routes()[0]}).Render())
	if strings.Contains(html, "fastr-docs-splash") {
		t.Fatalf("a normal page picked up the splash shell: %s", html)
	}
	if !strings.Contains(html, "fastr-docs-toc") {
		t.Fatalf("a normal page lost its table of contents: %s", html)
	}
}

// Hero links come from front matter, which is content, so they get the same
// treatment as a shortcode's href.
func TestSplashHeroRejectsScriptLinksAndEmptyActions(t *testing.T) {
	route := splashRoute(t, `---
template: splash
hero:
  title: Landing
  actions:
    - text: Bad
      link: "javascript:alert(1)"
    - text: ""
      link: /fine
    - text: Good
      link: /docs
---

Body.
`)
	html := string(renderPageHero(route.Metadata.Hero))
	if strings.Contains(strings.ToLower(html), "javascript:") {
		t.Fatalf("hero passed through a javascript: URL: %s", html)
	}
	if !strings.Contains(html, `href="/docs"`) {
		t.Fatalf("hero dropped a valid action: %s", html)
	}
	if strings.Contains(html, "/fine") {
		t.Fatalf("hero rendered an action with no label: %s", html)
	}
}

func TestPageWithoutHeroFrontMatterRendersNothingExtra(t *testing.T) {
	if got := renderPageHero(nil); got != "" {
		t.Fatalf("renderPageHero(nil) = %q, want empty", got)
	}
	if got := renderPageHero(&PageHero{}); got != "" {
		t.Fatalf("renderPageHero(empty) = %q, want empty", got)
	}
}

func TestSplashHeroesNeedTitlesAndLinkedActions(t *testing.T) {
	t.Run("a hero without a title is flagged", func(t *testing.T) {
		r := NewRouter()
		r.MustPage("/g", PageConfig{Title: "G", Description: "d", Order: 1, Source: "# G",
			Metadata: ContentMetadata{PageTemplate: PageTemplateSplash, Hero: &PageHero{Tagline: "t"}}})
		if err := r.Validate(); err == nil || !strings.Contains(err.Error(), "hero") {
			t.Fatalf("Validate() = %v, want a hero-title complaint", err)
		}
	})
	t.Run("hero actions require links", func(t *testing.T) {
		r := NewRouter()
		r.MustPage("/g", PageConfig{Title: "G", Description: "d", Order: 1, Source: "# G",
			Metadata: ContentMetadata{PageTemplate: PageTemplateSplash, Hero: &PageHero{Title: "T", Actions: []PageHeroAction{{Text: "Go"}}}}})
		if err := r.Validate(); err == nil || !strings.Contains(err.Error(), "action") {
			t.Fatalf("Validate() = %v, want a hero action link complaint", err)
		}
	})
}
