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

The export rewrites runtime, search, and OpenAPI asset URLs for the chosen base path. Serve the output as a static site or place it behind the GoFastr host.
