package docs

// Markdown and shortcodes: plain toc titles, validated props and tab
// labels, and case-folded names.
import (
	"strings"
	"testing"
)

func TestMarkdownShortcodes(t *testing.T) {
	t.Run("toc titles drop shortcode markers", func(t *testing.T) {
		for _, heading := range markdownHeadings("## Guía {{< callout >}}x{{< /callout >}}\n\nbody") {
			if strings.Contains(heading.Title, "{{<") {
				t.Fatalf("toc title carries shortcode syntax: %q", heading.Title)
			}
		}
	})

	t.Run("unknown shortcode props are named", func(t *testing.T) {
		r := NewRouter()
		r.MustPage("/g", PageConfig{Title: "G", Description: "d", Order: 1, Source: "{{< callout variant=\"nonsense\" >}}x{{< /callout >}}"})
		if err := r.Validate(); err == nil || !strings.Contains(err.Error(), "nonsense") {
			t.Fatalf("Validate() = %v, want the unknown prop named", err)
		}
	})

	t.Run("duplicate tab labels are named", func(t *testing.T) {
		r := NewRouter()
		r.MustPage("/g", PageConfig{Title: "G", Description: "d", Order: 1,
			Source: "{{< tabs >}}{{< tab label=\"Same\" >}}a{{< /tab >}}{{< tab label=\"Same\" >}}b{{< /tab >}}{{< /tabs >}}"})
		if err := r.Validate(); err == nil || !strings.Contains(err.Error(), "tab") {
			t.Fatalf("Validate() = %v, want a duplicate tab label named", err)
		}
	})

	t.Run("shortcode prop names fold case", func(t *testing.T) {
		html := renderScratchPage(t, "{{< callout Variant=\"warning\" >}}x{{< /callout >}}")
		if !strings.Contains(html, "warning") {
			t.Fatal("a capitalized prop name is silently ignored")
		}
	})
	t.Run("h4 headings carry anchors", func(t *testing.T) {
		html := renderScratchPage(t, "#### Deep section\n\nText.")
		if !strings.Contains(html, "heading-anchor") {
			t.Fatal("h4 joins the toc but cannot be linked")
		}
	})
	t.Run("toc titles drop math delimiters", func(t *testing.T) {
		for _, heading := range markdownHeadings("## The $x^2$ rule\n\nbody") {
			if strings.Contains(heading.Title, "$") {
				t.Fatalf("toc title carries math delimiters: %q", heading.Title)
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
	t.Run("h4-only pages still get a toc select", func(t *testing.T) {
		r := NewRouter()
		r.MustPage("/p", PageConfig{Title: "T", Description: "d", Order: 1, Source: "# Title\n\n#### Deep one\n\n#### Deep two"})
		if !strings.Contains(renderRouterPage(t, r, "/p"), "data-docs-toc-select") {
			t.Fatal("h4-only page has no toc select")
		}
	})
	t.Run("headings carry copyable anchor buttons", func(t *testing.T) {
		r := NewRouter()
		r.MustPage("/p", PageConfig{Title: "T", Description: "d", Order: 1, Source: "# T\n\n## Section"})
		if !strings.Contains(renderRouterPage(t, r, "/p"), "heading-anchor") {
			t.Fatal("no anchor affordance on headings")
		}
	})
	t.Run("heading ids fold apostrophes", func(t *testing.T) {
		html := renderScratchPage(t, "## What's new\n\nText.")
		if !strings.Contains(html, `id="whats-new"`) {
			t.Fatal("an apostrophe in a heading poisons its id")
		}
	})
}
