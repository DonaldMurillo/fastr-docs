# Deploy and customize

The generated project is white-label by construction. Keep the router and content model, then make the site yours through the site name, theme variables, logo, custom CSS, API host, and deployment base path.

## Run locally

```sh
fastr-docs check .
fastr-docs dev .
```

## Build a static release

```sh
fastr-docs export . --out dist --base /docs
```

Use `--base` when the site is served beneath a path. Verify direct navigation to nested routes, refresh behavior, asset URLs, the service worker scope, and search before publishing.

## Customize safely

- Replace the site name and icon in the generated project.
- Use the theme variables and GoFastr UI configuration before adding bespoke components.
- Keep CSP origins explicit through plugin or router configuration.
- Set `API_SERVER_URL` in the deployment environment.
- Preserve route titles, descriptions, order, and offline metadata when rebranding.

The fastest route to a coherent product is to change identity and content while preserving the proven interaction patterns.
