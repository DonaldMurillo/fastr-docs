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

	html, err := renderMarkdownWithComponents(source, markdownVocabulary{components: components}, nil)
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

	html, err := renderMarkdownWithComponents(source.String(), markdownVocabulary{components: components}, nil)
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

// A container has to see its nested shortcodes as separate items, which is the
// whole point: tabs, card grids, and steps cannot be built from one merged blob.
func TestContainersReceiveTheirChildrenSeparately(t *testing.T) {
	vocab := markdownVocabulary{
		components: map[string]MarkdownComponent{
			"tab": func(_ map[string]string, body render.HTML) render.HTML { return body },
		},
		containers: map[string]MarkdownContainer{
			"tabs": func(_ map[string]string, children []MarkdownChild, body render.HTML) render.HTML {
				parts := []render.HTML{render.Tag("i", nil, render.Text(fmt.Sprint(len(children))))}
				for _, child := range children {
					parts = append(parts, render.Tag("section",
						map[string]string{"data-name": child.Name, "data-label": child.Props["label"]}, child.Body))
				}
				parts = append(parts, render.Tag("footer", nil, body))
				return render.Tag("div", map[string]string{"class": "tabs"}, parts...)
			},
		},
	}
	source := "{{< tabs >}}\n\nbetween\n\n{{< tab label=\"Go\" >}}go body{{< /tab >}}\n\n{{< tab label=\"Shell\" >}}shell body{{< /tab >}}\n\n{{< /tabs >}}"

	html, err := renderMarkdownWithComponents(source, vocab, nil)
	if err != nil {
		t.Fatalf("renderMarkdownWithComponents() error = %v", err)
	}
	got := string(html)
	if strings.Contains(got, "FASTRDOCSSHORTCODESLOT") {
		t.Fatalf("container left an unreplaced marker: %s", got)
	}
	if !strings.Contains(got, "<i>2</i>") {
		t.Fatalf("container did not receive both children: %s", got)
	}
	for _, want := range []string{`data-label="Go"`, `data-label="Shell"`, "go body", "shell body", "between"} {
		if !strings.Contains(got, want) {
			t.Fatalf("container output missing %q: %s", want, got)
		}
	}
	if !strings.Contains(got, `data-name="tab"`) {
		t.Fatalf("container lost the child shortcode name: %s", got)
	}
}

// Registering a name as one kind must retire the other, so a shortcode never
// resolves to both a component and a container.
func TestRegisteringAComponentAndContainerUnderOneNameKeepsTheLast(t *testing.T) {
	r := NewRouter()
	if err := r.RegisterMarkdownComponent("thing", func(_ map[string]string, body render.HTML) render.HTML { return body }); err != nil {
		t.Fatalf("RegisterMarkdownComponent() error = %v", err)
	}
	if err := r.RegisterMarkdownContainer("thing", func(_ map[string]string, _ []MarkdownChild, body render.HTML) render.HTML { return body }); err != nil {
		t.Fatalf("RegisterMarkdownContainer() error = %v", err)
	}
	if _, ok := r.MarkdownComponents()["thing"]; ok {
		t.Fatal("registering a container left the component in place")
	}
	if !r.vocabulary(nil, nil).isContainer("thing") {
		t.Fatal("container did not take over the name")
	}

	if err := r.RegisterMarkdownComponent("thing", func(_ map[string]string, body render.HTML) render.HTML { return body }); err != nil {
		t.Fatalf("RegisterMarkdownComponent() error = %v", err)
	}
	if _, ok := r.MarkdownContainers()["thing"]; ok {
		t.Fatal("re-registering a component left the container in place")
	}
}
