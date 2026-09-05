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

## Install and update the docs

The generated host is installable on `localhost` and on HTTPS deployments. GoFastr supplies the manifest, install icons, service worker, offline screen, and registration script. A live install keeps the application shell available offline, while rendered document navigations remain network-first so they do not become stale.

For the full offline docs experience, install the static export after deploying it:

```sh
fastr-docs export . --out dist --pagefind
```

Every export gets a content-fingerprinted worker cache. When a changed export is deployed, the browser detects and installs the new worker in the background. GoFastr activates it after older tabs close, so the current session is not forced onto a different version. Open a new tab to use the updated docs.

When updating the project’s GoFastr dependency, use the framework CLI’s guided migration flow:

```sh
fastr-docs upgrade .
fastr-docs upgrade . --apply
```

The first command previews release notes and affected project lines. The second applies the dependency update and runs tidy, build, and test. Keep the installed `gofastr` CLI on the same release as the module in `go.mod`.

## Test the offline contract

The browser suite verifies that exported docs routes remain navigable and that the PWA assets are present. Add a test whenever a new runtime asset or route type becomes part of the offline experience.
