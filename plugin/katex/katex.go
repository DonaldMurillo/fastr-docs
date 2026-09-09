// Package katex renders TeX math in documentation pages.
//
// Unlike the Mermaid plugin next door, this one runs in the page. No iframe,
// no relaxed policy, nothing sandboxed.
//
// The reason is that KaTeX builds its layout as DOM nodes and assigns sizes
// through CSSOM (node.style.height = "0.68em"), and CSP polices style
// attributes in parsed markup and <style> elements, not CSSOM. So math renders
// under the same default-src 'self' every other page element lives with. The
// second reason is typographic: inline math has to sit on the text baseline
// mid-sentence, and an iframe cannot do that.
//
// The plugin claims two syntaxes. A {{< math >}} shortcode for display math,
// and, unless DisableDollarSyntax is set, the usual dollar delimiters:
// $x^2$ inline and $$...$$ for display. The dollar scanner leaves code fences,
// inline code spans, and escaped dollars alone.
package katex

import (
	"embed"
	"errors"
	"io/fs"
	"net/http"
	"strconv"
	"strings"

	docs "github.com/DonaldMurillo/fastr-docs"
	"github.com/DonaldMurillo/gofastr/core/render"
	gofastrRouter "github.com/DonaldMurillo/gofastr/core/router"
)

//go:embed assets
var assetsFS embed.FS

const (
	// AssetDir is the sub-directory of the docs asset prefix holding the
	// runtime, its stylesheet, and the fonts.
	AssetDir = "katex"
	// LoaderPath is the small script every page loads. It pulls ScriptPath in
	// only on a page that has math.
	LoaderPath = AssetDir + "/katex-loader.js"
	// ScriptPath is the renderer, relative to the docs asset prefix. It is not
	// a page script; the loader fetches it.
	ScriptPath = AssetDir + "/katex.js"
	// StylePath is the KaTeX stylesheet, relative to the docs asset prefix.
	StylePath = AssetDir + "/katex.css"

	// fontMaxAge is a year. The font files are content-addressed by KaTeX's
	// own release, so a stale one is not a risk worth a revalidation round
	// trip on every page.
	fontMaxAge = "public, max-age=31536000, immutable"
)

// Plugin registers the math syntaxes and serves the KaTeX runtime.
//
// Register it like any other plugin:
//
//	router.Use(katex.Plugin{})
//
// Then in Markdown:
//
//	The mass-energy equivalence is $E = mc^2$.
//
//	$$
//	\int_{-\infty}^{\infty} e^{-x^2} dx = \sqrt{\pi}
//	$$
type Plugin struct {
	// Shortcode overrides the display-math shortcode name. Empty registers
	// "math".
	Shortcode string
	// DisableDollarSyntax turns off the $...$ and $$...$$ scanner, leaving the
	// shortcode as the only way to write math. Set it for a site whose prose
	// is full of literal dollar amounts and shell variables.
	DisableDollarSyntax bool
}

func (Plugin) Name() string { return "katex" }

// Apply registers the display-math shortcode and, by default, the dollar
// scanner. The shortcode body is taken unrendered: TeX is riddled with
// characters Markdown would claim, starting with the backslash.
func (p Plugin) Apply(r *docs.Router) error {
	if r == nil {
		return errors.New("katex: router is nil")
	}
	name := strings.TrimSpace(p.Shortcode)
	if name == "" {
		name = "math"
	}
	if err := r.RegisterMarkdownRawComponent(name, func(props map[string]string, raw string) render.HTML {
		// Display is the default; display=false renders inline, so a
		// formula can sit mid-sentence through the shortcode too.
		display := true
		switch strings.ToLower(strings.TrimSpace(props["display"])) {
		case "false", "no", "0", "inline":
			display = false
		}
		return formula(stripFence(raw), display)
	}); err != nil {
		return err
	}
	if p.DisableDollarSyntax {
		return nil
	}
	return r.RegisterMarkdownSourceTransform("katex", extractDollarMath)
}

// RuntimeAssets contributes the runtime, the stylesheet, and the font files, so
// a project does not wire any of them into its own main.go.
func (Plugin) RuntimeAssets() (map[string][]byte, error) {
	assets := map[string][]byte{}
	err := fs.WalkDir(assetsFS, "assets", func(name string, entry fs.DirEntry, err error) error {
		if err != nil || entry.IsDir() {
			return err
		}
		body, readErr := fs.ReadFile(assetsFS, name)
		if readErr != nil {
			return readErr
		}
		assets[AssetDir+strings.TrimPrefix(name, "assets")] = body
		return nil
	})
	if err != nil {
		return nil, err
	}
	return assets, nil
}

// PageScripts is the loader alone, not the renderer.
//
// The renderer is 266KB and most documentation pages have no math on them, so
// loading it everywhere would make it the largest thing on the page and have it
// do nothing. The loader is under a kilobyte, checks for a placeholder, and
// fetches the renderer and its stylesheet only when it finds one.
//
// The stylesheet and the twenty font files are never page scripts either; they
// are served, and the browser fetches each font face only when a formula uses
// it.
func (Plugin) PageScripts() []string { return []string{LoaderPath} }

// MountAssets serves the fonts with a long immutable cache. Router.
// MountRuntimeAssets already serves them; this route takes precedence and only
// changes the caching, since a font re-fetched on every page load is the one
// avoidable cost in this plugin.
func (Plugin) MountAssets(r *docs.Router, httpRouter *gofastrRouter.Router) error {
	if httpRouter == nil {
		return errors.New("katex: MountAssets requires a GoFastr router")
	}
	entries, err := fs.ReadDir(assetsFS, "assets/fonts")
	if err != nil {
		return err
	}
	prefix := r.AssetPrefix() + "/" + AssetDir + "/fonts/"
	for _, entry := range entries {
		body, err := fs.ReadFile(assetsFS, "assets/fonts/"+entry.Name())
		if err != nil {
			return err
		}
		httpRouter.Get(prefix+entry.Name(), http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			w.Header().Set("Content-Type", "font/woff2")
			w.Header().Set("X-Content-Type-Options", "nosniff")
			w.Header().Set("Cache-Control", fontMaxAge)
			_, _ = w.Write(body)
		}))
	}
	return nil
}

// formula renders the placeholder the page runtime turns into typeset math.
//
// The TeX travels in an attribute rather than as element content, so no part of
// it is ever parsed as markup, and the same text is repeated as the element's
// body: without JavaScript a reader sees the source instead of a blank gap.
//
// Display math is a span carrying a modifier class rather than a div, because
// $$...$$ written mid-sentence would otherwise put a block element inside a
// paragraph.
func formula(tex string, display bool) render.HTML {
	tex = strings.TrimSpace(tex)
	if tex == "" {
		return ""
	}
	class := "fastr-docs-math"
	if display {
		class += " fastr-docs-math--display"
	}
	attrs := map[string]string{
		"class":                        class,
		"data-fastr-docs-math":         tex,
		"data-fastr-docs-katex":        "1",
		"aria-label":                   tex,
		"role":                         "math",
		"data-fastr-docs-math-display": strconv.FormatBool(display),
	}
	return render.Tag("span", attrs, render.Text(tex))
}

// stripFence removes an optional fenced-code wrapper, so a {{< math >}} block
// still reads as a code block in a plain Markdown editor.
func stripFence(raw string) string {
	lines := strings.Split(strings.ReplaceAll(raw, "\r\n", "\n"), "\n")
	start, end := 0, len(lines)
	for start < end && strings.TrimSpace(lines[start]) == "" {
		start++
	}
	for end > start && strings.TrimSpace(lines[end-1]) == "" {
		end--
	}
	if start < end && isFence(lines[start]) && isFence(lines[end-1]) {
		start++
		end--
	}
	return strings.Join(lines[start:end], "\n")
}

func isFence(line string) bool {
	trimmed := strings.TrimSpace(line)
	return strings.HasPrefix(trimmed, "```") || strings.HasPrefix(trimmed, "~~~")
}
