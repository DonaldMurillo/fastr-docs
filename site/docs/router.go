package docs

import (
	"os"

	docs "github.com/DonaldMurillo/fastr-docs"
	"github.com/DonaldMurillo/fastr-docs/plugin/katex"
	"github.com/DonaldMurillo/fastr-docs/plugin/mermaid"
	"github.com/DonaldMurillo/fastr-docs/plugin/openapi"
	uiapp "github.com/DonaldMurillo/gofastr/core-ui/app"
	"github.com/DonaldMurillo/gofastr/core/render"
	"github.com/DonaldMurillo/gofastr/framework/ui"
)

// NewRouter is the central project object. Every page, typed screen, and
// plugin contributes to this one tree, so navigation and local search cannot
// drift apart.
func NewRouter() *docs.Router {
	router := docs.NewRouter(
		docs.WithSiteName("fastr-docs"),
		docs.WithBrand(docs.BrandConfig{Name: "fastr-docs", FaviconURL: "/assets/favicon.svg"}),
		docs.WithIncludeDrafts(os.Getenv("DOCS_INCLUDE_DRAFTS") == "1"),
		docs.WithLanguage(os.Getenv("DOCS_LOCALE")),
		docs.WithLocale(os.Getenv("DOCS_CONTENT_LOCALE")),
		// Keeps a build for one locale usable while most pages are still
		// untranslated, which is the normal state of a real site.
		docs.WithLocaleFallback("en"),
		docs.WithVersion(os.Getenv("DOCS_CONTENT_VERSION")),
		docs.WithTemplate(docs.ParseTemplate(os.Getenv("DOCS_TEMPLATE"))),
		docs.WithSearchBackend(os.Getenv("DOCS_SEARCH_BACKEND")),
		// Last-updated dates come from git history, so no page has to carry a
		// hand-maintained date. Set DOCS_REPO_URL to also get edit links.
		docs.WithGitMetadata(docs.GitMetadataConfig{
			RepoURL: os.Getenv("DOCS_REPO_URL"),
			Branch:  os.Getenv("DOCS_REPO_BRANCH"),
		}),
	)
	// Diagrams render inside a sandboxed frame so the pages keep their strict
	// content policy. See docs/build/diagrams.
	if err := router.Use(mermaid.Plugin{}); err != nil {
		panic(err)
	}
	// Math renders in the page instead, because KaTeX needs no policy
	// relaxation and inline math has to sit on the text baseline. See
	// docs/build/math.
	if err := router.Use(katex.Plugin{}); err != nil {
		panic(err)
	}
	if err := router.Use(docs.MarkdownComponentsPlugin{Components: map[string]docs.MarkdownComponent{
		"callout": func(props map[string]string, body render.HTML) render.HTML {
			return ui.Callout(ui.CalloutConfig{Title: props["title"], Variant: ui.StatusInfo}, body)
		},
	}}); err != nil {
		panic(err)
	}
	// The landing page is front matter now, which is what a splash template is
	// for. HomeBody is kept in docs/home.go and mounted at /examples/home as the
	// worked example of the typed-screen route.
	router.MustPage("/", docs.PageConfig{
		Title:       "fastr-docs",
		Description: "A white-label documentation site.",
		SourcePath:  contentFile("index-splash.md"),
		Order:       1,
		Offline:     true,
	})
	router.MustPage("/docs", docs.PageConfig{
		Title:       "Documentation",
		Description: "Learn the route model, author content, and ship a docs site.",
		SourcePath:  contentFile("index.md"),
		Order:       2,
		Offline:     true,
		Badge:       docs.NavBadge{Label: "Popular", Tone: docs.NavBadgeToneInfo},
	})
	router.MustPage("/docs/getting-started", docs.PageConfig{
		Title:       "Getting started",
		Description: "Install fastr-docs and create your first route tree.",
		SourcePath:  contentFile("getting-started.md"),
		Order:       1,
		Offline:     true,
		Badge:       docs.NavBadge{Label: "Start here"},
	})
	concepts := router.MustGroup("/docs/concepts", docs.GroupConfig{
		Title:       "Concepts",
		Description: "The route model and composition primitives.",
		Order:       2,
	})
	concepts.MustPage("router", docs.PageConfig{
		Title:       "The router",
		Description: "Build navigation, search, and rendering from one tree.",
		SourcePath:  contentFile("concepts-router.md"),
		Order:       1,
		Offline:     true,
	})
	concepts.MustPage("content", docs.PageConfig{
		Title:       "Content authoring",
		Description: "Choose Markdown for durable, searchable documentation.",
		SourcePath:  contentFile("concepts-content.md"),
		Order:       2,
		Offline:     true,
	})
	concepts.MustPage("layouts", docs.PageConfig{
		Title:       "Layouts and navigation",
		Description: "Compose global shells and nested section layouts.",
		SourcePath:  contentFile("concepts-layouts.md"),
		Order:       3,
		Offline:     true,
	})

	build := router.MustGroup("/docs/build", docs.GroupConfig{
		Title:       "Build",
		Description: "Add screens, API references, and project extensions.",
		Order:       3,
		Badge:       docs.NavBadge{Label: "Build", Tone: docs.NavBadgeToneAccent},
	})
	build.MustPage("screens", docs.PageConfig{
		Title:       "Screens and components",
		Description: "Use typed GoFastr screens when Markdown is not enough.",
		SourcePath:  contentFile("build-screens.md"),
		Order:       1,
		Offline:     true,
	})
	build.MustPage("framework-ui", docs.PageConfig{
		Title:       "Framework UI",
		Description: "Use the native GoFastr components that power docs pages and screens.",
		SourcePath:  contentFile("build-ui.md"),
		Order:       2,
		Offline:     true,
	})
	build.MustPage("openapi", docs.PageConfig{
		Title:       "OpenAPI reference",
		Description: "Turn an API contract into a searchable request console.",
		SourcePath:  contentFile("build-openapi.md"),
		Order:       3,
		Offline:     true,
	})
	build.MustPage("plugins", docs.PageConfig{
		Title:       "Plugins and extensions",
		Description: "Extend the same router without forking the docs shell.",
		SourcePath:  contentFile("build-plugins.md"),
		Order:       4,
		Offline:     true,
	})
	build.MustPage("blog", docs.PageConfig{
		Title:       "Blog and RSS",
		Description: "Publish Markdown updates and expose the same posts as RSS.",
		SourcePath:  contentFile("build-blog.md"),
		Order:       5,
		Offline:     true,
	})
	build.MustPage("themes", docs.PageConfig{
		Title:       "Themes and templates",
		Description: "Choose one of five visual starting points and customize its tokens.",
		SourcePath:  contentFile("build-themes.md"),
		Order:       6,
		Offline:     true,
	})
	build.MustPage("diagrams", docs.PageConfig{
		Title:       "Diagrams",
		Description: "Render Mermaid diagrams without weakening the content policy.",
		SourcePath:  contentFile("build-diagrams.md"),
		Order:       8,
		Offline:     true,
		Badge:       docs.NavBadge{Label: "New", Tone: docs.NavBadgeToneInfo},
	})
	build.MustPage("math", docs.PageConfig{
		Title:       "Math",
		Description: "Write TeX with dollar delimiters and render it in the page.",
		SourcePath:  contentFile("build-math.md"),
		Order:       9,
		Offline:     true,
		Badge:       docs.NavBadge{Label: "New", Tone: docs.NavBadgeToneInfo},
	})
	build.MustPage("components", docs.PageConfig{
		Title:       "Markdown components",
		Description: "The shortcode vocabulary every project starts with.",
		SourcePath:  contentFile("build-components.md"),
		Order:       7,
		Offline:     true,
		Badge:       docs.NavBadge{Label: "New", Tone: docs.NavBadgeToneInfo},
	})
	// A partial Spanish translation. Only three pages exist in Spanish on
	// purpose: it demonstrates the language selector, and it demonstrates what a
	// half-translated site looks like, which is what every real one is.
	//
	// The paths mirror the English ones with an /es prefix so variantFamily
	// pairs them: /es/docs/getting-started strips its locale segment and
	// matches /docs/getting-started.
	router.MustPage("/es", docs.PageConfig{
		Title:       "Español",
		Description: "Documentación construida sobre un árbol de rutas.",
		SourcePath:  siteFile("content", "es", "index-splash.md"),
		Order:       6,
		Offline:     true,
		Badge:       docs.NavBadge{Label: "es", Tone: docs.NavBadgeToneAccent},
	})
	router.MustPage("/es/docs/getting-started", docs.PageConfig{
		Title:       "Primeros pasos",
		Description: "Crea tu primer proyecto y entiende qué se traduce.",
		SourcePath:  siteFile("content", "es", "getting-started.md"),
		Order:       1,
		Offline:     true,
	})
	router.MustPage("/es/docs/concepts/router", docs.PageConfig{
		Title:       "El Router",
		Description: "El objeto central: un solo árbol de rutas.",
		SourcePath:  siteFile("content", "es", "concepts-router.md"),
		Order:       2,
		Offline:     true,
	})

	if err := router.MarkdownBlog("/blog", siteFile("content", "blog"), docs.BlogConfig{
		Title: "Blog", Description: "Release notes and implementation updates for fastr-docs.", Order: 5,
		PostsPerPage: 10, RelatedPosts: 3,
		Offline: true,
	}); err != nil {
		panic(err)
	}

	operate := router.MustGroup("/docs/operate", docs.GroupConfig{
		Title:       "Operate",
		Description: "Search, ship, and verify your documentation site.",
		Order:       4,
	})
	operate.MustPage("search", docs.PageConfig{
		Title:       "Search",
		Description: "Make every route discoverable with a local index.",
		SourcePath:  contentFile("operate-search.md"),
		Order:       1,
		Offline:     true,
	})
	operate.MustPage("offline", docs.PageConfig{
		Title:       "Offline and PWA",
		Description: "Keep static content available when the network disappears.",
		SourcePath:  contentFile("operate-offline.md"),
		Order:       2,
		Offline:     true,
	})
	operate.MustPage("testing", docs.PageConfig{
		Title:       "Testing",
		Description: "Verify behavior at the router, browser, and export layers.",
		SourcePath:  contentFile("operate-testing.md"),
		Order:       3,
		Offline:     true,
	})
	operate.MustPage("deploy", docs.PageConfig{
		Title:       "Deploy and customize",
		Description: "Configure assets, hosts, and static deployment paths.",
		SourcePath:  contentFile("operate-deploy.md"),
		Order:       4,
		Offline:     true,
	})
	operate.MustPage("i18n", docs.PageConfig{
		Title:       "Languages",
		Description: "Translate content and chrome, and keep a partial translation usable.",
		SourcePath:  contentFile("operate-i18n.md"),
		Order:       5,
		Offline:     true,
		Badge:       docs.NavBadge{Label: "New", Tone: docs.NavBadgeToneInfo},
	})
	operate.MustPage("assets", docs.PageConfig{
		Title:       "Runtime assets",
		Description: "Collect the runtime, search index, and plugin assets in one call.",
		SourcePath:  contentFile("operate-assets.md"),
		Order:       6,
		Offline:     true,
	})
	operate.MustPage("feature-coverage", docs.PageConfig{
		Title:       "Feature coverage",
		Description: "See what fastr-docs supports and which tradeoffs are intentional.",
		SourcePath:  contentFile("feature-coverage.md"),
		Order:       7,
		Offline:     true,
	})

	collaborate := router.MustGroup("/docs/collaborate", docs.GroupConfig{
		Title:       "Collaborate",
		Description: "Keep humans and AI agents aligned with the route tree.",
		Order:       5,
	})
	collaborate.MustPage("ai-authoring", docs.PageConfig{
		Title:       "AI authoring",
		Description: "Give agents safe, explicit project conventions.",
		SourcePath:  contentFile("collaborate-ai-authoring.md"),
		Order:       1,
		Offline:     true,
	})

	examples := router.MustGroup("/examples", docs.GroupConfig{
		Title:       "Examples",
		Description: "See pages, screens, and plugins working together.",
		Order:       4,
		Badge:       docs.NavBadge{Label: "Explore", Tone: docs.NavBadgeToneSuccess},
	})
	examples.MustScreen("playground", docs.ScreenConfig{
		Title:       "Playground",
		Description: "Exercise GoFastr components inside a typed docs screen.",
		Component:   &playgroundScreen{},
		Order:       1,
		Offline:     true,
		Badge:       docs.NavBadge{Label: "Live", Tone: docs.NavBadgeToneSuccess},
	})
	examples.MustPage("route-tree", docs.PageConfig{
		Title:       "Route tree",
		Description: "A visual guide to how one tree drives the site.",
		SourcePath:  contentFile("examples-route-tree.md"),
		Order:       2,
		Offline:     true,
	})
	examples.MustPage("custom-surface", docs.PageConfig{
		Title:       "Custom surfaces",
		Description: "A checklist for turning a product feature into a docs surface.",
		SourcePath:  contentFile("examples-custom-surface.md"),
		Order:       3,
		Offline:     true,
	})
	// The hand-built landing surface, kept as the worked example of the typed
	// route for anyone who outgrows the front-matter splash template.
	examples.MustPage("typed-landing", docs.PageConfig{
		Title:       "Typed landing page",
		Description: "A landing surface built as Go rather than front matter.",
		Body:        func() render.HTML { return HomeBody(router) },
		Order:       4,
		Offline:     true,
		DisableTOC:  true,
	})
	if err := router.Use(openapi.Plugin{
		SpecPath:    contractFile(),
		Path:        "/api-reference",
		Title:       "Example API reference",
		Description: "A small contract used to exercise the in-project OpenAPI plugin.",
		ServerURL:   os.Getenv("API_SERVER_URL"),
		Order:       3,
		Badge:       docs.NavBadge{Label: "Demo", Tone: docs.NavBadgeToneNeutral},
	}); err != nil {
		panic(err)
	}
	return router
}

// Mount is deliberately tiny: the docs Router remains the source of truth,
// while GoFastr owns lifecycle, DI, policies, rendering, and the PWA shell.
func Mount(site *uiapp.App) error {
	router := NewRouter()
	return router.Mount(site, router.Layout())
}

// OpenAPICSS returns the bundled reference surface CSS. Projects can replace
// it entirely or append their brand overrides in the host setup.
func OpenAPICSS() string { return openapi.CSS() }
