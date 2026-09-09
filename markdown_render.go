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
// icon-first form. Keeping that adapter here lets fastr-docs use the
// released GoFastr API instead of requiring a framework fork or local replace.
var docsCopyButton = regexp.MustCompile(`(?s)<button\b[^>]*\bclass="[^"]*\bui-copy-btn\b[^"]*"[^>]*>.*?</button>`)

func renderDocsMarkdown(source string, attrs map[string]string) render.HTML {
	// Fences carrying options are lifted out first. GoFastr's parser would read
	// the whole info string as the language and lose highlighting entirely.
	source, fences := extractRichCodeFences(source)
	htmlBody := string(ui.Markdown(ui.MarkdownConfig{
		Source:     source,
		ExtraAttrs: html.Attrs(attrs),
	}))
	// Substituted before the copy-button pass so lifted blocks get the same
	// icon-only affordance as the ones GoFastr rendered.
	htmlBody = string(applyShortcodes(render.HTML(htmlBody), fences))
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
