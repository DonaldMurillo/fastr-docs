---
tags: [getting-started, authoring]
date: 2026-08-01
last_updated: 2026-08-03
---

# Getting started

The generated project is deliberately opinionated: one router, explicit order, strict validation, and a native GoFastr shell. You can keep the defaults or replace every brand and page while preserving the model.

## Run the project

```sh
fastr-docs check .
fastr-docs dev .
```

GoFastr's dev loop uses `http://localhost:8080` by default. Pass
`--addr localhost:3070` to choose another address. Use `Meta/Control + K` to
search the route tree, switch the color scheme from the header, and resize the
page to exercise the mobile drawer and in-page navigation. The loop rebuilds
and refreshes the browser when Go, Markdown, HTML, CSS, JavaScript, or
JSON/YAML contract files change.

## Add a page

Register a page with `router.Page` or `router.MustPage`, give it a clear title and description, and assign an explicit `Order` so sibling navigation stays intentional.

```go
router.MustPage("/docs/decisions", docs.PageConfig{
    Title:       "Architecture decisions",
    Description: "Why the project is shaped this way.",
    SourcePath:  "content/decisions.md",
    Order:       6,
    Offline:     true,
})
```

The path, title, description, order, search metadata, breadcrumbs, previous/next links, and offline eligibility all come from this registration.

## Add a navigation badge

Use a short badge when a route needs a visible qualifier such as `New`, `Popular`, or `Beta`. The badge stays attached to the route in both the desktop sidebar and mobile drawer; it does not change route behavior.

```go
router.MustPage("/docs/decisions", docs.PageConfig{
    Title:       "Architecture decisions",
    Description: "Why the project is shaped this way.",
    SourcePath:  "content/decisions.md",
    Order:       6,
    Badge:       docs.NavBadge{Label: "New", Tone: docs.NavBadgeToneAccent},
})
```

Groups and typed screens accept the same `Badge` field. Keep labels short so the navigation remains scannable at narrow widths.

## Add a screen

Use `router.Screen` when a page needs typed state, actions, or a custom GoFastr component. It still appears in the same navigation and search surfaces.

```go
router.MustScreen("/examples/playground", docs.ScreenConfig{
    Title:       "Playground",
    Description: "A typed interactive surface.",
    Component:   &PlaygroundScreen{},
    Order:       1,
})
```

Prefer Markdown when the content is durable prose. Use a screen when the user needs state, actions, data loading, or a framework component.

## Install and update it

The generated host is installable as a PWA on `localhost` and on HTTPS. For a
fully offline install, deploy the static export:

```sh
fastr-docs export . --out dist --pagefind
```

GoFastr fingerprints the static worker cache. After a redeploy, the browser
installs the new worker in the background and activates it after older tabs
close, so the current session is not interrupted.

`fastr-docs dev` delegates to GoFastr and adds JSON/YAML contract watching. If
the GoFastr CLI is installed, `gofastr dev` runs the framework watcher directly
without that extra contract bridge. Use `fastr-docs upgrade .` to review
framework migrations before updating the dependency in `go.mod`.


## Extend the project

Plugins implement `Name` and `Apply`. The bundled OpenAPI plugin registers `/api-reference` into the same router rather than creating a parallel documentation system. It reads the contract's `servers` URL and supports an `API_SERVER_URL` override for deployment-specific API hosts.

## Before you hand it off

Run `go test ./...`, `fastr-docs check .`, and the browser suite in `e2e/`. Keep strict validation enabled unless a documented integration requirement makes opting out necessary.
