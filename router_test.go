package docs

import (
	"context"
	"encoding/json"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	uiapp "github.com/DonaldMurillo/gofastr/core-ui/app"
	"github.com/DonaldMurillo/gofastr/core/render"
	gofastrRouter "github.com/DonaldMurillo/gofastr/core/router"
)

type testScreen struct{}

func (*testScreen) Render() render.HTML { return render.Raw("<section>screen</section>") }

func TestRouterBuildsOrderedNestedNavigationAndSearch(t *testing.T) {
	r := NewRouter(WithSiteName("Acme"))
	r.MustPage("/", PageConfig{Title: "Home", Description: "Welcome", Source: "# Welcome", Order: 1, Tags: []string{"overview"}})
	group := r.MustGroup("/docs", GroupConfig{Title: "Docs", Description: "Guides", Order: 2})
	group.MustPage("getting-started", PageConfig{Title: "Getting started", Description: "Start here", Source: "# Start here", Order: 2})
	group.MustScreen("playground", ScreenConfig{Title: "Playground", Description: "Interactive", Component: &testScreen{}, Order: 1})

	if err := r.Validate(); err != nil {
		t.Fatalf("Validate() error = %v", err)
	}
	nav := r.Navigation()
	if len(nav) != 4 {
		t.Fatalf("Navigation() length = %d, want 4", len(nav))
	}
	if got := nav[2].Path; got != "/docs/playground" {
		t.Fatalf("first nested nav path = %q, want /docs/playground", got)
	}
	if got := nav[3].Path; got != "/docs/getting-started" {
		t.Fatalf("second nested nav path = %q, want /docs/getting-started", got)
	}
	active := r.NavigationAt("/docs/getting-started")
	if !active[1].Active || !active[3].Active || active[0].Active {
		t.Fatalf("NavigationAt() active state = %#v", active)
	}
	if got := len(r.SearchIndex()); got != 3 {
		t.Fatalf("SearchIndex() length = %d, want 3", got)
	}
	if got := r.SiteName(); got != "Acme" {
		t.Fatalf("SiteName() = %q, want Acme", got)
	}
	bodyPage := NewRouter()
	bodyPage.MustPage("/interactive", PageConfig{Title: "Interactive", Description: "A typed docs page", Body: func() render.HTML { return render.Raw("<div>interactive</div>") }, SearchText: "interactive playground", Order: 1})
	if got := bodyPage.SearchIndex()[0].Text; got != "interactive playground" {
		t.Fatalf("Body page search text = %q", got)
	}
	results := r.Search("playground interactive")
	if len(results) != 1 || results[0].Entry.Path != "/docs/playground" || results[0].Score < 10 {
		t.Fatalf("ranked Search() results = %#v", results)
	}
}

func TestSearchProviderCanReplacePortableArtifact(t *testing.T) {
	type provider struct{}
	funcBuild := func(*Router) ([]byte, error) { return []byte(`{"custom":true}`), nil }
	_ = provider{}
	r := NewRouter(WithSearchProvider(searchProviderFunc(funcBuild)))
	r.MustPage("/guide", PageConfig{Title: "Guide", Description: "Guide", Source: "# Guide", Order: 1})
	body, err := r.SearchIndexJSON()
	if err != nil {
		t.Fatalf("SearchIndexJSON() error = %v", err)
	}
	if string(body) != `{"custom":true}` {
		t.Fatalf("custom search artifact = %s", body)
	}
}

func TestExportManifestContainsPublishedRoutesAndRedirects(t *testing.T) {
	r := NewRouter()
	r.MustPage("/guide", PageConfig{
		Title: "Guide", Description: "Guide", Source: "# Guide", Order: 1,
		Metadata: ContentMetadata{Redirects: []string{"/old-guide"}, Tags: []string{"guide"}},
	})
	r.MustPage("/private", PageConfig{
		Title: "Private", Description: "Private", Source: "# Private", Order: 2,
		Metadata: ContentMetadata{Draft: true},
	})
	body, err := r.ExportManifestJSON("/docs")
	if err != nil {
		t.Fatalf("ExportManifestJSON() error = %v", err)
	}
	var manifest ExportManifest
	if err := json.Unmarshal(body, &manifest); err != nil {
		t.Fatal(err)
	}
	if manifest.Schema != "fastr-docs/v1" || manifest.SearchIndex != "/docs/__fastr-docs/search.json" || manifest.SearchBackend != SearchBackendJSON {
		t.Fatalf("manifest header = %#v", manifest)
	}
	if len(manifest.Routes) != 1 || manifest.Routes[0].Path != "/docs/guide" {
		t.Fatalf("manifest routes = %#v", manifest.Routes)
	}
	if len(manifest.Redirects) != 1 || manifest.Redirects[0].From != "/docs/old-guide" || manifest.Redirects[0].To != "/docs/guide" {
		t.Fatalf("manifest redirects = %#v", manifest.Redirects)
	}
}

type searchProviderFunc func(*Router) ([]byte, error)

func (f searchProviderFunc) Build(r *Router) ([]byte, error) { return f(r) }

func TestSidebarDoesNotDuplicateAParentPageWithChildren(t *testing.T) {
	r := NewRouter()
	r.MustPage("/docs", PageConfig{Title: "Documentation", Description: "Docs", Source: "# Docs", Order: 1})
	r.MustPage("/docs/getting-started", PageConfig{Title: "Getting started", Description: "Start here", Source: "# Start here", Order: 1})

	items := r.sidebarItems(r.roots, "")
	if len(items) != 1 || items[0].Label != "Documentation" {
		t.Fatalf("sidebar roots = %#v, want one Documentation disclosure", items)
	}
	if len(items[0].Children) != 1 || items[0].Children[0].Label != "Getting started" {
		t.Fatalf("Documentation children = %#v, want only Getting started", items[0].Children)
	}
	if items[0].Children[0].Href != "/docs/getting-started" {
		t.Fatalf("nested child href = %q, want /docs/getting-started", items[0].Children[0].Href)
	}
}

func TestSidebarOpensTheActiveNestedRoute(t *testing.T) {
	r := NewRouter()
	r.MustPage("/docs", PageConfig{Title: "Documentation", Description: "Docs", Source: "# Docs", Order: 1})
	r.MustPage("/docs/getting-started", PageConfig{Title: "Getting started", Description: "Start here", Source: "# Start here", Order: 1})
	req := httptest.NewRequest("GET", "/docs/getting-started", nil)
	html := (&docsSidebar{router: r}).RenderCtx(uiapp.WithRequest(context.Background(), req))
	if !strings.Contains(string(html), `<details class="ui-sidebar__group"`) || !strings.Contains(string(html), " open>") {
		t.Fatalf("active nested route did not open its parent disclosure: %s", html)
	}
}

func TestSidebarShowsTopLevelRoutesOnHome(t *testing.T) {
	r := NewRouter()
	r.MustPage("/", PageConfig{Title: "Home", Description: "Home", Source: "# Home", Order: 1})
	r.MustPage("/guide", PageConfig{Title: "Guide", Description: "Guide", Source: "# Guide", Order: 2})
	r.MustScreen("/lab", ScreenConfig{Title: "Lab", Description: "Lab", Component: &testScreen{}, Order: 3})

	req := httptest.NewRequest("GET", "/", nil)
	html := (&docsSidebar{router: r}).RenderCtx(uiapp.WithRequest(context.Background(), req))
	if !strings.Contains(string(html), `href="/guide"`) || !strings.Contains(string(html), `href="/lab"`) {
		t.Fatalf("home sidebar omitted top-level routes: %s", html)
	}
}

func TestSidebarRendersRouteBadgesWithoutChangingNavigation(t *testing.T) {
	r := NewRouter()
	guides := r.MustGroup("/guides", GroupConfig{
		Title: "Guides", Description: "Guides", Order: 1,
		Badge: NavBadge{Label: "Popular", Tone: NavBadgeToneInfo},
	})
	guides.MustPage("patterns", PageConfig{
		Title: "Patterns", Description: "Patterns", Source: "# Patterns", Order: 1,
		Badge: NavBadge{Label: "New"},
	})

	items := r.Navigation()
	if len(items) != 2 || items[0].Badge.Label != "Popular" || items[1].Badge.Label != "New" {
		t.Fatalf("navigation badges = %#v, want group and page badges", items)
	}

	html := (&docsSidebar{router: r}).Render()
	for _, marker := range []string{
		`fastr-docs-nav-badge--info`,
		`fastr-docs-nav-badge--accent`,
		`data-badge-label="Popular"`,
		`data-badge-label="New"`,
	} {
		if !strings.Contains(string(html), marker) {
			t.Fatalf("sidebar missing badge marker %q: %s", marker, html)
		}
	}
}

func TestNavigationBadgeRejectsUnknownTone(t *testing.T) {
	r := NewRouter()
	if err := r.Page("/guide", PageConfig{
		Title: "Guide", Description: "Guide", Source: "# Guide", Order: 1,
		Badge: NavBadge{Label: "New", Tone: NavBadgeTone("electric")},
	}); err == nil || !strings.Contains(err.Error(), "invalid navigation badge tone") {
		t.Fatalf("invalid badge tone error = %v", err)
	}
}

func TestRouterRejectsBrokenRegistrationsInStrictMode(t *testing.T) {
	r := NewRouter()
	if err := r.Page("/missing", PageConfig{Title: "Missing"}); err != nil {
		t.Fatalf("Page() unexpectedly rejected registration: %v", err)
	}
	if err := r.Page("/missing", PageConfig{Title: "Duplicate", Description: "Duplicate", Source: "body", Order: 1}); err == nil {
		t.Fatal("duplicate Page() error = nil")
	}
	if err := r.Screen("/screen", ScreenConfig{Title: "Screen", Order: 3}); err != nil {
		t.Fatalf("Screen() unexpectedly rejected registration: %v", err)
	}
	if err := r.Page("/page", PageConfig{Title: "Page", Description: "Page", Source: "body", Order: 3}); err != nil {
		t.Fatalf("Page() unexpectedly rejected registration: %v", err)
	}
	if err := r.Validate(); err == nil || !strings.Contains(err.Error(), "Component is required") || !strings.Contains(err.Error(), "duplicate explicit order") {
		t.Fatalf("Validate() error = %v, want component and order diagnostics", err)
	}
}

func TestRouterRejectsAmbiguousOrExternalRoutePaths(t *testing.T) {
	r := NewRouter()
	for _, path := range []string{"//external.example", "/docs?draft=1", "/docs#setup"} {
		if err := r.Page(path, PageConfig{Title: "Invalid", Description: "Invalid", Source: "# Invalid", Order: 1}); err == nil {
			t.Fatalf("Page(%q) accepted an unsafe/ambiguous route path", path)
		}
	}
}

func TestRouteIDsRemainUniqueForSimilarPaths(t *testing.T) {
	r := NewRouter()
	r.MustPage("/", PageConfig{Title: "Home", Description: "Home", Source: "# Home", Order: 1})
	r.MustPage("/home", PageConfig{Title: "Home route", Description: "Home route", Source: "# Home route", Order: 2})
	r.MustPage("/foo-bar", PageConfig{Title: "Hyphen", Description: "Hyphen", Source: "# Hyphen", Order: 3})
	r.MustPage("/foo/bar", PageConfig{Title: "Nested", Description: "Nested", Source: "# Nested", Order: 4})
	ids := map[string]bool{}
	for _, route := range r.Routes() {
		if ids[route.ID] {
			t.Fatalf("duplicate route ID %q", route.ID)
		}
		ids[route.ID] = true
	}
}

func TestStrictValidationCanBeExplicitlyOptedOut(t *testing.T) {
	r := NewRouter(WithStrictValidation(false))
	r.MustPage("/draft", PageConfig{Title: ""})
	if err := r.Validate(); err != nil {
		t.Fatalf("opted-out validation error = %v", err)
	}
}

func TestOpenAPIPluginContributesToTheSameRouter(t *testing.T) {
	spec := []byte(`{"openapi":"3.1.0","info":{"title":"Public API","description":"The contract."},"paths":{"/v1/projects":{"get":{"summary":"List projects","operationId":"listProjects"}}}}`)
	r := NewRouter()
	if err := r.Use(OpenAPIPlugin{Spec: spec}); err != nil {
		t.Fatalf("Use(OpenAPIPlugin) error = %v", err)
	}
	route := r.Routes()[0]
	if route.Path != "/api-reference" || route.Kind != KindPlugin || route.Plugin != "openapi" {
		t.Fatalf("plugin route = %#v", route)
	}
	entries := r.SearchIndex()
	if len(entries) != 1 || !strings.Contains(entries[0].Text, "listProjects") {
		t.Fatalf("plugin search entry = %#v", entries)
	}
}

func TestPageFileFeedsValidationAndSearch(t *testing.T) {
	file := t.TempDir() + string(os.PathSeparator) + "guide.md"
	if err := os.WriteFile(file, []byte("# File-backed guide\n\nRouter source."), 0o644); err != nil {
		t.Fatal(err)
	}
	r := NewRouter()
	if err := r.PageFile("/guide", PageConfig{Title: "Guide", Description: "A guide", SourcePath: file, Order: 1}); err != nil {
		t.Fatalf("PageFile() error = %v", err)
	}
	if err := r.Validate(); err != nil {
		t.Fatalf("Validate() error = %v", err)
	}
	if got := r.SearchIndex()[0].Text; !strings.Contains(got, "Router source") {
		t.Fatalf("SearchIndex text = %q", got)
	}
}

func TestMountRegistersPagesAndTypedScreensWithGoFastr(t *testing.T) {
	r := NewRouter()
	r.MustPage("/", PageConfig{Title: "Home", Description: "Home", Source: "# Home\n\n## Setup\n\nRead this guide.", Order: 1})
	r.MustScreen("/playground", ScreenConfig{Title: "Playground", Description: "Interactive", Component: &testScreen{}, Order: 2})
	site := uiapp.NewApp("Test docs")
	if err := r.Mount(site, r.Layout()); err != nil {
		t.Fatalf("Mount() error = %v", err)
	}
	if len(site.Routes()) != 2 {
		t.Fatalf("GoFastr route count = %d, want 2", len(site.Routes()))
	}
	if _, _, ok := site.Router.Resolve("/playground"); !ok {
		t.Fatal("typed screen was not registered")
	}
	html, err := site.RenderPage(context.Background(), "/")
	if err != nil {
		t.Fatalf("RenderPage() error = %v", err)
	}
	for _, marker := range []string{"ui-sidebar", "data-fui-open=\"fastr-docs-sections\"", "fastr-docs-brand", "data-fui-theme-toggle", "fastr-docs-command-trigger", "ui-doc-layout", "ui-anchored-rail", "data-fui-scrollspy", "data-docs-toc-select", "#setup", "Playground"} {
		if !strings.Contains(string(html), marker) {
			t.Fatalf("default docs layout missing %q: %s", marker, html)
		}
	}
	for _, marker := range []string{"Extensions", "Router extensions", "Agent-ready project", "offline shell ready"} {
		if strings.Contains(string(html), marker) {
			t.Fatalf("default white-label chrome leaked implementation metadata %q: %s", marker, html)
		}
	}
}

func TestDefaultThemeIncludesEditorialLightAndDarkTokens(t *testing.T) {
	css := DefaultTheme().CSSCustomProperties()
	for _, marker := range []string{
		"#f7f5ef",
		"#fffdfa",
		"#ec7131",
		`[data-color-scheme="dark"]`,
		"#131917",
		"#ffab74",
	} {
		if !strings.Contains(css, marker) {
			t.Fatalf("DefaultTheme() CSS missing %q", marker)
		}
	}
}

func TestContentSecurityPolicyOnlyAllowsDeclaredConnectOrigins(t *testing.T) {
	r := NewRouter()
	r.AllowConnectOrigin("https://api.example.com/v1")
	r.AllowConnectOrigin("https://api.example.com/other")
	r.AllowConnectOrigin("javascript:alert(1)")

	if got := r.ConnectOrigins(); len(got) != 1 || got[0] != "https://api.example.com" {
		t.Fatalf("ConnectOrigins() = %#v, want one normalized origin", got)
	}
	policy := ContentSecurityPolicy(r.ConnectOrigins()...)
	for _, marker := range []string{
		"default-src 'self'",
		"object-src 'none'",
		"frame-ancestors 'none'",
		"connect-src 'self' https://api.example.com",
	} {
		if !strings.Contains(policy, marker) {
			t.Fatalf("ContentSecurityPolicy() missing %q: %s", marker, policy)
		}
	}
	if strings.Contains(policy, "javascript:") || strings.Contains(policy, "https://api.example.com/v1") {
		t.Fatalf("ContentSecurityPolicy() leaked an unsafe or path-specific source: %s", policy)
	}
}

func TestResponsiveDocsCSSKeepsChromeAndTOCAligned(t *testing.T) {
	css := NewRouter().CSS()
	for _, marker := range []string{
		"ui-site-header__mobile",
		"fastr-docs-doc-layout",
		".layout-docs .ui-sidebar__inline { display: block;",
		".layout-docs .ui-sidebar__nav { margin-top: 8px; }",
		"> .scrollspy",
		"ui-anchored-rail",
		"position: sticky",
		"--nav-h",
		"body { min-width: 0;",
		"overflow-x: clip",
	} {
		if !strings.Contains(css, marker) {
			t.Fatalf("responsive docs CSS missing %q", marker)
		}
	}
}

func TestMarkdownTOCMatchesRenderedHeadingIDsAndDeduplicatesTargets(t *testing.T) {
	headings := markdownHeadings("# Guide\n\n## Setup\n\n## Setup\n\n```\n## Not a section\n```\n\n### API / clients")
	if len(headings) != 3 {
		t.Fatalf("markdownHeadings() length = %d, want 3: %#v", len(headings), headings)
	}
	if headings[0].ID != "setup" || headings[1].ID != "setup-2" || headings[2].ID != "api-clients" {
		t.Fatalf("markdown heading IDs = %#v", headings)
	}
	rendered := dedupeMarkdownHeadingIDs(`<div><h2 id="setup">Setup</h2><h2 id="setup">Setup</h2><h3 id="api-clients">API</h3></div>`)
	if !strings.Contains(rendered, `<h2 id="setup-2">`) {
		t.Fatalf("duplicate Markdown heading ID was not rewritten: %s", rendered)
	}
}

func TestMountNavigationRegistersTheFrameworkDrawer(t *testing.T) {
	r := NewRouter()
	r.MustPage("/", PageConfig{Title: "Home", Description: "Home", Source: "# Home", Order: 1})
	httpRouter := gofastrRouter.New()
	if err := r.MountNavigation(httpRouter); err != nil {
		t.Fatalf("MountNavigation() error = %v", err)
	}
	var found bool
	for _, route := range httpRouter.Routes() {
		if strings.Contains(route.Pattern, "fastr-docs-sections") {
			found = true
			break
		}
	}
	if !found {
		t.Fatal("framework Sidebar drawer route was not registered")
	}
}

func TestHeaderItemsRepresentTopLevelGroups(t *testing.T) {
	r := NewRouter()
	guides := r.MustGroup("/guides", GroupConfig{Title: "Guides", Description: "Guides", Order: 1})
	guides.MustPage("getting-started", PageConfig{Title: "Getting started", Description: "Start here", Source: "# Start", Order: 1})
	r.MustPage("/api-reference", PageConfig{Title: "API reference", Description: "API", Source: "# API", Order: 2})
	items := r.headerItems()
	if len(items) != 2 || items[0].Label != "Guides" || items[0].Href != "/guides/getting-started" || items[1].Label != "API reference" {
		t.Fatalf("headerItems() = %#v, want top-level group and route", items)
	}
}

func TestHeaderVariantSelectorsResolveLocaleAndVersionSiblings(t *testing.T) {
	r := NewRouter()
	for order, item := range []struct {
		path    string
		locale  string
		version string
	}{
		{path: "/docs/en/v1/guide", locale: "en", version: "v1"},
		{path: "/docs/fr/v1/guide", locale: "fr", version: "v1"},
		{path: "/docs/en/v2/guide", locale: "en", version: "v2"},
		{path: "/docs/fr/v2/guide", locale: "fr", version: "v2"},
	} {
		r.MustPage(item.path, PageConfig{
			Title: "Guide", Description: "A localized versioned guide", Source: "# Guide", Order: order + 1,
			Metadata: ContentMetadata{Locale: item.locale, Version: item.version},
		})
	}
	header := string((&docsHeader{router: r}).render("/docs/en/v1/guide"))
	for _, marker := range []string{
		`data-docs-variant-select="locale"`,
		`data-docs-variant-select="version"`,
		`value="/docs/fr/v1/guide"`,
		`value="/docs/en/v2/guide"`,
	} {
		if !strings.Contains(header, marker) {
			t.Fatalf("variant selector missing %q: %s", marker, header)
		}
	}
}

func TestHeaderVariantActiveScriptGroupsLocaleAndVersionSiblings(t *testing.T) {
	r := NewRouter()
	variants := r.MustGroup("/locales", GroupConfig{Title: "Localized guides", Description: "Variants", Order: 1})
	for order, item := range []struct {
		path    string
		locale  string
		version string
	}{
		{path: "pt-BR/v1/guide", locale: "pt-BR", version: "v1"},
		{path: "fr/v1/guide", locale: "fr", version: "v1"},
		{path: "pt-BR/v2/guide", locale: "pt-BR", version: "v2"},
		{path: "fr/v2/guide", locale: "fr", version: "v2"},
	} {
		variants.MustPage(item.path, PageConfig{
			Title: "Guide", Description: "Variant", Source: "# Guide", Order: order + 1,
			Metadata: ContentMetadata{Locale: item.locale, Version: item.version},
		})
	}
	header := string((&docsHeader{router: r}).render("/locales/fr/v2/guide"))
	for _, marker := range []string{
		`name="fastr-docs-active-families"`,
		`/locales/fr/v2/guide`,
		`/locales/pt-BR/v1/guide`,
	} {
		if !strings.Contains(header, marker) {
			t.Fatalf("variant active-state script missing %q: %s", marker, header)
		}
	}
	if strings.Contains(header, `<script data-fastr-docs-active-families>`) {
		t.Fatalf("variant active-state config must not emit an inline script under the framework CSP: %s", header)
	}
}

func TestLanguageDoesNotFilterRuntimeVariants(t *testing.T) {
	r := NewRouter(WithLanguage("pt-BR"))
	r.MustPage("/en/guide", PageConfig{Title: "Guide", Description: "English", Source: "# Guide", Order: 1, Metadata: ContentMetadata{Locale: "en"}})
	r.MustPage("/pt-BR/guide", PageConfig{Title: "Guide", Description: "Portuguese", Source: "# Guide", Order: 2, Metadata: ContentMetadata{Locale: "pt-BR"}})
	if r.Language() != "pt-BR" || len(r.PublishedRoutes()) != 2 {
		t.Fatalf("WithLanguage unexpectedly filtered routes: language=%q routes=%d", r.Language(), len(r.PublishedRoutes()))
	}
}

func TestLayoutUsesNativeCommandPaletteSearch(t *testing.T) {
	r := NewRouter(WithSiteName("Acme"))
	r.MustPage("/docs", PageConfig{Title: "Documentation", Description: "Docs", Source: "# Docs", Order: 1})
	header := string((&docsHeader{router: r}).Render())
	for _, marker := range []string{"fastr-docs-command-trigger", "data-fui-open=\"fastr-docs-command-palette\"", "data-fui-shortcut-click=\"Meta+K\"", "ui-shortcut-hint"} {
		if !strings.Contains(header, marker) {
			t.Fatalf("native command palette header missing %q: %s", marker, header)
		}
	}
	if strings.Contains(header, "fastr-docs-search") || strings.Contains(header, "ui-search-input") {
		t.Fatalf("legacy docs search leaked into header: %s", header)
	}
}

func TestPagefindSearchBackendAnnotatesNativePalette(t *testing.T) {
	r := NewRouter(WithPagefind())
	r.MustPage("/docs", PageConfig{Title: "Documentation", Description: "Docs", Source: "# Docs", Order: 1})
	header := string((&docsHeader{router: r}).Render())
	if r.SearchBackend() != SearchBackendPagefind || !strings.Contains(header, `data-fastr-docs-backend="pagefind"`) || !strings.Contains(header, `data-fastr-docs-pagefind-path="/pagefind/"`) {
		t.Fatalf("Pagefind backend was not exposed to the native palette: %s", header)
	}
	if got := string(RuntimeJS()); !strings.Contains(got, "pagefind.js") || !strings.Contains(got, "data-fastr-docs-backend") {
		t.Fatalf("Pagefind runtime integration is incomplete: %s", got)
	}
	manifest, err := r.ExportManifestJSON("")
	if err != nil || !strings.Contains(string(manifest), `"searchBackend": "pagefind"`) || !strings.Contains(string(manifest), `"searchPath": "/pagefind/"`) {
		t.Fatalf("Pagefind manifest metadata is incomplete: %s (%v)", manifest, err)
	}
}

func TestSearchBackendOptionsNormalizeAndFallback(t *testing.T) {
	r := NewRouter(WithPagefind(), WithSearchBackend("unsupported"), WithPagefindPath("assets/pagefind"))
	if r.SearchBackend() != SearchBackendJSON {
		t.Fatalf("invalid search backend did not fall back to JSON: %q", r.SearchBackend())
	}
	if r.PagefindPath() != "/assets/pagefind/" {
		t.Fatalf("Pagefind path was not normalized: %q", r.PagefindPath())
	}

	r = NewRouter(WithSearchBackend("PAGEFIND"), WithPagefindPath("/search/"))
	if r.SearchBackend() != SearchBackendPagefind || r.PagefindPath() != "/search/" {
		t.Fatalf("normalized Pagefind options were not retained: backend=%q path=%q", r.SearchBackend(), r.PagefindPath())
	}
}
