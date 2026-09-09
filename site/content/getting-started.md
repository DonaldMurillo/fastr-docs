---
tags: [getting-started, installation, authoring]
locale: en
---

# Getting started

Create a documentation site with the CLI, then keep the route tree in `docs/router.go` and the prose in `content/`.

## Create a project

Install the development CLI from this repository or from the published module:

```sh
go install github.com/DonaldMurillo/fastr-docs/cmd/fastr-docs@latest
fastr-docs init acme-docs --name "Acme Docs" --module example.com/acme-docs
cd acme-docs
```

The generator creates a runnable project with a Router, Markdown content, a typed example screen, an OpenAPI spec, PWA assets, and agent references.

## Run the checks

```sh
fastr-docs doctor .
fastr-docs dev .
```

`doctor` validates the project, tidies dependencies, and runs its Go tests.
`dev` watches Go, Markdown, HTML, CSS, JavaScript, and JSON/YAML spec
files. It rebuilds the server and refreshes open browser tabs after a
successful change. Open the local server and use `Ctrl+K` or `⌘K` to search
the route tree.

## Register a page

Pages are the default rendering strategy. Point one at a Markdown file and give sibling routes an explicit order:

```go
router.MustPage("/docs/decisions", docs.PageConfig{
    Title:       "Architecture decisions",
    Description: "Why the project is shaped this way.",
    SourcePath:  "content/decisions.md",
    Order:       6,
    Offline:     true,
})
```

The registration supplies the URL, title, description, navigation item, search record, breadcrumbs, previous/next links, and export metadata.

## Add a navigation badge

Use a short badge when a route needs a visible qualifier such as `New`, `Popular`, or `Beta`:

```go
Badge: docs.NavBadge{Label: "New", Tone: docs.NavBadgeToneAccent},
```

Groups, pages, screens, and OpenAPI plugin routes accept the same metadata. It appears in the desktop rail and mobile drawer without changing route behavior.

## Export a static site

```sh
fastr-docs export . --out dist --pagefind
```

The export includes route HTML, the PWA shell, local search assets, agent discovery files, and the OpenAPI runtime. Use `--base /docs` when deploying below a path prefix.

## Install and update it

The generated host is installable as a PWA on `localhost` and on HTTPS. For a
fully offline install, deploy the static `dist` directory. GoFastr fingerprints
that export and installs a new worker after a redeploy; existing tabs finish on
their current version and new tabs use the update.

`fastr-docs dev` delegates development to GoFastr and adds JSON/YAML spec
watching. If you install the GoFastr CLI separately, the direct framework
command is:

```sh
gofastr dev
```

Use `fastr-docs upgrade .` to review framework migrations before updating the
GoFastr dependency in `go.mod`.

## Next

Read [The Router](/docs/concepts/router) to understand the source of truth, then compare the [page and screen example](/examples/playground).
