package docs

// docs.js behavior: anchors, search, palette focus, theme color, scroll
// restoration, accessibility state, and announcements.
import (
	"strings"
	"testing"
)

func TestRuntimeBehavior(t *testing.T) {
	js := RuntimeJS()
	t.Run("heading anchors are focusable and named", func(t *testing.T) {
		html := renderScratchPage(t, "## Sección uno\n\nText.")
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
		start := strings.Index(js, "function syncSectionSelect")
		end := strings.Index(js[start:], "function bindSectionSelect") + start
		if !strings.Contains(js[start:end], "version") {
			t.Fatal("the section select marks the first section on versioned pages")
		}
	})

	t.Run("theme color follows the page", func(t *testing.T) {
		if !strings.Contains(js, "theme-color") {
			t.Fatal("the runtime never touches meta theme-color; browser chrome keeps the first theme")
		}
	})

	t.Run("scroll position survives back navigation", func(t *testing.T) {
		if !strings.Contains(js, "scrollRestoration") {
			t.Fatal("SPA navigation drops the reader back at the top of long pages")
		}
	})

	t.Run("the blog search folds accents", func(t *testing.T) {
		start := strings.Index(js, "function blogSearchTerms")
		end := start + strings.Index(js[start:], "}")
		if !strings.Contains(js[start:end], ".normalize(") {
			t.Fatal(`the blog filter cannot find "diseno" under the tag written "diseño"`)
		}
	})

	t.Run("stale json searches are aborted", func(t *testing.T) {
		if !strings.Contains(js, "AbortController") {
			t.Fatal("a slow earlier query can overwrite a later result")
		}
	})

	t.Run("the blog search announces its result count", func(t *testing.T) {
		start := strings.Index(js, "function initBlogSearch")
		end := strings.Index(js[start:], "function blogShareURL") + start
		if start < 0 || !strings.Contains(js[start:end], "announce") {
			t.Fatal("filtering the blog list is silent; screen readers hear nothing about the result count")
		}
	})

	t.Run("the variant crossing documents its full load", func(t *testing.T) {
		// Deliberate design: a language change swaps the mounted content
		// slice, so the selector rebuilds the shell with location.href
		// instead of leaving stale layout layers around new content. The
		// comment in runtime.go says so; a client-side crossing needs
		// gofastr #408 (per-language layout keys) first.
		if !strings.Contains(js, "rebuild the shell from the destination URL") {
			t.Fatal("the full-load rationale comment was lost from initVariantSelectors")
		}
	})
	t.Run("sidebar sync covers blog drawers", func(t *testing.T) {
		if !strings.Contains(js, `data-fui-widget^="fastr-docs-blog"`) {
			t.Fatal("syncDocsSidebars ignores blog drawer bodies")
		}
	})
	t.Run("chrome sync carries document direction", func(t *testing.T) {
		if !strings.Contains(js, `setAttribute('dir'`) {
			t.Fatal("syncChrome never touches dir")
		}
	})
	t.Run("the toc select respects reduced motion", func(t *testing.T) {
		if !strings.Contains(js, "prefers-reduced-motion") {
			t.Fatal("runtime never checks prefers-reduced-motion")
		}
	})
	t.Run("the section select api is exported", func(t *testing.T) {
		if !strings.Contains(js, "fastrDocs.initSectionSelects") {
			t.Fatal("initSectionSelects is not on window.fastrDocs")
		}
	})
	t.Run("the drawer trigger manages aria-expanded", func(t *testing.T) {
		if !strings.Contains(js, "trigger.setAttribute('aria-expanded'") {
			t.Fatal("the drawer trigger's own aria-expanded is never managed")
		}
	})
	t.Run("path normalization decodes percent escapes", func(t *testing.T) {
		if !strings.Contains(js, "decodeURIComponent") {
			t.Fatal("normalizeDocsPath does not decode")
		}
	})
	t.Run("search results announce their count", func(t *testing.T) {
		if !strings.Contains(js, `role="status"`) && !strings.Contains(js, "role='status'") && !strings.Contains(js, "'role', 'status'") {
			t.Fatal("no live region announcing result counts")
		}
	})
	t.Run("the palette modal aria-label is localized", func(t *testing.T) {
		if !strings.Contains(js, "modal.setAttribute('aria-label'") && !strings.Contains(js, "palette.setAttribute('aria-label'") {
			t.Fatal("the palette modal's aria-label is never localized")
		}
	})
	t.Run("focus returns to the trigger after close", func(t *testing.T) {
		if !strings.Contains(js, "trigger.focus()") {
			t.Fatal("focus never returns to the drawer trigger")
		}
	})
	t.Run("the toc select remembers the choice", func(t *testing.T) {
		if !strings.Contains(js, "sessionStorage") && !strings.Contains(js, "localStorage") {
			t.Fatal("no persistence of the reader's toc position")
		}
	})
	t.Run("palette options name their section", func(t *testing.T) {
		if !strings.Contains(js, "data-fastr-docs-section") {
			t.Fatal("results carry no section marker for grouping")
		}
	})
	t.Run("the palette empty state is localized", func(t *testing.T) {
		if !strings.Contains(js, "data-fastr-docs-search-empty") {
			t.Fatal("the empty result row has no label hook to localize through")
		}
	})
	t.Run("the theme storage key is namespaced", func(t *testing.T) {
		if !strings.Contains(js, "fastr-docs-theme") {
			t.Fatal("the runtime watches for theme changes by any key containing 'theme'")
		}
	})
	t.Run("theme changes re-sync the chrome color", func(t *testing.T) {
		if !strings.Contains(js, ":theme") {
			t.Fatal("a theme change never reaches the browser chrome color")
		}
	})
	t.Run("scroll listeners are passive", func(t *testing.T) {
		if !strings.Contains(js, "passive:") {
			t.Fatal("scroll and touch listeners block the main thread")
		}
	})
	t.Run("the blog search input is debounced", func(t *testing.T) {
		start := strings.Index(js, "function initBlogSearch")
		end := strings.Index(js[start:], "function blogShareURL") + start
		if start < 0 || !strings.Contains(js[start:end], "setTimeout") {
			t.Fatal("every keystroke re-filters and re-announces the whole list")
		}
	})
	t.Run("the openapi console marks busy in markup", func(t *testing.T) {
		if !strings.Contains(js, "aria-busy") {
			t.Fatal("the docs runtime never marks its surfaces busy")
		}
	})
	t.Run("the runtime prepares overlays for print", func(t *testing.T) {
		if !strings.Contains(js, "beforeprint") {
			t.Fatal("printing leaves drawers and the palette open over the page")
		}
	})
	t.Run("the runtime exposes its base helpers", func(t *testing.T) {
		if !strings.Contains(js, "fastrDocs.withBase") || !strings.Contains(js, "fastrDocs.base") {
			t.Fatal("plugins cannot share the runtime's base-path helpers")
		}
	})
	t.Run("the runtime exposes path normalization", func(t *testing.T) {
		if !strings.Contains(js, "fastrDocs.normalizePath") {
			t.Fatal("plugins cannot compare route paths the way the runtime does")
		}
	})
}
