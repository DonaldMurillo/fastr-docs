package docs

import (
	"strconv"
	"strings"
	"unicode"

	"github.com/DonaldMurillo/gofastr/core-ui/html"
	"github.com/DonaldMurillo/gofastr/core/render"
	"github.com/DonaldMurillo/gofastr/framework/ui"
)

// Heading is a Markdown heading exposed to a page table of contents.
type Heading struct {
	ID    string
	Title string
	Level int
}

func (r *Router) wrapDocPage(route *Route, body render.HTML, headings []Heading) render.HTML {
	if editURL := safeMetadataURL(route.Metadata.EditURL); editURL != "" {
		body = render.Join(body, render.Tag("p", map[string]string{"class": "fastr-docs-edit-link"},
			render.Tag("a", map[string]string{
				"href": editURL, "rel": "nofollow noopener", "target": "_blank",
			}, render.Text("Edit this page ↗")),
		))
	}
	cfg := ui.DocLayoutConfig{
		Crumbs: r.docCrumbs(route),
		Pager:  r.docPager(route),
		Class:  "fastr-docs-doc-layout",
	}
	if len(headings) > 0 {
		items := make([]ui.RailItem, 0, len(headings))
		for _, heading := range headings {
			items = append(items, ui.RailItem{Anchor: heading.ID, Text: heading.Title})
		}
		if len(items) > 0 {
			rail := ui.AnchoredRail(ui.AnchoredRailConfig{
				Label:           "On this page",
				Items:           items,
				ObserveSelector: ".ui-doc-layout__content",
				TargetSelector:  "h2[id], h3[id]",
				Class:           "fastr-docs-toc fastr-docs-toc--rail",
			})
			cfg.Toc = render.Join(rail, docsTocSelect(headings))
		}
	}
	return ui.DocLayout(cfg, body)
}

func docsTocSelect(headings []Heading) render.HTML {
	options := make([]ui.SelectOption, 0, len(headings))
	for i, heading := range headings {
		options = append(options, ui.SelectOption{
			Value:    "#" + heading.ID,
			Text:     heading.Title,
			Selected: i == 0,
		})
	}
	return ui.Select(ui.SelectConfig{
		Name:    "docs-toc",
		ID:      "fastr-docs-toc-select",
		Label:   "On this page",
		Options: options,
		Class:   "fastr-docs-toc-select",
		ExtraAttrs: html.Attrs{
			"data-docs-toc-select": "true",
		},
	})
}

func (r *Router) docCrumbs(route *Route) []ui.DocCrumb {
	if route == nil {
		return nil
	}
	if route.Path == "/" {
		return nil
	}
	crumbs := []ui.DocCrumb{{Label: r.SiteName(), Href: "/"}}
	var ancestors []*Route
	for parent := route.Parent; parent != nil; parent = parent.Parent {
		if parent.Kind != KindGroup && parent.Path != "/" && r.routePublished(parent) {
			ancestors = append(ancestors, parent)
		}
	}
	for i := len(ancestors) - 1; i >= 0; i-- {
		crumbs = append(crumbs, ui.DocCrumb{Label: ancestors[i].Title, Href: ancestors[i].Path})
	}
	crumbs = append(crumbs, ui.DocCrumb{Label: route.Title})
	return crumbs
}

func (r *Router) docPager(route *Route) *ui.DocPager {
	if route == nil || route.Path == "/" {
		return nil
	}
	visible := r.PublishedRoutes()
	index := -1
	for i, candidate := range visible {
		if candidate.Path == route.Path {
			index = i
			break
		}
	}
	if index <= 0 {
		return nil
	}
	pager := &ui.DocPager{PrevHref: visible[index-1].Path, PrevLabel: visible[index-1].Title}
	if index+1 < len(visible) {
		pager.NextHref = visible[index+1].Path
		pager.NextLabel = visible[index+1].Title
	}
	return pager
}

func markdownHeadings(source string) []Heading {
	var headings []Heading
	seen := make(map[string]int)
	inFence := false
	for _, line := range strings.Split(strings.ReplaceAll(source, "\r\n", "\n"), "\n") {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "```") || strings.HasPrefix(trimmed, "~~~") {
			inFence = !inFence
			continue
		}
		if inFence || !strings.HasPrefix(trimmed, "##") {
			continue
		}
		level := 0
		for level < len(trimmed) && trimmed[level] == '#' {
			level++
		}
		if level < 2 || level > 3 {
			continue
		}
		title := strings.TrimSpace(strings.TrimLeft(trimmed, "# "))
		if title == "" {
			continue
		}
		id := headingSlug(title)
		if id == "" {
			continue
		}
		seen[id]++
		if seen[id] > 1 {
			id += "-" + strconv.Itoa(seen[id])
		}
		headings = append(headings, Heading{ID: id, Title: title, Level: level})
	}
	return headings
}

// dedupeMarkdownHeadingIDs keeps the framework Markdown renderer as the
// source of HTML, then repairs only the generated h2/h3 IDs so duplicate
// headings remain valid HTML and every scrollspy item has a unique target.
// GoFastr's slugger is intentionally stable and does not suffix duplicates.
func dedupeMarkdownHeadingIDs(markdown string) string {
	counts := make(map[string]int)
	var out strings.Builder
	for offset := 0; offset < len(markdown); {
		next2 := strings.Index(markdown[offset:], `<h2 id="`)
		next3 := strings.Index(markdown[offset:], `<h3 id="`)
		next := -1
		if next2 >= 0 {
			next = next2
		}
		if next3 >= 0 && (next < 0 || next3 < next) {
			next = next3
		}
		if next < 0 {
			out.WriteString(markdown[offset:])
			break
		}
		next += offset
		idStart := next + strings.Index(markdown[next:], ` id="`) + len(` id="`)
		idEndRel := strings.IndexByte(markdown[idStart:], '"')
		if idEndRel < 0 {
			out.WriteString(markdown[offset:])
			break
		}
		idEnd := idStart + idEndRel
		out.WriteString(markdown[offset:idStart])
		base := markdown[idStart:idEnd]
		counts[base]++
		if counts[base] == 1 {
			out.WriteString(base)
		} else {
			out.WriteString(base + "-" + strconv.Itoa(counts[base]))
		}
		offset = idEnd
	}
	return out.String()
}

func headingSlug(text string) string {
	var out strings.Builder
	previousDash := true
	for _, char := range strings.ToLower(text) {
		switch {
		case unicode.IsLetter(char) || unicode.IsDigit(char):
			out.WriteRune(char)
			previousDash = false
		case char == ' ' || char == '-' || char == '_':
			if !previousDash {
				out.WriteByte('-')
				previousDash = true
			}
		}
	}
	return strings.Trim(out.String(), "-")
}
