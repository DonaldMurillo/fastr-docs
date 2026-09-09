package docs

import (
	"fmt"

	"github.com/DonaldMurillo/gofastr/core/render"
)

// MarkdownSourceTransform rewrites Markdown source before the parser runs.
//
// A shortcode can only claim text a writer wrapped in {{< >}}. Some syntax is
// ambient instead: math delimited by $ appears mid-sentence, and rewriting it
// after parsing would mean pattern-matching rendered HTML, which cannot tell a
// dollar sign in prose from one in a code span.
//
// place takes rendered HTML and returns an opaque marker to splice into the
// source in its stead. The Markdown parser only ever sees the marker, so the
// fragment survives untouched and nothing inside it is parsed as Markdown.
// This is the same mechanism option-carrying code fences already use.
type MarkdownSourceTransform func(source string, place func(render.HTML) string) string

// RegisterMarkdownSourceTransform adds a source rewriter, which runs on every
// Markdown page before shortcodes expand. Registering the same name twice
// replaces the earlier transform rather than running both.
//
// Transforms run in registration order. Keep one narrow: a transform that
// claims text another one wanted makes the two order-dependent.
func (r *Router) RegisterMarkdownSourceTransform(name string, transform MarkdownSourceTransform) error {
	name, err := r.prepareShortcodeRegistration("RegisterMarkdownSourceTransform", "Markdown source transform", name)
	if err != nil {
		return err
	}
	if transform == nil {
		return fmt.Errorf("docs: Markdown source transform %q is nil", name)
	}
	for i, existing := range r.markdownTransforms {
		if existing.name == name {
			r.markdownTransforms[i].transform = transform
			return nil
		}
	}
	r.markdownTransforms = append(r.markdownTransforms, namedSourceTransform{name: name, transform: transform})
	return nil
}

// MarkdownSourceTransformNames lists the registered transforms in the order
// they run.
func (r *Router) MarkdownSourceTransformNames() []string {
	if r == nil {
		return nil
	}
	names := make([]string, 0, len(r.markdownTransforms))
	for _, entry := range r.markdownTransforms {
		names = append(names, entry.name)
	}
	return names
}

type namedSourceTransform struct {
	name      string
	transform MarkdownSourceTransform
}

// applyMarkdownTransforms runs every transform over the source and returns the
// fragments each one lifted out.
func applyMarkdownTransforms(source string, transforms []namedSourceTransform) (string, []shortcodeReplacement) {
	if len(transforms) == 0 {
		return source, nil
	}
	var replacements []shortcodeReplacement
	place := func(html render.HTML) string {
		marker := fmt.Sprintf("FASTRDOCSTRANSFORMSLOT%dEND", len(replacements))
		replacements = append(replacements, shortcodeReplacement{marker: marker, html: html})
		return marker
	}
	for _, entry := range transforms {
		source = entry.transform(source, place)
	}
	return source, replacements
}
