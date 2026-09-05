---
tags: [overview, fastr-docs]
---

# Documentation

fastr-docs is a reusable documentation framework for GoFastr applications. A central `docs.Router` registers pages, typed screens, groups, and plugins. The same tree drives navigation, search, breadcrumbs, previous/next links, static export, and validation.

## Choose a path

- [Getting started](/docs/getting-started) — install the CLI and create a working site.
- [The Router](/docs/concepts/router) — understand the object that owns the documentation tree.
- [Content authoring](/docs/concepts/content) — use Markdown, front matter, collections, and components.
- [Screens and components](/docs/build/screens) — add typed GoFastr surfaces when Markdown is not enough.
- [Framework UI](/docs/build/framework-ui) — use the native components available to pages and screens.
- [Example API reference](/api-reference) — see the OpenAPI plugin mounted into the same site.

## What belongs to the Router

Routes carry titles, descriptions, explicit sibling order, visibility, offline eligibility, search text, content metadata, and optional sidebar badges. Registering a route once gives every adapter the same source of truth.

## Next steps

1. Read [Getting started](/docs/getting-started).
2. Try the [typed playground](/examples/playground).
3. Review the [route tree example](/examples/route-tree).
4. Press `Ctrl+K` or `⌘K` to search this site.
