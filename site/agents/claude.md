# Working on the fastr-docs documentation site

This is fastr-docs’ self-hosted documentation project. The package and CLI
live in the parent directory; this site consumes the checked-out package
through the local replace in `go.mod`.

## Source of truth

- `docs/router.go` owns the site route tree and explicit order.
- `content/` owns the Markdown pages.
- `docs/home.go` owns the landing surface.
- `docs/examples.go` owns the typed playground.
- `openapi.json` is a small OpenAPI plugin fixture, not a product API.

## Agent connection

The live host serves GoFastr's `/mcp` endpoint with read-only introspection
tools, and the agent card advertises that endpoint. Static exports do not
include the live transport.

## Authoring rules

1. Keep route registration and content changes together.
2. Give every route a clear title, description, and explicit sibling order.
3. Use `Badge` only for short navigation metadata such as `New`, `Popular`, or `Live`.
4. Prefer Markdown for durable explanation and typed screens for state or actions.
5. Use `MarkdownCollection` for disk-backed content and `MarkdownCollectionFS`
   for `embed.FS` or another virtual source; do not create a second route tree.
6. Keep the self-hosted site white-label; do not add framework implementation metadata to the chrome.
7. Run `go run ../cmd/fastr-docs check .`, `go test ./...`, and the root E2E suite for behavior changes.

## Running the site

Run the standalone site from its own directory with `go run .` for a single
process. To use the watcher from the repository root, run
`go run ./cmd/fastr-docs dev site`. The site resolves content and public assets
relative to its own directory. Validate it with
`go run ./cmd/fastr-docs check site` and `go run ./cmd/fastr-docs doctor site`.
