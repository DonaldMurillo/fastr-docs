package docs

import (
	"strings"
	"testing"
)

// A fence with no options must not be touched at all. The interception is only
// worth having if it is provably opt-in.
func TestPlainFencesRenderExactlyAsBefore(t *testing.T) {
	for name, source := range map[string]string{
		"language":    "# Title\n\n```go\nfunc main() {}\n```\n",
		"no language": "```\nplain text\n```\n",
		"tilde":       "~~~js\nconst a = 1;\n~~~\n",
		"two blocks":  "```go\na := 1\n```\n\ntext\n\n```sh\nls -la\n```\n",
		"indented":    "- item\n\n  ```go\n  a := 1\n  ```\n",
	} {
		t.Run(name, func(t *testing.T) {
			stripped, replacements := extractRichCodeFences(source)
			if len(replacements) != 0 {
				t.Fatalf("plain fence was lifted out: %v", replacements)
			}
			if stripped != strings.ReplaceAll(source, "\r\n", "\n") {
				t.Fatalf("plain fence source was rewritten:\nwant %q\ngot  %q", source, stripped)
			}
		})
	}
}

func TestFenceOptionsAreParsed(t *testing.T) {
	for name, testCase := range map[string]struct {
		info string
		want codeFence
	}{
		"title": {
			`go title="main.go"`,
			codeFence{language: "go", title: "main.go"},
		},
		"single quoted title": {
			`go title='cmd/main.go'`,
			codeFence{language: "go", title: "cmd/main.go"},
		},
		"highlight list": {
			"go {1,3}",
			codeFence{language: "go", highlight: map[int]bool{1: true, 3: true}},
		},
		"highlight range": {
			"go {2-4}",
			codeFence{language: "go", highlight: map[int]bool{2: true, 3: true, 4: true}},
		},
		"mixed": {
			`ts title="app.ts" {1,4-5} showLineNumbers`,
			codeFence{language: "ts", title: "app.ts", lineNumbers: true, highlight: map[int]bool{1: true, 4: true, 5: true}},
		},
		"title with spaces": {
			`sh title="deploy to staging"`,
			codeFence{language: "sh", title: "deploy to staging"},
		},
		"scroll": {
			"json scroll",
			codeFence{language: "json", scroll: true},
		},
		"no language": {
			`title="notes.txt"`,
			codeFence{title: "notes.txt"},
		},
	} {
		t.Run(name, func(t *testing.T) {
			language, options := splitFenceInfo(testCase.info)
			got := parseFenceOptions(language, options)
			if got.language != testCase.want.language {
				t.Fatalf("language = %q, want %q", got.language, testCase.want.language)
			}
			if got.title != testCase.want.title {
				t.Fatalf("title = %q, want %q", got.title, testCase.want.title)
			}
			if got.lineNumbers != testCase.want.lineNumbers || got.scroll != testCase.want.scroll {
				t.Fatalf("flags = %+v, want lineNumbers=%v scroll=%v", got, testCase.want.lineNumbers, testCase.want.scroll)
			}
			for line := range testCase.want.highlight {
				if !got.highlight[line] {
					t.Fatalf("line %d not highlighted: %v", line, got.highlight)
				}
			}
			if len(got.highlight) != len(testCase.want.highlight) {
				t.Fatalf("highlight = %v, want %v", got.highlight, testCase.want.highlight)
			}
		})
	}
}

func TestRichFencesRenderTitleNumbersAndHighlights(t *testing.T) {
	html := string(renderDocsMarkdown("```go title=\"main.go\" {2} showLineNumbers\npackage main\n\nfunc main() {}\n```\n", nil))
	for _, want := range []string{
		"main.go",
		"ui-code-block--numbered",
		`class="fastr-docs-code-hl"`,
		`<span class="tk-kw">package</span> main`,
	} {
		if !strings.Contains(html, want) {
			t.Fatalf("rich fence output missing %q: %s", want, html)
		}
	}
	// The options must never survive as the language class.
	if strings.Contains(html, "language-go title") || strings.Contains(html, "{2}") {
		t.Fatalf("fence options leaked into the output: %s", html)
	}
	if strings.Contains(html, "FASTRDOCSFENCESLOT") {
		t.Fatalf("fence marker was not substituted: %s", html)
	}
}

// This is the behaviour the whole milestone exists for: before, an option in
// the info string cost you syntax highlighting without any warning.
func TestFenceOptionsNoLongerCostSyntaxHighlighting(t *testing.T) {
	plain := string(renderDocsMarkdown("```go\nfunc main() { return }\n```\n", nil))
	titled := string(renderDocsMarkdown("```go title=\"main.go\"\nfunc main() { return }\n```\n", nil))
	countTokens := func(html string) int { return strings.Count(html, "tk-kw") }
	if countTokens(plain) == 0 {
		t.Skip("this GoFastr build does not tokenize go")
	}
	if countTokens(titled) < countTokens(plain) {
		t.Fatalf("a titled fence lost syntax highlighting: plain=%d titled=%d", countTokens(plain), countTokens(titled))
	}
}

func TestRichFencesWorkInsideShortcodeBodies(t *testing.T) {
	r := NewRouter()
	html, err := renderMarkdownWithComponents("{{< note title=\"See\" >}}\n```go title=\"inner.go\"\na := 1\n```\n{{< /note >}}", r.vocabulary(nil, nil), nil)
	if err != nil {
		t.Fatalf("renderMarkdownWithComponents() error = %v", err)
	}
	if !strings.Contains(string(html), "inner.go") {
		t.Fatalf("fence inside a shortcode body lost its title: %s", html)
	}
}

func TestMalformedFenceOptionsDoNotBreakTheDocument(t *testing.T) {
	for name, source := range map[string]string{
		"unterminated":     "```go title=\"main.go\"\nnever closed\n",
		"bad range":        "```go {4-2}\na\n```\n",
		"non numeric":      "```go {abc}\na\n```\n",
		"huge range":       "```go {1-99999999}\na\n```\n",
		"unbalanced quote": "```go title=\"unclosed\na\n```\n",
		"empty braces":     "```go {}\na\n```\n",
	} {
		t.Run(name, func(t *testing.T) {
			html := string(renderDocsMarkdown(source, nil))
			if strings.Contains(html, "FASTRDOCSFENCESLOT") {
				t.Fatalf("marker survived: %s", html)
			}
		})
	}
}

// A fence shown inside a longer fence is example content. Lifting it out would
// replace the example with the thing it is documenting.
func TestFenceInsideALongerFenceIsLeftAlone(t *testing.T) {
	source := "````md\n```go title=\"main.go\" {1}\na := 1\n```\n````\n"
	stripped, replacements := extractRichCodeFences(source)
	if len(replacements) != 0 {
		t.Fatalf("a fence inside a ```` block was lifted out: %v", replacements)
	}
	if stripped != source[:len(source)-1] && stripped != source {
		t.Fatalf("example block was rewritten:\nwant %q\ngot  %q", source, stripped)
	}

	// Only the extractor is asserted here. GoFastr's Markdown parser has no
	// support for four-backtick fences at all: it closes the block on the first
	// ``` inside and renders the rest as prose. Showing a fenced example inside
	// another fence is not possible on this engine, whatever this code does.
}

// A plain fence following another must still be scanned.
func TestFencesAfterAPlainFenceAreStillProcessed(t *testing.T) {
	source := "```go\na := 1\n```\n\n```go title=\"second.go\"\nb := 2\n```\n"
	_, replacements := extractRichCodeFences(source)
	if len(replacements) != 1 {
		t.Fatalf("expected the second fence to be lifted, got %d", len(replacements))
	}
}
