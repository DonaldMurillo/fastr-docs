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

func validateMarkdownComponents(source string, components map[string]MarkdownComponent) error {
	_, _, err := expandMarkdownShortcodes(source, components, false)
	return err
}

func expandMarkdownShortcodes(source string, components map[string]MarkdownComponent, renderBody bool) (string, []shortcodeReplacement, error) {
	if !markdownShortcodeToken.MatchString(source) {
		return source, nil, nil
	}
	type frame struct {
		node  markdownShortcode
		start int
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
	appendRendered := func(node markdownShortcode) error {
		component := components[node.name]
		if component == nil {
			return fmt.Errorf("Markdown component %q is not registered", node.name)
		}
		bodySource, nested, err := expandMarkdownShortcodes(node.body, components, true)
		if err != nil {
			return err
		}
		body := render.HTML("")
		if renderBody || strings.TrimSpace(bodySource) != "" {
			body = renderDocsMarkdown(bodySource, nil)
			for _, replacement := range nested {
				body = replaceShortcode(body, replacement)
			}
		}
		html := component(node.props, body)
		marker := fmt.Sprintf(`FASTRDOCSSHORTCODESLOT%d`, len(replacements))
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
			frame := stack[len(stack)-1]
			stack = stack[:len(stack)-1]
			if err := appendRendered(frame.node); err != nil {
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
				if err := appendRendered(node); err != nil {
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

func renderMarkdownWithComponents(source string, components map[string]MarkdownComponent, attrs map[string]string) (render.HTML, error) {
	expanded, replacements, err := expandMarkdownShortcodes(source, components, false)
	if err != nil {
		return "", err
	}
	html := renderDocsMarkdown(expanded, attrs)
	for _, replacement := range replacements {
		html = replaceShortcode(html, replacement)
	}
	return html, nil
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
