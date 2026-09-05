package docs

import (
	"strings"
	"testing"

	"github.com/DonaldMurillo/gofastr/core/render"
)

func renderDefaults(t *testing.T, source string) string {
	t.Helper()
	r := NewRouter()
	html, err := renderMarkdownWithComponents(source, r.vocabulary(nil, nil), nil)
	if err != nil {
		t.Fatalf("renderMarkdownWithComponents(%q) error = %v", source, err)
	}
	return string(html)
}

func TestDefaultComponentsAreAvailableWithoutRegistration(t *testing.T) {
	for name, testCase := range map[string]struct {
		source string
		want   []string
	}{
		"note":     {"{{< note title=\"Heads up\" >}}Body text{{< /note >}}", []string{"Heads up", "Body text"}},
		"danger":   {"{{< danger >}}Careful{{< /danger >}}", []string{"Careful"}},
		"callout":  {"{{< callout variant=\"warning\" title=\"T\" >}}W{{< /callout >}}", []string{"T", "W"}},
		"card":     {"{{< card title=\"Routing\" description=\"How it works\" href=\"/docs\" >}}More{{< /card >}}", []string{"Routing", "How it works", `href="/docs"`}},
		"details":  {"{{< details summary=\"Show\" >}}Hidden{{< /details >}}", []string{"Show", "Hidden", "<details"}},
		"badge":    {"{{< badge label=\"New\" />}}", []string{"New"}},
		"tag":      {"{{< tag label=\"go\" href=\"/tags/go\" />}}", []string{"go", "/tags/go"}},
		"banner":   {"{{< banner title=\"Release\" body=\"v1 is out\" />}}", []string{"Release", "v1 is out"}},
		"icon":     {"{{< icon name=\"check\" />}}", []string{"<svg"}},
		"steps":    {"{{< steps >}}\n1. First\n2. Second\n{{< /steps >}}", []string{"fastr-docs-steps", "First", "Second"}},
		"filetree": {"{{< filetree >}}\n- src\n  - main.go\n{{< /filetree >}}", []string{"fastr-docs-filetree", "main.go"}},
	} {
		t.Run(name, func(t *testing.T) {
			got := renderDefaults(t, testCase.source)
			for _, want := range testCase.want {
				if !strings.Contains(got, want) {
					t.Fatalf("%s output missing %q: %s", name, want, got)
				}
			}
			if strings.Contains(got, "FASTRDOCSSHORTCODESLOT") {
				t.Fatalf("%s left an unreplaced marker: %s", name, got)
			}
		})
	}
}

func TestDefaultTabsAndGridUseTheirChildren(t *testing.T) {
	got := renderDefaults(t, "{{< tabs >}}\n{{< tab label=\"Go\" >}}go body{{< /tab >}}\n{{< tab label=\"Shell\" >}}shell body{{< /tab >}}\n{{< /tabs >}}")
	for _, want := range []string{"Go", "Shell", "go body", "shell body", "<details"} {
		if !strings.Contains(got, want) {
			t.Fatalf("tabs output missing %q: %s", want, got)
		}
	}

	grid := renderDefaults(t, "{{< cards >}}\n{{< card title=\"One\" >}}a{{< /card >}}\n{{< card title=\"Two\" >}}b{{< /card >}}\n{{< /cards >}}")
	for _, want := range []string{"One", "Two", "ui-grid"} {
		if !strings.Contains(grid, want) {
			t.Fatalf("cards output missing %q: %s", want, grid)
		}
	}
}

func TestTabGroupsGetDistinctNamesOnOnePage(t *testing.T) {
	source := strings.Repeat("{{< tabs >}}{{< tab label=\"A\" >}}x{{< /tab >}}{{< /tabs >}}\n\n", 2)
	got := renderDefaults(t, source)
	names := map[string]bool{}
	for _, part := range strings.Split(got, `name="`)[1:] {
		names[part[:strings.Index(part, `"`)]] = true
	}
	if len(names) < 2 {
		t.Fatalf("two tab sets shared a group name, which breaks native exclusivity: %s", got)
	}
}

// Props come from Markdown, so they come from an author. Several GoFastr
// components panic on input they reject; none of that may reach a reader.
func TestDefaultComponentsSurviveHostileOrEmptyProps(t *testing.T) {
	for name, source := range map[string]string{
		"unknown callout variant": `{{< callout variant="bogus" >}}body{{< /callout >}}`,
		"unknown banner variant":  `{{< banner title="T" variant="bogus" />}}`,
		"unknown card variant":    `{{< card title="T" variant="bogus" >}}b{{< /card >}}`,
		"unknown grid gap":        `{{< cards gap="bogus" >}}{{< card title="T" >}}b{{< /card >}}{{< /cards >}}`,
		"unknown icon":            `{{< icon name="definitely-not-an-icon" />}}`,
		"empty tab set":           `{{< tabs >}}nothing here{{< /tabs >}}`,
		"empty grid":              `{{< cards >}}{{< /cards >}}`,
		"banner without title":    `{{< banner />}}`,
		"badge without label":     `{{< badge />}}`,
		"icon without name":       `{{< icon />}}`,
		"action without href":     `{{< hero title="T" >}}{{< action label="Go" />}}{{< /hero >}}`,
		"tab without label":       `{{< tabs >}}{{< tab >}}body{{< /tab >}}{{< /tabs >}}`,
	} {
		t.Run(name, func(t *testing.T) {
			// A panic here would take down a page render, so the assertion is
			// simply that this returns.
			got := renderDefaults(t, source)
			if strings.Contains(got, "FASTRDOCSSHORTCODESLOT") {
				t.Fatalf("%s left an unreplaced marker: %s", name, got)
			}
		})
	}
}

func TestAuthorSuppliedLinksCannotCarryScript(t *testing.T) {
	for name, source := range map[string]string{
		"card":   `{{< card title="T" href="javascript:alert(1)" >}}b{{< /card >}}`,
		"tag":    `{{< tag label="T" href="javascript:alert(1)" />}}`,
		"action": `{{< hero title="T" >}}{{< action label="Go" href="javascript:alert(1)" />}}{{< /hero >}}`,
	} {
		t.Run(name, func(t *testing.T) {
			got := renderDefaults(t, source)
			if strings.Contains(strings.ToLower(got), "javascript:") {
				t.Fatalf("%s passed through a javascript: URL: %s", name, got)
			}
		})
	}
}

func TestHeroBuildsActionsFromItsChildren(t *testing.T) {
	got := renderDefaults(t, "{{< hero eyebrow=\"Docs\" title=\"Build fast\" subtitle=\"Ship it\" >}}\n{{< action label=\"Start\" href=\"/docs\" />}}\n{{< action label=\"API\" href=\"/api\" variant=\"secondary\" />}}\n{{< /hero >}}")
	for _, want := range []string{"Docs", "Build fast", "Ship it", "Start", "/docs", "API", "/api"} {
		if !strings.Contains(got, want) {
			t.Fatalf("hero output missing %q: %s", want, got)
		}
	}
}

func TestDiffShortcodeRecoversThePatchFromItsBody(t *testing.T) {
	got := renderDefaults(t, "{{< diff >}}\n```\n-old line\n+new line\n```\n{{< /diff >}}")
	for _, want := range []string{"old line", "new line"} {
		if !strings.Contains(got, want) {
			t.Fatalf("diff output missing %q: %s", want, got)
		}
	}
}

func TestWithoutDefaultComponentsLeavesShortcodesUnregistered(t *testing.T) {
	r := NewRouter(WithoutDefaultComponents())
	if len(r.MarkdownComponents()) != 0 || len(r.MarkdownContainers()) != 0 || len(r.MarkdownRawComponents()) != 0 {
		t.Fatalf("WithoutDefaultComponents() left a vocabulary behind")
	}
	if err := validateMarkdownComponents("{{< note >}}x{{< /note >}}", r.vocabulary(nil, nil)); err == nil {
		t.Fatal("expected an unregistered shortcode error")
	}
}

func TestProjectComponentsOverrideTheDefaults(t *testing.T) {
	r := NewRouter()
	if err := r.RegisterMarkdownComponent("note", func(_ map[string]string, _ render.HTML) render.HTML {
		return "<p data-custom-note>replaced</p>"
	}); err != nil {
		t.Fatalf("RegisterMarkdownComponent() error = %v", err)
	}
	html, err := renderMarkdownWithComponents("{{< note >}}original{{< /note >}}", r.vocabulary(nil, nil), nil)
	if err != nil {
		t.Fatalf("renderMarkdownWithComponents() error = %v", err)
	}
	if !strings.Contains(string(html), "data-custom-note") {
		t.Fatalf("project component did not override the default: %s", html)
	}
}

// The file tree and diff shortcodes take their bodies unrendered. Rendering
// first collapses every line into one, and pulls the code block's own chrome
// ("N lines", the copy glyphs) into the text.
func TestRawShortcodesKeepTheirLineStructure(t *testing.T) {
	tree := renderDefaults(t, "{{< filetree >}}\n```\nmy-docs/\n  content/\n    index.md\n  main.go\n```\n{{< /filetree >}}")
	if strings.Contains(tree, "lines") || strings.Contains(tree, "ui-code-block") {
		t.Fatalf("file tree captured the code block chrome: %s", tree)
	}
	if strings.Count(tree, "<ul>") < 2 {
		t.Fatalf("file tree did not nest: %s", tree)
	}
	for _, want := range []string{
		`<span class="fastr-docs-filetree__dir">my-docs/</span>`,
		`<span class="fastr-docs-filetree__dir">content/</span>`,
		`<span class="fastr-docs-filetree__file">index.md</span>`,
		`<span class="fastr-docs-filetree__file">main.go</span>`,
	} {
		if !strings.Contains(tree, want) {
			t.Fatalf("file tree missing %q: %s", want, tree)
		}
	}

	diff := renderDefaults(t, "{{< diff >}}\n```\n context\n-removed\n+added\n```\n{{< /diff >}}")
	if strings.Contains(diff, "lines") {
		t.Fatalf("diff captured the code block chrome: %s", diff)
	}
	for _, want := range []string{"ui-diff-viewer__line--remove", "ui-diff-viewer__line--add", "removed", "added"} {
		if !strings.Contains(diff, want) {
			t.Fatalf("diff missing %q: %s", want, diff)
		}
	}
}

func TestFileTreeReadsIndentationWithoutAFence(t *testing.T) {
	got := renderDefaults(t, "{{< filetree >}}\nsrc/\n  main.go\n{{< /filetree >}}")
	if strings.Count(got, "<ul>") < 2 {
		t.Fatalf("unfenced file tree did not nest: %s", got)
	}
}

func TestProjectRawComponentsOverrideTheDefaults(t *testing.T) {
	r := NewRouter()
	if err := r.RegisterMarkdownRawComponent("diff", func(_ map[string]string, raw string) render.HTML {
		return render.HTML("<pre data-custom-diff>" + raw + "</pre>")
	}); err != nil {
		t.Fatalf("RegisterMarkdownRawComponent() error = %v", err)
	}
	html, err := renderMarkdownWithComponents("{{< diff >}}-a\n+b{{< /diff >}}", r.vocabulary(nil, nil), nil)
	if err != nil {
		t.Fatalf("renderMarkdownWithComponents() error = %v", err)
	}
	if !strings.Contains(string(html), "data-custom-diff") {
		t.Fatalf("project raw component did not override the default: %s", html)
	}
}
