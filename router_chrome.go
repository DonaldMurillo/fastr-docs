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
	searchTrigger := h.router.searchTrigger(currentPath)
	labels := h.router.uiAt(currentPath)
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
		NavItems:     h.router.headerItems(currentPath),
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
	// label is what the reader sees. For a locale it is the language's own
	// name, because "es" is a worse label than "Espanol" for the one person who
	// needs the selector most.
	label string
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
		selectors = append(selectors, r.variantSelect(r.uiAt(currentPath).Language, "locale", options, currentPath))
	}
	if options := r.variantOptions(current, "version"); len(options) > 1 {
		selectors = append(selectors, r.variantSelect(r.uiAt(currentPath).Version, "version", options, currentPath))
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
		text := option.label
		if text == "" {
			text = option.value
		}
		items = append(items, render.Tag("option", attrs, render.Text(text)))
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
		value := r.effectiveLocale(route)
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
			label := value
			if dimension == "locale" {
				label = r.LocaleName(value)
			}
			options = append(options, docsVariantOption{value: value, label: label, href: target.Path})
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
		candidateValue := r.effectiveLocale(candidate)
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

// headerItems builds the primary nav for the page being read.
//
// Sections are keyed by variantFamily so a translated section does not appear
// beside the original as a tab of its own. Without that, a site with one
// Spanish subtree grows an "Espanol" tab that reads as a topic rather than a
// language; the language selector is the affordance for that.
//
// The labels stay in the language the sections were registered in. Translating
// a route title is the project's call, not the framework's, and a partly
// translated site is the normal case.
func (r *Router) headerItems(currentPath string) []ui.SiteHeaderLink {
	locale := ""
	if current := r.routeAtPath(currentPath); current != nil {
		locale = r.effectiveLocale(current)
	}
	items := make([]ui.SiteHeaderLink, 0, len(r.roots))
	seen := make(map[string]bool)
	for _, route := range r.sorted(r.roots) {
		if !r.routeVisible(route) {
			continue
		}
		family := variantFamily(route)
		// The home route is not a tab, but its family still has to be claimed
		// here, or a translated home lands in the nav as one.
		if route.Path == "/" {
			seen[family] = true
			continue
		}
		if seen[family] {
			continue
		}
		chosen := r.headerVariant(route, locale)
		target := r.headerTarget(chosen)
		if target == nil {
			continue
		}
		seen[family] = true
		items = append(items, ui.SiteHeaderLink{Label: chosen.Title, Href: target.Path, MatchPrefix: true})
	}
	return items
}

// headerVariant swaps a section for its translation when the page being read is
// in another language, so the tabs are in the reader's language wherever one
// exists. A section with no translation stays as it is, which is the normal
// state of a partly translated site.
//
// It searches every route rather than only the roots. A translated section is
// usually not a root: /es/docs sits under /es, so a roots-only search finds
// nothing and the whole nav stays in the source language.
func (r *Router) headerVariant(route *Route, locale string) *Route {
	if locale == "" || r.effectiveLocale(route) == locale {
		return route
	}
	family := variantFamily(route)
	// The whole tree, not Routes(): that returns pages only, and a translated
	// section is usually a group. /es/examples is a group, so a Routes() search
	// found nothing and the tab stayed in the source language.
	if match := r.findVariant(r.roots, route, family, locale); match != nil {
		return match
	}
	return route
}

func (r *Router) findVariant(routes []*Route, exclude *Route, family, locale string) *Route {
	for _, candidate := range routes {
		if candidate == exclude || !r.routeVisible(candidate) {
			continue
		}
		if variantFamily(candidate) == family && r.effectiveLocale(candidate) == locale {
			return candidate
		}
		if match := r.findVariant(candidate.Children, exclude, family, locale); match != nil {
			return match
		}
	}
	return nil
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
	r.commandPaletteVisible = r.searchTrigger("")
	r.commandPalette = palette
	return r.commandPaletteVisible, palette
}

// searchTrigger builds the header's search button for one page.
//
// It is deliberately not memoized with the palette. The palette modal is
// mounted once for the whole site, so its placeholder is fixed, but the trigger
// is rendered into every page and can carry that page's language.
func (r *Router) searchTrigger(currentPath string) render.HTML {
	labels := r.uiAt(currentPath)
	return render.Tag("button", map[string]string{
		"type":                          "button",
		"class":                         "fastr-docs-command-trigger",
		"data-fui-open":                 "fastr-docs-command-palette",
		"data-fui-shortcut-click":       "Meta+K",
		"data-fastr-docs-backend":       string(r.SearchBackend()),
		"data-fastr-docs-pagefind-path": r.PagefindPath(),
		"data-fastr-docs-index-path":    r.SearchIndexPath(),
		"aria-label":                    labels.OpenSearch,
	},
		render.Raw(`<svg width="15" height="15" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.8" stroke-linecap="round" aria-hidden="true"><circle cx="10.8" cy="10.8" r="6.3"/><path d="m16 16 4.6 4.6"/></svg>`),
		render.Tag("span", map[string]string{"class": "fastr-docs-command-trigger__label"}, render.Text(labels.Search)),
		ui.ShortcutHint(ui.ShortcutHintConfig{Chord: "Mod+K", SROnlyLabel: labels.OpenSearch, Class: "fastr-docs-command-trigger__hint"}),
	)
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
		Title:                 r.uiAt(currentPath).Contents,
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
	// A locale home such as /es is the translation of "/", not a section of the
	// site. Left as the active root it wraps the whole translated tree in one
	// extra level that the default locale does not have, so the reader gets
	// "Espanol > Documentacion > ..." where an English reader gets
	// "Documentation > ...". Descend past it to the section actually being read.
	if active.Path != "/" && variantFamily(active) == "" {
		if section := r.sectionForPath(active.Children, currentPath); section != nil {
			active = section
		}
	}
	if active.Path == "/" {
		return r.roots
	}
	roots := make([]*Route, 0, 2)
	if home := r.localeHome(currentPath); home != nil && r.routeVisible(home) {
		roots = append(roots, home)
	}
	return append(roots, active)
}

// sectionForPath finds the child that contains the current path, so a locale
// home can hand over to the section beneath it.
func (r *Router) sectionForPath(routes []*Route, currentPath string) *Route {
	path := normalizePath(currentPath)
	var best *Route
	for _, route := range routes {
		if !r.routeVisible(route) || !pathActive(route.Path, path) {
			continue
		}
		if best == nil || len(route.Path) > len(best.Path) {
			best = route
		}
	}
	return best
}

// localeHome returns the home route for the language being read.
//
// Translating the label alone is worse than not translating it: the link still
// took the reader to the default-locale home, so "Inicio" quietly left the
// Spanish site.
func (r *Router) localeHome(currentPath string) *Route {
	home := r.routes["/"]
	current := r.routeAtPath(currentPath)
	if current == nil {
		return home
	}
	locale := r.effectiveLocale(current)
	if home != nil && r.effectiveLocale(home) == locale {
		return home
	}
	if match := r.findVariant(r.roots, nil, "", locale); match != nil {
		return match
	}
	return home
}

// homeFirst puts the home of the current locale at the top of the sidebar.
//
// Explicit Order decides the rest, but a translated home cannot compete on it:
// /es sits among the site's top-level sections and would sort wherever its
// number falls, landing "Inicio" under the section it introduces.
func homeFirst(routes []*Route) []*Route {
	for i, route := range routes {
		if variantFamily(route) != "" {
			continue
		}
		if i == 0 {
			return routes
		}
		ordered := make([]*Route, 0, len(routes))
		ordered = append(ordered, route)
		ordered = append(ordered, routes[:i]...)
		return append(ordered, routes[i+1:]...)
	}
	return routes
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
	for _, route := range homeFirst(r.sorted(routes)) {
		if !r.routeVisible(route) {
			continue
		}
		children := r.sidebarItems(route.Children, currentPath)
		label := route.Title
		// Every locale has a home, and each one is labelled "Home" in its own
		// language rather than by its route title.
		//
		// It is also a link, never a section. A locale home such as /es owns
		// the whole translated tree, so recursing into it would list the entire
		// site again underneath the word "Inicio".
		if variantFamily(route) == "" {
			label = r.uiAt(currentPath).Home
			children = nil
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
