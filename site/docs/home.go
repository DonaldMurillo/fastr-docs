package docs

import (
	"strconv"
	"strings"

	fastrdocs "github.com/DonaldMurillo/fastr-docs"
	"github.com/DonaldMurillo/gofastr/core/render"
)

// HomeBody is the fastr-docs project's own landing surface. It is registered
// with the same Router API that this site documents.
func HomeBody(router *fastrdocs.Router) render.HTML {
	markup := `
<div class="fastr-docs-home">
  <div class="fastr-docs-home__hero">
    <div class="fastr-docs-home__hero-copy">
      <span class="fastr-docs-home__eyebrow">Reusable documentation for GoFastr</span>
      <h1>Build docs around one route tree.</h1>
      <p>fastr-docs keeps pages, screens, search, OpenAPI, and offline delivery derived from one explicit Router.</p>
      <div class="fastr-docs-home__actions">
        <a class="fastr-docs-home__button fastr-docs-home__button--primary" href="/docs">Read the docs <span aria-hidden="true">→</span></a>
        <a class="fastr-docs-home__button fastr-docs-home__button--ghost" href="/docs/getting-started">Start building <span aria-hidden="true">↗</span></a>
      </div>
    </div>
    <div class="fastr-docs-home__hero-art" aria-label="Route graph illustration" role="img">
      <div class="fastr-docs-home__grid"></div>
      <div class="fastr-docs-home__orbit fastr-docs-home__orbit--one"></div>
      <div class="fastr-docs-home__orbit fastr-docs-home__orbit--two"></div>
      <div class="fastr-docs-home__pin"><span></span></div>
      <span class="fastr-docs-home__art-label">ROUTER / 001 · STATIC READY</span>
    </div>
  </div>

  <div class="fastr-docs-home__metrics" aria-label="fastr-docs capabilities">
    <div><strong>01</strong><span>route tree</span></div>
    <div><strong>02</strong><span>page + screen</span></div>
    <div><strong>∞</strong><span>plugins</span></div>
  </div>

  <section class="fastr-docs-home__section">
    <div class="fastr-docs-home__section-intro"><h2>One router, every surface</h2><p>Choose Markdown for durable content or a typed screen for interaction. Navigation, search, breadcrumbs, and export stay in sync.</p></div>
    <div class="fastr-docs-home__feature-grid">
      <article><span class="fastr-docs-home__feature-icon">▤</span><h3>Pages</h3><p>Markdown with front matter, headings, local search, table of contents, and offline eligibility.</p></article>
      <article><span class="fastr-docs-home__feature-icon">□</span><h3>Screens</h3><p>Typed GoFastr components for playgrounds, API explorers, and product-specific workflows.</p></article>
      <article><span class="fastr-docs-home__feature-icon">⌘</span><h3>Plugins</h3><p>OpenAPI and project extensions contribute routes, search text, assets, and validation to the same tree.</p></article>
    </div>
  </section>

  <section class="fastr-docs-home__section">
    <div class="fastr-docs-home__section-intro"><h2>Project anatomy</h2><p>The CLI creates a runnable site with explicit routes, content, agent references, and a complete browser test surface.</p></div>
    <div class="fastr-docs-home__anatomy">
      <div class="fastr-docs-home__route-panel"><div class="fastr-docs-home__panel-head"><strong>registered routes</strong><span>__ROUTE_COUNT__ routes</span></div>__ROUTE_ROWS__</div>
      <pre class="fastr-docs-home__code"><span class="fastr-docs-home__panel-head"><strong>docs/router.go</strong><span>Go</span></span><code><span>router := docs.NewRouter(</span>
  <span>docs.WithSiteName(<em>"fastr-docs"</em>),</span>
  <span>docs.WithStrictValidation(<b>true</b>),</span>
<span>)</span>
  <span>router.MustPage(<em>"/docs/guide"</em>, docs.PageConfig{...})</span>
<span>router.Use(openapi.Plugin{...})</span></code></pre>
    </div>
  </section>

  <aside class="fastr-docs-home__callout"><span class="fastr-docs-home__callout-mark">◎</span><div><strong>Designed for humans and agents</strong><p>Every generated project includes agent references, authoring skills, strict validation, and browser checks tied to the route tree.</p></div></aside>
</div>`
	markup = strings.Replace(markup, "__ROUTE_COUNT__", strconv.Itoa(len(router.PublishedRoutes())), 1)
	markup = strings.Replace(markup, "__ROUTE_ROWS__", string(homeRouteRows(router)), 1)
	return render.Raw(markup)
}

func homeRouteRows(router *fastrdocs.Router) render.HTML {
	rows := make([]render.HTML, 0)
	for _, item := range router.Navigation() {
		if item.Depth != 0 {
			continue
		}
		path := item.Path
		if item.Kind == fastrdocs.KindGroup {
			path += "/*"
		}
		active := ""
		if item.Path == "/" {
			active = " fastr-docs-home__route-row--active"
		}
		rows = append(rows, render.Raw(`<div class="fastr-docs-home__route-row`+active+`"><span>`+render.Escape(string(item.Kind))+`</span><b>`+render.Escape(item.Title)+`</b><code>`+render.Escape(path)+`</code></div>`))
	}
	return render.Join(rows...)
}
