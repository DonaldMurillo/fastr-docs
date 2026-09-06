---
tags: [deploy, branding, hosting]
---

# Deploy and customize

The generated site is white-label. Keep the Router and host model, then replace the brand, public assets, runtime origin, and deployment path for the project.

## Set the brand

```go
router := docs.NewRouter(
    docs.WithSiteName("Acme Docs"),
    docs.WithBrand(docs.BrandConfig{
        Name: "Acme Docs", FaviconURL: "/assets/favicon.svg",
    }),
)
```

Use `WithCustomCSS` for product-specific visual changes and `MountAssets` for safe public files. The default theme exposes light and dark tokens for the shell and GoFastr components.

## Configure deployment origins

Set `PUBLIC_SITE_URL` for absolute sitemap and robots URLs. Set `API_SERVER_URL` when the OpenAPI server differs between local, staging, and production. Declared API origins are included in the host’s CSP; arbitrary network destinations are not.

## Deploy below a path

```sh
fastr-docs export . --out dist --base /docs
```

The export rewrites runtime, search, and OpenAPI asset URLs for the chosen base
path, and stamps the base into every page so the language selector, the search
results, and the sidebar's active state work below it too. Serve the output as
a static site or place it behind the GoFastr host.

## GitHub Pages

A repository named `<user>.github.io` is served at the root of that domain.
Every other repository is a project page under `https://<user>.github.io/<repo>/`,
which is the base-path case above. The workflow in `.github/workflows/pages.yml`
handles both: it derives the base from the repository name, exports, runs
Pagefind over the output, and publishes with `actions/deploy-pages`.

```yaml title=".github/workflows/pages.yml" {14,15}
- name: Export
  run: |
    case "$GITHUB_REPOSITORY" in
      */*.github.io) base="" ;;
      *) base="/${GITHUB_REPOSITORY#*/}" ;;
    esac
    fastr-docs export . --out dist --base "$base" --pagefind
```

Enable Pages for the repository with GitHub Actions as the source, and push.
The export writes `404.html` at the root, which is the file Pages serves for a
missing URL, and carries its content security policy as a meta tag because
Pages sets no headers. A custom domain makes the base empty again; nothing else
changes.
