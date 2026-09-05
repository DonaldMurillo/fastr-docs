# Content authoring

Markdown is the default authoring format because it is portable, reviewable, searchable, and easy for both humans and agents to edit. A page registration supplies the product metadata; the Markdown file supplies the reader experience.

## A durable page

```go
router.MustPage("/docs/decisions", docs.PageConfig{
    Title:       "Architecture decisions",
    Description: "The choices behind the product.",
    SourcePath:  "content/decisions.md",
    SearchText:  "architecture decisions tradeoffs",
    Tags:        []string{"architecture", "reference"},
    Order:       6,
    Offline:     true,
})
```

Use `Description` for navigation and search summaries. Use `SearchText` when the page contains important terms that are not written verbatim in the visible article. Tags make future external indexers or filtered surfaces possible without changing the route model.

Collection files can set `slug` when the URL should not match the filename:

```md
---
title: The router
slug: concepts/router
---
```

That file publishes at `/docs/concepts/router` when the collection prefix is
`/docs`. Explicit `Router` paths remain the source of truth for manually
registered pages.

## Load packaged content

Use `MarkdownCollectionFS` when content comes from `embed.FS`, a test
filesystem, or another virtual source:

```go
import "embed"

//go:embed content
var contentFS embed.FS

if err := router.MarkdownCollectionFS(
    "/docs", contentFS, "content",
    docs.CollectionConfig{Offline: true},
); err != nil {
    panic(err)
}
```

Use `MarkdownCollection` for disk-backed content during development. Run
`fastr-docs dev` to rebuild the Router when those files change. Use
`MarkdownCollectionFS` for content that should travel with the binary or
another virtual filesystem. Embedded content still requires a rebuild because
the files are compiled into the binary. Both collection APIs apply the same
metadata, ordering, locale, version, draft, search, and validation rules.

## Organize released versions

Keep immutable snapshots in one child directory per version and register them
with `MarkdownVersionedCollection`:

```text
content/versions/
├── v1/guide.md
└── v2/guide.md
```

```go
if err := router.MarkdownVersionedCollection("/docs", "content/versions", docs.VersionedCollectionConfig{
    Current:  "v2",
    Versions: []string{"v2", "v1"},
    Collection: docs.CollectionConfig{Offline: true},
}); err != nil {
    panic(err)
}
```

The current version keeps the normal `/docs/guide` URL. Older versions use
`/docs/v1/guide`. The same helper is available as
`MarkdownVersionedCollectionFS` for embedded content. The version selector,
filters, search index, static export, and manifest all read these routes from
the same Router.

## Markdown conventions

- Start with one page-level heading.
- Use headings to create useful in-page navigation.
- Keep code examples complete enough to copy.
- Link to route paths rather than duplicating product URLs.
- Keep paragraphs short and lead with the decision or outcome.
- Prefer an explicit warning or note over hidden assumptions.

The shell derives a sticky table of contents from the headings. On narrower layouts it becomes the **On this page** selector so the document remains readable without a second column.

## When to use a screen

Markdown should explain a feature. A [screen](/docs/build/screens) should let the reader operate it, inspect state, or trigger an action. Keeping those roles separate makes the content easy to export while preserving room for rich examples.

## Reusable content components

When prose needs a consistent visual treatment, register a
`MarkdownComponentsPlugin` and use a shortcode. The component receives parsed
string props plus a server-rendered Markdown body:

```md
{{< note title="Decision" >}}
The body can contain **Markdown**, links, and nested registered components.
{{< /note >}}
```

Unknown names and unclosed blocks fail strict validation before the site is
served. This keeps the authoring format portable while still allowing typed
GoFastr surfaces where they add real value.

{{< callout title="A typed authoring escape hatch" >}}
The generated project registers this shared component vocabulary while keeping the page body Markdown.
{{< /callout >}}
