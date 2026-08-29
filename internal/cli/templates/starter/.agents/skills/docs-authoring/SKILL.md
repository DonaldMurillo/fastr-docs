---
name: docs-authoring
description: Author and extend this documentation site through docs.Router.
---

# Docs authoring

Use `docs/router.go` for route registration. Prefer Markdown for durable static
content and typed GoFastr screens for interactive experiences. Every new route
needs a title, description, explicit sibling order, and a navigation destination.

Markdown front matter is the metadata API. Use `draft`, `noindex`, `edit_url`,
`canonical`, `authors`, `date`, `last_updated`, `locale`, `version`,
`alternates`, `tags`, and `redirects` rather than inventing page-local config.
For directory-based content, use `router.MarkdownCollection`; do not create a
second navigation manifest. Verify the intended locale/version and that drafts
are absent from `SearchIndex`, navigation, mounting, and export.

Run `fastr-docs doctor .`, `go test ./...`, and the repository E2E suite after
changes. Inspect a static export for `sitemap.xml`, `robots.txt`, assets,
manifest, service worker, and representative route pages.
