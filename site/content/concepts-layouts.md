---
tags: [layouts, navigation, responsive]
---

# Layouts and navigation

fastr-docs uses nested GoFastr layouts. The outer layout owns white-label chrome; each top-level route section can own its contextual sidebar and screen group.

## Global and section layout

```go
router.Layout()
```

The global layout renders the brand, theme control, native command palette, PWA shell, and header links derived from top-level routes. `Router.Mount` creates a section layout for each active route branch.

The persistent sidebar narrows to the active section on documentation pages. On small screens, GoFastr’s drawer exposes the same route tree with its active branch expanded and current page selected.

## Header groups are not duplicate sidebar items

The header links represent top-level sections. The sidebar represents the active section’s local tree. A header link can remain active on nested pages because matching uses the section route prefix.

## In-page navigation

The document layout derives headings from Markdown and renders an anchored rail on wide screens. At narrower widths it uses a dropdown selector. Scrollspy updates the selected heading while preserving normal document scrolling.

Use the [route tree example](/examples/route-tree) to see how a single registration produces all three navigation surfaces.
