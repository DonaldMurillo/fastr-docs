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
fastr-docs dev .
# run once without the watcher
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
fastr-docs dev .
fastr-docs upgrade .
fastr-docs export . --out dist
fastr-docs sync-skills .
```

Generated projects carry a set of agent skills, authored under
`.agents/skills/` and mirrored to `.claude/skills/` where Claude Code loads
them. `fastr-docs check` fails when the two copies disagree, naming each file
that is stale, missing, or orphaned, so an edit cannot land in one and leave
the other agent following older rules. `fastr-docs sync-skills` copies the
authored version across; add `--prune` to also delete files that exist only
under `.claude/skills`.

`fastr-docs dev` delegates to GoFastr's development loop. It watches Go,
Markdown, HTML, CSS, JavaScript, and JSON/YAML contract files, rebuilds the
server, and refreshes open browser tabs after a successful rebuild. Use
`go run .` when you want a single process without file watching.

`fastr-docs build` delegates to GoFastr's build gates and compiler. Use
`fastr-docs upgrade . --apply` to apply GoFastr framework migrations through
the same project wrapper.

Generated sites are installable PWAs through GoFastr's `WithPWA` host option.
Deploy the static export when the installed app must include every document
offline:

```sh
fastr-docs export . --out dist --pagefind
```

GoFastr fingerprints the exported worker cache and updates it on the next
visit after a redeploy. Existing tabs finish on their current version; new
tabs use the updated docs. Use `fastr-docs upgrade .` to review and apply
GoFastr framework and CLI migrations when updating a generated project.

The generated project includes two top-level pages, a nested getting-started
page, the in-project OpenAPI plugin, `agents/claude.md`, and skills covering
authoring, blogging, theming, OpenAPI, and publishing. It also emits an
installable PWA/static export with the
OpenAPI and docs runtimes plus the router search index precached. The live
host publishes GoFastr's agent-ready `/llms.txt`, agent card, and `/mcp`
discovery surfaces alongside the page-level Markdown references. Static
exports retain the documentation surfaces; `/mcp` is a live server endpoint.

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

The first-party OpenAPI plugin accepts JSON or YAML, reads the contract's
`servers` URL, and exposes an optional request console. Set `ServerURL` to
override it per deployment;
the generated starter maps that override to `API_SERVER_URL`.

```go
router.Use(openapi.Plugin{
    SpecPath:  "openapi.json",
    ServerURL: os.Getenv("API_SERVER_URL"),
})
```

Generated hosts also enable GoFastr's read-only MCP introspection surface:
`framework.WithMCP()` mounts `/mcp`, and
`framework.WithMCPIntrospection()` exposes tools that describe the running
app. The generated agent card advertises the same endpoint. Remove both
options and `AgentCard.MCPEndpoint` together when a deployment must not expose
an MCP endpoint.

## Content components and plugins

Every Router registers a shortcode vocabulary by default, so Markdown authors
get components without writing Go:

```md
{{< warning title="Important" >}}
This body is still Markdown and can contain links and emphasis.
{{< /warning >}}
```

`note`, `info`, `tip`, `success`, `warning`, `caution`, and `danger` are
admonitions; `callout` takes the tone as a prop. `cards` lays out `card`
children, `tabs` holds `tab` children, and `hero` holds `action` children.
`steps`, `filetree`, `details`, `badge`, `tag`, `banner`, `icon`, and `diff`
round it out.

Shortcodes come in three shapes. `MarkdownComponent` receives its body as
rendered Markdown. `MarkdownContainer` receives its nested shortcodes as
separate children, which is what tabs and card grids need. `MarkdownRawComponent`
receives the body unrendered, for content that is data rather than prose:
`diff` and `filetree` both read line structure that Markdown rendering would
destroy.

Registering a name replaces whatever held it:

```go
router.Use(docs.MarkdownComponentsPlugin{Components: map[string]docs.MarkdownComponent{
    "note": func(props map[string]string, body render.HTML) render.HTML {
        return ui.Callout(ui.CalloutConfig{Title: props["title"], Variant: ui.StatusNeutral}, body)
    },
}})
```

`docs.WithoutDefaultComponents()` starts from an empty vocabulary, where any
unregistered shortcode fails the build instead of rendering.

`MarkdownCollectionPlugin` packages filesystem content discovery for projects
that want a plugin-owned collection. Use `MarkdownCollectionFS` for `embed.FS`
or another virtual filesystem, and `MarkdownCollection` for disk-backed
development content. Other extensions implement `Plugin`, and
may additionally implement `ValidatingPlugin` or `AssetPlugin`; route content,
search metadata, origins, validation, and assets should remain attached to the
same `Router`.

When a project contains localized or versioned route siblings, the header emits
route-aware Language and Version selectors that preserve the current page
family.

## Dates and edit links from git

`edit_url` and `last_updated` can both be derived from history instead of being
typed into every page and left to rot:

```go
router := docs.NewRouter(docs.WithGitMetadata(docs.GitMetadataConfig{
    RepoURL: "https://github.com/acme/docs",
    Branch:  "main",
}))
```

Front matter always wins, so a page that declares either keeps what it
declared. One `git log` covers the whole repository rather than one process per
page. Pages served from an `embed.FS` have no file on disk and are skipped.

It degrades silently by design: no git binary, no repository, or a shallow
clone leaves every page exactly as authored, because a documentation build must
not require version-control history to be present. `EditURLFor` replaces the
built-in GitHub-style link for hosts that shape edit URLs differently.

Use `WithPagefindPath` when the generated search bundle is served from a
custom asset prefix, such as a reverse-proxy mount:

```go
router := docs.NewRouter(docs.WithPagefind(), docs.WithPagefindPath("/assets/pagefind/"))
```

`WithSearchIndexPath` does the same for the JSON search index. The default,
`/__fastr-docs/search.json`, matches the prefix the generated starter mounts.
A host that serves the docs assets somewhere else must set it, or the command
palette falls back to an unfiltered route list:

```go
router := docs.NewRouter(docs.WithSearchIndexPath("/assets/search.json"))
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
