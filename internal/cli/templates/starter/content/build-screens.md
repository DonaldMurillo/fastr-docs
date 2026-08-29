# Screens and components

Pages are ideal for durable prose. Screens are the escape hatch for product surfaces that need typed state, actions, data loading, or custom GoFastr components.

## Register a typed surface

```go
router.MustScreen("/examples/playground", docs.ScreenConfig{
    Title:       "Playground",
    Description: "Try the route model in a live surface.",
    Component:   &PlaygroundScreen{},
    SearchText:  "interactive playground screen state actions",
    Order:       1,
})
```

The screen uses the same breadcrumbs, sidebar, search index, layout selection, and previous/next navigation as a Markdown page. The rendering strategy changes; the product contract does not.

## Keep state close to the interaction

Use GoFastr components for controls, feedback, disclosure, forms, and data views. Keep server actions and validation in the screen or host layer, not in ad-hoc page JavaScript. This lets the framework own accessibility, navigation, policy, and lifecycle behavior.

## A screen checklist

1. Give it a title, description, explicit order, and search terms.
2. Render a useful empty state before data arrives.
3. Make every action observable by keyboard and screen readers.
4. Test the success, error, and navigation paths in a real browser.
5. Verify the surface at the same breakpoints as the surrounding docs.

Use [custom surfaces](/examples/custom-surface) as the design checklist before introducing a new interactive route.
