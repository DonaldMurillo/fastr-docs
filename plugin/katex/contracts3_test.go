package katex

// Math contracts from the third audit cycle: shortcode props, CRLF
// bodies, tables, semantics, and the loader.

import (
	"context"
	"strings"
	"testing"

	docs "github.com/DonaldMurillo/fastr-docs"
	uiapp "github.com/DonaldMurillo/gofastr/core-ui/app"
)

func renderPage(t *testing.T, r *docs.Router, path string) string {
	t.Helper()
	site := uiapp.NewApp("T")
	if err := r.Mount(site, r.Layout()); err != nil {
		t.Fatal(err)
	}
	html, err := site.RenderPage(context.Background(), path)
	if err != nil {
		t.Fatal(err)
	}
	return string(html)
}

func TestRed283To287(t *testing.T) {
	t.Run("the math shortcode honors a display prop", func(t *testing.T) {
		r := mathRouter(t, Plugin{})
		html := renderPage(t, r, "/m")
		if !strings.Contains(html, `data-fastr-docs-math-display="false"`) {
			t.Fatal("an inline math shortcode still renders as display")
		}
	})
	t.Run("crlf shortcode bodies still render", func(t *testing.T) {
		r := docs.NewRouter()
		if err := r.Use(Plugin{}); err != nil {
			t.Fatal(err)
		}
		if err := r.Page("/m", docs.PageConfig{Title: "M", Description: "d", Order: 1, Source: "{{< math >}}\r\nx^2\r\n{{< /math >}}"}); err != nil {
			t.Fatalf("a CRLF-wrapped shortcode body fails to register: %v", err)
		}
		html := renderPage(t, r, "/m")
		if !strings.Contains(html, "fastr-docs-math") || !strings.Contains(html, "x^2") {
			t.Fatal("a CRLF-wrapped shortcode body loses its formula")
		}
	})
	t.Run("math renders inside table cells", func(t *testing.T) {
		r := docs.NewRouter()
		if err := r.Use(Plugin{}); err != nil {
			t.Fatal(err)
		}
		r.MustPage("/m", docs.PageConfig{Title: "M", Description: "d", Order: 1, Source: "| a | b |\n| --- | --- |\n| $x$ | y |"})
		html := renderPage(t, r, "/m")
		if !strings.Contains(html, "fastr-docs-math") {
			t.Fatal("inline math inside a table cell never renders")
		}
	})
	t.Run("formulas carry a math role", func(t *testing.T) {
		r := docs.NewRouter()
		if err := r.Use(Plugin{}); err != nil {
			t.Fatal(err)
		}
		r.MustPage("/m", docs.PageConfig{Title: "M", Description: "d", Order: 1, Source: "$x^2$"})
		html := renderPage(t, r, "/m")
		if !strings.Contains(html, `role="math"`) {
			t.Fatal("the placeholder span does not announce itself as math")
		}
	})
	t.Run("the loader announces failures inline", func(t *testing.T) {
		assets, err := Plugin{}.RuntimeAssets()
		if err != nil {
			t.Fatal(err)
		}
		loader := string(assets[LoaderPath])
		if !strings.Contains(loader, "aria") {
			t.Fatal("a failed render leaves a silent placeholder")
		}
	})
}

func mathRouter(t *testing.T, p Plugin) *docs.Router {
	t.Helper()
	r := docs.NewRouter()
	if err := r.Use(p); err != nil {
		t.Fatal(err)
	}
	r.MustPage("/m", docs.PageConfig{Title: "M", Description: "d", Order: 1,
		Source: "{{< math display=false >}}x^2{{< /math >}}"})
	return r
}
