package docs

// Layout chrome: the header, sidebar, locale/version selectors, and command
// palette the Router renders around page content. Route registration, search,
// and mounting live in router.go.

import (
	"context"
	"encoding/json"
	"sort"
	"strings"

	uiapp "github.com/DonaldMurillo/gofastr/core-ui/app"
	"github.com/DonaldMurillo/gofastr/core-ui/widget"
	"github.com/DonaldMurillo/gofastr/core/render"
	"github.com/DonaldMurillo/gofastr/framework/ui"
)

type docsHeader struct{ router *Router }

func (h *docsHeader) Render() render.HTML {
	return h.render("")
}

func (h *docsHeader) RenderCtx(ctx context.Context) render.HTML {
	currentPath := ""
	if request := uiapp.RequestFromContext(ctx); request != nil && request.URL != nil {
		currentPath = request.URL.Path
	}
	return h.render(currentPath)
}

func (h *docsHeader) render(currentPath string) render.HTML {
	searchTrigger, _ := h.router.ensureCommandPalette()
	labels := h.router.UIStrings()
	drawerName := "fastr-docs-sections"
	if root := h.router.rootForPath(currentPath); root != nil && h.router.isBlogPrefix(root.Path) {
		drawerName = blogDrawerName(root.Path)
	}
	brand := render.Join(
		render.Tag("button", map[string]string{
			"class":                         "fastr-docs-mobile-nav-trigger",
			"type":                          "button",
			"data-fui-open":                 drawerName,
			"data-fastr-docs-global-drawer": "fastr-docs-sections",
			"data-fastr-docs-blog-drawer":   "fastr-docs-blog-sections",
			"data-fastr-docs-blog-prefixes": h.router.blogPrefixes(),
			"data-fastr-docs-blog-drawers":  h.router.blogDrawerNames(),
			"aria-label":                    labels.OpenNavigation,
		}, render.Raw(`<svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.8" stroke-linecap="round" aria-hidden="true"><line x1="4" y1="6" x2="20" y2="6"/><line x1="4" y1="12" x2="20" y2="12"/><line x1="4" y1="18" x2="20" y2="18"/></svg>`)),
		render.Tag("a", map[string]string{"href": "/", "class": "fastr-docs-brand", "aria-label": h.router.SiteName() + " home"},
			h.brandMark(),
			render.Tag("span", map[string]string{"class": "fastr-docs-brand__name"}, render.Text(h.router.SiteName())),
		),
	)
	actions := render.Join(
		render.Tag("div", map[string]string{"class": "fastr-docs-command-search"},
			searchTrigger,
		),
		h.router.variantSelectors(currentPath),
		ui.ThemeToggle(ui.ThemeToggleConfig{Variant: ui.ThemeToggleIcon, Class: "fastr-docs-theme-toggle"}),
	)
	header := ui.SiteHeader(ui.SiteHeaderConfig{
		Brand: brand,
		MobileBrand: render.Join(
			render.Tag("button", map[string]string{
				"class":                         "fastr-docs-mobile-nav-trigger",
				"type":                          "button",
				"data-fui-open":                 drawerName,
				"data-fastr-docs-global-drawer": "fastr-docs-sections",
				"data-fastr-docs-blog-drawer":   "fastr-docs-blog-sections",
				"data-fastr-docs-blog-prefixes": h.router.blogPrefixes(),
				"data-fastr-docs-blog-drawers":  h.router.blogDrawerNames(),
				"aria-label":                    labels.OpenNavigation,
			}, render.Raw(`<svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.8" stroke-linecap="round" aria-hidden="true"><line x1="4" y1="6" x2="20" y2="6"/><line x1="4" y1="12" x2="20" y2="12"/><line x1="4" y1="18" x2="20" y2="18"/></svg>`)),
			render.Tag("a", map[string]string{"href": "/", "class": "fastr-docs-brand fastr-docs-brand--mobile", "aria-label": h.router.SiteName() + " home"},
				h.brandMark(),
			),
		),
		NavItems:     h.router.headerItems(),
		Actions:      actions,
		Class:        "fastr-docs-site-header",
		Drawer:       ui.SiteHeaderDrawerPopover,
		NavUnderline: true,
	})
	return render.Join(header, h.router.headerVariantActiveConfig())
}

// headerVariantActiveScript extends GoFastr's normal exact/prefix active-link
// matching for route families whose published pages are siblings, such as
// /locales/pt-BR/v1/guide and /locales/fr/v2/guide. The header still links to
// the first published variant, but every sibling should keep that section
// active while the reader switches language or version.
func (r *Router) headerVariantActiveScript() render.HTML {
	type activeConfig struct {
		Links  map[string]string `json:"links"`
		Routes map[string]string `json:"routes"`
	}
	cfg := activeConfig{Links: map[string]string{}, Routes: map[string]string{}}
	for _, route := range r.sorted(r.roots) {
		if !r.routeVisible(route) || route.Path == "/" {
			continue
		}
		target := r.headerTarget(route)
		if target == nil || (target.Metadata.Locale == "" && target.Metadata.Version == "") {
			continue
		}
		family := variantFamily(target)
		if family == "" {
			continue
		}
		cfg.Links[target.Path] = family
		for _, candidate := range r.Routes() {
			if r.variantPublished(candidate) && variantFamily(candidate) == family {
				cfg.Routes[candidate.Path] = family
			}
		}
	}
	if len(cfg.Links) == 0 || len(cfg.Routes) == 0 {
		return render.Text("")
	}
	payload, err := json.Marshal(cfg)
	if err != nil {
		return render.Text("")
	}
	return render.Raw(`<script data-fastr-docs-active-families>(function(){
  'use strict';
  if (window.__fastrDocsActiveFamilies) return;
  window.__fastrDocsActiveFamilies = true;
  const config = ` + string(payload) + `;
  const normalize = (value) => {
    const path = (value || '/').split(/[?#]/, 1)[0] || '/';
    return path.length > 1 ? path.replace(/\/+$/, '') : path;
  };
  const update = (path) => {
    const family = config.routes[normalize(path)] || '';
    document.querySelectorAll('header nav a').forEach((link) => {
      const linkFamily = config.links[normalize(link.getAttribute('href'))];
      if (!linkFamily) return;
      link.setAttribute('data-fui-activelink-skip', '');
      if (family && linkFamily === family) {
        link.setAttribute('aria-current', 'page');
        link.classList.add('active');
      } else {
        link.removeAttribute('aria-current');
        link.classList.remove('active');
      }
    });
  };
  const updateCurrent = () => update(location.pathname + location.search);
  const root = document.body || document.documentElement;
  if (root && window.MutationObserver) {
    new MutationObserver(updateCurrent).observe(root, {childList: true, subtree: true});
  }
  window.addEventListener('gofastr:navigate', (event) => {
    update((event.detail && event.detail.path) || (location.pathname + location.search));
    setTimeout(updateCurrent, 0);
  });
  updateCurrent();
  setTimeout(updateCurrent, 0);
})();</script>`)
}

// headerVariantActiveConfig gives the external docs runtime the route-family
// data it needs to extend GoFastr's normal exact/prefix active-link matching
// for published siblings such as /locales/pt-BR/v1/guide and
// /locales/fr/v2/guide. The header still links to the first published variant,
// but every sibling should keep that section active while the reader switches
// language or version.
func (r *Router) headerVariantActiveConfig() render.HTML {
	links := map[string]string{}
	routes := map[string]string{}
	for _, route := range r.sorted(r.roots) {
		if !r.routeVisible(route) || route.Path == "/" {
			continue
		}
		target := r.headerTarget(route)
		if target == nil || (target.Metadata.Locale == "" && target.Metadata.Version == "") {
			continue
		}
		family := variantFamily(target)
		if family == "" {
			continue
		}
		links[target.Path] = family
		for _, candidate := range r.Routes() {
			if r.variantPublished(candidate) && variantFamily(candidate) == family {
				routes[candidate.Path] = family
			}
		}
	}
	if len(links) == 0 || len(routes) == 0 {
		return render.Text("")
	}
	payload, err := json.Marshal(map[string]map[string]string{"links": links, "routes": routes})
	if err != nil {
		return render.Text("")
	}
	return render.VoidTag("meta", map[string]string{
		"name":    "fastr-docs-active-families",
		"content": string(payload),
	})
}

type docsVariantOption struct {
	value string
	href  string
}

func (r *Router) variantSelectors(currentPath string) render.HTML {
	currentPath = normalizePath(currentPath)
	if currentPath == "" {
		return render.Text("")
	}
	current := r.routeAtPath(currentPath)
	if current == nil {
		return render.Text("")
	}
	var selectors []render.HTML
	if options := r.variantOptions(current, "locale"); len(options) > 1 {
		selectors = append(selectors, r.variantSelect(r.UIStrings().Language, "locale", options, currentPath))
	}
	if options := r.variantOptions(current, "version"); len(options) > 1 {
		selectors = append(selectors, r.variantSelect(r.UIStrings().Version, "version", options, currentPath))
	}
	if len(selectors) == 0 {
		return render.Text("")
	}
	return render.Tag("div", map[string]string{"class": "fastr-docs-variant-selectors"}, selectors...)
}

func (r *Router) variantSelect(label, dimension string, options []docsVariantOption, currentPath string) render.HTML {
	items := make([]render.HTML, 0, len(options))
	for _, option := range options {
		attrs := map[string]string{"value": option.href, "data-docs-variant-value": option.value}
		if normalizePath(option.href) == currentPath {
			attrs["selected"] = ""
		}
		items = append(items, render.Tag("option", attrs, render.Text(option.value)))
	}
	return render.Tag("label", map[string]string{"class": "fastr-docs-variant-select"},
		render.Tag("span", map[string]string{"class": "fastr-docs-variant-select__label"}, render.Text(label)),
		render.Tag("select", map[string]string{
			"aria-label":               label,
			"data-docs-variant-select": dimension,
			"data-docs-current-path":   currentPath,
		}, items...),
	)
}

func (r *Router) variantOptions(current *Route, dimension string) []docsVariantOption {
	values := make(map[string]bool)
	for _, route := range r.Routes() {
		if !r.variantPublished(route) || variantFamily(route) != variantFamily(current) {
			continue
		}
		value := route.Metadata.Locale
		if dimension == "version" {
			value = route.Metadata.Version
		}
		if strings.TrimSpace(value) != "" {
			values[value] = true
		}
	}
	if len(values) == 0 {
		return nil
	}
	ordered := make([]string, 0, len(values))
	for value := range values {
		ordered = append(ordered, value)
	}
	sort.Strings(ordered)
	options := make([]docsVariantOption, 0, len(ordered))
	for _, value := range ordered {
		if target := r.variantTarget(current, dimension, value); target != nil {
			options = append(options, docsVariantOption{value: value, href: target.Path})
		}
	}
	return options
}

func (r *Router) variantTarget(current *Route, dimension, value string) *Route {
	bestScore := -1
	var best *Route
	for _, candidate := range r.Routes() {
		if !r.variantPublished(candidate) || variantFamily(candidate) != variantFamily(current) {
			continue
		}
		candidateValue := candidate.Metadata.Locale
		if dimension == "version" {
			candidateValue = candidate.Metadata.Version
		}
		if candidateValue != value {
			continue
		}
		score := 0
		if candidate.Kind == current.Kind {
			score += 2
		}
		if candidate.Title == current.Title {
			score++
		}
		if candidate.Metadata.Locale == current.Metadata.Locale {
			score++
		}
		if candidate.Metadata.Version == current.Metadata.Version {
			score++
		}
		if score > bestScore {
			bestScore, best = score, candidate
		}
	}
	return best
}

func (r *Router) routeAtPath(path string) *Route {
	path = normalizePath(path)
	for _, route := range r.Routes() {
		if route.Path == path {
			return route
		}
	}
	return nil
}

func (r *Router) variantPublished(route *Route) bool {
	return route != nil && route.Kind != KindGroup && !route.Hidden && (!route.Metadata.Draft || r.includeDrafts || route.includeDrafts)
}

func variantFamily(route *Route) string {
	if route == nil {
		return ""
	}
	parts := strings.Split(strings.Trim(route.Path, "/"), "/")
	remove := map[string]bool{}
	if route.Metadata.Locale != "" {
		remove[route.Metadata.Locale] = true
	}
	if route.Metadata.Version != "" {
		remove[route.Metadata.Version] = true
	}
	filtered := parts[:0]
	for _, part := range parts {
		if !remove[part] {
			filtered = append(filtered, part)
		}
	}
	return strings.Join(filtered, "/")
}

type docsSidebar struct{ router *Router }

func (s *docsSidebar) Render() render.HTML {
	return s.render("")
}

func (s *docsSidebar) RenderCtx(ctx context.Context) render.HTML {
	currentPath := ""
	if request := uiapp.RequestFromContext(ctx); request != nil && request.URL != nil {
		currentPath = request.URL.Path
	}
	return s.render(currentPath)
}

func (s *docsSidebar) render(currentPath string) render.HTML {
	cfg := s.router.sidebarConfig(currentPath)
	cfg.CurrentPath = currentPath
	return ui.Sidebar(cfg).Render()
}

func (r *Router) headerItems() []ui.SiteHeaderLink {
	items := make([]ui.SiteHeaderLink, 0, len(r.roots))
	for _, route := range r.sorted(r.roots) {
		if !r.routeVisible(route) || route.Path == "/" {
			continue
		}
		target := r.headerTarget(route)
		if target == nil {
			continue
		}
		items = append(items, ui.SiteHeaderLink{Label: route.Title, Href: target.Path, MatchPrefix: true})
	}
	return items
}

func (r *Router) headerTarget(route *Route) *Route {
	if route == nil {
		return nil
	}
	if route.Kind == KindGroup {
		return r.firstVisibleDescendant(route)
	}
	return route
}

func (r *Router) firstVisibleDescendant(route *Route) *Route {
	if route == nil {
		return nil
	}
	for _, child := range r.sorted(route.Children) {
		if !r.routeVisible(child) {
			continue
		}
		if child.Kind != KindGroup {
			return child
		}
		if descendant := r.firstVisibleDescendant(child); descendant != nil {
			return descendant
		}
	}
	return nil
}

func (r *Router) ensureCommandPalette() (render.HTML, *widget.Builder) {
	if r.commandPalette != nil {
		return r.commandPaletteVisible, r.commandPalette
	}
	commands := make([]ui.PaletteCommand, 0, len(r.Routes()))
	for _, route := range r.PublishedRoutes() {
		if !r.routePublished(route) {
			continue
		}
		meta := route.Path
		if route.Description != "" {
			meta += " · " + route.Description
		}
		commands = append(commands, ui.PaletteCommand{Label: route.Title, Href: route.Path, Meta: meta})
	}
	if len(commands) == 0 {
		commands = append(commands, ui.PaletteCommand{Label: r.SiteName(), Href: "/", Meta: r.UIStrings().Home})
	}
	_, palette := ui.CommandPalette(ui.CommandPaletteConfig{
		Name:         "fastr-docs-command-palette",
		Placeholder:  r.UIStrings().SearchPlaceholder,
		TriggerLabel: r.UIStrings().OpenSearch,
		Commands:     commands,
	})
	// GoFastr's native palette supplies Escape and backdrop dismissal. Add a
	// visible close affordance as a later slot so the search input remains the
	// first focus target when the modal reopens. CSS positions this control at
	// the palette's top-right edge on both desktop and mobile.
	palette.Slot("footer", docsCommandPaletteClose{label: r.UIStrings().CloseSearch})
	visible := render.Tag("button", map[string]string{
		"type":                          "button",
		"class":                         "fastr-docs-command-trigger",
		"data-fui-open":                 "fastr-docs-command-palette",
		"data-fui-shortcut-click":       "Meta+K",
		"data-fastr-docs-backend":       string(r.SearchBackend()),
		"data-fastr-docs-pagefind-path": r.PagefindPath(),
		"data-fastr-docs-index-path":    r.SearchIndexPath(),
		"aria-label":                    r.UIStrings().OpenSearch,
	},
		render.Raw(`<svg width="15" height="15" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.8" stroke-linecap="round" aria-hidden="true"><circle cx="10.8" cy="10.8" r="6.3"/><path d="m16 16 4.6 4.6"/></svg>`),
		render.Tag("span", map[string]string{"class": "fastr-docs-command-trigger__label"}, render.Text(r.UIStrings().Search)),
		ui.ShortcutHint(ui.ShortcutHintConfig{Chord: "Mod+K", SROnlyLabel: r.UIStrings().OpenSearch, Class: "fastr-docs-command-trigger__hint"}),
	)
	r.commandPaletteVisible = visible
	r.commandPalette = palette
	return visible, palette
}

type docsCommandPaletteClose struct{ label string }

func (c docsCommandPaletteClose) Render() render.HTML {
	if c.label == "" {
		c.label = defaultUIStrings.CloseSearch
	}
	return render.Tag("div", map[string]string{"class": "fastr-docs-command-palette__close-slot"},
		render.Tag("button", map[string]string{
			"type":            "button",
			"class":           "fastr-docs-command-palette__close",
			"data-fui-action": "close",
			"aria-label":      c.label,
		}, ui.Icon("close", ui.IconConfig{Size: "18"})),
	)
}

func (r *Router) sidebarConfig(currentPath string) ui.SidebarConfig {
	return ui.SidebarConfig{
		Title:                 r.UIStrings().Contents,
		Items:                 r.sidebarItems(r.sidebarRoots(currentPath), currentPath),
		DrawerName:            "fastr-docs-sections",
		SuppressDrawerTrigger: true,
	}
}

// sidebarRoots keeps the persistent contents rail scoped to the active
// top-level section. The home route is the global index, so it exposes every
// top-level route group; section pages then narrow the rail to Home plus the
// active section's local tree.
func (r *Router) sidebarRoots(currentPath string) []*Route {
	if strings.TrimSpace(currentPath) == "" {
		return r.roots
	}
	active := r.rootForPath(currentPath)
	if active == nil {
		return r.roots
	}
	if active.Path == "/" {
		return r.roots
	}
	roots := make([]*Route, 0, 2)
	if home := r.routes["/"]; home != nil && r.routeVisible(home) {
		roots = append(roots, home)
	}
	return append(roots, active)
}

func (r *Router) rootForPath(currentPath string) *Route {
	path := normalizePath(currentPath)
	var active *Route
	for _, root := range r.roots {
		if !r.routeVisible(root) || !pathActive(root.Path, path) {
			continue
		}
		if active == nil || len(root.Path) > len(active.Path) {
			active = root
		}
	}
	return active
}

func (r *Router) sidebarItems(routes []*Route, currentPath string) []ui.SidebarItem {
	items := make([]ui.SidebarItem, 0, len(routes))
	for _, route := range r.sorted(routes) {
		if !r.routeVisible(route) {
			continue
		}
		children := r.sidebarItems(route.Children, currentPath)
		label := route.Title
		if route.Path == "/" {
			label = r.UIStrings().Home
		}
		item := ui.SidebarItem{
			Label:    label,
			Icon:     sidebarIcon(route.Kind, route.Badge, route.Path, len(children) > 0 && normalizePath(currentPath) == normalizePath(route.Path)),
			Children: children,
			Active:   currentPath != "" && pathActive(route.Path, normalizePath(currentPath)),
		}
		if len(children) == 0 && route.Kind != KindGroup {
			item.Href = route.Path
		}
		if item.Href == "" && len(item.Children) == 0 {
			continue
		}
		items = append(items, item)
	}
	return items
}

func sidebarIcon(kind RouteKind, badge NavBadge, navPath string, activeParent bool) render.HTML {
	path := `<rect x="5" y="4" width="14" height="16" rx="2"/><path d="M8 8h8M8 12h8M8 16h5"/>`
	if kind == KindGroup {
		path = `<path d="M4 7.5h6l1.5 2H20v8.5a2 2 0 0 1-2 2H6a2 2 0 0 1-2-2z"/><path d="M4 9h16"/>`
	}
	activeAttr := ""
	if activeParent {
		activeAttr = ` data-fastr-docs-active="true"`
	}
	markup := `<svg width="15" height="15" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.7" stroke-linecap="round" stroke-linejoin="round" aria-hidden="true" data-fastr-docs-nav-path="` + render.Escape(navPath) + `"` + activeAttr + `>` + path + `</svg>`
	if badge.Label != "" {
		markup += `<span class="fastr-docs-nav-badge fastr-docs-nav-badge--` + string(badge.Tone) + `" aria-hidden="true" data-badge-label="` + render.Escape(badge.Label) + `" title="` + render.Escape(badge.Label) + `"></span>`
	}
	return render.Raw(markup)
}
