package docs

// CSS returns the white-label visual layer around GoFastr's UI components.
// The framework still owns structure and interaction; these rules own the
// fastr-docs editorial identity, responsive placement, and semantic tokens.
func (r *Router) CSS() string {
	if r == nil {
		return docsCSS + "\n" + (&Router{}).ThemeCSS()
	}
	return docsCSS + "\n" + r.ThemeCSS()
}

const docsCSS = `
:root { color-scheme: light dark; }
html { scroll-behavior: smooth; }
body { min-width: 0; margin: 0; overflow-x: clip; color: var(--color-text, #1c1f1d); background: var(--color-background, #f7f5ef); font-family: var(--font-body, Inter, ui-sans-serif, system-ui, sans-serif); -webkit-font-smoothing: antialiased; }
button, input, select, textarea { font: inherit; }
code, kbd, pre { font-family: var(--font-mono, ui-monospace, monospace); }

.layout-docs {
  --nav-h: 68px;
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
  --docs-shadow: 0 22px 60px rgb(36 43 36 / .08);
  background: radial-gradient(circle at 50% -10%, color-mix(in srgb, var(--docs-paper) 65%, transparent), transparent 44rem);
  overflow-x: clip;
}

/* Header */
.layout-docs > header { position: sticky; top: 0; z-index: 30; height: 68px; border-bottom: 1px solid var(--docs-line); background: color-mix(in srgb, var(--docs-canvas) 90%, transparent); backdrop-filter: blur(18px); }
.fastr-docs-site-header { gap: 20px; padding-inline: 30px; }
.fastr-docs-site-header .ui-site-header__brand { flex: 0 0 auto; gap: 9px; }
.fastr-docs-brand { display: inline-flex; align-items: center; gap: 9px; color: var(--docs-ink) !important; text-decoration: none; }
.fastr-docs-brand__mark { display: inline-flex; align-items: flex-end; justify-content: center; gap: 2px; width: 27px; height: 27px; padding: 4px; border: 2px solid var(--docs-ink); border-radius: 8px; transform: rotate(-9deg); }
.fastr-docs-brand__mark i { display: block; width: 4px; border-radius: 3px; background: var(--docs-orange); transform: skew(-12deg); }
.fastr-docs-brand__mark i:nth-child(1) { height: 8px; }
.fastr-docs-brand__mark i:nth-child(2) { height: 14px; }
.fastr-docs-brand__mark i:nth-child(3) { height: 19px; }
.fastr-docs-brand__logo { display: block; width: auto; max-width: 150px; max-height: 31px; object-fit: contain; }
.fastr-docs-brand__name { color: var(--docs-ink); font-family: var(--font-heading, Inter, sans-serif); font-size: 18px; font-weight: 700; letter-spacing: -.065em; }
.fastr-docs-site-header .ui-site-header__links { gap: 4px; margin-inline-start: 28px; }
.fastr-docs-site-header .ui-site-header__links a { padding: 8px 11px; border-radius: 6px; color: var(--docs-muted); font-size: 12px; transition: .18s ease; }
.fastr-docs-site-header .ui-site-header__links a:hover, .fastr-docs-site-header .ui-site-header__links a[aria-current="page"] { color: var(--docs-ink); background: var(--docs-paper-2); }
.fastr-docs-site-header .ui-site-header__links a[aria-current="page"] { font-weight: 650; }
.fastr-docs-site-header .ui-site-header__right { gap: 7px; }
.fastr-docs-command-search { position: relative; display: flex; align-items: center; margin-left: auto; }
.fastr-docs-command-trigger { display: inline-flex; align-items: center; gap: 8px; width: 195px; height: 36px; padding: 0 10px; border: 1px solid var(--docs-line); border-radius: 7px; color: var(--docs-muted); background: var(--docs-paper); cursor: pointer; text-align: left; transition: .18s ease; }
.fastr-docs-command-trigger:hover, .fastr-docs-command-trigger:focus-visible { border-color: var(--docs-line-strong); color: var(--docs-ink); box-shadow: 0 7px 25px rgb(30 36 30 / .06); }
.fastr-docs-command-trigger__label { flex: 1; font-size: 11px; }
.fastr-docs-command-trigger__hint { flex: 0 0 auto; color: var(--docs-faint); font-size: 9px; }
.fastr-docs-command-trigger__hint .ui-shortcut-hint__key { min-width: 15px; padding: 1px 4px; border-color: var(--docs-line); border-bottom-width: 1px; border-radius: 4px; color: var(--docs-muted); background: var(--docs-paper-2); font-size: 9px; }
.fastr-docs-command-search > .ui-visually-hidden { position: absolute; }
/* The native GoFastr palette owns search behavior. This shell owns the
   discoverable close affordance and keeps the long route list inside the
   viewport, including the framework's full-screen mobile variant. */
[data-fui-widget="fastr-docs-command-palette"] > .fui-panel { position: relative; display: flex; flex-direction: column; width: min(36rem, 92vw); max-width: 100%; max-height: calc(100vh - 32px); overflow: hidden; border: 1px solid var(--docs-line); border-radius: 12px; background: var(--docs-paper); box-shadow: var(--docs-shadow); }
[data-fui-widget="fastr-docs-command-palette"] > .fui-panel > .fui-slot { min-width: 0; }
.fastr-docs-command-palette__close-slot { position: absolute; z-index: 2; top: 8px; right: 12px; }
.fastr-docs-command-palette__close { display: inline-flex; align-items: center; justify-content: center; flex: 0 0 auto; width: 40px; height: 40px; padding: 0; border: 1px solid var(--docs-line); border-radius: 7px; color: var(--docs-muted); background: transparent; cursor: pointer; }
.fastr-docs-command-palette__close:hover, .fastr-docs-command-palette__close:focus-visible { border-color: var(--docs-line-strong); color: var(--docs-ink); background: var(--docs-paper-2); outline: none; }
[data-fui-widget="fastr-docs-command-palette"] > .fui-panel > .fui-slot-body { min-height: 0; overflow: hidden; }
[data-fui-widget="fastr-docs-command-palette"] .ui-cmd-palette { display: flex; flex-direction: column; width: 100%; min-height: 0; max-height: 100%; box-sizing: border-box; border: 0; border-radius: 0; box-shadow: none; }
[data-fui-widget="fastr-docs-command-palette"] .ui-cmd-palette__combobox { display: flex; flex-direction: column; min-height: 0; }
[data-fui-widget="fastr-docs-command-palette"] .ui-cmd-palette__combobox .combobox__form { padding-right: 64px; }
[data-fui-widget="fastr-docs-command-palette"] .ui-cmd-palette__combobox .combobox__listbox { min-height: 0; overflow-y: auto; }
/* GoFastr's static combobox runtime marks filtered options as hidden. Keep
   the framework's option layout rule from overriding the browser's hidden
   state so the palette shows only matching routes. */
[data-fui-static-options] [role="option"][hidden] { display: none !important; }
.fastr-docs-variant-selectors { display: inline-flex; align-items: center; gap: 5px; }
.fastr-docs-variant-select { display: inline-flex; align-items: center; }
.fastr-docs-variant-select__label { position: absolute; width: 1px; height: 1px; overflow: hidden; clip: rect(0 0 0 0); white-space: nowrap; }
.fastr-docs-variant-select select { min-width: 74px; max-width: 112px; height: 34px; padding: 0 24px 0 9px; border: 1px solid var(--docs-line); border-radius: 7px; color: var(--docs-muted); background: var(--docs-paper); font: 11px var(--font-mono, monospace); }
.fastr-docs-variant-select select:focus-visible { border-color: var(--docs-line-strong); color: var(--docs-ink); outline: 2px solid color-mix(in srgb, var(--docs-orange) 45%, transparent); outline-offset: 2px; }
.fastr-docs-theme-toggle { width: 34px; height: 34px; padding: 0; border: 1px solid transparent; border-radius: 7px; color: var(--docs-muted); background: transparent; }
.fastr-docs-theme-toggle:hover, .fastr-docs-theme-toggle:focus-visible { color: var(--docs-ink); border-color: var(--docs-line); background: var(--docs-paper); }
.fastr-docs-mobile-nav-trigger { display: none; }

/* Outer shell and route tree */
.layout-docs .layout-body { min-height: calc(100vh - 68px); }
.layout-docs .layout-body > nav { flex: 0 0 254px; min-width: 0; border-right: 1px solid var(--docs-line); background: color-mix(in srgb, var(--docs-canvas) 80%, transparent); }
.layout-docs .ui-sidebar__inline { display: block; min-width: 0; height: calc(100vh - 68px); padding: 21px 16px 19px; overflow-y: auto; position: sticky; top: 68px; scrollbar-width: thin; scrollbar-color: var(--docs-line-strong) transparent; }
.layout-docs .ui-sidebar__title { display: flex; align-items: center; gap: 7px; padding: 26px 8px 9px; margin: 0; color: var(--docs-muted); font-size: 10px; font-weight: 650; letter-spacing: .1em; text-transform: uppercase; }
.layout-docs .ui-sidebar__title::before { width: 14px; height: 1px; background: var(--docs-orange); content: ""; }
.layout-docs .ui-sidebar__nav { margin-top: 8px; }
.layout-docs .ui-sidebar__list, .layout-docs .ui-sidebar__sublist { gap: 6px; }
.layout-docs .ui-sidebar__sublist { margin-top: 5px; margin-inline-start: 18px; padding-left: 0; border-left: 1px solid var(--docs-line-strong); }
.layout-docs .ui-sidebar__link { min-height: 40px; padding: 10px; border-radius: 8px; color: var(--docs-muted); font-size: 12px; line-height: 1.25; transition: .16s ease; }
.layout-docs .ui-sidebar__link:hover, .layout-docs .ui-sidebar__link:focus-visible { color: var(--docs-ink); background: var(--docs-paper-2); }
.layout-docs .ui-sidebar__link[aria-current="page"] { color: var(--docs-orange-deep); background: var(--docs-orange-wash); font-weight: 650; }
.layout-docs .ui-sidebar__link[aria-current="page"] .ui-sidebar__icon { color: var(--docs-orange); }
.layout-docs .ui-sidebar__group > summary.ui-sidebar__link:has(svg[data-fastr-docs-active="true"]) { color: var(--docs-orange-deep); background: var(--docs-orange-wash); font-weight: 650; }
.layout-docs .ui-sidebar__group > summary.ui-sidebar__link:has(svg[data-fastr-docs-active="true"]) .ui-sidebar__icon { color: var(--docs-orange); }
.layout-docs .ui-sidebar__item--sub .ui-sidebar__link { min-height: 38px; padding: 9px 10px; border-radius: 0 7px 7px 0; font-size: 11px; }
.layout-docs .ui-sidebar__icon { display: inline-flex; flex: 0 0 auto; color: var(--docs-faint); }
.layout-docs .ui-sidebar__link[aria-current="page"] .ui-sidebar__icon { color: var(--docs-orange); }
.layout-docs .ui-sidebar__icon:has(> .fastr-docs-nav-badge), .ui-sidebar--drawer-body .ui-sidebar__icon:has(> .fastr-docs-nav-badge) { display: contents; }
.layout-docs .ui-sidebar__link > .ui-sidebar__label, .ui-sidebar--drawer-body .ui-sidebar__link > .ui-sidebar__label { order: 1; min-width: 0; }
.layout-docs .fastr-docs-nav-badge, .ui-sidebar--drawer-body .fastr-docs-nav-badge { order: 2; display: inline-flex; align-items: center; max-width: 82px; margin-left: auto; padding: 3px 6px; overflow: hidden; border: 1px solid var(--docs-line); border-radius: 999px; color: var(--docs-muted); background: var(--docs-paper-2); font-family: var(--font-mono, monospace); font-size: 8px; font-weight: 650; letter-spacing: .03em; line-height: 1; text-overflow: ellipsis; white-space: nowrap; }
.layout-docs .fastr-docs-nav-badge::after, .ui-sidebar--drawer-body .fastr-docs-nav-badge::after { content: attr(data-badge-label); }
.layout-docs .fastr-docs-nav-badge--accent, .ui-sidebar--drawer-body .fastr-docs-nav-badge--accent { border-color: color-mix(in srgb, var(--docs-orange) 34%, var(--docs-line)); color: var(--docs-orange-deep); background: var(--docs-orange-wash); }
.layout-docs .fastr-docs-nav-badge--info, .ui-sidebar--drawer-body .fastr-docs-nav-badge--info { border-color: color-mix(in srgb, var(--docs-blue) 34%, var(--docs-line)); color: color-mix(in srgb, var(--docs-blue) 72%, var(--docs-ink)); background: color-mix(in srgb, var(--docs-blue) 12%, transparent); }
.layout-docs .fastr-docs-nav-badge--success, .ui-sidebar--drawer-body .fastr-docs-nav-badge--success { border-color: color-mix(in srgb, var(--docs-sage) 34%, var(--docs-line)); color: color-mix(in srgb, var(--docs-sage) 76%, var(--docs-ink)); background: var(--docs-sage-wash); }
.layout-docs .fastr-docs-nav-badge--warning, .ui-sidebar--drawer-body .fastr-docs-nav-badge--warning { border-color: color-mix(in srgb, var(--docs-orange) 42%, var(--docs-line)); color: color-mix(in srgb, var(--docs-orange) 82%, var(--docs-ink)); background: color-mix(in srgb, var(--docs-orange) 16%, transparent); }
.layout-docs .ui-sidebar__footer { margin-top: 17px; padding-top: 22px; border-top: 1px solid var(--docs-line); }

/* The mobile drawer is mounted at document level by GoFastr, outside the
   layout-docs wrapper. Keep its spacing, hierarchy, and color treatment in
   lockstep with the persistent rail above. */
.ui-sidebar--drawer-body {
  --docs-paper: var(--color-surface, #fffdfa);
  --docs-paper-2: var(--color-surface-soft, #f1f0e9);
  --docs-ink: var(--color-text, #1c1f1d);
  --docs-muted: var(--color-text-muted, #7c827a);
  --docs-faint: var(--color-text-subtle, #a7aca4);
  --docs-line: var(--color-border, #e1e2db);
  --docs-line-strong: var(--color-border-strong, #c9cdc3);
  --docs-orange: var(--color-accent, #ec7131);
  --docs-orange-deep: var(--color-primary, #c9501e);
  --docs-orange-wash: color-mix(in srgb, var(--docs-orange) 12%, transparent);
  box-sizing: border-box;
  width: 100%;
  min-width: 0;
  padding: 26px 18px 32px;
  color: var(--docs-ink);
  background: var(--docs-canvas, var(--color-background, #f7f5ef));
}
.ui-sidebar--drawer-body .ui-sidebar__title { display: flex; align-items: center; gap: 8px; padding: 0 10px 16px; margin: 0; color: var(--docs-muted); font-size: 10px; font-weight: 650; letter-spacing: .1em; line-height: 1.2; text-transform: uppercase; }
.ui-sidebar--drawer-body .ui-sidebar__title::before { width: 14px; height: 1px; flex: 0 0 auto; background: var(--docs-orange); content: ""; }
.ui-sidebar--drawer-body .ui-sidebar__nav { margin-top: 0; }
.ui-sidebar--drawer-body .ui-sidebar__list, .ui-sidebar--drawer-body .ui-sidebar__sublist { gap: 6px; }
.ui-sidebar--drawer-body .ui-sidebar__sublist { margin-top: 5px; margin-inline-start: 18px; padding-left: 0; border-left: 1px solid var(--docs-line-strong); }
.ui-sidebar--drawer-body .ui-sidebar__link { min-height: 44px; padding: 10px 12px; gap: 10px; border-radius: 8px; color: var(--docs-muted); font-size: 14px; line-height: 1.25; }
.ui-sidebar--drawer-body .ui-sidebar__link:hover, .ui-sidebar--drawer-body .ui-sidebar__link:focus-visible { color: var(--docs-ink); background: var(--docs-paper-2); }
.ui-sidebar--drawer-body .ui-sidebar__link[aria-current="page"] { color: var(--docs-orange-deep); background: var(--docs-orange-wash); font-weight: 650; }
.ui-sidebar--drawer-body .ui-sidebar__link[aria-current="page"] .ui-sidebar__icon { color: var(--docs-orange); }
.ui-sidebar--drawer-body .ui-sidebar__group > summary.ui-sidebar__link.fastr-docs-nav-group--active { color: var(--docs-orange-deep); background: var(--docs-orange-wash); font-weight: 650; }
.ui-sidebar--drawer-body .ui-sidebar__group > summary.ui-sidebar__link.fastr-docs-nav-group--active .ui-sidebar__icon { color: var(--docs-orange); }
.ui-sidebar--drawer-body .ui-sidebar__item--sub .ui-sidebar__link { min-height: 42px; padding: 10px 12px; border-radius: 0 7px 7px 0; font-size: 13px; }
.ui-sidebar--drawer-body .ui-sidebar__icon { color: var(--docs-faint); }
.ui-sidebar--drawer-body .fastr-docs-nav-badge { max-width: 96px; padding: 4px 7px; font-size: 9px; }
.ui-sidebar--drawer-body .ui-sidebar__footer { margin-top: 20px; padding-top: 24px; border-top: 1px solid var(--docs-line); }

/* Generated landing surface */
.fastr-docs-home { width: min(900px, calc(100% - 88px)); margin-inline: auto; padding: 47px 0 75px; }
.fastr-docs-home__hero { display: grid; grid-template-columns: minmax(0, 1.05fr) minmax(280px, .95fr); gap: 44px; align-items: center; min-height: 450px; }
.fastr-docs-home__eyebrow { color: var(--docs-orange-deep); font-family: var(--font-mono, monospace); font-size: 10px; font-weight: 650; letter-spacing: .1em; text-transform: uppercase; }
.fastr-docs-home h1 { max-width: 610px; margin: 15px 0 16px; color: var(--docs-ink); font-family: var(--font-heading, Inter, sans-serif); font-size: clamp(42px, 5vw, 67px); letter-spacing: -.065em; line-height: .98; }
.fastr-docs-home__hero-copy > p { max-width: 470px; margin: 0; color: var(--docs-muted); font-size: 17px; line-height: 1.55; }
.fastr-docs-home__actions { display: flex; flex-wrap: wrap; gap: 9px; margin-top: 27px; }
.fastr-docs-home__button { display: inline-flex; align-items: center; gap: 8px; padding: 10px 13px; border: 1px solid var(--docs-line); border-radius: 6px; color: var(--docs-ink); background: var(--docs-paper); font-size: 11px; font-weight: 650; text-decoration: none !important; transition: .18s ease; }
.fastr-docs-home__button:hover { border-color: var(--docs-orange); transform: translateY(-1px); }
.fastr-docs-home__button--primary { border-color: var(--docs-orange-deep); color: #fff; background: var(--docs-orange-deep); }
.fastr-docs-home__button--ghost { color: var(--docs-orange-deep); }
.fastr-docs-home__hero-art { position: relative; min-height: 270px; overflow: hidden; border: 1px solid var(--docs-line); border-radius: 9px; background: var(--docs-paper); box-shadow: var(--docs-shadow); }
.fastr-docs-home__grid { position: absolute; inset: 0; opacity: .65; background-image: linear-gradient(var(--docs-line) 1px, transparent 1px), linear-gradient(90deg, var(--docs-line) 1px, transparent 1px); background-size: 25px 25px; }
.fastr-docs-home__hero-art::before, .fastr-docs-home__hero-art::after { position: absolute; width: 155%; height: 1px; background: color-mix(in srgb, var(--docs-blue) 50%, transparent); content: ""; transform: rotate(-19deg); }
.fastr-docs-home__hero-art::before { left: -30%; top: 53%; }
.fastr-docs-home__hero-art::after { left: -17%; top: 76%; background: color-mix(in srgb, var(--docs-orange) 50%, transparent); transform: rotate(23deg); }
.fastr-docs-home__orbit { position: absolute; width: 150px; height: 150px; border: 1px solid color-mix(in srgb, var(--docs-orange) 48%, transparent); border-radius: 50%; }
.fastr-docs-home__orbit--one { right: -72px; top: -80px; }
.fastr-docs-home__orbit--two { right: -38px; top: -46px; width: 82px; height: 82px; }
.fastr-docs-home__pin { position: absolute; left: 43%; top: 48%; display: flex; align-items: center; justify-content: center; width: 46px; height: 46px; border: 2px solid var(--docs-orange); border-radius: 50%; box-shadow: 0 0 0 9px color-mix(in srgb, var(--docs-orange) 12%, transparent); }
.fastr-docs-home__pin span { width: 9px; height: 9px; border-radius: 50%; background: var(--docs-orange); }
.fastr-docs-home__art-label { position: absolute; right: 14px; bottom: 12px; color: var(--docs-muted); font-family: var(--font-mono, monospace); font-size: 8px; letter-spacing: .08em; }
.fastr-docs-home__metrics { display: grid; grid-template-columns: repeat(3, 1fr); gap: 0; margin: 28px 0 66px; padding: 17px 0; border-top: 1px solid var(--docs-line); border-bottom: 1px solid var(--docs-line); }
.fastr-docs-home__metrics div { display: grid; gap: 5px; padding-left: 18px; border-left: 1px solid var(--docs-line); }
.fastr-docs-home__metrics div:first-child { padding-left: 0; border-left: 0; }
.fastr-docs-home__metrics strong { color: var(--docs-ink); font-family: var(--font-heading, Inter, sans-serif); font-size: 27px; letter-spacing: -.05em; }
.fastr-docs-home__metrics span { color: var(--docs-muted); font-family: var(--font-mono, monospace); font-size: 9px; text-transform: uppercase; letter-spacing: .08em; }
.fastr-docs-home__section { margin-top: 59px; }
.fastr-docs-home__section-intro { display: grid; grid-template-columns: 1fr 1fr; gap: 24px; align-items: end; margin-bottom: 19px; }
.fastr-docs-home__section-intro h2 { margin: 0; color: var(--docs-ink); font-family: var(--font-heading, Inter, sans-serif); font-size: 26px; letter-spacing: -.045em; }
.fastr-docs-home__section-intro p { margin: 0; color: var(--docs-muted); font-size: 11px; line-height: 1.55; }
.fastr-docs-home__feature-grid { display: grid; grid-template-columns: repeat(3, 1fr); gap: 10px; }
.fastr-docs-home__feature-grid article { min-height: 142px; padding: 16px; border: 1px solid var(--docs-line); border-radius: 7px; background: var(--docs-paper); }
.fastr-docs-home__feature-icon { display: inline-flex; align-items: center; justify-content: center; width: 28px; height: 28px; margin-bottom: 16px; border-radius: 6px; color: var(--docs-orange-deep); background: var(--docs-orange-wash); font-family: var(--font-mono, monospace); }
.fastr-docs-home__feature-grid h3 { margin: 0 0 6px; color: var(--docs-ink); font-family: var(--font-heading, Inter, sans-serif); font-size: 14px; }
.fastr-docs-home__feature-grid p { margin: 0; color: var(--docs-muted); font-size: 10px; line-height: 1.5; }
.fastr-docs-home__anatomy { display: grid; grid-template-columns: .8fr 1.2fr; gap: 10px; }
.fastr-docs-home__route-panel, .fastr-docs-home__code { overflow: hidden; margin: 0; border: 1px solid var(--docs-line); border-radius: 7px; background: var(--docs-paper); }
.fastr-docs-home__panel-head { display: flex; align-items: center; justify-content: space-between; padding: 11px 13px; border-bottom: 1px solid var(--docs-line); color: var(--docs-ink); font-family: var(--font-mono, monospace); font-size: 9px; }
.fastr-docs-home__panel-head span { color: var(--docs-faint); font-weight: 400; }
.fastr-docs-home__route-row { display: grid; grid-template-columns: 43px 1fr auto; gap: 8px; align-items: center; padding: 10px 13px; border-bottom: 1px solid var(--docs-line); color: var(--docs-muted); font-size: 10px; }
.fastr-docs-home__route-row:last-child { border-bottom: 0; }
.fastr-docs-home__route-row > span { color: var(--docs-faint); font-family: var(--font-mono, monospace); font-size: 8px; }
.fastr-docs-home__route-row b { color: var(--docs-ink-soft); font-weight: 500; }
.fastr-docs-home__route-row code { color: var(--docs-faint); font-size: 8px; }
.fastr-docs-home__route-row--active { background: var(--docs-orange-wash); }
.fastr-docs-home__route-row--active b { color: var(--docs-orange-deep); }
.fastr-docs-home__code { padding: 0 13px 16px; color: var(--docs-code-text); background: var(--docs-code); font-size: 10px; line-height: 1.8; }
.fastr-docs-home__code .fastr-docs-home__panel-head { padding-inline: 0; border-color: #2a342c; color: var(--docs-code-text); }
.fastr-docs-home__code code { display: block; overflow-x: auto; }
.fastr-docs-home__code em { color: #b9d78e; font-style: normal; }
.fastr-docs-home__code b { color: #ed9d68; font-weight: 400; }
.fastr-docs-home__callout { display: flex; gap: 13px; align-items: flex-start; margin-top: 34px; padding: 16px; border: 1px solid color-mix(in srgb, var(--docs-orange) 28%, var(--docs-line)); border-radius: 8px; background: linear-gradient(135deg, var(--docs-orange-wash), var(--docs-paper)); }
.fastr-docs-home__callout-mark { display: inline-flex; align-items: center; justify-content: center; width: 28px; height: 28px; border: 1px solid var(--docs-orange); border-radius: 50%; color: var(--docs-orange); font-size: 18px; }
.fastr-docs-home__callout strong { display: block; margin-bottom: 4px; color: var(--docs-ink); font-size: 12px; }
.fastr-docs-home__callout p { max-width: 600px; margin: 0; color: var(--docs-muted); font-size: 11px; line-height: 1.55; }

/* Article and in-page rail */
.layout-docs main { min-width: 0; }
.fastr-docs-doc-layout { --ui-doc-layout-pad: 0; --ui-doc-layout-gap: 0; display: grid; grid-template-columns: minmax(0, 1fr) 198px; gap: 0; max-width: none; padding: 0; }
.layout-docs .fastr-docs-doc-layout[data-fui-comp="ui-doc-layout"] { grid-template-columns: minmax(0, 1fr) 198px; }
.fastr-docs-doc-layout > .ui-doc-layout__content { flex: 1 1 auto; width: 100%; min-width: 0; max-width: none; box-sizing: border-box; padding: 0; }
.fastr-docs-doc-layout .ui-doc-layout__crumbs, .fastr-docs-doc-layout .ui-markdown { width: min(900px, calc(100% - 88px)); margin-inline: auto; }
.fastr-docs-doc-layout .ui-doc-layout__crumbs { padding-top: 47px; color: var(--docs-faint); font-family: var(--font-mono, monospace); font-size: 9px; }
.fastr-docs-doc-layout .ui-doc-layout__crumbs a { color: var(--docs-muted); }
.fastr-docs-doc-layout .ui-markdown { padding: 20px 0 75px; }
.fastr-docs-doc-layout > .scrollspy { position: sticky; top: 68px; align-self: start; height: max-content; max-height: calc(100vh - 100px); overflow-y: auto; padding: 47px 21px 24px 20px; border-left: 1px solid var(--docs-line); scrollbar-width: thin; scrollbar-color: var(--docs-line-strong) transparent; }
.fastr-docs-doc-layout > .scrollspy .fastr-docs-toc { position: static; }
.fastr-docs-doc-layout > .scrollspy > .ui-anchored-rail { min-width: 0; }
.fastr-docs-doc-layout .ui-anchored-rail__label { margin-bottom: 13px; color: var(--docs-faint); font-family: var(--font-mono, monospace); font-size: 9px; font-weight: 650; letter-spacing: .11em; text-transform: uppercase; }
.fastr-docs-doc-layout .ui-anchored-rail__list { display: grid; gap: 0; padding: 0; margin: 0; border-left: 1px solid var(--docs-line); list-style: none; }
.fastr-docs-doc-layout .ui-anchored-rail__list a { display: block; padding: 6px 0 6px 13px; border-left: 1px solid transparent; margin-left: -1px; color: var(--docs-muted); font-size: 10px; line-height: 1.35; transition: .16s ease; }
.fastr-docs-doc-layout .ui-anchored-rail__list a:hover, .fastr-docs-doc-layout .ui-anchored-rail__list a[aria-current="true"] { border-left-color: var(--docs-orange); color: var(--docs-orange-deep); }
.fastr-docs-doc-layout > .fastr-docs-toc-select { display: none; }
.fastr-docs-toc-select .ui-select__label { color: var(--docs-faint); font-family: var(--font-mono, monospace); font-size: 9px; font-weight: 650; letter-spacing: .11em; text-transform: uppercase; }
.fastr-docs-toc-select .ui-select__input { min-width: 0; min-height: 34px; padding: 7px 34px 7px 10px; border-color: var(--docs-line); border-radius: 6px; color: var(--docs-ink); background-color: var(--docs-paper); background-image: linear-gradient(45deg, transparent 50%, var(--docs-muted) 50%), linear-gradient(135deg, var(--docs-muted) 50%, transparent 50%); background-position: calc(100% - 15px) 15px, calc(100% - 10px) 15px; background-size: 5px 5px, 5px 5px; }

/* Markdown is the portable Page surface. */
.layout-docs .ui-markdown { color: var(--docs-ink-soft); font-size: 14px; line-height: 1.72; }
.layout-docs .ui-markdown h1, .layout-docs .ui-markdown h2, .layout-docs .ui-markdown h3, .layout-docs .ui-markdown h4 { color: var(--docs-ink); font-family: var(--font-heading, Inter, sans-serif); letter-spacing: -.045em; }
.layout-docs .ui-markdown h1 { margin: 0 0 16px; font-size: clamp(40px, 5vw, 64px); line-height: 1.02; }
.layout-docs .ui-markdown h2 { margin-top: 58px; margin-bottom: 13px; padding-top: 0; scroll-margin-top: 110px; font-size: 26px; line-height: 1.1; }
.layout-docs .ui-markdown h3 { margin-top: 34px; margin-bottom: 9px; scroll-margin-top: 110px; font-size: 18px; }
.layout-docs .ui-markdown p { max-width: 73ch; margin: 0 0 16px; }
.layout-docs .ui-markdown a { color: var(--docs-orange-deep); text-decoration: underline; text-decoration-color: color-mix(in srgb, var(--docs-orange) 45%, transparent); text-underline-offset: 3px; }
.layout-docs .ui-markdown code { padding: 2px 5px; border-radius: 4px; color: var(--docs-orange-deep); background: var(--docs-orange-wash); font-size: .86em; }
.layout-docs .ui-markdown pre:not(.ui-code-block):not(.ui-code-block__body) { overflow-x: auto; padding: 17px 19px 19px; border: 1px solid var(--color-code-border, #2d392f); border-radius: 8px; color: var(--docs-code-text); background: var(--docs-code); box-shadow: var(--docs-shadow); font-size: 11px; line-height: 1.8; }
.layout-docs .ui-markdown pre code { padding: 0; color: inherit; background: transparent; }
.layout-docs .ui-markdown [data-fui-comp="ui-code-block"] { margin: 20px 0 24px; border-radius: 8px; box-shadow: var(--docs-shadow); }
.layout-docs .ui-markdown [data-fui-comp="ui-code-block"] .ui-code-block__head { padding: 2px 12px; }
.layout-docs .ui-markdown [data-fui-comp="ui-code-block"] .ui-code-block__body { padding: 12px 16px 13px; border: 0; border-radius: 0 0 8px 8px; font-size: 11px; line-height: 1.5; }
.layout-docs .ui-markdown blockquote { margin: 24px 0; padding: 13px 17px; border-left: 2px solid var(--docs-orange); color: var(--docs-muted); background: var(--docs-orange-wash); }
.layout-docs .ui-markdown table { width: 100%; border-collapse: collapse; }
.layout-docs .ui-markdown th, .layout-docs .ui-markdown td { padding: 9px 10px; border-bottom: 1px solid var(--docs-line); text-align: left; }
.layout-docs .ui-markdown th { color: var(--docs-faint); font-family: var(--font-mono, monospace); font-size: 10px; text-transform: uppercase; letter-spacing: .08em; }
.layout-docs .ui-doc-layout__foot { width: min(900px, calc(100% - 88px)); margin: 0 auto; padding-bottom: 65px; }
.layout-docs .ui-doc-layout__prev, .layout-docs .ui-doc-layout__next { border-color: var(--docs-line); border-radius: 7px; background: var(--docs-paper); }
.layout-docs .ui-doc-layout__prev:hover, .layout-docs .ui-doc-layout__next:hover { border-color: var(--docs-orange); transform: translateY(-1px); }
.layout-docs .ui-doc-layout__pager-dir { color: var(--docs-faint); font-family: var(--font-mono, monospace); font-size: 9px; text-transform: uppercase; letter-spacing: .08em; }
.layout-docs .ui-doc-layout__pager-ttl { color: var(--docs-orange-deep); font-weight: 650; }
.layout-docs .fastr-docs-edit-link { width: min(900px, calc(100% - 88px)); margin: 7px auto 0; color: var(--docs-faint); font-family: var(--font-mono, monospace); font-size: 10px; }
.layout-docs .fastr-docs-edit-link a { color: var(--docs-muted); }
.layout-docs .fastr-docs-edit-link a:hover { color: var(--docs-orange-deep); }
.layout-docs .fastr-docs-page-meta { width: min(900px, calc(100% - 88px)); margin: 0 auto 14px; color: var(--docs-muted); font-family: var(--font-mono, monospace); font-size: 10px; line-height: 1.5; }
.layout-docs .fastr-docs-page-meta time { color: var(--docs-ink-soft); }

/* Branded fallback for URLs that are not in the Router. */
.fastr-docs-not-found { width: min(620px, calc(100% - 38px)); margin: 0 auto; padding: 96px 0 120px; }
.fastr-docs-not-found__code { margin: 0 0 12px; color: var(--docs-orange-deep, var(--color-primary)); font: 650 11px var(--font-mono, monospace); letter-spacing: .12em; }
.fastr-docs-not-found h1 { margin: 0 0 14px; color: var(--docs-ink, var(--color-text)); font: 650 clamp(40px, 7vw, 68px)/1 var(--font-heading, sans-serif); letter-spacing: -.06em; }
.fastr-docs-not-found__message { margin: 0 0 26px; color: var(--docs-muted, var(--color-text-muted)); font-size: 16px; line-height: 1.6; }
.fastr-docs-not-found__link { display: inline-flex; padding: 10px 13px; border: 1px solid var(--docs-line, var(--color-border)); border-radius: 7px; color: var(--docs-orange-deep, var(--color-primary)); background: var(--docs-paper, var(--color-surface)); font-size: 12px; font-weight: 650; text-decoration: none; }
.fastr-docs-not-found__link:hover, .fastr-docs-not-found__link:focus-visible { border-color: var(--docs-orange, var(--color-accent)); }

/* Framework controls use the same editorial surfaces. */
.layout-docs .ui-sidebar__hamburger { border-color: var(--docs-line); color: var(--docs-muted); background: var(--docs-paper); }
.layout-docs .ui-sidebar__hamburger:hover { color: var(--docs-ink); border-color: var(--docs-line-strong); }
.layout-docs [data-fui-comp="ui-button"] { border-radius: 6px; }

@media (max-width: 1120px) {
  .fastr-docs-doc-layout { display: flex; flex-direction: column; }
  .fastr-docs-doc-layout > * { min-width: 0; max-width: 100%; box-sizing: border-box; }
  .fastr-docs-doc-layout > .scrollspy { display: none; }
  .fastr-docs-doc-layout > .fastr-docs-toc-select { display: flex; order: -1; position: sticky; top: 68px; z-index: 12; width: min(900px, calc(100% - 72px)); box-sizing: border-box; align-items: center; gap: 12px; height: auto; margin: 18px auto 8px; padding: 11px 13px; border: 1px solid var(--docs-line); border-radius: 7px; background: var(--docs-paper); }
  .fastr-docs-doc-layout > .fastr-docs-toc-select .ui-select__label { flex: 0 0 auto; margin: 0; }
  .fastr-docs-doc-layout > .fastr-docs-toc-select .ui-select__input { flex: 1 1 auto; width: auto; }
  .layout-docs .ui-markdown h2, .layout-docs .ui-markdown h3 { scroll-margin-top: 164px; }
  .fastr-docs-doc-layout .ui-doc-layout__crumbs, .fastr-docs-doc-layout .ui-markdown, .layout-docs .ui-doc-layout__foot { width: min(900px, calc(100% - 72px)); }
  .fastr-docs-doc-layout .ui-doc-layout__crumbs { padding-top: 12px; }
  .fastr-docs-home { width: min(900px, calc(100% - 72px)); }
}

@media (min-width: 1121px) {
  .fastr-docs-doc-layout.ui-doc-layout--narrow > .ui-doc-layout__content { margin-inline: 0; }
  .fastr-docs-doc-layout .ui-doc-layout__crumbs, .fastr-docs-doc-layout .ui-markdown { width: min(940px, calc(100% - 48px)); }
}

@media (max-width: 767px) {
  .layout-docs > header { height: 62px; }
  .fastr-docs-site-header { height: 62px; padding-inline: 16px; gap: 12px; }
  .fastr-docs-site-header .ui-site-header__mobile { display: none !important; }
  .fastr-docs-site-header .ui-site-header__links { display: none; }
  .fastr-docs-site-header .ui-site-header__right { gap: 2px; }
  .fastr-docs-site-header .ui-site-header__bar-actions { display: flex !important; align-items: center; gap: 2px; }
  .fastr-docs-site-header .ui-site-header__brand { flex: 1 1 auto; }
  .fastr-docs-site-header .ui-site-header__brand--mobile { gap: 2px; }
  .fastr-docs-site-header .ui-site-header__brand--mobile .fastr-docs-brand { flex: 0 0 auto; }
  .fastr-docs-mobile-nav-trigger { display: inline-flex; align-items: center; justify-content: center; width: 34px; height: 34px; margin-right: 2px; padding: 0; border: 1px solid transparent; border-radius: 7px; color: var(--docs-muted); background: transparent; cursor: pointer; }
  .fastr-docs-mobile-nav-trigger:hover, .fastr-docs-mobile-nav-trigger:focus-visible { color: var(--docs-ink); border-color: var(--docs-line); background: var(--docs-paper); }
  .fastr-docs-brand__name { font-size: 16px; }
  .fastr-docs-command-trigger { width: 36px; padding-inline: 0; justify-content: center; border-color: transparent; background: transparent; }
  .fastr-docs-command-trigger__label, .fastr-docs-command-trigger__hint { display: none; }
  .fastr-docs-variant-selectors { gap: 2px; }
  .fastr-docs-variant-select select { min-width: 40px; width: 40px; padding-inline: 4px 15px; font-size: 10px; }
  .layout-docs .layout-body { min-height: calc(100vh - 62px); }
  .layout-docs .layout-body > nav { flex: 0 0 auto; border-right: 0; }
  .layout-docs .ui-sidebar__inline { display: none; }
  .layout-docs .ui-sidebar__hamburger { position: fixed; z-index: 45; top: 76px; left: 16px; width: 34px; height: 34px; padding: 0; border-color: transparent; background: transparent; font-size: 0; }
  .layout-docs .ui-sidebar__hamburger span { font-size: 0; }
  .layout-docs .ui-sidebar__hamburger::before { width: 17px; height: 13px; border-top: 1.5px solid currentColor; border-bottom: 1.5px solid currentColor; content: ""; }
  .layout-docs .ui-sidebar__hamburger::after { position: absolute; width: 17px; height: 1.5px; background: currentColor; content: ""; }
  .fastr-docs-doc-layout .ui-doc-layout__crumbs, .fastr-docs-doc-layout .ui-markdown, .layout-docs .ui-doc-layout__foot { width: calc(100% - 38px); }
  .layout-docs .fastr-docs-edit-link, .layout-docs .fastr-docs-page-meta { width: calc(100% - 38px); }
  .fastr-docs-home { width: calc(100% - 38px); padding-top: 64px; }
  .fastr-docs-home__hero, .fastr-docs-home__anatomy { grid-template-columns: 1fr; gap: 25px; }
  .fastr-docs-home__hero-art { min-height: 190px; }
  .fastr-docs-home h1 { font-size: 41px; }
  .fastr-docs-home__hero-copy > p { font-size: 15px; }
  .fastr-docs-home__section-intro { display: block; }
  .fastr-docs-home__section-intro p { margin-top: 8px; }
  .fastr-docs-home__feature-grid { grid-template-columns: 1fr; }
  .fastr-docs-home__metrics { margin-bottom: 40px; }
  .fastr-docs-doc-layout .ui-doc-layout__crumbs { padding-top: 0; }
  .fastr-docs-doc-layout > .fastr-docs-toc-select { top: 62px; width: calc(100% - 38px); margin-top: 13px; }
  .fastr-docs-doc-layout > .fastr-docs-toc-select { display: block; margin-bottom: 0; padding: 10px 11px 11px; }
  .fastr-docs-doc-layout > .fastr-docs-toc-select .ui-select__label { display: block; margin: 0 0 6px; }
  .fastr-docs-doc-layout > .fastr-docs-toc-select .ui-select__input { display: block; width: 100%; box-sizing: border-box; }
  .fastr-docs-doc-layout .ui-doc-layout__content { margin-top: 10px; }
  .fastr-docs-doc-layout .ui-markdown { padding-top: 18px; }
  .layout-docs .ui-markdown h1 { font-size: 41px; }
  .layout-docs .ui-markdown h2 { margin-top: 45px; scroll-margin-top: 144px; font-size: 23px; }
  .layout-docs .ui-doc-layout__foot-nav { grid-template-columns: 1fr; }
  .layout-docs .ui-doc-layout__next { text-align: left; }
}

@media (max-width: 540px) {
  [data-fui-widget="fastr-docs-command-palette"].fui-pos-center { align-items: stretch; justify-content: stretch; padding: 0; }
  [data-fui-widget="fastr-docs-command-palette"] > .fui-panel { width: 100%; max-width: none; height: 100%; max-height: none; border: 0; border-radius: 0; }
  [data-fui-widget="fastr-docs-command-palette"] .ui-cmd-palette { flex: 1 1 auto; }
  [data-fui-widget="fastr-docs-command-palette"] .ui-cmd-palette__combobox { flex: 1 1 auto; }
  [data-fui-widget="fastr-docs-command-palette"] .ui-cmd-palette__combobox .combobox__listbox { flex: 1 1 auto; max-height: none; }
}

@media (prefers-reduced-motion: reduce) {
  html { scroll-behavior: auto; }
  .fastr-docs-command-trigger, .layout-docs .ui-sidebar__link, .layout-docs .ui-doc-layout__prev, .layout-docs .ui-doc-layout__next { transition: none; }
}

/* Blog publication surface. Blog is a sibling layout to documentation: it
   keeps the global brand and command palette, but owns its own navigation,
   aggregation views, cards, and post reading shape. */
.layout-blog { --blog-content-max: 1040px; min-width: 0; background: transparent; }
.layout-blog .layout-body { display: grid; grid-template-columns: 224px minmax(0, 1fr); min-height: calc(100vh - var(--nav-h, 68px)); }
.layout-blog .layout-body > nav { width: auto; min-width: 0; border-right: 1px solid var(--docs-line); background: color-mix(in srgb, var(--docs-canvas) 80%, transparent); }
.layout-blog .ui-sidebar__inline { position: sticky; top: 68px; height: calc(100vh - 68px); box-sizing: border-box; padding: 30px 16px 24px; overflow-y: auto; }
.layout-blog .ui-sidebar__title { padding: 0 8px 16px; }
.layout-blog .ui-sidebar__nav { margin-top: 0; }
.layout-blog .ui-sidebar__list { gap: 4px; }
.layout-blog .ui-sidebar__link { min-height: 38px; padding: 9px 10px; border-radius: 7px; font-size: 12px; }
.layout-blog .ui-sidebar__item:last-child { margin-top: 24px; padding-top: 18px; border-top: 1px solid var(--docs-line); }
.layout-blog .ui-sidebar__item:last-child > .ui-sidebar__link { color: var(--docs-faint); font-family: var(--font-mono, monospace); font-size: 10px; text-transform: uppercase; letter-spacing: .06em; }
.layout-blog .layout-content { min-width: 0; overflow: hidden; }
.fastr-docs-blog-page { width: min(var(--blog-content-max), calc(100% - 96px)); margin: 0 auto; padding: 72px 0 96px; color: var(--docs-ink-soft); }
.fastr-docs-blog__header { max-width: 760px; margin-bottom: 44px; }
.fastr-docs-blog__eyebrow { margin: 0 0 14px; color: var(--docs-orange-deep); font: 650 10px/1.2 var(--font-mono, monospace); letter-spacing: .12em; text-transform: uppercase; }
.fastr-docs-blog__header h1 { margin: 0 0 16px; color: var(--docs-ink); font: 700 clamp(42px, 6vw, 76px)/.98 var(--font-heading, Inter, sans-serif); letter-spacing: -.065em; }
.fastr-docs-blog__lede { max-width: 64ch; margin: 0; color: var(--docs-muted); font-size: 18px; line-height: 1.55; }
.fastr-docs-blog__toolbar { margin-top: 28px; }
.fastr-docs-blog__toolbar-nav { display: flex; flex-wrap: wrap; align-items: center; gap: 7px; }
.fastr-docs-blog__toolbar-link, .fastr-docs-blog__feed-link { display: inline-flex; min-height: 32px; align-items: center; padding: 0 10px; border: 1px solid var(--docs-line); border-radius: 6px; color: var(--docs-muted); background: var(--docs-paper); font: 10px var(--font-mono, monospace); text-decoration: none; }
.fastr-docs-blog__toolbar-link:hover, .fastr-docs-blog__toolbar-link:focus-visible, .fastr-docs-blog__feed-link:hover, .fastr-docs-blog__feed-link:focus-visible { border-color: var(--docs-orange); color: var(--docs-orange-deep); }
.fastr-docs-blog__feed-link { margin-left: auto; color: var(--docs-orange-deep); }
.fastr-docs-blog__intro { max-width: 68ch; margin: -12px 0 45px; color: var(--docs-muted); }
.fastr-docs-blog__intro.ui-markdown p { margin-bottom: 12px; }
.fastr-docs-blog__section-label { margin: 0 0 12px; color: var(--docs-faint); font: 650 10px var(--font-mono, monospace); letter-spacing: .1em; text-transform: uppercase; }
.fastr-docs-blog__featured { margin-bottom: 54px; }
.fastr-docs-blog__section-head { display: flex; align-items: baseline; justify-content: space-between; gap: 20px; margin-bottom: 16px; }
.fastr-docs-blog__section-head h2 { margin: 0; color: var(--docs-ink); font: 650 25px/1.1 var(--font-heading, Inter, sans-serif); letter-spacing: -.045em; }
.fastr-docs-blog__count, .fastr-docs-blog__search-summary { color: var(--docs-faint); font: 10px var(--font-mono, monospace); }
.fastr-docs-blog__cards { display: grid; grid-template-columns: repeat(2, minmax(0, 1fr)); gap: 14px; }
.fastr-docs-blog-card-wrap { min-width: 0; }
.fastr-docs-blog-card { min-width: 0; border-color: var(--docs-line); border-radius: 8px; background: color-mix(in srgb, var(--docs-paper) 74%, transparent); box-shadow: none; text-decoration: none; transition: border-color .18s ease, transform .18s ease, background .18s ease; }
.fastr-docs-blog-card:hover, .fastr-docs-blog-card:focus-visible { border-color: var(--docs-orange); background: var(--docs-paper); transform: translateY(-2px); }
.fastr-docs-blog-card--featured { border-color: color-mix(in srgb, var(--docs-orange) 46%, var(--docs-line)); background: linear-gradient(135deg, var(--docs-orange-wash), var(--docs-paper)); }
.fastr-docs-blog-card .ui-card__inner { min-width: 0; }
.fastr-docs-blog-card .ui-card__header { padding: 20px 20px 0; }
.fastr-docs-blog-card .ui-card__body { padding: 11px 20px 18px; }
.fastr-docs-blog-card__tags { display: flex; flex-wrap: wrap; gap: 6px; padding: 10px 20px 0; }
.fastr-docs-blog-card__meta { margin-bottom: 10px; color: var(--docs-faint); font: 10px var(--font-mono, monospace); }
.fastr-docs-blog-card h3 { margin: 0; color: var(--docs-ink); font: 650 23px/1.08 var(--font-heading, Inter, sans-serif); letter-spacing: -.045em; }
.fastr-docs-blog-card__excerpt { margin: 0; color: var(--docs-muted); font-size: 13px; line-height: 1.6; }
.fastr-docs-blog-search-item { min-width: 0; }
.fastr-docs-blog-card__tag { font-size: 9px; }
.fastr-docs-blog__empty { margin: 20px 0; padding: 20px; border: 1px dashed var(--docs-line-strong); color: var(--docs-muted); }
.fastr-docs-blog__pagination { display: flex; justify-content: space-between; gap: 12px; margin-top: 28px; padding-top: 18px; border-top: 1px solid var(--docs-line); }
.fastr-docs-blog__pager-link { color: var(--docs-orange-deep); font-size: 12px; text-decoration: none; }
.fastr-docs-blog__pager-link:hover { text-decoration: underline; text-underline-offset: 3px; }
.fastr-docs-blog-search { display: flex; align-items: end; gap: 9px; max-width: 660px; margin-top: 26px; }
.fastr-docs-blog-search label { position: absolute; width: 1px; height: 1px; overflow: hidden; clip: rect(0 0 0 0); }
.fastr-docs-blog-search input { flex: 1 1 auto; min-width: 0; height: 42px; box-sizing: border-box; padding: 0 13px; border: 1px solid var(--docs-line); border-radius: 7px; color: var(--docs-ink); background: var(--docs-paper); outline: none; }
.fastr-docs-blog-search input:focus { border-color: var(--docs-orange); box-shadow: 0 0 0 3px var(--docs-orange-wash); }
.fastr-docs-blog-search button { height: 42px; padding: 0 15px; border: 1px solid var(--docs-orange); border-radius: 7px; color: var(--docs-paper); background: var(--docs-orange-deep); cursor: pointer; font-size: 12px; font-weight: 650; }
.fastr-docs-blog__search-summary { margin: 31px 0 14px; }
.fastr-docs-blog-terms { display: grid; grid-template-columns: repeat(3, minmax(0, 1fr)); gap: 10px; }
.fastr-docs-blog-term { display: flex; align-items: center; justify-content: space-between; gap: 12px; padding: 16px; border: 1px solid var(--docs-line); border-radius: 7px; color: var(--docs-ink); background: var(--docs-paper); text-decoration: none; }
.fastr-docs-blog-term:hover, .fastr-docs-blog-term:focus-visible { border-color: var(--docs-orange); }
.fastr-docs-blog-term strong { font-size: 13px; }
.fastr-docs-blog-term span { color: var(--docs-faint); font: 10px var(--font-mono, monospace); }
.fastr-docs-blog-post__crumbs { margin-bottom: 45px; color: var(--docs-faint); font: 10px var(--font-mono, monospace); }
.fastr-docs-blog-post__crumbs a { color: var(--docs-muted); text-decoration: none; }
.fastr-docs-blog-post__crumbs a:hover { color: var(--docs-orange-deep); }
.fastr-docs-blog-post__grid { display: grid; grid-template-columns: minmax(0, 1fr) 190px; gap: 64px; align-items: start; }
.fastr-docs-blog-post__content { min-width: 0; }
.fastr-docs-blog-post__header { margin-bottom: 36px; }
.fastr-docs-blog-post__header h1 { margin: 0 0 18px; color: var(--docs-ink); font: 700 clamp(42px, 5vw, 68px)/1 var(--font-heading, Inter, sans-serif); letter-spacing: -.065em; }
.fastr-docs-blog-post__lede { max-width: 68ch; margin: 0 0 17px; color: var(--docs-muted); font-size: 17px; line-height: 1.55; }
.fastr-docs-blog-post__meta { color: var(--docs-faint); font: 10px var(--font-mono, monospace); }
.fastr-docs-blog-post__meta-item a { color: var(--docs-muted); }
.fastr-docs-blog-post__tags { display: flex; flex-wrap: wrap; gap: 6px; margin-top: 15px; }
.fastr-docs-blog-post__actions { display: flex; flex-wrap: wrap; align-items: center; gap: 7px; margin-top: 18px; }
.fastr-docs-blog-post__actions .ui-button { min-height: 38px; padding-inline: 13px; }
.fastr-docs-blog-post__share::before { margin-right: 7px; color: currentColor; content: "↗"; font-size: 15px; line-height: 1; }
.fastr-docs-blog-post__actions [data-fui-comp="ui-copy-btn"] { display: inline-flex; }
.fastr-docs-blog-post__copy .ui-copy-btn { min-height: 38px; }
.fastr-docs-blog-post__article { min-width: 0; }
.fastr-docs-blog-post__content.ui-markdown { color: var(--docs-ink-soft); font-size: 15px; line-height: 1.78; }
.fastr-docs-blog-post__content.ui-markdown h2 { margin-top: 55px; scroll-margin-top: 92px; color: var(--docs-ink); font-size: 30px; }
.fastr-docs-blog-post__content.ui-markdown h3 { margin-top: 35px; scroll-margin-top: 92px; color: var(--docs-ink); font-size: 21px; }
.fastr-docs-blog-post__content.ui-markdown p { max-width: 72ch; }
.fastr-docs-blog-post__content.ui-markdown pre { max-width: 100%; overflow-x: auto; }
.fastr-docs-blog-post__toc { position: sticky; top: 92px; min-width: 0; max-height: calc(100vh - 120px); overflow-y: auto; }
.fastr-docs-blog-post__toc .fastr-docs-toc--rail { position: static; }
.fastr-docs-blog-post__toc > .fastr-docs-toc-select { display: none; }
.fastr-docs-blog-post__edit { margin-top: 28px; color: var(--docs-faint); font: 10px var(--font-mono, monospace); }
.fastr-docs-blog-post__edit a { color: var(--docs-muted); }
.fastr-docs-blog-post__related { margin-top: 62px; padding-top: 28px; border-top: 1px solid var(--docs-line); }
.fastr-docs-blog-post__related h2 { margin: 0 0 16px; color: var(--docs-ink); font: 650 25px/1.1 var(--font-heading, Inter, sans-serif); letter-spacing: -.045em; }
.layout-blog .ui-doc-layout__foot { width: 100%; margin-top: 38px; padding: 28px 0 0; border-top: 1px solid var(--docs-line); }
.layout-blog .ui-doc-layout__foot-nav { display: grid; grid-template-columns: 1fr 1fr; gap: 14px; }
.layout-blog .ui-doc-layout__prev, .layout-blog .ui-doc-layout__next { display: flex; min-width: 0; flex-direction: column; gap: 5px; padding: 16px; border: 1px solid var(--docs-line); border-radius: 7px; color: var(--docs-ink); background: var(--docs-paper); text-decoration: none; }
.layout-blog .ui-doc-layout__prev:hover, .layout-blog .ui-doc-layout__next:hover, .layout-blog .ui-doc-layout__prev:focus-visible, .layout-blog .ui-doc-layout__next:focus-visible { border-color: var(--docs-orange); }
.layout-blog .ui-doc-layout__next { text-align: right; }
.layout-blog .ui-doc-layout__pager-dir { color: var(--docs-faint); font: 10px var(--font-mono, monospace); }
.layout-blog .ui-doc-layout__pager-ttl { overflow: hidden; color: var(--docs-ink); font-size: 13px; font-weight: 650; text-overflow: ellipsis; white-space: nowrap; }

@media (max-width: 900px) {
  .fastr-docs-blog-page { width: min(760px, calc(100% - 48px)); padding-top: 48px; }
  .fastr-docs-blog__cards { grid-template-columns: 1fr; }
  .fastr-docs-blog-terms { grid-template-columns: repeat(2, minmax(0, 1fr)); }
  .fastr-docs-blog-post__grid { grid-template-columns: 1fr; gap: 0; }
  .fastr-docs-blog-post__toc { position: static; order: -1; max-height: none; margin-bottom: 34px; }
  .fastr-docs-blog-post__toc > [data-fui-comp="scrollspy"] { display: none; }
  .fastr-docs-blog-post__toc > .fastr-docs-toc-select { display: block; width: 100%; box-sizing: border-box; padding: 11px 13px; border: 1px solid var(--docs-line); border-radius: 7px; background: var(--docs-paper); }
  .fastr-docs-blog-post__toc .ui-select__label { display: block; margin-bottom: 7px; color: var(--docs-faint); font: 650 9px var(--font-mono, monospace); letter-spacing: .1em; text-transform: uppercase; }
  .fastr-docs-blog-post__toc .ui-select__input { width: 100%; box-sizing: border-box; }
}

@media (max-width: 767px) {
  .layout-blog .layout-body { display: block; }
  .layout-blog .layout-body > nav { border-right: 0; }
  .layout-blog .ui-sidebar__inline { display: none; }
}

@media (max-width: 540px) {
  .fastr-docs-blog-page { width: calc(100% - 36px); padding-top: 37px; padding-bottom: 64px; }
  .fastr-docs-blog__header { margin-bottom: 32px; }
  .fastr-docs-blog__header h1, .fastr-docs-blog-post__header h1 { font-size: 43px; }
  .fastr-docs-blog__lede, .fastr-docs-blog-post__lede { font-size: 15px; }
  .fastr-docs-blog__toolbar-nav { gap: 5px; }
  .fastr-docs-blog__feed-link { margin-left: 0; }
  .fastr-docs-blog-terms { grid-template-columns: 1fr; }
  .fastr-docs-blog-search { align-items: stretch; flex-direction: column; }
  .fastr-docs-blog-search button { width: 100%; }
  .fastr-docs-blog-post__crumbs { margin-bottom: 32px; }
  .fastr-docs-blog-post__content.ui-markdown { font-size: 14px; }
  .layout-blog .ui-doc-layout__foot-nav { grid-template-columns: 1fr; }
  .layout-blog .ui-doc-layout__next { text-align: left; }
}
`
