package docs

// Markdown and shortcode contracts: plain toc titles,
// validated props and tab labels, and case-folded names.
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
		r := NewRouter()
		r.MustPage("/g", PageConfig{Title: "G", Description: "d", Order: 1,
			Source: "{{< callout Variant=\"warning\" >}}x{{< /callout >}}"})
		html := r2Render(t, r, "/g")
		if !strings.Contains(html, "warning") {
			t.Fatal("a capitalized prop name is silently ignored")
		}
	})
	t.Run("h4 headings carry anchors", func(t *testing.T) {
		r := NewRouter()
		r.MustPage("/g", PageConfig{Title: "G", Description: "d", Order: 1, Source: "#### Deep section\n\nText."})
		html := r2Render(t, r, "/g")
		if !strings.Contains(html, "heading-anchor") {
			t.Fatal("h4 joins the toc but cannot be linked")
		}
	})
}
