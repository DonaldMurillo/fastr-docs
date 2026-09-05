package docs

import (
	"fmt"
	"strings"
	"testing"

	"github.com/DonaldMurillo/gofastr/core/render"
)

// A shortcode nested inside another shortcode has to survive marker
// substitution. The outer component's HTML carries the inner component's
// marker, so replacing markers in completion order alone leaves the inner
// marker stranded as visible text.
func TestNestedShortcodesResolveInsideTheirParent(t *testing.T) {
	components := map[string]MarkdownComponent{
		"outer": func(_ map[string]string, body render.HTML) render.HTML {
			return render.Tag("section", map[string]string{"class": "outer"}, body)
		},
		"inner": func(props map[string]string, body render.HTML) render.HTML {
			return render.Tag("span", map[string]string{"data-label": props["label"]}, body)
		},
	}
	source := "{{< outer >}}\n\n{{< inner label=\"a\" >}}deep{{< /inner >}}\n\n{{< /outer >}}"

	html, err := renderMarkdownWithComponents(source, components, nil)
	if err != nil {
		t.Fatalf("renderMarkdownWithComponents() error = %v", err)
	}
	got := string(html)
	if strings.Contains(got, "FASTRDOCSSHORTCODESLOT") {
		t.Fatalf("nested shortcode left an unreplaced marker: %s", got)
	}
	for _, want := range []string{`class="outer"`, `data-label="a"`, "deep"} {
		if !strings.Contains(got, want) {
			t.Fatalf("nested shortcode output missing %q: %s", want, got)
		}
	}
	if strings.Index(got, `data-label="a"`) < strings.Index(got, `class="outer"`) {
		t.Fatalf("inner component rendered outside its parent: %s", got)
	}
}

// Marker 1 used to be a prefix of marker 10, so a page with eleven or more
// shortcodes corrupted the eleventh onward.
func TestManyShortcodesOnOnePageKeepTheirOwnMarkers(t *testing.T) {
	const count = 14
	components := map[string]MarkdownComponent{
		"chip": func(props map[string]string, _ render.HTML) render.HTML {
			return render.Tag("b", map[string]string{"data-n": props["n"]}, render.Text(props["n"]))
		},
	}
	var source strings.Builder
	for i := range count {
		fmt.Fprintf(&source, "{{< chip n=\"%d\" />}}\n\n", i)
	}

	html, err := renderMarkdownWithComponents(source.String(), components, nil)
	if err != nil {
		t.Fatalf("renderMarkdownWithComponents() error = %v", err)
	}
	got := string(html)
	if strings.Contains(got, "FASTRDOCSSHORTCODESLOT") {
		t.Fatalf("unreplaced marker with %d shortcodes: %s", count, got)
	}
	for i := range count {
		if !strings.Contains(got, fmt.Sprintf(`data-n="%d"`, i)) {
			t.Fatalf("shortcode %d did not render: %s", i, got)
		}
	}
}
