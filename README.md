# fastr-docs

Reusable, white-label documentation sites built on GoFastr.

The central API is `docs.Router`. It owns the route tree, explicit sibling
order, navigation metadata, local-search records, Markdown pages, typed
GoFastr screens, and plugins such as OpenAPI. GoFastr remains responsible for
screen rendering, DI, policies, lifecycle, and the PWA/offline shell. The
default layout includes router-derived navigation, a sticky Markdown table of
contents with scrollspy, and a local search index. Static builds can run
Pagefind as a first-class export backend while JSON remains the zero-dependency
development fallback.

## Create a site

From this repository while developing the package:

```sh
go run ./cmd/fastr-docs init my-docs --name "Acme Docs" --module example.com/acme-docs
cd my-docs
fastr-docs check .
go run .
# or export a static PWA with chunked Pagefind search
fastr-docs export . --out dist --pagefind
```

After publishing the module, the CLI is intended to be installed with:

```sh
go install github.com/DonaldMurillo/fastr-docs/cmd/fastr-docs@latest
```

The generated project exposes the same workflow through the CLI:

```sh
fastr-docs check .
fastr-docs build .
fastr-docs export . --out dist
```

The generated project includes two top-level pages, a nested getting-started
page, the in-project OpenAPI plugin, `agents/claude.md`, and a docs-authoring
skill reference. It also emits an installable PWA/static export with the
OpenAPI and docs runtimes plus the router search index precached. The host
publishes GoFastr's agent-ready `/llms.txt` and agent-card discovery surfaces
alongside the page-level Markdown references.

## Registering content

```go
router := docs.NewRouter(docs.WithSiteName("Acme Docs"))
router.MustPage("/", docs.PageConfig{
    Title: "Acme Docs",
    Source: "# Welcome\n\nYour documentation starts here.",
    Order: 1,
    Offline: true,
})
router.MustScreen("/playground", docs.ScreenConfig{
    Title: "Playground",
    Component: &PlaygroundScreen{},
    Order: 2,
})
```

Use `router.Mount(site, layout)` to register the tree into a GoFastr
`core-ui/app.App`. Plugins implement `Name` and `Apply`, so extensions stay in
the same navigation and search model instead of creating a parallel system.

When serving through a GoFastr HTTP host, mount the framework-owned mobile
navigation drawer and native command palette once after creating the host:

```go
router.MountNavigation(server.Router())
router.MountCommandPalette(server.Router())
```

The layout uses GoFastr `SiteHeader`, `Sidebar`, `DocLayout`, `AnchoredRail`,
`CommandPalette`, and `ShortcutHint` components. The primary header follows
top-level route groups, while `Meta/Control + K` opens the native command
palette for route navigation. Their drawer behavior, active route state, and
scrollspy should be extended through GoFastr components rather than replaced
with page-specific JavaScript.

The first-party OpenAPI plugin reads the contract's `servers` URL and exposes
an optional request console. Set `ServerURL` to override it per deployment;
the generated starter maps that override to `API_SERVER_URL`.

```go
router.Use(openapi.Plugin{
    SpecPath:  "openapi.json",
    ServerURL: os.Getenv("API_SERVER_URL"),
})
```

## Content components and plugins

Markdown can embed typed GoFastr components with a small shortcode syntax:

```go
router.Use(docs.MarkdownComponentsPlugin{Components: map[string]docs.MarkdownComponent{
    "note": func(props map[string]string, body render.HTML) render.HTML {
        return ui.Callout(ui.CalloutConfig{Title: props["title"], Variant: ui.StatusInfo}, body)
    },
}})
```

```md
{{< note title="Important" >}}
This body is still Markdown and can contain links and emphasis.
{{< /note >}}
```

`MarkdownCollectionPlugin` packages filesystem content discovery for projects
that want a plugin-owned collection. Other extensions implement `Plugin`, and
may additionally implement `ValidatingPlugin` or `AssetPlugin`; route content,
search metadata, origins, validation, and assets should remain attached to the
same `Router`.

When a project contains localized or versioned route siblings, the header emits
route-aware Language and Version selectors that preserve the current page
family.

Use `WithPagefindPath` when the generated search bundle is served from a
custom asset prefix, such as a reverse-proxy mount:

```go
router := docs.NewRouter(docs.WithPagefind(), docs.WithPagefindPath("/assets/pagefind/"))
```

## End-to-end tests

The Playwright suite generates a fresh starter project for every run and
covers desktop/mobile navigation, search, theme persistence, sticky TOC,
OpenAPI server requests, static export, and PWA assets:

```sh
cd e2e
npm install
npx playwright install chromium
npm test
```
