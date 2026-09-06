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
			}, render.Text(r.uiForRoute(route).EditPage)),
		))
	}
	if metadata := r.docMetadata(route); metadata != "" {
		body = render.Join(metadata, body)
	}
	cfg := ui.DocLayoutConfig{
		Crumbs: r.docCrumbs(route),
		Class:  "fastr-docs-doc-layout",
		Pager:  r.docPager(route),
	}
	if route.Metadata.PageTemplate == PageTemplateSplash {
		// A landing page drops the breadcrumb trail and widens its column. The
		// table of contents is already suppressed before this is called.
		cfg.Crumbs = nil
		cfg.Class += " fastr-docs-splash"
	}
	if len(headings) > 0 {
		items := make([]ui.RailItem, 0, len(headings))
		for _, heading := range headings {
			items = append(items, ui.RailItem{Anchor: heading.ID, Text: heading.Title})
		}
		if len(items) > 0 {
			rail := ui.AnchoredRail(ui.AnchoredRailConfig{
				Label:           r.uiForRoute(route).OnThisPage,
				Items:           items,
				ObserveSelector: ".ui-doc-layout__content",
				TargetSelector:  "h2[id], h3[id]",
				Class:           "fastr-docs-toc fastr-docs-toc--rail",
			})
			cfg.Toc = render.Join(rail, r.docsTocSelect(headings, r.uiForRoute(route).OnThisPage))
		}
	}
	return ui.DocLayout(cfg, body)
}

func (r *Router) docsTocSelect(headings []Heading, label string) render.HTML {
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
		Label:   label,
		Options: options,
		Class:   "fastr-docs-toc-select",
		ExtraAttrs: html.Attrs{
			"data-docs-toc-select": "true",
		},
	})
}

func (r *Router) docMetadata(route *Route) render.HTML {
	if route == nil {
		return ""
	}
	meta := route.Metadata
	if meta.DateModified == "" && meta.DatePublished == "" && len(meta.Authors) == 0 {
		return ""
	}
	labels := r.uiForRoute(route)
	parts := make([]render.HTML, 0, 3)
	if meta.DateModified != "" {
		parts = append(parts, render.Join(
			render.Text(labels.LastUpdated+" "),
			render.Tag("time", map[string]string{"datetime": meta.DateModified}, render.Text(meta.DateModified)),
		))
	} else if meta.DatePublished != "" {
		parts = append(parts, render.Join(
			render.Text(labels.Published+" "),
			render.Tag("time", map[string]string{"datetime": meta.DatePublished}, render.Text(meta.DatePublished)),
		))
	}
	if len(meta.Authors) > 0 {
		parts = append(parts, render.Text(labels.By+" "+strings.Join(meta.Authors, ", ")))
	}
	content := make([]render.HTML, 0, len(parts)*2-1)
	for i, part := range parts {
		if i > 0 {
			content = append(content, render.Text(" · "))
		}
		content = append(content, part)
	}
	return render.Tag("p", map[string]string{
		"class": "fastr-docs-page-meta", "data-docs-page-meta": "true",
	}, content...)
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

// docPager builds the previous/next footer.
//
// Neighbours are drawn from the reader's own language. Walking every published
// route sent a reader at the end of the Spanish tree into the English one, and
// the card showed an English title with no hint that the language had changed.
// The direction lines come from that language's strings.
func (r *Router) docPager(route *Route) *ui.DocPager {
	if route == nil || route.Path == "/" {
		return nil
	}
	locale := r.effectiveLocale(route)
	siblings := make([]*Route, 0, len(r.PublishedRoutes()))
	for _, candidate := range r.PublishedRoutes() {
		if r.effectiveLocale(candidate) == locale {
			siblings = append(siblings, candidate)
		}
	}
	index := -1
	for i, candidate := range siblings {
		if candidate.Path == route.Path {
			index = i
			break
		}
	}
	if index <= 0 {
		return nil
	}
	labels := r.uiForRoute(route)
	pager := &ui.DocPager{
		PrevHref: siblings[index-1].Path, PrevLabel: siblings[index-1].Title,
		PrevDirLabel: labels.Previous, NextDirLabel: labels.Next,
	}
	if index+1 < len(siblings) {
		pager.NextHref, pager.NextLabel = siblings[index+1].Path, siblings[index+1].Title
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
