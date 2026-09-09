package docs

// CSS coverage: print, forced colors, contrast, focus rings, dynamic
// viewport, RTL, wrapping, and the accent.
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
	t.Run("h4 scroll margin on mobile", func(t *testing.T) {
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
	t.Run("the toc select input rule sets 16px", func(t *testing.T) {
		block, ok := cssMediaBlock(docsCSS, "@media (max-width: 767px)")
		if !ok {
			t.Fatal("no mobile block")
		}
		rule, ok := cssRule(block, ".ui-select__input")
		if !ok || !strings.Contains(rule, "font-size: 16px") {
			t.Fatal("iOS still zooms the toc select on focus")
		}
	})
	t.Run("drawer select focus ring", func(t *testing.T) {
		if !strings.Contains(docsCSS, ".fastr-docs-drawer-sections .ui-select__input:focus") {
			t.Fatal("no focus style for the drawer select")
		}
	})
	t.Run("print styles", func(t *testing.T) {
		if !strings.Contains(docsCSS, "@media print") {
			t.Fatal("no print stylesheet at all")
		}
	})
	t.Run("forced colors mode", func(t *testing.T) {
		if !strings.Contains(docsCSS, "forced-colors") {
			t.Fatal("no forced-colors handling for Windows high contrast")
		}
	})
	t.Run("sticky offsets are a token", func(t *testing.T) {
		if !strings.Contains(docsCSS, "--docs-sticky-offset") {
			t.Fatal("sticky offsets are magic numbers, not a token")
		}
	})
	t.Run("the sticky offset token is consumed", func(t *testing.T) {
		css := NewRouter().CSS()
		if !strings.Contains(css[strings.Index(css, "scroll-margin-top"):strings.Index(css, "scroll-margin-top")+80], "var(--docs-sticky-offset") {
			t.Fatal("scroll margins hardcode their offsets instead of the token")
		}
	})
	t.Run("drawer body scrolls its own region", func(t *testing.T) {
		rule, ok := cssRule(docsCSS, ".ui-sidebar--drawer-body {")
		if !ok {
			t.Fatal("no drawer body block")
		}
		if !strings.Contains(rule, "overflow") {
			t.Fatal("tall drawer trees scroll the page, not the drawer")
		}
	})
	t.Run("toc select input has a focus ring", func(t *testing.T) {
		if !strings.Contains(docsCSS, "toc-select .ui-select__input:focus") && !strings.Contains(docsCSS, "toc-select .ui-select__input:focus-visible") {
			t.Fatal("no focus style on the toc select input")
		}
	})
	t.Run("the trigger focus ring is an outline", func(t *testing.T) {
		rule, ok := cssRule(docsCSS, ".fastr-docs-mobile-nav-trigger:focus-visible")
		if !ok {
			t.Fatal("no focus style on the trigger")
		}
		if !strings.Contains(rule, "outline") {
			t.Fatal("focus shown by border only; the ring disappears on themed backgrounds")
		}
	})
	t.Run("focus styling comes from a token", func(t *testing.T) {
		if !strings.Contains(NewRouter().CSS(), "--docs-focus") {
			t.Fatal("focus ring colors are hardcoded at every use site")
		}
	})
	t.Run("the palette loading row is styled", func(t *testing.T) {
		if !strings.Contains(NewRouter().CSS(), ".fastr-docs-search-loading") {
			t.Fatal("the loading row renders as an unstyled list entry")
		}
	})
	t.Run("the palette empty row is styled", func(t *testing.T) {
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
	t.Run("view transitions are styled", func(t *testing.T) {
		if !strings.Contains(NewRouter().CSS(), "view-transition") {
			t.Fatal("SPA swaps cut hard with no crossfade")
		}
	})
	t.Run("selections follow the theme", func(t *testing.T) {
		if !strings.Contains(NewRouter().CSS(), "::selection") {
			t.Fatal("selected text renders in the browser default blue")
		}
	})
	t.Run("display math gets block styling", func(t *testing.T) {
		if !strings.Contains(NewRouter().CSS(), "fastr-docs-math--display") {
			t.Fatal("display math renders inline with no block treatment from the docs layer")
		}
	})
	t.Run("kbd chips are styled", func(t *testing.T) {
		css := NewRouter().CSS()
		i := strings.Index(css, "\nkbd {")
		if i < 0 || !strings.Contains(css[i:i+220], "border") {
			t.Fatal("keyboard keys render as plain text")
		}
	})
	t.Run("sidebar group disclosures show focus", func(t *testing.T) {
		if !strings.Contains(NewRouter().CSS(), "summary.ui-sidebar__link:focus-visible") && !strings.Contains(NewRouter().CSS(), "summary:focus-visible") {
			t.Fatal("sidebar group summaries are invisible to keyboard focus")
		}
	})
	t.Run("share buttons show a focus ring", func(t *testing.T) {
		if !strings.Contains(NewRouter().CSS(), "fastr-docs-blog-share:focus-visible") && !strings.Contains(NewRouter().CSS(), "[data-fastr-docs-share]:focus-visible") {
			t.Fatal("share controls are invisible to keyboard focus")
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
	t.Run("print declares a font size", func(t *testing.T) {
		css := NewRouter().CSS()
		start := strings.Index(css, "@media print")
		end := start + strings.Index(css[start:], "\n}")
		if !strings.Contains(css[start:end], "font-size") {
			t.Fatal("printed pages keep screen type sizes")
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
	t.Run("the toc select hides without enough headings", func(t *testing.T) {
		r := NewRouter()
		r.MustPage("/g", PageConfig{Title: "G", Description: "d", Order: 1, Source: "## Only\n\nOne heading."})
		html := renderRouterPage(t, r, "/g")
		if strings.Contains(html, "fastr-docs-toc-select") {
			t.Fatal("a one-heading page still renders a toc select")
		}
	})
}
