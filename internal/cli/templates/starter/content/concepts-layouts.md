# Layouts and navigation

fastr-docs uses nested GoFastr layouts. The outer layout owns white-label chrome such as the brand, theme toggle, search, PWA shell, and global header. A section layout owns its local sidebar and screen group.

## Global and section concerns

```go
func (r *Router) Layout() *uiapp.Layout {
    return uiapp.NewLayout("docs").WithHeader(&docsHeader{router: r})
}

func (r *Router) sectionLayout(route *docs.Route) *uiapp.Layout {
    return uiapp.NewLayout("docs-section").WithSidebar(&docsSidebar{router: r})
}
```

The header switches between top-level sections. The sidebar stays focused on the active section, so a large documentation tree does not compete with an API explorer or examples area.

## Why nested layouts matter

Different route groups can have different local navigation, content density, or interaction patterns while still sharing the same brand and command palette. Layouts inside layouts let the product grow without cloning the shell for every feature.

## Responsive behavior

- Wide documents use a readable article column plus a sticky in-page rail.
- Tablet layouts collapse the rail into a dropdown selector.
- Mobile layouts move the section tree into the native GoFastr drawer.
- Header navigation and the local sidebar remain separate responsibilities.

The browser suite checks these behaviors at desktop, tablet, and mobile widths. When changing layout CSS, inspect all three states visually as well as asserting their DOM behavior.
