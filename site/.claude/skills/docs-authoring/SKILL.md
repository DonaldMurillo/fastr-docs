---
name: docs-authoring
description: Add or change pages, navigation, and content metadata in this fastr-docs site.
---

# Docs authoring

`docs/router.go` owns the route tree. It is the single source of truth for
navigation, breadcrumbs, previous/next links, the search index, and the static
export. There is no second sidebar file to keep in sync, so do not add one.

## Choosing a route type

Markdown for durable prose. `router.MustPage` with a `Source`, or
`router.MarkdownCollection(prefix, dir, cfg)` to pick up a whole directory.
Typed GoFastr screens for anything interactive: `router.MustScreen` with a
`Component`. Both land on the same tree and both are searchable.

Every route needs a title, a description, and an explicit sibling `Order`.
Order is not inferred from filename or registration order, and duplicate order
values inside one group fail validation.

## Front matter is the metadata API

These keys are read from Markdown front matter. Use them instead of inventing
page-local configuration:

`title`, `slug`, `description`, `excerpt`, `draft`, `noindex`, `edit_url`,
`canonical`, `image`, `authors`, `date`, `last_updated`, `locale`, `version`,
`alternates`, `tags`, `redirects`, `order`.

`draft: true` removes a page from navigation, search, mounting, export, and
RSS. After changing draft status, confirm the page is absent from all five,
not just the sidebar.

## Embedding components in Markdown

Register a component adapter once, then use the shortcode:

```md
{{< note title="Important" >}}
The body is still Markdown and can contain links and emphasis.
{{< /note >}}
```

## Before you call it done

Run `fastr-docs check .` and `go test ./...`. `fastr-docs doctor .` reports
route and content problems that validation alone does not surface.

See the `docs-publishing` skill for export and deploy, and `docs-blog` for
posts and feeds.
