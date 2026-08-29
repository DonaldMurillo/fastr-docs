# OpenAPI reference

The bundled OpenAPI plugin turns a contract into a first-class route. It contributes the reference index, operation filters, request console, server selection, and search records to the same router as the handbook.

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

The contract remains the source of truth for operations and schemas. The optional `ServerURL` override lets the same build point at a local mock, staging API, or production host without rewriting the document.

## Try it safely

The generated example contract exposes a mock project endpoint. Filter operations, open an operation, inspect its parameters, and send a request from the console. The E2E suite uses a local mock server so this behavior is deterministic and does not transmit data to a third-party API.

## Production checklist

- Keep the OpenAPI document versioned with the service.
- Set `API_SERVER_URL` per deployment environment.
- Declare external origins through the router so the CSP remains strict.
- Test both a successful request and a visible error response.
- Keep the API section in the same top-level navigation model as the handbook.
