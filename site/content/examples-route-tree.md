---
tags: [examples, route-tree, navigation]
---

# Route tree example

This site uses one Router for its handbook, examples, and OpenAPI reference. The top-level sections become header groups; each section owns a contextual sidebar branch.

```text
/
├── docs
│   ├── getting-started
│   ├── concepts
│   │   ├── router
│   │   ├── content
│   │   └── layouts
│   ├── build
│   │   ├── screens
│   │   ├── framework-ui
│   │   ├── openapi
│   │   └── plugins
│   └── operate
├── examples
│   ├── playground
│   ├── route-tree
│   └── custom-surface
└── api-reference
```

The Router preserves explicit sibling order. The active section narrows the desktop rail while the mobile drawer exposes the same branch, expands the active group, and marks the current page.

Open [Getting started](/docs/getting-started), then use the header to move between Documentation, Examples, and the API reference.
