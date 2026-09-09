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

// A fence shown inside a longer fence is example content. Its options belong to
// the example being documented, not to this document, so they must not apply
// here.
//
// The whole block is lifted as one unit because GoFastr's parser reads only
// three fence characters: left to it, ````md is ``` with a language of "`md",
// and the first inner ``` closes the block, scattering the example across three
// pieces of output.
func TestFenceInsideALongerFenceIsContentNotAFence(t *testing.T) {
	source := "````md\n```go title=\"main.go\" {1}\na := 1\n```\n````\n"
	// A plain fence of any length passes through to GoFastr, which reads the
	// marker length since v0.83.0, so the inner optioned fence is never seen
	// as a fence of its own.
	if out, replacements := extractRichCodeFences(source); len(replacements) != 0 || out != source {
		t.Fatalf("a plain four-backtick fence was lifted: %q -> %q (%d replacements)", source, out, len(replacements))
	}
	html := string(renderDocsMarkdown(source, nil))
	for _, want := range []string{"title=&quot;main.go&quot; {1}", "a := 1"} {
		if !strings.Contains(html, want) {
			t.Fatalf("example lost %q: %s", want, html)
		}
	}
	if blocks := strings.Count(html, "ui-code-block__body"); blocks != 1 {
		t.Fatalf("the inner fence became %d blocks of its own: %s", blocks, html)
	}
	// The example's own title must stay text. A filename header here would mean
	// the document adopted the options it is trying to show.
	if strings.Contains(html, "ui-code-block__filename") {
		t.Fatalf("the example's options were applied to this document: %s", html)
	}
}

// A plain fence following another must still be scanned.
func TestFencesAfterAPlainFenceAreStillProcessed(t *testing.T) {
	source := "```go\na := 1\n```\n\n```go title=\"second.go\"\nb := 2\n```\n"
	_, replacements := extractRichCodeFences(source)
	if len(replacements) != 1 {
		t.Fatalf("expected the second fence to be lifted, got %d", len(replacements))
	}
}

// GoFastr's parser only ever reads three fence characters, so it takes ````md
// as ``` with a language of "`md" and then lets the first inner ``` close the
// block. A Markdown example showing a fenced block inside a shortcode came out
// as three broken pieces, which is exactly how the components page is written.
func TestLongerFenceKeepsANestedExampleWhole(t *testing.T) {
	source := "````md\n{{< filetree >}}\n```\nmy-docs/\n```\n{{< /filetree >}}\n````\n"
	out := string(renderDocsMarkdown(source, nil))

	if blocks := strings.Count(out, "ui-code-block__body"); blocks != 1 {
		t.Fatalf("rendered %d code blocks, want 1: %s", blocks, out)
	}
	for _, want := range []string{"{{&lt; filetree &gt;}}", "my-docs/", "{{&lt; /filetree &gt;}}"} {
		if !strings.Contains(out, want) {
			t.Fatalf("output lost %q: %s", want, out)
		}
	}
	// The language is "md", not "`md": the fourth backtick belongs to the
	// marker.
	if !strings.Contains(out, `aria-label="md source"`) {
		t.Fatalf("the fence marker leaked into the language: %s", out)
	}
}

// A plain three-character fence must still pass through to GoFastr untouched,
// so existing output does not change.
func TestOrdinaryFenceIsStillNotLifted(t *testing.T) {
	source := "```go\nx := 1\n```\n"
	if out, replacements := extractRichCodeFences(source); len(replacements) != 0 || out != source {
		t.Fatalf("a plain fence was lifted: %q -> %q (%d replacements)", source, out, len(replacements))
	}
}

func TestFenceLanguagesAreLenient(t *testing.T) {
	t.Run("fence languages fold case", func(t *testing.T) {
		language, options := splitFenceInfo("GO {1}")
		fence := parseFenceOptions(language, options)
		if fence.language != "go" || len(fence.highlight) != 1 {
			t.Fatalf("fence = %+v, want go with its highlight", fence)
		}
	})
	t.Run("fence languages tolerate trailing space", func(t *testing.T) {
		language, _ := splitFenceInfo("go ")
		if language != "go" {
			t.Fatalf("splitFenceInfo = %q, want go", language)
		}
	})
}
