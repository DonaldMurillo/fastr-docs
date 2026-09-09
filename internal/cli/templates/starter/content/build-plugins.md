# Plugins and extensions

A plugin is a focused contribution to the central router. It can register routes, search metadata, connection origins, or other composition data without replacing the docs shell.

## Plugin shape

```go
type Plugin interface {
    Name() string
    Apply(*docs.Router) error
}
```

Apply the plugin during project setup. Fail loudly when its configuration is invalid; strict startup errors are safer than publishing a half-mounted extension.

The package includes two reusable plugin building blocks:

```go
router.Use(docs.MarkdownCollectionPlugin{
    Path: "/docs",
    Dir:  "content",
    Config: docs.CollectionConfig{Offline: true},
})

router.Use(docs.MarkdownComponentsPlugin{
    Components: map[string]docs.MarkdownComponent{
        "note": renderNote,
    },
})
```

For packaged content, use `FS` and `Root` instead of `Dir`:

```go
router.Use(docs.MarkdownCollectionPlugin{
    Path: "/docs",
    FS:   contentFS,
    Root: "content",
    Config: docs.CollectionConfig{Offline: true},
})
```

The plugin delegates to the same collection APIs, so metadata, ordering,
drafts, search, and validation stay consistent across disk and virtual files.

The OpenAPI plugin is another ordinary contribution to this same tree. An
extension may add typed screens, Markdown pages, global shortcodes, validation,
assets, or allowed connect origins, but it should not create a second route or
search registry.

## Keep extensions white-label

An extension should contribute its content and behavior, not impose a product identity. Use the host theme and GoFastr primitives for typography, controls, focus behavior, mobile layout, and color scheme.

## Extension checklist

- Register every visible route in the shared router.
- Give generated routes stable titles, descriptions, tags, and order values.
- Add only the external origins the browser must contact.
- Include extension assets in the PWA precache when they are needed offline.
- Add E2E coverage for install, navigation, happy path, error path, and mobile behavior.
- Document the extension in `agents/claude.md` and the authoring skill when agents need to modify it.
