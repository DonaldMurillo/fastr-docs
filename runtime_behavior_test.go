package docs

// docs.js behavior contracts: anchors, search, palette focus,
// theme color, scroll restoration, and announcements.
import (
	"strings"
	"testing"
)

func TestRuntimeBehavior(t *testing.T) {
	js := RuntimeJS()
	t.Run("heading anchors are focusable and named", func(t *testing.T) {
		r := NewRouter()
		r.MustPage("/g", PageConfig{Title: "G", Description: "d", Order: 1, Source: "## Sección uno\n\nText."})
		html := r2Render(t, r, "/g")
		index := strings.Index(html, "heading-anchor")
		if index < 0 {
			t.Fatal("no heading anchor rendered")
		}
		chunk := html[index:min(index+220, len(html))]
		if strings.Contains(chunk, "aria-hidden") || !strings.Contains(chunk, "aria-label") {
			t.Fatalf("anchor is hidden from keyboards: %s", chunk)
		}
	})

	t.Run("json search input is debounced", func(t *testing.T) {
		if !strings.Contains(js, "debounce") {
			t.Fatal("every keystroke fetches; the search input has no debounce")
		}
	})

	t.Run("json results highlight matched terms", func(t *testing.T) {
		if !strings.Contains(js, "<mark>") {
			t.Fatal("result rows do not highlight the terms that matched")
		}
	})

	t.Run("pagefind urls pass through the base", func(t *testing.T) {
		start := strings.Index(js, "function renderPagefindResults")
		end := strings.Index(js[start:], "function renderJSONResults") + start
		if !strings.Contains(js[start:end], "withBase(") {
			t.Fatal("pagefind result urls navigate without the export base")
		}
	})

	t.Run("heading anchors copy rather than navigate", func(t *testing.T) {
		if !strings.Contains(js, "heading-anchor") || !strings.Contains(js, "clipboard") {
			t.Fatal("no runtime binds the heading anchors to the clipboard")
		}
	})

	t.Run("the toc select follows a deep-linked hash", func(t *testing.T) {
		start := strings.Index(js, "function initTocSelect")
		end := strings.Index(js[start:], "function docsBase") + start
		if !strings.Contains(js[start:end], "location.hash") {
			t.Fatal("landing on #section leaves the toc select on the first entry")
		}
	})

	t.Run("the palette gives focus back to its trigger", func(t *testing.T) {
		start := strings.Index(js, "watchPaletteForLocalization")
		end := strings.Index(js[start:], "function pagefindList") + start
		if !strings.Contains(js[start:end], "focus()") {
			t.Fatal("closing the palette strands focus at the top of the document")
		}
	})

	t.Run("the drawer select syncs the active version", func(t *testing.T) {
		js := RuntimeJS()
		start := strings.Index(js, "function syncSectionSelect")
		end := strings.Index(js[start:], "function bindSectionSelect") + start
		if !strings.Contains(js[start:end], "version") {
			t.Fatal("the section select marks the first section on versioned pages")
		}
	})

	t.Run("theme color follows the page", func(t *testing.T) {
		if !strings.Contains(RuntimeJS(), "theme-color") {
			t.Fatal("the runtime never touches meta theme-color; browser chrome keeps the first theme")
		}
	})

	t.Run("scroll position survives back navigation", func(t *testing.T) {
		if !strings.Contains(RuntimeJS(), "scrollRestoration") {
			t.Fatal("SPA navigation drops the reader back at the top of long pages")
		}
	})

	t.Run("the blog search folds accents", func(t *testing.T) {
		js := RuntimeJS()
		start := strings.Index(js, "function blogSearchTerms")
		end := start + strings.Index(js[start:], "}")
		if !strings.Contains(js[start:end], ".normalize(") {
			t.Fatal(`the blog filter cannot find "diseno" under the tag written "diseño"`)
		}
	})

	t.Run("stale json searches are aborted", func(t *testing.T) {
		if !strings.Contains(RuntimeJS(), "AbortController") {
			t.Fatal("a slow earlier query can overwrite a later result")
		}
	})

	t.Run("the blog search announces its result count", func(t *testing.T) {
		js := RuntimeJS()
		start := strings.Index(js, "function initBlogSearch")
		end := strings.Index(js[start:], "function blogShareURL") + start
		if start < 0 || !strings.Contains(js[start:end], "announce") {
			t.Fatal("filtering the blog list is silent; screen readers hear nothing about the result count")
		}
	})
}
