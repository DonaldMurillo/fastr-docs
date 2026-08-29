# Content authoring

Markdown is the default authoring format because it is portable, reviewable, searchable, and easy for both humans and agents to edit. A page registration supplies the product metadata; the Markdown file supplies the reader experience.

## A durable page

```go
router.MustPage("/docs/decisions", docs.PageConfig{
    Title:       "Architecture decisions",
    Description: "The choices behind the product.",
    SourcePath:  "content/decisions.md",
    SearchText:  "architecture decisions tradeoffs",
    Tags:        []string{"architecture", "reference"},
    Order:       6,
    Offline:     true,
})
```

Use `Description` for navigation and search summaries. Use `SearchText` when the page contains important terms that are not written verbatim in the visible article. Tags make future external indexers or filtered surfaces possible without changing the route model.

## Markdown conventions

- Start with one page-level heading.
- Use headings to create useful in-page navigation.
- Keep code examples complete enough to copy.
- Link to route paths rather than duplicating product URLs.
- Keep paragraphs short and lead with the decision or outcome.
- Prefer an explicit warning or note over hidden assumptions.

The shell derives a sticky table of contents from the headings. On narrower layouts it becomes the **On this page** selector so the document remains readable without a second column.

## When to use a screen

Markdown should explain a feature. A [screen](/docs/build/screens) should let the reader operate it, inspect state, or trigger an action. Keeping those roles separate makes the content easy to export while preserving room for rich examples.

## Reusable content components

When prose needs a consistent visual treatment, register a
`MarkdownComponentsPlugin` and use a shortcode. The component receives parsed
string props plus a server-rendered Markdown body:

```md
{{< note title="Decision" >}}
The body can contain **Markdown**, links, and nested registered components.
{{< /note >}}
```

Unknown names and unclosed blocks fail strict validation before the site is
served. This keeps the authoring format portable while still allowing typed
GoFastr surfaces where they add real value.

{{< callout title="A typed authoring escape hatch" >}}
The generated project registers this shared component vocabulary while keeping the page body Markdown.
{{< /callout >}}
