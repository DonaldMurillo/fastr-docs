package docs

import "strings"

// ThemeCSS returns the template-specific structural layer. GoFastr emits the
// semantic color, typography, radius, and dark-mode variables from Router.Theme;
// this layer applies those variables to the documentation shell and adds the
// small visual signature that distinguishes each template.
func (r *Router) ThemeCSS() string {
	config := r.ThemeConfig()
	return strings.TrimSpace(strings.Join([]string{
		themeTokenBridge,
		templateCSS(config.Template),
		config.CustomCSS,
	}, "\n"))
}

const themeTokenBridge = `
:is(.layout-docs, .layout-blog) {
  --docs-canvas: var(--color-background, #f7f5ef);
  --docs-paper: var(--color-surface, #fffdfa);
  --docs-paper-2: var(--color-surface-soft, #f1f0e9);
  --docs-ink: var(--color-text, #1c1f1d);
  --docs-ink-soft: var(--color-text-muted, #3e4540);
  --docs-muted: var(--color-text-muted, #7c827a);
  --docs-faint: var(--color-text-subtle, #a7aca4);
  --docs-line: var(--color-border, #e1e2db);
  --docs-line-strong: var(--color-border-strong, #c9cdc3);
  --docs-orange: var(--color-accent, #ec7131);
  --docs-orange-deep: var(--color-primary, #c9501e);
  --docs-orange-wash: color-mix(in srgb, var(--docs-orange) 12%, transparent);
  --docs-sage: var(--color-success, #6c9873);
  --docs-sage-wash: color-mix(in srgb, var(--docs-sage) 13%, transparent);
  --docs-blue: var(--color-info, #687bc3);
  --docs-code: var(--color-code-surface, #172019);
  --docs-code-text: var(--color-code-text, #e7eee4);
  --docs-shadow: 0 22px 60px color-mix(in srgb, var(--docs-ink) 9%, transparent);
  --docs-radius-sm: var(--radius-sm, 6px);
  --docs-radius-md: var(--radius-md, 8px);
  --docs-radius-lg: var(--radius-lg, 10px);
  --fastr-docs-template: editorial;
  background: radial-gradient(circle at 50% -10%, color-mix(in srgb, var(--docs-paper) 65%, transparent), transparent 44rem);
}
`

func templateCSS(template Template) string {
	switch ParseTemplate(string(template)) {
	case TemplateTerminal:
		return terminalTemplateCSS
	case TemplateBlueprint:
		return blueprintTemplateCSS
	case TemplateStudio:
		return studioTemplateCSS
	case TemplateNotebook:
		return notebookTemplateCSS
	default:
		return editorialTemplateCSS
	}
}

const editorialTemplateCSS = `
:is(.layout-docs, .layout-blog) { --fastr-docs-template: editorial; }
`

const terminalTemplateCSS = `
:is(.layout-docs, .layout-blog) {
  --fastr-docs-template: terminal;
  --docs-shadow: 0 12px 30px color-mix(in srgb, var(--docs-ink) 12%, transparent);
  background-image: linear-gradient(color-mix(in srgb, var(--docs-line) 28%, transparent) 1px, transparent 1px), linear-gradient(90deg, color-mix(in srgb, var(--docs-line) 28%, transparent) 1px, transparent 1px);
  background-size: 28px 28px;
}
:is(.layout-docs, .layout-blog) > header { border-bottom-style: dashed; }
:is(.layout-docs, .layout-blog) .fastr-docs-brand__mark { border-radius: 2px; border-style: dashed; transform: none; }
:is(.layout-docs, .layout-blog) .fastr-docs-brand__name,
:is(.layout-docs, .layout-blog) .ui-markdown h1,
:is(.layout-docs, .layout-blog) .ui-markdown h2,
:is(.layout-docs, .layout-blog) .ui-markdown h3 { letter-spacing: -.02em; }
:is(.layout-docs, .layout-blog) .ui-site-header__links a,
:is(.layout-docs, .layout-blog) .fastr-docs-command-trigger,
:is(.layout-docs, .layout-blog) .ui-sidebar__link,
:is(.layout-docs, .layout-blog) .ui-doc-layout__prev,
:is(.layout-docs, .layout-blog) .ui-doc-layout__next { border-radius: var(--docs-radius-sm); }
:is(.layout-docs, .layout-blog) .ui-sidebar__title,
:is(.layout-docs, .layout-blog) .fastr-docs-blog__eyebrow { letter-spacing: .16em; }
`

const blueprintTemplateCSS = `
:is(.layout-docs, .layout-blog) {
  --fastr-docs-template: blueprint;
  background-image: linear-gradient(color-mix(in srgb, var(--docs-line) 30%, transparent) 1px, transparent 1px), linear-gradient(90deg, color-mix(in srgb, var(--docs-line) 30%, transparent) 1px, transparent 1px);
  background-size: 32px 32px;
}
:is(.layout-docs, .layout-blog) > header { border-bottom-width: 2px; }
:is(.layout-docs, .layout-blog) .fastr-docs-brand__mark { border-radius: 1px; transform: none; }
:is(.layout-docs, .layout-blog) .ui-site-header__links a,
:is(.layout-docs, .layout-blog) .fastr-docs-command-trigger,
:is(.layout-docs, .layout-blog) .ui-sidebar__link,
:is(.layout-docs, .layout-blog) .ui-doc-layout__prev,
:is(.layout-docs, .layout-blog) .ui-doc-layout__next { border-radius: 0; }
:is(.layout-docs, .layout-blog) .ui-markdown h1,
:is(.layout-docs, .layout-blog) .ui-markdown h2,
:is(.layout-docs, .layout-blog) .ui-markdown h3 { letter-spacing: -.025em; }
:is(.layout-docs, .layout-blog) .ui-sidebar__title::before { width: 22px; }
`

const studioTemplateCSS = `
:is(.layout-docs, .layout-blog) {
  --fastr-docs-template: studio;
  --docs-shadow: 0 25px 70px color-mix(in srgb, var(--docs-orange-deep) 12%, transparent);
  background: radial-gradient(circle at 10% 4%, color-mix(in srgb, var(--docs-orange) 10%, transparent), transparent 28rem), radial-gradient(circle at 92% 24%, color-mix(in srgb, var(--docs-orange-deep) 8%, transparent), transparent 34rem), var(--docs-canvas);
}
:is(.layout-docs, .layout-blog) .fastr-docs-brand__mark { border-radius: var(--docs-radius-lg); transform: rotate(-5deg); }
:is(.layout-docs, .layout-blog) .ui-site-header__links a,
:is(.layout-docs, .layout-blog) .fastr-docs-command-trigger,
:is(.layout-docs, .layout-blog) .ui-sidebar__link,
:is(.layout-docs, .layout-blog) .ui-doc-layout__prev,
:is(.layout-docs, .layout-blog) .ui-doc-layout__next,
:is(.layout-docs, .layout-blog) [data-fui-comp="ui-card"] { border-radius: var(--docs-radius-md); }
:is(.layout-docs, .layout-blog) .ui-markdown h1,
:is(.layout-docs, .layout-blog) .ui-markdown h2,
:is(.layout-docs, .layout-blog) .ui-markdown h3 { letter-spacing: -.055em; }
:is(.layout-docs, .layout-blog) .ui-sidebar__title::before { border-radius: 999px; }
`

const notebookTemplateCSS = `
:is(.layout-docs, .layout-blog) {
  --fastr-docs-template: notebook;
  background-image: linear-gradient(color-mix(in srgb, var(--docs-line) 24%, transparent) 1px, transparent 1px);
  background-size: 100% 32px;
}
:is(.layout-docs, .layout-blog) .fastr-docs-brand__mark { border-radius: 2px; transform: rotate(-3deg); }
:is(.layout-docs, .layout-blog) .ui-site-header__links a,
:is(.layout-docs, .layout-blog) .fastr-docs-command-trigger,
:is(.layout-docs, .layout-blog) .ui-sidebar__link,
:is(.layout-docs, .layout-blog) .ui-doc-layout__prev,
:is(.layout-docs, .layout-blog) .ui-doc-layout__next { border-radius: var(--docs-radius-sm); }
:is(.layout-docs, .layout-blog) .ui-markdown h1,
:is(.layout-docs, .layout-blog) .ui-markdown h2,
:is(.layout-docs, .layout-blog) .ui-markdown h3 { letter-spacing: -.035em; }
:is(.layout-docs, .layout-blog) .ui-markdown blockquote { border-left-width: 3px; }
:is(.layout-docs, .layout-blog) .ui-sidebar__title::before { width: 20px; }
`
