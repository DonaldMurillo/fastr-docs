package docs

import (
	"strings"
	"testing"
)

func TestRenderDocsMarkdownUsesReleasedCopyButtonAsIcon(t *testing.T) {
	html := string(renderDocsMarkdown("```go\nfmt.Println(\"hi\")\n```", nil))
	for _, marker := range []string{
		"ui-copy-btn--icon",
		`aria-label="Copy to clipboard"`,
		`ui-copy-btn__label" aria-hidden="true">⧉`,
		`ui-copy-btn__copied" aria-hidden="true">✓`,
	} {
		if !strings.Contains(html, marker) {
			t.Errorf("docs Markdown is missing %q:\n%s", marker, html)
		}
	}
	if strings.Contains(html, `ui-copy-btn__label">copy</span>`) {
		t.Errorf("docs Markdown should not render a text copy label:\n%s", html)
	}
}

func TestRenderDocsMarkdownPreservesCodeCopyTargetAndAttributes(t *testing.T) {
	html := string(renderDocsMarkdown("```go\nfmt.Println(\"hi\")\n```", map[string]string{
		"data-docs-route": "/docs/example",
	}))
	if !strings.Contains(html, `data-docs-route="/docs/example"`) {
		t.Fatalf("docs Markdown lost caller attributes:\n%s", html)
	}
	if !strings.Contains(html, `data-fui-copy-text-from="#ui-code-block-`) {
		t.Fatalf("docs Markdown lost the framework copy target:\n%s", html)
	}
}
