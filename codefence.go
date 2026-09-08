package docs

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/DonaldMurillo/gofastr/core/render"
	"github.com/DonaldMurillo/gofastr/framework/ui"
)

// Fenced code blocks carry options in their info string:
//
//	```go title="main.go" {2,4-6} showLineNumbers
//
// GoFastr's Markdown parser takes the whole remainder of the fence line as the
// language, so it would emit class="language-go title=&quot;main.go&quot;
// {2,4-6}" and then fail to match "go" at all. Writing an option today does not
// merely get ignored, it silently costs you syntax highlighting.
//
// So blocks that carry options are lifted out before the Markdown parser runs
// and rendered directly through ui.CodeBlock. A fence with no options is left
// completely untouched, which keeps existing output byte-identical.

type codeFence struct {
	language    string
	title       string
	highlight   map[int]bool
	lineNumbers bool
	scroll      bool
	code        string
}

// extractRichCodeFences replaces every option-carrying fence with a marker and
// returns the rendered blocks to substitute back after Markdown rendering.
func extractRichCodeFences(source string) (string, []shortcodeReplacement) {
	if !strings.Contains(source, "```") && !strings.Contains(source, "~~~") {
		return source, nil
	}
	lines := strings.Split(strings.ReplaceAll(source, "\r\n", "\n"), "\n")
	var out []string
	var replacements []shortcodeReplacement
	for i := 0; i < len(lines); i++ {
		marker, info, ok := openingFence(lines[i])
		if !ok {
			out = append(out, lines[i])
			continue
		}
		language, options := splitFenceInfo(info)
		end := closingFenceIndex(lines, i+1, marker)
		if options == "" {
			// A plain fence stays exactly as written, and the scan resumes
			// after its closing line. Without that skip, a fence shown inside
			// a longer ````md example block would be lifted out of the example.
			out = append(out, lines[i:min(end+1, len(lines))]...)
			i = end
			continue
		}
		body := lines[i+1 : end]
		fence := parseFenceOptions(language, options)
		fence.code = strings.Join(body, "\n")
		token := fmt.Sprintf("FASTRDOCSFENCESLOT%dEND", len(replacements))
		replacements = append(replacements, shortcodeReplacement{marker: token, html: renderCodeFence(fence)})
		out = append(out, "", token, "")
		i = end
	}
	return strings.Join(out, "\n"), replacements
}

// openingFence reports whether a line opens a fence, and returns the fence
// marker plus its info string.
func openingFence(line string) (string, string, bool) {
	trimmed := strings.TrimLeft(line, " \t")
	for _, char := range []byte{'`', '~'} {
		run := 0
		for run < len(trimmed) && trimmed[run] == char {
			run++
		}
		if run < 3 {
			continue
		}
		// The marker keeps its real length. A ```` block is closed only by a
		// run of four or more, so a ``` fence inside it is content, not a fence.
		return trimmed[:run], strings.TrimSpace(trimmed[run:]), true
	}
	return "", "", false
}

func closingFenceIndex(lines []string, from int, marker string) int {
	for i := from; i < len(lines); i++ {
		trimmed := strings.TrimSpace(lines[i])
		// Only a run of the same character, at least as long as the opener.
		if len(trimmed) >= len(marker) && strings.HasPrefix(trimmed, marker) && strings.TrimLeft(trimmed, marker[:1]) == "" {
			return i
		}
	}
	// An unterminated fence runs to the end of the document, which is what the
	// Markdown parser does too.
	return len(lines)
}

// splitFenceInfo separates the language from everything after it.
func splitFenceInfo(info string) (string, string) {
	info = strings.TrimSpace(info)
	if info == "" {
		return "", ""
	}
	cut := strings.IndexAny(info, " \t{")
	first, rest := info, ""
	if cut >= 0 {
		first = strings.TrimSpace(info[:cut])
		rest = strings.TrimSpace(info[cut:])
	}
	// A fence may carry options with no language at all, as in
	// ```title="notes.txt". Without this the option becomes the language.
	if strings.Contains(first, "=") {
		return "", info
	}
	// Language matching downstream is exact, so GO and go highlight alike.
	return strings.ToLower(first), rest
}

func parseFenceOptions(language, options string) codeFence {
	// Language matching downstream is exact; GO, Go, and go highlight alike.
	fence := codeFence{language: strings.ToLower(strings.TrimSpace(language)), highlight: map[int]bool{}}
	for _, field := range splitFenceOptions(options) {
		switch {
		case strings.HasPrefix(field, "{") && strings.HasSuffix(field, "}"):
			parseHighlightRanges(strings.TrimSuffix(strings.TrimPrefix(field, "{"), "}"), fence.highlight)
		case strings.HasPrefix(field, "title="):
			fence.title = strings.Trim(strings.TrimPrefix(field, "title="), `"'`)
		default:
			switch strings.ToLower(field) {
			case "shownumbers", "showlinenumbers", "linenumbers", "numbered", "nums":
				fence.lineNumbers = true
			case "scroll", "collapse":
				fence.scroll = true
			}
		}
	}
	return fence
}

// splitFenceOptions splits on whitespace while keeping quoted values and
// {1,3-5} ranges intact.
func splitFenceOptions(options string) []string {
	var fields []string
	var current strings.Builder
	quote := byte(0)
	braces := 0
	flush := func() {
		if current.Len() > 0 {
			fields = append(fields, current.String())
			current.Reset()
		}
	}
	for i := 0; i < len(options); i++ {
		c := options[i]
		switch {
		case quote != 0:
			current.WriteByte(c)
			if c == quote {
				quote = 0
			}
		case c == '"' || c == '\'':
			quote = c
			current.WriteByte(c)
		case c == '{':
			braces++
			current.WriteByte(c)
		case c == '}':
			if braces > 0 {
				braces--
			}
			current.WriteByte(c)
			if braces == 0 {
				flush()
			}
		case (c == ' ' || c == '\t') && braces == 0:
			flush()
		default:
			current.WriteByte(c)
		}
	}
	flush()
	return fields
}

func parseHighlightRanges(spec string, into map[int]bool) {
	for _, part := range strings.Split(spec, ",") {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}
		from, to, isRange := strings.Cut(part, "-")
		start, err := strconv.Atoi(strings.TrimSpace(from))
		if err != nil {
			continue
		}
		end := start
		if isRange {
			parsed, err := strconv.Atoi(strings.TrimSpace(to))
			if err != nil {
				continue
			}
			end = parsed
		}
		if end < start {
			start, end = end, start
		}
		// A range large enough to be a typo would otherwise allocate for every
		// line in it.
		if end-start > 10000 {
			continue
		}
		for line := start; line <= end; line++ {
			if line > 0 {
				into[line] = true
			}
		}
	}
}

func renderCodeFence(fence codeFence) render.HTML {
	lines := ui.HighlightLines(fence.code, fence.language)
	if len(fence.highlight) > 0 {
		// CodeBlock wraps each entry in its own .ui-code-block__line span and
		// gives no way to mark one, so the emphasis has to live inside the
		// line, styled from styles.go.
		for i := range lines {
			if fence.highlight[i+1] {
				lines[i] = render.Tag("span", map[string]string{"class": "fastr-docs-code-hl"}, lines[i])
			}
		}
	}
	return ui.CodeBlock(ui.CodeBlockConfig{
		Lines:       lines,
		Language:    fence.language,
		Filename:    fence.title,
		ShowCopy:    true,
		LineNumbers: fence.lineNumbers,
		Scroll:      fence.scroll,
	})
}
