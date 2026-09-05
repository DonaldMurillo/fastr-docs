package docs

import (
	"strings"
	"testing"

	"github.com/DonaldMurillo/gofastr/core/render"
)

func TestRegisterMarkdownSourceTransformRejectsBadInput(t *testing.T) {
	r := NewRouter()
	if err := r.RegisterMarkdownSourceTransform("", func(s string, _ func(render.HTML) string) string { return s }); err == nil {
		t.Fatal("an empty name was accepted")
	}
	if err := r.RegisterMarkdownSourceTransform("math", nil); err == nil {
		t.Fatal("a nil transform was accepted")
	}
}

// Registering the same name twice replaces the transform. Running both would
// mean a plugin re-registered on a shared Router silently doubles its work.
func TestRegisterMarkdownSourceTransformReplacesByName(t *testing.T) {
	r := NewRouter()
	for i := 0; i < 3; i++ {
		if err := r.RegisterMarkdownSourceTransform("math", func(s string, _ func(render.HTML) string) string { return s }); err != nil {
			t.Fatalf("RegisterMarkdownSourceTransform() error = %v", err)
		}
	}
	if names := r.MarkdownSourceTransformNames(); len(names) != 1 || names[0] != "math" {
		t.Fatalf("names = %v, want [math]", names)
	}
}

func TestTransformsRunInRegistrationOrder(t *testing.T) {
	r := NewRouter()
	var order []string
	for _, name := range []string{"first", "second", "third"} {
		if err := r.RegisterMarkdownSourceTransform(name, func(s string, _ func(render.HTML) string) string {
			order = append(order, name)
			return s
		}); err != nil {
			t.Fatalf("RegisterMarkdownSourceTransform() error = %v", err)
		}
	}
	if _, _ = applyMarkdownTransforms("x", r.markdownTransforms); strings.Join(order, ",") != "first,second,third" {
		t.Fatalf("ran in order %v", order)
	}
}

// The point of the hook: a transform's HTML reaches the page whole, without the
// Markdown parser ever seeing it.
func TestTransformOutputSurvivesTheMarkdownParser(t *testing.T) {
	r := NewRouter()
	err := r.RegisterMarkdownSourceTransform("shout", func(source string, place func(render.HTML) string) string {
		return strings.ReplaceAll(source, "!!", place(render.HTML(`<b class="shout">*not markdown*</b>`)))
	})
	if err != nil {
		t.Fatalf("RegisterMarkdownSourceTransform() error = %v", err)
	}
	html, err := renderMarkdownWithComponents("Hello !! world", r.vocabulary(nil, nil), nil)
	if err != nil {
		t.Fatalf("renderMarkdownWithComponents() error = %v", err)
	}
	body := string(html)
	if !strings.Contains(body, `<b class="shout">*not markdown*</b>`) {
		t.Fatalf("fragment did not survive: %s", body)
	}
	// The asterisks would be emphasis had the parser seen them.
	if strings.Contains(body, "<em>") {
		t.Fatalf("the parser reached inside the fragment: %s", body)
	}
	if !strings.Contains(body, "Hello ") || !strings.Contains(body, " world") {
		t.Fatalf("surrounding prose was lost: %s", body)
	}
}

// A transform and a shortcode on the same page must not overwrite each other's
// markers. Both mechanisms number their slots from zero.
func TestTransformAndShortcodeMarkersDoNotCollide(t *testing.T) {
	r := NewRouter()
	if err := r.RegisterMarkdownSourceTransform("mark", func(source string, place func(render.HTML) string) string {
		return strings.ReplaceAll(source, "@@", place(render.HTML(`<i>transformed</i>`)))
	}); err != nil {
		t.Fatalf("RegisterMarkdownSourceTransform() error = %v", err)
	}
	if err := r.RegisterMarkdownComponent("shout", func(_ map[string]string, body render.HTML) render.HTML {
		return render.HTML(`<b>` + string(body) + `</b>`)
	}); err != nil {
		t.Fatalf("RegisterMarkdownComponent() error = %v", err)
	}
	source := "Start @@ then {{< shout >}}loud{{< /shout >}} and @@ again."
	html, err := renderMarkdownWithComponents(source, r.vocabulary(nil, nil), nil)
	if err != nil {
		t.Fatalf("renderMarkdownWithComponents() error = %v", err)
	}
	body := string(html)
	if strings.Count(body, "<i>transformed</i>") != 2 {
		t.Fatalf("transform fragments were lost or duplicated: %s", body)
	}
	if !strings.Contains(body, "<b>") || !strings.Contains(body, "loud") {
		t.Fatalf("shortcode did not render: %s", body)
	}
	if strings.Contains(body, "SLOT") {
		t.Fatalf("a marker leaked to the reader: %s", body)
	}
}
