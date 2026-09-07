// Package docs is the route and content layer for reusable documentation
// sites. It deliberately keeps navigation, search metadata, and page
// registration in one Router so an extension contributes to the same tree as
// ordinary documentation pages.
package docs

import (
	"context"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	stdhtml "html"
	"io"
	"net/url"
	"os"
	pathpkg "path"
	"reflect"
	"sort"
	"strings"

	uiapp "github.com/DonaldMurillo/gofastr/core-ui/app"
	"github.com/DonaldMurillo/gofastr/core-ui/component"
	"github.com/DonaldMurillo/gofastr/core-ui/widget"
	"github.com/DonaldMurillo/gofastr/core-ui/widget/preset"
	"github.com/DonaldMurillo/gofastr/core/render"
	gofastrRouter "github.com/DonaldMurillo/gofastr/core/router"
	"github.com/DonaldMurillo/gofastr/framework/ui"
)

// RouteKind identifies how a route is rendered.
type RouteKind string

const (
	KindGroup  RouteKind = "group"
	KindPage   RouteKind = "page"
	KindScreen RouteKind = "screen"
	KindPlugin RouteKind = "plugin"
)

// NavBadgeTone controls the visual emphasis of a route badge. The label is
// project-owned, so a badge can say "New", "Popular", "Hot", "Beta", or any
// other short piece of navigation metadata without coupling the Router to a
// particular product vocabulary.
type NavBadgeTone string

const (
	NavBadgeToneNeutral NavBadgeTone = "neutral"
	NavBadgeToneAccent  NavBadgeTone = "accent"
	NavBadgeToneInfo    NavBadgeTone = "info"
	NavBadgeToneSuccess NavBadgeTone = "success"
	NavBadgeToneWarning NavBadgeTone = "warning"
)

// NavBadge adds concise, optional metadata to a route in the sidebar.
// Badges are presentation metadata only: they do not affect URLs, search,
// ordering, active state, or route visibility.
type NavBadge struct {
	// Label is the badge text; it also titles the badge for
	// hover discovery.
	Label string
	// Tone selects the badge color.
	Tone NavBadgeTone
}

func normalizeNavBadge(badge NavBadge) (NavBadge, error) {
	badge.Label = strings.TrimSpace(badge.Label)
	if badge.Label == "" {
		return NavBadge{}, nil
	}
	if badge.Tone == "" {
		badge.Tone = NavBadgeToneAccent
	}
	switch badge.Tone {
	case NavBadgeToneNeutral, NavBadgeToneAccent, NavBadgeToneInfo, NavBadgeToneSuccess, NavBadgeToneWarning:
		return badge, nil
	default:
		return NavBadge{}, fmt.Errorf("invalid navigation badge tone %q", badge.Tone)
	}
}

// PageConfig describes a Markdown or server-rendered documentation page.
// Source is Markdown when Body is nil. Body is useful when a page needs a
// custom GoFastr component-like render function but is still a docs page.
type PageConfig struct {
	// Title is the page heading and nav label.
	Title string
	// Description is the page's meta description and search entry.
	Description string
	// Source is the Markdown body as a string, for pages authored in Go.
	Source string
	// SourcePath reads the Markdown body from a file; relative to the
	// project, resolved by the caller.
	SourcePath string
	// Body supplies the rendered HTML directly, bypassing Markdown
	// entirely.
	Body func() render.HTML
	// ContextBody is the request-aware form of Body. It is useful for
	// server-rendered aggregate views such as a blog search page. During
	// static export it receives a background context and should render the
	// canonical, query-free view.
	ContextBody func(context.Context) render.HTML
	// SearchText overrides the text fed to the search index; empty uses
	// the rendered content.
	SearchText string
	// Preload asks GoFastr to prefetch this route on hover, visibility, or
	// idle. Leave empty to keep the route opt-in and never prefetched.
	Preload string
	// DisableTOC drops the heading rail and its mobile select.
	DisableTOC bool
	// Order sorts the page among its siblings; lower comes first.
	Order int
	// Offline marks the page precacheable by the service worker.
	Offline bool
	// Hidden keeps the page out of navigation while it still serves.
	Hidden bool
	// Tags label the page for blog tag views.
	Tags []string
	// Metadata is the page's front-matter equivalent: locale, version,
	// translation pairing, redirects, and the rest.
	Metadata ContentMetadata
	// Badge pins a small badge beside the page in sidebars and drawers.
	Badge NavBadge
	// Components exposes typed GoFastr renderers to Markdown through the
	// {{< name key="value" >}} shortcode syntax. Components are rendered
	// server-side and may contain nested Markdown content.
	Components map[string]MarkdownComponent
	// Containers are shortcodes that receive their nested shortcodes as
	// separate children rather than one merged body, which is what tabs,
	// card grids, and steps need.
	Containers map[string]MarkdownContainer
}

// ScreenConfig registers an arbitrary typed GoFastr component in the same
// documentation tree as Markdown pages.
type ScreenConfig struct {
	// Title is the screen heading and nav label.
	Title string
	// Description is the screen's meta description.
	Description string
	// Component renders the screen's body; it owns everything inside the
	// content column.
	Component component.Component
	// SearchText is what the search index stores for the screen; screens
	// have no Markdown to derive it from.
	SearchText string
	// Preload asks GoFastr to prefetch this route on hover, visibility, or
	// idle. Leave empty to keep the route opt-in and never prefetched.
	Preload string
	// Plugin names the plugin that contributed the screen, surfaced for
	// provenance.
	Plugin string
	// Order sorts the screen among its siblings; lower comes first.
	Order int
	// Offline marks the screen precacheable by the service worker.
	Offline bool
	// Hidden keeps the screen out of navigation while it still serves.
	Hidden bool
	// Tags label the screen for blog tag views.
	Tags []string
	// Metadata carries locale and version for the screen, the same fields
	// front matter would give a page; a translated plugin mount sets it so its
	// section pairs with the original.
	Metadata ContentMetadata
	// Badge pins a small badge beside the screen in sidebars and drawers.
	Badge NavBadge
}

// GroupConfig describes a navigation group. Groups are metadata-only routes;
// their children are mounted as the actual GoFastr screens.
type GroupConfig struct {
	// Title is the group's label in navigation and search.
	Title string
	// Description is the group's meta description.
	Description string
	// Order sorts the group among its siblings; lower comes first.
	Order int
	// Hidden keeps the group out of navigation without unregistering its
	// pages.
	Hidden bool
	// Badge pins a small badge beside the group in sidebars and drawers.
	Badge NavBadge
	// Locale and Version place the group in a translated or versioned slice of
	// the site.
	//
	// A page takes these from its front matter, but a group is pure Go and has
	// no front matter to read, so without them a translated section could never
	// be paired with its original: variantFamily had no locale segment to
	// strip, and the language selector and header nav both went looking for a
	// counterpart that did not exist.
	Locale  string
	Version string
	// TranslationOf names the group this one translates, for a translated
	// section whose path does not mirror the original's. See
	// ContentMetadata.TranslationOf.
	TranslationOf string
}

// Route is a node in the docs route tree. Children are always returned in
// explicit order, with registration order as the stable tie-breaker.
type Route struct {
	// ID is the stable identifier derived from the path.
	ID string
	// Path is the route path, root-relative.
	Path string
	// Title is the nav label and page heading.
	Title string
	// Description is the meta description.
	Description string
	// Kind separates pages, screens, and metadata-only groups.
	Kind RouteKind
	// Order sorts the route among its siblings.
	Order int
	// Hidden keeps the route out of navigation while it still serves.
	Hidden bool
	// Offline marks the route precacheable by the service worker.
	Offline bool
	// Tags label the route for tag views.
	Tags []string
	// Plugin names the plugin that contributed the route, empty for
	// project-owned routes.
	Plugin string
	// SearchText is what the search index stores; empty derives it
	// from the rendered content.
	SearchText string
	// Preload asks GoFastr to prefetch the route on hover or idle.
	Preload string
	// Metadata carries locale, version, pairing, and redirects.
	Metadata ContentMetadata
	// Badge pins a small badge beside the route in sidebars and
	// drawers.
	Badge NavBadge
	// Blog marks a route as a post or archive entry created by MarkdownBlog.
	// BlogIndex distinguishes the archive route from individual posts.
	Blog bool
	// BlogIndex marks a publication collection's generated landing.
	BlogIndex bool
	// Parent links up the tree; nil at the root.
	Parent *Route
	// Children are the nested routes, in explicit order with
	// registration order as tie-breaker.
	Children []*Route

	// includeDrafts is set by a content collection that explicitly opts into
	// draft previews. It keeps that choice local to the collection instead of
	// changing visibility for every route registered on the Router.
	includeDrafts bool
	page          *PageConfig
	screen        *ScreenConfig
	seq           int
}

// NavItem is a flattened, render-agnostic navigation item. Consumers can use
// it to render a sidebar, breadcrumbs, mobile navigation, or an API index.
type NavItem struct {
	// ID mirrors Route.ID for the flattened view.
	ID string
	// Path is the route path, root-relative.
	Path string
	// Title is the nav label.
	Title string
	// Kind separates pages, screens, and groups.
	Kind RouteKind
	// Depth is the nesting level, zero at the top.
	Depth int
	// Active says this item is the current page.
	Active bool
	// Children are the nested items.
	Children int
	// Badge pins a small badge beside the item.
	Badge NavBadge
}

// SearchEntry is the portable local-search record produced by the Router.
// A static indexer such as Pagefind can consume these records without knowing
// anything about GoFastr.
type SearchEntry struct {
	// ID is the stable record identifier.
	ID string `json:"id"`
	// Path is the page the entry points at.
	Path string `json:"path"`
	// Title heads the result.
	Title string `json:"title"`
	// Description is the result's summary line.
	Description string `json:"description"`
	// Text is the searchable body the index matches against.
	Text string `json:"text"`
	// Tags join the searchable text.
	Tags []string `json:"tags,omitempty"`
	// Locale is the page's language; the JSON backend filters
	// results to the reader's language with it.
	Locale string `json:"locale,omitempty"`
	// Version is the page's version variant.
	Version string `json:"version,omitempty"`
	// EditURL links the result to its source.
	EditURL string `json:"editUrl,omitempty"`
	// Canonical is the page's canonical URL.
	Canonical string `json:"canonical,omitempty"`
	// NoIndex says the page asked search engines to skip it.
	NoIndex bool `json:"noIndex,omitempty"`
	// Kind separates pages, screens, and groups.
	Kind RouteKind `json:"kind,omitempty"`
	// Order breaks score ties the way the tree does.
	Order int `json:"order,omitempty"`
	// Headings carry the page's table of contents into the entry.
	Headings []string `json:"headings,omitempty"`
	// Alternates maps language tags to the entry's translations, the
	// hreflang pairs derived from the route tree.
	Alternates map[string]string `json:"alternates,omitempty"`
}

// SearchResult is a ranked local-search result. The Router's built-in search
// is deliberately small and dependency-free, while SearchIndexJSON remains a
// portable hand-off point for Pagefind or another static indexer.
type SearchResult struct {
	// Entry is the matched record.
	Entry SearchEntry
	// Score ranks the result; higher wins.
	Score int
	// Matches are the terms that hit, for highlighting.
	Matches []string
}

// SearchProvider lets a project replace the JSON artifact with a richer index
// such as Pagefind while keeping route registration in the central Router.
type SearchProvider interface {
	Build(*Router) ([]byte, error)
}

// SearchBackend selects how browser search is served. JSON is the portable
// zero-dependency fallback; Pagefind is a static export enhancement that
// indexes generated HTML into a chunked browser-local bundle.
// defaultSearchIndexPath is the JSON search index URL mounted by the
// generated starter and by ExportStatic.
const defaultSearchIndexPath = "/__fastr-docs/search.json"

// SearchBackend selects how the command palette answers on this site: the
// JSON index the Router builds, or a Pagefind index built over the export.
// WithSearchBackend sets it.
type SearchBackend string

const (
	SearchBackendJSON     SearchBackend = "json"
	SearchBackendPagefind SearchBackend = "pagefind"
)

// WithSearchBackend selects the browser search integration. Invalid values
// fall back to JSON so a typo cannot make a development server unusable.
func WithSearchBackend(backend string) Option {
	return func(r *Router) {
		if strings.EqualFold(strings.TrimSpace(backend), string(SearchBackendPagefind)) {
			r.searchBackend = SearchBackendPagefind
			return
		}
		r.searchBackend = SearchBackendJSON
	}
}

// WithPagefind is the explicit form of WithSearchBackend for projects that
// want Pagefind in static exports and the native command palette.
func WithPagefind() Option {
	return WithSearchBackend(string(SearchBackendPagefind))
}

// WithPagefindPath changes the URL prefix used to load the generated
// Pagefind runtime. The default is /pagefind/; custom prefixes are useful when
// an export is mounted below a reverse-proxy asset path.
func WithPagefindPath(path string) Option {
	return func(r *Router) {
		path = strings.TrimSpace(path)
		if path == "" {
			return
		}
		if !strings.HasPrefix(path, "/") {
			path = "/" + path
		}
		if !strings.HasSuffix(path, "/") {
			path += "/"
		}
		r.pagefindPath = path
	}
}

// WithSearchIndexPath changes the URL the browser runtime fetches for the JSON
// search index. The default is /__fastr-docs/search.json, which matches the
// asset prefix the generated starter mounts. Hosts that serve the docs assets
// from another prefix must set this, or the command palette silently falls
// back to an unfiltered route list.
func WithSearchIndexPath(path string) Option {
	return func(r *Router) {
		path = strings.TrimSpace(path)
		if path == "" {
			return
		}
		if !strings.HasPrefix(path, "/") {
			path = "/" + path
		}
		r.searchIndexPath = path
	}
}

// JSONSearchProvider is the default portable search provider.
type JSONSearchProvider struct{}

func (JSONSearchProvider) Build(r *Router) ([]byte, error) {
	return json.MarshalIndent(r.searchIndexEntries(), "", "  ")
}

// WithSearchProvider replaces the default local JSON search artifact.
func WithSearchProvider(provider SearchProvider) Option {
	return func(r *Router) { r.searchProvider = provider }
}

// Plugin contributes routes or metadata to a Router.
type Plugin interface {
	Name() string
	Apply(*Router) error
}

// ValidatingPlugin adds plugin-specific checks to the Router's strict build
// validation without forcing every plugin to own a second validation pass.
type ValidatingPlugin interface {
	Plugin
	Validate(*Router) error
}

// AssetPlugin lets an extension mount its own browser assets through the
// project's GoFastr HTTP router. This is optional; ordinary route plugins need
// only implement Plugin.
type AssetPlugin interface {
	Plugin
	MountAssets(*Router, *gofastrRouter.Router) error
}

// Option configures a Router.
type Option func(*Router)

// LayoutConfig lets a project replace the generated global or section shell
// while keeping route registration, navigation, and search in Router. A
// section factory receives the top-level route that owns the active layout.
// Returning nil uses the built-in shell for that level.
type LayoutConfig struct {
	// Global builds the outermost layout wrapped around every
	// section; nil uses the built-in docs shell.
	Global func(*Router) *uiapp.Layout
	// Section builds the layout wrapping one top-level section;
	// nil keeps the built-in section shell with its sidebar.
	Section func(*Router, *Route) *uiapp.Layout
}

// WithLayouts installs optional layout factories. This is the extension point
// for a footer, custom section chrome, or a product-specific nested layout;
// callers do not need to fork Mount or duplicate route traversal.
func WithLayouts(config LayoutConfig) Option {
	return func(r *Router) {
		if config.Global != nil {
			r.layoutConfig.Global = config.Global
		}
		if config.Section != nil {
			r.layoutConfig.Section = config.Section
		}
	}
}

// WithSiteName sets the generated site name used by adapters and templates.
func WithSiteName(name string) Option {
	return func(r *Router) {
		if strings.TrimSpace(name) != "" {
			r.siteName = strings.TrimSpace(name)
		}
	}
}

// WithIncludeDrafts makes draft routes available to navigation, search, and
// mounting. It is intended for local preview and content review; production
// projects should leave the default (false) in place.
func WithIncludeDrafts(enabled bool) Option {
	return func(r *Router) { r.includeDrafts = enabled }
}

// WithLocale limits published content to the requested content locale.
// Routes without a locale remain shared across every locale. Use
// WithLanguage when the goal is only to set the document language while
// keeping runtime locale variants mounted.
func WithLocale(locale string) Option {
	return func(r *Router) { r.locale = strings.TrimSpace(locale) }
}

// WithLanguage sets the document language used by host adapters. It does not
// filter routes, which makes it safe for sites with runtime locale selectors.
func WithLanguage(language string) Option {
	return func(r *Router) {
		if strings.TrimSpace(language) != "" {
			r.language = strings.TrimSpace(language)
		}
	}
}

// WithVersion limits published content to the requested documentation
// version. Routes without a version remain shared across every version.
func WithVersion(version string) Option {
	return func(r *Router) { r.version = strings.TrimSpace(version) }
}

// WithStrictValidation keeps the default safety checks enabled or explicitly
// disables them for an advanced project. The CLI never disables strict mode
// unless the project opts out in code.
func WithStrictValidation(enabled bool) Option {
	return func(r *Router) { r.strict = enabled }
}

// Router is the central object for a documentation site. Register pages,
// screens, groups, and plugins here; adapters and generators derive their
// navigation and search surfaces from the same object.
type Router struct {
	routes                map[string]*Route
	roots                 []*Route
	seq                   int
	strict                bool
	siteName              string
	registrationErr       error
	navigationMounted     bool
	connectOrigins        []string
	commandPalette        *widget.Builder
	commandPaletteVisible render.HTML
	commandPaletteMounted bool
	includeDrafts         bool
	locale                string
	language              string
	version               string
	searchProvider        SearchProvider
	searchWeights         map[string]int
	searchBackend         SearchBackend
	pagefindPath          string
	searchIndexPath       string
	markdownComponents    map[string]MarkdownComponent
	markdownContainers    map[string]MarkdownContainer
	markdownRaws          map[string]MarkdownRawComponent
	markdownTransforms    []namedSourceTransform
	localeUI              map[string]UIStrings
	customNotFound        NotFoundScreen
	pageScripts           []PageScript
	customCSP             string
	localeNames           map[string]string
	gitMeta               *GitMetadataConfig
	gitMetaResolved       bool
	localeFallback        bool
	fallbackLocale        string
	localeFamilies        map[string]map[string]bool
	brand                 BrandConfig
	themeConfig           ThemeConfig
	ui                    UIStrings
	layoutConfig          LayoutConfig
	plugins               []Plugin
	blogs                 map[string]blogCollection
	blogViews             map[string]blogView
}

// NewRouter creates a strict Router. Strict validation is intentionally the
// default so broken links, missing titles, duplicate order, and unusable
// content fail during startup or build instead of shipping silently.
func NewRouter(options ...Option) *Router {
	r := &Router{
		routes:          make(map[string]*Route),
		strict:          true,
		siteName:        "Documentation",
		language:        "en",
		searchBackend:   SearchBackendJSON,
		pagefindPath:    "/pagefind/",
		searchIndexPath: defaultSearchIndexPath,
		// Registered before options run so WithoutDefaultComponents, and any
		// project registering the same name, still wins.
		markdownComponents: DefaultMarkdownComponents(),
		markdownContainers: DefaultMarkdownContainers(),
		markdownRaws:       DefaultMarkdownRawComponents(),
		blogs:              make(map[string]blogCollection),
		blogViews:          make(map[string]blogView),
		themeConfig:        ThemeConfig{Template: TemplateEditorial},
		ui:                 defaultUIStrings,
	}
	for _, option := range options {
		if option != nil {
			option(r)
		}
	}
	return r
}

// SiteName returns the configured site name.
func (r *Router) SiteName() string { return r.siteName }

// Locale returns the active content locale filter, if one was configured.
func (r *Router) Locale() string { return r.locale }

// Language returns the document language for the host shell. It is separate
// from Locale, which is an optional build-time content filter.
func (r *Router) Language() string {
	if r == nil || strings.TrimSpace(r.language) == "" {
		return "en"
	}
	return r.language
}

// Version returns the active documentation version filter, if one was
// configured.
func (r *Router) Version() string { return r.version }

// Locales returns the sorted locale values present in published route
// variants. It describes the whole registered site, not only the active
// WithLocale filter, so headers, export tools, and external adapters can build
// their own selectors without walking Route nodes.
func (r *Router) Locales() []string { return r.variantValues("locale") }

// Versions returns the sorted documentation version values present in
// published route variants. Version lifecycle tools can use this inventory to
// show or validate available versions while Router remains the source of
// truth for the actual pages.
func (r *Router) Versions() []string { return r.variantValues("version") }

func (r *Router) variantValues(dimension string) []string {
	if r == nil {
		return nil
	}
	seen := map[string]bool{}
	for _, route := range r.Routes() {
		if !r.variantPublished(route) {
			continue
		}
		value := route.Metadata.Locale
		if dimension == "version" {
			value = route.Metadata.Version
		}
		if value != "" {
			seen[value] = true
		}
	}
	values := make([]string, 0, len(seen))
	for value := range seen {
		values = append(values, value)
	}
	sort.Strings(values)
	return values
}

// RegisterMarkdownComponent adds a reusable typed component to every
// Markdown page. PageConfig.Components remains available for page-local
// registrations; global components are useful for a project's shared
// callouts, tabs, code samples, or custom content blocks.
func (r *Router) RegisterMarkdownComponent(name string, component MarkdownComponent) error {
	if r == nil {
		return errors.New("docs: RegisterMarkdownComponent requires a Router")
	}
	name = strings.TrimSpace(name)
	if !markdownShortcodeName.MatchString(name) || strings.HasPrefix(name, "/") {
		return fmt.Errorf("docs: invalid Markdown component name %q", name)
	}
	if component == nil {
		return fmt.Errorf("docs: Markdown component %q is nil", name)
	}
	if r.markdownComponents == nil {
		r.markdownComponents = make(map[string]MarkdownComponent)
	}
	delete(r.markdownContainers, name)
	delete(r.markdownRaws, name)
	r.markdownComponents[name] = component
	return nil
}

// MarkdownComponents returns a copy of the globally registered Markdown
// component registry for adapters, tooling, and plugin validation.
// RegisterMarkdownContainer registers a shortcode that receives its nested
// shortcodes as separate children. A name may resolve to a component or a
// container, so registering one clears the other.
func (r *Router) RegisterMarkdownContainer(name string, container MarkdownContainer) error {
	if r == nil {
		return errors.New("docs: RegisterMarkdownContainer requires a Router")
	}
	name = strings.TrimSpace(name)
	if !markdownShortcodeName.MatchString(name) {
		return fmt.Errorf("docs: invalid Markdown container name %q", name)
	}
	if container == nil {
		return fmt.Errorf("docs: Markdown container %q is nil", name)
	}
	if r.markdownContainers == nil {
		r.markdownContainers = make(map[string]MarkdownContainer)
	}
	delete(r.markdownComponents, name)
	delete(r.markdownRaws, name)
	r.markdownContainers[name] = container
	return nil
}

// RegisterMarkdownRawComponent registers a shortcode that receives its body
// unrendered. A name resolves to one kind, so registering one retires the rest.
func (r *Router) RegisterMarkdownRawComponent(name string, raw MarkdownRawComponent) error {
	if r == nil {
		return errors.New("docs: RegisterMarkdownRawComponent requires a Router")
	}
	name = strings.TrimSpace(name)
	if !markdownShortcodeName.MatchString(name) {
		return fmt.Errorf("docs: invalid Markdown component name %q", name)
	}
	if raw == nil {
		return fmt.Errorf("docs: Markdown raw component %q is nil", name)
	}
	if r.markdownRaws == nil {
		r.markdownRaws = make(map[string]MarkdownRawComponent)
	}
	delete(r.markdownComponents, name)
	delete(r.markdownContainers, name)
	r.markdownRaws[name] = raw
	return nil
}

// MarkdownRawComponents returns the registered raw vocabulary.
func (r *Router) MarkdownRawComponents() map[string]MarkdownRawComponent {
	if r == nil || len(r.markdownRaws) == 0 {
		return nil
	}
	return cloneMarkdownRaws(r.markdownRaws)
}

// MarkdownContainers returns the registered container vocabulary.
func (r *Router) MarkdownContainers() map[string]MarkdownContainer {
	if r == nil || len(r.markdownContainers) == 0 {
		return nil
	}
	return cloneMarkdownContainers(r.markdownContainers)
}

// vocabulary merges the Router's shortcode namespace with a page's local
// additions. Page-local entries win.
func (r *Router) vocabulary(components map[string]MarkdownComponent, containers map[string]MarkdownContainer) markdownVocabulary {
	return markdownVocabulary{
		components: mergeMarkdownComponents(r.markdownComponents, components),
		containers: mergeMarkdownContainers(r.markdownContainers, containers),
		raws:       mergeMarkdownRaws(r.markdownRaws, nil),
		transforms: r.markdownTransforms,
	}
}

// MarkdownComponents returns a copy of the registered Markdown components by
// name, so callers can inspect or compose them without being able to mutate
// the Router's registry.
func (r *Router) MarkdownComponents() map[string]MarkdownComponent {
	if r == nil || len(r.markdownComponents) == 0 {
		return nil
	}
	return cloneMarkdownComponents(r.markdownComponents)
}

// SearchBackend returns the configured browser search backend.
func (r *Router) SearchBackend() SearchBackend {
	if r == nil || r.searchBackend == "" {
		return SearchBackendJSON
	}
	return r.searchBackend
}

// PagefindPath returns the static Pagefind bundle directory used by the
// browser runtime.
func (r *Router) PagefindPath() string {
	if r == nil || strings.TrimSpace(r.pagefindPath) == "" {
		return "/pagefind/"
	}
	return r.pagefindPath
}

// SearchIndexPath returns the URL the browser runtime fetches for the JSON
// search index.
func (r *Router) SearchIndexPath() string {
	if r == nil || strings.TrimSpace(r.searchIndexPath) == "" {
		return defaultSearchIndexPath
	}
	return r.searchIndexPath
}

// SearchIndexCacheControl returns the Cache-Control header a host should
// serve the JSON search index with. The index is content-addressed per
// build, so it is immutable and safe to cache hard.
func (r *Router) SearchIndexCacheControl() string {
	return "public, max-age=31536000, immutable"
}

// PageScript is one host-mountable browser script contributed by the project
// alongside the framework runtime.
type PageScript struct {
	Name string
	JS   string
}

// WithPageScript contributes a project script to the runtime asset set: it
// is served beside docs.js and listed by RuntimeScriptNames, so analytics or
// product glue ships without editing the generated host.
func WithPageScript(name, js string) Option {
	return func(r *Router) {
		if r.pageScripts == nil {
			r.pageScripts = []PageScript{}
		}
		r.pageScripts = append(r.pageScripts, PageScript{Name: name, JS: js})
	}
}

// PageScripts lists the project-contributed scripts in registration order.
func (r *Router) PageScripts() []PageScript {
	if r == nil || len(r.pageScripts) == 0 {
		return nil
	}
	return append([]PageScript(nil), r.pageScripts...)
}

// Sitemap builds a sitemap.xml URL set from the published routes: every
// servable path that search engines may index, with the locale alternates
// each page pairs with. NoIndex pages are left out.
func (r *Router) Sitemap() []byte {
	var b strings.Builder
	b.WriteString(`<?xml version="1.0" encoding="UTF-8"?>` + "\n")
	b.WriteString(`<urlset xmlns="http://www.sitemaps.org/schemas/sitemap/0.9">` + "\n")
	for _, route := range r.PublishedRoutes() {
		if route == nil || route.Metadata.NoIndex {
			continue
		}
		b.WriteString(`<url><loc>` + stdhtml.EscapeString(route.Path) + `</loc>`)
		for lang, target := range r.alternatesFor(route) {
			b.WriteString(`<xhtml:link rel="alternate" hreflang="` + stdhtml.EscapeString(lang) + `" href="` + stdhtml.EscapeString(target) + `"/>`)
		}
		if date := strings.TrimSpace(route.Metadata.DateModified); date != "" {
			b.WriteString(`<lastmod>` + stdhtml.EscapeString(date) + `</lastmod>`)
		} else if date := strings.TrimSpace(route.Metadata.DatePublished); date != "" {
			b.WriteString(`<lastmod>` + stdhtml.EscapeString(date) + `</lastmod>`)
		}
		b.WriteString(`</url>` + "\n")
	}
	b.WriteString(`</urlset>`)
	return []byte(b.String())
}

// AllowConnectOrigin records an origin that a browser-side extension is
// expected to contact. The generated GoFastr host can use ConnectOrigins to
// extend its strict CSP without allowing arbitrary network destinations.
func (r *Router) AllowConnectOrigin(raw string) {
	origin := normalizeConnectOrigin(raw)
	if origin == "" {
		return
	}
	for _, existing := range r.connectOrigins {
		if existing == origin {
			return
		}
	}
	r.connectOrigins = append(r.connectOrigins, origin)
}

// ConnectOrigins returns the normalized external origins declared by
// extensions on this router.
func (r *Router) ConnectOrigins() []string {
	return append([]string(nil), r.connectOrigins...)
}

// ContentSecurityPolicy returns the strict default GoFastr policy with only
// the explicitly declared browser connect origins added to connect-src.
func ContentSecurityPolicy(connectOrigins ...string) string {
	sources := []string{"'self'"}
	seen := map[string]bool{"'self'": true}
	for _, raw := range connectOrigins {
		origin := normalizeConnectOrigin(raw)
		if origin == "" || seen[origin] {
			continue
		}
		seen[origin] = true
		sources = append(sources, origin)
	}
	return "default-src 'self'; img-src 'self' data:; object-src 'none'; form-action 'self'; frame-ancestors 'none'; base-uri 'self'; upgrade-insecure-requests; connect-src " + strings.Join(sources, " ")
}

// WithContentSecurityPolicy replaces the default policy wholesale, for hosts
// that must match an organization-wide header. The framework's strict
// default is a good starting point: copy it before loosening.
func WithContentSecurityPolicy(policy string) Option {
	return func(r *Router) {
		r.customCSP = policy
	}
}

// ContentSecurityPolicy returns the effective policy for this Router: the
// project's override when set, otherwise the strict default with the
// declared connect origins.
func (r *Router) ContentSecurityPolicy() string {
	if r != nil && strings.TrimSpace(r.customCSP) != "" {
		return r.customCSP
	}
	return ContentSecurityPolicy(r.ConnectOrigins()...)
}

func normalizeConnectOrigin(raw string) string {
	parsed, err := url.Parse(strings.TrimSpace(raw))
	if err != nil || parsed.User != nil || parsed.Host == "" {
		return ""
	}
	if parsed.Scheme != "http" && parsed.Scheme != "https" {
		return ""
	}
	return parsed.Scheme + "://" + parsed.Host
}

// Group adds a metadata group and returns a scoped registrar for its children.
func (r *Router) Group(path string, cfg GroupConfig) (*Group, error) {
	badge, err := normalizeNavBadge(cfg.Badge)
	if err != nil {
		return nil, fmt.Errorf("docs: group %q: %w", path, err)
	}
	route, err := r.addRoute(path, Route{
		Title:       cfg.Title,
		Description: cfg.Description,
		Kind:        KindGroup,
		Order:       cfg.Order,
		Hidden:      cfg.Hidden,
		Badge:       badge,
		Metadata: ContentMetadata{
			Locale:        strings.TrimSpace(cfg.Locale),
			Version:       strings.TrimSpace(cfg.Version),
			TranslationOf: strings.TrimSpace(cfg.TranslationOf),
		},
	})
	if err != nil {
		return nil, err
	}
	return &Group{router: r, route: route}, nil
}

// MustGroup is the concise form for declarative app setup. It panics during
// startup when the route tree is malformed, which is preferable to a server
// running with an incomplete navigation tree.
func (r *Router) MustGroup(path string, cfg GroupConfig) *Group {
	g, err := r.Group(path, cfg)
	if err != nil {
		panic(err)
	}
	return g
}

// Page registers a page at path.
func (r *Router) Page(path string, cfg PageConfig) error {
	if err := validatePreload(cfg.Preload); err != nil {
		return fmt.Errorf("docs: page %q: %w", path, err)
	}
	badge, err := normalizeNavBadge(cfg.Badge)
	if err != nil {
		return fmt.Errorf("docs: page %q: %w", path, err)
	}
	metadata, source, err := pageMetadata(cfg)
	if err != nil {
		return fmt.Errorf("docs: page %q: %w", path, err)
	}
	if err := validateMarkdownComponents(source, r.vocabulary(cfg.Components, cfg.Containers)); err != nil {
		return fmt.Errorf("docs: page %q: %w", path, err)
	}
	cfg.Metadata = metadata
	if source != "" {
		cfg.Source = source
	}
	if cfg.Title == "" {
		cfg.Title = metadata.Title
	}
	if cfg.Description == "" {
		cfg.Description = metadata.Description
		if cfg.Description == "" && cfg.Source != "" {
			cfg.Description = firstParagraph(cfg.Source)
		}
	}
	if len(cfg.Tags) == 0 {
		cfg.Tags = cloneStrings(metadata.Tags)
	}
	if cfg.Order < 1 && metadata.Order > 0 {
		cfg.Order = metadata.Order
	}
	_, err = r.addRoute(path, Route{
		Title:       cfg.Title,
		Description: cfg.Description,
		Kind:        KindPage,
		Order:       cfg.Order,
		Offline:     cfg.Offline,
		Hidden:      cfg.Hidden,
		Tags:        cloneStrings(cfg.Tags),
		SearchText:  cfg.SearchText,
		Preload:     cfg.Preload,
		Metadata:    metadata,
		Badge:       badge,
		page:        &cfg,
	})
	return err
}

// MustPage registers a page and panics on invalid configuration.
func (r *Router) MustPage(path string, cfg PageConfig) {
	if err := r.Page(path, cfg); err != nil {
		panic(err)
	}
}

// Markdown is a named convenience for Page when the source is Markdown.
func (r *Router) Markdown(path string, title string, source string) error {
	return r.Page(path, PageConfig{Title: title, Source: source})
}

// PageFile registers a Markdown page whose source is loaded and validated
// from SourcePath during registration. Keeping the path in the route config
// means a generated project does not need boilerplate file-reading helpers.
func (r *Router) PageFile(path string, cfg PageConfig) error {
	return r.Page(path, cfg)
}

// MustPageFile is the panic-on-error form of PageFile.
func (r *Router) MustPageFile(path string, cfg PageConfig) {
	if err := r.PageFile(path, cfg); err != nil {
		panic(err)
	}
}

// Screen registers a typed GoFastr component in the same route tree.
func (r *Router) Screen(path string, cfg ScreenConfig) error {
	if err := validatePreload(cfg.Preload); err != nil {
		return fmt.Errorf("docs: screen %q: %w", path, err)
	}
	badge, err := normalizeNavBadge(cfg.Badge)
	if err != nil {
		return fmt.Errorf("docs: screen %q: %w", path, err)
	}
	metadata := mergeContentMetadata(cfg.Metadata, ContentMetadata{
		Title: cfg.Title, Description: cfg.Description, Order: cfg.Order,
		Tags: cfg.Tags,
	})
	cfg.Metadata = metadata
	if cfg.Title == "" {
		cfg.Title = metadata.Title
	}
	if cfg.Description == "" {
		cfg.Description = metadata.Description
	}
	if len(cfg.Tags) == 0 {
		cfg.Tags = cloneStrings(metadata.Tags)
	}
	if cfg.Order < 1 && metadata.Order > 0 {
		cfg.Order = metadata.Order
	}
	_, err = r.addRoute(path, Route{
		Title:       cfg.Title,
		Description: cfg.Description,
		Kind:        KindScreen,
		Order:       cfg.Order,
		Offline:     cfg.Offline,
		Hidden:      cfg.Hidden,
		Tags:        cloneStrings(cfg.Tags),
		SearchText:  cfg.SearchText,
		Preload:     cfg.Preload,
		Plugin:      cfg.Plugin,
		Metadata:    metadata,
		Badge:       badge,
		screen:      &cfg,
	})
	return err
}

// MustScreen registers a typed screen and panics on invalid configuration.
func (r *Router) MustScreen(path string, cfg ScreenConfig) {
	if err := r.Screen(path, cfg); err != nil {
		panic(err)
	}
}

// Use applies a plugin to this Router. A plugin is not a separate app: its
// routes become ordinary children of the same route tree.
func (r *Router) Use(plugin Plugin) error {
	if plugin == nil {
		return errors.New("docs: nil plugin")
	}
	if err := plugin.Apply(r); err != nil {
		return fmt.Errorf("docs: plugin %q: %w", plugin.Name(), err)
	}
	r.plugins = append(r.plugins, plugin)
	return nil
}

// MountPluginAssets invokes optional asset hooks after the host HTTP router
// exists. It is safe to call more than once for idempotent plugins, but a
// project should normally call it once during host setup.
func (r *Router) MountPluginAssets(httpRouter *gofastrRouter.Router) error {
	if httpRouter == nil {
		return errors.New("docs: MountPluginAssets requires a GoFastr router")
	}
	for _, plugin := range r.plugins {
		if assetPlugin, ok := plugin.(AssetPlugin); ok {
			if err := assetPlugin.MountAssets(r, httpRouter); err != nil {
				return fmt.Errorf("docs: plugin %q assets: %w", plugin.Name(), err)
			}
		}
	}
	return nil
}

// Routes returns all navigable (non-group) routes in stable tree order.
func (r *Router) Routes() []*Route {
	var out []*Route
	var walk func([]*Route)
	walk = func(nodes []*Route) {
		for _, node := range r.sorted(nodes) {
			if node.Kind != KindGroup {
				out = append(out, node)
			}
			walk(node.Children)
		}
	}
	walk(r.roots)
	return out
}

// PublishedRoutes returns routes safe for a public build. Drafts, hidden
// routes, and routes outside the active locale/version are excluded.
func (r *Router) PublishedRoutes() []*Route {
	var out []*Route
	for _, route := range r.Routes() {
		if r.routePublished(route) {
			out = append(out, route)
		}
	}
	return out
}

// SitemapExcludePaths returns route prefixes that must not be emitted in a
// sitemap: hidden, draft, locale/version-inactive, and explicit noindex
// content. Hosts can pass the result directly to uihost.SitemapConfig.
func (r *Router) SitemapExcludePaths() []string {
	seen := map[string]bool{}
	var out []string
	for _, route := range r.Routes() {
		if !r.routePublished(route) || route.Metadata.NoIndex {
			if route.Path != "" && !seen[route.Path] {
				seen[route.Path] = true
				out = append(out, route.Path)
			}
		}
	}
	sort.Strings(out)
	return out
}

// Tree returns a copy of the root slice. Route nodes are immutable after
// registration by convention; callers should not mutate them.
func (r *Router) Tree() []*Route { return append([]*Route(nil), r.sorted(r.roots)...) }

// Navigation returns a flattened navigation model, including groups, in the
// same order the site should render them.
func (r *Router) Navigation() []NavItem {
	return r.NavigationAt("")
}

// NavigationAt returns the flattened navigation model with active state
// derived from currentPath. Prefix matching keeps a parent page/group active
// while a nested route is open; the root route only matches exactly.
func (r *Router) NavigationAt(currentPath string) []NavItem {
	currentPath = normalizePath(currentPath)
	var out []NavItem
	var walk func([]*Route, int)
	walk = func(nodes []*Route, depth int) {
		for _, node := range r.sorted(nodes) {
			if !r.routeVisible(node) {
				continue
			}
			out = append(out, NavItem{ID: node.ID, Path: node.Path, Title: node.Title, Kind: node.Kind, Depth: depth, Active: pathActive(node.Path, currentPath), Children: len(node.Children), Badge: node.Badge})
			if len(node.Children) > 0 {
				walk(node.Children, depth+1)
			}
		}
	}
	walk(r.roots, 0)
	return out
}

// SearchIndex returns records for visible navigable routes. Markdown source is
// preserved as text so a local indexer can tokenize it later.
func (r *Router) SearchIndex() []SearchEntry {
	entries := make([]SearchEntry, 0, len(r.Routes()))
	for _, route := range r.PublishedRoutes() {
		if !r.routePublished(route) || route.Metadata.NoIndex {
			continue
		}
		entry := SearchEntry{
			ID: route.ID, Path: route.Path, Title: route.Title,
			Description: route.Description, Tags: cloneStrings(route.Tags),
			// The effective locale, not the declared one. A page that declares
			// none still belongs to the default language, and a search filter
			// comparing raw values would drop every unmarked page, including
			// typed screens, which have no front matter to declare one in.
			Locale: r.effectiveLocale(route), Version: route.Metadata.Version,
			EditURL: route.Metadata.EditURL, Canonical: route.Metadata.CanonicalURL,
			NoIndex: route.Metadata.NoIndex, Kind: route.Kind, Order: route.Order,
			Alternates: r.alternatesFor(route),
		}
		if route.page != nil {
			entry.Text = pageSource(route.page)
			for _, heading := range markdownHeadings(entry.Text) {
				entry.Headings = append(entry.Headings, heading.Title)
			}
		}
		if entry.Text == "" {
			entry.Text = route.SearchText
			// Screens carry no Markdown page, but their SearchText often
			// does; the headings ride along when it does.
			for _, heading := range markdownHeadings(entry.Text) {
				entry.Headings = append(entry.Headings, heading.Title)
			}
		}
		entries = append(entries, entry)
	}
	return entries
}

// SearchIndexJSON serializes the portable search records for a local static
// indexer such as Pagefind or a small client-side search worker.
func (r *Router) SearchIndexJSON() ([]byte, error) {
	if err := r.Validate(); err != nil {
		return nil, err
	}
	if r.searchProvider != nil {
		return r.searchProvider.Build(r)
	}
	return json.MarshalIndent(r.searchIndexEntries(), "", "  ")
}

func (r *Router) searchIndexEntries() []SearchEntry { return r.SearchIndex() }

// Search ranks published content using title, description, tags, headings,
// and body text. Exact title matches receive the strongest weight; ties are
// stable in intentional route order.
func (r *Router) Search(query string) []SearchResult {
	terms := searchTerms(query)
	if len(terms) == 0 {
		return nil
	}
	var results []SearchResult
	for _, entry := range r.SearchIndex() {
		fields := []struct {
			name  string
			value string
		}{
			{"title", entry.Title}, {"description", entry.Description},
			// Tags are matched folded, so "diseno" finds a tag written
			// "diseño" the same way body text already does.
			{"tags", strings.Join(entry.Tags, " ")}, {"headings", strings.Join(entry.Headings, " ")},
			{"body", entry.Text},
		}
		score := 0
		matched := make([]string, 0, len(terms))
		for _, term := range terms {
			termScore := 0
			for _, field := range fields {
				if strings.Contains(foldSearchText(field.value), term) {
					termScore = maxInt(termScore, r.searchWeight(field.name))
				}
			}
			if termScore == 0 {
				continue
			}
			score += termScore
			matched = append(matched, term)
		}
		if len(matched) == len(terms) {
			results = append(results, SearchResult{Entry: entry, Score: score, Matches: matched})
		}
	}
	sort.SliceStable(results, func(i, j int) bool {
		if results[i].Score != results[j].Score {
			return results[i].Score > results[j].Score
		}
		return results[i].Entry.Order < results[j].Entry.Order
	})
	return results
}

// foldSearchText lowercases and strips diacritics so "traduccion" matches
// "traducción". Ascii-folded keys are how search stays usable in languages
// whose readers type without accents; the display text is never altered.
func foldSearchText(value string) string {
	var foldTable = map[rune]rune{
		'à': 'a', 'á': 'a', 'â': 'a', 'ã': 'a', 'ä': 'a', 'å': 'a', 'ā': 'a', 'ă': 'a', 'ą': 'a',
		'ç': 'c', 'ć': 'c', 'ĉ': 'c', 'ċ': 'c', 'č': 'c',
		'è': 'e', 'é': 'e', 'ê': 'e', 'ë': 'e', 'ē': 'e', 'ĕ': 'e', 'ė': 'e', 'ę': 'e', 'ě': 'e',
		'ì': 'i', 'í': 'i', 'î': 'i', 'ï': 'i', 'ĩ': 'i', 'ī': 'i', 'ĭ': 'i', 'į': 'i', 'ı': 'i',
		'ñ': 'n', 'ń': 'n', 'ņ': 'n', 'ň': 'n',
		'ò': 'o', 'ó': 'o', 'ô': 'o', 'õ': 'o', 'ö': 'o', 'ø': 'o', 'ō': 'o', 'ŏ': 'o', 'ő': 'o',
		'ù': 'u', 'ú': 'u', 'û': 'u', 'ü': 'u', 'ũ': 'u', 'ū': 'u', 'ŭ': 'u', 'ů': 'u', 'ű': 'u', 'ų': 'u',
		'ý': 'y', 'ÿ': 'y', 'ŷ': 'y',
		'đ': 'd', 'ď': 'd', 'ð': 'd',
		'ł': 'l', 'ĺ': 'l', 'ľ': 'l', 'ŀ': 'l',
		'ŕ': 'r', 'ŗ': 'r', 'ř': 'r',
		'ś': 's', 'ŝ': 's', 'ş': 's', 'š': 's', 'ß': 's',
		't': 't', 'ţ': 't', 'ť': 't',
		'ź': 'z', 'ż': 'z', 'ž': 'z',
		'ĝ': 'g', 'ğ': 'g',
		'ĥ': 'h', 'ħ': 'h',
		'ĵ': 'j',
		'ķ': 'k',
		'æ': 'a', 'œ': 'o',
	}
	folded := make([]rune, 0, len(value))
	for _, r := range strings.ToLower(value) {
		if base, ok := foldTable[r]; ok {
			r = base
		}
		folded = append(folded, r)
	}
	return string(folded)
}

// SearchLimited bounds the result count, so a caller showing three hints
// under a search box does not rank the whole index and truncate by hand.
func (r *Router) SearchLimited(query string, limit int) []SearchResult {
	results := r.Search(query)
	if limit > 0 && len(results) > limit {
		results = results[:limit]
	}
	return results
}

func searchTerms(query string) []string {
	var terms []string
	seen := map[string]bool{}
	for _, term := range strings.Fields(foldSearchText(query)) {
		term = strings.Trim(term, ".,:;!?()[]{}\"")
		if term != "" && !seen[term] {
			seen[term] = true
			terms = append(terms, term)
		}
	}
	return terms
}

func (r *Router) searchWeight(field string) int {
	if weight, ok := r.searchWeights[field]; ok {
		return weight
	}
	switch field {
	case "title":
		return 12
	case "headings":
		return 8
	case "description":
		return 6
	case "tags":
		return 5
	}
	return 1
}

// WithSearchWeights tunes the built-in ranker. The map names a field (title,
// headings, description, tags, body) and the weight a term match there adds;
// fields left out keep their defaults, and unknown names are ignored.
func WithSearchWeights(weights map[string]int) Option {
	return func(r *Router) {
		if r.searchWeights == nil {
			r.searchWeights = map[string]int{}
		}
		for field, weight := range weights {
			r.searchWeights[field] = weight
		}
	}
}

func maxInt(a, b int) int {
	if a > b {
		return a
	}
	return b
}

// WriteSearchIndex writes SearchIndexJSON to an output stream.
func (r *Router) WriteSearchIndex(w io.Writer) error {
	if w == nil {
		return errors.New("docs: search index writer is nil")
	}
	body, err := r.SearchIndexJSON()
	if err != nil {
		return err
	}
	_, err = w.Write(body)
	return err
}

// Warnings reports advisory findings that never fail Validate: draft
// pages counting down to publication, and locale label sets that are only
// partly translated. A build stays usable while incomplete; these are the
// notes that say what is still rough.
func (r *Router) Warnings() []string {
	if r == nil {
		return nil
	}
	var warnings []string
	for _, route := range r.Routes() {
		if route.Metadata.Draft {
			warnings = append(warnings, fmt.Sprintf("route %q is a draft; it is excluded from every build until drafts are included", route.Path))
		}
	}
	locales := make([]string, 0, len(r.localeUI))
	for locale := range r.localeUI {
		locales = append(locales, locale)
	}
	sort.Strings(locales)
	for _, locale := range locales {
		set := r.localeUI[locale]
		blog := reflect.ValueOf(set.Blog)
		blogType := blog.Type()
		set_, total := 0, 0
		for i := range blogType.NumField() {
			if blogType.Field(i).Type.Kind() != reflect.String {
				continue
			}
			total++
			if blog.Field(i).String() != "" {
				set_++
			}
		}
		if set_ > 0 && set_ < total {
			warnings = append(warnings, fmt.Sprintf("locale %q translates %d of %d blog labels", locale, set_, total))
		}
	}
	return warnings
}

// Validate checks the full route tree and returns all problems together.
func (r *Router) Validate() error {
	// Fills in edit links and last-updated dates before anything reads route
	// metadata. Mount calls Validate, so both paths are covered.
	r.resolveGitMetadata()
	var problems []string
	if r.registrationErr != nil {
		problems = append(problems, r.registrationErr.Error())
	}
	for _, plugin := range r.plugins {
		if validator, ok := plugin.(ValidatingPlugin); ok {
			if err := validator.Validate(r); err != nil {
				problems = append(problems, fmt.Sprintf("plugin %q: %v", plugin.Name(), err))
			}
		}
	}
	var walk func([]*Route)
	walk = func(nodes []*Route) {
		seenOrders := map[int]string{}
		seenTitles := map[string]string{}
		for _, node := range nodes {
			if r.strict && node.Order < 1 {
				problems = append(problems, fmt.Sprintf("route %q: a positive explicit Order is required", node.Path))
			}
			if r.strict && node.Title == "" && node.Kind != KindGroup {
				problems = append(problems, fmt.Sprintf("route %q: title is required", node.Path))
			}
			if r.strict && node.Kind == KindGroup && node.Title == "" {
				problems = append(problems, fmt.Sprintf("group %q: title is required", node.Path))
			}
			if r.strict && node.Kind != KindGroup && node.Description == "" {
				problems = append(problems, fmt.Sprintf("route %q: description is required", node.Path))
			}
			if r.strict && node.Order != 0 {
				if previous, ok := seenOrders[node.Order]; ok {
					problems = append(problems, fmt.Sprintf("routes %q and %q: duplicate explicit order %d", previous, node.Path, node.Order))
				}
				seenOrders[node.Order] = node.Path
			}
			// Titles may repeat across languages: an untranslated
			// section keeps its original title beside the translated
			// tree. One language naming two siblings the same thing is
			// the mistake worth catching.
			if r.strict && node.Title != "" {
				if previous, ok := seenTitles[node.Title]; ok {
					if before := r.routes[previous]; before == nil || r.effectiveLocale(before) == r.effectiveLocale(node) {
						problems = append(problems, fmt.Sprintf("routes %q and %q: siblings share the title %q in one language", previous, node.Path, node.Title))
					}
				}
				seenTitles[node.Title] = node.Path
			}
			if r.strict && (node.Kind == KindPage || node.Kind == KindPlugin) && node.page != nil && node.page.Source == "" && node.page.Body == nil && node.page.ContextBody == nil && node.page.SourcePath == "" {
				problems = append(problems, fmt.Sprintf("page %q: Source, SourcePath, or Body is required", node.Path))
			}
			if r.strict && (node.Kind == KindPage || node.Kind == KindPlugin) && node.page != nil && node.page.Source == "" && node.page.Body == nil && node.page.SourcePath != "" {
				if _, err := os.Stat(node.page.SourcePath); err != nil {
					problems = append(problems, fmt.Sprintf("page %q: source file %q is not readable: %v", node.Path, node.page.SourcePath, err))
				}
			}
			if node.page != nil && node.page.Body == nil {
				if err := validateMarkdownComponents(pageSource(node.page), r.vocabulary(node.page.Components, node.page.Containers)); err != nil {
					problems = append(problems, fmt.Sprintf("page %q: %v", node.Path, err))
				}
			}
			if r.strict && node.Kind == KindScreen && (node.screen == nil || node.screen.Component == nil) {
				problems = append(problems, fmt.Sprintf("screen %q: Component is required", node.Path))
			}
			for _, redirect := range node.Metadata.Redirects {
				if safeRedirectPath(redirect) == "" {
					problems = append(problems, fmt.Sprintf("route %q: redirect %q must be an internal path", node.Path, redirect))
				}
				if redirect == node.Path {
					problems = append(problems, fmt.Sprintf("route %q redirects onto itself", node.Path))
				}
				if target := r.routes[normalizePath(redirect)]; target != nil && target != node {
					for _, chained := range target.Metadata.Redirects {
						if chained == redirect || chained == node.Path {
							problems = append(problems, fmt.Sprintf("routes %q and %q form a redirect cycle through %q", node.Path, target.Path, redirect))
						}
					}
				}
			}
			if r.strict {
				if canonical := strings.TrimSpace(node.Metadata.CanonicalURL); canonical != "" && !strings.HasPrefix(canonical, "http://") && !strings.HasPrefix(canonical, "https://") {
					problems = append(problems, fmt.Sprintf("route %q: canonical URL %q must be absolute", node.Path, canonical))
				}
				for lang, target := range node.Metadata.Alternates {
					if !localeShape.MatchString(lang) {
						problems = append(problems, fmt.Sprintf("route %q: alternate key %q is not a language tag", node.Path, lang))
					}
					if r.routes[normalizePath(target)] == nil {
						problems = append(problems, fmt.Sprintf("route %q: alternate %q points at %q, which no route serves", node.Path, lang, target))
					}
				}
			}
			walk(node.Children)
		}
	}
	walk(r.roots)
	if r.strict {
		for _, issue := range r.ContentIssues() {
			problems = append(problems, issue.Error())
		}
	}
	if len(problems) > 0 {
		return errors.New(strings.Join(problems, "; "))
	}
	return nil
}

// Mount registers every navigable route with a GoFastr UI app. The route metadata
// remains owned by docs.Router; GoFastr owns rendering, DI, policies, and
// lifecycle after registration.
func (r *Router) Mount(site *uiapp.App, layout *uiapp.Layout) error {
	if site == nil {
		return errors.New("docs: Mount requires a GoFastr app")
	}
	if err := r.Validate(); err != nil {
		return err
	}
	if layout == nil {
		layout = r.Layout()
	}
	for _, route := range r.PublishedRoutes() {
		for _, redirect := range route.Metadata.Redirects {
			if from := safeRedirectPath(redirect); from != "" && from != route.Path {
				site.Redirect(from, route.Path)
			}
		}
	}
	// The supplied layout is the global shell. Each top-level route section
	// gets its own nested ScreenGroup layout below, so switching sections can
	// replace the section shell while preserving the shared header/search shell.
	site.SetDefaultLayout(layout)
	groups := make(map[string]*uiapp.ScreenGroup)
	var orderedGroups []*uiapp.ScreenGroup
	for _, route := range r.PublishedRoutes() {
		var screen *uiapp.Screen
		switch route.Kind {
		case KindPage, KindPlugin:
			screen = uiapp.NewScreen(route.Path, &pageComponent{router: r, route: route}).
				WithTitle(route.Title).
				WithDescription(route.Description)
		case KindScreen:
			// Every screen is wrapped, not only those with metadata: the
			// wrapper also carries the page's chrome template, which the
			// runtime needs on every page to refresh the header after a
			// client-side navigation.
			screenComponent := metadataScreenComponent(r, route, route.screen.Component)
			screen = uiapp.NewScreen(route.Path, screenComponent).
				WithTitle(route.Title).
				WithDescription(route.Description)
		default:
			continue
		}
		screen.Preload = route.Preload
		root := r.rootRoute(route)
		// A blog collection gets its own group whatever its depth, so its
		// pages take the blog layout and sidebar. /es/blog sits under the
		// Spanish home; grouped by tree root it wore the docs sidebar.
		if blogRoot := r.blogRootFor(route); blogRoot != nil {
			root = blogRoot
		}
		if root == nil {
			site.RegisterScreen(screen, layout)
			continue
		}
		group := groups[root.Path]
		if group == nil {
			group = uiapp.NewScreenGroup(root.Path, r.sectionLayout(root))
			groups[root.Path] = group
			orderedGroups = append(orderedGroups, group)
		}
		group.Screen(screen, nil)
	}
	for _, group := range orderedGroups {
		site.Router.ScreenGroup(group)
	}
	return nil
}

// MountNavigation mounts the framework-owned mobile docs Sidebar drawers.
// Call it on the GoFastr HTTP router after creating the UI host and before
// starting or exporting the application. The route tree remains owned by
// this Router; GoFastr owns drawer focus, escape handling, backdrop, and
// scroll locking.
//
// One drawer is mounted per language, because a widget's body is fixed at
// mount time and a single drawer would serve every page the default
// locale's tree. Each drawer also carries a section select above the tree:
// below md the header tabs are hidden and the drawer is the only
// navigation left.
func (r *Router) MountNavigation(httpRouter *gofastrRouter.Router) error {
	if httpRouter == nil {
		return errors.New("docs: MountNavigation requires a GoFastr router")
	}
	if r.navigationMounted {
		return nil
	}
	mounter := routerMounter{router: httpRouter}
	for i, home := range r.localeDrawerHomes() {
		// The first entry is the default drawer whatever the project
		// declared: the header trigger's global-drawer fallback aims at
		// the pinned short name, so it must exist on every site, home
		// page or none, fallback locale declared or not. Every later
		// entry derives its name from its locale.
		drawer := docsDrawerName("")
		if i > 0 {
			drawer = docsDrawerName(r.effectiveLocale(home))
		}
		currentPath := ""
		if home != nil && home.Path != "/" {
			currentPath = home.Path
		}
		mountNavigationDrawer(mounter, r.docsDrawerConfig(home, drawer), r.docsSectionSelect(currentPath, drawer+"-section"))
	}
	for _, prefix := range r.blogPrefixesList() {
		cfg := (&blogSidebar{router: r, prefix: prefix}).config("", blogDrawerName(prefix))
		mountNavigationDrawer(mounter, cfg, r.docsSectionSelect(prefix, cfg.DrawerName+"-section"))
	}
	r.navigationMounted = true
	return nil
}

// NavigationDrawerNames lists every navigation drawer this Router mounts:
// the default drawer, one per additional locale, and one per blog
// collection. The export manifest publishes them so an offline host can
// precache exactly the chrome that exists.
func (r *Router) NavigationDrawerNames() []string {
	return r.navigationDrawerNames()
}

func (r *Router) navigationDrawerNames() []string {
	if r == nil {
		return nil
	}
	names := []string{}
	for i, home := range r.localeDrawerHomes() {
		if i == 0 {
			names = append(names, docsDrawerName(""))
			continue
		}
		if home != nil {
			names = append(names, docsDrawerName(r.effectiveLocale(home)))
		}
	}
	for _, prefix := range r.blogPrefixesList() {
		names = append(names, blogDrawerName(prefix))
	}
	sort.Strings(names[1:])
	return names
}

// mountNavigationDrawer mounts one drawer widget with the section select
// above the sidebar tree. It is what ui.MountSidebar builds, minus its
// fixed body slot: SidebarConfig has no way to put content above the nav
// (gofastr #405), so the body is navigationDrawerBody instead.
func mountNavigationDrawer(mount ui.WidgetMounter, cfg ui.SidebarConfig, sections render.HTML) {
	b := preset.Drawer(cfg.DrawerName).
		Hidden().
		Slot("body", navigationDrawerBody{cfg: cfg, sections: sections})
	def := b.Build()
	mount.MountWidget(&def)
}

// MountCommandPalette mounts GoFastr's native command/search surface. The
// visible trigger is rendered by Layout; the widget itself must be mounted on
// the host router once, just like any other GoFastr widget.
func (r *Router) MountCommandPalette(httpRouter *gofastrRouter.Router) error {
	if httpRouter == nil {
		return errors.New("docs: MountCommandPalette requires a GoFastr router")
	}
	if r.commandPaletteMounted {
		return nil
	}
	_, palette := r.ensureCommandPalette()
	widget.Mount(httpRouter, palette.Definition())
	r.commandPaletteMounted = true
	return nil
}

// Layout returns the global white-label shell derived from this Router.
// Top-level route sections are nested inside this shell by Mount, allowing
// each section to own its own layout and navigation without duplicating the
// global header, search, theme, or PWA chrome.
func (r *Router) Layout() *uiapp.Layout {
	if r != nil && r.layoutConfig.Global != nil {
		if layout := r.layoutConfig.Global(r); layout != nil {
			return layout
		}
	}
	return uiapp.NewLayout("docs").WithHeader(&docsHeader{router: r})
}

func (r *Router) sectionLayout(route *Route) *uiapp.Layout {
	if r != nil && r.layoutConfig.Section != nil {
		if layout := r.layoutConfig.Section(r, route); layout != nil {
			return layout
		}
	}
	if route != nil && route.Path != "/" && r.isBlogPrefix(route.Path) {
		return uiapp.NewLayout("blog" + r.layoutLocaleSuffix(route)).WithSidebar(&blogSidebar{router: r, prefix: route.Path})
	}
	return uiapp.NewLayout(sectionLayoutName(route) + r.layoutLocaleSuffix(route)).WithSidebar(&docsSidebar{router: r})
}

// layoutLocaleSuffix keys a section layout by the language of the pages it
// wraps, once a site has more than one.
//
// GoFastr re-renders a layout layer on client-side navigation only when its
// key changes. The docs sidebar lives in the section layer, so a Spanish page
// reached from an English one by a content link kept the English sidebar
// around a Spanish article. A single-language site keeps its layout names,
// and the CSS classes derived from them.
func (r *Router) layoutLocaleSuffix(route *Route) string {
	if r == nil || !r.multilingual() {
		return ""
	}
	if locale := r.effectiveLocale(route); locale != "" {
		return "-" + locale
	}
	return ""
}

// multilingual reports whether published pages are in more than one language,
// counting the fallback locale an unmarked page is treated as being in.
// Locales() lists declared values only, and a site that marks its
// translations but not its originals declares exactly one.
func (r *Router) multilingual() bool {
	seen := map[string]bool{}
	for _, route := range r.Routes() {
		if !r.variantPublished(route) {
			continue
		}
		if locale := r.effectiveLocale(route); locale != "" {
			seen[locale] = true
		}
	}
	return len(seen) > 1
}

func sectionLayoutName(route *Route) string {
	if route == nil || route.Path == "/" {
		return "docs-home"
	}
	if route.Path == "/api-reference" {
		return "docs-api"
	}
	return "docs-section"
}

// blogRootFor returns the blog collection a route belongs to, or nil.
func (r *Router) blogRootFor(route *Route) *Route {
	for current := route; current != nil; current = current.Parent {
		if r.isBlogPrefix(current.Path) {
			return current
		}
	}
	return nil
}

func (r *Router) rootRoute(route *Route) *Route {
	for route != nil && route.Parent != nil {
		route = route.Parent
	}
	return route
}

type routerMounter struct{ router *gofastrRouter.Router }

func (m routerMounter) MountWidget(def *widget.Definition) { widget.Mount(m.router, def) }

// Group scopes page and screen registration beneath a group path.
type Group struct {
	router *Router
	route  *Route
}

// Page adds a page beneath the group.
func (g *Group) Page(path string, cfg PageConfig) error {
	return g.router.Page(joinPath(g.route.Path, path), cfg)
}

// MustPage adds a page beneath the group and panics on invalid config.
func (g *Group) MustPage(path string, cfg PageConfig) {
	g.router.MustPage(joinPath(g.route.Path, path), cfg)
}

// PageFile adds a Markdown file beneath the group.
func (g *Group) PageFile(path string, cfg PageConfig) error {
	return g.router.PageFile(joinPath(g.route.Path, path), cfg)
}

// MustPageFile adds a Markdown file beneath the group and panics on error.
func (g *Group) MustPageFile(path string, cfg PageConfig) {
	g.router.MustPageFile(joinPath(g.route.Path, path), cfg)
}

// Screen adds a typed screen beneath the group.
func (g *Group) Screen(path string, cfg ScreenConfig) error {
	return g.router.Screen(joinPath(g.route.Path, path), cfg)
}

// MustScreen adds a typed screen beneath the group and panics on invalid config.
func (g *Group) MustScreen(path string, cfg ScreenConfig) {
	g.router.MustScreen(joinPath(g.route.Path, path), cfg)
}

// Route returns the group node.
func (g *Group) Route() *Route { return g.route }

type pageComponent struct {
	router *Router
	route  *Route
}

type screenMetadataComponent struct {
	router    *Router
	component component.Component
	route     *Route
}

func (s *screenMetadataComponent) Render() render.HTML {
	return render.Join(s.component.Render(), s.router.chromeTemplate(context.Background(), s.route.Path))
}

func (s *screenMetadataComponent) RenderCtx(ctx context.Context) render.HTML {
	return render.Join(component.RenderComponentCtx(ctx, s.component), s.router.chromeTemplate(ctx, s.route.Path))
}

func (s *screenMetadataComponent) Actions() {
	if interactive, ok := s.component.(component.InteractiveComponent); ok {
		interactive.Actions()
	}
}

// ComponentID preserves explicit action component IDs across the metadata
// wrapper. An empty result falls back to GoFastr's route-derived ID.
func (s *screenMetadataComponent) ComponentID() string {
	if identified, ok := s.component.(uiapp.ScreenComponentID); ok {
		if id := strings.TrimSpace(identified.ComponentID()); id != "" {
			return id
		}
	}
	path := strings.Trim(strings.TrimSpace(s.route.Path), "/")
	if path == "" {
		return "home"
	}
	return strings.NewReplacer("/", "-", ":", "").Replace(path)
}

// SetParams preserves dynamic route parameters when metadata requires a
// wrapper around a GoFastr screen.
func (s *screenMetadataComponent) ScreenTitle() string { return s.route.Title }

func (s *screenMetadataComponent) ScreenDescription() string { return s.route.Description }

func (s *screenMetadataComponent) ScreenType() uiapp.ScreenType {
	if typer, ok := s.component.(uiapp.ScreenTyper); ok {
		return typer.ScreenType()
	}
	return uiapp.ScreenPage
}

func (s *screenMetadataComponent) ScreenArticle() uiapp.ArticleMeta {
	meta := s.route.Metadata
	return uiapp.ArticleMeta{
		Headline: s.route.Title, Description: s.route.Description,
		Author: strings.Join(meta.Authors, ", "), DatePublished: meta.DatePublished,
		DateModified: meta.DateModified, Image: meta.Image,
	}
}

func (s *screenMetadataComponent) HeadHTML() string {
	custom := ""
	if seo, ok := s.component.(interface{ HeadHTML() string }); ok {
		custom = seo.HeadHTML()
	}
	return custom + s.router.metadataHeadHTML(s.route)
}

func (s *screenMetadataComponent) SetParams(params map[string]string) {
	if setter, ok := s.component.(uiapp.ParamSetter); ok {
		setter.SetParams(params)
	}
}

func metadataScreenComponent(r *Router, route *Route, original component.Component) component.Component {
	base := &screenMetadataComponent{router: r, component: original, route: route}
	loader, hasLoader := original.(uiapp.ScreenLoader)
	provider, hasStaticPaths := original.(uiapp.StaticPathsProvider)
	switch {
	case hasLoader && hasStaticPaths:
		return &metadataScreenWithLoaderAndStaticPaths{screenMetadataComponent: base, loader: loader, provider: provider}
	case hasLoader:
		return &metadataScreenWithLoader{screenMetadataComponent: base, loader: loader}
	case hasStaticPaths:
		return &metadataScreenWithStaticPaths{screenMetadataComponent: base, provider: provider}
	default:
		return base
	}
}

type metadataScreenWithLoader struct {
	*screenMetadataComponent
	loader uiapp.ScreenLoader
}

func (s *metadataScreenWithLoader) Load(ctx context.Context) error { return s.loader.Load(ctx) }

type metadataScreenWithStaticPaths struct {
	*screenMetadataComponent
	provider uiapp.StaticPathsProvider
}

func (s *metadataScreenWithStaticPaths) StaticPaths(ctx context.Context) []map[string]string {
	return s.provider.StaticPaths(ctx)
}

type metadataScreenWithLoaderAndStaticPaths struct {
	*screenMetadataComponent
	loader   uiapp.ScreenLoader
	provider uiapp.StaticPathsProvider
}

func (s *metadataScreenWithLoaderAndStaticPaths) Load(ctx context.Context) error {
	return s.loader.Load(ctx)
}

func (s *metadataScreenWithLoaderAndStaticPaths) StaticPaths(ctx context.Context) []map[string]string {
	return s.provider.StaticPaths(ctx)
}

func (p *pageComponent) Render() render.HTML {
	return p.render(context.Background())
}

func (p *pageComponent) RenderCtx(ctx context.Context) render.HTML {
	return p.render(ctx)
}

func (p *pageComponent) render(ctx context.Context) render.HTML {
	return render.Join(p.content(ctx), p.router.chromeTemplate(ctx, p.route.Path))
}

func (p *pageComponent) content(ctx context.Context) render.HTML {
	if p.route.page == nil {
		return render.Text("")
	}
	if p.route.Blog {
		if p.route.page.ContextBody != nil {
			return p.route.page.ContextBody(ctx)
		}
		if p.route.page.Body != nil {
			return p.route.page.Body()
		}
	}
	if p.route.page.Body != nil {
		return p.router.wrapDocPage(p.route, p.route.page.Body(), nil)
	}
	source := pageSource(p.route.page)
	if source == "" {
		panic(fmt.Sprintf("docs: page %q has no Markdown source", p.route.Path))
	}
	blogPost := p.route.Blog && !p.route.BlogIndex
	if blogPost {
		source = stripLeadingMarkdownTitle(source)
	}
	markdown := renderDocsMarkdown(source, map[string]string{
		"data-docs-route": p.route.Path,
		"data-offline":    fmt.Sprintf("%t", p.route.Offline),
	})
	vocab := p.router.vocabulary(p.route.page.Components, p.route.page.Containers)
	if !vocab.empty() {
		var err error
		markdown, err = renderMarkdownWithComponents(source, vocab, map[string]string{
			"data-docs-route": p.route.Path,
			"data-offline":    fmt.Sprintf("%t", p.route.Offline),
		})
		if err != nil {
			panic(fmt.Sprintf("docs: render Markdown components for %q: %v", p.route.Path, err))
		}
	}
	markdown = render.HTML(headingAnchorButtons(dedupeMarkdownHeadingIDs(string(markdown))))
	if blogPost {
		return p.router.wrapBlogPost(p.route, markdown, source)
	}
	if hero := renderPageHero(p.route.Metadata.Hero); hero != "" {
		markdown = render.Join(hero, markdown)
	}
	var headings []Heading
	if !p.route.page.DisableTOC && p.route.Metadata.PageTemplate != PageTemplateSplash {
		headings = markdownHeadings(source)
	}
	return p.router.wrapDocPage(p.route, markdown, headings)
}

func pageSource(page *PageConfig) string {
	if page == nil {
		return ""
	}
	if page.Source != "" {
		document, err := ParseMarkdown(page.Source)
		if err != nil {
			return page.Source
		}
		return document.Body
	}
	if page.SourcePath != "" {
		body, err := os.ReadFile(page.SourcePath)
		if err != nil {
			return ""
		}
		document, err := ParseMarkdown(string(body))
		if err != nil {
			return string(body)
		}
		return document.Body
	}
	return ""
}

func (p *pageComponent) ScreenTitle() string       { return p.route.Title }
func (p *pageComponent) ScreenDescription() string { return p.route.Description }

// ScreenArticle exposes Markdown metadata to GoFastr's Reader Mode and SEO
// synthesis without making content authors implement a second interface.
func (p *pageComponent) ScreenArticle() uiapp.ArticleMeta {
	meta := p.route.Metadata
	author := strings.Join(meta.Authors, ", ")
	return uiapp.ArticleMeta{
		Headline: p.route.Title, Description: p.route.Description, Author: author,
		DatePublished: meta.DatePublished, DateModified: meta.DateModified,
		Image: meta.Image,
	}
}

// HeadHTML contributes page-local SEO metadata that is not synthesized by
// GoFastr's ScreenArticle integration. GoFastr owns the description, Open
// Graph, article wrapper, and JSON-LD derived from ScreenArticle; this hook
// supplies robots, canonical/alternate links, Twitter, and article fields.
func (p *pageComponent) HeadHTML() string {
	return p.router.metadataHeadHTML(p.route)
}

func (r *Router) metadataHeadHTML(route *Route) string {
	if route == nil {
		return ""
	}
	meta := route.Metadata
	alternates := r.alternatesFor(route)
	var tags []string
	title := route.Title
	description := route.Description
	canonical := safeMetadataURL(meta.CanonicalURL)
	if canonical != "" {
		tags = append(tags, `<link rel="canonical" href="`+stdhtml.EscapeString(canonical)+`">`)
	}
	if meta.NoIndex || meta.Draft || route.Hidden {
		tags = append(tags, `<meta name="robots" content="noindex,nofollow">`)
	}
	if title != "" {
		tags = append(tags, `<meta name="twitter:title" content="`+stdhtml.EscapeString(title)+`">`)
	}
	if description != "" {
		tags = append(tags, `<meta name="twitter:description" content="`+stdhtml.EscapeString(description)+`">`)
	}
	if title != "" || description != "" || meta.Image != "" {
		card := "summary"
		if safeMetadataURL(meta.Image) != "" {
			card = "summary_large_image"
		}
		tags = append(tags, `<meta name="twitter:card" content="`+card+`">`)
	}
	if image := safeMetadataURL(meta.Image); image != "" {
		tags = append(tags, `<meta name="twitter:image" content="`+stdhtml.EscapeString(image)+`">`)
	}
	for _, author := range meta.Authors {
		if author = strings.TrimSpace(author); author != "" {
			tags = append(tags, `<meta name="author" content="`+stdhtml.EscapeString(author)+`">`)
		}
	}
	for _, keyword := range meta.Tags {
		if keyword = strings.TrimSpace(keyword); keyword != "" {
			tags = append(tags, `<meta property="article:tag" content="`+stdhtml.EscapeString(keyword)+`">`)
		}
	}
	if meta.DatePublished != "" {
		tags = append(tags, `<meta property="article:published_time" content="`+stdhtml.EscapeString(meta.DatePublished)+`">`)
	}
	if meta.DateModified != "" {
		tags = append(tags, `<meta property="article:modified_time" content="`+stdhtml.EscapeString(meta.DateModified)+`">`)
	}
	locales := make([]string, 0, len(alternates))
	for locale := range alternates {
		locales = append(locales, locale)
	}
	sort.Strings(locales)
	for _, locale := range locales {
		href := alternates[locale]
		if cleanHref := safeMetadataURL(href); cleanHref != "" && strings.TrimSpace(locale) != "" {
			tags = append(tags, `<link rel="alternate" hreflang="`+stdhtml.EscapeString(locale)+`" href="`+stdhtml.EscapeString(cleanHref)+`">`)
		}
	}
	return strings.Join(tags, "")
}

func (r *Router) routeHasMetadata(route *Route) bool {
	if route == nil {
		return false
	}
	meta := route.Metadata
	return meta.Draft || meta.NoIndex || meta.EditURL != "" || meta.CanonicalURL != "" || meta.Image != "" || len(meta.Authors) > 0 || meta.DatePublished != "" || meta.DateModified != "" || meta.Locale != "" || meta.Version != "" || len(r.alternatesFor(route)) > 0
}

func safeMetadataURL(raw string) string {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return ""
	}
	parsed, err := url.Parse(raw)
	if err != nil || parsed.User != nil || strings.ContainsAny(raw, "<>\"'") {
		return ""
	}
	if parsed.IsAbs() && parsed.Scheme != "http" && parsed.Scheme != "https" {
		return ""
	}
	if parsed.IsAbs() || strings.HasPrefix(raw, "/") {
		return raw
	}
	return ""
}

func safeRedirectPath(raw string) string {
	raw = strings.TrimSpace(raw)
	if raw == "" || strings.HasPrefix(raw, "//") || strings.ContainsAny(raw, "?#") {
		return ""
	}
	path := normalizePath(raw)
	if path == "" || !strings.HasPrefix(path, "/") {
		return ""
	}
	return path
}

func (r *Router) addRoute(path string, route Route) (*Route, error) {
	path = normalizePath(path)
	// Slug renames the last segment wherever it is set, not only inside
	// collections: the field's contract is "overrides the last segment of
	// the route path", and a standalone page honoring it only some of the
	// time is how two pages silently share one intended address.
	// Blogs resolve their own slugs while building post paths, so the
	// rewrite below must not move a post after the blog has recorded where
	// it expects to find it.
	underBlog := false
	for _, prefix := range r.blogPrefixesList() {
		if pathActive(prefix, path) {
			underBlog = true
			break
		}
	}
	if route.Kind != KindGroup && !underBlog && route.Metadata.Slug != "" && path != "/" &&
		!strings.Contains(route.Metadata.Slug, "/") && pathpkg.Base(path) != route.Metadata.Slug {
		slug, err := collectionSlug(route.Metadata.Slug)
		if err != nil {
			return nil, fmt.Errorf("docs: route %q: %w", path, err)
		}
		path = normalizePath(pathpkg.Join(pathpkg.Dir(path), slug))
	}
	if path == "" {
		return nil, errors.New("docs: route path is required")
	}
	if strings.HasPrefix(path, "//") {
		return nil, fmt.Errorf("docs: route path %q must use one leading slash", path)
	}
	if strings.ContainsAny(path, "?#") {
		return nil, fmt.Errorf("docs: route path %q must not contain query or fragment data", path)
	}
	if _, exists := r.routes[path]; exists {
		err := fmt.Errorf("docs: duplicate route %q", path)
		r.registrationErr = err
		return nil, err
	}
	r.seq++
	route.Path = path
	route.ID = routeID(path)
	// Locale and Version are compared as lowercased slugs everywhere
	// downstream (pairing, search, the manifest); storing them raw lets a
	// mixed-case front matter value fork one language into two.
	route.Metadata.Locale = normalizeLocale(route.Metadata.Locale)
	route.Metadata.Version = strings.TrimSpace(route.Metadata.Version)
	route.seq = r.seq
	if route.Title == "" && route.Kind == KindPage && route.page != nil {
		route.Title = titleFromPath(path)
	}
	r.routes[path] = &route
	r.rebuildTree()
	return &route, nil
}

func (r *Router) rebuildTree() {
	// The locale family index describes the tree, so it cannot outlive it.
	r.localeFamilies = nil
	r.roots = nil
	for _, route := range r.routes {
		route.Parent = nil
		route.Children = nil
	}
	ordered := make([]*Route, 0, len(r.routes))
	for _, route := range r.routes {
		ordered = append(ordered, route)
	}
	sort.SliceStable(ordered, func(i, j int) bool { return ordered[i].seq < ordered[j].seq })
	for _, route := range ordered {
		parent := r.parentFor(route.Path)
		if parent != nil {
			route.Parent = parent
			parent.Children = append(parent.Children, route)
		} else {
			r.roots = append(r.roots, route)
		}
	}
}

func (r *Router) parentFor(path string) *Route {
	best := ""
	var parent *Route
	for candidate, route := range r.routes {
		if candidate == path || !strings.HasPrefix(path, candidate+"/") {
			continue
		}
		if len(candidate) > len(best) {
			best, parent = candidate, route
		}
	}
	return parent
}

func (r *Router) sorted(nodes []*Route) []*Route {
	out := append([]*Route(nil), nodes...)
	sort.SliceStable(out, func(i, j int) bool {
		if out[i].Order != out[j].Order {
			return out[i].Order < out[j].Order
		}
		return out[i].seq < out[j].seq
	})
	return out
}

func normalizePath(path string) string {
	path = strings.TrimSpace(path)
	if path == "" {
		return ""
	}
	if !strings.HasPrefix(path, "/") {
		path = "/" + path
	}
	path = strings.TrimRight(path, "/")
	if path == "" {
		return "/"
	}
	return path
}

func validatePreload(mode string) error {
	switch strings.TrimSpace(mode) {
	case "", uiapp.PreloadHover, uiapp.PreloadVisible, uiapp.PreloadEager:
		return nil
	default:
		return fmt.Errorf("invalid preload mode %q (use %q, %q, or %q)", mode, uiapp.PreloadHover, uiapp.PreloadVisible, uiapp.PreloadEager)
	}
}

func joinPath(prefix, path string) string {
	if strings.HasPrefix(path, "/") {
		return normalizePath(path)
	}
	return normalizePath(strings.TrimRight(prefix, "/") + "/" + path)
}

func routeID(path string) string {
	if path == "/" {
		return "home"
	}
	// Route IDs are consumed by search indexes and client adapters, so they
	// must be deterministic and collision-free. A readable slug would make
	// `/foo-bar` and `/foo/bar` (or `/home` and `/`) collide. Keep the root
	// friendly and encode every other path as a compact, URL-safe byte ID.
	return "route-" + hex.EncodeToString([]byte(path))
}

func titleFromPath(path string) string {
	part := path[strings.LastIndex(path, "/")+1:]
	part = strings.ReplaceAll(part, "-", " ")
	if part == "" {
		return "Home"
	}
	return strings.ToUpper(part[:1]) + part[1:]
}

func cloneStrings(in []string) []string { return append([]string(nil), in...) }

func (r *Router) routePublished(route *Route) bool {
	if route == nil || route.Kind == KindGroup || route.Hidden {
		return false
	}
	if route.Metadata.Draft && !r.includeDrafts && !route.includeDrafts {
		return false
	}
	if !r.localeAllows(route) {
		return false
	}
	if r.version != "" && route.Metadata.Version != "" && route.Metadata.Version != r.version {
		return false
	}
	return true
}

func (r *Router) routeVisible(route *Route) bool {
	if route == nil || route.Hidden {
		return false
	}
	if route.Kind != KindGroup {
		return r.routePublished(route)
	}
	for _, child := range route.Children {
		if r.routeVisible(child) {
			return true
		}
	}
	return false
}

func pathActive(routePath, currentPath string) bool {
	if currentPath == "" {
		return false
	}
	if routePath == "/" {
		return currentPath == "/"
	}
	return currentPath == routePath || strings.HasPrefix(currentPath, routePath+"/")
}
