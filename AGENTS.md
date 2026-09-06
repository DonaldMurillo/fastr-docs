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

### Fences longer than three characters

A plain fence of any length passes through to GoFastr untouched. Since v0.83.0
its parser reads the marker length and closes a block only on a run at least as
long as the opener, so ````md holding a ``` example renders as one block.
`extractRichCodeFences` still has to know the length: it skips to the real
closing line so an optioned fence shown inside the example is not lifted out of
it.

Before v0.83.0 the parser read exactly three characters, took ````md as ``` with
a language of "`md", and let the first inner ``` close the block. fastr-docs
lifted every longer fence itself to work around that, and no longer does.

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

## Chrome in more than one language

`WithUIStrings` sets labels for the whole Router, which is all a site needs when
one build serves one language. A build serving several needs
`WithLocaleUIStrings(locale, UIStrings{...})`, because otherwise a Spanish page
arrives wrapped in English furniture: "Contents", "On this page", "Search".

Resolution is per route, not per Router: `uiAt(path)` and `uiForRoute(route)`
read the route's own locale. Chrome renderers that have a path or a route use
those; anything built once for the whole site cannot.

Two traps here:

- `mergeUIStrings` substitutes the English defaults whenever its base looks
  uninitialized. That is right for a Router built with no options and wrong for
  accumulating a locale: it fills every field of a partial translation with
  English, which then overwrites the project's own Router-wide labels.
  `overlayUIStrings` is the one to use for a locale.
- The command palette modal is mounted once for the whole site, so its
  placeholder cannot vary by locale. The header's search trigger is rendered
  into every page, so `searchTrigger(path)` is deliberately not memoized with
  the palette and does follow the page.

`WithLocaleNames` gives the selector real language names. Go's standard library
has no locale display names, so the project supplies them rather than the
framework guessing; without it the selector shows "es", which is the worst
possible label for the reader who needs it.

`headerItems` keys sections by `variantFamily` so a translated section does not
appear beside the original as its own tab. The home route is not a tab but still
claims its family, or a translated home lands in the nav as one. Tab labels stay
in the language they were registered in: translating a route title is the
project's call.

## Pairing a translation with its original

Three things had to be true before a translated section could replace the
original rather than sit beside it, and each was a gap rather than a decision.

`GroupConfig.Locale` exists because a page reads its locale from front matter
and **a group has no front matter to read**. Without it `variantFamily` had no
locale segment to strip, so `/es/examples` and `/examples` were unrelated
families: the selector had nowhere to go and the nav tab stayed English.

`effectiveLocale` treats a route with no declared locale as being in
`WithLocaleFallback`'s default. Requiring `locale: en` on thirty original pages
to get a language selector is busywork, and omitting it fails silently. This is
deliberately only about pairing; `localeAllows` decides publication and is
untouched, so an unmarked route still serves in every locale build.

Pairing is resolved on the Router, in `familyOf`, never by calling
`variantFamily` directly. `variantFamily` is the path-shaped family, the path
with its locale and version segments removed, and it is all a folder-for-folder
translation needs. `ContentMetadata.TranslationOf` (`translation_of:` in front
matter, `GroupConfig.TranslationOf` for a group) joins the named route's family
instead, which is what a translated slug needs: `/es/ejemplos/arbol-de-rutas`
cannot pair with `/examples/route-tree` by shape. The reference is followed
through its target, so pointing at another translation lands in the same
family. A reference nothing serves is kept as the family, so the page pairs
with nothing rather than with the wrong thing, and `translationIssues` names
it: dangling, self, or same-language references are `ContentIssues`.

`alternatesFor` derives the `hreflang` map from the family: every published
variant in another language, in both directions, under any declared
`alternates:`. `routeHasMetadata` counts it, or a page with no metadata of its
own would carry no head and the pairing would be invisible from that side.

`headerVariant` walks the whole tree, not `Routes()`. `Routes()` returns pages
only, and a translated section is usually a group, so a `Routes()` search found
nothing.

### The locale home

A locale home such as `/es` is the translation of `/`, not a section of the
site. Two things follow, and both were wrong before:

`sidebarRoots` descends past it to the section actually being read. Left as the
active root it wraps the whole translated tree in an extra level the default
locale does not have, so a Spanish reader got "Espanol > Documentacion > ..."
where an English one got "Documentation > ...".

`localeHome` resolves the home link per language. Translating the label alone
was worse than not translating it: "Inicio" still pointed at `/` and quietly
took the reader out of the Spanish site. `sidebarItems` renders any home-family
route as a link with no children, and `homeFirst` puts it at the top, because a
translated home sits among the top-level sections and cannot compete on explicit
`Order`.

The test for this compares the two sidebars: same number of links, each home
pointing inside its own language.

### Search and the document language

Pagefind decides which language index a page belongs to by reading
`<html lang>`, and a screen reader reads the same attribute to choose
pronunciation rules (WCAG 3.1.1). `Router.LanguageFor(path)` answers per page,
and the hosts hand it to `app.WithLangFunc`, which GoFastr added in v0.83.0.
It resolves the route's effective locale, and for a path with no route, such as
a 404 under `/es`, the language of the deepest route above it, so a missed
Spanish URL still answers in Spanish. The live server and the export both carry
it.

Verified with Pagefind 1.5.2 over this site's export: "Discovered 2 languages:
en, es", separate `pagefind.en_*.pf_meta` / `pagefind.es_*.pf_meta` and
separate `en_*.pf_index` / `es_*.pf_index` chunks. Searching "traduccion" from
a Spanish page returns hits only under `/es`; "translation" from an English page
returns none of them.

Before v0.83.0 GoFastr wrote one host-wide value, and `WriteExportLocales`
rewrote the attribute per page after the export. That is gone: there is nothing
left for it to fix.

The JSON backend is a different mechanism and needed its own fix: one index
holds every locale, so `renderJSONResults` filters to the page's language. The
runtime reads that from `data-fastr-docs-locale` on the search trigger, which
is rendered into every page. A site that declares no locales is unfiltered.

### The pager

`docPager` builds a `ui.DocPager` for `DocLayoutConfig.Pager`, with the
direction lines from the page's `UIStrings.Previous` and `Next` through
`PrevDirLabel` and `NextDirLabel`, which GoFastr added in v0.83.0. Neighbours
are drawn from the reader's own language; walking every published route stepped
a reader at the edge of the Spanish tree into the English one.

### Surfaces built once for the whole site

Three surfaces have no route to read a language from, and each needed a
different answer.

The **404** is handed to uihost at startup, but `RenderNotFound(path)` does
receive the requested path per request, and that is enough:
`NotFoundScreen.Locales` matches the URL by longest prefix.
`Router.NotFoundScreen()` fills it in from the registered locales so a project
does not restate what the Router already knows.

The **command palette modal** is mounted once at an HTTP route, so its
placeholder cannot be rendered per language. Its strings ride on the search
trigger, which is per page, and the runtime applies them when the modal opens.
Localizing at init is too early: the modal is not in the document until it is
opened, so `watchPaletteForLocalization` hooks the trigger click and the input's
focusin instead.

The **blog** is a collection per language. Every label in `blog_surface.go`
is read through `blogLabels(prefix)`, which is `uiAt(prefix).Blog`, so a
collection registered with `DefaultLocale: "es"` renders in the Spanish
`BlogStrings` while the English one beside it does not; `formatBlogDate` takes
the prefix for the same reason. The generated views (archive, tags, authors,
search, pages) carry the collection's locale as metadata, which is what makes
`/es/blog/archive` pair with `/blog/archive` and read the right strings.
`blogRootFor` groups a collection's screens by its prefix rather than its tree
root, or `/es/blog`, which sits under the Spanish home, wore the docs sidebar.
The site has both collections; a grep for `UIStrings().Blog` should stay at
zero.

### The header after a client-side navigation

GoFastr's runtime keeps every layout layer it shares with the destination and
swaps only the deepest changed slot. The header is in the outermost layer, so
it was rendered for the first page loaded and never again: a Spanish reader who
clicked into an untranslated section kept Spanish tabs, a Spanish search label,
a language selector still aimed at the previous page, and `lang="es"` over
English content.

Every page therefore carries its own header in
`<template data-fastr-docs-chrome>` inside the swapped region
(`chromeTemplate`, appended by `pageComponent` and by the screen wrapper, which
now wraps every screen). `syncChrome` in `runtime.go` runs first in `init()`
on every navigation and copies the parts that change into the live header: the
two nav lists, the selector container, the search trigger's `data-fastr-docs-*`
attributes and label, and `<html lang>`. The selector container is rendered
even when empty so there is always somewhere to copy into. On a full load the
copy is a no-op.

The sidebar is in the section layer, which is shared across languages by name,
so `layoutLocaleSuffix` keys section layouts by language on a multilingual site
(`docs-section-es`, `blog-en`). A single-language site keeps its layout names
and the `.layout-*` classes derived from them; `styles.go` selects the blog
layout with `[data-fui-layout^="blog"]` for that reason.

One trap when adding `data-fastr-docs-search-*` attributes:
`TestLayoutUsesNativeCommandPaletteSearch` guarded against the legacy search
component with `strings.Contains(header, "fastr-docs-search")`, which the new
placeholder attribute tripped. It matches a class token now.

## Serving below a path prefix

Route paths inside a page are root-relative (`/es/docs/guide`), and the
export rewrites the URL attributes GoFastr knows about. Everything the runtime
does with a path is not an attribute: the language selector navigates to an
`<option value>`, the drawer trigger and sidebar compare `data-fastr-docs-*`
route paths against `location.pathname`, JSON search results are built from
`entry.path`, and Pagefind's result URLs are relative to the folder it
indexed. On a GitHub project page every one of those was off by the
repository name.

The base is stamped into `data-fastr-docs-base` on the search trigger by the
export's `rewriteRuntimeURLs` (site and starter), and the runtime has three
helpers around it: `docsBase()`, `withBase(path)` for navigation, and
`normalizeDocsPath(path)`, which strips it so a path from the location and a
path from the page compare equal. Pagefind gets `options({baseUrl})` before
the first search. The e2e harness exports the site below `/prefix` and serves
it from a folder of that name, so the static tests exercise the real shape.

Two more things were wrong on a static host and are fixed in GoFastr v0.84.1,
so they need nothing here now: the exported route graph kept root-relative
paths, so the runtime never recognised a link below a base and every click
was a full page load (#403); and the exported per-component stylesheet links
lacked the `data-fui-style` marker the runtime dedupes on, so every static
page loaded each component stylesheet a second time after `app.css` and the
site's overrides lost the cascade (#401). The e2e tests on the prefixed
export keep both pinned from this side.

## Runtime assets

`runtime_assets.go` collects `docs.js`, the search index, the export manifest,
and every `RuntimeAssetPlugin`'s files. `MountRuntimeAssets` serves them,
`WriteRuntimeAssets` exports them, `RuntimeAssetNames` feeds the precache list
and `RuntimeScriptNames` the host's script tags. See "Page scripts vs served
assets" for why those are two lists.

Before it, the generated `main.go` hand-wrote each file twice, once for serving
and once for export, so any new asset-bearing plugin meant editing every
project. `internal/cli/cli.go`'s `main.go` marker list asserts both calls are
present, which is what keeps a generated project on this path.

Plugin file names are validated. They come from a plugin rather than the
project, so a collision with `docs.js` or a `../` escape is an error.

The e2e fixture at `e2e/fixtures/non-cli/` still wires its assets by hand, on
purpose: it proves the lower-level API works without the convenience wrapper,
and it uses a non-default `/__manual` prefix.

## Diagrams (plugin/mermaid)

Resolved the CSP problem the earlier spike found, using the isolation model of
the editor plugin in `gofastr-plugins/mermaid`, reduced to what a docs page
needs: no save path, no capability handshake, no document store.

The diagram renders in a frame document loaded with `sandbox="allow-scripts"`
and no `allow-same-origin`. Only that document is served with
`style-src 'unsafe-inline'`; pages keep `default-src 'self'`. Tests assert both
halves of that.

Two things are easy to get wrong here, and both fail silently:

- The frame's opaque origin makes its *own* stylesheet and bundle cross-origin
  requests. Without `Cross-Origin-Resource-Policy: cross-origin` on them the
  browser blocks with `ERR_BLOCKED_BY_RESPONSE.NotSameOrigin` and the diagram
  never appears. The host adapter must **not** carry that header; it is an
  ordinary page script.
- The frame cannot read the host stylesheet, so dark mode is mirrored by asking
  it to re-render, not by CSS.

`assets/frame/` is a generated esbuild bundle, split rather than single-file.
Mermaid loads each diagram type through a dynamic import, so `format: 'esm'` with
`splitting` emits them as separate chunks and a flowchart fetches the flowchart
code instead of all 3.4MB. Rebuild with
`cd plugin/mermaid/js && npm install && npm run build`. It is a separate Go
package, so nothing links it in unless the project imports it.

Splitting matters more here than it would in a page, because of a property that
took measuring to find: **every frame is a distinct opaque origin, so it gets its
own HTTP cache partition.** Two diagrams on one page cannot share a download, and
a reload cannot reuse either. This repo's diagrams page transferred 6.9MB, warm
cache or not; it is now 1.7MB. Shrinking what one frame needs is the only lever
that exists, so do not "simplify" the build back to a single file.

Three more things that fail quietly here:

- A module entry is fetched in **CORS mode**, and the frame's origin is literally
  `null`, so the chunks need `Access-Control-Allow-Origin: *` on top of CORP.
  Without it the entry fails with a CORS error and nothing renders. They are
  public static files with no credentials, which is the case the wildcard is for.
- The framed files had **no `Cache-Control` and no validator**, so nothing cached
  them even within a single frame. They are content-addressed now and immutable
  for a year; the frame document stays revalidated because it carries the hashes.
- `iframe loading="lazy"` does **not** defer in practice. Chrome's threshold is
  generous enough that both diagrams on a normal page loaded immediately. The
  adapter gates frame creation on an `IntersectionObserver` instead, and forces
  every diagram in on `beforeprint`, since printing has no viewport to scroll.

## Math (plugin/katex)

Takes the opposite approach to diagrams, on purpose. Math renders in the page,
with no frame and no relaxed policy.

KaTeX builds its layout as DOM nodes and assigns sizes through CSSOM
(`node.style.height = "0.68em"`). CSP polices style attributes in parsed markup
and `<style>` elements, not CSSOM, so the strict page policy holds untouched.
Verified rather than assumed: the rendered page carries about 150 inline style
attributes and reports zero policy violations. The e2e spec asserts both numbers
together, because either alone proves nothing.

Rendering in the page is also the only way inline math works. A frame is a
rectangle and cannot share a line box with the sentence around it.

The one place KaTeX would trip the policy is its own error path, which calls
`setAttribute("style", "color:...")`. The runtime sets `throwOnError` and renders
its own error text instead, so a mistyped formula never produces a violation.
`\fcolorbox` and `\textsf` shadows take the same path and are the known
exception; nothing else does.

The dollar scanner in `dollar.go` follows Pandoc's rules: no whitespace after
the opening `$`, none before the closing `$`, no digit after it, and a backslash
escapes. Fenced blocks and inline code spans are skipped outright. Those rules
are why `from $5 to $10` and `$HOME` stay prose, and the table in
`site/content/build-math.md` documents them for writers.

`assets/` is generated from two entry points. `katex-loader.js` is 783 bytes and
is the only file every page carries; it looks for a placeholder and pulls
`katex.js` (266KB) and the stylesheet in only when it finds one, re-checking on
DOM mutation so a client-side navigation into a math page still works. Measured:
a page without math transfers 783 bytes of KaTeX, this repo's math page about
347KB. `PageScripts()` returns the loader, never the renderer.

The stylesheet also declares 20 font faces and the browser fetches only the ones
a formula uses, so the fonts need no lazy-loading machinery of their own.

Rebuild with
`cd plugin/katex/js && npm install && npm run build`. The placeholder CSS lives
in `build.mjs` and is appended to the stylesheet, so the plugin owns its
presentation and a project needs no CSS of its own.

## Markdown source transforms

`markdown_transform.go` lets a plugin rewrite Markdown source before the parser
runs. A shortcode can only claim text a writer wrapped in `{{< >}}`; `$x^2$` in
the middle of a sentence needs this instead.

The transform receives a `place` function, hands it rendered HTML, and splices
the returned marker into the source. The parser only ever sees the marker. It is
the same mechanism option-carrying code fences already use, and the three marker
namespaces (`FASTRDOCSSHORTCODESLOT`, `FASTRDOCSFENCESLOT`,
`FASTRDOCSTRANSFORMSLOT`) are distinct so slot zero of one cannot eat slot zero
of another.

Transforms run before shortcodes expand, so a transform can claim text inside a
shortcode body. That is why the math scanner skips code but a shortcode cannot
opt out of it.

## Page scripts vs served assets

`RuntimeAssetNames` lists everything served. `RuntimeScriptNames` lists only what
belongs in a `<script>` tag, and they are not the same set.

Picking page scripts by file extension put the 3.3MB Mermaid frame bundle on
every page, where nothing loaded it: the frame document fetches it itself. A
plugin now declares `PageScripts()` to narrow this; one that does not still
contributes all its `.js`, which is right for a single-file runtime like
`openapi.js`.

## Known GoFastr limitations worked around here

Ticket these rather than re-discovering them:

- `GroupConfig` had no locale, so a translated section could not pair with its
  original. Worked around by carrying `Locale`/`Version` on the group route.
- `ui.Markdown` maps only `title=` and `showLineNumbers` from a fence's
  options; the rest of the info string lands in `data-meta`. Line highlights
  and `scroll` still need `extractRichCodeFences`.
- `core/markdown` has no nested-list support, and flattens one into a single
  `<li>` joined by `<br>`. This is why `filetree` parses raw indentation.
- `CodeBlock` has no line-highlight, diff, word-highlight, or wrap option.

Fixed upstream in v0.83.0, and the workarounds removed here: the per-page
document language (`app.WithLangFunc`), the pager's direction labels
(`ui.DocPager.PrevDirLabel`/`NextDirLabel`), and fences longer than three
characters.
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
