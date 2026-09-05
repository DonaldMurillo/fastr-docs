---
tags: [screens, components, gofastr]
---

# Screens and components

Use a typed screen when a route needs state, actions, data loading, or a richer GoFastr component composition. The route still participates in the same navigation, search, breadcrumbs, ordering, and export model as a Markdown page.

## Register a screen

```go
router.MustScreen("/examples/playground", docs.ScreenConfig{
    Title:       "Playground",
    Description: "Try the framework components.",
    Component:   &PlaygroundScreen{},
    Order:       1,
    Offline:     true,
})
```

The `Component` implements the normal GoFastr rendering contract. Build the surface with framework UI primitives and keep route-specific CSS next to the screen.

## Choose the boundary

Start with a Page when the user is reading or searching durable prose. Promote the route to a Screen when the user needs a live control, request console, chart, form, or server-backed state. The URL and its place in the route tree do not need to change.

## Test the surface

The [typed playground](/examples/playground) is a real screen inside this site. Its tabs, disclosure, counter, switch, and segmented control are covered by browser tests. Use the same approach for product-specific interactions: assert the user-visible result, not only the rendered markup.
