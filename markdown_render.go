package docs

import (
	"regexp"
	"strings"

	"github.com/DonaldMurillo/gofastr/core-ui/html"
	"github.com/DonaldMurillo/gofastr/core/render"
	"github.com/DonaldMurillo/gofastr/framework/ui"
)

// Markdown rendering belongs to GoFastr, including parsing, highlighting,
// component CSS, and the clipboard runtime. fastr-docs only changes the
// presentation of the code-block copy affordance to match the docs shell's
// icon-first contract. Keeping that adapter here lets fastr-docs use the
// released GoFastr API instead of requiring a framework fork or local replace.
var docsCopyButton = regexp.MustCompile(`(?s)<button\b[^>]*\bclass="[^"]*\bui-copy-btn\b[^"]*"[^>]*>.*?</button>`)

func renderDocsMarkdown(source string, attrs map[string]string) render.HTML {
	htmlBody := string(ui.Markdown(ui.MarkdownConfig{
		Source:     source,
		ExtraAttrs: html.Attrs(attrs),
	}))
	return render.HTML(docsCopyButton.ReplaceAllStringFunc(htmlBody, iconOnlyCopyButton))
}

func iconOnlyCopyButton(button string) string {
	end := strings.IndexByte(button, '>')
	if end < 0 {
		return button
	}
	opening := button[:end]
	if !strings.Contains(opening, "ui-copy-btn--icon") {
		opening = strings.Replace(opening, `class="`, `class="ui-copy-btn--icon `, 1)
	}
	if !strings.Contains(opening, "aria-label=") {
		opening += ` aria-label="Copy to clipboard"`
	}
	return opening + `>` +
		`<span class="ui-copy-btn__label" aria-hidden="true">⧉</span>` +
		`<span class="ui-copy-btn__copied" aria-hidden="true">✓</span>` +
		`</button>`
}
