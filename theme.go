package docs

import (
	"github.com/DonaldMurillo/gofastr/core-ui/style"
	uitheme "github.com/DonaldMurillo/gofastr/framework/ui/theme"
)

// DefaultTheme is the white-label visual foundation for generated docs sites.
// It keeps GoFastr's semantic tokens as the source of truth while matching the
// fastr-docs editorial palette in both color schemes.
func DefaultTheme() style.Theme {
	return uitheme.Default(uitheme.Overrides{
		Background:   "#f7f5ef",
		Surface:      "#fffdfa",
		SurfaceSoft:  "#f1f0e9",
		Border:       "#e1e2db",
		BorderStrong: "#c9cdc3",
		Text:         "#1c1f1d",
		TextMuted:    "#7c827a",
		TextSubtle:   "#a7aca4",
		Primary:      "#c9501e",
		PrimaryFg:    "#fffdfa",
		Accent:       "#ec7131",
		Success:      "#6c9873",
		Warning:      "#d2ae5c",
		Info:         "#687bc3",
		CodeSurface:  "#172019",
		CodeText:     "#e7eee4",
		CodeBorder:   "#2d392f",
		FontBody:     "Inter, ui-sans-serif, system-ui, -apple-system, BlinkMacSystemFont, \"Segoe UI\", sans-serif",
		FontHeading:  "\"SF Pro Display\", \"Avenir Next\", Inter, ui-sans-serif, system-ui, sans-serif",
		FontMono:     "\"SFMono-Regular\", Consolas, \"Liberation Mono\", monospace",
		RadiusSm:     6,
		RadiusMd:     8,
		RadiusLg:     10,
		DarkColors: map[string]string{
			"background":    "#131917",
			"surface":       "#1b211e",
			"surface-soft":  "#222923",
			"border":        "#303931",
			"border-strong": "#455147",
			"text":          "#edf1eb",
			"text-muted":    "#929c93",
			"text-subtle":   "#68736a",
			"primary":       "#ffab74",
			"primary-fg":    "#131917",
			"accent":        "#f28a4b",
			"success":       "#84bd8d",
			"warning":       "#e0c074",
			"info":          "#93a3ec",
			"code-surface":  "#0e1410",
			"code-text":     "#e5eee2",
			"code-border":   "#455147",
		},
	})
}
