---
tags: [ui, components, gofastr, reference]
---

# Framework UI

fastr-docs uses GoFastr's native UI components for the docs shell, content blocks, forms, data views, and interaction states. A page and a typed screen use the same component library, so a custom surface does not need a second visual system.

## Docs shell

The default Router layout provides these native pieces:

| Area | Components | Used for |
| --- | --- | --- |
| Site chrome | `SiteHeader`, `SiteFooter`, `DocLayout` | Brand, section navigation, article framing |
| Navigation | `Sidebar`, `AnchoredRail`, `TableOfContents`, `StepRail` | Route tree, in-page headings, guided sequences |
| Search | `CommandPalette`, `GlobalSearch`, `SearchInput` | `Ctrl+K` or `⌘K` route search |
| Utility | `ThemeToggle`, `BackToTop`, `SkipLink`, `Tooltip`, `Menu` | Theme, keyboard access, page controls |

The Router supplies the route data. GoFastr owns drawer behavior, focus, keyboard shortcuts, scroll locking, and component lifecycle.

## Native component catalog

The framework UI package is the complete native vocabulary available to typed
screens and Markdown component adapters. Components are grouped here by the
job they do; the names are the public GoFastr constructors.

### Docs shell and navigation

| Components | Purpose |
| --- | --- |
| `AnchoredRail`, `StepRail`, `TableOfContents` | Sticky in-page, step, and heading navigation |
| `BackToTop`, `SkipLink` | Long-page and keyboard navigation |
| `CommandPalette`, `GlobalSearch`, `SearchInput`, `ShortcutHint` | Search, command access, and keyboard hints |
| `DocLayout`, `Sidebar`, `SidebarBody` | Documentation page framing and responsive route navigation |
| `Menu`, `Tooltip` | Compact menus and contextual help |
| `SiteHeader`, `SiteFooter`, `ThemeToggle` | Site chrome, identity, and light/dark/auto mode |

### Reading and technical content

| Components | Purpose |
| --- | --- |
| `Markdown`, `Section`, `PageHeader`, `Hero`, `HeroSplit` | Prose, sections, page introductions, and landing layouts |
| `CodeBlock`, `CodeTabs`, `CopyButton`, `TerminalBlock` | Source, language variants, clipboard actions, and transcripts |
| `Callout`, `Banner`, `FactBox`, `StatusBadge`, `StatusPill`, `Muted` | Context, state, and secondary information |
| `Card`, `DetailList`, `MetricBand`, `RecordSummary`, `StatCard` | Structured explanations and compact facts |
| `Collapsible`, `Tabs`, `DiffViewer`, `JSONViewer` | Optional detail, alternatives, changes, and structured output |
| `Divider`, `EmptyState` | Content separation and no-content states |

### Layout and composition

| Components | Purpose |
| --- | --- |
| `Container`, `Stack`, `Cluster`, `Grid`, `Center`, `Box`, `Sticky`, `Responsive` | Responsive page composition without page-specific flex hacks |
| `PaneHost`, `Workbench`, `Toolbar` | Dense reference views, inspectors, and grouped actions |
| `AspectRatioComponent` (`AspectRatio`), `Spacer` | Stable media geometry and intentional whitespace |
| `Themed` | Scoped theme override for a component subtree |

### Forms and controls

| Components | Purpose |
| --- | --- |
| `AuthCard`, `SignOut` | Authentication shells and sign-out action |
| `Form`, `FormSection`, `FormField`, `ValidationSummary` | Form structure, grouping, and validation feedback |
| `TextField`, `TextArea`, `PasswordInput`, `NumberField`, `NumberInput`, `DateField`, `TimePicker`, `ColorField`, `ColorPicker` | Typed native inputs |
| `Select`, `Checkbox`, `CheckboxGroup`, `Radio`, `RadioGroup`, `Switch`, `SignalToggle` | Selection and boolean controls |
| `Slider`, `RangeSlider`, `RatingInput`, `SegmentedControl` | Bounded values, ratings, and mutually exclusive choices |
| `InputGroup`, `TagInput`, `Tag`, `FileUpload`, `FileDropzone` | Compound fields, chips, and file input surfaces |
| `ConditionalField`, `ConditionalFieldVisible`, `FormRepeater`, `Repeater`, `StepWizard` | Conditional, repeating, and multi-step authoring flows |

### Data, actions, and feedback

| Components | Purpose |
| --- | --- |
| `DataTable`, `FilterToolbar`, `FilterChipBar` | Sortable data and URL-driven filtering |
| `Button`, `Link`, `LinkButton`, `ConfirmAction`, `ToggleAction`, `OptimisticAction` | Safe actions and navigation affordances |
| `Notification`, `NotificationBell`, `NetworkRetryBanner`, `PollingIndicator` | Feedback, unread state, retry, and active polling |
| `Spinner`, `SkeletonRow`, `SkeletonCard`, `SkeletonAvatar`, `ProgressSteps` | Loading and progress states |
| `Counter`, `AnimatedCounter` | Signal-driven and scroll-triggered values |

### Data visualization and media

| Components | Purpose |
| --- | --- |
| `LineChart`, `BarChart`, `PieChart`, `Sparkline` | SVG charts and compact trends |
| `Gallery`, `Lightbox`, `Carousel`, `Avatar`, `AvatarGroup` | Image, identity, and browsing surfaces |
| `OptimizedImage`, `PipelineImage` | Responsive images and framework image variants |
| `PricingCard`, `Timeline` | Plans, milestones, and chronological records |

The inventory is intentionally a reference, not a second component system.
GoFastr owns rendering, accessibility behavior, focus management, widget
lifecycle, and theme tokens. fastr-docs supplies route data and content
composition around those components.

## Markdown content blocks

Markdown pages can use the native renderer and its framed code blocks. A
project can expose any of the catalogued components through a typed shortcode
adapter:

Use a Markdown component shortcode when a project needs a reusable content block that is not part of the default vocabulary:

```go
router.Use(docs.MarkdownComponentsPlugin{Components: map[string]docs.MarkdownComponent{
    "note": func(props map[string]string, body render.HTML) render.HTML {
        return ui.Callout(ui.CalloutConfig{
            Title: props["title"], Variant: ui.StatusInfo,
        }, body)
    },
}})
```

```md
{{< note title="Important" >}}
This body remains Markdown and can contain links and emphasis.
{{< /note >}}
```

## Interactive screens

Typed screens use the same native controls for application behavior:

- `Form`, `FormField`, `TextField`, `TextArea`, `Select`, `Checkbox`, `Radio`, and `Switch`
- `FilterToolbar`, `FilterChipBar`, `SegmentedControl`, and `SearchInput`
- `DataTable`, `JSONViewer`, `LineChart`, `BarChart`, `PieChart`, and `Sparkline`
- `Notification`, `Banner`, `Spinner`, `Skeleton`, `ProgressSteps`, and `PollingIndicator`
- `Carousel`, `PricingCard`, `StatCard`, `Timeline`, and `RecordSummary`

Register the screen in the same Router:

```go
router.MustScreen("/examples/playground", docs.ScreenConfig{
    Title:       "Playground",
    Description: "Exercise the component library.",
    Component:   &PlaygroundScreen{},
    Order:       1,
})
```

The route remains searchable, appears in the right section navigation, receives breadcrumbs and previous/next links, and participates in static export and PWA precaching.

## Layout primitives

Use `Container`, `Stack`, `Cluster`, `Grid`, `Center`, `Box`, `Sticky`, and `Responsive` to compose a screen. Prefer the component's own layout contract before adding page-specific flex rules. `PaneHost`, `Workbench`, and `Toolbar` cover denser tool and reference views.

## Framework contracts

The UI library is only one layer of the product. A complete docs site also depends on the Router contracts:

- One route tree for pages, screens, groups, plugins, search, and export
- Explicit sibling order and strict validation
- Front matter for titles, descriptions, drafts, tags, versions, locales, redirects, and SEO metadata
- Local JSON search in development and Pagefind for chunked static indexes
- OpenAPI reference pages with an optional server-backed request console
- Light and dark tokens, white-label branding, PWA delivery, and offline content
- Generated and non-CLI project tests that exercise the rendered user flows

See [The router](/docs/concepts/router), [Content authoring](/docs/concepts/content), [OpenAPI reference](/docs/build/openapi), and [Testing](/docs/operate/testing) for the contracts behind these components.

## When the native set is not enough

Keep product-specific behavior in a typed screen or plugin. A plugin can register routes, add Markdown components, validate its own configuration, provide search data, declare browser origins, and mount assets through the GoFastr router. That keeps new capabilities in the same navigation, search, export, and validation pipeline instead of creating a parallel app.
