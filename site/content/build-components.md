---
tags: [guide, authoring, components]
---

# Markdown components

Every fastr-docs project starts with this shortcode vocabulary registered. No
Go code, no imports, no configuration. Register the same name yourself to
replace any of them.

## Admonitions

Five names, one component. Each is a callout with its tone fixed.

```md
{{< note title="Heads up" >}}
`note` and `info` are the same callout.
{{< /note >}}
```

Which renders as:

{{< note title="Heads up" >}}
`note` and `info` are the same callout. Use whichever reads better.
{{< /note >}}

{{< tip >}}
`tip` and `success` share a tone. Bodies are still Markdown, so [links](/docs)
and **emphasis** work.
{{< /tip >}}

{{< warning title="Check your order" >}}
Sibling routes need explicit `Order` values.
{{< /warning >}}

{{< danger >}}
`danger` renders with `role="alert"`.
{{< /danger >}}

`callout` takes the tone as a prop when you would rather not pick a name:

```md
{{< callout variant="warning" title="Important" >}}
Body text.
{{< /callout >}}
```

## Tabs

```md
{{< tabs >}}
{{< tab label="Go" >}}
Tabs are native `<details>` elements.
{{< /tab >}}
{{< tab label="Shell" >}}
Each set gets its own group name.
{{< /tab >}}
{{< /tabs >}}
```

{{< tabs >}}
{{< tab label="Go" >}}
Tabs are native `<details>` elements. No JavaScript runs to switch them.
{{< /tab >}}
{{< tab label="Shell" >}}
Each set gets its own group name, so two sets on one page stay independent.
{{< /tab >}}
{{< tab label="Notes" >}}
The first tab opens by default.
{{< /tab >}}
{{< /tabs >}}

## Cards

```md
{{< cards >}}
{{< card title="Routing" description="One tree owns navigation and search." href="/docs/concepts/router" >}}
{{< /card >}}
{{< /cards >}}
```

{{< cards >}}
{{< card title="Routing" description="One tree owns navigation and search." href="/docs/concepts/router" >}}
{{< /card >}}
{{< card title="Content" description="Markdown with front matter." href="/docs/concepts/content" >}}
{{< /card >}}
{{< card title="Publishing" description="Check, export, deploy." href="/docs/operate/deploy" >}}
{{< /card >}}
{{< /cards >}}

## Steps

An ordinary numbered list inside the shortcode. The numbers come from the list,
so reordering the steps renumbers them.

```md
{{< steps >}}
1. Run `fastr-docs init my-docs` to scaffold a project.
2. Edit `content/` and register routes in `docs/router.go`.
{{< /steps >}}
```

{{< steps >}}
1. Run `fastr-docs init my-docs` to scaffold a project.
2. Edit `content/` and register routes in `docs/router.go`.
3. Run `fastr-docs check .` before you commit.
{{< /steps >}}

## File tree

Write the tree as a fenced block and indent it. `filetree` reads its body
unrendered, so the indentation survives; nesting comes from it, and a trailing
`/` marks a directory.

````md
{{< filetree >}}
```
my-docs/
  content/
    index.md
  main.go
```
{{< /filetree >}}
````

{{< filetree >}}
```
my-docs/
  content/
    index.md
    getting-started.md
  docs/
    router.go
  main.go
  openapi.json
```
{{< /filetree >}}

## Smaller pieces

Collapsible sections, badges, tags, and icons are all one-liners:

```md
{{< details summary="What does strict validation check?" >}}
Broken internal links and missing heading anchors.
{{< /details >}}

{{< badge label="New" variant="success" />}}
{{< tag label="routing" href="/docs/concepts/router" />}}
{{< icon name="check" label="Supported" />}}
```

{{< details summary="What does strict validation check?" >}}
Broken internal links, missing heading anchors, duplicate sibling order, and
pages with no title or description.
{{< /details >}}

Badges and tags: {{< badge label="New" variant="success" />}} {{< badge label="Beta" variant="warning" />}} {{< tag label="routing" href="/docs/concepts/router" />}}

Icons come from the GoFastr registry: {{< icon name="check" label="Supported" />}} {{< icon name="info" />}}

## Diffs

The body is a patch, read unrendered so its line structure survives.

````md
{{< diff left="Before" right="After" >}}
```
 router := docs.NewRouter(
-    docs.WithSiteName("Docs"),
+    docs.WithSiteName("Acme Docs"),
 )
```
{{< /diff >}}
````

{{< diff left="Before" right="After" >}}
```
 router := docs.NewRouter(
-    docs.WithSiteName("Docs"),
+    docs.WithSiteName("Acme Docs"),
+    docs.WithTemplate(docs.TemplateBlueprint),
 )
```
{{< /diff >}}

## Code block options

Fenced blocks take options after the language. A fence with no options renders
exactly as it always did.

{{< filetree >}}
```
title="path/to/file"   a filename header
{1,4-6}                highlight those lines
showLineNumbers        number the gutter
scroll                 cap the height and scroll inside
```
{{< /filetree >}}

Opening a fence with `go title="docs/router.go" {3,4} showLineNumbers` gives:

```go title="docs/router.go" {3,4} showLineNumbers
router := docs.NewRouter(
    docs.WithSiteName("Acme Docs"),
    docs.WithTemplate(docs.TemplateBlueprint),
    docs.WithSearchIndexPath("/assets/search.json"),
)
```

A title with no language works too:

```title="notes.txt"
Plain text, with a filename header and nothing tokenized.
```

## Three shapes

A shortcode is registered as one of three kinds, and a name resolves to exactly
one of them.

| Kind | Receives | Used by |
| --- | --- | --- |
| `MarkdownComponent` | its body as rendered Markdown | most of them |
| `MarkdownContainer` | its nested shortcodes as separate children | `tabs`, `cards`, `hero` |
| `MarkdownRawComponent` | its body unrendered | `diff`, `filetree` |

The container exists because a component receives one merged blob and cannot
tell where one child ends and the next begins, which is exactly what a tab set
needs to know.

The raw kind exists because rendering a body as Markdown destroys line
structure. A patch collapses onto one line, and the code block's own chrome
ends up inside the text. Write those two as a fenced block.

## Replacing a default

Registering a name replaces whatever held it:

```go
router.RegisterMarkdownComponent("note", func(props map[string]string, body render.HTML) render.HTML {
    return ui.Callout(ui.CalloutConfig{Title: props["title"], Variant: ui.StatusNeutral}, body)
})
```

Pass `docs.WithoutDefaultComponents()` to start from an empty vocabulary, where
any unregistered shortcode fails the build instead of rendering.
