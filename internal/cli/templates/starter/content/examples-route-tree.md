# Route tree

The generated site is easiest to understand as a tree:

```text
/
├── docs
│   ├── getting-started
│   ├── concepts/
│   │   ├── router
│   │   ├── content
│   │   └── layouts
│   ├── build/
│   ├── operate/
│   └── collaborate/
├── examples/
└── api-reference
```

The header chooses the top-level section. The contextual sidebar shows the active branch. The article gets breadcrumbs, previous/next controls, search metadata, and a table of contents from the same registration.

## Why this scales

Adding a route is one coherent change: register it, author its content, choose its order, and add its behavior test. The navigation does not need a manual copy of the route, and the search index does not need a second update.

## Explore the handbook

Use the top navigation to switch between Documentation, Examples, and API reference. On a narrow viewport, open the native drawer and select a different section; the shell should close and the destination should remain in the correct layout.
