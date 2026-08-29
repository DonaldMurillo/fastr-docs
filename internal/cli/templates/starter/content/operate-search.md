# Search

Search is derived from the router, not maintained as a second list. Every visible page, screen, and plugin route can contribute a title, description, body text, and tags to the local search index.

## Use the native command palette

Press `Meta/Control + K` or activate **Search** in the header. Results are keyboard-friendly route commands, and selecting one navigates through the GoFastr router while preserving the shell where possible.

## Improve result quality

```go
router.MustPage("/docs/decisions", docs.PageConfig{
    Title:      "Architecture decisions",
    SearchText: "routing layouts static export offline PWA",
    Tags:       []string{"architecture", "deployment"},
    // ...
})
```

Use `SearchText` for concepts that readers will search for but that are intentionally absent from the prose. Keep route titles precise; a clever title is harder to discover than a literal one.

## Portable JSON and Pagefind

The router emits portable JSON records and the default runtime keeps search
available without an extra build dependency. For a larger static site, install
the Pagefind CLI and run:

```sh
fastr-docs export . --out dist --pagefind
```

That command exports the HTML first, then runs Pagefind against the generated
site. The native `Meta/Control + K` palette loads the chunked Pagefind browser
API when its bundle is present and falls back to the JSON route list if a dev
server or deployment does not include it.
