package docs

import (
	"fmt"
	"math"
	"strconv"
	"strings"

	"github.com/DonaldMurillo/gofastr/core-ui/style"
	uitheme "github.com/DonaldMurillo/gofastr/framework/ui/theme"
)

// Template selects the visual starting point for a documentation site.
// Templates are build-time choices: they change the visual system while the
// Router, route tree, content, search, PWA, and plugin behavior stay the same.
type Template string

const (
	// TemplateEditorial is the default warm, compact documentation theme.
	TemplateEditorial Template = "editorial"
	// TemplateTerminal is a monospace, high-contrast command-line theme.
	TemplateTerminal Template = "terminal"
	// TemplateBlueprint is a cool, technical theme with grid details.
	TemplateBlueprint Template = "blueprint"
	// TemplateStudio is a softer, expressive product-studio theme.
	TemplateStudio Template = "studio"
	// TemplateNotebook is a reading-first paper and serif theme.
	TemplateNotebook Template = "notebook"
)

// ThemeOverrides is the GoFastr semantic token set exposed through the
// fastr-docs wrapper. It includes light and dark color tokens, font families,
// and the three shared radii. Empty fields retain the selected template value.
type ThemeOverrides = uitheme.Overrides

// ThemeConfig selects a template and optionally overrides its semantic tokens.
// CustomCSS is appended after the fastr-docs visual layer, so it is suitable
// for project-specific components and small layout adjustments. Keep it in
// source control and treat it as trusted CSS.
// Variables returns the theme's CSS custom properties as data: the
// template's baseline with this config's overrides applied. Tooling can
// diff or preview a theme without parsing CSS.
func (c ThemeConfig) Variables() map[string]string {
	vars := map[string]string{
		// The template's own name travels with the variables, so tooling
		// and agents can say which theme a page wears without guessing
		// from colors.
		"template": string(ParseTemplate(string(c.Template))),
	}
	for k, v := range templateVariables(c.Template) {
		vars[k] = v
	}
	for k, v := range overridesVariables(c.Overrides) {
		vars[k] = v
	}
	if strings.TrimSpace(c.CustomCSS) != "" {
		vars["custom-css-bytes"] = strconv.Itoa(len(c.CustomCSS))
	}
	return vars
}

func overridesVariables(o ThemeOverrides) map[string]string {
	vars := map[string]string{}
	set := func(token, value string) {
		if strings.TrimSpace(value) != "" {
			vars[token] = value
		}
	}
	set("color-background", o.Background)
	set("color-surface", o.Surface)
	set("color-surface-soft", o.SurfaceSoft)
	set("color-border", o.Border)
	set("color-border-strong", o.BorderStrong)
	set("color-text", o.Text)
	set("color-text-muted", o.TextMuted)
	set("color-text-subtle", o.TextSubtle)
	set("color-primary", o.Primary)
	set("color-primary-fg", o.PrimaryFg)
	set("color-accent", o.Accent)
	set("color-success", o.Success)
	set("color-warning", o.Warning)
	set("color-danger", o.Danger)
	set("color-info", o.Info)
	set("font-body", o.FontBody)
	set("font-heading", o.FontHeading)
	set("font-mono", o.FontMono)
	for k, v := range o.DarkColors {
		vars["dark-"+k] = v
	}
	return vars
}

type ThemeConfig struct {
	// Template is the starting point; one of ThemeTemplates.
	Template Template
	// Overrides replace individual theme variables.
	Overrides ThemeOverrides
	// CustomCSS is appended after the theme, for brand rules.
	CustomCSS string
}

// ThemeTemplates returns the five supported starting points in gallery order.
func ThemeTemplates() []Template {
	return []Template{
		TemplateEditorial,
		TemplateTerminal,
		TemplateBlueprint,
		TemplateStudio,
		TemplateNotebook,
	}
}

// ParseTemplate normalizes a template name from configuration or an
// environment variable. Unknown or empty values intentionally fall back to
// the safe editorial default.
// ParseTemplate folds case and whitespace; an unknown name still resolves
// to the editorial default, and Warnings says so.
func ParseTemplate(value string) Template {
	value = strings.ToLower(strings.TrimSpace(value))
	switch strings.ToLower(strings.TrimSpace(value)) {
	case string(TemplateTerminal):
		return TemplateTerminal
	case string(TemplateBlueprint):
		return TemplateBlueprint
	case string(TemplateStudio):
		return TemplateStudio
	case string(TemplateNotebook):
		return TemplateNotebook
	default:
		return TemplateEditorial
	}
}

// TemplateDescription gives configuration UIs and generated docs a concise
// explanation of each visual starting point.
func TemplateDescription(template Template) string {
	switch ParseTemplate(string(template)) {
	case TemplateTerminal:
		return "Monospace, high-contrast, and command-line inspired."
	case TemplateBlueprint:
		return "Technical, cool-toned, and structured around a blueprint grid."
	case TemplateStudio:
		return "Expressive, softly rounded, and suited to product documentation."
	case TemplateNotebook:
		return "Reading-first, paper-toned, and led by serif typography."
	default:
		return "Warm, compact, and editorial with a focused reading rhythm."
	}
}

// WithTemplate selects one of the built-in visual templates. It preserves any
// token overrides supplied by an earlier WithTheme option, which makes option
// order safe for composition.
func WithTemplate(template Template) Option {
	return func(r *Router) {
		if r == nil {
			return
		}
		raw := strings.TrimSpace(string(template))
		parsed := ParseTemplate(raw)
		if raw != "" && parsed != Template(raw) {
			// Keep the typo for Warnings to name; the rendered theme is
			// still the safe default.
			r.templateRaw = raw
		}
		r.themeConfig.Template = parsed
	}
}

// WithTheme selects a template and/or applies semantic token overrides. A
// zero Template keeps the current template, allowing callers to write only
// the custom values they need:
//
//	docs.WithTheme(docs.ThemeConfig{
//		Template: docs.TemplateBlueprint,
//		Overrides: docs.ThemeOverrides{Primary: "#155E9A"},
//	})
func WithTheme(config ThemeConfig) Option {
	return func(r *Router) {
		if r == nil {
			return
		}
		if config.Template != "" {
			r.themeConfig.Template = ParseTemplate(string(config.Template))
		}
		r.themeConfig.Overrides = mergeThemeOverrides(r.themeConfig.Overrides, config.Overrides)
		if strings.TrimSpace(config.CustomCSS) != "" {
			r.themeConfig.CustomCSS = strings.TrimSpace(config.CustomCSS)
		}
	}
}

// Template returns the active visual template.
func (r *Router) Template() Template {
	if r == nil {
		return TemplateEditorial
	}
	return ParseTemplate(string(r.themeConfig.Template))
}

// ThemeConfig returns a copy of the active theme configuration.
func (r *Router) ThemeConfig() ThemeConfig {
	if r == nil {
		return ThemeConfig{Template: TemplateEditorial}
	}
	config := r.themeConfig
	config.Template = r.Template()
	config.Overrides.DarkColors = cloneStringMap(config.Overrides.DarkColors)
	return config
}

// Theme returns the complete adaptive GoFastr theme for the active template.
// Hosts should pass this directly to uiapp.NewApp(...).WithTheme(...).
func (r *Router) Theme() style.Theme {
	if r == nil {
		return themeForTemplate(TemplateEditorial, ThemeOverrides{})
	}
	return themeForTemplate(r.Template(), r.themeConfig.Overrides)
}

// ThemeColor returns the browser chrome color for the active theme. A brand
// override wins; otherwise the light background token is used as the stable
// metadata value for both color schemes.
func (r *Router) ThemeColor() string {
	if r != nil && strings.TrimSpace(r.brand.ThemeColor) != "" {
		return strings.TrimSpace(r.brand.ThemeColor)
	}
	return r.Theme().Colors.Background.Value
}

// DefaultTheme is kept as a convenient compatibility entry point for projects
// that do not need template selection. It is the same Editorial theme used by
// a new Router.
func DefaultTheme() style.Theme {
	return themeForTemplate(TemplateEditorial, ThemeOverrides{})
}

func themeForTemplate(template Template, overrides ThemeOverrides) style.Theme {
	return uitheme.Default(templateOverrides(template), overrides)
}

func templateVariables(template Template) map[string]string {
	return overridesVariables(templateOverrides(template))
}

func templateOverrides(template Template) ThemeOverrides {
	switch ParseTemplate(string(template)) {
	case TemplateTerminal:
		return terminalThemeOverrides()
	case TemplateBlueprint:
		return blueprintThemeOverrides()
	case TemplateStudio:
		return studioThemeOverrides()
	case TemplateNotebook:
		return notebookThemeOverrides()
	default:
		return editorialThemeOverrides()
	}
}

func editorialThemeOverrides() ThemeOverrides {
	return ThemeOverrides{
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
	}
}

func terminalThemeOverrides() ThemeOverrides {
	return ThemeOverrides{
		Background:   "#f4f7f2",
		Surface:      "#fbfdf9",
		SurfaceSoft:  "#e7eee8",
		Border:       "#c5d2c6",
		BorderStrong: "#839887",
		Text:         "#172019",
		TextMuted:    "#526457",
		TextSubtle:   "#7f9182",
		Primary:      "#217951",
		PrimaryFg:    "#ffffff",
		Accent:       "#b6671f",
		Success:      "#26794f",
		Warning:      "#a3611a",
		Danger:       "#b84436",
		Info:         "#2f6798",
		CodeSurface:  "#101812",
		CodeText:     "#dcebdd",
		CodeBorder:   "#2f4736",
		FontBody:     "\"IBM Plex Mono\", \"SFMono-Regular\", Consolas, monospace",
		FontHeading:  "\"IBM Plex Mono\", \"SFMono-Regular\", Consolas, monospace",
		FontMono:     "\"IBM Plex Mono\", \"SFMono-Regular\", Consolas, monospace",
		RadiusSm:     2,
		RadiusMd:     3,
		RadiusLg:     4,
		DarkColors: map[string]string{
			"background":    "#0b110e",
			"surface":       "#111a14",
			"surface-soft":  "#17231b",
			"border":        "#294032",
			"border-strong": "#48634e",
			"text":          "#e8f3e8",
			"text-muted":    "#9bb09d",
			"text-subtle":   "#708575",
			"primary":       "#83e1a8",
			"primary-fg":    "#0b110e",
			"accent":        "#ffbd70",
			"success":       "#83e1a8",
			"warning":       "#f0c477",
			"danger":        "#ff9184",
			"info":          "#8ac8f0",
			"code-surface":  "#070c09",
			"code-text":     "#d9f2dc",
			"code-border":   "#36513d",
		},
	}
}

func blueprintThemeOverrides() ThemeOverrides {
	return ThemeOverrides{
		Background:   "#f4f8fb",
		Surface:      "#ffffff",
		SurfaceSoft:  "#e8f0f5",
		Border:       "#c5d6e1",
		BorderStrong: "#91aabb",
		Text:         "#142b3b",
		TextMuted:    "#536a79",
		TextSubtle:   "#7f96a5",
		Primary:      "#176aa8",
		PrimaryFg:    "#ffffff",
		Accent:       "#20a5c2",
		Success:      "#277f6d",
		Warning:      "#ae6e1e",
		Danger:       "#b94c4c",
		Info:         "#4078c4",
		CodeSurface:  "#14242f",
		CodeText:     "#e0f0f7",
		CodeBorder:   "#365363",
		FontBody:     "\"IBM Plex Sans\", Inter, ui-sans-serif, system-ui, sans-serif",
		FontHeading:  "\"IBM Plex Sans\", Inter, ui-sans-serif, system-ui, sans-serif",
		FontMono:     "\"IBM Plex Mono\", \"SFMono-Regular\", Consolas, monospace",
		RadiusSm:     1,
		RadiusMd:     2,
		RadiusLg:     3,
		DarkColors: map[string]string{
			"background":    "#0c1720",
			"surface":       "#13232d",
			"surface-soft":  "#1b303c",
			"border":        "#294352",
			"border-strong": "#486675",
			"text":          "#edf7fb",
			"text-muted":    "#a2b8c5",
			"text-subtle":   "#708997",
			"primary":       "#75c9f1",
			"primary-fg":    "#0c1720",
			"accent":        "#58d5df",
			"success":       "#72c9ad",
			"warning":       "#efc477",
			"danger":        "#ff9b91",
			"info":          "#91b8f2",
			"code-surface":  "#081118",
			"code-text":     "#dff5ff",
			"code-border":   "#365462",
		},
	}
}

func studioThemeOverrides() ThemeOverrides {
	return ThemeOverrides{
		Background:   "#fbf8f7",
		Surface:      "#ffffff",
		SurfaceSoft:  "#f2ebee",
		Border:       "#e5d8df",
		BorderStrong: "#c6aebb",
		Text:         "#291c25",
		TextMuted:    "#6e5e67",
		TextSubtle:   "#a08f99",
		Primary:      "#8950d8",
		PrimaryFg:    "#ffffff",
		Accent:       "#ed6a5d",
		Success:      "#3e8b74",
		Warning:      "#c48431",
		Danger:       "#c94c57",
		Info:         "#557dcc",
		CodeSurface:  "#251b2c",
		CodeText:     "#f2e8f6",
		CodeBorder:   "#49374f",
		FontBody:     "Inter, ui-sans-serif, system-ui, sans-serif",
		FontHeading:  "\"Space Grotesk\", Inter, ui-sans-serif, system-ui, sans-serif",
		FontMono:     "\"SFMono-Regular\", Consolas, monospace",
		RadiusSm:     10,
		RadiusMd:     15,
		RadiusLg:     22,
		DarkColors: map[string]string{
			"background":    "#1a151d",
			"surface":       "#241d27",
			"surface-soft":  "#302632",
			"border":        "#493b4d",
			"border-strong": "#68566d",
			"text":          "#f7eef7",
			"text-muted":    "#beaaba",
			"text-subtle":   "#8e788f",
			"primary":       "#c9a2ff",
			"primary-fg":    "#1a151d",
			"accent":        "#ff9b8f",
			"success":       "#7ec8aa",
			"warning":       "#efc17a",
			"danger":        "#ff928d",
			"info":          "#9bb7f4",
			"code-surface":  "#140f18",
			"code-text":     "#f4e9fa",
			"code-border":   "#4c3a55",
		},
	}
}

func notebookThemeOverrides() ThemeOverrides {
	return ThemeOverrides{
		Background:   "#f6f0e4",
		Surface:      "#fffaf0",
		SurfaceSoft:  "#efe5d4",
		Border:       "#ded0b9",
		BorderStrong: "#b6a48a",
		Text:         "#2c2a26",
		TextMuted:    "#6e665a",
		TextSubtle:   "#998f80",
		Primary:      "#8e4f2e",
		PrimaryFg:    "#fffaf0",
		Accent:       "#c87944",
		Success:      "#4f8065",
		Warning:      "#b8792e",
		Danger:       "#b64e42",
		Info:         "#5574a2",
		CodeSurface:  "#2b2823",
		CodeText:     "#f4ecdc",
		CodeBorder:   "#51483b",
		FontBody:     "\"Iowan Old Style\", \"Palatino Linotype\", Palatino, Georgia, serif",
		FontHeading:  "\"Fraunces\", \"Iowan Old Style\", Georgia, serif",
		FontMono:     "\"SFMono-Regular\", Consolas, monospace",
		RadiusSm:     2,
		RadiusMd:     3,
		RadiusLg:     4,
		DarkColors: map[string]string{
			"background":    "#1b1916",
			"surface":       "#25221e",
			"surface-soft":  "#302b25",
			"border":        "#4b4337",
			"border-strong": "#675947",
			"text":          "#f6efdf",
			"text-muted":    "#c6bba9",
			"text-subtle":   "#958a78",
			"primary":       "#efb086",
			"primary-fg":    "#1b1916",
			"accent":        "#e7955e",
			"success":       "#91c69f",
			"warning":       "#e4bc72",
			"danger":        "#ff9e8f",
			"info":          "#9eb8e9",
			"code-surface":  "#12100e",
			"code-text":     "#f5eddd",
			"code-border":   "#554839",
		},
	}
}

func mergeThemeOverrides(base, override ThemeOverrides) ThemeOverrides {
	merged := base
	if override.Background != "" {
		merged.Background = override.Background
	}
	if override.Surface != "" {
		merged.Surface = override.Surface
	}
	if override.SurfaceSoft != "" {
		merged.SurfaceSoft = override.SurfaceSoft
	}
	if override.Border != "" {
		merged.Border = override.Border
	}
	if override.BorderStrong != "" {
		merged.BorderStrong = override.BorderStrong
	}
	if override.Text != "" {
		merged.Text = override.Text
	}
	if override.TextMuted != "" {
		merged.TextMuted = override.TextMuted
	}
	if override.TextSubtle != "" {
		merged.TextSubtle = override.TextSubtle
	}
	if override.Primary != "" {
		merged.Primary = override.Primary
	}
	if override.PrimaryFg != "" {
		merged.PrimaryFg = override.PrimaryFg
	}
	if override.Accent != "" {
		merged.Accent = override.Accent
	}
	if override.Success != "" {
		merged.Success = override.Success
	}
	if override.Warning != "" {
		merged.Warning = override.Warning
	}
	if override.Danger != "" {
		merged.Danger = override.Danger
	}
	if override.Info != "" {
		merged.Info = override.Info
	}
	if override.CodeSurface != "" {
		merged.CodeSurface = override.CodeSurface
	}
	if override.CodeText != "" {
		merged.CodeText = override.CodeText
	}
	if override.CodeBorder != "" {
		merged.CodeBorder = override.CodeBorder
	}
	if override.FontBody != "" {
		merged.FontBody = override.FontBody
	}
	if override.FontHeading != "" {
		merged.FontHeading = override.FontHeading
	}
	if override.FontMono != "" {
		merged.FontMono = override.FontMono
	}
	if override.RadiusSm > 0 {
		merged.RadiusSm = override.RadiusSm
	}
	if override.RadiusMd > 0 {
		merged.RadiusMd = override.RadiusMd
	}
	if override.RadiusLg > 0 {
		merged.RadiusLg = override.RadiusLg
	}
	if len(override.DarkColors) > 0 {
		merged.DarkColors = cloneStringMap(merged.DarkColors)
		if merged.DarkColors == nil {
			merged.DarkColors = make(map[string]string, len(override.DarkColors))
		}
		for key, value := range override.DarkColors {
			if strings.TrimSpace(value) != "" {
				merged.DarkColors[key] = value
			}
		}
	}
	return merged
}

// templateNames lists the available templates for warnings.
func templateNames() []string {
	names := make([]string, 0, len(ThemeTemplates()))
	for _, template := range ThemeTemplates() {
		names = append(names, string(template))
	}
	return names
}

// lowContrastOverrides reports "text on background" when the configured
// override pair fails WCAG AA, or "" when it passes or is unmeasurable.
func (r *Router) lowContrastOverrides() string {
	if r == nil {
		return ""
	}
	config := r.ThemeConfig()
	vars := config.Variables()
	text, hasText := vars["color-text"]
	background, hasBackground := vars["color-background"]
	if !hasText || !hasBackground {
		return ""
	}
	ratio, ok := contrastRatio(text, background)
	if !ok || ratio >= 4.5 {
		return ""
	}
	return fmt.Sprintf("text %q on background %q (ratio %.2f:1)", text, background, ratio)
}

// contrastRatio computes the WCAG relative-luminance ratio of two hex
// colors. Non-hex values report ok=false rather than a wrong number.
func contrastRatio(fg, bg string) (float64, bool) {
	f, ok1 := hexLuminance(fg)
	g, ok2 := hexLuminance(bg)
	if !ok1 || !ok2 {
		return 0, false
	}
	lighter, darker := f, g
	if darker > lighter {
		lighter, darker = darker, lighter
	}
	return (lighter + 0.05) / (darker + 0.05), true
}

func hexLuminance(hex string) (float64, bool) {
	value := strings.TrimPrefix(strings.TrimSpace(hex), "#")
	if len(value) == 3 {
		expanded := make([]byte, 0, 6)
		for i := range value {
			expanded = append(expanded, value[i], value[i])
		}
		value = string(expanded)
	}
	if len(value) != 6 {
		return 0, false
	}
	var channels [3]float64
	for i := range 3 {
		parsed := 0
		for _, c := range []byte(value[i*2 : i*2+2]) {
			digit, ok := hexDigit(c)
			if !ok {
				return 0, false
			}
			parsed = parsed*16 + digit
		}
		linear := float64(parsed) / 255.0
		if linear <= 0.04045 {
			linear /= 12.92
		} else {
			linear = math.Pow((linear+0.055)/1.055, 2.4)
		}
		channels[i] = linear
	}
	return 0.2126*channels[0] + 0.7152*channels[1] + 0.0722*channels[2], true
}

func hexDigit(c byte) (int, bool) {
	switch {
	case c >= '0' && c <= '9':
		return int(c - '0'), true
	case c >= 'a' && c <= 'f':
		return int(c-'a') + 10, true
	case c >= 'A' && c <= 'F':
		return int(c-'A') + 10, true
	}
	return 0, false
}
