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

type lifecycleScreen struct{}

func (*lifecycleScreen) Render() render.HTML        { return render.Raw("<section>dynamic screen</section>") }
func (*lifecycleScreen) Load(context.Context) error { return nil }
func (*lifecycleScreen) StaticPaths(context.Context) []map[string]string {
	return []map[string]string{{"slug": "one"}}
}
func (*lifecycleScreen) ComponentID() string          { return "lifecycle-screen" }
func (*lifecycleScreen) ScreenTitle() string          { return "Lifecycle screen" }
func (*lifecycleScreen) ScreenDescription() string    { return "Lifecycle screen" }
func (*lifecycleScreen) ScreenType() uiapp.ScreenType { return uiapp.ScreenPage }
func (*lifecycleScreen) HeadHTML() string             { return `<meta name="x-screen" content="custom">` }

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

func TestRouterExposesLocalizedChromeLabels(t *testing.T) {
	r := NewRouter(WithSiteName("Acme"), WithUIStrings(UIStrings{
		Contents: "Contenu", Home: "Accueil", OnThisPage: "Sur cette page",
		Search: "Rechercher", SearchPlaceholder: "Rechercher la documentation…",
		OpenSearch: "Ouvrir la recherche", CloseSearch: "Fermer la recherche",
		OpenNavigation: "Ouvrir la navigation", EditPage: "Modifier cette page",
		LastUpdated: "Mis à jour", Published: "Publié", By: "Par",
		Language: "Langue", Version: "Version",
	}))
	r.MustPage("/", PageConfig{Title: "Home", Description: "Home", Source: "# Home", Order: 1})
	r.MustPage("/guide", PageConfig{
		Title: "Guide", Description: "Guide", Source: "# Guide\n\n## Setup",
		Order: 1, Metadata: ContentMetadata{DateModified: "2026-08-30", Authors: []string{"Ada"}, EditURL: "https://example.com/edit"},
	})
	header := string((&docsHeader{router: r}).Render())
	if !strings.Contains(header, `aria-label="Ouvrir la recherche"`) || !strings.Contains(header, "Rechercher") {
		t.Fatalf("localized search chrome missing: %s", header)
	}
	sidebar := string((&docsSidebar{router: r}).Render())
	if !strings.Contains(sidebar, ">Contenu<") || !strings.Contains(sidebar, ">Accueil<") {
		t.Fatalf("localized sidebar chrome missing: %s", sidebar)
	}
	page := r.Routes()[1]
	rendered := string((&pageComponent{router: r, route: page}).Render())
	for _, marker := range []string{"data-docs-page-meta", "Mis à jour", "2026-08-30", "Par Ada", "Modifier cette page"} {
		if !strings.Contains(rendered, marker) {
			t.Fatalf("localized page chrome missing %q: %s", marker, rendered)
		}
	}
}

func TestRouterPreservesGoFastrScreenCapabilitiesAndPreload(t *testing.T) {
	r := NewRouter()
	r.MustScreen("/items/:slug", ScreenConfig{
		Title: "Item", Description: "A dynamic item", Component: &lifecycleScreen{},
		Preload: "hover", Metadata: ContentMetadata{DateModified: "2026-08-30"}, Order: 1,
	})
	site := uiapp.NewApp("Docs")
	if err := r.Mount(site, r.Layout()); err != nil {
		t.Fatalf("Mount() error = %v", err)
	}
	screen, ok := site.Router.ScreenByPattern("/items/:slug")
	if !ok {
		t.Fatal("dynamic screen was not registered")
	}
	if screen.Preload != uiapp.PreloadHover {
		t.Fatalf("screen preload = %q, want %q", screen.Preload, uiapp.PreloadHover)
	}
	if _, ok := screen.Component.(uiapp.ScreenLoader); !ok {
		t.Fatal("metadata wrapper hid ScreenLoader")
	}
	provider, ok := screen.Component.(uiapp.StaticPathsProvider)
	if !ok || len(provider.StaticPaths(context.Background())) != 1 {
		t.Fatal("metadata wrapper hid StaticPathsProvider")
	}
	identified, ok := screen.Component.(uiapp.ScreenComponentID)
	if !ok || identified.ComponentID() != "lifecycle-screen" {
		t.Fatal("metadata wrapper hid ScreenComponentID")
	}
	if seo, ok := screen.Component.(interface{ HeadHTML() string }); !ok || !strings.Contains(seo.HeadHTML(), `x-screen`) {
		t.Fatal("metadata wrapper hid custom head HTML")
	}
}

func TestRouterPreloadRejectsUnknownModes(t *testing.T) {
	if err := NewRouter().Page("/guide", PageConfig{Title: "Guide", Description: "Guide", Source: "# Guide", Preload: "sometimes"}); err == nil {
		t.Fatal("unknown page preload mode was accepted")
	}
	if err := NewRouter().Screen("/guide", ScreenConfig{Title: "Guide", Description: "Guide", Component: &testScreen{}, Preload: "sometimes"}); err == nil {
		t.Fatal("unknown screen preload mode was accepted")
	}
}

func TestRouterExportsLocaleAndVersionInventories(t *testing.T) {
	r := NewRouter()
	for order, variant := range []struct{ path, locale, version string }{
		{"en/v1/guide", "en", "v1"}, {"fr/v2/guide", "fr", "v2"}, {"es/v1/guide", "es", "v1"},
	} {
		r.MustPage("/"+variant.path, PageConfig{
			Title: "Guide", Description: "Guide", Source: "# Guide", Order: order + 1,
			Metadata: ContentMetadata{Locale: variant.locale, Version: variant.version},
		})
	}
	if got := strings.Join(r.Locales(), ","); got != "en,es,fr" {
		t.Fatalf("Locales() = %q", got)
	}
	if got := strings.Join(r.Versions(), ","); got != "v1,v2" {
		t.Fatalf("Versions() = %q", got)
	}
	manifest, err := r.ExportManifestJSON("")
	if err != nil || !strings.Contains(string(manifest), `"locales": [`) || !strings.Contains(string(manifest), `"versions": [`) {
		t.Fatalf("manifest variant inventory missing: %s (%v)", manifest, err)
	}
}

func TestRouterLayoutFactoriesCustomizeGlobalAndSectionShells(t *testing.T) {
	r := NewRouter(WithLayouts(LayoutConfig{
		Global: func(*Router) *uiapp.Layout { return uiapp.NewLayout("custom-global") },
		Section: func(_ *Router, route *Route) *uiapp.Layout {
			return uiapp.NewLayout("custom-section-" + strings.TrimPrefix(route.Path, "/"))
		},
	}))
	r.MustPage("/docs", PageConfig{Title: "Docs", Description: "Docs", Source: "# Docs", Order: 1})
	if got := r.Layout().Name; got != "custom-global" {
		t.Fatalf("global layout = %q", got)
	}
	if got := r.sectionLayout(r.Routes()[0]).Name; got != "custom-section-docs" {
		t.Fatalf("section layout = %q", got)
	}
}

func TestNoIndexRoutesStayOutOfPublicSearch(t *testing.T) {
	r := NewRouter()
	r.MustPage("/public", PageConfig{Title: "Public", Description: "Public", Source: "# Public", Order: 1})
	r.MustPage("/private", PageConfig{Title: "Private", Description: "Private", Source: "# Private", Order: 2, Metadata: ContentMetadata{NoIndex: true}})
	if err := r.Validate(); err != nil {
		t.Fatalf("Validate() error = %v", err)
	}
	entries := r.SearchIndex()
	if len(entries) != 1 || entries[0].Path != "/public" {
		t.Fatalf("SearchIndex() = %#v, want only public route", entries)
	}
	if got := r.Search("private"); len(got) != 0 {
		t.Fatalf("Search() returned noindex route: %#v", got)
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

func TestSidebarMarksAnActiveParentPageWithChildren(t *testing.T) {
	r := NewRouter()
	r.MustPage("/docs", PageConfig{Title: "Documentation", Description: "Docs", Source: "# Docs", Order: 1})
	r.MustPage("/docs/getting-started", PageConfig{Title: "Getting started", Description: "Start here", Source: "# Start here", Order: 1})

	req := httptest.NewRequest("GET", "/docs", nil)
	html := (&docsSidebar{router: r}).RenderCtx(uiapp.WithRequest(context.Background(), req))
	if !strings.Contains(string(html), `data-fastr-docs-active="true"`) {
		t.Fatalf("active parent page marker missing: %s", html)
	}

	req = httptest.NewRequest("GET", "/docs/getting-started", nil)
	html = (&docsSidebar{router: r}).RenderCtx(uiapp.WithRequest(context.Background(), req))
	if strings.Contains(string(html), `data-fastr-docs-active="true"`) {
		t.Fatalf("nested page incorrectly marked its parent page as exact active: %s", html)
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

func TestPageMetadataEmitsSEOWithoutDuplicatingNativeArticleData(t *testing.T) {
	r := NewRouter()
	r.MustPage("/guide", PageConfig{
		Title:       "Deployment guide",
		Description: "Ship the documentation site.",
		Metadata: ContentMetadata{
			CanonicalURL:  "https://docs.example.com/guide",
			Image:         "/assets/guide.png",
			Authors:       []string{"Ada Lovelace", "Grace Hopper"},
			DatePublished: "2026-08-01",
			DateModified:  "2026-08-29",
			Locale:        "en-US",
			Tags:          []string{"deployment", "hosting"},
			NoIndex:       true,
		},
		Source: "# Deployment guide\n\nShip the site.",
		Order:  1,
	})

	html := metadataHeadHTML(r.Routes()[0])
	for _, marker := range []string{
		`<meta name="twitter:title" content="Deployment guide">`,
		`<meta name="twitter:description" content="Ship the documentation site.">`,
		`<meta name="twitter:image" content="/assets/guide.png">`,
		`<meta name="twitter:card" content="summary_large_image">`,
		`<meta name="author" content="Ada Lovelace">`,
		`<meta property="article:published_time" content="2026-08-01">`,
		`<meta property="article:tag" content="deployment">`,
		`<meta name="robots" content="noindex,nofollow">`,
		`<link rel="canonical" href="https://docs.example.com/guide">`,
	} {
		if !strings.Contains(html, marker) {
			t.Fatalf("metadata head missing %q: %s", marker, html)
		}
	}

	r = NewRouter()
	r.MustPage("/unsafe", PageConfig{
		Title: "Unsafe metadata", Description: "A page", Source: "# Unsafe", Order: 1,
		Metadata: ContentMetadata{Image: "javascript:alert(1)"},
	})
	html = metadataHeadHTML(r.Routes()[0])
	if strings.Contains(html, "javascript:") || strings.Contains(html, "og:image") {
		t.Fatalf("unsafe image metadata was emitted: %s", html)
	}
	article := (&pageComponent{router: r, route: r.Routes()[0]}).ScreenArticle()
	if article.Headline != "Unsafe metadata" || article.Description != "A page" || article.Image != "javascript:alert(1)" {
		t.Fatalf("page did not expose native article metadata: %#v", article)
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

func TestThemeTemplatesExposeFiveCompleteDistinctPresets(t *testing.T) {
	templates := ThemeTemplates()
	if len(templates) != 5 {
		t.Fatalf("ThemeTemplates() returned %d templates, want 5", len(templates))
	}
	seen := make(map[Template]bool, len(templates))
	primaries := make(map[string]bool, len(templates))
	for _, template := range templates {
		if seen[template] {
			t.Fatalf("ThemeTemplates() repeated %q", template)
		}
		seen[template] = true
		r := NewRouter(WithTemplate(template))
		if r.Template() != template {
			t.Fatalf("Router.Template() = %q, want %q", r.Template(), template)
		}
		th := r.Theme()
		if th.Colors.Background.Value == "" || th.Colors.Primary.Value == "" || th.DarkColors["background"] == "" {
			t.Fatalf("template %q did not produce a complete adaptive theme", template)
		}
		primaries[th.Colors.Primary.Value] = true
		if !strings.Contains(r.CSS(), "--fastr-docs-template: "+string(template)) {
			t.Fatalf("template %q CSS marker is missing", template)
		}
		if strings.TrimSpace(TemplateDescription(template)) == "" {
			t.Fatalf("template %q has no description", template)
		}
	}
	if len(primaries) != len(templates) {
		t.Fatalf("templates do not have distinct primary tokens: %#v", primaries)
	}
}

func TestWithThemeAppliesCustomTokensAndCSSAfterTemplate(t *testing.T) {
	r := NewRouter(
		WithTheme(ThemeConfig{
			Template: TemplateBlueprint,
			Overrides: ThemeOverrides{
				Primary:    "#123456",
				DarkColors: map[string]string{"primary": "#abcdef"},
			},
			CustomCSS: ".project-docs { --example: 1; }",
		}),
	)
	if r.Template() != TemplateBlueprint {
		t.Fatalf("Router.Template() = %q, want %q", r.Template(), TemplateBlueprint)
	}
	if got := r.Theme().Colors.Primary.Value; got != "#123456" {
		t.Fatalf("custom light primary = %q, want #123456", got)
	}
	if got := r.Theme().DarkColors["primary"]; got != "#abcdef" {
		t.Fatalf("custom dark primary = %q, want #abcdef", got)
	}
	if !strings.Contains(r.CSS(), ".project-docs { --example: 1; }") {
		t.Fatal("custom CSS was not appended to the router stylesheet")
	}

	ordered := NewRouter(
		WithTheme(ThemeConfig{Overrides: ThemeOverrides{Primary: "#654321"}}),
		WithTemplate(TemplateStudio),
	)
	if got := ordered.Theme().Colors.Primary.Value; got != "#654321" {
		t.Fatalf("custom primary was lost when WithTemplate followed WithTheme: %q", got)
	}
}

func TestThemeColorFollowsTemplateAndBrandOverride(t *testing.T) {
	blueprint := NewRouter(WithTemplate(TemplateBlueprint))
	if got := blueprint.ThemeColor(); got != blueprint.Theme().Colors.Background.Value {
		t.Fatalf("ThemeColor() = %q, want the active light background %q", got, blueprint.Theme().Colors.Background.Value)
	}
	branded := NewRouter(
		WithTemplate(TemplateNotebook),
		WithBrand(BrandConfig{ThemeColor: "#123456"}),
	)
	if got := branded.ThemeColor(); got != "#123456" {
		t.Fatalf("branded ThemeColor() = %q, want #123456", got)
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

func TestSearchIndexPathDefaultsAndOverrides(t *testing.T) {
	if got := NewRouter().SearchIndexPath(); got != "/__fastr-docs/search.json" {
		t.Fatalf("SearchIndexPath() = %q, want the mounted starter default", got)
	}
	for _, testCase := range []struct{ in, want string }{
		{"/__manual/search.json", "/__manual/search.json"},
		{"assets/search.json", "/assets/search.json"},
		{"   ", "/__fastr-docs/search.json"},
	} {
		if got := NewRouter(WithSearchIndexPath(testCase.in)).SearchIndexPath(); got != testCase.want {
			t.Fatalf("WithSearchIndexPath(%q) = %q, want %q", testCase.in, got, testCase.want)
		}
	}
}

func TestCommandPaletteTriggerPublishesConfiguredSearchIndexPath(t *testing.T) {
	r := NewRouter(WithSearchIndexPath("/__manual/search.json"))
	r.MustPage("/", PageConfig{Title: "Home", Source: "# Home"})

	visible, _ := r.ensureCommandPalette()
	markup := string(visible)
	if !strings.Contains(markup, `data-fastr-docs-index-path="/__manual/search.json"`) {
		t.Fatalf("command palette trigger did not publish the configured index path: %s", markup)
	}
	if strings.Contains(markup, "/__fastr-docs/search.json") {
		t.Fatalf("command palette trigger kept the hardcoded index path: %s", markup)
	}
}
