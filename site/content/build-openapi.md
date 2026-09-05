---
tags: [openapi, plugin, api-reference]
---

# OpenAPI reference

The first-party OpenAPI plugin turns a contract into a route in the same Router. It renders an operation index, endpoint cards, parameters, schemas, and an optional request console. The API reference is a normal route: it participates in the top navigation, contextual sidebar, local search, static export, and PWA shell.

## Mount the plugin

```go
router.Use(openapi.Plugin{
    SpecPath:  "openapi.json",
    Path:      "/api-reference",
    Title:     "API reference",
    ServerURL: os.Getenv("API_SERVER_URL"),
    Order:     3,
})
```

The plugin accepts JSON and YAML contracts. The first `servers` URL is used by default. `ServerURL` overrides it for staging or production without changing the checked-in contract or route tree.

## Use the request console

The console follows the selected operation and renders the values that the contract describes:

- path, query, header, and cookie parameters
- required markers and client-side required-value checks
- request bodies with content type and JSON validation
- readable response output and network/CORS errors

Path parameters are substituted into the URL, query values are encoded, and request bodies are sent only when the selected operation declares one. Cookie parameters are shown for reference but cannot be injected into a cross-origin browser request.

## Keep requests explicit

The request console runs in the browser and requires the API server to allow CORS. The Router records the resolved origin so the GoFastr host can include it in its strict `connect-src` policy. If no server URL is configured, the reference remains readable and the console explains why requests are unavailable.

This site includes a small example contract and local mock server so the plugin can be exercised end to end. They are fixtures, not an HTTP API offered by fastr-docs itself.

## Extend the plugin

The plugin contributes a normal screen route. Projects can set its title, description, order, server URL, and sidebar badge just like any other route. Add a different API adapter by implementing `docs.Plugin` and registering it with the same Router.
