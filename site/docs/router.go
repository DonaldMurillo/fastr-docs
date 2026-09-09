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
		// The chrome is translated per locale, so a Spanish page is not wrapped
		// in English furniture. Only the labels a reader of the translated
		// pages will actually meet are set here; the rest fall back to English,
		// which is what a partly translated site looks like.
		docs.WithLocaleNames(map[string]string{"en": "English", "es": "Español"}),
		docs.WithLocaleUIStrings("es", docs.UIStrings{
			Contents:          "Contenido",
			Home:              "Inicio",
			SkipToContent:     "Saltar al contenido principal",
			SectionHelp:       "Salta a una sección de primer nivel.",
			AnchorLabel:       "Copiar enlace a esta sección",
			Sections:          "Secciones",
			OnThisPage:        "En esta página",
			Search:            "Buscar",
			SearchPlaceholder: "Buscar en la documentación",
			OpenSearch:        "Abrir la búsqueda",
			CloseSearch:       "Cerrar la búsqueda",
			OpenNavigation:    "Abrir la navegación",
			EditPage:          "Editar esta página",
			LastUpdated:       "Actualizado",
			Published:         "Publicado",
			By:                "por",
			Language:          "Idioma",
			Previous:          "← Anterior",
			Next:              "Siguiente →",
			Version:           "Versión",
			DateFormat:        "2 de January de 2006",
			Months:            []string{"enero", "febrero", "marzo", "abril", "mayo", "junio", "julio", "agosto", "septiembre", "octubre", "noviembre", "diciembre"},
			Blog: docs.BlogStrings{
				Title:                  "Blog",
				Navigation:             "Navegación del blog",
				Breadcrumb:             "Ruta de navegación",
				AllPosts:               "Todas las entradas",
				LatestPosts:            "Últimas entradas",
				RecentPosts:            "Entradas recientes",
				KeepReading:            "Sigue leyendo",
				NoPostsYet:             "Todavía no hay entradas publicadas aquí.",
				NothingClassified:      "Todavía no hay nada clasificado.",
				CollectionDescription:  "Entradas publicadas en esta colección.",
				Search:                 "Buscar",
				SearchDescription:      "Busca entre las entradas publicadas.",
				SearchPosts:            "Buscar entradas",
				SearchThePublication:   "Buscar en la publicación",
				ResultsFor:             "Resultados para “%s”",
				NoPostsMatched:         "Ninguna entrada coincide con esa búsqueda.",
				MatchesSummary:         "%d coincidencias",
				Archive:                "Archivo",
				ArchiveDescription:     "Recorre todas las entradas publicadas por año.",
				ArchiveYear:            "Archivo de %s",
				ArchiveYearDescription: "Entradas publicadas en %s.",
				Tag:                    "Etiqueta",
				Tags:                   "Etiquetas",
				TagsDescription:        "Recorre las entradas por tema.",
				ExploreTerms:           "Explora %s en toda la publicación.",
				PostCount:              "%d entradas",
				Publication:            "Publicación",
				ReadingTime:            "%s min de lectura",
				Featured:               "Destacada",
				Latest:                 "Últimas",
				Feed:                   "RSS",
				TopicsPrefix:           "Temas: ",
				Lede:                   "Notas de versión, ensayos y cambios de implementación.",
				TagDescription:         "Entradas etiquetadas con %s.",
				Authors:                "Autores",
				AuthorsDescription:     "Recorre las entradas por autor.",
				Author:                 "Autor",
				AuthorDescription:      "Entradas de %s.",
				Page:                   "Página %d",
				PagedTitle:             "Blog · Página %d",
				PageDescription:        "Más entradas de %s.",
				OlderPosts:             "Entradas anteriores →",
				NewerPosts:             "← Entradas más recientes",
				Pagination:             "Paginación del blog",
				PostActions:            "Acciones de la entrada",
				Share:                  "Compartir",
				ShareThisPost:          "Compartir esta entrada",
				CopyLink:               "Copiar enlace",
				LinkCopied:             "Enlace copiado",
			},
			NotFound: docs.NotFoundStrings{
				Heading:       "Página no encontrada",
				Message:       "Esta página no existe.",
				MessageForURL: "Ninguna página coincide con %s.",
				BackTo:        "Volver a %s",
				SiteFallback:  "la documentación",
			},
		}),
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
		Description: "Turn an API spec into a searchable request console.",
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
	// The Spanish tree mirrors the English one segment for segment, groups
	// included, so variantFamily pairs every page with its original and the
	// language selector always has somewhere to go.
	router.MustPage("/es", docs.PageConfig{
		Title:       "Español",
		Description: "Documentación construida sobre un árbol de rutas.",
		SourcePath:  siteFile("content", "es", "index-splash.md"),
		Order:       6,
		Offline:     true,
	})
	router.MustPage("/es/docs", docs.PageConfig{
		Title:       "Documentación",
		Description: "Aprende el modelo de rutas, escribe contenido y publica.",
		SourcePath:  siteFile("content", "es", "index.md"),
		Order:       1,
		Offline:     true,
	})
	router.MustPage("/es/docs/getting-started", docs.PageConfig{
		Title:       "Primeros pasos",
		Description: "Crea tu primer proyecto y entiende qué se traduce.",
		SourcePath:  siteFile("content", "es", "getting-started.md"),
		Order:       2,
		Offline:     true,
	})
	conceptosES := router.MustGroup("/es/docs/concepts", docs.GroupConfig{
		Title:       "Conceptos",
		Description: "El modelo de rutas y sus primitivas de composición.",
		Order:       3,
		Locale:      "es",
	})
	conceptosES.MustPage("router", docs.PageConfig{
		Title:       "El Router",
		Description: "El objeto central: un solo árbol de rutas.",
		SourcePath:  siteFile("content", "es", "concepts-router.md"),
		Order:       1,
		Offline:     true,
	})
	conceptosES.MustPage("content", docs.PageConfig{
		Title:       "Escribir contenido",
		Description: "Markdown con front matter, y el campo que enlaza una traducción.",
		SourcePath:  siteFile("content", "es", "concepts-content.md"),
		Order:       2,
		Offline:     true,
	})
	conceptosES.MustPage("layouts", docs.PageConfig{
		Title:       "Layouts y navegación",
		Description: "Compón envolturas globales y barras laterales por sección.",
		SourcePath:  siteFile("content", "es", "concepts-layouts.md"),
		Order:       3,
		Offline:     true,
	})
	construirES := router.MustGroup("/es/docs/build", docs.GroupConfig{
		Title:       "Construir",
		Description: "Añade pantallas, referencias de API y extensiones.",
		Order:       4,
		Locale:      "es",
	})
	construirES.MustPage("screens", docs.PageConfig{
		Title:       "Pantallas y componentes",
		Description: "Usa pantallas tipadas cuando Markdown no basta.",
		SourcePath:  siteFile("content", "es", "build-screens.md"),
		Order:       1,
		Offline:     true,
	})
	construirES.MustPage("framework-ui", docs.PageConfig{
		Title:       "UI del framework",
		Description: "Los componentes nativos de GoFastr en páginas y pantallas.",
		SourcePath:  siteFile("content", "es", "build-ui.md"),
		Order:       2,
		Offline:     true,
	})
	construirES.MustPage("openapi", docs.PageConfig{
		Title:       "Referencia OpenAPI",
		Description: "Convierte una especificación de API en una consola de peticiones.",
		SourcePath:  siteFile("content", "es", "build-openapi.md"),
		Order:       3,
		Offline:     true,
	})
	construirES.MustPage("plugins", docs.PageConfig{
		Title:       "Plugins y extensiones",
		Description: "Extiende el mismo router sin bifurcar la envoltura.",
		SourcePath:  siteFile("content", "es", "build-plugins.md"),
		Order:       4,
		Offline:     true,
	})
	construirES.MustPage("blog", docs.PageConfig{
		Title:       "Blog y RSS",
		Description: "Publica novedades en Markdown y expónlas como RSS.",
		SourcePath:  siteFile("content", "es", "build-blog.md"),
		Order:       5,
		Offline:     true,
	})
	construirES.MustPage("themes", docs.PageConfig{
		Title:       "Temas y plantillas",
		Description: "Elige un punto de partida visual y ajusta sus tokens.",
		SourcePath:  siteFile("content", "es", "build-themes.md"),
		Order:       6,
		Offline:     true,
	})
	construirES.MustPage("components", docs.PageConfig{
		Title:       "Componentes de Markdown",
		Description: "El vocabulario de shortcodes con el que arranca todo proyecto.",
		SourcePath:  siteFile("content", "es", "build-components.md"),
		Order:       7,
		Offline:     true,
	})
	construirES.MustPage("diagrams", docs.PageConfig{
		Title:       "Diagramas",
		Description: "Dibuja diagramas de Mermaid sin debilitar la política de contenido.",
		SourcePath:  siteFile("content", "es", "build-diagrams.md"),
		Order:       8,
		Offline:     true,
	})
	construirES.MustPage("math", docs.PageConfig{
		Title:       "Matemáticas",
		Description: "Escribe TeX con dólares y renderízalo en la página.",
		SourcePath:  siteFile("content", "es", "build-math.md"),
		Order:       9,
		Offline:     true,
	})
	operarES := router.MustGroup("/es/docs/operate", docs.GroupConfig{
		Title:       "Operar",
		Description: "Busca, publica y verifica tu sitio de documentación.",
		Order:       5,
		Locale:      "es",
	})
	operarES.MustPage("search", docs.PageConfig{
		Title:       "Búsqueda",
		Description: "Haz que cada ruta se pueda encontrar con un índice local.",
		SourcePath:  siteFile("content", "es", "operate-search.md"),
		Order:       1,
		Offline:     true,
	})
	operarES.MustPage("offline", docs.PageConfig{
		Title:       "Sin conexión y PWA",
		Description: "Mantén el contenido estático disponible sin red.",
		SourcePath:  siteFile("content", "es", "operate-offline.md"),
		Order:       2,
		Offline:     true,
	})
	operarES.MustPage("testing", docs.PageConfig{
		Title:       "Pruebas",
		Description: "Verifica el comportamiento en el router, el navegador y la exportación.",
		SourcePath:  siteFile("content", "es", "operate-testing.md"),
		Order:       3,
		Offline:     true,
	})
	operarES.MustPage("deploy", docs.PageConfig{
		Title:       "Desplegar y personalizar",
		Description: "Configura recursos, hosts y rutas de despliegue estático.",
		SourcePath:  siteFile("content", "es", "operate-deploy.md"),
		Order:       4,
		Offline:     true,
	})
	operarES.MustPage("i18n", docs.PageConfig{
		Title:       "Idiomas y traducción",
		Description: "Traduce el contenido y la interfaz, y deja usable una traducción parcial.",
		SourcePath:  siteFile("content", "es", "operate-i18n.md"),
		Order:       5,
		Offline:     true,
	})
	operarES.MustPage("assets", docs.PageConfig{
		Title:       "Recursos de ejecución",
		Description: "Reúne el runtime, el índice y los recursos de plugin en una llamada.",
		SourcePath:  siteFile("content", "es", "operate-assets.md"),
		Order:       6,
		Offline:     true,
	})
	operarES.MustPage("feature-coverage", docs.PageConfig{
		Title:       "Qué cubre",
		Description: "Mira qué soporta fastr-docs y qué compromisos son deliberados.",
		SourcePath:  siteFile("content", "es", "feature-coverage.md"),
		Order:       7,
		Offline:     true,
	})
	colaborarES := router.MustGroup("/es/docs/collaborate", docs.GroupConfig{
		Title:       "Colaborar",
		Description: "Mantén alineados a las personas y a los agentes.",
		Order:       6,
		Locale:      "es",
	})
	colaborarES.MustPage("ai-authoring", docs.PageConfig{
		Title:       "Escribir con agentes",
		Description: "Dale a los agentes convenciones explícitas y seguras.",
		SourcePath:  siteFile("content", "es", "collaborate-ai-authoring.md"),
		Order:       1,
		Offline:     true,
	})
	// The examples are translated slug and all, so their paths do not mirror
	// the English ones. The group names its original here, and each page
	// names its own in front matter with translation_of. The rest of the
	// Spanish tree pairs by path shape; a site can use both.
	ejemplosES := router.MustGroup("/es/ejemplos", docs.GroupConfig{
		Title:         "Ejemplos",
		Description:   "Superficies trabajadas dentro de este mismo Router.",
		Order:         7,
		Locale:        "es",
		TranslationOf: "/examples",
	})
	ejemplosES.MustPage("arbol-de-rutas", docs.PageConfig{
		Title:       "Ejemplo de árbol de rutas",
		Description: "Una guía visual de cómo un solo árbol mueve el sitio.",
		SourcePath:  siteFile("content", "es", "examples-route-tree.md"),
		Order:       1,
		Offline:     true,
	})
	ejemplosES.MustPage("superficies-propias", docs.PageConfig{
		Title:       "Superficies propias",
		Description: "Una lista para convertir una función del producto en documentación.",
		SourcePath:  siteFile("content", "es", "examples-custom-surface.md"),
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
	// The Spanish blog is a second collection in its own language. Its
	// landing, archive, tags, authors and search pair with the English ones by
	// path shape, and so does a post that keeps its file name; a post with a
	// translated slug names its original with translation_of.
	if err := router.MarkdownBlog("/es/blog", siteFile("content", "es", "blog"), docs.BlogConfig{
		Title: "Blog", Description: "Notas de versión y cambios de implementación de fastr-docs.", Order: 8,
		PostsPerPage: 10, RelatedPosts: 3,
		DefaultLocale: "es",
		Offline:       true,
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
		SpecPath:    specFile(),
		Path:        "/api-reference",
		Title:       "Example API reference",
		Description: "A small spec used to exercise the in-project OpenAPI plugin.",
		ServerURL:   os.Getenv("API_SERVER_URL"),
		Order:       3,
		Badge:       docs.NavBadge{Label: "Demo", Tone: docs.NavBadgeToneNeutral},
	}); err != nil {
		panic(err)
	}
	if err := router.Use(openapi.Plugin{
		SpecPath:    specFileES(),
		ServerURL:   os.Getenv("API_SERVER_URL"),
		Path:        "/es/api-reference",
		Title:       "Referencia de la API de ejemplo",
		Description: "La misma especificación pequeña, traducida: cada montura del plugin declara su idioma y sus etiquetas.",
		Locale:      "es",
		Order:       3,
		Badge:       docs.NavBadge{Label: "Demo", Tone: docs.NavBadgeToneNeutral},
		Strings: openapi.Strings{
			Eyebrow:              "Referencia OpenAPI",
			OperationsLabel:      "operaciones",
			SchemasLabel:         "esquemas",
			NoServer:             "Sin servidor configurado",
			FilterPlaceholder:    "Filtrar endpoints…",
			TryRequest:           "Prueba una petición",
			ConsoleNoServer:      "Sin URL de servidor OpenAPI configurada.",
			OperationLabel:       "Operación",
			ServerLabel:          "Servidor",
			TokenPlaceholder:     "Token bearer",
			DeprecatedLabel:      "Obsoleta",
			ResponseHeadersLabel: "Cabeceras de respuesta",
			CopyCurlLabel:        "Copiar como cURL",
			SendRequest:          "Enviar la petición",
			ResponsePrompt:       "Elige una operación y envía una petición.",
			NoOperations:         "Esta especificación no tiene operaciones.",
			CORSNote:             "Las peticiones salen del navegador y necesitan que el servidor de la API permita CORS.",
			NoInputs:             "Esta operación no tiene entradas.",
			OperationIDPrefix:    "ID de operación: ",
			ParametersLabel:      "Parámetros",
			RequestBodyLabel:     "Cuerpo de la petición",
			RequestBodyNote:      "La especificación exige cuerpo de petición.",
			ResponseLabel:        "Respuesta",
			ValuePlaceholder:     "Valor",
			ParameterWord:        "parámetro",
			RequiredWord:         "obligatorio",
		},
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
