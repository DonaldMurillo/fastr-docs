package main

import (
	"context"
	"embed"
	"fmt"
	"io/fs"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing/fstest"

	docs "github.com/DonaldMurillo/fastr-docs"
	"github.com/DonaldMurillo/fastr-docs/plugin/openapi"
	uiapp "github.com/DonaldMurillo/gofastr/core-ui/app"
	"github.com/DonaldMurillo/gofastr/core/middleware"
	"github.com/DonaldMurillo/gofastr/core/render"
	"github.com/DonaldMurillo/gofastr/core/static"
	"github.com/DonaldMurillo/gofastr/framework"
	fwimage "github.com/DonaldMurillo/gofastr/framework/image"
	"github.com/DonaldMurillo/gofastr/framework/ui"
	"github.com/DonaldMurillo/gofastr/framework/uihost"
)

// The fixture is deliberately hand-authored. It does not call the fastr-docs
// CLI or reuse the starter template; it exercises the public Router and
// GoFastr component APIs exactly as an adopters' project would.
//
//go:embed openapi.json
var contractFS embed.FS

func main() {
	built, err := buildSite()
	if err != nil {
		panic(err)
	}
	if dir := exportDir(os.Args[1:]); dir != "" {
		if err := built.server.ExportStatic(context.Background(), dir, normalizeBase(exportBase(os.Args[1:]))); err != nil {
			panic(err)
		}
		if err := writeRuntimeAssets(dir, built.openAPIRuntime, built.searchIndex, built.manifest); err != nil {
			panic(err)
		}
		if err := writeManualAssets(dir); err != nil {
			panic(err)
		}
		if err := docs.WriteAgentAssets(dir, normalizeBase(exportBase(os.Args[1:])), built.server.Router()); err != nil {
			panic(err)
		}
		if err := rewriteStaticCSP(dir, docs.ContentSecurityPolicy(built.connectOrigins...)); err != nil {
			panic(err)
		}
		fmt.Println("static site exported to " + dir)
		return
	}

	addr := ":3078"
	if port := os.Getenv("PORT"); port != "" {
		addr = ":" + port
	}
	fmt.Println("Manual fixture listening on http://localhost" + addr)
	if err := built.server.Start(addr); err != nil {
		panic(err)
	}
}

type builtSite struct {
	server         *framework.App
	openAPIRuntime []byte
	searchIndex    []byte
	manifest       []byte
	connectOrigins []string
}

func buildSite() (*builtSite, error) {
	contract, err := fs.ReadFile(contractFS, "openapi.json")
	if err != nil {
		return nil, err
	}

	router := docs.NewRouter(docs.WithSiteName("Manual Docs"), docs.WithLanguage(os.Getenv("DOCS_LOCALE")))
	if err := router.Use(docs.MarkdownComponentsPlugin{Components: map[string]docs.MarkdownComponent{
		"note": func(props map[string]string, body render.HTML) render.HTML {
			return ui.Callout(ui.CalloutConfig{Title: props["title"], Variant: ui.StatusInfo}, body)
		},
	}}); err != nil {
		panic(err)
	}
	router.MustPage("/", docs.PageConfig{
		Title:       "Manual Docs",
		Description: "A hand-authored project exercising the public Router API.",
		Source:      "# Manual Docs\n\nThis is a hand-authored documentation site built with the public `docs.Router` API. Start with the guide, try the interactive framework surface, or browse the API reference.\n\n## Start here\n\nNew to this project? Read the [Getting started guide](/guides/getting-started) to follow the route tree from a nested Markdown page to a typed screen.\n\n- [Getting started](/guides/getting-started) — register pages, groups, and screens.\n- [Patterns](/guides/patterns) — see how explicit order keeps a growing tree readable.\n- [Framework lab](/framework-lab) — try the interactive GoFastr surfaces used by this fixture.\n\n## Choose a surface\n\nUse the route that matches what you need to do:\n\n- [Guides](/guides/getting-started) — learn the authoring model and navigation tree.\n- [Framework lab](/framework-lab) — exercise tabs, forms, charts, filters, and typed state.\n- [Localized guides](/locales/pt-BR/v1/guide) — switch language and version while staying in the same page family.\n- [Manual fixture API](/api-reference) — search the OpenAPI contract and send a safe GET request.\n\n## What this project demonstrates\n\nMarkdown pages, typed screens, nested groups, local search, PWA export, and OpenAPI all share one route tree. The sidebar shows the full tree here, then narrows to the active section as you read.\n\n## Next steps\n\n1. [Read Getting started](/guides/getting-started).\n2. [Open the Framework lab](/framework-lab).\n3. [Inspect the API reference](/api-reference).\n4. Press `Ctrl+K` or select Search in the header to jump to any route.",
		Order:       1,
		Offline:     true,
	})
	guides := router.MustGroup("/guides", docs.GroupConfig{Title: "Guides", Description: "Hand-authored guides", Order: 2, Badge: docs.NavBadge{Label: "Popular", Tone: docs.NavBadgeToneInfo}})
	guides.MustPage("getting-started", docs.PageConfig{
		Title:       "Getting started",
		Description: "The first guide in a nested group.",
		Source:      "# Getting started\n\nThis page proves nested Markdown navigation and the sticky in-page rail.\n\n{{< note title=\"Manual component\" >}}\nThe hand-authored fixture registers a typed component vocabulary through the public plugin contract.\n{{< /note >}}\n\n## Compose a route tree\n\nRegister pages and screens against one router.\n\n## Add a screen\n\nTyped GoFastr screens can sit beside Markdown.",
		Order:       1,
	})
	guides.MustPage("patterns", docs.PageConfig{
		Title:       "Patterns",
		Description: "A second nested page with a different explicit order.",
		Source:      "# Patterns\n\nUse groups to make large documentation trees legible.\n\n## Keep order explicit\n\nThe Router preserves the author’s order.",
		Order:       2,
		Badge:       docs.NavBadge{Label: "New"},
		Metadata:    docs.ContentMetadata{Tags: []string{"patterns", "routing"}, NoIndex: true, CanonicalURL: "https://docs.example/patterns", EditURL: "https://github.com/acme/docs/edit/main/patterns.md", Redirects: []string{"/old-patterns"}},
	})
	variants := router.MustGroup("/locales", docs.GroupConfig{Title: "Localized guides", Description: "Runtime locale and version switching.", Order: 4})
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
		variants.MustPage(item.path, docs.PageConfig{
			Title: "Variant guide", Description: "A localized and versioned guide.", Source: "# Variant guide\n\nThis page proves the runtime locale and version selectors.", Order: order + 1,
			Metadata: docs.ContentMetadata{Locale: item.locale, Version: item.version},
		})
	}
	router.MustScreen("/framework-lab", docs.ScreenConfig{
		Title:       "Framework lab",
		Description: "Interactive GoFastr component surfaces.",
		Component:   &frameworkLabScreen{},
		Order:       3,
		Badge:       docs.NavBadge{Label: "Live", Tone: docs.NavBadgeToneSuccess},
	})
	if err := router.Use(openapi.Plugin{
		Spec:        contract,
		ServerURL:   os.Getenv("API_SERVER_URL"),
		Title:       "Manual fixture API",
		Description: "The OpenAPI plugin is mounted into the same hand-authored Router.",
	}); err != nil {
		return nil, err
	}
	if err := router.Validate(); err != nil {
		return nil, err
	}

	application := uiapp.NewApp("Manual Docs").WithTheme(docs.DefaultTheme()).WithLang(router.Language())
	layout := router.Layout()
	if err := router.Mount(application, layout); err != nil {
		return nil, err
	}

	host := uihost.New(application,
		uihost.WithDescription("A hand-authored GoFastr documentation project."),
		uihost.WithCustomCSS(router.CSS()+"\n"+openapi.CSS()+"\n"+frameworkLabCSS),
		uihost.WithPublicLLMMD(),
		uihost.WithAgentReady(uihost.AgentReadyConfig{
			Title:     "Manual Docs",
			Summary:   "Hand-authored documentation with a typed framework lab.",
			WhenToUse: "Use this site to read the guides, exercise the framework lab, and inspect the API reference.",
		}),
		uihost.WithSitemap(uihost.SitemapConfig{BaseURL: manualSiteURL(), ExcludePaths: append([]string{"/__manual/"}, router.SitemapExcludePaths()...)}),
		uihost.WithRobots(uihost.RobotsConfig{Disallow: []string{"/__manual/"}}),
		uihost.WithExtraScripts("/__manual/docs.js", "/__manual/openapi.js"),
		uihost.WithAppIcon(manualAppIconPNG()),
		uihost.WithPWA(uihost.PWAConfig{
			Name:        "Manual Docs",
			ShortName:   "Manual Docs",
			Description: "A hand-authored GoFastr documentation project.",
			Precache:    []string{"/__manual/docs.js", "/__manual/openapi.js", "/__manual/search.json", "/__manual/manifest.json"},
		}),
	)
	server := framework.NewUIHostApp(host, framework.WithConfig(framework.AppConfig{
		Name: "Manual Docs",
		SecurityHeaders: middleware.SecurityHeadersConfig{
			ContentSecurityPolicy: docs.ContentSecurityPolicy(router.ConnectOrigins()...),
		},
	}))
	if err := router.MountNavigation(server.Router()); err != nil {
		return nil, err
	}
	if err := router.MountCommandPalette(server.Router()); err != nil {
		return nil, err
	}
	server.Router().PostFunc("/manual-lab/save", savePreferences)

	openAPIRuntime, err := fs.ReadFile(openapi.Assets(), "openapi.js")
	if err != nil {
		return nil, err
	}
	searchIndex, err := router.SearchIndexJSON()
	if err != nil {
		return nil, err
	}
	manifest, err := router.ExportManifestJSON("")
	if err != nil {
		return nil, err
	}
	assets := fstest.MapFS{
		"docs.js":       &fstest.MapFile{Data: []byte(docs.RuntimeJS())},
		"openapi.js":    &fstest.MapFile{Data: openAPIRuntime},
		"search.json":   &fstest.MapFile{Data: searchIndex},
		"manifest.json": &fstest.MapFile{Data: manifest},
	}
	server.Router().Get("/__manual/{path...}", static.Handler(static.Config{FS: assets, Prefix: "/__manual"}))
	if err := router.MountAssets(server.Router(), docs.AssetConfig{FS: fstest.MapFS{"fixture.txt": &fstest.MapFile{Data: []byte("manual asset")}}, Prefix: "/assets"}); err != nil {
		return nil, err
	}
	return &builtSite{server: server, openAPIRuntime: openAPIRuntime, searchIndex: searchIndex, manifest: manifest, connectOrigins: router.ConnectOrigins()}, nil
}

func manualAppIconPNG() []byte {
	image, err := fwimage.NewGradient(512, 512, "#f97316", "#9a3412")
	if err != nil {
		return nil
	}
	data, err := image.PNG().Bytes()
	if err != nil {
		return nil
	}
	return data
}

func savePreferences(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, "invalid form", http.StatusBadRequest)
		return
	}
	http.Redirect(w, r, "/framework-lab?saved=1", http.StatusSeeOther)
}

type frameworkLabScreen struct{}

func (*frameworkLabScreen) ScreenTitle() string { return "Framework lab" }
func (*frameworkLabScreen) ScreenDescription() string {
	return "Interactive GoFastr component surfaces."
}

func (s *frameworkLabScreen) Render() render.HTML { return s.render(nil, false) }

func (s *frameworkLabScreen) RenderCtx(ctx context.Context) render.HTML {
	saved := false
	if request := uiapp.RequestFromContext(ctx); request != nil {
		saved = request.URL.Query().Get("saved") == "1"
	}
	return s.render(ctx, saved)
}

func (*frameworkLabScreen) render(ctx context.Context, saved bool) render.HTML {
	return render.Tag("div", map[string]string{"class": "manual-lab", "data-testid": "framework-lab"},
		render.Tag("header", map[string]string{"class": "manual-lab__hero"},
			render.Tag("span", map[string]string{"class": "manual-lab__eyebrow"}, render.Text("Hand-authored fixture")),
			render.Tag("h1", nil, render.Text("Framework lab")),
			render.Tag("p", nil, render.Text("A real GoFastr screen composed directly from framework/ui and core-ui primitives. Every section below is an E2E user surface.")),
		),
		ui.Banner(ui.BannerConfig{Title: "Live framework surface", Body: "This route was registered with docs.Router.Screen, not generated by the CLI.", Variant: ui.BannerInfo, Dismissible: true, DismissID: "manual-fixture-banner", ID: "manual-fixture-banner", Ctx: ctx}),
		render.Tag("section", map[string]string{"class": "manual-lab__stats", "data-testid": "stats-surface"},
			ui.StatCard(ui.StatCardConfig{Label: "Routes", Value: "7", Trend: "+2 nested guides", Direction: ui.TrendUp}),
			ui.StatCard(ui.StatCardConfig{Label: "Components", Value: "24", Trend: "SSR + hydrated", Direction: ui.TrendUp}),
			ui.StatCard(ui.StatCardConfig{Label: "Offline", Value: "Ready", Trend: "PWA shell enabled", Direction: ui.TrendFlat}),
		),
		render.Tag("section", map[string]string{"class": "manual-lab__section", "data-testid": "interactive-surface"},
			render.Tag("h2", nil, render.Text("Interactive primitives")),
			render.Tag("div", map[string]string{"class": "manual-lab__grid"},
				ui.Card(ui.CardConfig{Heading: "Tabs and disclosure", HeadingLevel: 3},
					ui.Tabs(ui.TabsConfig{SignalName: "manual-tabs", Tabs: []ui.TabItem{
						{Label: "Overview", Content: render.Tag("p", nil, render.Text("The first tab is active on the server."))},
						{Label: "Runtime", Content: render.Tag("p", nil, render.Text("The signal runtime owns the selection."))},
					}}),
					ui.Collapsible(ui.CollapsibleConfig{Summary: "Show implementation note"}, render.Tag("p", nil, render.Text("Disclosure uses native details semantics with framework behavior."))),
				),
				ui.Card(ui.CardConfig{Heading: "Local controls", HeadingLevel: 3},
					ui.Counter(ui.CounterConfig{SignalName: "manual-counter", Step: 1}),
					ui.Switch(ui.ToggleConfig{Name: "notifications", Label: "Enable notifications", Checked: true, ID: "manual-notifications"}),
					ui.SegmentedControl(ui.SegmentedControlConfig{Name: "density", Label: "Density", Selected: "comfortable", Options: []ui.SegmentedOption{{Label: "Compact", Value: "compact"}, {Label: "Comfortable", Value: "comfortable"}, {Label: "Spacious", Value: "spacious"}}}),
				),
			),
		),
		render.Tag("section", map[string]string{"class": "manual-lab__section", "data-testid": "form-surface"},
			render.Tag("h2", nil, render.Text("Form controls")),
			ui.Form(ui.FormConfig{Action: "/manual-lab/save", Method: "POST", SubmitLabel: "Save preferences", ID: "manual-preferences"},
				ui.TextField(ui.TextFieldConfig{Name: "name", Label: "Display name", ID: "manual-name", Value: "Ada Lovelace", Required: true}),
				ui.Select(ui.SelectConfig{Name: "theme", Label: "Preferred theme", ID: "manual-theme", Options: []ui.SelectOption{{Value: "system", Text: "System", Selected: true}, {Value: "light", Text: "Light"}, {Value: "dark", Text: "Dark"}}}),
				ui.NumberInput(ui.NumberInputConfig{Name: "sessions", Label: "Sessions per week", ID: "manual-sessions", Value: 3, Min: 0, Max: 10}),
				ui.TextArea(ui.TextAreaConfig{Name: "notes", Label: "Notes", ID: "manual-notes", Value: "Keep the route tree intentional.", Rows: 3, Autogrow: true}),
				ui.RatingInput(ui.RatingConfig{Name: "rating", Label: "Rate this fixture", Value: 4, ID: "manual-rating"}),
				ui.TagInput(ui.TagInputConfig{Name: "tags", Label: "Tags", Values: []string{"docs", "gofastr"}, Placeholder: "Add a tag", ID: "manual-tags"}),
			),
			func() render.HTML {
				if saved {
					return ui.Callout(ui.CalloutConfig{Title: "Preferences saved", Variant: ui.StatusSuccess, ID: "manual-saved", Landmark: boolPtr(false)}, render.Text("The native POST completed and redirected back to this screen."))
				}
				return render.Text("")
			}(),
		),
		render.Tag("section", map[string]string{"class": "manual-lab__section", "data-testid": "content-surface"},
			render.Tag("h2", nil, render.Text("Content and data")),
			render.Tag("div", map[string]string{"class": "manual-lab__grid"},
				ui.Card(ui.CardConfig{Heading: "Code", HeadingLevel: 3},
					ui.CodeTabs(ui.CodeTabsConfig{Name: "manual-code", Label: "Examples", LineNumbers: true},
						ui.CodeSample{Label: "Go", Language: "go", Filename: "main.go", Code: "router.MustScreen(\"/framework-lab\", docs.ScreenConfig{...})"},
						ui.CodeSample{Label: "JSON", Language: "json", Code: "{ \"route\": \"/framework-lab\" }"},
					),
					ui.CopyButton(ui.CopyButtonConfig{Target: "#manual-copy-source", Label: "Copy route", ToastOnCopy: true}),
					render.Tag("code", map[string]string{"id": "manual-copy-source", "class": "manual-lab__copy-source"}, render.Text("/framework-lab")),
				),
				ui.Card(ui.CardConfig{Heading: "Carousel", HeadingLevel: 3},
					ui.Carousel(ui.CarouselConfig{ID: "manual-carousel", Label: "Fixture highlights", Loop: true, Slides: []ui.CarouselSlide{
						{Label: "First highlight", Content: render.Tag("p", nil, render.Text("Router-owned composition."))},
						{Label: "Second highlight", Content: render.Tag("p", nil, render.Text("Framework-owned interaction."))},
						{Label: "Third highlight", Content: render.Tag("p", nil, render.Text("Exportable PWA shell."))},
					}}),
				),
			),
		),
		render.Tag("section", map[string]string{"class": "manual-lab__section", "data-testid": "data-surface"},
			render.Tag("h2", nil, render.Text("Tables, filters, and charts")),
			ui.FilterToolbar(ui.FilterToolbarConfig{Action: "/framework-lab", Search: &ui.FilterSearch{Name: "q", Placeholder: "Search fixture"}, Facets: []ui.Facet{{Name: "status", Label: "Status", Options: []ui.FacetOption{{Label: "Ready", Value: "ready"}, {Label: "Draft", Value: "draft"}}}}, Sort: []ui.SortOption{{Label: "Name", Value: "name"}, {Label: "Recent", Value: "recent"}}, ID: "manual-filters"}),
			ui.DataTable(ui.DataTableConfig{ID: "manual-table", Caption: "Fixture routes", Columns: []ui.Column{{Key: "route", Header: "Route", Sortable: true}, {Key: "kind", Header: "Kind"}, {Key: "status", Header: "Status"}}, SortHrefPattern: "?sort=%s&dir=%s", Rows: []ui.Row{{ID: "route-home", Cells: map[string]render.HTML{"route": render.Text("/"), "kind": render.Text("Markdown"), "status": ui.StatusBadge(ui.StatusBadgeConfig{Label: "Ready", Variant: ui.StatusSuccess})}}, {ID: "route-lab", Cells: map[string]render.HTML{"route": render.Text("/framework-lab"), "kind": render.Text("Screen"), "status": ui.StatusBadge(ui.StatusBadgeConfig{Label: "Ready", Variant: ui.StatusSuccess})}}}, Responsive: ui.ResponsiveCards}),
			render.Tag("div", map[string]string{"class": "manual-lab__charts"},
				ui.LineChart(ui.LineChartConfig{ID: "manual-line-chart", Series: []ui.LineSeries{{Name: "Routes", Values: []float64{2, 4, 5, 7}, Area: true}}, Labels: []string{"Q1", "Q2", "Q3", "Q4"}, ShowLegend: true, LabelledBy: "manual-line-title"}),
				ui.BarChart(ui.BarChartConfig{ID: "manual-bar-chart", Bars: []ui.BarChartBar{{Label: "Docs", Value: 5, Color: "primary"}, {Label: "Screens", Value: 2, Color: "success"}}, ShowAxis: true, ShowLabels: true, LabelledBy: "manual-bar-title"}),
				ui.PieChart(ui.PieChartConfig{ID: "manual-pie-chart", Slices: []ui.PieSlice{{Label: "Pages", Value: 5, Color: "primary"}, {Label: "Screens", Value: 2, Color: "success"}}, InnerRadius: .55, CenterLabel: "7", CenterSubtext: "routes", LabelledBy: "manual-pie-title"}),
			),
		),
	)
}

func boolPtr(value bool) *bool { return &value }

const frameworkLabCSS = `.manual-lab { max-width: 1120px; margin: 0 auto; padding: 48px clamp(20px, 5vw, 64px) 96px; }
.manual-lab__hero { max-width: 700px; margin-bottom: 32px; }
.manual-lab__eyebrow { color: var(--color-primary); font: 700 .72rem/1.2 var(--font-mono, monospace); letter-spacing: .12em; text-transform: uppercase; }
.manual-lab h1 { margin: 10px 0; font-size: clamp(2.5rem, 6vw, 5rem); letter-spacing: -.06em; }
.manual-lab h2 { margin: 0 0 18px; font-size: 1.45rem; letter-spacing: -.035em; }
.manual-lab__hero p { color: var(--color-text-muted); font-size: 1.05rem; line-height: 1.7; }
.manual-lab__stats, .manual-lab__grid, .manual-lab__charts { display: grid; grid-template-columns: repeat(3, minmax(0, 1fr)); gap: 14px; }
.manual-lab__section { margin-top: 54px; }
.manual-lab__grid { grid-template-columns: repeat(2, minmax(0, 1fr)); }
.manual-lab__charts { align-items: center; margin-top: 24px; }
.manual-lab__copy-source { display: inline-block; margin-top: 12px; padding: 6px 8px; border-radius: 5px; background: var(--color-surface-soft); }
.manual-lab [data-fui-comp="ui-card"] { height: 100%; }
.manual-lab form { display: grid; gap: 16px; max-width: 720px; padding: 24px; border: 1px solid var(--color-border); border-radius: 10px; background: var(--color-surface); }
@media (max-width: 860px) { .manual-lab__stats, .manual-lab__grid, .manual-lab__charts { grid-template-columns: 1fr; } }
`

func writeRuntimeAssets(dir string, openAPIRuntime, searchIndex, manifest []byte) error {
	assetDir := filepath.Join(dir, "__manual")
	if err := os.MkdirAll(assetDir, 0o755); err != nil {
		return err
	}
	if err := os.WriteFile(filepath.Join(assetDir, "docs.js"), []byte(docs.RuntimeJS()), 0o644); err != nil {
		return err
	}
	if err := os.WriteFile(filepath.Join(assetDir, "openapi.js"), openAPIRuntime, 0o644); err != nil {
		return err
	}
	if err := os.WriteFile(filepath.Join(assetDir, "search.json"), searchIndex, 0o644); err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(assetDir, "manifest.json"), manifest, 0o644)
}

func writeManualAssets(dir string) error {
	if err := os.MkdirAll(filepath.Join(dir, "assets"), 0o755); err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(dir, "assets", "fixture.txt"), []byte("manual asset"), 0o644)
}

func manualSiteURL() string {
	if value := strings.TrimSpace(os.Getenv("PUBLIC_SITE_URL")); value != "" {
		return strings.TrimRight(value, "/")
	}
	return "http://localhost:3078"
}

func rewriteStaticCSP(dir, policy string) error {
	oldContent := `content="default-src 'self'; img-src 'self' data:; object-src 'none'; form-action 'self'; frame-ancestors 'none'; base-uri 'self'"`
	newContent := `content="` + policy + `"`
	return filepath.WalkDir(dir, func(path string, entry fs.DirEntry, err error) error {
		if err != nil || entry.IsDir() || filepath.Ext(path) != ".html" {
			return err
		}
		body, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		return os.WriteFile(path, []byte(strings.ReplaceAll(string(body), oldContent, newContent)), 0o644)
	})
}

func exportDir(args []string) string {
	for i := 0; i < len(args); i++ {
		if args[i] == "--export" && i+1 < len(args) {
			return args[i+1]
		}
		if strings.HasPrefix(args[i], "--export=") {
			return strings.TrimPrefix(args[i], "--export=")
		}
	}
	return ""
}

func exportBase(args []string) string {
	for i := 0; i < len(args); i++ {
		if args[i] == "--export-base" && i+1 < len(args) {
			return args[i+1]
		}
		if strings.HasPrefix(args[i], "--export-base=") {
			return strings.TrimPrefix(args[i], "--export-base=")
		}
	}
	return ""
}

func normalizeBase(base string) string {
	base = strings.TrimSpace(base)
	if base == "" || base == "/" {
		return ""
	}
	if !strings.HasPrefix(base, "/") {
		base = "/" + base
	}
	return strings.TrimRight(base, "/")
}
