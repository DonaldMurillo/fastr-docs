---
tags: [reference, features, architecture]
---

# Feature coverage

fastr-docs covers the core jobs expected from a documentation generator:
author content, organize it into navigable sections, search it, publish it,
and keep the same route model available to humans and agents.

## Supported in the Router

| Capability | fastr-docs implementation |
| --- | --- |
| Content hierarchy | One `docs.Router` with nested groups, pages, typed screens, and plugin routes. |
| Markdown authoring | Filesystem Markdown, embedded `fs.FS`, front matter, custom slugs, links, code blocks, and registered shortcodes. |
| Interactive content | Native GoFastr screens and components on the same route tree as Markdown. |
| Layout composition | Global and section layouts, including layouts nested inside other layouts through GoFastr. |
| Navigation | Explicit sibling `Order`, active ancestor state, breadcrumbs, previous/next links, responsive drawers, and in-page heading navigation. |
| Search | JSON index during development, Pagefind export support, and GoFastr's native `Ctrl+K` or `⌘K` command palette. |
| Localization | Locale metadata, alternate page links, locale filters, localized UI strings, and a route manifest inventory. |
| Documentation versions | Version metadata, version filters, route-aware selectors, `MarkdownVersionedCollection`, and a version inventory in the export manifest. |
| SEO and publishing | Titles, descriptions, canonical URLs, `noindex`, authors, published/updated dates, social images, redirects, sitemap, robots rules, Markdown blogs, and RSS feeds. |
| Offline delivery | GoFastr PWA integration and static export of offline-eligible routes and assets. |
| API references | In-project OpenAPI plugin with JSON or YAML specs and an optional server URL override. |
| Extensibility | Plugins, Markdown component adapters, custom layout factories, assets, validation, and search contributions. |
| Agent workflows | Generated `agents/claude.md`, authoring guidance, `/llms.txt`, agent card, and optional MCP discovery. |
| Quality gates | Strict route/content validation, CLI checks, static export checks, and desktop/mobile browser tests. |

## Extend the generated site

Use `WithUIStrings` for translated framework labels, `WithLayouts` for custom
global or section shells, and `PageConfig.Preload` or `ScreenConfig.Preload`
for GoFastr's `hover`, `visible`, and `eager` route loading modes. Typed
screens retain their GoFastr loader, static-path, action, component ID, and
custom head HTML capabilities.

Install `NotFoundScreen` with GoFastr's host to give missing routes a branded
recovery page:

```go
host := uihost.New(site,
    uihost.WithNotFoundScreen(docs.NotFoundScreen{
        SiteName: router.SiteName(),
    }),
)
server := framework.NewUIHostApp(host)
```

The live host returns a branded 404 with the requested path and a link home.
Static exports can call `docs.WriteStaticNotFound` after `ExportStatic`; the
generated starter already writes this `404.html` file.

Use `MarkdownVersionedCollection` when each release has its own content
directory. The current version keeps the normal route family while archived
versions receive a version path segment. `MarkdownVersionedCollectionFS` gives
embedded sites the same behavior.

## Intentional tradeoffs

Docusaurus and Starlight are useful references for Markdown, search,
localization, versioning, and plugins. fastr-docs keeps those expectations in
the GoFastr runtime. It uses shortcodes and typed Go screens instead of a
React, MDX, or Markdoc runtime, and it keeps analytics and vendor integrations
in project plugins.

Version metadata, selectors, and versioned collections are supported today.
Snapshot creation stays in the content or VCS workflow so the framework does
not rewrite arbitrary Go route code.

The starter and the non-CLI fixture are both tested through real browser flows.
Keep both paths covered when changing the public Router API.
