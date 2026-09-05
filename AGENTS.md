# Working in the fastr-docs repository

This is the fastr-docs package itself, not a generated site. If you are looking
for guidance on authoring a docs project, that lives in the skills under
`internal/cli/templates/starter/.agents/skills/`.

## Toolchain

`go.mod` requires Go 1.27, which arrives through toolchain switching rather
than a local install. The only working local toolchain is Go 1.25.5 at
`C:\Users\Dom\sdk\go`, and `GOROOT` and `PATH` point there.

If `go` ever goes missing from `PATH` again, that is the cause: two Go 1.26.3
trees, `C:\Program Files\Go` and `C:\Users\Dom\go\go`, still sit on `PATH` with
`go.exe` deleted from them and only `gofmt.exe` left. They are dead but
harmless as long as `C:\Users\Dom\sdk\go\bin` comes first.

## Modules

Three, deliberately separate:

- The repository root is `github.com/DonaldMurillo/fastr-docs`, package `docs`.
- `site/` is the dogfood documentation project. It has its own `go.mod` with a
  local `replace` pointing at the parent, so it can be developed standalone.
- `e2e/fixtures/non-cli/` is a hand-authored fixture with its own `go.mod`. It
  exercises the public API without the CLI generator, which is how the two
  paths are kept from drifting.

Changing the public API means updating all three.

## Layout of the root package

Everything is package `docs` at the repository root. The rough split:

| Area | Files |
| --- | --- |
| Route tree and options | `router.go` (types, registration, search, validation, mounting) |
| Layout rendering | `router_chrome.go` (header, sidebar, variant selectors, palette) |
| Markdown content | `content.go`, `content_validation.go`, `markdown_render.go`, `shortcodes.go` |
| Publication | `blog.go`, `blog_surface.go`, `versions.go` |
| Presentation | `theme.go`, `theme_css.go`, `styles.go`, `branding.go`, `labels.go`, `toc.go` |
| Browser runtime | `runtime.go` (the JS served as `docs.js`) |
| Export surfaces | `manifest.go`, `assets.go`, `agent_assets.go`, `notfound.go`, `openapi.go` |
| Extension points | `plugins.go`, `plugin/openapi/` |
| Generator | `internal/cli/`, `cmd/fastr-docs/` |

There is no `pages/` directory. Root-level `app.js`, `docs.js`, `index.html`,
and `styles.css` were a manual lab and are gone. The browser runtime is
generated from Go by `RuntimeJS()`, and CSS from `styles.go` and `theme_css.go`.

## The starter template

`internal/cli/templates/starter/` is copied verbatim by `fastr-docs init`.
`.tmpl` files are rendered with `templateData`; everything else is copied as-is.

The embed directive uses `all:` because a plain glob skips every name starting
with a dot. That previously dropped the template's `.gitignore` from generated
projects without any error.

Skills are authored once under `.agents/skills/` and mirrored to
`.claude/skills/` at init time by `mirrorSkills`. Add a skill by creating the
directory under `.agents/skills/`; `starterSkills()` reads the embedded
filesystem, so no code change is needed.

## Checks

```sh
go build ./... && go vet ./... && go test ./...
cd site && go test ./...
cd e2e && npm test          # Playwright, ~3 minutes, generates a fresh starter
```

The e2e suite has two browser projects, `chromium` and `mobile-chromium`, and
runs serially. `global-setup.mjs` builds the starter, the fixture, and a static
export, then serves them on ports 4175-4181. It probes for a Go binary itself,
including `~/sdk/go/bin/go.exe`.

Visual snapshots live in `e2e/tests/visual.spec.mjs-snapshots/` and are
platform-specific. A theme or layout change will need them regenerated.

## Conventions worth keeping

Strict validation is the default and should stay that way. Asset paths that
the browser runtime fetches must be configurable options, not string literals:
`WithPagefindPath` and `WithSearchIndexPath` exist because a hardcoded index
path silently degraded the command palette to an unfiltered route list.

Prose in this repository follows the unslop rules in the user's global
configuration. No em dashes, no puffery, sentence-case headings.
