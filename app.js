const iconPaths = {
  search: '<circle cx="11" cy="11" r="6.5"></circle><path d="m16 16 4.5 4.5"></path>',
  github: '<path d="M15 22v-4a4.8 4.8 0 0 0-1-3.5c3.3-.4 6.8-1.6 6.8-7a5.5 5.5 0 0 0-1.5-3.8A5.1 5.1 0 0 0 19.2 2S18 1.6 15 3.4a13.3 13.3 0 0 0-7 0C5 1.6 3.8 2 3.8 2a5.1 5.1 0 0 0-.1 3.7 5.5 5.5 0 0 0-1.5 3.8c0 5.4 3.5 6.6 6.8 7A4.8 4.8 0 0 0 8 18v4"></path><path d="M8 18c-3 .9-3-1.4-4.2-1.8"></path>',
  sun: '<circle cx="12" cy="12" r="4"></circle><path d="M12 2v2M12 20v2M4.93 4.93l1.41 1.41M17.66 17.66l1.41 1.41M2 12h2M20 12h2M4.93 19.07l1.41-1.41M17.66 6.34l1.41-1.41"></path>',
  moon: '<path d="M21 12.8A8.5 8.5 0 1 1 11.2 3 6.5 6.5 0 0 0 21 12.8Z"></path>',
  menu: '<path d="M4 6h16M4 12h16M4 18h16"></path>',
  chevron: '<path d="m6 9 6 6 6-6"></path>',
  arrow: '<path d="M5 12h14M13 6l6 6-6 6"></path>',
  searchDoc: '<path d="M4 4h16v16H4z"></path><path d="M8 8h8M8 12h5M8 16h3"></path>',
  book: '<path d="M4 19.5A2.5 2.5 0 0 1 6.5 17H20"></path><path d="M6.5 2H20v20H6.5A2.5 2.5 0 0 1 4 19.5v-15A2.5 2.5 0 0 1 6.5 2Z"></path>',
  code: '<path d="m8 9-4 3 4 3M16 9l4 3-4 3M14 5l-4 14"></path>',
  screen: '<rect x="3" y="4" width="18" height="16" rx="2"></rect><path d="M8 8h8M8 12h5M8 16h7"></path>',
  plug: '<path d="M12 22v-5M8 8V2M16 8V2M6 8h12v3a6 6 0 0 1-12 0V8Z"></path>',
  cloud: '<path d="M17.5 19H9a7 7 0 1 1 6.7-9h1.8a4 4 0 1 1 0 8Z"></path>',
  pencil: '<path d="m12 20 8-8-4-4-8 8-1 5 5-1Z"></path><path d="m14 6 4 4M4 20h4"></path>',
  copy: '<rect x="9" y="9" width="11" height="11" rx="2"></rect><path d="M5 15H4a2 2 0 0 1-2-2V4a2 2 0 0 1 2-2h9a2 2 0 0 1 2 2v1"></path>',
  check: '<path d="m5 12 4 4L19 6"></path>',
  zap: '<path d="m13 2-9 12h7l-1 8 9-12h-7l1-8Z"></path>',
  shield: '<path d="M12 22s8-4 8-10V5l-8-3-8 3v7c0 6 8 10 8 10Z"></path><path d="m9 12 2 2 4-4"></path>',
  agent: '<circle cx="12" cy="12" r="8"></circle><path d="M8 12h8M12 8v8M5 5l2 2M19 5l-2 2"></path>',
};

const icon = (name, size = 14) => `<svg width="${size}" height="${size}" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.7" stroke-linecap="round" stroke-linejoin="round" aria-hidden="true">${iconPaths[name] || iconPaths.book}</svg>`;
const siteName = 'starter-docs';

const routes = [
  { id: 'home', path: '/', title: 'Home', kind: 'Page', icon: 'screen', description: 'A router-first documentation project.', group: 'Project' },
  { id: 'docs', path: '/docs', title: 'Docs index', kind: 'Page', icon: 'book', description: 'The generated documentation hub.', group: 'Documentation' },
  { id: 'getting-started', path: '/docs/getting-started', title: 'Getting started', kind: 'Page', icon: 'book', description: 'Create your first Page or Screen route.', group: 'Documentation', nested: true },
  { id: 'openapi', path: '/api-reference', title: 'API reference', kind: 'Screen', icon: 'plug', description: 'Generated endpoint reference from openapi.json.', group: 'Reference' },
];

const code = {
  router: `<span class="line"><span class="ln">01</span><span class="kw">site</span> := docs.<span class="fn">NewRouter</span>(docs.Config{</span><span class="line"><span class="ln">02</span>  Name: <span class="str">"starter-docs"</span>,</span><span class="line"><span class="ln">03</span>  PWA: docs.PWAConfig{OfflineStaticContent: <span class="kw">true</span>},</span><span class="line"><span class="ln">04</span>})</span><span class="line"><span class="ln">05</span></span><span class="line"><span class="ln">06</span>site.<span class="fn">Group</span>(<span class="str">"/docs"</span>, docs.GroupConfig{</span><span class="line"><span class="ln">07</span>  Title: <span class="str">"Documentation"</span>,</span><span class="line"><span class="ln">08</span>}, <span class="kw">func</span>(r *docs.Router) {</span><span class="line"><span class="ln">09</span>  r.<span class="fn">Page</span>(<span class="str">"getting-started"</span>, docs.PageConfig{...})</span><span class="line"><span class="ln">10</span>})</span>`,
  page: `<span class="line"><span class="ln">01</span>r.<span class="fn">Page</span>(<span class="str">"getting-started"</span>, docs.PageConfig{</span><span class="line"><span class="ln">02</span>  Title: <span class="str">"Getting started"</span>,</span><span class="line"><span class="ln">03</span>  Body: docs.<span class="fn">MarkdownFile</span>(</span><span class="line"><span class="ln">04</span>    <span class="str">"content/getting-started.md"</span>,</span><span class="line"><span class="ln">05</span>  ),</span><span class="line"><span class="ln">06</span>  Offline: docs.Static,</span><span class="line"><span class="ln">07</span>})</span>`,
  screen: `<span class="line"><span class="ln">01</span>r.<span class="fn">Screen</span>(<span class="str">"playground"</span>, docs.ScreenConfig{</span><span class="line"><span class="ln">02</span>  Title: <span class="str">"Playground"</span>,</span><span class="line"><span class="ln">03</span>  Screen: &amp;PlaygroundScreen{},</span><span class="line"><span class="ln">04</span>  Offline: docs.NetworkRequired,</span><span class="line"><span class="ln">05</span>})</span>`,
  openapi: `<span class="line"><span class="ln">01</span>site.<span class="fn">Use</span>(openapi.<span class="fn">Plugin</span>(openapi.Config{</span><span class="line"><span class="ln">02</span>  Spec: <span class="str">"openapi.json"</span>,</span><span class="line"><span class="ln">03</span>  Route: <span class="str">"/api-reference"</span>,</span><span class="line"><span class="ln">04</span>}))</span>`
};

function routeHref(id) {
  const route = routes.find(item => item.id === id);
  return route?.href || `#${route?.path || '/'}`;
}

function renderRouteTree(activeId) {
  document.querySelector('[data-route-tree]').innerHTML = `<div class="route-tree__group"><div class="route-tree__group-label">Home</div>${routeLink('home', activeId)}</div><div class="route-tree__group"><div class="route-tree__group-label">Docs</div>${routeLink('docs', activeId)}${routeLink('getting-started', activeId)}</div><div class="route-tree__group"><div class="route-tree__group-label">Reference</div>${routeLink('openapi', activeId)}</div>`;
}

function routeLink(id, activeId) {
  const route = routes.find(item => item.id === id);
  return `<a class="route-tree__link ${route.nested ? 'route-tree__link--nested' : ''} ${id === activeId ? 'route-tree__link--active' : ''}" href="${routeHref(id)}" data-route="${id}"><span class="route-tree__icon">${icon(route.icon, 13)}</span><span>${route.title}</span><span class="route-tree__path">${route.path}</span></a>`;
}

function pageFrame({ section, title, lede, tag, time = '5 min read', body, home = false }) {
  return `<div class="view ${home ? 'view--home' : ''}"><div class="route-header"><div class="crumbs"><span>starter-docs</span><span>/</span><span>${section}</span></div><span class="eyebrow">${tag}</span><h1>${title}</h1><p class="route-header__lede">${lede}</p><div class="route-meta"><span class="tag">${tag}</span><span>${time}</span><span>·</span><span>static-ready</span></div></div><div class="body">${body}</div></div>`;
}

function renderHome() {
  return pageFrame({
    section: 'project',
    title: 'A docs site that starts as a router.',
    lede: 'This starter gives every project a single, inspectable composition point for pages, screens, navigation, search, API reference, and offline content.',
    tag: 'Project overview',
    time: '2 min read',
    home: true,
    body: `<div class="hero-grid"><div><div class="button-row"><a class="button button--primary" href="#/docs">Open the docs ${icon('arrow', 13)}</a><button class="button button--ghost" type="button" data-scroll-to="get-router">See the router <span>↓</span></button></div></div><div class="hero-mark" aria-hidden="true"><div class="hero-mark__grid"></div><div class="hero-mark__pin"></div><span class="hero-mark__label"><strong>ROUTER / 001</strong> · local preview</span></div></div><div class="metrics"><div class="metric"><span class="metric__value">03</span><span class="metric__label">starter routes</span></div><div class="metric"><span class="metric__value">02</span><span class="metric__label">render modes</span></div><div class="metric"><span class="metric__value">01</span><span class="metric__label">offline shell</span></div></div><div class="section-intro" id="get-router"><h2>One router, every surface</h2><p>The router is the product. Pages and screens are the rendering strategies it composes.</p></div><div class="feature-grid"><div class="feature"><span class="feature__icon">${icon('book', 15)}</span><strong class="feature__title">Page routes</strong><p class="feature__copy">Static Markdown pages with explicit order, metadata, search, and offline export.</p></div><div class="feature"><span class="feature__icon">${icon('screen', 15)}</span><strong class="feature__title">Screen routes</strong><p class="feature__copy">Typed GoFastr screens for API explorers, playgrounds, and custom experiences.</p></div><div class="feature"><span class="feature__icon">${icon('plug', 15)}</span><strong class="feature__title">Project plugins</strong><p class="feature__copy">OpenAPI, changelog, and search plugins extend this same router.</p></div></div><div class="section-intro"><h2>Generated project anatomy</h2><p>Start small. Keep every decision visible in the generated project.</p></div><div class="router-preview"><div class="panel"><div class="panel__head"><strong>registered routes</strong><span class="panel__right">3 routes</span></div><div class="panel__body"><div class="route-mini route-mini--selected"><span class="route-mini__kind">page</span><span>Home</span><span class="route-mini__path">/</span></div><div class="route-mini"><span class="route-mini__kind">page</span><span>Docs index</span><span class="route-mini__path">/docs</span></div><div class="route-mini"><span class="route-mini__kind">page</span><span>Getting started</span><span class="route-mini__path">/docs/…</span></div></div></div><div class="panel panel--dark"><div class="panel__head"><strong>docs/router.go</strong><span class="panel__right">Go</span></div><pre class="panel__code">${code.router}</pre></div></div><div class="callout" id="get-agents"><span class="callout__icon">${icon('agent', 14)}</span><div><span class="callout__title">AI-ready from the first commit</span><p>The generated project includes <code>agents/claude.md</code>, Codex-compatible skills, and project references that explain how to author pages without guessing.</p></div></div>`
  });
}

function renderDocs() {
  return pageFrame({ section: 'docs', title: 'The project handbook.', lede: 'A small generated docs tree that demonstrates the authoring model, the router, and the extensions that make a real documentation product.', tag: 'Documentation index', time: '3 min read', body: `<div class="page-index"><a class="page-index__item" href="#/docs/getting-started"><span class="page-index__number">01</span><span><strong class="page-index__title">Getting started</strong><span class="page-index__desc">Create your router, add your first Page, and preview the project.</span></span><span class="page-index__arrow">${icon('arrow', 14)}</span></a><a class="page-index__item" href="#/docs#get-openapi"><span class="page-index__number">02</span><span><strong class="page-index__title">OpenAPI plugin</strong><span class="page-index__desc">Generate API reference pages inside the same project and navigation tree.</span></span><span class="page-index__arrow">${icon('arrow', 14)}</span></a><a class="page-index__item" href="#/docs#get-agents"><span class="page-index__number">03</span><span><strong class="page-index__title">AI project references</strong><span class="page-index__desc">Teach agents the route conventions, checks, and authoring workflow.</span></span><span class="page-index__arrow">${icon('arrow', 14)}</span></a></div><div class="section-intro" id="get-openapi"><h2>OpenAPI, in the same router</h2><p>The API plugin is part of the fastr-docs project, but remains optional for generic docs sites.</p></div><div class="plugin-banner"><span class="plugin-banner__icon">${icon('plug', 16)}</span><div><strong>OpenAPI plugin installed</strong><p>It reads <code>openapi.json</code>, registers an API reference screen, and contributes endpoint pages to the local search index.</p></div></div><div class="code-card"><div class="code-card__bar"><span class="code-dots"><i></i><i></i><i></i></span><span class="code-card__file">docs/router.go</span><button type="button" class="copy-button" data-copy="openapi-code">${icon('copy', 11)} Copy</button></div><pre id="openapi-code">${code.openapi}</pre></div><div class="section-intro" id="get-page"><h2>Two rendering strategies</h2><p>Use the smallest strategy that fits the page, while keeping both routes in one tree.</p></div><div class="router-preview"><div class="panel panel--dark"><div class="panel__head"><strong>Page</strong><span class="panel__right">static / offline</span></div><pre class="panel__code">${code.page}</pre></div><div class="panel panel--dark"><div class="panel__head"><strong id="get-screen">Screen</strong><span class="panel__right">typed / interactive</span></div><pre class="panel__code">${code.screen}</pre></div></div><div class="article-footer"><div class="feedback"><span>Useful page?</span><button type="button" data-feedback="yes" aria-label="Yes">${icon('check', 13)}</button><button type="button" data-feedback="no" aria-label="No">${icon('pencil', 12)}</button></div><span class="tag tag--sage">offline-ready</span></div><div class="pager"><a class="pager__item pager__item--next" href="#/docs/getting-started"><span class="pager__direction">Next chapter →</span><span class="pager__title">Getting started</span></a></div>` });
}

function renderGettingStarted() {
  return pageFrame({ section: 'docs / getting started', title: 'Create your first route.', lede: 'The generated starter project is intentionally small: one router, two top-level pages, and one nested page to make navigation and ordering visible immediately.', tag: 'Tutorial', time: '7 min read', body: `<p>When you run <code>fastr-docs init</code>, the generator creates a project that is useful on its first boot. You can read the route tree, click through the navigation, and then replace the sample content with your own.</p><div class="callout"><span class="callout__icon">${icon('zap', 14)}</span><div><span class="callout__title">The source of truth is the router</span><p>Do not maintain a separate sidebar file. Route registration creates the URL, navigation item, search record, breadcrumbs, and export manifest together.</p></div></div><h2 id="create-router">Create the router</h2><p>The starter exposes a single composition point in <code>docs/router.go</code>. Sections give the tree shape; explicit order keeps it stable as the project grows.</p><div class="code-card"><div class="tabs"><button class="tab tab--active" type="button" data-tab="router-tab">router.go</button><button class="tab" type="button" data-tab="config-tab">fastr-docs.yaml</button></div><div class="tab-panel tab-panel--active" data-panel="router-tab"><pre>${code.router}</pre></div><div class="tab-panel" data-panel="config-tab"><pre><span class="line"><span class="ln">01</span><span class="prop">theme</span>: signal-atlas</span><span class="line"><span class="ln">02</span><span class="prop">default_mode</span>: light</span><span class="line"><span class="ln">03</span><span class="prop">search</span>:</span><span class="line"><span class="ln">04</span>  <span class="prop">provider</span>: local</span><span class="line"><span class="ln">05</span><span class="prop">checks</span>:</span><span class="line"><span class="ln">06</span>  <span class="prop">broken_links</span>: error</span></pre></div></div><h2 id="choose-rendering">Choose Page or Screen</h2><p>Most docs routes are Pages. Use a Screen when the page needs typed rendering, live data, or a richer GoFastr interaction.</p><div class="step-list"><div class="step"><div><strong>Start with a Page</strong><p>Point the route at a Markdown file and get static rendering, search, SEO, and offline caching immediately.</p></div></div><div class="step"><div><strong>Promote when needed</strong><p>Replace the body with a typed Screen without changing the route’s place in the tree.</p></div></div><div class="step"><div><strong>Run the checks</strong><p>Use <code>fastr-docs check</code> before publishing. Failures are errors unless explicitly opted out.</p></div></div></div><h2 id="agents">Use the included agent references</h2><p>The starter includes agent instructions so an AI can work safely in the project without reverse-engineering conventions.</p><div class="feature-grid"><div class="feature"><span class="feature__icon">${icon('agent', 15)}</span><strong class="feature__title">agents/claude.md</strong><p class="feature__copy">Claude-specific project orientation and safe edit workflow.</p></div><div class="feature"><span class="feature__icon">${icon('code', 15)}</span><strong class="feature__title">skills/docs-authoring</strong><p class="feature__copy">Reusable instructions for adding and reviewing documentation routes.</p></div><div class="feature"><span class="feature__icon">${icon('shield', 15)}</span><strong class="feature__title">Strict checks</strong><p class="feature__copy">Agents are told to run validation before considering work complete.</p></div></div><div class="article-footer"><div class="feedback"><span>Useful page?</span><button type="button" data-feedback="yes" aria-label="Yes">${icon('check', 13)}</button><button type="button" data-feedback="no" aria-label="No">${icon('pencil', 12)}</button></div><span class="tag tag--sage">static + offline</span></div><div class="pager"><a class="pager__item" href="#/docs"><span class="pager__direction">← Previous</span><span class="pager__title">Docs index</span></a><a class="pager__item pager__item--next" href="#/"><span class="pager__direction">Back to →</span><span class="pager__title">Project home</span></a></div>` });
}

function renderOpenAPI() {
  return `<div class="view view--api"><div class="route-header"><div class="crumbs"><span>${siteName}</span><span>/</span><span>reference</span><span>/</span><span>api</span></div><span class="eyebrow">OpenAPI plugin</span><h1>API reference.</h1><p class="route-header__lede">A generated contract, made readable. Browse operations, inspect schemas, and try a request without leaving the documentation.</p><div class="route-meta"><span class="tag tag--blue">OpenAPI 3.1</span><span>12 operations</span><span>·</span><span>3 schemas</span><span>·</span><span>updated 4 min ago</span></div></div><div class="api-toolbar"><div><strong>Public API</strong><span>Generated from openapi.json</span></div><label class="api-filter-wrap"><span data-icon="search"></span><input class="api-filter" type="search" placeholder="Filter endpoints…" aria-label="Filter endpoints"></label></div><div class="api-reference__layout"><aside class="api-index"><div class="api-index__head"><span>Endpoints</span><span>12</span></div><div class="api-index__group">Projects</div><a href="#api-list-projects" data-api-operation-link><span class="method-badge method-badge--get">GET</span><code>/v1/projects</code></a><a href="#api-create-project" data-api-operation-link><span class="method-badge method-badge--post">POST</span><code>/v1/projects</code></a><a href="#api-project-detail" data-api-operation-link><span class="method-badge method-badge--get">GET</span><code>/v1/projects/{id}</code></a><div class="api-index__group">Accounts</div><a href="#api-accounts" data-api-operation-link><span class="method-badge method-badge--get">GET</span><code>/v1/account</code></a><div class="api-index__group">Models</div><a href="#api-model-project" data-api-operation-link><span class="api-index__model">{ }</span><span>Project</span></a><a href="#api-model-error" data-api-operation-link><span class="api-index__model">{ }</span><span>Error</span></a></aside><section class="api-content"><div class="api-content__intro"><span class="tag tag--sage">stable</span><span>Base URL</span><code>https://api.example.com</code></div><article class="api-operation api-operation--open is-expanded" id="api-list-projects" data-api-operation data-api-search="get list projects v1 projects"><div class="api-operation__top"><div class="api-operation__route"><span class="method-badge method-badge--get">GET</span><code>/v1/projects</code></div><span class="status-code status-code--success">200</span><button class="api-expand" type="button" data-api-expand>Collapse ${icon('chevron', 13)}</button></div><p class="api-operation__summary">List projects visible to the authenticated account.</p><div class="api-operation__details"><div class="api-columns"><div><h4>Query parameters</h4><div class="api-table"><div><code>cursor</code><span>string · optional</span></div><div><code>limit</code><span>integer · default 20</span></div><div><code>status</code><span>string · optional</span></div></div></div><div><h4>Response</h4><pre class="api-json"><span class="json-brace">{</span><span class="json-line"><span class="json-key">"data"</span>: [</span><span class="json-line indent"><span class="json-brace">{</span> <span class="json-key">"id"</span>: <span class="json-string">"prj_01"</span>,</span><span class="json-line indent"><span class="json-key">"name"</span>: <span class="json-string">"Northstar"</span> <span class="json-brace">}</span></span><span class="json-line">  ],</span><span class="json-line"><span class="json-key">"next_cursor"</span>: <span class="json-null">null</span></span><span class="json-brace">}</span></pre></div></div><div class="code-card api-code-card"><div class="tabs"><button class="tab tab--active" type="button" data-tab="curl-tab">cURL</button><button class="tab" type="button" data-tab="go-tab">Go</button><button class="tab" type="button" data-tab="js-tab">JavaScript</button></div><div class="tab-panel tab-panel--active" data-panel="curl-tab"><pre><span class="line"><span class="ln">01</span>curl https://api.example.com/v1/projects \</span><span class="line"><span class="ln">02</span>  -H <span class="str">"Authorization: Bearer $TOKEN"</span></span></pre></div><div class="tab-panel" data-panel="go-tab"><pre><span class="line"><span class="ln">01</span>projects, err := client.Projects.<span class="fn">List</span>(ctx)</span></pre></div><div class="tab-panel" data-panel="js-tab"><pre><span class="line"><span class="ln">01</span><span class="kw">const</span> projects = <span class="kw">await</span> client.projects.<span class="fn">list</span>()</span></pre></div></div></div></article><article class="api-operation" id="api-create-project" data-api-operation data-api-search="post create projects v1 projects"><div class="api-operation__top"><div class="api-operation__route"><span class="method-badge method-badge--post">POST</span><code>/v1/projects</code></div><span class="status-code status-code--success">201</span><button class="api-expand" type="button" data-api-expand>Expand ${icon('chevron', 13)}</button></div><p class="api-operation__summary">Create a project for the authenticated account.</p><div class="api-operation__details"><div class="api-request-row"><span>Request body</span><code>ProjectCreate</code><span class="api-required">required</span></div><pre class="api-json"><span class="json-brace">{</span><span class="json-line"><span class="json-key">"name"</span>: <span class="json-string">"Northstar"</span>,</span><span class="json-line"><span class="json-key">"slug"</span>: <span class="json-string">"northstar"</span></span><span class="json-brace">}</span></pre></div></article><article class="api-operation" id="api-project-detail" data-api-operation data-api-search="get project detail id v1 projects"><div class="api-operation__top"><div class="api-operation__route"><span class="method-badge method-badge--get">GET</span><code>/v1/projects/{id}</code></div><span class="status-code status-code--success">200</span><button class="api-expand" type="button" data-api-expand>Expand ${icon('chevron', 13)}</button></div><p class="api-operation__summary">Fetch one project by ID.</p><div class="api-operation__details"><p>Returns a single <code>Project</code> resource or <code>404 not_found</code>.</p></div></article><section class="api-model" id="api-model-project"><div class="api-model__head"><span class="api-index__model">{ }</span><div><span class="tag tag--blue">schema</span><h2>Project</h2></div><code>object</code></div><div class="schema-table"><div><code>id</code><span>string</span><em>read-only</em></div><div><code>name</code><span>string</span><em>required</em></div><div><code>slug</code><span>string</span><em>required</em></div><div><code>created_at</code><span>date-time</span><em>read-only</em></div></div></section></section><aside class="api-console"><div class="api-console__head"><div><span class="tag tag--orange">Live console</span><strong>Try a request</strong></div><span class="api-console__lock">${icon('shield', 13)}</span></div><label>Endpoint<select><option>GET · /v1/projects</option><option>POST · /v1/projects</option></select></label><label>Authorization<div class="api-input"><span>Bearer</span><input value="demo_token" aria-label="Demo token"></div></label><label>Query parameters<div class="api-input"><span>limit</span><input value="20" aria-label="Limit"></div></label><button class="button button--primary api-send" type="button" data-try-request>Send request ${icon('arrow', 13)}</button><div class="api-response" data-api-response><div class="api-response__head"><span>Response</span><span>ready</span></div><p>Send a request to see a simulated response.</p></div><p class="api-console__note">Requests run against the configured API origin. This static preview uses a local response fixture.</p></aside></div></div>`;
}

const pageRenderers = { home: renderHome, docs: renderDocs, 'getting-started': renderGettingStarted, openapi: renderOpenAPI };

function currentRoute() {
  const path = window.location.hash.replace(/^#/, '') || '/';
  const routePath = path.split('#')[0] || '/';
  if (path.includes('#get-openapi') || path === '/api-reference') return 'openapi';
  if (routePath === '/api-reference') return 'openapi';
  return routes.find(route => route.path === routePath)?.id || (routePath.startsWith('/docs/getting-started') ? 'getting-started' : routePath.startsWith('/docs') ? 'docs' : 'home');
}

function render() {
  const id = currentRoute();
  document.querySelector('[data-view]').innerHTML = pageRenderers[id]();
  document.querySelectorAll('[data-icon]').forEach(element => { if (!element.innerHTML) element.innerHTML = icon(element.dataset.icon); });
  if (id === 'openapi') ensureOpenAPIPrototypeDetails();
  if (id === 'docs') {
    const apiLink = document.querySelector('.page-index__item:nth-child(2)');
    if (apiLink) {
      apiLink.href = routeHref('openapi');
      apiLink.querySelector('.page-index__title').textContent = 'OpenAPI reference';
      apiLink.querySelector('.page-index__desc').textContent = 'Explore generated operations and schemas inside the same project.';
    }
  }
  const active = routes.find(route => route.id === id);
  renderRouteTree(id);
  document.querySelector('[data-mobile-route]').textContent = active.title;
  document.title = `${active.title} — ${siteName}`;
  bindPageActions();
  renderMobileToc();
  requestAnimationFrame(() => {
    const hashPath = window.location.hash.replace(/^#/, '');
    const marker = hashPath.indexOf('#');
    const anchor = marker >= 0 ? hashPath.slice(marker + 1) : '';
    if (anchor) setTimeout(() => document.getElementById(anchor)?.scrollIntoView({ behavior: 'smooth', block: 'start' }), 0);
    else window.scrollTo({ top: 0, behavior: 'instant' });
    updateRail();
  });
}

function ensureOpenAPIPrototypeDetails() {
  const content = document.querySelector('.api-content');
  if (!content) return;
  if (!document.getElementById('api-accounts')) {
    const account = document.createElement('article');
    account.className = 'api-operation';
    account.id = 'api-accounts';
    account.dataset.apiOperation = 'true';
    account.dataset.apiSearch = 'get account v1 account';
    account.innerHTML = `<div class="api-operation__top"><div class="api-operation__route"><span class="method-badge method-badge--get">GET</span><code>/v1/account</code></div><span class="status-code status-code--success">200</span><button class="api-expand" type="button" data-api-expand>Expand ${icon('chevron', 13)}</button></div><p class="api-operation__summary">Fetch the authenticated account.</p><div class="api-operation__details"><p>Returns the account associated with the bearer token.</p></div>`;
    const firstModel = content.querySelector('.api-model');
    if (firstModel) firstModel.before(account); else content.append(account);
  }
  if (!document.getElementById('api-model-error')) {
    const error = document.createElement('section');
    error.className = 'api-model';
    error.id = 'api-model-error';
    error.innerHTML = `<div class="api-model__head"><span class="api-index__model">{ }</span><div><span class="tag tag--blue">schema</span><h2>Error</h2></div><code>object</code></div><div class="schema-table"><div><code>code</code><span>string</span><em>required</em></div><div><code>message</code><span>string</span><em>required</em></div></div>`;
    content.append(error);
  }
}

function renderMobileToc() {
  const toc = document.querySelector('[data-mobile-toc]');
  const links = document.querySelector('[data-mobile-toc-links]');
  const count = document.querySelector('[data-mobile-toc-count]');
  if (window.__mobileTocObserver) window.__mobileTocObserver.disconnect();
  const headings = [...document.querySelectorAll('[data-view] h2[id], [data-view] h3[id]')];
  if (!headings.length) {
    toc.hidden = true;
    return;
  }
  toc.hidden = false;
  count.textContent = `${headings.length} sections`;
  const anchorBase = routeHref(currentRoute());
  links.innerHTML = headings.map(heading => `<a href="${anchorBase}#${heading.id}" data-mobile-toc-link="${heading.id}" class="${heading.tagName === 'H3' ? 'mobile-toc__subitem' : ''}">${heading.textContent}</a>`).join('');
  links.querySelectorAll('a').forEach(link => link.addEventListener('click', () => { toc.open = false; }));
  const tocLinks = [...links.querySelectorAll('[data-mobile-toc-link]')];
  window.__mobileTocObserver = new IntersectionObserver(entries => entries.forEach(entry => {
    if (entry.isIntersecting) tocLinks.forEach(link => link.classList.toggle('is-active', link.dataset.mobileTocLink === entry.target.id));
  }), { rootMargin: '-14% 0px -70% 0px' });
  headings.forEach(heading => window.__mobileTocObserver.observe(heading));
}

function bindPageActions() {
  document.querySelectorAll('[data-copy]').forEach(button => button.addEventListener('click', async () => {
    const element = document.getElementById(button.dataset.copy);
    const text = element?.innerText.replace(/^\d+\s+/gm, '') || '';
    try { await navigator.clipboard.writeText(text); } catch { /* preview may not have clipboard permission */ }
    button.innerHTML = `${icon('check', 11)} Copied`;
    showToast('Code copied to clipboard');
    setTimeout(() => { button.innerHTML = `${icon('copy', 11)} Copy`; }, 1400);
  }));
  document.querySelectorAll('[data-tab]').forEach(tab => tab.addEventListener('click', () => {
    const parent = tab.closest('.code-card');
    parent.querySelectorAll('[data-tab]').forEach(item => item.classList.toggle('tab--active', item === tab));
    parent.querySelectorAll('[data-panel]').forEach(panel => panel.classList.toggle('tab-panel--active', panel.dataset.panel === tab.dataset.tab));
  }));
  document.querySelectorAll('[data-feedback]').forEach(button => button.addEventListener('click', () => { button.parentElement.querySelectorAll('button').forEach(item => item.classList.remove('is-selected')); button.classList.add('is-selected'); showToast('Thanks for the feedback'); }));
  document.querySelectorAll('[data-scroll-to]').forEach(button => button.addEventListener('click', () => document.getElementById(button.dataset.scrollTo)?.scrollIntoView({ behavior: 'smooth' })));
  document.querySelectorAll('[data-api-expand]').forEach(button => button.addEventListener('click', () => {
    const operation = button.closest('[data-api-operation]');
    const open = operation.classList.toggle('is-expanded');
    button.innerHTML = `${open ? 'Collapse' : 'Expand'} ${icon('chevron', 13)}`;
  }));
  const apiFilter = document.querySelector('.api-filter');
  if (apiFilter) apiFilter.addEventListener('input', event => {
    const query = event.target.value.trim().toLowerCase();
    document.querySelectorAll('[data-api-operation]').forEach(operation => operation.hidden = query !== '' && !operation.dataset.apiSearch.includes(query));
  });
  document.querySelectorAll('[data-api-operation-link]').forEach(link => link.addEventListener('click', event => {
    event.preventDefault();
    document.querySelectorAll('[data-api-operation-link]').forEach(item => item.classList.remove('is-active'));
    link.classList.add('is-active');
    const targetId = link.getAttribute('href')?.slice(1);
    const target = targetId ? document.getElementById(targetId) : null;
    if (target) {
      history.replaceState(null, '', `${routeHref('openapi')}#${targetId}`);
      target.scrollIntoView({ behavior: 'smooth', block: 'start' });
    }
  }));
  const tryRequest = document.querySelector('[data-try-request]');
  if (tryRequest) tryRequest.addEventListener('click', () => {
    const response = document.querySelector('[data-api-response]');
    response.innerHTML = `<div class="api-response__head"><span>Response</span><span class="api-response__success">200 · 128ms</span></div><pre>{ <span class="json-key">"data"</span>: [<span class="json-string">"prj_01"</span>, <span class="json-string">"prj_02"</span>] }</pre>`;
    showToast('Request simulated locally');
  });
}

function updateRail() {
  if (window.__railObserver) window.__railObserver.disconnect();
  const links = [...document.querySelectorAll('[data-rail-link]')];
  const targets = links.map(link => document.getElementById(link.dataset.railLink)).filter(Boolean);
  if (!targets.length) return;
  window.__railObserver = new IntersectionObserver(entries => entries.forEach(entry => { if (entry.isIntersecting) links.forEach(link => link.classList.toggle('is-active', link.dataset.railLink === entry.target.id)); }), { rootMargin: '-12% 0px -72% 0px' });
  targets.forEach(target => window.__railObserver.observe(target));
}

function setupSearch() {
  const overlay = document.querySelector('[data-search-overlay]');
  const input = document.querySelector('[data-search-input]');
  const results = document.querySelector('[data-search-results]');
  let selected = -1;
  const all = [...routes, { id: 'agents', path: '#get-agents', title: 'AI project references', kind: 'Guide', icon: 'agent', description: 'Claude, Codex, and docs-authoring instructions.', group: 'Project' }];
  const renderResults = query => {
    const q = query.trim().toLowerCase();
    const matches = q ? all.filter(item => `${item.title} ${item.description} ${item.group} ${item.path}`.toLowerCase().includes(q)).slice(0, 8) : all.slice(0, 5);
    results.innerHTML = matches.length ? matches.map(item => `<div class="search-result" data-search-id="${item.id}" tabindex="0"><span class="search-result__icon">${icon(item.icon, 14)}</span><div><div class="search-result__title">${item.title}</div><div class="search-result__meta">${item.group} · ${item.description}</div></div></div>`).join('') : '<p class="search-placeholder">No matching route or reference.</p>';
    selected = -1;
    results.querySelectorAll('[data-search-id]').forEach(item => item.addEventListener('click', () => { const target = all.find(route => route.id === item.dataset.searchId); if (target.id === 'agents') window.location.hash = '#/#get-agents'; else window.location.hash = routeHref(target.id); close(); }));
  };
  const open = () => { overlay.hidden = false; document.body.classList.add('is-searching'); input.value = ''; renderResults(''); setTimeout(() => input.focus(), 20); };
  const close = () => { overlay.hidden = true; document.body.classList.remove('is-searching'); };
  document.querySelector('[data-open-search]').addEventListener('click', open);
  document.querySelector('[data-close-search]').addEventListener('click', close);
  input.addEventListener('input', event => renderResults(event.target.value));
  input.addEventListener('keydown', event => { const items = [...results.querySelectorAll('[data-search-id]')]; if (event.key === 'Escape') close(); if (event.key === 'ArrowDown' && items.length) { event.preventDefault(); selected = (selected + 1) % items.length; items.forEach((item, index) => item.classList.toggle('is-focused', index === selected)); } if (event.key === 'ArrowUp' && items.length) { event.preventDefault(); selected = (selected - 1 + items.length) % items.length; items.forEach((item, index) => item.classList.toggle('is-focused', index === selected)); } if (event.key === 'Enter' && items[selected]) items[selected].click(); });
  document.addEventListener('keydown', event => { if ((event.metaKey || event.ctrlKey) && event.key.toLowerCase() === 'k') { event.preventDefault(); overlay.hidden ? open() : close(); } if (event.key === 'Escape' && !overlay.hidden) close(); });
}

function setupTheme() {
  const button = document.querySelector('[data-theme-toggle]');
  const stored = localStorage.getItem('fastr-docs-theme');
  if (stored) document.documentElement.dataset.theme = stored;
  const sync = () => { const dark = document.documentElement.dataset.theme === 'dark'; button.innerHTML = icon(dark ? 'moon' : 'sun', 14); button.setAttribute('aria-label', dark ? 'Switch to light theme' : 'Switch to dark theme'); };
  sync();
  button.addEventListener('click', () => { const dark = document.documentElement.dataset.theme === 'dark'; document.documentElement.dataset.theme = dark ? 'light' : 'dark'; localStorage.setItem('fastr-docs-theme', dark ? 'light' : 'dark'); sync(); });
}

function setupMobileNav() {
  const sidebar = document.querySelector('[data-sidebar]');
  const toggle = document.querySelector('[data-mobile-toggle]');
  const open = () => { sidebar.classList.add('is-open'); toggle.setAttribute('aria-expanded', 'true'); };
  const close = () => { sidebar.classList.remove('is-open'); toggle.setAttribute('aria-expanded', 'false'); };
  toggle.addEventListener('click', () => sidebar.classList.contains('is-open') ? close() : open());
  document.querySelector('[data-open-drawer]').addEventListener('click', open);
  document.querySelector('[data-route-tree]').addEventListener('click', event => { if (event.target.closest('a')) close(); });
}

function showToast(message) { const toast = document.querySelector('[data-toast]'); toast.textContent = message; toast.classList.add('is-visible'); clearTimeout(window.__toast); window.__toast = setTimeout(() => toast.classList.remove('is-visible'), 1700); }

document.querySelectorAll('[data-icon]').forEach(element => { if (!element.innerHTML) element.innerHTML = icon(element.dataset.icon); });
window.addEventListener('hashchange', render);
setupTheme();
setupSearch();
setupMobileNav();
render();
