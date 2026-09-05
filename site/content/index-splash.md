---
template: splash
locale: en
hero:
  eyebrow: Reusable documentation for GoFastr
  title: Build docs around one route tree.
  tagline: Pages, screens, search, OpenAPI, and offline delivery all derive from one explicit Router.
  actions:
    - text: Read the docs
      link: /docs
      variant: primary
    - text: Start building
      link: /docs/getting-started
      variant: secondary
---

## One router, every surface

Choose Markdown for durable content or a typed screen for interaction.
Navigation, search, breadcrumbs, and export stay in sync because they all read
the same tree.

{{< cards >}}
{{< card title="Pages" description="Markdown with front matter, headings, local search, table of contents, and offline eligibility." href="/docs/concepts/content" >}}
{{< /card >}}
{{< card title="Screens" description="Typed GoFastr components for playgrounds, API explorers, and product workflows." href="/docs/build/screens" >}}
{{< /card >}}
{{< card title="Plugins" description="Extend the same router without forking the docs shell." href="/docs/build/plugins" >}}
{{< /card >}}
{{< /cards >}}

## Written for agents as much as people

Generated projects ship `agents/claude.md`, a skill set under `.agents/skills`
mirrored to `.claude/skills`, and the live host publishes `/llms.txt`, an agent
card, and an MCP introspection endpoint.

{{< cards >}}
{{< card title="Markdown components" description="A shortcode vocabulary registered by default: admonitions, tabs, cards, steps, file trees." href="/docs/build/components" >}}
{{< /card >}}
{{< card title="Offline and PWA" description="Export a static, installable site with every document precached." href="/docs/operate/offline" >}}
{{< /card >}}
{{< /cards >}}

## Start here

{{< steps >}}
1. `go run ./cmd/fastr-docs init my-docs` scaffolds a strict project.
2. Edit `content/` and register routes in `docs/router.go`.
3. `fastr-docs check .` validates links, anchors, order, and skill drift.
{{< /steps >}}
