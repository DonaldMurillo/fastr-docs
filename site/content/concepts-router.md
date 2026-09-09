---
tags: [router, concepts, architecture]
locale: en
---

# The Router

`docs.Router` is the central object in a fastr-docs project. Register content once and let the host derive the navigation tree, search records, layouts, export manifest, and validation results.

## Register the tree

```go
router := docs.NewRouter(docs.WithSiteName("Acme Docs"))

router.MustPage("/", docs.PageConfig{
    Title: "Acme Docs", Source: "# Acme Docs", Order: 1,
})

guides := router.MustGroup("/guides", docs.GroupConfig{
    Title: "Guides", Description: "Learn the product.", Order: 2,
})
guides.MustPage("getting-started", docs.PageConfig{
    Title: "Getting started", SourcePath: "content/getting-started.md", Order: 1,
})
```

Groups are metadata-only nodes. They provide hierarchy and disclosure behavior; their children are the navigable pages or screens.

## Pick a route kind

- `Page` renders Markdown or a page body.
- `Screen` renders a typed GoFastr component.
- `Group` adds a named branch to the tree.
- `Plugin` contributes routes through the same registration interface.

Every route can also carry search text, tags, locale/version metadata, offline eligibility, a canonical URL, redirects, and a short sidebar badge.

## Keep order intentional

`Order` controls sibling order. Registration order is only the stable tie-breaker when two routes share a value. Strict validation reports duplicate explicit orders so a growing tree does not silently reshuffle.

## Mount the Router

```go
site := uiapp.NewApp("Acme Docs").WithTheme(docs.DefaultTheme())
if err := router.Mount(site, router.Layout()); err != nil {
    return err
}
```

The layout composes GoFastr’s header, sidebar, document layout, in-page rail, command palette, theme toggle, and PWA host. The Router remains the source of truth; the framework owns rendering and interaction.

Continue with [Content authoring](/docs/concepts/content) or inspect the [route tree example](/examples/route-tree).
