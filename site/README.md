# fastr-docs documentation site

This is fastr-docs’ own documentation project. It is a standalone generated
site inside the repository, with a local replace pointing at the package in
the parent directory. It is also the first dogfood project for the Router,
OpenAPI plugin, PWA export, command palette, badges, and agent references.
The live host also exposes GoFastr's read-only MCP introspection and discovery
surfaces; static exports contain the docs and agent assets, not the live MCP
transport.

## Run it

```sh
# from the repository root
go run ./cmd/fastr-docs dev site
```

The GoFastr dev loop opens <http://localhost:8080> by default. Running
`go run .` directly still uses the site's fallback port, <http://localhost:3079>,
but does not watch or refresh on changes.

The site is installable as a PWA on `localhost` and on HTTPS. For the complete
offline experience, export and deploy `dist`; GoFastr rotates the static worker
cache when the exported content changes. Use `fastr-docs upgrade site` to
review framework migrations when updating the GoFastr dependency.

## Validate it

From the repository root:

```sh
go run ./cmd/fastr-docs check site
go run ./cmd/fastr-docs doctor site
```

The site has its own `go.mod`, so it can also be developed independently:

```sh
cd site
go test ./...
go run .
```

## Export it

```sh
cd site
go run . --export site-dist --export-base /docs
```

For Pagefind search, set the search backend before exporting or use the CLI
from the site directory:

```sh
cd site
DOCS_SEARCH_BACKEND=pagefind go run . --export site-dist
```

## Source layout

- `docs/router.go` registers the self-documentation route tree.
- `content/` contains the Markdown reference and guide pages.
- `docs/examples.go` contains the typed GoFastr playground.
- `openapi.json` is an intentionally small plugin fixture, not a fastr-docs HTTP API.
- `agents/claude.md` and `.agents/skills/` describe the authoring workflow.
