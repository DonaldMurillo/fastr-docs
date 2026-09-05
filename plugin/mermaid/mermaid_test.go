package mermaid

import (
	"io/fs"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	docs "github.com/DonaldMurillo/fastr-docs"
	gofastrRouter "github.com/DonaldMurillo/gofastr/core/router"
)

func routerWithPlugin(t *testing.T, plugin Plugin) *docs.Router {
	t.Helper()
	r := docs.NewRouter(docs.WithSiteName("Diagrams"))
	if err := r.Use(plugin); err != nil {
		t.Fatalf("Use() error = %v", err)
	}
	return r
}

// renderDiagram invokes the shortcode the plugin registered, which is exactly
// what a page render calls.
func renderDiagram(t *testing.T, r *docs.Router, props map[string]string, raw string) string {
	t.Helper()
	component := r.MarkdownRawComponents()["mermaid"]
	if component == nil {
		t.Fatal("plugin did not register the mermaid shortcode")
	}
	return string(component(props, raw))
}

func TestShortcodeEmitsAPlaceholderCarryingTheSource(t *testing.T) {
	r := routerWithPlugin(t, Plugin{})
	html := renderDiagram(t, r, map[string]string{"title": "Flow"}, "\n```\ngraph TD\n    A --> B\n```\n")

	for _, want := range []string{
		`class="fastr-docs-mermaid"`,
		`data-fastr-docs-mermaid-title="Flow"`,
		"/__fastr-docs/mermaid/diagram.html",
		"graph TD",
	} {
		if !strings.Contains(html, want) {
			t.Fatalf("output missing %q: %s", want, html)
		}
	}
	// The fence is a readability wrapper, not part of the diagram.
	if strings.Contains(html, "```") {
		t.Fatalf("fence markers survived into the diagram source: %s", html)
	}
	// The page must not render an iframe itself; the adapter creates it, so a
	// reader without JavaScript still sees the source.
	if strings.Contains(html, "<iframe") {
		t.Fatalf("plugin emitted an iframe server-side: %s", html)
	}
	if !strings.Contains(html, "fastr-docs-mermaid__source") {
		t.Fatalf("no no-JavaScript fallback: %s", html)
	}
}

func TestShortcodeFollowsACustomAssetPrefix(t *testing.T) {
	r := docs.NewRouter(docs.WithSearchIndexPath("/__custom/search.json"))
	if err := r.Use(Plugin{}); err != nil {
		t.Fatalf("Use() error = %v", err)
	}
	html := renderDiagram(t, r, nil, "graph TD\n A-->B")
	if !strings.Contains(html, "/__custom/mermaid/diagram.html") {
		t.Fatalf("frame src did not follow the asset prefix: %s", html)
	}
}

func TestEmptyDiagramRendersNothing(t *testing.T) {
	r := routerWithPlugin(t, Plugin{})
	html := renderDiagram(t, r, nil, "\n```\n\n```\n")
	if strings.Contains(html, "fastr-docs-mermaid") {
		t.Fatalf("empty diagram produced a placeholder: %s", html)
	}
}

func TestRuntimeAssetsCarryTheFrameAndAdapter(t *testing.T) {
	assets, err := Plugin{}.RuntimeAssets()
	if err != nil {
		t.Fatalf("RuntimeAssets() error = %v", err)
	}
	for _, want := range []string{"mermaid/diagram.html", "mermaid/diagram.css", "mermaid/diagram.js", "mermaid/adapter.js"} {
		if len(assets[want]) == 0 {
			t.Fatalf("RuntimeAssets() missing or empty %q", want)
		}
	}
	// The bundle carries Mermaid itself; a stub would defeat the point.
	if len(assets["mermaid/diagram.js"]) < 500_000 {
		t.Fatalf("bundle is %d bytes, too small to contain Mermaid", len(assets["mermaid/diagram.js"]))
	}
	// Nothing may be fetched at runtime: the frame is an opaque origin under
	// script-src 'self'.
	if strings.Contains(string(assets["mermaid/diagram.html"]), "http://") ||
		strings.Contains(string(assets["mermaid/diagram.html"]), "https://") {
		t.Fatal("frame document references an absolute URL")
	}
}

// A frame with no allow-same-origin has an opaque origin, so its own bundle and
// stylesheet are cross-origin requests. Without this header the browser blocks
// them and the diagram silently never renders.
func TestFramedAssetsAreServedCrossOriginAndTheAdapterIsNot(t *testing.T) {
	r := routerWithPlugin(t, Plugin{})
	httpRouter := gofastrRouter.New()
	if err := r.MountRuntimeAssets(httpRouter, r.AssetPrefix(), ""); err != nil {
		t.Fatalf("MountRuntimeAssets() error = %v", err)
	}
	if err := r.MountPluginAssets(httpRouter); err != nil {
		t.Fatalf("MountPluginAssets() error = %v", err)
	}
	server := httptest.NewServer(httpRouter)
	defer server.Close()

	for _, path := range []string{"/diagram.html", "/diagram.css", "/diagram.js"} {
		response, err := http.Get(server.URL + "/__fastr-docs/mermaid" + path)
		if err != nil {
			t.Fatalf("GET %s: %v", path, err)
		}
		body := response.Header
		_ = response.Body.Close()
		if response.StatusCode != http.StatusOK {
			t.Fatalf("GET %s = %d", path, response.StatusCode)
		}
		if got := body.Get("Cross-Origin-Resource-Policy"); got != "cross-origin" {
			t.Fatalf("%s CORP = %q, want cross-origin", path, got)
		}
	}

	response, err := http.Get(server.URL + "/__fastr-docs/mermaid/adapter.js")
	if err != nil {
		t.Fatalf("GET adapter.js: %v", err)
	}
	defer response.Body.Close()
	if got := response.Header.Get("Cross-Origin-Resource-Policy"); got == "cross-origin" {
		t.Fatal("the host adapter must not be served cross-origin; it is an ordinary page script")
	}
}

// The relaxation Mermaid needs must reach the frame document and nothing else.
func TestFrameDocumentCarriesItsOwnRelaxedPolicy(t *testing.T) {
	r := routerWithPlugin(t, Plugin{})
	httpRouter := gofastrRouter.New()
	if err := r.MountPluginAssets(httpRouter); err != nil {
		t.Fatalf("MountPluginAssets() error = %v", err)
	}
	server := httptest.NewServer(httpRouter)
	defer server.Close()

	response, err := http.Get(server.URL + "/__fastr-docs/mermaid/diagram.html")
	if err != nil {
		t.Fatalf("GET diagram.html: %v", err)
	}
	defer response.Body.Close()
	policy := response.Header.Get("Content-Security-Policy")
	for _, want := range []string{"style-src 'self' 'unsafe-inline'", "script-src 'self'", "frame-ancestors 'self'"} {
		if !strings.Contains(policy, want) {
			t.Fatalf("frame policy missing %q: %q", want, policy)
		}
	}
	// The frame has no reason to reach the network or be reframed off-site.
	for _, unwanted := range []string{"script-src 'self' 'unsafe-inline'", "frame-ancestors *"} {
		if strings.Contains(policy, unwanted) {
			t.Fatalf("frame policy is too permissive (%q): %q", unwanted, policy)
		}
	}
	if !strings.Contains(policy, "connect-src 'none'") {
		t.Fatalf("frame should not be able to reach the network: %q", policy)
	}

	// The page policy is untouched by any of this.
	if page := docs.ContentSecurityPolicy(); strings.Contains(page, "unsafe-inline") {
		t.Fatalf("page policy was relaxed: %q", page)
	}
}

func TestStripFenceKeepsDiagramLines(t *testing.T) {
	for name, testCase := range map[string]struct{ in, want string }{
		"fenced":   {"\n```\ngraph TD\n  A-->B\n```\n", "graph TD\n  A-->B"},
		"unfenced": {"graph TD\n  A-->B", "graph TD\n  A-->B"},
		"tilde":    {"~~~\ngraph TD\n~~~", "graph TD"},
		"empty":    {"\n\n", ""},
	} {
		t.Run(name, func(t *testing.T) {
			if got := stripFence(testCase.in); got != testCase.want {
				t.Fatalf("stripFence() = %q, want %q", got, testCase.want)
			}
		})
	}
}

func TestBundledAssetsAreEmbedded(t *testing.T) {
	for _, name := range []string{"assets/diagram.html", "assets/diagram.css", "assets/diagram.js"} {
		if _, err := fs.Stat(assetsFS, name); err != nil {
			t.Fatalf("%s is not embedded: %v", name, err)
		}
	}
	if len(adapterJS) == 0 {
		t.Fatal("host adapter is not embedded")
	}
}
