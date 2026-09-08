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
	"github.com/DonaldMurillo/gofastr/core-ui/html"
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
	return render.Join(h.siteHeader(currentPath), h.router.headerVariantActiveConfig())
}

// siteHeader is the header as rendered for one page: its nav tabs, language
// and version selectors, and search trigger all follow that page's language.
func (h *docsHeader) siteHeader(currentPath string) render.HTML {
	searchTrigger := h.router.searchTrigger(currentPath)
	labels := h.router.uiAt(currentPath)
	// Both trigger buttons carry the same drawer map; it is computed once
	// because it walks the route tree per locale.
	localeDrawers := h.router.localeDrawerNames()
	drawerName := "fastr-docs-sections"
	// The collection, not the tree root: /es/blog sits under the Spanish
	// home, and its trigger must open the Spanish blog's drawer.
	if root := h.router.blogRootFor(h.router.routeAtPath(currentPath)); root != nil {
		drawerName = blogDrawerName(root.Path)
	}
	brand := render.Join(
		render.Tag("button", map[string]string{
			"class":                          "fastr-docs-mobile-nav-trigger",
			"type":                           "button",
			"data-fui-open":                  drawerName,
			"data-fastr-docs-global-drawer":  "fastr-docs-sections",
			"data-fastr-docs-blog-drawer":    "fastr-docs-blog-sections",
			"data-fastr-docs-blog-prefixes":  h.router.blogPrefixes(),
			"data-fastr-docs-blog-drawers":   h.router.blogDrawerNames(),
			"data-fastr-docs-locale-drawers": localeDrawers,
			"aria-label":                     labels.OpenNavigation,
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
				"class":                          "fastr-docs-mobile-nav-trigger",
				"type":                           "button",
				"data-fui-open":                  drawerName,
				"data-fastr-docs-global-drawer":  "fastr-docs-sections",
				"data-fastr-docs-blog-drawer":    "fastr-docs-blog-sections",
				"data-fastr-docs-blog-prefixes":  h.router.blogPrefixes(),
				"data-fastr-docs-blog-drawers":   h.router.blogDrawerNames(),
				"data-fastr-docs-locale-drawers": localeDrawers,
				"aria-label":                     labels.OpenNavigation,
				// The trigger names the widget it opens, so assistive tech
				// pairs them without guessing from data attributes.
				"aria-controls": drawerName,
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
	return header
}

// chromeTemplate is the page's own header, carried inside the region GoFastr
// swaps on client-side navigation.
//
// The header lives in the outermost layout layer, which the runtime keeps
// across navigations, so after a client-side move from a Spanish page to an
// English one the header stayed Spanish: nav tabs, search label, the
// language selector still aimed at the previous page's translation, and
// <html lang>. The runtime reads this template on every navigation and
// copies the parts that change into the live header (syncChrome in
// runtime.go). On a full page load it is a no-op. The path comes from the
// request when there is one, because a dynamic screen's route path is a
// pattern.
func (r *Router) chromeTemplate(ctx context.Context, fallbackPath string) render.HTML {
	currentPath := fallbackPath
	if request := uiapp.RequestFromContext(ctx); request != nil && request.URL != nil {
		currentPath = request.URL.Path
	}
	return render.Tag("template", map[string]string{
		"data-fastr-docs-chrome": "",
		"data-fastr-docs-lang":   r.LanguageFor(currentPath),
		"data-fastr-docs-dir":    r.DirectionFor(currentPath),
		"data-fastr-docs-skip":   r.uiAt(currentPath).SkipToContent,
	}, render.Join(
		// syncChrome copies this into the live head, so the browser
		// chrome follows the page's theme instead of the first page's.
		render.Raw(`<meta name="theme-color" content="`+r.ThemeColor()+`">`),
		(&docsHeader{router: r}).siteHeader(currentPath)))
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
		family := r.familyOf(target)
		if family == "" {
			continue
		}
		cfg.Links[target.Path] = family
		for _, candidate := range r.Routes() {
			if r.variantPublished(candidate) && r.familyOf(candidate) == family {
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
		family := r.familyOf(target)
		if family == "" {
			continue
		}
		links[target.Path] = family
		for _, candidate := range r.Routes() {
			if r.variantPublished(candidate) && r.familyOf(candidate) == family {
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
	// The container is rendered even when empty, so the runtime can swap a
	// page's selectors into it after a client-side navigation whichever page
	// came before; CSS hides it while it has nothing.
	return render.Tag("div", map[string]string{"class": "fastr-docs-variant-selectors"}, selectors...)
}

func (r *Router) variantSelect(label, dimension string, options []docsVariantOption, currentPath string) render.HTML {
	items := make([]render.HTML, 0, len(options))
	for _, option := range options {
		text := option.label
		if text == "" {
			text = option.value
		}
		// A phone-width header has room for a code, not a language name. The
		// option carries both spellings and the runtime shows the one that
		// fits (syncVariantShortLabels); a version is already short.
		short := option.value
		if dimension == "locale" {
			short = strings.ToUpper(option.value)
		}
		attrs := map[string]string{
			"value":                   option.href,
			"data-docs-variant-value": option.value,
			"data-docs-variant-label": text,
			"data-docs-variant-short": short,
		}
		// A language option says which language it is, so screen readers
		// announce "Spanish" rather than reading the label in the page's
		// own language.
		if dimension == "locale" {
			attrs["lang"] = option.value
		}
		if normalizePath(option.href) == currentPath {
			attrs["selected"] = ""
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
		if !r.variantPublished(route) || r.familyOf(route) != r.familyOf(current) {
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
		if !r.variantPublished(candidate) || r.familyOf(candidate) != r.familyOf(current) {
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
	path = normalizePath(strings.SplitN(strings.SplitN(path, "?", 2)[0], "#", 2)[0])
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

// variantFamily is the path-shaped family: the route's path with its locale
// and version segments removed. Callers want familyOf, which also honours
// TranslationOf; this is the fallback it uses.
func variantFamily(route *Route) string {
	if route == nil {
		return ""
	}
	parts := strings.Split(strings.Trim(route.Path, "/"), "/")
	// Segments compare case-insensitively: locales are normalized to
	// lowercase at registration, while the path itself keeps the spelling
	// the project chose (pt-BR stays pt-BR in the URL).
	remove := map[string]bool{}
	if route.Metadata.Locale != "" {
		remove[strings.ToLower(route.Metadata.Locale)] = true
	}
	if route.Metadata.Version != "" {
		remove[strings.ToLower(route.Metadata.Version)] = true
	}
	filtered := parts[:0]
	for _, part := range parts {
		if !remove[strings.ToLower(part)] {
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
		family := r.familyOf(route)
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
	family := r.familyOf(route)
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
		if r.familyOf(candidate) == family && r.effectiveLocale(candidate) == locale {
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
// searchLocale is the locale search results should be limited to, and is empty
// on a site that declares no locales at all.
func (r *Router) searchLocale(currentPath string) string {
	if r == nil || len(r.localeFamilyLocales()) == 0 {
		return ""
	}
	route := r.routeAtPath(currentPath)
	if route == nil {
		return ""
	}
	return r.effectiveLocale(route)
}

func (r *Router) searchTrigger(currentPath string) render.HTML {
	labels := r.uiAt(currentPath)
	return render.Tag("button", map[string]string{
		"type":                          "button",
		"class":                         "fastr-docs-command-trigger",
		"aria-expanded":                 "false",
		"aria-haspopup":                 "dialog",
		"data-fui-open":                 "fastr-docs-command-palette",
		"data-fui-shortcut-click":       "Meta+K",
		"data-fastr-docs-backend":       string(r.SearchBackend()),
		"data-fastr-docs-pagefind-path": r.PagefindPath(),
		"data-fastr-docs-index-path":    r.SearchIndexPath(),
		// The path prefix the site is served under, empty on the live host.
		// A static export below a prefix stamps it in (RewriteStaticBase),
		// and the runtime strips it when matching a route and adds it when
		// navigating to one: route paths in the page are root-relative,
		// location.pathname is not.
		"data-fastr-docs-base": "",
		// The runtime filters results to this locale. The trigger carries it
		// so the palette does not depend on what the document declares.
		"data-fastr-docs-locale": r.searchLocale(currentPath),
		// The palette modal is mounted once for the whole site, so it cannot be
		// rendered per language. The runtime applies these to it for the page
		// being read.
		"data-fastr-docs-search-placeholder": labels.SearchPlaceholder,
		"data-fastr-docs-search-close":       labels.CloseSearch,
		"aria-label":                         labels.OpenSearch,
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

// docsDrawerName is the widget name for one locale's drawer. The default
// locale keeps the short name (public selectors and the service worker
// precache depend on it); every other locale derives one, the way
// blogDrawerName does for a collection.
func docsDrawerName(locale string) string {
	value := strings.ToLower(strings.TrimSpace(locale))
	if value == "" {
		return "fastr-docs-sections"
	}
	value = strings.NewReplacer("/", "-", "_", "-", ".", "-").Replace(value)
	return "fastr-docs-sections-" + value
}

// localeDrawerHomes lists the home routes behind the mounted drawers: the
// default drawer's home first (nil when the site registers no route at "/"),
// then one home per additional locale, sorted. The first entry owns the
// pinned fastr-docs-sections name whatever locales the project declares, so
// callers decide the drawer name by position, not by comparing locale
// spellings. A drawer is mounted per home because a widget's body is fixed
// at mount time: one drawer for the whole site would hand a translated
// reader the default locale's tree with a home link leading out of their
// language. A locale with no home route has no drawer; its pages fall back
// to the default locale's, which is the state before per-locale drawers
// existed.
func (r *Router) localeDrawerHomes() []*Route {
	if r == nil {
		return nil
	}
	defaultHome := r.routes["/"]
	if defaultHome != nil && defaultHome.Hidden {
		// A hidden home cannot label the default drawer; mount it from the
		// bare roots list instead, the way a home-less site does.
		defaultHome = nil
	}
	seen := make(map[string]bool, 4)
	locales := make([]string, 0, 4)
	for _, family := range r.localeFamilyLocales() {
		for locale := range family {
			locale = normalizeLocale(locale)
			if locale == "" || locale == normalizeLocale(r.fallbackLocale) || seen[locale] {
				continue
			}
			seen[locale] = true
			locales = append(locales, locale)
		}
	}
	sort.Strings(locales)
	homes := []*Route{defaultHome}
	drawers := map[string]bool{docsDrawerName(""): true}
	seenVersionFamily := map[string]bool{}
	for _, home := range r.versionHomes() {
		if home == nil || home == defaultHome {
			continue
		}
		family := r.familyOf(home)
		if family != "" && seenVersionFamily[family] {
			continue
		}
		// The current version's tree is already the default drawer; a
		// second drawer for it would duplicate every link.
		if r.isCurrentVersionHome(home) {
			continue
		}
		if drawer := docsDrawerName("v-" + home.Metadata.Version); !drawers[drawer] {
			drawers[drawer] = true
			if family != "" {
				seenVersionFamily[family] = true
			}
			homes = append(homes, home)
		}
	}
	for _, locale := range locales {
		home := r.findVariant(r.roots, nil, "", locale)
		if home == nil || home == defaultHome {
			continue
		}
		// Two spellings of one locale can fold to the same drawer name
		// ("pt-BR" and "pt_BR"); the second mount would silently replace
		// the first, so the first sorted spelling wins.
		drawer := docsDrawerName(locale)
		if drawers[drawer] {
			continue
		}
		drawers[drawer] = true
		homes = append(homes, home)
	}
	return homes
}

// localeDrawerNames maps each non-default locale's home path to its drawer
// name, as semicolon-separated prefix=drawer pairs. The runtime re-aims the
// trigger per page with it, the same shape as the blog drawer map.
func (r *Router) localeDrawerNames() string {
	homes := r.localeDrawerHomes()
	if len(homes) < 2 {
		return ""
	}
	pairs := make([]string, 0, len(homes)-1)
	for _, home := range homes[1:] {
		pairs = append(pairs, home.Path+"="+docsDrawerName(r.effectiveLocale(home)))
	}
	return strings.Join(pairs, ";")
}

// versionHomes lists one home per version family: the route that opens the
// version's tree, so a versioned site gets a drawer per version exactly the
// way a translated one gets one per language.
func (r *Router) versionHomes() []*Route {
	if r == nil {
		return nil
	}
	// One home per version FAMILY, not per version: /guide (v2) and
	// /v1/guide pair as variants of one guide, and the drawer count follows
	// the families the reader can cross between.
	seen := map[string]bool{}
	var homes []*Route
	for _, route := range r.Routes() {
		version := strings.TrimSpace(route.Metadata.Version)
		if version == "" || seen[version] {
			continue
		}
		if !r.routeVisible(route) {
			continue
		}
		seen[version] = true
		homes = append(homes, route)
	}
	return homes
}

// drawerRoots is the whole navigation tree for one locale's drawer. The
// desktop rail narrows to the section being read (sidebarRoots); the drawer
// is the only navigation a phone has, so it carries the locale's home plus
// every top-level section of that locale.
func (r *Router) drawerRoots(home *Route) []*Route {
	if home == nil || home.Path == "/" || len(home.Children) == 0 {
		return r.roots
	}
	return append([]*Route{home}, home.Children...)
}

// docsDrawerConfig builds the sidebar config for one locale's drawer. The
// inline rail keeps sidebarConfig, which narrows to the active section; this
// is the drawer's whole-language twin.
func (r *Router) docsDrawerConfig(home *Route, drawer string) ui.SidebarConfig {
	currentPath := ""
	if home != nil && home.Path != "/" {
		currentPath = home.Path
	}
	// On a single-language site whose only locale is not the default
	// ("everything is Spanish", no fallback declared), the default drawer
	// must still speak that language: read labels from the home's locale.
	title := r.uiAt(currentPath).Contents
	if currentPath == "" && home != nil && home.Path == "/" {
		title = r.uiAt(home.Path).Contents
	}
	return ui.SidebarConfig{
		Title:                 title,
		Items:                 r.sidebarItems(r.drawerRoots(home), currentPath),
		DrawerName:            drawer,
		SuppressDrawerTrigger: true,
	}
}

// docsSectionSelect renders the dropdown at the top of a drawer: the home of
// the drawer's language plus every section, aimed and labelled the way the
// header tabs are. The tabs are hidden below md and the drawer is the only
// navigation left, so this is what lets a reader jump between sections
// without scrolling the whole tree to find the next collapsed group.
func (r *Router) docsSectionSelect(currentPath, id string) render.HTML {
	labels := r.uiAt(currentPath)
	options := make([]ui.SelectOption, 0, 8)
	items := r.headerItems(currentPath)
	if home := r.localeHome(currentPath); home != nil {
		options = append(options, ui.SelectOption{Value: home.Path, Text: labels.Home})
		for _, item := range items {
			options = append(options, ui.SelectOption{Value: item.Href, Text: item.Label})
		}
	} else if len(items) > 0 {
		// A site with no home route still needs a first option meaning
		// "the beginning": the first section stands in, labeled Home.
		options = append(options, ui.SelectOption{Value: items[0].Href, Text: labels.Home})
		for _, item := range items[1:] {
			options = append(options, ui.SelectOption{Value: item.Href, Text: item.Label})
		}
	}
	// A versioned family puts one sibling per version in the select, so the
	// reader crosses versions from the drawer the way the selector crosses
	// languages.
	for _, home := range r.versionHomes() {
		if home == nil || home.Path == "" {
			continue
		}
		duplicate := false
		for _, option := range options {
			if option.Value == home.Path {
				duplicate = true
				break
			}
		}
		if !duplicate {
			options = append(options, ui.SelectOption{Value: home.Path, Text: home.Title})
		}
	}
	// A select with nothing to choose between is noise; a one-section
	// site keeps its drawer to the tree alone.
	if len(options) < 2 {
		return ""
	}
	versions := r.Versions()
	return ui.Select(ui.SelectConfig{
		Name:    "docs-section",
		ID:      id,
		Label:   labels.Sections,
		Help:    labels.SectionHelp,
		Options: options,
		Class:   "fastr-docs-section-select",
		ExtraAttrs: html.Attrs{
			"data-fastr-docs-section-select": "true",
			"autocomplete":                   "off",
			// Each option's route path rides along, so the runtime can
			// prefix-match and re-sync without parsing option text.
			"data-fastr-docs-section-paths": strings.Join(sectionPaths(options), ","),
			// The site's version segments, so the runtime can match an
			// option to the reader's current version even when the
			// option's own path is the default-family spelling.
			"data-fastr-docs-versions": strings.Join(versions, ","),
		},
	})
}

func sectionPaths(options []ui.SelectOption) []string {
	paths := make([]string, 0, len(options))
	for _, option := range options {
		if option.Value != "" {
			paths = append(paths, option.Value)
		}
	}
	return paths
}

// navigationDrawerBody is a drawer widget's content: the section select
// above the sidebar tree. ui.MountSidebar offers no way to put anything
// above the nav inside the drawer body (gofastr #405), so MountNavigation
// mounts the drawer from preset.Drawer directly and slots this in.
type navigationDrawerBody struct {
	cfg      ui.SidebarConfig
	sections render.HTML
}

func (b navigationDrawerBody) Render() render.HTML {
	return render.Join(
		render.Tag("div", map[string]string{"class": "fastr-docs-drawer-sections"}, b.sections),
		// The drawer-body class is the styling contract gofastr's own
		// drawer slot uses; styles.go keys a block of drawer styling on
		// it. ui.SidebarBody supplies the nav and the component style
		// marker, and its own wrapper class is unstyled.
		render.Tag("div", map[string]string{"class": "ui-sidebar ui-sidebar--drawer-body"}, ui.SidebarBody(b.cfg)),
	)
}

// docsDrawerBody renders one locale's drawer content: the section select and
// the whole tree for that language.
func (r *Router) docsDrawerBody(home *Route, drawer string) render.HTML {
	currentPath := ""
	if home != nil && home.Path != "/" {
		currentPath = home.Path
	}
	return navigationDrawerBody{cfg: r.docsDrawerConfig(home, drawer), sections: r.docsSectionSelect(currentPath, drawer+"-section")}.Render()
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
	if active.Path != "/" && r.familyOf(active) == "" {
		if section := r.sectionForPath(active.Children, currentPath); section != nil {
			active = section
		} else {
			// Standing on the locale home itself, which is the translation of
			// "/" and so lists that language's sections, exactly as the default
			// locale's home lists r.roots. Without this the home is both the
			// home entry and the active section, and renders twice.
			return append([]*Route{active}, active.Children...)
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
func (r *Router) homeFirst(routes []*Route) []*Route {
	for i, route := range routes {
		if r.familyOf(route) != "" {
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
	for _, route := range r.homeFirst(r.sorted(routes)) {
		if !r.routeVisible(route) {
			continue
		}
		children := r.sidebarItems(route.Children, currentPath)
		label := route.Title
		if r.familyOf(route) == "" {
			// Another language's home is not a section of this one. Left in, it
			// appears as a second "Home" leading out of the site, dragging that
			// language's whole tree behind it; the language selector is the
			// affordance for switching.
			if route != r.localeHome(currentPath) {
				continue
			}
			// The home of the language being read is labelled "Home" in that
			// language, and is a link rather than a section: a locale home owns
			// the whole translated tree, so recursing into it would list the
			// entire site again underneath the word "Inicio".
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

// familyHasCurrentVariant reports whether the family contains a route with
// no version segment: the tree the default drawer already serves.
func (r *Router) familyHasCurrentVariant(family string) bool {
	if family == "" {
		return false
	}
	for _, route := range r.Routes() {
		if r.familyOf(route) == family && strings.TrimSpace(route.Metadata.Version) == "" && r.variantPublished(route) {
			return true
		}
	}
	return false
}

// isCurrentVersionHome reports whether home belongs to a versioned
// collection whose current version is mounted at the plain prefix, which
// the default drawer already serves.
func (r *Router) isCurrentVersionHome(home *Route) bool {
	version := strings.TrimSpace(home.Metadata.Version)
	if version == "" {
		return false
	}
	for prefix, current := range r.currentVersions {
		if current == version && pathActive(prefix, home.Path) {
			return true
		}
	}
	return false
}
