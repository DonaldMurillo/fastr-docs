---
tags: [search, pagefind, command-palette]
---

# Search

Search is derived from the Router, so a route is searchable as soon as it is registered and published. Each record includes its path, title, description, kind, tags, headings, and searchable text.

## Use the native command palette

`Router.MountCommandPalette` mounts GoFastr’s command palette. The header trigger and `Ctrl+K` / `⌘K` shortcut open the same surface. Results navigate with normal route URLs and preserve the app shell.

```go
if err := router.MountCommandPalette(server.Router()); err != nil {
    return err
}
```

## Choose a backend

The default JSON backend is dependency-free. The command palette fetches the generated `search.json` index on its first query, so body text is searchable in development and static exports. Static builds can opt into Pagefind for a chunked index:

```go
router := docs.NewRouter(docs.WithPagefind())
```

```sh
fastr-docs export . --out dist --pagefind
```

Use `WithPagefindPath` when a reverse proxy serves the Pagefind bundle below a custom asset prefix. The command palette keeps the same user-facing behavior.

## Search rules

Hidden, draft, locale-inactive, version-inactive, and no-index content is excluded from public search. A custom `SearchProvider` can replace the JSON artifact while keeping the portable `SearchIndex` model.
