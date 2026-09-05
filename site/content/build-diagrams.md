---
tags: [diagrams, mermaid, plugins]
locale: en
---

# Diagrams

Mermaid diagrams render from Markdown, without relaxing the content policy that
protects the rest of the site.

```go title="docs/router.go"
router.Use(mermaid.Plugin{})
```

That is the whole setup. The plugin contributes its own assets, so nothing in
`main.go` changes.

## Writing one

{{< mermaid title="How a page is rendered" >}}
```
graph LR
    MD[Markdown file] --> R[docs.Router]
    R --> NAV[Navigation]
    R --> IDX[Search index]
    R --> PAGE[Rendered page]
    PAGE --> EXP[Static export]
```
{{< /mermaid >}}

The fence inside the shortcode is optional. It keeps the source readable in a
plain Markdown editor, and the plugin strips it.

## Why it runs in a frame

Mermaid renders by injecting a `<style>` element and emitting SVG that carries
inline styles. Documentation pages are served under `default-src 'self'`, which
blocks both, so running Mermaid in the page would mean weakening that policy
everywhere.

{{< mermaid title="Where the policy boundary sits" >}}
```
flowchart TD
    subgraph page["Docs page — default-src 'self'"]
        A[Placeholder with the diagram source]
        B[Host adapter]
    end
    subgraph frame["Sandboxed frame — style-src 'unsafe-inline'"]
        C[Mermaid]
    end
    A --> B
    B -- postMessage: source --> C
    C -- postMessage: height --> B
```
{{< /mermaid >}}

The frame is loaded with `sandbox="allow-scripts"` and no `allow-same-origin`,
which gives it an opaque origin: it cannot read the host's DOM, cookies, or
storage. `postMessage` is the only channel, and it carries two things, the
diagram source in and the rendered height back.

{{< note title="The relaxation is scoped" >}}
Only the frame document is served with `style-src 'unsafe-inline'`. Every page
on this site keeps `default-src 'self'`, and you can confirm it in the network
panel: the two responses carry different policies.
{{< /note >}}

## Without JavaScript

The page renders the diagram source as preformatted text, and the adapter
replaces it with the frame. A reader without JavaScript sees the definition
rather than an empty box.

## Themes

The frame cannot read the host's stylesheet, so the host mirrors the theme by
asking for a re-render whenever the colour scheme changes. Toggle the theme in
the header and the diagrams above follow.

## What a page downloads

Nothing, until a diagram is close to the viewport. The adapter builds each frame
on an `IntersectionObserver`; `loading="lazy"` alone does not defer, because the
browser's threshold is generous enough to load every diagram on a normal page
immediately.

The frame bundle is split by diagram type, so a flowchart fetches the flowchart
code rather than all of Mermaid. This page transfers about 1.7 MB across its two
diagrams, down from 6.9 MB before the split.

Two diagrams cost roughly twice one, and that is not a bug we can fix. Each frame
is a separate opaque origin, which is what keeps it isolated, and that also gives
it its own HTTP cache partition: frames cannot share a download, and a reload
cannot reuse one. Isolation and caching are the same tradeoff seen from two
sides.
