package docs

import (
	"github.com/DonaldMurillo/gofastr/core/render"
	"github.com/DonaldMurillo/gofastr/framework/ui"
)

// playgroundScreen is a typed surface that lives in the same route tree as
// Markdown. It gives the fastr-docs site a concrete place to demonstrate and
// exercise GoFastr components.
type playgroundScreen struct{}

func (*playgroundScreen) Render() render.HTML {
	return render.Tag("div", map[string]string{
		"class":       "fastr-docs-example-screen",
		"data-testid": "examples-playground",
	},
		ui.PageHeader(ui.PageHeaderConfig{
			Eyebrow:  "Typed screen",
			Title:    "Playground",
			Subtitle: "A real GoFastr component surface registered beside the fastr-docs handbook.",
		}),
		render.Tag("div", map[string]string{"class": "fastr-docs-example-screen__stats"},
			ui.StatCard(ui.StatCardConfig{Label: "Source of truth", Value: "Router", Trend: "One route tree", Direction: ui.TrendFlat}),
			ui.StatCard(ui.StatCardConfig{Label: "Content", Value: "Pages + screens", Trend: "Same navigation", Direction: ui.TrendUp}),
			ui.StatCard(ui.StatCardConfig{Label: "Delivery", Value: "PWA ready", Trend: "Static content works offline", Direction: ui.TrendUp}),
		),
		render.Tag("div", map[string]string{"class": "fastr-docs-example-screen__grid"},
			ui.Card(ui.CardConfig{Heading: "Compose a surface", HeadingLevel: 2, Description: "Framework primitives can be assembled without leaving the docs router."},
				ui.Tabs(ui.TabsConfig{SignalName: "docs-playground-tabs", Tabs: []ui.TabItem{
					{Label: "Overview", Content: render.Tag("p", nil, render.Text("Tabs, disclosures, and cards keep rich content discoverable."))},
					{Label: "Implementation", Content: render.Tag("p", nil, render.Text("Move the component into your product package when it becomes a real feature."))},
				}}),
				ui.Collapsible(ui.CollapsibleConfig{Summary: "Show the design rule"}, render.Tag("p", nil, render.Text("Start with a route, give it an explicit order, and let the Router derive navigation and search metadata."))),
			),
			ui.Card(ui.CardConfig{Heading: "Try the controls", HeadingLevel: 2, Description: "These are framework-owned interactions rendered inside a typed screen."},
				ui.Counter(ui.CounterConfig{SignalName: "docs-playground-counter", Step: 1}),
				ui.Switch(ui.ToggleConfig{Name: "playground-notifications", Label: "Enable notifications", Checked: true, ID: "playground-notifications"}),
				ui.SegmentedControl(ui.SegmentedControlConfig{Name: "playground-density", Label: "Density", Selected: "comfortable", Options: []ui.SegmentedOption{
					{Label: "Compact", Value: "compact"}, {Label: "Comfortable", Value: "comfortable"}, {Label: "Spacious", Value: "spacious"},
				}}),
			),
		),
		render.Tag("p", map[string]string{"class": "fastr-docs-example-screen__next"},
			render.Raw(`Read <a href="/docs/build/screens">Screens and components</a> for the registration pattern.`)),
	)
}

// ExamplesCSS is the small amount of layout chrome specific to this typed
// example. Framework components continue to own their own styling.
func ExamplesCSS() string {
	return `.fastr-docs-example-screen { max-width: 1120px; margin: 0 auto; padding: 48px clamp(24px, 5vw, 72px) 96px; color: var(--docs-ink-soft); }
.fastr-docs-example-screen__stats, .fastr-docs-example-screen__grid { display: grid; gap: 14px; }
.fastr-docs-example-screen__stats { grid-template-columns: repeat(3, minmax(0, 1fr)); margin: 28px 0 40px; }
.fastr-docs-example-screen__grid { grid-template-columns: repeat(2, minmax(0, 1fr)); align-items: stretch; }
.fastr-docs-example-screen__grid > [data-fui-comp="ui-card"] { height: 100%; }
.fastr-docs-example-screen__next { margin: 30px 0 0; color: var(--docs-muted); }
.fastr-docs-example-screen__next a { color: var(--docs-orange-deep); }
@media (max-width: 820px) { .fastr-docs-example-screen { padding-inline: 24px; } .fastr-docs-example-screen__stats, .fastr-docs-example-screen__grid { grid-template-columns: 1fr; } }
@media (max-width: 420px) { .fastr-docs-example-screen { padding-inline: 19px; } }`
}
