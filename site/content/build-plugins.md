---
tags: [plugins, extensions, router]
---

# Plugins and extensions

Plugins extend a project without creating a second content system. A plugin can register pages or screens, add search text, declare external origins, validate project rules, and mount browser assets.

## Implement the basic contract

```go
type Plugin interface {
    Name() string
    Apply(*Router) error
}
```

`Apply` should register everything the extension owns. Route titles, descriptions, order, visibility, search metadata, and badges then flow into the normal adapters.

The built-in collection plugin accepts either a disk directory or a virtual
filesystem. Choose one source per plugin:

```go
router.Use(docs.MarkdownCollectionPlugin{
    Path: "/docs",
    FS:   contentFS,
    Root: "content",
    Config: docs.CollectionConfig{Offline: true},
})
```

Use `Dir` for development-time files and `FS` plus `Root` for `embed.FS` or
another packaged source. The plugin delegates to the same collection APIs,
so metadata, ordering, drafts, search, and validation stay consistent.

## Add validation or assets

Use `ValidatingPlugin` when the extension has checks that must run with the Router’s strict validation. Use `AssetPlugin` when it needs to mount static files through the GoFastr HTTP router.

```go
type ValidatingPlugin interface {
    Plugin
    Validate(*Router) error
}
```

Call `router.MountPluginAssets(server.Router())` after the host router exists. Keep external origins explicit with `router.AllowConnectOrigin` so the generated CSP stays narrow.

## The bundled example

The [OpenAPI reference](/docs/build/openapi) is implemented as an in-project plugin. Its screen is registered alongside Markdown routes and is covered by the same search, navigation, export, and browser test surfaces.
