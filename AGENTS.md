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

## Markdown shortcodes

Three registrable shapes, all in `shortcodes.go`, merged per page by
`Router.vocabulary`. A name resolves to exactly one kind, and registering it
retires the others.

`MarkdownComponent` gets its body as rendered Markdown. `MarkdownContainer`
gets its nested shortcodes as `[]MarkdownChild` plus any prose between them,
which is the only way to build tabs or card grids. `MarkdownRawComponent` gets
the body unrendered, for content that is data rather than prose.

`diff` and `filetree` are raw for a concrete reason: rendering their bodies
first collapses every line into one and pulls the code block's own chrome
("3 lines", the copy glyphs) into the text. `filetree` also cannot wrap a
Markdown list, because `core/markdown` does not support nested lists and
flattens them into one item joined by `<br>`.

The defaults live in `markdown_components.go` and are registered in
`NewRouter` before options run, so a project still overrides any name.
Every prop is sanitized through a fixed table: `ui.Callout` panics on an
unregistered `StatusVariant`, `tabs.New` panics without a name or tabs, and
`ui.Banner` panics without a title. Those inputs come from Markdown files, so
none of them may reach a component unchecked.

Shortcode bodies are themselves `.ui-markdown` blocks nested inside the page's.
`styles.go` resets `.ui-markdown .ui-markdown` because the page-level prose
rules otherwise add a max width and 75px of bottom padding inside every
callout and card.

## Code fences

`codefence.go` lifts option-carrying fences out of the source before GoFastr's
Markdown parser sees them, and renders them through `ui.CodeBlock`. Supported
after the language: `title="..."`, `{1,3-5}`, `showLineNumbers`, `scroll`.

A fence with no options is passed through completely untouched, and a test
asserts the source is byte-identical, so the interception stays provably
opt-in.

This exists because `core/markdown`'s `renderFence` takes the whole remainder
of the fence line as the language. Writing ` ```go title="x" ` there does not
merely ignore the option, it emits `class="language-go title=&quot;x&quot;"`
and then fails to match `go`, silently costing you syntax highlighting.

Line highlighting wraps the line's own HTML, because `CodeBlock` owns the
`.ui-code-block__line` wrapper and offers no hook to mark one.

## Translation

`labels.go` holds every framework-owned string. There are no hardcoded UI
strings left in `blog_surface.go`, `notfound.go`, `toc.go`, `router_chrome.go`,
or `versions.go`; a grep for `render.Text("Capitalized")` in those files should
stay at zero.

`mergeUIStrings` uses reflection rather than one if-statement per field. There
are over eighty labels, and the previous hand-written merge was already 47
lines for thirteen.

Labels holding `%s`/`%d` go through `formatLabel`, which returns the label
unchanged when a translation dropped the placeholder, instead of emitting Go's
`%!(EXTRA ...)` into a page. `go vet` treats `formatLabel` as a printf wrapper,
so a test passing an argument to a label without a directive has to hide the
literal behind a table.

The blog's client-side search filter reads its labels from `data-*` attributes
on the search root, because the runtime JS is a constant and cannot see the
Router.

Locale fallback lives in `i18n.go`. `localeAllows` replaces the old inline
locale filter in `routePublished`, and the family index it consults is
invalidated in `rebuildTree`. Coverage reporting is advisory on purpose:
wiring it into `ContentIssues` made `Validate` fail on any partially
translated site, which a test caught immediately.

## Runtime assets

`runtime_assets.go` collects `docs.js`, the search index, the export manifest,
and every `RuntimeAssetPlugin`'s files. `MountRuntimeAssets` serves them,
`WriteRuntimeAssets` exports them, `RuntimeAssetNames` feeds the precache list
and the host's script tags.

Before it, the generated `main.go` hand-wrote each file twice, once for serving
and once for export, so any new asset-bearing plugin meant editing every
project. `internal/cli/cli.go`'s `main.go` marker list asserts both calls are
present, which is what keeps a generated project on this path.

Plugin file names are validated. They come from a plugin rather than the
project, so a collision with `docs.js` or a `../` escape is an error.

The e2e fixture at `e2e/fixtures/non-cli/` still wires its assets by hand, on
purpose: it proves the lower-level API works without the convenience wrapper,
and it uses a non-default `/__manual` prefix.

## Mermaid and KaTeX: blocked on a CSP decision

Measured in a real browser against the live policy, not reasoned about:
`default-src 'self'` blocks both dynamically injected `<style>` elements and
inline `style` attributes. Chrome logged two violations, and neither the
injected stylesheet nor the inline attribute applied.

Mermaid renders by injecting a `<style>` element and emitting SVG carrying
inline styles, so it cannot work under the current policy. There is no way to
ship it without one of:

1. Adding `style-src 'unsafe-inline'`, which weakens the strict CSP for every
   generated site.
2. Mounting it through GoFastr's `framework/pluginhost` sandboxed iframe, which
   keeps the policy intact but is substantially more work.

Vendoring is its own decision: mermaid.min.js is roughly 3MB, embedded into the
binary of any project that opts in. `plugin/openapi` shows the pattern, and a
separate package means unused ones are never linked in, so the cost only lands
on projects that ask for it.

This is a product and security-posture call, so it is deliberately not made
here. `RuntimeAssetPlugin` exists and is tested, so whichever way it goes the
plumbing is ready.

## Known GoFastr limitations worked around here

Ticket these rather than re-discovering them:

- `core/markdown` does not parse fence info strings beyond the language.
- It has no nested-list support, and flattens one into a single `<li>` joined
  by `<br>`. This is why `filetree` parses raw indentation.
- It does not support four-backtick fences, so a fenced example cannot be shown
  inside another fence. Do not write ` ````md ` blocks in content.
- `CodeBlock` has no line-highlight, diff, word-highlight, or wrap option.
- There is no `FileTree`, content `Steps`, `CardGrid`, `LinkCard`, or generic
  `ui.Tabs`; `core-ui/patterns/tabs` is the only generic tabset.

## Skill drift

The same skills live in four places: the template source, a generated
project's two copies, and `site/`'s two copies. Two checks keep them honest,
and both must stay in place.

`fastr-docs check` fails when a project's `.claude/skills` disagrees with its
`.agents/skills`, reporting stale, missing, and orphaned files by name.
`fastr-docs sync-skills` repairs it; `--prune` also deletes files that exist
only under `.claude/skills`, which is opt-in because such a file may be a
deliberate Claude-only skill.

`TestDogfoodSiteSkillsMatchTheTemplate` covers this repository: it fails if
either of `site/`'s copies drifts from the starter template. After editing a
template skill, copy it to `site/.agents/skills/` and `site/.claude/skills/`.
It normalizes line endings, because the checked-in site copies arrive with
CRLF on Windows while the embedded template keeps LF.

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
