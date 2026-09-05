package katex

import (
	"io/fs"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	docs "github.com/DonaldMurillo/fastr-docs"
	"github.com/DonaldMurillo/gofastr/core/render"
	gofastrRouter "github.com/DonaldMurillo/gofastr/core/router"
)

func routerWithPlugin(t *testing.T, plugin Plugin) *docs.Router {
	t.Helper()
	r := docs.NewRouter(docs.WithSiteName("Math"))
	if err := r.Use(plugin); err != nil {
		t.Fatalf("Use() error = %v", err)
	}
	return r
}

// place stands in for the marker allocator the docs package passes to a source
// transform, recording what the transform lifted out.
func place(collected *[]string) func(render.HTML) string {
	return func(html render.HTML) string {
		*collected = append(*collected, string(html))
		return "MARK" + string(rune('0'+len(*collected)-1))
	}
}

func TestShortcodeEmitsDisplayMath(t *testing.T) {
	r := routerWithPlugin(t, Plugin{})
	component := r.MarkdownRawComponents()["math"]
	if component == nil {
		t.Fatal("plugin did not register the math shortcode")
	}
	html := string(component(nil, "\n```\nE = mc^2\n```\n"))

	for _, want := range []string{
		`class="fastr-docs-math fastr-docs-math--display"`,
		`data-fastr-docs-math="E = mc^2"`,
		`data-fastr-docs-math-display="true"`,
	} {
		if !strings.Contains(html, want) {
			t.Fatalf("output missing %q: %s", want, html)
		}
	}
	// The fence is a readability wrapper, not part of the formula.
	if strings.Contains(html, "```") {
		t.Fatalf("fence markers survived into the formula: %s", html)
	}
	// Without JavaScript the source stays readable rather than leaving a gap.
	if !strings.Contains(html, ">E = mc^2<") {
		t.Fatalf("no no-JavaScript fallback: %s", html)
	}
}

func TestDollarSyntaxIsRegisteredUnlessDisabled(t *testing.T) {
	if names := routerWithPlugin(t, Plugin{}).MarkdownSourceTransformNames(); len(names) != 1 || names[0] != "katex" {
		t.Fatalf("transforms = %v, want [katex]", names)
	}
	if names := routerWithPlugin(t, Plugin{DisableDollarSyntax: true}).MarkdownSourceTransformNames(); len(names) != 0 {
		t.Fatalf("transforms = %v, want none", names)
	}
}

func TestInlineAndDisplayDollarMath(t *testing.T) {
	var lifted []string
	source := "Energy is $E = mc^2$ exactly.\n\n$$\na^2 + b^2 = c^2\n$$\n\nAnd $$x_1$$ inline."
	out := extractDollarMath(source, place(&lifted))

	if len(lifted) != 3 {
		t.Fatalf("lifted %d formulas, want 3: %v", len(lifted), lifted)
	}
	if strings.Contains(out, "$") {
		t.Fatalf("a dollar delimiter survived: %s", out)
	}
	if !strings.Contains(lifted[0], `data-fastr-docs-math-display="false"`) {
		t.Fatalf("inline math was marked as display: %s", lifted[0])
	}
	if !strings.Contains(lifted[1], "a^2 + b^2 = c^2") {
		t.Fatalf("block math body is wrong: %s", lifted[1])
	}
	if !strings.Contains(lifted[1], `data-fastr-docs-math-display="true"`) {
		t.Fatalf("block math was not marked as display: %s", lifted[1])
	}
	if !strings.Contains(lifted[2], `data-fastr-docs-math-display="true"`) {
		t.Fatalf("inline $$ was not marked as display: %s", lifted[2])
	}
	// The surrounding prose is untouched.
	if !strings.HasPrefix(out, "Energy is MARK0 exactly.") {
		t.Fatalf("prose around the formula changed: %s", out)
	}
}

// Prose is full of dollar signs that are not math, and claiming one silently
// deletes the reader's text.
func TestDollarsInProseAreLeftAlone(t *testing.T) {
	for name, source := range map[string]string{
		"price range":      "Plans run from $5 to $10 a month.",
		"space after":      "Costs $ 5 in total.",
		"space before":     "Pay $5 or $ 6.",
		"single dollar":    "It costs $5.",
		"escaped":          `A literal \$5 stays.`,
		"shell variable":   "Set $HOME and $PATH first.",
		"empty delimiters": "Nothing here: $$.",
	} {
		t.Run(name, func(t *testing.T) {
			var lifted []string
			if out := extractDollarMath(source, place(&lifted)); out != source || len(lifted) != 0 {
				t.Fatalf("claimed prose as math: %q -> %q (%v)", source, out, lifted)
			}
		})
	}
}

func TestCodeIsNeverScanned(t *testing.T) {
	for name, source := range map[string]string{
		"fenced":      "```sh\necho $x^2$\n```",
		"tilde":       "~~~\n$a$\n~~~",
		"code span":   "Use `$a$` to write it.",
		"double span": "Here ``$a$`` too.",
	} {
		t.Run(name, func(t *testing.T) {
			var lifted []string
			if out := extractDollarMath(source, place(&lifted)); out != source || len(lifted) != 0 {
				t.Fatalf("scanned code: %q -> %q (%v)", source, out, lifted)
			}
		})
	}
}

// A code span on the same line as real math must not swallow the math, and the
// math must not swallow the code span.
func TestMathAndCodeSpanCoexistOnOneLine(t *testing.T) {
	var lifted []string
	out := extractDollarMath("Call `f($x)` when $y = 2$ holds.", place(&lifted))
	if len(lifted) != 1 {
		t.Fatalf("lifted %d formulas, want 1: %v", len(lifted), lifted)
	}
	if !strings.Contains(lifted[0], `data-fastr-docs-math="y = 2"`) {
		t.Fatalf("wrong formula lifted: %s", lifted[0])
	}
	if !strings.Contains(out, "`f($x)`") {
		t.Fatalf("code span was altered: %s", out)
	}
}

func TestRuntimeAssetsCarryTheRuntimeStylesheetAndFonts(t *testing.T) {
	assets, err := Plugin{}.RuntimeAssets()
	if err != nil {
		t.Fatalf("RuntimeAssets() error = %v", err)
	}
	for _, want := range []string{"katex/katex.js", "katex/katex.css"} {
		if len(assets[want]) == 0 {
			t.Fatalf("RuntimeAssets() missing or empty %q", want)
		}
	}
	// The bundle carries KaTeX itself; a stub would defeat the point.
	if len(assets["katex/katex.js"]) < 150_000 {
		t.Fatalf("bundle is %d bytes, too small to contain KaTeX", len(assets["katex/katex.js"]))
	}
	fonts := 0
	for name := range assets {
		if strings.HasPrefix(name, "katex/fonts/") {
			fonts++
			if !strings.HasSuffix(name, ".woff2") {
				t.Fatalf("a non-woff2 font shipped: %q", name)
			}
		}
	}
	if fonts == 0 {
		t.Fatal("no fonts shipped; KaTeX would fall back to the browser's default face")
	}
	// The stylesheet must resolve its fonts under the asset prefix, not from a
	// CDN: default-src 'self' would block that, and it would leak readers.
	css := string(assets["katex/katex.css"])
	if strings.Contains(css, "http://") || strings.Contains(css, "https://") {
		t.Fatal("stylesheet references an absolute URL")
	}
	if strings.Contains(css, ".ttf") || strings.Contains(css, ".woff)") {
		t.Fatal("stylesheet still lists font formats that are not shipped")
	}
}

// Only the runtime belongs in a <script> tag. Selecting page scripts by file
// extension would put the stylesheet's neighbours there too.
func TestOnlyTheRuntimeIsAPageScript(t *testing.T) {
	scripts := Plugin{}.PageScripts()
	if len(scripts) != 1 || scripts[0] != "katex/katex.js" {
		t.Fatalf("PageScripts() = %v, want [katex/katex.js]", scripts)
	}
	r := routerWithPlugin(t, Plugin{})
	names, err := r.RuntimeScriptNames("")
	if err != nil {
		t.Fatalf("RuntimeScriptNames() error = %v", err)
	}
	for _, name := range names {
		if strings.HasPrefix(name, "katex/fonts/") || strings.HasSuffix(name, ".css") {
			t.Fatalf("RuntimeScriptNames() included %q", name)
		}
	}
}

func TestFontsAreServedAsWoff2AndCachedHard(t *testing.T) {
	r := routerWithPlugin(t, Plugin{})
	httpRouter := gofastrRouter.New()
	if err := r.MountPluginAssets(httpRouter); err != nil {
		t.Fatalf("MountPluginAssets() error = %v", err)
	}
	server := httptest.NewServer(httpRouter)
	defer server.Close()

	response, err := http.Get(server.URL + "/__fastr-docs/katex/fonts/KaTeX_Main-Regular.woff2")
	if err != nil {
		t.Fatalf("GET font: %v", err)
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusOK {
		t.Fatalf("GET font = %d", response.StatusCode)
	}
	if got := response.Header.Get("Content-Type"); got != "font/woff2" {
		t.Fatalf("font Content-Type = %q", got)
	}
	if got := response.Header.Get("Cache-Control"); !strings.Contains(got, "immutable") {
		t.Fatalf("font Cache-Control = %q, want immutable", got)
	}
}

// The whole reason this plugin renders in the page rather than in a frame is
// that it needs no relaxation. If that ever stops being true, this test is
// where it should show up.
func TestThePagePolicyStaysStrict(t *testing.T) {
	if policy := docs.ContentSecurityPolicy(); strings.Contains(policy, "unsafe-inline") || strings.Contains(policy, "unsafe-eval") {
		t.Fatalf("page policy was relaxed: %q", policy)
	}
	// KaTeX's own error path writes a style attribute, which the policy would
	// block, so the runtime must catch errors rather than let KaTeX render
	// them.
	runtime, err := fs.ReadFile(assetsFS, "assets/katex.js")
	if err != nil {
		t.Fatalf("read runtime: %v", err)
	}
	if !strings.Contains(string(runtime), "throwOnError:!0") && !strings.Contains(string(runtime), "throwOnError: true") {
		t.Fatal("the runtime does not set throwOnError, so KaTeX would render its own error markup")
	}
}
