# Offline and PWA

Static documentation should remain useful when a reader is on a train, behind a restrictive network, or browsing a deployed preview without a live API. Mark durable routes as offline-eligible and export the site as an installable PWA.

## Mark content for offline delivery

```go
router.MustPage("/docs/decisions", docs.PageConfig{
    SourcePath: "content/decisions.md",
    Offline:    true,
    // ...
})
```

The generated host precaches the docs runtime, OpenAPI runtime, search index, icons, and exported static pages. Interactive API requests still need a network; the handbook itself does not.

## Export the site

```sh
fastr-docs export . --out dist
```

The export keeps route paths, breadcrumbs, in-page navigation, the manifest, service worker, icons, runtime assets, and the local search records. Deploy `dist` to a static host or CDN.

## Test the offline contract

The browser suite verifies that exported docs routes remain navigable and that the PWA assets are present. Add a test whenever a new runtime asset or route type becomes part of the offline experience.
