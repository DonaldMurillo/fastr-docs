---
tags: [reference, features, architecture]
---

# Feature coverage

fastr-docs covers the core jobs people expect from a documentation generator:
author content, organize it into navigable sections, search it, publish it,
and keep the same route model available to humans and agents.

This page explains what is supported, where the implementation deliberately
differs from JavaScript-first documentation tools, and which choices belong to
the project using fastr-docs.

## Supported in the Router

| Capability | fastr-docs implementation |
| --- | --- |
| Content hierarchy | One `docs.Router` with nested groups, pages, typed screens, and plugin routes. |
| Markdown authoring | Filesystem Markdown, embedded `fs.FS`, front matter, custom slugs, links, and a shortcode vocabulary registered by default. |
| Interactive content | Native GoFastr screens and components on the same route tree as Markdown. |
| Layout composition | Global and section layouts, including layouts nested inside other layouts through GoFastr. |
| Navigation | Explicit sibling `Order`, active ancestor state, breadcrumbs, previous/next links, responsive drawers, and in-page heading navigation. |
| Search | JSON index during development, Pagefind export support, and GoFastr's native `Ctrl+K` or `⌘K` command palette. |
| Localization | Locale metadata, pairing by path or by `translation_of`, derived `hreflang` alternates, locale filters, a fully translatable UI string surface, configurable date formats, fallback to a default locale, and translation-coverage reporting. |
| Documentation versions | Version metadata, version filters, route-aware selectors, `MarkdownVersionedCollection`, and a version inventory in the export manifest. |
| SEO and publishing | Titles, descriptions, canonical URLs, `noindex`, authors, published/updated dates, social images, redirects, sitemap, robots rules, Markdown blogs, and RSS feeds. |
| Offline delivery | GoFastr PWA integration and static export of offline-eligible routes and assets. |
| API references | In-project OpenAPI plugin with JSON or YAML contracts and an optional server URL override. |
| Extensibility | Plugins, three shortcode shapes, custom layout factories, validation, search contributions, and a collected runtime-asset pipeline. |
| Page templates | A front-matter `splash` shell for landing pages, alongside typed screens for anything it cannot express. |
| Code blocks | Fence options for a title, line numbers, line highlighting, and internal scrolling. |
| Repository metadata | `last_updated` and `edit_url` derived from git history, with front matter overriding. |
| Agent workflows | Generated `agents/claude.md`, authoring guidance, `/llms.txt`, agent card, and optional MCP discovery. |
| Quality gates | Strict route/content validation, CLI checks, static export checks, and desktop/mobile browser tests. |

## The APIs that shape a project

The defaults are useful for a generated site, but the public API is intended
to be extended by the project that owns the content.

### Localize the framework chrome

`WithUIStrings` translates labels supplied by the docs shell without changing
route titles or content metadata:

```go
router := docs.NewRouter(
    docs.WithUIStrings(docs.UIStrings{
        Contents: "Contenu",
        Search:   "Rechercher",
        Language: "Langue",
    }),
)
```

### Preserve GoFastr route capabilities

`PageConfig.Preload` and `ScreenConfig.Preload` accept GoFastr's `hover`,
`visible`, and `eager` modes. Typed screens can keep their GoFastr loader,
static-path, action, custom component ID, and head HTML capabilities when the
Router adds documentation metadata.

### Customize the shell

`WithLayouts` lets a project provide global and section layout factories while
the Router continues to own route selection and documentation metadata:

```go
router := docs.NewRouter(docs.WithLayouts(docs.LayoutConfig{
    Global:  buildGlobalLayout,
    Section: buildSectionLayout,
}))
```

### Make missing routes useful

Install `NotFoundScreen` in the GoFastr host, or replace it with a branded
screen that follows the same contract:

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

### Publish versioned snapshots

Use `MarkdownVersionedCollection` when each release has its own content
directory. The current version keeps the normal route family while archived
versions receive a version path segment. `MarkdownVersionedCollectionFS` gives
embedded sites the same behavior.

## How this differs from Docusaurus and Starlight

Docusaurus and Starlight are strong references for information architecture,
Markdown authoring, search, localization, versioning, and extension points.
fastr-docs keeps those product expectations where they fit the GoFastr runtime,
but does not copy every ecosystem feature.

| Area | fastr-docs today | Product decision |
| --- | --- | --- |
| Docs sidebars | Router groups and explicit order | Keep the route tree as the source of truth instead of maintaining a second sidebar file. |
| Markdown components | Shortcodes and typed GoFastr screens | Use native Go components. React, MDX, and Markdoc runtimes are not required. |
| Search | Local JSON, Pagefind, and native command palette | Keep search deployable without a hosted service. |
| i18n | Metadata, filters, alternates, selectors, translated shell labels, and fallback for untranslated pages | The project supplies translated content; the framework does not invent translations, and never machine-translates. |
| Versioning | Metadata, filtering, selectors, versioned collections, and manifest inventory | Snapshot creation remains a content or VCS workflow; the Router publishes immutable directories once they are authored. |
| Plugins | Go plugins contribute routes, content components, validation, assets, and search data | Extensions stay inside the same Router lifecycle. |
| Blog and RSS | `MarkdownBlog`, `BlogPosts`, `RSSXML`, live feed mounting, and static feed output | Keep posts in the Router so publishing metadata, navigation, search, locale/version filters, and export stay in sync. |
| Analytics and third-party integrations | Host or plugin responsibility | Keep the white-label core free of vendor accounts and tracking assumptions. |

Snapshot creation is intentionally a content or VCS workflow rather than a
command that rewrites arbitrary Go route code. Blog publishing is opt-in by
calling `MarkdownBlog`; sites that do not publish updates do not carry a feed.

## What the generated project proves

The starter contains Markdown pages, nested groups, a typed playground,
framework UI examples, an OpenAPI reference, local search, PWA export, agent
references, and browser tests. The non-CLI fixture exercises the same public
Router API by hand. Keep both projects in the test suite when changing a
public contract.

Continue with [The router](/docs/concepts/router), [Content authoring](/docs/concepts/content), [Framework UI](/docs/build/framework-ui), or [Testing](/docs/operate/testing).
