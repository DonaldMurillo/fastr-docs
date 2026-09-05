package katex

import (
	"strings"

	"github.com/DonaldMurillo/gofastr/core/render"
)

// extractDollarMath lifts $...$ and $$...$$ out of Markdown source before the
// parser runs, replacing each with a marker that the docs package substitutes
// back afterwards.
//
// Dollar signs are common in prose that has nothing to do with math, so the
// scanner is deliberately hard to trigger by accident. It follows the rules
// Pandoc settled on:
//
//   - the character after the opening $ is not whitespace, so "$ 5" is prose
//   - the character before the closing $ is not whitespace, so "cost $5 and
//     $10 shipping" finds no closing delimiter
//   - the character after the closing $ is not a digit, which is what rules
//     out "from $5 to $10"
//   - a backslash escapes a dollar, so writing one literally is always
//     possible
//
// Code is skipped outright: fenced blocks, and inline code spans within a line.
func extractDollarMath(source string, place func(render.HTML) string) string {
	if !strings.Contains(source, "$") {
		return source
	}
	lines := strings.Split(strings.ReplaceAll(source, "\r\n", "\n"), "\n")
	out := make([]string, 0, len(lines))
	fenced := false
	for i := 0; i < len(lines); i++ {
		line := lines[i]
		if isFence(line) {
			fenced = !fenced
			out = append(out, line)
			continue
		}
		if fenced {
			out = append(out, line)
			continue
		}
		// A line that is nothing but $$ opens display math that runs until the
		// next such line. This is the shape almost every writer uses, and
		// handling it here means the inline scanner never has to span lines.
		if strings.TrimSpace(line) == "$$" {
			if end := closingDisplayLine(lines, i+1); end > 0 {
				body := strings.Join(lines[i+1:end], "\n")
				out = append(out, place(formula(body, true)))
				i = end
				continue
			}
		}
		out = append(out, scanLine(line, place))
	}
	return strings.Join(out, "\n")
}

func closingDisplayLine(lines []string, from int) int {
	for i := from; i < len(lines); i++ {
		if isFence(lines[i]) {
			return -1
		}
		if strings.TrimSpace(lines[i]) == "$$" {
			return i
		}
	}
	return -1
}

// scanLine replaces the math delimiters in one line, leaving escapes and
// inline code spans alone.
func scanLine(line string, place func(render.HTML) string) string {
	if !strings.Contains(line, "$") {
		return line
	}
	var out strings.Builder
	for i := 0; i < len(line); {
		switch {
		case line[i] == '\\' && i+1 < len(line):
			// An escaped dollar stays a dollar. The backslash is left in place
			// for the Markdown parser to consume as it normally would.
			out.WriteString(line[i : i+2])
			i += 2
		case line[i] == '`':
			end := closingCodeSpan(line, i)
			out.WriteString(line[i:end])
			i = end
		case line[i] == '$':
			text, display, end := readFormula(line, i)
			if end < 0 {
				out.WriteByte('$')
				i++
				continue
			}
			out.WriteString(place(formula(text, display)))
			i = end
		default:
			out.WriteByte(line[i])
			i++
		}
	}
	return out.String()
}

// closingCodeSpan returns the offset just past a run of backticks and whatever
// it delimits. An unclosed run is treated as ordinary text, which is what the
// Markdown parser will do with it too.
func closingCodeSpan(line string, start int) int {
	run := 0
	for start+run < len(line) && line[start+run] == '`' {
		run++
	}
	ticks := strings.Repeat("`", run)
	rest := start + run
	for {
		next := strings.Index(line[rest:], ticks)
		if next < 0 {
			return start + run
		}
		at := rest + next
		after := at + run
		if after < len(line) && line[after] == '`' {
			// A longer run does not close a shorter one.
			rest = after
			continue
		}
		return after
	}
}

// readFormula reads one delimited formula starting at the dollar sign at
// start. It returns the TeX, whether it is display math, and the offset just
// past the closing delimiter, or -1 when this dollar opens nothing.
func readFormula(line string, start int) (string, bool, int) {
	display := start+1 < len(line) && line[start+1] == '$'
	open := start + 1
	if display {
		open = start + 2
	}
	if open >= len(line) || isSpace(line[open]) {
		return "", false, -1
	}
	for i := open; i < len(line); i++ {
		if line[i] == '\\' {
			i++
			continue
		}
		if line[i] != '$' {
			continue
		}
		if display {
			if i+1 >= len(line) || line[i+1] != '$' {
				continue
			}
		}
		if isSpace(line[i-1]) {
			continue
		}
		end := i + 1
		if display {
			end = i + 2
		}
		// "$5 to $10" is money, not math. Only the inline form needs this
		// check; $$ is unambiguous enough on its own.
		if !display && end < len(line) && isDigit(line[end]) {
			return "", false, -1
		}
		text := line[open:i]
		if strings.TrimSpace(text) == "" {
			return "", false, -1
		}
		return text, display, end
	}
	return "", false, -1
}

func isSpace(c byte) bool { return c == ' ' || c == '\t' }

func isDigit(c byte) bool { return c >= '0' && c <= '9' }
