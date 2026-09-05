package docs

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/DonaldMurillo/gofastr/core/render"
)

// MarkdownComponent renders a typed component embedded in Markdown. Props are
// parsed from the shortcode's key/value attributes and body is already
// rendered Markdown, so a component can wrap headings, links, and prose with
// GoFastr UI primitives.
//
// Example:
//
//	{{< callout variant="warning" >}}
//	Keep the route order explicit.
//	{{< /callout >}}
type MarkdownComponent func(props map[string]string, body render.HTML) render.HTML

// MarkdownChild is one shortcode nested directly inside a container. Body is
// that child's own rendered output, so a child may itself be a component or
// another container.
type MarkdownChild struct {
	Name  string
	Props map[string]string
	Body  render.HTML
}

// MarkdownContainer renders a shortcode that needs its nested shortcodes as
// separate items rather than as one merged blob. Tabs, card grids, steps, and
// file trees all have to know where one child ends and the next begins, which
// a MarkdownComponent cannot see.
//
// Children are the shortcodes nested directly inside it. Body is whatever
// Markdown sat between them, so a container can support both shapes.
//
// Example:
//
//	{{< tabs >}}
//	{{< tab label="Go" >}}...{{< /tab >}}
//	{{< tab label="Shell" >}}...{{< /tab >}}
//	{{< /tabs >}}
type MarkdownContainer func(props map[string]string, children []MarkdownChild, body render.HTML) render.HTML

// markdownVocabulary is the merged shortcode namespace for one page. A name
// resolves to either a component or a container, never both.
type markdownVocabulary struct {
	components map[string]MarkdownComponent
	containers map[string]MarkdownContainer
}

func (v markdownVocabulary) empty() bool {
	return len(v.components) == 0 && len(v.containers) == 0
}

func (v markdownVocabulary) knows(name string) bool {
	if _, ok := v.containers[name]; ok {
		return true
	}
	_, ok := v.components[name]
	return ok
}

func (v markdownVocabulary) isContainer(name string) bool {
	_, ok := v.containers[name]
	return ok
}

func (v markdownVocabulary) render(node markdownShortcode, children []MarkdownChild, body render.HTML) render.HTML {
	if container, ok := v.containers[node.name]; ok {
		return container(node.props, children, body)
	}
	return v.components[node.name](node.props, body)
}

type markdownShortcode struct {
	name  string
	props map[string]string
	body  string
}

type shortcodeReplacement struct {
	marker string
	html   render.HTML
}

var markdownShortcodeName = regexp.MustCompile(`^[A-Za-z][A-Za-z0-9_-]*$`)
var markdownShortcodeToken = regexp.MustCompile(`\{\{<\s*(/)?([A-Za-z][A-Za-z0-9_-]*)([^>]*)>\}\}`)
var markdownShortcodeProp = regexp.MustCompile(`([A-Za-z][A-Za-z0-9_-]*)\s*=\s*(?:"([^"]*)"|'([^']*)'|([^\s]+))`)

func validateMarkdownComponents(source string, vocab markdownVocabulary) error {
	_, _, err := expandMarkdownShortcodes(source, vocab, false)
	return err
}

func expandMarkdownShortcodes(source string, vocab markdownVocabulary, renderBody bool) (string, []shortcodeReplacement, error) {
	if !markdownShortcodeToken.MatchString(source) {
		return source, nil, nil
	}
	type frame struct {
		node     markdownShortcode
		start    int
		children []MarkdownChild
	}
	var stack []frame
	var root strings.Builder
	var replacements []shortcodeReplacement
	last := 0
	appendText := func(text string) {
		if len(stack) == 0 {
			root.WriteString(text)
			return
		}
		stack[len(stack)-1].node.body += text
	}
	appendRendered := func(current frame) error {
		node := current.node
		if !vocab.knows(node.name) {
			return fmt.Errorf("Markdown component %q is not registered", node.name)
		}
		bodySource, nested, err := expandMarkdownShortcodes(node.body, vocab, true)
		if err != nil {
			return err
		}
		body := render.HTML("")
		if renderBody || strings.TrimSpace(bodySource) != "" {
			body = applyShortcodes(renderDocsMarkdown(bodySource, nil), nested)
		}
		html := vocab.render(node, current.children, body)
		// Inside a container the rendered child is handed over as its own item
		// rather than flattened into the parent's body, so the parent can tell
		// one child from the next.
		if len(stack) > 0 && vocab.isContainer(stack[len(stack)-1].node.name) {
			parent := &stack[len(stack)-1]
			parent.children = append(parent.children, MarkdownChild{Name: node.name, Props: node.props, Body: html})
			return nil
		}
		marker := fmt.Sprintf(`FASTRDOCSSHORTCODESLOT%dEND`, len(replacements))
		replacements = append(replacements, shortcodeReplacement{marker: marker, html: html})
		appendText(marker)
		return nil
	}
	for _, match := range markdownShortcodeToken.FindAllStringSubmatchIndex(source, -1) {
		appendText(source[last:match[0]])
		if markdownOffsetInFence(source, match[0]) {
			appendText(source[match[0]:match[1]])
			last = match[1]
			continue
		}
		closing := match[2] >= 0 && source[match[2]:match[3]] == "/"
		name := source[match[4]:match[5]]
		attrs := strings.TrimSpace(source[match[6]:match[7]])
		selfClosing := strings.HasSuffix(attrs, "/")
		attrs = strings.TrimSpace(strings.TrimSuffix(attrs, "/"))
		if closing {
			if len(stack) == 0 || stack[len(stack)-1].node.name != name {
				return "", nil, fmt.Errorf("Markdown component closing tag %q does not match its opener", name)
			}
			closed := stack[len(stack)-1]
			stack = stack[:len(stack)-1]
			if err := appendRendered(closed); err != nil {
				return "", nil, err
			}
		} else {
			props := make(map[string]string)
			for _, prop := range markdownShortcodeProp.FindAllStringSubmatch(attrs, -1) {
				value := prop[2]
				if value == "" {
					value = prop[3]
				}
				if value == "" {
					value = prop[4]
				}
				props[prop[1]] = value
			}
			node := markdownShortcode{name: name, props: props}
			if selfClosing {
				if err := appendRendered(frame{node: node, start: match[0]}); err != nil {
					return "", nil, err
				}
			} else {
				stack = append(stack, frame{node: node, start: match[0]})
			}
		}
		last = match[1]
	}
	if len(stack) > 0 {
		return "", nil, fmt.Errorf("Markdown component %q is missing a closing tag", stack[len(stack)-1].node.name)
	}
	appendText(source[last:])
	return root.String(), replacements, nil
}

func markdownOffsetInFence(source string, offset int) bool {
	inFence := false
	position := 0
	for _, line := range strings.SplitAfter(strings.ReplaceAll(source, "\r\n", "\n"), "\n") {
		end := position + len(line)
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "```") || strings.HasPrefix(trimmed, "~~~") {
			if offset >= position && offset < end {
				return inFence
			}
			inFence = !inFence
		}
		if offset >= position && offset < end {
			return inFence
		}
		position = end
	}
	return inFence
}

func renderMarkdownWithComponents(source string, vocab markdownVocabulary, attrs map[string]string) (render.HTML, error) {
	expanded, replacements, err := expandMarkdownShortcodes(source, vocab, false)
	if err != nil {
		return "", err
	}
	return applyShortcodes(renderDocsMarkdown(expanded, attrs), replacements), nil
}

// applyShortcodes substitutes rendered components back into the document.
//
// Order matters. Replacements are collected as each shortcode closes, so a
// child always lands before its parent. A child's marker only appears in the
// output once the parent's HTML has been inserted, so substituting in
// collection order leaves the child marker stranded as visible text. Walking
// backwards inserts each parent first and resolves its children afterwards, to
// any depth. It also settles the prefix collision between marker 1 and marker
// 10, since the higher index is always replaced first.
func applyShortcodes(html render.HTML, replacements []shortcodeReplacement) render.HTML {
	for i := len(replacements) - 1; i >= 0; i-- {
		html = replaceShortcode(html, replacements[i])
	}
	return html
}

func replaceShortcode(source render.HTML, replacement shortcodeReplacement) render.HTML {
	html := string(source)
	// A block shortcode is commonly surrounded by blank lines and Goldmark
	// emits it as a paragraph. Remove that paragraph before inserting the
	// component so the result remains valid HTML.
	html = strings.ReplaceAll(html, "<p>"+replacement.marker+"</p>", string(replacement.html))
	return render.HTML(strings.ReplaceAll(html, replacement.marker, string(replacement.html)))
}

func cloneMarkdownComponents(in map[string]MarkdownComponent) map[string]MarkdownComponent {
	if len(in) == 0 {
		return nil
	}
	out := make(map[string]MarkdownComponent, len(in))
	for name, component := range in {
		out[name] = component
	}
	return out
}

func mergeMarkdownComponents(global, local map[string]MarkdownComponent) map[string]MarkdownComponent {
	if len(global) == 0 && len(local) == 0 {
		return nil
	}
	out := cloneMarkdownComponents(global)
	if out == nil {
		out = make(map[string]MarkdownComponent)
	}
	for name, component := range local {
		out[name] = component
	}
	return out
}

func cloneMarkdownContainers(in map[string]MarkdownContainer) map[string]MarkdownContainer {
	if len(in) == 0 {
		return nil
	}
	out := make(map[string]MarkdownContainer, len(in))
	for name, container := range in {
		out[name] = container
	}
	return out
}

func mergeMarkdownContainers(global, local map[string]MarkdownContainer) map[string]MarkdownContainer {
	if len(global) == 0 && len(local) == 0 {
		return nil
	}
	out := cloneMarkdownContainers(global)
	if out == nil {
		out = make(map[string]MarkdownContainer)
	}
	for name, container := range local {
		out[name] = container
	}
	return out
}
