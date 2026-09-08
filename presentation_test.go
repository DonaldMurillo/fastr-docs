package docs

// CSS contracts: print, forced colors, contrast, focus rings,
// dynamic viewport, RTL, wrapping, and the accent.
import (
	"strings"
	"testing"
)

func TestPresentation(t *testing.T) {
	css := NewRouter().CSS()
	t.Run("print hides heading anchors", func(t *testing.T) {
		css := NewRouter().CSS()
		start := strings.Index(css, "@media print")
		end := start + strings.Index(css[start:], "}")
		if start < 0 || !strings.Contains(css[start:end], "heading-anchor") {
			t.Fatal("the print block leaves the # glyphs in the output")
		}
	})

	t.Run("reduced motion stills the blog cards", func(t *testing.T) {
		css := NewRouter().CSS()
		start := strings.Index(css, "prefers-reduced-motion")
		if start < 0 || !strings.Contains(css[start:start+400], "fastr-docs-blog-card") {
			t.Fatal("blog card transforms ignore prefers-reduced-motion")
		}
	})

	t.Run("deep-linked headings clear the sticky bars", func(t *testing.T) {
		if !strings.Contains(css, ":target") {
			t.Fatal("scroll-margin exists for h2/h3 but not for :target arrivals")
		}
	})

	t.Run("a contrast preference block exists", func(t *testing.T) {
		if !strings.Contains(css, "prefers-contrast") {
			t.Fatal("increased-contrast readers get the default palette")
		}
	})

	t.Run("the drawer heights track the dynamic viewport", func(t *testing.T) {
		if !strings.Contains(css, "dvh") {
			t.Fatal("the mobile drawer sizes against the small static viewport")
		}
	})

	t.Run("toc rail links show a focus ring", func(t *testing.T) {
		if !strings.Contains(css, ".ui-anchored-rail__list a:focus-visible") {
			t.Fatal("the on-this-page rail is invisible to keyboard focus")
		}
	})

	t.Run("highlighted code survives forced colors", func(t *testing.T) {
		start := strings.Index(css, "forced-colors")
		if start < 0 || !strings.Contains(css[start:start+600], "fastr-docs-code-hl") {
			t.Fatal("code highlighting lives on background color alone")
		}
	})

	t.Run("rtl documents have a style story", func(t *testing.T) {
		if !strings.Contains(css, "[dir=") && !strings.Contains(css, ":dir(") {
			t.Fatal("DirectionFor sets dir but no rule answers it")
		}
	})

	t.Run("the heading anchor is styled", func(t *testing.T) {
		if !strings.Contains(NewRouter().CSS(), ".heading-anchor") {
			t.Fatal("anchor links render as raw blue underlined text")
		}
	})

	t.Run("the section select help line is styled", func(t *testing.T) {
		if !strings.Contains(NewRouter().CSS(), ".ui-select__help") {
			t.Fatal("the drawer select help paragraph renders unstyled")
		}
	})

	t.Run("long tokens wrap in prose", func(t *testing.T) {
		if !strings.Contains(NewRouter().CSS(), "overflow-wrap") {
			t.Fatal("a long URL or hash blows out the content column")
		}
	})

	t.Run("print hides the blog toolbar", func(t *testing.T) {
		css := NewRouter().CSS()
		start := strings.Index(css, "@media print")
		end := start + strings.Index(css[start:], "}")
		if !strings.Contains(css[start:end], "blog") {
			t.Fatal("feed links and tag filters print beside the post")
		}
	})

	t.Run("form controls follow the accent", func(t *testing.T) {
		if !strings.Contains(NewRouter().CSS(), "accent-color") {
			t.Fatal("checkboxes and radios ignore the theme accent")
		}
	})
	t.Run("the search count region is styled", func(t *testing.T) {
		if !strings.Contains(NewRouter().CSS(), ".fastr-docs-search-count") {
			t.Fatal("the announced result count renders as bare paragraph text")
		}
	})
}
