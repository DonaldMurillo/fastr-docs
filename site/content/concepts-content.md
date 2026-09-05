---
tags: [content, markdown, front-matter]
locale: en
---

# Content authoring

Use Markdown for documentation that should be durable, searchable, exportable, and easy to review. fastr-docs removes front matter before rendering so page content, search text, and headings stay consistent.

## Put metadata in front matter

```md
---
title: Configure deployment
description: Set the public URL and static base path.
tags: [deploy, hosting]
order: 4
last_updated: 2026-08-29
---

# Configure deployment
```

Supported metadata includes titles, slugs, descriptions, drafts, no-index rules, edit and canonical URLs, authors, dates, locale, version, alternates, tags, redirects, and order.

## Register a file

```go
router.MustPage("/docs/deploy", docs.PageConfig{
    Title:      "Configure deployment",
    SourcePath: "content/deploy.md",
    Order:      4,
    Offline:    true,
})
```

Use `PageFile` when the source is selected at runtime, or `MarkdownCollection` when a directory should become a route family. Collection routes inherit front matter and can apply locale/version prefixes.

Set `slug` when a collection file needs a URL that does not match its
filename. The value may contain nested path segments; explicit `Router` paths
remain the source of truth for manually registered pages:

```md
---
title: The router
slug: concepts/router
---
```

That file publishes at `/docs/concepts/router` when the collection prefix is
`/docs`.

For packaged content, use `MarkdownCollectionFS` with any `fs.FS`. This keeps the same route and metadata behavior when content comes from `embed.FS`, a test filesystem, or another virtual source:

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
the files are compiled into the binary. Both APIs use the same front matter,
ordering, locale, version, draft, search, and validation rules.

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

## Add typed Markdown components

Register shared components through `MarkdownComponentsPlugin`:

```go
router.Use(docs.MarkdownComponentsPlugin{Components: map[string]docs.MarkdownComponent{
    "note": func(props map[string]string, body render.HTML) render.HTML {
        return ui.Callout(ui.CalloutConfig{
            Title: props["title"], Variant: ui.StatusInfo,
        }, body)
    },
}})
```

Then use the component in Markdown:

```md
{{< note title="Important" >}}
This body remains part of the Markdown page.
{{< /note >}}
```

Keep shared vocabulary in the Router and page-specific components in `PageConfig.Components`. See [Screens and components](/docs/build/screens) when the whole route needs typed state or actions.
