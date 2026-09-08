package docs

import (
	"fmt"
	"strconv"
	"strings"
	"unicode"

	"github.com/DonaldMurillo/gofastr/core-ui/html"
	"github.com/DonaldMurillo/gofastr/core/render"
	"github.com/DonaldMurillo/gofastr/framework/ui"
)

// Heading is a Markdown heading exposed to a page table of contents.
type Heading struct {
	// ID is the anchor fragment the heading renders with.
	ID string
	// Title is the heading text.
	Title string
	// Level is 1 to 6; the rail and select carry the 2s and 3s.
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
	// A landing page is not a document. Its git date is the date the hero was
	// last edited, and "Updated" floating above a hero reads as noise.
	if route.Metadata.PageTemplate != PageTemplateSplash {
		if metadata := r.docMetadata(route); metadata != "" {
			body = render.Join(metadata, body)
		}
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
	// A date from git is an RFC 3339 timestamp; the reader gets it in the
	// page's language and DateFormat, and the machine-readable value stays
	// on the element.
	display := func(raw string) string {
		if date, ok := parseBlogDate(raw); ok {
			return labels.formatDate(date)
		}
		return raw
	}
	if meta.DateModified != "" {
		parts = append(parts, render.Join(
			render.Text(labels.LastUpdated+" "),
			render.Tag("time", map[string]string{"datetime": meta.DateModified}, render.Text(display(meta.DateModified))),
		))
	} else if meta.DatePublished != "" {
		parts = append(parts, render.Join(
			render.Text(labels.Published+" "),
			render.Tag("time", map[string]string{"datetime": meta.DatePublished}, render.Text(display(meta.DatePublished))),
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
		// h2 and h3 drive the rail; h4 joins when it is the deepest level
		// the page uses, so a page organized entirely in h4s still gets a
		// table of contents instead of none at all.
		if level < 2 || level > 4 {
			continue
		}
		title := plainHeadingTitle(strings.TrimSpace(strings.TrimLeft(trimmed, "# ")))
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
// headingAnchorButtons adds a copyable anchor to every h2 and h3 the page
// carries. Headings come from GoFastr's renderer without one, and a reader
// linking a section by hand has to fish the URL out of the TOC otherwise.
func headingAnchorButtons(markdown string) string {
	var out strings.Builder
	for offset := 0; offset < len(markdown); {
		next := nextHeadingWithID(markdown[offset:], 2, 3, 4)
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
		closeRel := strings.Index(markdown[idEnd:], "</h")
		if closeRel < 0 {
			out.WriteString(markdown[offset:])
			break
		}
		id := markdown[idStart:idEnd]
		// The button lands AFTER the closing tag, not inside the heading:
		// a labeled control inside a heading joins the heading's own
		// accessible name, and "Start here" stops matching exactly.
		closer := markdown[idEnd+closeRel:]
		closerEnd := strings.IndexByte(closer, '>')
		if closerEnd < 0 {
			out.WriteString(markdown[offset:])
			break
		}
		out.WriteString(markdown[offset : idEnd+closeRel+closerEnd+1])
		// A button, not a link: the anchor copies the URL rather than
		// scrolling (the reader is already here), and it is reachable by
		// keyboard with a name a screen reader announces.
		out.WriteString(`<button type="button" class="heading-anchor" data-fastr-docs-anchor="#` + render.Escape(id) + `" aria-label="Copy link to this section">#</button>`)
		offset = idEnd + closeRel + closerEnd + 1
	}
	return out.String()
}

func dedupeMarkdownHeadingIDs(markdown string) string {
	counts := make(map[string]int)
	var out strings.Builder
	for offset := 0; offset < len(markdown); {
		next := nextHeadingWithID(markdown[offset:], 2, 3, 4)
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
		closeRel := strings.Index(markdown[idEnd:], "</h")
		if closeRel < 0 {
			out.WriteString(markdown[offset:])
			break
		}
		out.WriteString(markdown[offset:idStart])
		// The rendered id becomes the folded slug of the heading's own
		// text, so the id, the toc, the rail, and a reader typing the
		// ASCII form of an accented word all agree. GoFastr's slugger
		// keeps the accents; the rail aims at the folded form, and a
		// mismatch there crashes the scrollspy.
		base := headingSlug(plainText(markdown[idEnd : idEnd+closeRel]))
		if base == "" {
			base = markdown[idStart:idEnd]
		}
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

// plainText strips the tags from a rendered fragment, leaving the words a
// reader sees.
func plainText(fragment string) string {
	var b strings.Builder
	depth := 0
	for _, r := range fragment {
		switch {
		case r == '<':
			depth++
		case r == '>':
			if depth > 0 {
				depth--
			}
		case depth == 0:
			b.WriteRune(r)
		}
	}
	return b.String()
}

func headingSlug(text string) string {
	var out strings.Builder
	previousDash := true
	// Folded like search and tags: the anchor for "Diseño" is #diseno, so
	// a reader typing the ASCII form lands on the section.
	for _, char := range foldRunes(text) {
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

// nextHeadingWithID finds the next hN-with-id opening tag for any of the
// levels given, or -1.
func nextHeadingWithID(markdown string, levels ...int) int {
	next := -1
	for _, level := range levels {
		at := strings.Index(markdown, fmt.Sprintf(`<h%d id="`, level))
		if at >= 0 && (next < 0 || at < next) {
			next = at
		}
	}
	return next
}

// plainHeadingTitle strips the syntax a writer can leave in a heading:
// shortcode markers and emphasis pairs. The table of contents and the slug
// should show words, not markup.
func plainHeadingTitle(title string) string {
	title = markdownShortcodeToken.ReplaceAllString(title, "")
	for _, marker := range []string{"**", "__"} {
		title = strings.ReplaceAll(title, marker, "")
	}
	return strings.TrimSpace(title)
}
