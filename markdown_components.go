package docs

import (
	"strconv"
	"strings"
	"sync/atomic"

	"github.com/DonaldMurillo/gofastr/core-ui/patterns/tabs"
	"github.com/DonaldMurillo/gofastr/core/render"
	"github.com/DonaldMurillo/gofastr/framework/ui"
)

// Shortcode props come from Markdown, which means they come from whoever is
// writing the docs. Several GoFastr components panic on input they consider
// invalid: an unregistered StatusVariant panics at render, and tabs.New panics
// on an empty group name or zero tabs. Every value below is therefore mapped
// through a fixed table or guarded, so a typo in a Markdown file renders
// something reasonable instead of taking down the server.

// tabGroupSeq names each tab set. tabs.New requires a name that is unique
// within the page, and these renderers are shared across requests, so the
// counter has to be atomic.
var tabGroupSeq atomic.Uint64

// DefaultMarkdownComponents is the shortcode vocabulary every Router registers
// unless WithoutDefaultComponents is passed. A project overrides any of these
// by registering the same name, which replaces rather than conflicts.
func DefaultMarkdownComponents() map[string]MarkdownComponent {
	components := map[string]MarkdownComponent{
		"callout": calloutComponent(""),
		"card":    cardComponent,
		"details": detailsComponent,
		"badge":   badgeComponent,
		"tag":     tagComponent,
		"banner":  bannerComponent,
		"icon":    iconComponent,
		"steps":   wrapperComponent("fastr-docs-steps"),
		// Slots. They carry props for a parent container and render as their
		// own body when used on their own.
		"tab":    slotComponent,
		"action": actionComponent,
	}
	// The admonition names authors actually reach for, each pinned to a tone.
	for name, variant := range map[string]ui.StatusVariant{
		"note":    ui.StatusInfo,
		"info":    ui.StatusInfo,
		"tip":     ui.StatusSuccess,
		"success": ui.StatusSuccess,
		"warning": ui.StatusWarning,
		"caution": ui.StatusWarning,
		"danger":  ui.StatusDanger,
	} {
		components[name] = calloutComponent(variant)
	}
	return components
}

// DefaultMarkdownRawComponents is the shortcode vocabulary whose bodies are
// data rather than prose, so they must not be rendered as Markdown first.
func DefaultMarkdownRawComponents() map[string]MarkdownRawComponent {
	return map[string]MarkdownRawComponent{
		"diff":     diffComponent,
		"filetree": fileTreeComponent,
	}
}

// DefaultMarkdownContainers is the shortcode vocabulary that needs to see its
// children individually.
func DefaultMarkdownContainers() map[string]MarkdownContainer {
	return map[string]MarkdownContainer{
		"tabs":  tabsContainer,
		"cards": gridContainer,
		"grid":  gridContainer,
		"hero":  heroContainer,
	}
}

// WithoutDefaultComponents drops the built-in shortcode vocabulary. Use it when
// a project wants Markdown to fail on any shortcode it did not register itself.
func WithoutDefaultComponents() Option {
	return func(r *Router) {
		r.markdownComponents = make(map[string]MarkdownComponent)
		r.markdownContainers = make(map[string]MarkdownContainer)
		r.markdownRaws = make(map[string]MarkdownRawComponent)
	}
}

func calloutComponent(fixed ui.StatusVariant) MarkdownComponent {
	return func(props map[string]string, body render.HTML) render.HTML {
		variant := fixed
		if variant == "" {
			variant = statusVariant(props["variant"], ui.StatusInfo)
		}
		return ui.Callout(ui.CalloutConfig{
			Title:   prop(props, "title"),
			Variant: variant,
			Class:   prop(props, "class"),
		}, body)
	}
}

func cardComponent(props map[string]string, body render.HTML) render.HTML {
	return ui.Card(ui.CardConfig{
		Heading:     firstProp(props, "title", "heading"),
		Description: prop(props, "description"),
		Href:        safeLinkProp(props, "href"),
		Variant:     cardVariant(props["variant"]),
		Class:       prop(props, "class"),
	}, body)
}

func detailsComponent(props map[string]string, body render.HTML) render.HTML {
	summary := firstProp(props, "summary", "title", "label")
	if summary == "" {
		summary = "Details"
	}
	return ui.Collapsible(ui.CollapsibleConfig{
		Summary: summary,
		Open:    boolProp(props, "open"),
		Class:   prop(props, "class"),
	}, body)
}

func badgeComponent(props map[string]string, _ render.HTML) render.HTML {
	label := firstProp(props, "label", "text")
	if label == "" {
		return ""
	}
	return ui.StatusBadge(ui.StatusBadgeConfig{
		Label:   label,
		Variant: statusVariant(props["variant"], ui.StatusNeutral),
		Class:   prop(props, "class"),
	})
}

func tagComponent(props map[string]string, _ render.HTML) render.HTML {
	label := firstProp(props, "label", "text")
	if label == "" {
		return ""
	}
	return ui.Tag(ui.TagConfig{
		Label:   label,
		Href:    safeLinkProp(props, "href"),
		Variant: statusVariant(props["variant"], ui.StatusNeutral),
		Class:   prop(props, "class"),
	})
}

func bannerComponent(props map[string]string, _ render.HTML) render.HTML {
	title := prop(props, "title")
	if title == "" {
		// Banner requires a title and panics without one.
		return ""
	}
	dismissID := firstProp(props, "dismiss-id", "dismiss")
	return ui.Banner(ui.BannerConfig{
		Title:       title,
		Body:        prop(props, "body"),
		Variant:     bannerVariant(props["variant"]),
		Dismissible: dismissID != "" || boolProp(props, "dismissible"),
		DismissID:   dismissID,
		Class:       prop(props, "class"),
	})
}

// diffComponent renders a unified diff. It takes the raw body: a patch loses
// its line structure the moment it is rendered as Markdown, and the surrounding
// code-block chrome ends up in the text.
func diffComponent(props map[string]string, raw string) render.HTML {
	patch := stripShortcodeFence(raw)
	if strings.TrimSpace(patch) == "" {
		return ""
	}
	mode := ui.DiffUnified
	if strings.EqualFold(prop(props, "mode"), "split") {
		mode = ui.DiffSplit
	}
	return ui.DiffViewer(ui.DiffViewerConfig{
		Patch:      patch,
		Mode:       mode,
		LeftLabel:  prop(props, "left"),
		RightLabel: prop(props, "right"),
		Class:      prop(props, "class"),
	})
}

func iconComponent(props map[string]string, _ render.HTML) render.HTML {
	name := prop(props, "name")
	if name == "" {
		return ""
	}
	// Icon returns an empty string for names it does not know, so an unknown
	// icon degrades to nothing rather than erroring.
	return ui.Icon(name, ui.IconConfig{
		Size:      prop(props, "size"),
		AriaLabel: firstProp(props, "label", "alt"),
		Class:     prop(props, "class"),
	})
}

// wrapperComponent styles a plain Markdown list. Steps and file trees are both
// authored as ordinary nested lists so the Markdown stays readable in a plain
// editor; the class carries the presentation.
func wrapperComponent(class string) MarkdownComponent {
	return func(props map[string]string, body render.HTML) render.HTML {
		classes := class
		if extra := prop(props, "class"); extra != "" {
			classes += " " + extra
		}
		return render.Tag("div", map[string]string{"class": classes}, body)
	}
}

// fileTreeComponent draws a directory listing from an indented block.
//
// It cannot wrap a Markdown list the way steps does, for two reasons: GoFastr's
// Markdown parser does not support nested lists and flattens them into one item
// joined by <br>, and rendering the body first would lose the indentation the
// nesting is read from. So it takes the raw body.
func fileTreeComponent(props map[string]string, raw string) render.HTML {
	classes := "fastr-docs-filetree"
	if extra := prop(props, "class"); extra != "" {
		classes += " " + extra
	}
	nodes := parseFileTree(stripShortcodeFence(raw))
	if len(nodes) == 0 {
		return ""
	}
	return render.Tag("div", map[string]string{"class": classes}, renderFileTree(nodes))
}

type fileTreeNode struct {
	name     string
	indent   int
	children []*fileTreeNode
}

func parseFileTree(text string) []*fileTreeNode {
	var roots []*fileTreeNode
	var stack []*fileTreeNode
	for _, line := range strings.Split(strings.ReplaceAll(text, "\r\n", "\n"), "\n") {
		expanded := strings.ReplaceAll(line, "\t", "  ")
		name := strings.TrimSpace(expanded)
		if name == "" {
			continue
		}
		indent := len(expanded) - len(strings.TrimLeft(expanded, " "))
		// Accept an optional list marker so a tree copied out of Markdown still
		// reads correctly.
		name = strings.TrimSpace(strings.TrimPrefix(strings.TrimPrefix(name, "- "), "* "))
		if name == "" {
			continue
		}
		node := &fileTreeNode{name: name, indent: indent}
		for len(stack) > 0 && stack[len(stack)-1].indent >= indent {
			stack = stack[:len(stack)-1]
		}
		if len(stack) == 0 {
			roots = append(roots, node)
		} else {
			parent := stack[len(stack)-1]
			parent.children = append(parent.children, node)
		}
		stack = append(stack, node)
	}
	return roots
}

func renderFileTree(nodes []*fileTreeNode) render.HTML {
	items := make([]render.HTML, 0, len(nodes))
	for _, node := range nodes {
		class := "fastr-docs-filetree__file"
		if strings.HasSuffix(node.name, "/") || len(node.children) > 0 {
			class = "fastr-docs-filetree__dir"
		}
		parts := []render.HTML{render.Tag("span", map[string]string{"class": class}, render.Text(node.name))}
		if len(node.children) > 0 {
			parts = append(parts, renderFileTree(node.children))
		}
		items = append(items, render.Tag("li", nil, parts...))
	}
	return render.Tag("ul", nil, items...)
}

// slotComponent renders as its own body. It exists so a child shortcode can
// carry props for its parent container and still be meaningful on its own.
func slotComponent(_ map[string]string, body render.HTML) render.HTML { return body }

func actionComponent(props map[string]string, _ render.HTML) render.HTML {
	label := firstProp(props, "label", "text")
	href := safeLinkProp(props, "href")
	if label == "" || href == "" {
		return ""
	}
	return ui.LinkButton(ui.LinkButtonConfig{
		Label:    label,
		Href:     href,
		Variant:  buttonVariant(props["variant"]),
		External: strings.HasPrefix(href, "http"),
	})
}

func tabsContainer(props map[string]string, children []MarkdownChild, body render.HTML) render.HTML {
	items := make([]tabs.Tab, 0, len(children))
	for _, child := range children {
		label := firstProp(child.Props, "label", "title")
		if label == "" {
			label = "Tab " + strconv.Itoa(len(items)+1)
		}
		items = append(items, tabs.Tab{Label: label, Content: child.Body, Open: len(items) == 0})
	}
	if len(items) == 0 {
		// tabs.New panics without at least one tab, so a tab set that was
		// written empty falls back to whatever prose it contained.
		return body
	}
	return tabs.New(tabs.Config{
		Name:  "fastr-docs-tabs-" + strconv.FormatUint(tabGroupSeq.Add(1), 10),
		Label: prop(props, "label"),
		Class: prop(props, "class"),
	}, items...)
}

func gridContainer(props map[string]string, children []MarkdownChild, body render.HTML) render.HTML {
	cells := make([]render.HTML, 0, len(children))
	for _, child := range children {
		cells = append(cells, child.Body)
	}
	if len(cells) == 0 {
		// No child shortcodes, so treat the body as the grid's content. This is
		// what lets {{< cards >}} wrap hand-written Markdown.
		cells = append(cells, body)
	}
	return ui.Grid(ui.GridConfig{
		Min:   prop(props, "min"),
		Gap:   gridGap(props["gap"]),
		Class: prop(props, "class"),
	}, cells...)
}

func heroContainer(props map[string]string, children []MarkdownChild, body render.HTML) render.HTML {
	actions := make([]render.HTML, 0, len(children))
	for _, child := range children {
		if child.Name == "action" && child.Body != "" {
			actions = append(actions, child.Body)
		}
	}
	hero := ui.Hero(ui.HeroConfig{
		Eyebrow:  prop(props, "eyebrow"),
		Title:    prop(props, "title"),
		Subtitle: firstProp(props, "subtitle", "tagline"),
		Actions:  actions,
		Class:    prop(props, "class"),
	})
	if strings.TrimSpace(string(body)) == "" {
		return hero
	}
	return render.Tag("div", map[string]string{"class": "fastr-docs-hero"}, hero, body)
}

func prop(props map[string]string, name string) string {
	return strings.TrimSpace(props[name])
}

func firstProp(props map[string]string, names ...string) string {
	for _, name := range names {
		if value := prop(props, name); value != "" {
			return value
		}
	}
	return ""
}

func boolProp(props map[string]string, name string) bool {
	switch strings.ToLower(prop(props, name)) {
	case "true", "yes", "1", "on":
		return true
	}
	// A valueless attribute such as {{< details open >}} parses to no entry at
	// all, so treat the bare presence of the key as true.
	_, present := props[name]
	return present && props[name] == ""
}

// safeLinkProp keeps javascript: and other script-bearing URLs out of hrefs
// that a content author supplied.
func safeLinkProp(props map[string]string, name string) string {
	return safeContentLink(prop(props, name))
}

func safeContentLink(value string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return ""
	}
	lower := strings.ToLower(value)
	if strings.HasPrefix(lower, "/") || strings.HasPrefix(lower, "#") || strings.HasPrefix(lower, "./") || strings.HasPrefix(lower, "../") {
		return value
	}
	if strings.HasPrefix(lower, "http://") || strings.HasPrefix(lower, "https://") || strings.HasPrefix(lower, "mailto:") {
		return value
	}
	return ""
}

func statusVariant(value string, fallback ui.StatusVariant) ui.StatusVariant {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "info":
		return ui.StatusInfo
	case "success", "tip":
		return ui.StatusSuccess
	case "warning", "warn", "caution":
		return ui.StatusWarning
	case "danger", "error":
		return ui.StatusDanger
	case "neutral", "note":
		return ui.StatusNeutral
	}
	return fallback
}

func bannerVariant(value string) ui.BannerVariant {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "success":
		return ui.BannerSuccess
	case "warning", "warn", "caution":
		return ui.BannerWarn
	case "danger", "error":
		return ui.BannerDanger
	}
	return ui.BannerInfo
}

func cardVariant(value string) ui.CardVariant {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "outlined":
		return ui.CardOutlined
	case "flat":
		return ui.CardFlat
	}
	return ui.CardElevated
}

func buttonVariant(value string) ui.ButtonVariant {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "secondary":
		return ui.ButtonSecondary
	case "ghost":
		return ui.ButtonGhost
	}
	return ui.ButtonPrimary
}

func gridGap(value string) ui.Gap {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "none":
		return ui.GapNone
	case "xs":
		return ui.GapXS
	case "sm", "small":
		return ui.GapSM
	case "lg", "large":
		return ui.GapLG
	case "xl":
		return ui.GapXL
	case "2xl":
		return ui.Gap2XL
	}
	return ui.GapMD
}

// renderPageHero draws the hero of a splash page from its front matter, so a
// landing page needs no typed screen.
func renderPageHero(hero *PageHero) render.HTML {
	if hero == nil {
		return ""
	}
	if hero.Title == "" && hero.Tagline == "" && hero.Eyebrow == "" && len(hero.Actions) == 0 {
		return ""
	}
	actions := make([]render.HTML, 0, len(hero.Actions))
	for _, action := range hero.Actions {
		link := safeContentLink(action.Link)
		if action.Text == "" || link == "" {
			continue
		}
		actions = append(actions, ui.LinkButton(ui.LinkButtonConfig{
			Label:    action.Text,
			Href:     link,
			Variant:  buttonVariant(action.Variant),
			External: strings.HasPrefix(link, "http"),
		}))
	}
	return ui.Hero(ui.HeroConfig{
		Eyebrow:  hero.Eyebrow,
		Title:    hero.Title,
		Subtitle: hero.Tagline,
		Actions:  actions,
		Class:    "fastr-docs-page-hero",
	})
}
