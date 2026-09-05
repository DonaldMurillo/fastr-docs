// Package mermaid renders Mermaid diagrams in documentation pages.
//
// Mermaid renders by injecting a <style> element and emitting SVG carrying
// inline styles. A fastr-docs site serves pages under default-src 'self', which
// blocks both, so running Mermaid in the page itself is not an option without
// weakening the policy for every page.
//
// Instead the diagram renders inside a same-origin frame document loaded in an
// <iframe sandbox="allow-scripts"> with no allow-same-origin. Only that
// document carries the relaxed style policy; the pages stay strict. The frame
// has an opaque origin, so it cannot reach the host's DOM, cookies, or storage,
// and postMessage is the only channel between them.
//
// This follows the isolation model of the editor plugin in gofastr-plugins,
// reduced to what a documentation page needs: no save path, no capability
// handshake, no document store. The host sends a diagram, the frame reports its
// height.
package mermaid

import (
	"embed"
	"errors"
	"io/fs"
	"net/http"
	"strings"

	docs "github.com/DonaldMurillo/fastr-docs"
	"github.com/DonaldMurillo/gofastr/core/render"
	gofastrRouter "github.com/DonaldMurillo/gofastr/core/router"
)

//go:embed assets
var assetsFS embed.FS

//go:embed host/adapter.js
var adapterJS []byte

const (
	// AssetDir is the sub-directory of the docs asset prefix that holds the
	// frame document and its bundle.
	AssetDir = "mermaid"
	// FramePath is the frame document, relative to the docs asset prefix.
	FramePath = AssetDir + "/diagram.html"
	// AdapterPath is the host-side bridge, relative to the docs asset prefix.
	AdapterPath = AssetDir + "/adapter.js"

	// framePolicy is the only place the strict policy is relaxed, and it
	// applies to the frame document alone. Mermaid needs inline styles;
	// frame-ancestors keeps the document from being embedded off-site.
	framePolicy = "default-src 'self'; img-src 'self' data:; style-src 'self' 'unsafe-inline'; " +
		"script-src 'self'; font-src 'self' data:; connect-src 'none'; object-src 'none'; " +
		"base-uri 'none'; form-action 'none'; frame-ancestors 'self'"
)

// Plugin registers the `mermaid` shortcode and serves the sandboxed renderer.
//
// Register it like any other plugin:
//
//	router.Use(mermaid.Plugin{})
//
// Then in Markdown:
//
//	{{< mermaid >}}
//	```
//	graph TD
//	    A --> B
//	```
//	{{< /mermaid >}}
type Plugin struct {
	// Name overrides the shortcode name. Empty registers "mermaid".
	Shortcode string
	// FrameSrc overrides the URL of the frame document. Empty derives it from
	// the Router's asset prefix, which is what a generated project wants.
	FrameSrc string
}

func (Plugin) Name() string { return "mermaid" }

// Apply registers the shortcode. The body is taken unrendered, because a
// diagram is data: rendering it as Markdown would destroy its line structure,
// which is the whole syntax.
func (p Plugin) Apply(r *docs.Router) error {
	if r == nil {
		return errors.New("mermaid: router is nil")
	}
	name := strings.TrimSpace(p.Shortcode)
	if name == "" {
		name = "mermaid"
	}
	frameSrc := strings.TrimSpace(p.FrameSrc)
	if frameSrc == "" {
		frameSrc = r.AssetPrefix() + "/" + FramePath
	}
	return r.RegisterMarkdownRawComponent(name, func(props map[string]string, raw string) render.HTML {
		return diagram(raw, frameSrc, props["title"])
	})
}

// RuntimeAssets contributes the frame document, its bundle, and the host
// adapter, so a project does not wire any of them into its own main.go.
func (Plugin) RuntimeAssets() (map[string][]byte, error) {
	assets := map[string][]byte{AdapterPath: adapterJS}
	entries, err := fs.ReadDir(assetsFS, "assets")
	if err != nil {
		return nil, err
	}
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		body, err := fs.ReadFile(assetsFS, "assets/"+entry.Name())
		if err != nil {
			return nil, err
		}
		assets[AssetDir+"/"+entry.Name()] = body
	}
	return assets, nil
}

// MountAssets serves the three framed files with the headers the sandbox
// requires. Router.MountRuntimeAssets already serves everything the plugin
// contributed; these routes take precedence for the files the frame loads.
//
// A frame with no allow-same-origin has an opaque origin, so its own
// stylesheet and bundle are cross-origin requests as far as the browser is
// concerned. Without Cross-Origin-Resource-Policy: cross-origin they are
// blocked with ERR_BLOCKED_BY_RESPONSE.NotSameOrigin and the frame renders
// nothing. The host adapter deliberately does not carry that header: it is an
// ordinary same-origin script on the page.
func (p Plugin) MountAssets(r *docs.Router, httpRouter *gofastrRouter.Router) error {
	if httpRouter == nil {
		return errors.New("mermaid: MountAssets requires a GoFastr router")
	}
	prefix := r.AssetPrefix() + "/" + AssetDir + "/"
	for _, framed := range []struct{ file, contentType string }{
		{"diagram.html", "text/html; charset=utf-8"},
		{"diagram.css", "text/css; charset=utf-8"},
		{"diagram.js", "text/javascript; charset=utf-8"},
	} {
		body, err := fs.ReadFile(assetsFS, "assets/"+framed.file)
		if err != nil {
			return err
		}
		contentType := framed.contentType
		isDocument := strings.HasSuffix(framed.file, ".html")
		httpRouter.Get(prefix+framed.file, http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			w.Header().Set("Content-Type", contentType)
			w.Header().Set("X-Content-Type-Options", "nosniff")
			w.Header().Set("Cross-Origin-Resource-Policy", "cross-origin")
			if isDocument {
				w.Header().Set("Content-Security-Policy", framePolicy)
			}
			_, _ = w.Write(body)
		}))
	}
	return nil
}

// diagram renders the placeholder the host adapter turns into an iframe. The
// source is carried in an attribute rather than as markup, so nothing in it is
// ever parsed as HTML by the page.
func diagram(raw, frameSrc, title string) render.HTML {
	source := stripFence(raw)
	if strings.TrimSpace(source) == "" {
		return ""
	}
	label := strings.TrimSpace(title)
	if label == "" {
		label = "Diagram"
	}
	attrs := map[string]string{
		"class":                         "fastr-docs-mermaid",
		"data-fastr-docs-mermaid":       source,
		"data-fastr-docs-mermaid-frame": frameSrc,
		"data-fastr-docs-mermaid-title": label,
	}
	// Without JavaScript the source stays visible as preformatted text, which
	// is more useful than an empty box.
	return render.Tag("div", attrs, render.Tag("pre", map[string]string{"class": "fastr-docs-mermaid__source"}, render.Text(source)))
}

// stripFence removes an optional fenced-code wrapper, so a diagram still reads
// as a code block in a plain Markdown editor.
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
