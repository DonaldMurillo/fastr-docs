# The router

`docs.Router` is the central object of a fastr-docs project. It is intentionally more than a URL registry: it is the interface from which the shell, navigation, search index, breadcrumbs, previous/next links, plugins, and static export are derived.

## Register once

```go
router := docs.NewRouter(docs.WithSiteName("Acme Docs"))
router.MustPage("/", docs.PageConfig{Title: "Acme Docs", Body: HomeBody, Order: 1})
router.MustPage("/docs", docs.PageConfig{Title: "Documentation", SourcePath: "content/index.md", Order: 2})
```

The same `router` is mounted into the GoFastr app and used to mount framework navigation and the command palette. A plugin contributes to that object instead of creating a parallel registry.

## Order is a product decision

Every sibling route receives an explicit `Order`. Registration order is only a stable tie-breaker; it is never a substitute for deciding how a reader should move through the product.

```go
guides := router.MustGroup("/docs/guides", docs.GroupConfig{
    Title: "Guides",
    Order: 4,
})
guides.MustPage("deploy", docs.PageConfig{
    Title: "Deploy",
    Order: 1,
    SourcePath: "content/deploy.md",
})
```

## Route kinds

| Kind | Best for | Example |
| --- | --- | --- |
| Page | Durable prose and reference | Markdown content |
| Screen | State, actions, and custom UI | Playground or dashboard |
| Group | Navigation hierarchy | Concepts or Guides |
| Plugin | External spec-driven surfaces | OpenAPI reference |

Read [content authoring](/docs/concepts/content) for the page model and [screens and components](/docs/build/screens) for interactive surfaces.

## Validate early

Strict validation is on by default. Missing titles, missing content, unusable screen components, unsafe paths, duplicate routes, and duplicate explicit orders fail during startup or export instead of becoming silent navigation defects.
