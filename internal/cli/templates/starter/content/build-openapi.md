# OpenAPI reference

The bundled OpenAPI plugin turns a spec into a first-class route. It contributes the reference index, operation filters, request console, and search records to the same router as the handbook.

## Add the plugin

```go
if err := router.Use(openapi.Plugin{
    SpecPath:  "openapi.json",
    Path:      "/api-reference",
    Title:     "API reference",
    ServerURL: os.Getenv("API_SERVER_URL"),
    Order:     3,
}); err != nil {
    panic(err)
}
```

The spec remains the source of truth for operations and schemas. JSON and YAML specs are supported. The optional `ServerURL` override lets the same build point at a local mock, staging API, or production host without rewriting the document.

## Try it safely

The generated example spec exposes a mock project endpoint. Filter operations, open an operation, inspect its path, query, header, or cookie parameters, and send a request from the console. Request bodies are validated as JSON before they are sent. The E2E suite uses a local mock server so this behavior is deterministic and does not transmit data to a third-party API.

Cookie parameters are documented in the console but cannot be injected into a cross-origin browser request. The server must also allow the docs origin through CORS.

## Production checklist

- Keep the OpenAPI document versioned with the service.
- Set `API_SERVER_URL` per deployment environment.
- Declare external origins through the router so the CSP remains strict.
- Test both a successful request and a visible error response.
- Keep the API section in the same top-level navigation model as the handbook.
