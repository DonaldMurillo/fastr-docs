---
tags: [offline, pwa, static-export]
---

# Offline and PWA delivery

Static documentation should remain readable when the network disappears. Mark durable routes `Offline: true`, then export the site so GoFastr can precache the shell and fastr-docs can include the route assets.

## Export the site

```sh
fastr-docs export . --out dist --pagefind
```

The output includes route HTML, the manifest, service worker, generated install icons, the docs runtime, search records, OpenAPI runtime, agent assets, and project public assets.

## Install and update the docs

The live host is installable on `localhost` and on HTTPS deployments because the generated layout enables GoFastr's PWA support and derives the required install icons. A live install keeps the application shell and offline screen available, while rendered document navigations stay network-first so published content is never stale by design.

For a fully offline documentation app, install the static export instead:

```sh
fastr-docs export . --out dist --pagefind
```

Deploy `dist` to a static host or CDN. The exported worker precaches every exported route, and its content fingerprint changes whenever the export changes. On the next visit, the browser installs the new worker and cache in the background. GoFastr waits until old tabs close before activating it, so an open documentation session is not interrupted. A new tab then uses the updated site.

Keep the framework dependency and its CLI aligned when updating the project:

```sh
fastr-docs upgrade .
fastr-docs upgrade . --apply
```

The first command shows the release migrations between the version in `go.mod` and the target release. The second applies the dependency, tidy, build, and test steps. The standalone `gofastr` binary and the Go module are versioned separately, so install the target CLI before applying a multi-release upgrade.

## Understand the boundary

Markdown pages, navigation, local search, and exported UI can work offline. Interactive API requests still need a network and a reachable server. The OpenAPI console reports that limitation when no server URL is configured.

## Keep the cache intentional

The generated host precaches only its known runtime and route assets. If a plugin owns browser assets, mount them through the Router so the host and export pipeline can account for them.

Test the exported site in a real browser with the network disabled. The E2E suite for this project covers navigation, Pagefind search, PWA assets, and offline route rendering.
