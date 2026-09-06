import { test, expect } from '@playwright/test';
import { runtime } from '../support/runtime.mjs';

const selfPage = (path = '/') => `${runtime().selfURL}${path}`;
const isMobileProject = (testInfo) => testInfo.project.name === 'mobile-chromium';

const openPrimaryNav = async (page, testInfo) => {
  const nav = page.locator(isMobileProject(testInfo) ? 'nav.ui-site-header__mobile-links' : 'nav.ui-site-header__links');
  if (isMobileProject(testInfo) && await nav.isHidden()) {
    await page.locator('summary[aria-label="Toggle navigation"]:visible').first().click();
  }
  await expect(nav).toBeVisible();
  return nav;
};

const openSectionNav = async (page, testInfo) => {
  const nav = page.locator(isMobileProject(testInfo) ? '[data-fui-widget="fastr-docs-sections"]' : '.ui-sidebar__inline');
  if (isMobileProject(testInfo) && await nav.isHidden()) {
    await page.locator('[data-fui-open="fastr-docs-sections"]:visible').first().click();
  }
  await expect(nav).toBeVisible();
  return nav;
};

// Each blog collection has its own drawer; the English one keeps the short
// name, a second collection's is derived from its prefix.
const openBlogNav = async (page, testInfo, drawer = 'fastr-docs-blog-sections') => {
  const nav = page.locator(isMobileProject(testInfo) ? `[data-fui-widget="${drawer}"]` : '[data-fui-layout^="blog"] .ui-sidebar__inline');
  if (isMobileProject(testInfo) && await nav.isHidden()) {
    await page.locator(`[data-fui-open="${drawer}"]:visible`).first().click();
  }
  await expect(nav).toBeVisible();
  return nav;
};

test('self-hosted docs connect the landing page, route groups, and command palette', async ({ page }, testInfo) => {
  await page.goto(selfPage('/'));
  await expect(page.getByRole('heading', { name: 'Build docs around one route tree.', exact: true })).toBeVisible();

  let headerNav = null;
  if (isMobileProject(testInfo)) {
    await page.getByRole('link', { name: 'Read the docs', exact: true }).click();
  } else {
    headerNav = await openPrimaryNav(page, testInfo);
    await headerNav.getByRole('link', { name: 'Documentation', exact: true }).click();
  }
  await expect(page).toHaveURL(/\/docs\/?$/);
  await expect(page.getByRole('heading', { name: 'Documentation', exact: true })).toBeVisible();
  const sectionNav = await openSectionNav(page, testInfo);
  const documentationGroup = sectionNav.locator('details.ui-sidebar__group').filter({ hasText: /^Documentation/ });
  await expect(documentationGroup).toHaveAttribute('open', '');

  const buildGroup = sectionNav.locator('summary.ui-sidebar__link').filter({ hasText: /^Build$/ }).locator('xpath=..');
  await buildGroup.locator('summary.ui-sidebar__link').click();
  await buildGroup.locator('a.ui-sidebar__link[href="/docs/build/framework-ui"]').click();
  await expect(page).toHaveURL(/\/docs\/build\/framework-ui\/?$/);
  await expect(page.getByRole('heading', { name: 'Framework UI', exact: true })).toBeVisible();
  await expect(page.getByText('Site chrome', { exact: true })).toBeVisible();
  await expect(page.locator('code', { hasText: 'CommandPalette' }).first()).toBeVisible();
  for (const heading of ['Docs shell and navigation', 'Reading and technical content', 'Layout and composition', 'Forms and controls', 'Data, actions, and feedback', 'Data visualization and media']) {
    await expect(page.getByRole('heading', { name: heading, exact: true })).toBeVisible();
  }
  for (const component of ['AuthCard', 'DataTable', 'PipelineImage', 'ThemeToggle', 'ValidationSummary']) {
    await expect(page.getByText(component, { exact: true }).first()).toBeVisible();
  }

  if (isMobileProject(testInfo)) {
    await page.goto(selfPage('/'));
    const landingSidebar = await openSectionNav(page, testInfo);
    await landingSidebar.getByRole('link', { name: 'Example API reference', exact: true }).click();
  } else {
    await openPrimaryNav(page, testInfo);
    await headerNav.getByRole('link', { name: 'Example API reference', exact: true }).click();
  }
  await expect(page).toHaveURL(/\/api-reference\/?$/);
  await expect(page.getByRole('heading', { name: 'Example API reference', exact: true })).toBeVisible();

  await page.goto(selfPage('/docs/getting-started'));
  if (!isMobileProject(testInfo)) {
    const activeHeaderNav = await openPrimaryNav(page, testInfo);
    await expect(activeHeaderNav.getByRole('link', { name: 'Documentation', exact: true })).toHaveAttribute('aria-current', 'page');
  }
  const activeSectionNav = await openSectionNav(page, testInfo);
  await expect(activeSectionNav.locator('a.ui-sidebar__link[href="/docs/getting-started"]')).toHaveAttribute('aria-current', 'page');

  await page.keyboard.press('Control+K');
  const paletteInput = page.locator('#fastr-docs-command-palette-input:visible').first();
  await expect(paletteInput).toBeVisible();
  await paletteInput.fill('playground');
  await page.locator('[role="option"][data-fui-push-state="/examples/playground"]:visible').click();
  await expect(page).toHaveURL(/\/examples\/playground\/?$/);
});

test('self-hosted OpenAPI and typed screen surfaces behave as documented', async ({ page }) => {
  await page.goto(selfPage('/api-reference'));
  await expect(page.locator('[data-openapi-reference]')).toBeVisible();
  await expect(page.locator('[data-openapi-operation]').first()).toContainText('listProjects');
  await page.locator('[data-openapi-filter]').fill('projects');
  await expect(page.locator('[data-openapi-operation]').first()).toBeVisible();

  await page.goto(selfPage('/examples/playground'));
  await expect(page.locator('[data-testid="examples-playground"]')).toBeVisible();
  await page.getByText('Implementation', { exact: true }).click();
  await expect(page.getByText('Move the component into your product package when it becomes a real feature.', { exact: true })).toBeVisible();
  await page.getByText('Show the design rule', { exact: true }).click();
  await expect(page.getByText('Start with a route, give it an explicit order, and let the Router derive navigation and search metadata.', { exact: true })).toBeVisible();

  const counter = page.locator('.fui-counter').first();
  await expect(counter).toContainText('0');
  await counter.getByRole('button').last().click();
  await expect(counter).toContainText('1');
});

test('self-hosted navigation opens the active nested route without overflow', async ({ page }, testInfo) => {
  await page.goto(selfPage('/docs/concepts/layouts'));
  await expect(page.locator('html')).toHaveJSProperty('scrollWidth', await page.locator('html').evaluate((node) => node.clientWidth));

  const sidebar = await openSectionNav(page, testInfo);
  const current = sidebar.getByRole('link', { name: 'Layouts and navigation', exact: true });
  await expect(current).toHaveAttribute('aria-current', 'page');
  await expect(current).toHaveClass(/active/);

  const toc = page.locator('[data-docs-toc-select]');
  await page.evaluate(() => document.querySelector('#header-groups-are-not-duplicate-sidebar-items').scrollIntoView({ behavior: 'instant', block: 'start' }));
  await expect.poll(() => toc.inputValue()).toBe('#header-groups-are-not-duplicate-sidebar-items');
});

test('self-hosted blog archive and RSS feed share the Router content', async ({ page, request }, testInfo) => {
  await page.addInitScript(() => {
    Object.defineProperty(navigator, 'share', { configurable: true, writable: true, value: undefined });
    Object.defineProperty(navigator, 'clipboard', { configurable: true, writable: true, value: { writeText: async () => {} } });
  });
  await page.goto(selfPage('/blog'));
  await expect(page.locator('h1').filter({ hasText: 'Blog' })).toBeVisible();
  await expect(page.locator('[data-fui-layout^="blog"]')).toBeVisible();
  const blogNav = await openBlogNav(page, testInfo);
  await expect(blogNav.getByRole('link', { name: 'All posts', exact: true })).toHaveAttribute('aria-current', 'page');
  if (isMobileProject(testInfo)) {
    await page.keyboard.press('Escape');
  }
  await expect(page.locator('a.fastr-docs-blog-card[href="/blog/route-tree"]').first()).toBeVisible();
  await expect(page.getByRole('link', { name: 'Tags', exact: true }).first()).toBeVisible();
  await page.getByRole('link', { name: 'Search', exact: true }).first().click();
  await expect(page).toHaveURL(/\/blog\/search\/?$/);
  await page.getByRole('searchbox', { name: 'Search posts' }).fill('router');
  await page.getByRole('button', { name: 'Search', exact: true }).click();
  await expect(page).toHaveURL(/\/blog\/search\?q=router$/);
  await expect(page.getByText(/Results for “router”/)).toBeVisible();
  await page.goto(selfPage('/blog/tags/routing'));
  await expect(page.locator('h1').filter({ hasText: 'Tag: routing' })).toBeVisible();

  await page.goto(selfPage('/blog/route-tree'));
  const article = page.locator('article.fastr-docs-blog-post__article');
  await expect(article).toHaveCount(1);
  await expect(article.locator('h1')).toHaveAttribute('id', /-title$/);
  await expect(article.locator('time[datetime="2026-08-28"]')).toBeVisible();
  await expect(article.locator('[data-fastr-docs-share]')).toHaveAttribute('aria-label', 'Share this post');
  await expect(article.locator('[data-fui-copy-text-from]')).toHaveAttribute('aria-label', 'Copy link');
  await expect(article.locator('[data-fastr-docs-share-target]')).toContainText('/blog/route-tree');
  await expect(page.locator('.fastr-docs-blog-post__related')).toContainText('A publication surface of its own');
  await expect(article.locator('.fastr-docs-blog-post__tags a[href="/blog/tags/architecture"]')).toBeVisible();
  if (isMobileProject(testInfo)) {
    const toc = page.locator('.fastr-docs-blog-post__toc [data-docs-toc-select]');
    await expect(toc).toBeVisible();
    await toc.selectOption({ label: 'A better reading surface' });
    await expect(page).toHaveURL(/#a-better-reading-surface$/);
  } else {
    await expect(page.locator('.fastr-docs-blog-post__toc .fastr-docs-toc--rail')).toBeVisible();
  }
  await article.locator('[data-fastr-docs-share]').click();
  await expect(article.locator('[data-fastr-docs-share-status]')).toHaveText('Link copied');
  await page.evaluate(() => { navigator.share = async () => {}; });
  await article.locator('[data-fastr-docs-share]').click();
  await expect(article.locator('[data-fastr-docs-share-status]')).toHaveText('Post shared');
  await article.locator('[data-fui-copy-text-from]').click();
  await expect(article.locator('[data-fui-copy-status]')).toHaveText('Link copied');

  const post = await request.get(selfPage('/blog/route-tree'));
  expect(post.ok()).toBeTruthy();
  expect(await post.text()).toContain('One tree for docs and publishing');

  const feed = await request.get(selfPage('/blog/feed.xml'));
  expect(feed.ok()).toBeTruthy();
  expect(feed.headers()['content-type']).toContain('application/rss+xml');
  const body = await feed.text();
  expect(body).toContain('<rss');
  expect(body).toContain('One tree for docs and publishing');
});

// Mermaid renders inside a sandboxed frame because the pages' own policy blocks
// the inline styles it emits. These assert the isolation actually holds, not
// just that a picture appeared.
test('diagrams render inside a sandboxed frame', async ({ page }) => {
  const cspViolations = [];
  page.on('console', (m) => { if (/Content Security Policy/i.test(m.text())) cspViolations.push(m.text()); });

  await page.goto(selfPage('/docs/build/diagrams'));
  const roots = page.locator('.fastr-docs-mermaid');
  await expect(roots).toHaveCount(2);

  const first = roots.first();
  await first.scrollIntoViewIfNeeded();
  const frame = first.locator('iframe');
  await expect(frame).toHaveAttribute('sandbox', 'allow-scripts');

  // Rendering happens in the frame; the host only learns the height.
  await expect.poll(async () => (await frame.getAttribute('style')) || '', { timeout: 20_000 }).toContain('height');
  const rendered = page.frameLocator('.fastr-docs-mermaid iframe').first().locator('svg');
  await expect(rendered).toBeVisible({ timeout: 20_000 });

  // The relaxation the frame needs must not reach the page.
  expect(cspViolations, cspViolations.join('\n')).toHaveLength(0);
});

test('the diagram frame carries its own policy and the page does not', async ({ request }) => {
  const frame = await request.get(selfPage('/__fastr-docs/mermaid/diagram.html'));
  expect(frame.ok()).toBeTruthy();
  const framePolicy = frame.headers()['content-security-policy'] || '';
  expect(framePolicy).toContain("style-src 'self' 'unsafe-inline'");
  expect(framePolicy).toContain("frame-ancestors 'self'");
  // Its own bundle is cross-origin to an opaque-origin frame, so it needs this
  // or the browser blocks it and nothing renders.
  expect(frame.headers()['cross-origin-resource-policy']).toBe('cross-origin');

  const entry = await request.get(selfPage('/__fastr-docs/mermaid/frame/frame.js'));
  expect(entry.ok()).toBeTruthy();
  expect(entry.headers()['cross-origin-resource-policy']).toBe('cross-origin');
  // The entry is a module, fetched in CORS mode from an origin of "null".
  expect(entry.headers()['access-control-allow-origin']).toBe('*');
  expect(entry.headers()['cache-control']).toContain('immutable');

  const adapter = await request.get(selfPage('/__fastr-docs/mermaid/adapter.js'));
  expect(adapter.ok()).toBeTruthy();
  expect(adapter.headers()['cross-origin-resource-policy']).not.toBe('cross-origin');
});

// The frame used to pull all 3.4MB of Mermaid, once per frame, because every
// diagram type was in one file. Splitting it means a flowchart fetches the
// flowchart code and nothing else.
test('a diagram fetches only the chunks it needs', async ({ page }) => {
  let bytes = 0;
  page.on('response', async (r) => {
    if (!r.url().includes('__fastr-docs/mermaid')) return;
    try { bytes += (await r.body()).length; } catch {}
  });

  await page.goto(selfPage('/docs/build/diagrams'), { waitUntil: 'networkidle' });
  await page.evaluate(() => window.scrollTo(0, document.body.scrollHeight));
  await expect.poll(async () => (await page.locator('.fastr-docs-mermaid iframe').last().getAttribute('style')) || '',
    { timeout: 20_000 }).toContain('height');
  await page.waitForTimeout(1500);

  // Two frames, each an isolated origin that cannot reuse the other's download.
  // Before splitting this page transferred 6.9MB.
  expect(bytes).toBeLessThan(3_000_000);
  expect(bytes).toBeGreaterThan(100_000);
});

test('a diagram falls back to its source without JavaScript', async ({ browser }) => {
  const context = await browser.newContext({ javaScriptEnabled: false });
  const page = await context.newPage();
  await page.goto(selfPage('/docs/build/diagrams'));
  await expect(page.locator('.fastr-docs-mermaid__source').first()).toContainText('graph LR');
  await expect(page.locator('.fastr-docs-mermaid iframe')).toHaveCount(0);
  await context.close();
});

// Math takes the opposite approach to diagrams: it renders in the page, with no
// frame and no relaxed policy, because KaTeX applies layout through CSSOM
// rather than style attributes. These assert that claim rather than trusting it.
test('math renders in the page under the strict policy', async ({ page }) => {
  const cspViolations = [];
  page.on('console', (m) => { if (/Content Security Policy/i.test(m.text())) cspViolations.push(m.text()); });

  await page.goto(selfPage('/docs/build/math'));
  const roots = page.locator('.fastr-docs-math');
  await expect(roots).toHaveCount(6);
  await expect(roots.first().locator('.katex')).toBeVisible({ timeout: 20_000 });

  const state = await page.evaluate(() => {
    const all = [...document.querySelectorAll('.fastr-docs-math')];
    return {
      rendered: all.filter((r) => r.querySelector('.katex')).length,
      failed: all.filter((r) => r.hasAttribute('data-fastr-docs-math-failed')).length,
      // KaTeX sizes struts and rules with these. Their presence together with
      // an empty violation list is the whole argument for rendering in-page.
      styleAttrs: all.reduce((n, r) => n + r.querySelectorAll('[style]').length, 0),
      frames: document.querySelectorAll('.fastr-docs-math iframe').length,
    };
  });
  expect(state.rendered).toBe(6);
  expect(state.failed).toBe(0);
  expect(state.styleAttrs).toBeGreaterThan(50);
  expect(state.frames).toBe(0);
  expect(cspViolations, cspViolations.join('\n')).toHaveLength(0);
});

test('inline math shares a line with the sentence around it', async ({ page }) => {
  await page.goto(selfPage('/docs/build/math'));
  const inline = page.locator('.fastr-docs-math:not(.fastr-docs-math--display)').first();
  await expect(inline.locator('.katex')).toBeVisible({ timeout: 20_000 });
  const shares = await inline.evaluate((el) => {
    const paragraph = el.closest('p');
    return paragraph ? paragraph.getBoundingClientRect().height < el.getBoundingClientRect().height * 3 : false;
  });
  expect(shares).toBeTruthy();
});

test('the KaTeX stylesheet and fonts are served from the site itself', async ({ request }) => {
  const css = await request.get(selfPage('/__fastr-docs/katex/katex.css'));
  expect(css.ok()).toBeTruthy();
  const body = await css.text();
  // A CDN reference would be blocked by default-src 'self' and would leak
  // readers to a third party.
  expect(body).not.toContain('https://');
  expect(body).not.toContain('.ttf');

  const font = await request.get(selfPage('/__fastr-docs/katex/fonts/KaTeX_Main-Regular.woff2'));
  expect(font.ok()).toBeTruthy();
  expect(font.headers()['content-type']).toBe('font/woff2');
  expect(font.headers()['cache-control']).toContain('immutable');
});

test('math falls back to its source without JavaScript', async ({ browser }) => {
  const context = await browser.newContext({ javaScriptEnabled: false });
  const page = await context.newPage();
  await page.goto(selfPage('/docs/build/math'));
  await expect(page.locator('.fastr-docs-math').first()).toContainText('gamma');
  await expect(page.locator('.fastr-docs-math .katex')).toHaveCount(0);
  await context.close();
});

// A plugin bundle that belongs somewhere other than the page must not be
// server-rendered into it. The Mermaid bundle is megabytes and the frame
// fetches it itself; the KaTeX renderer is fetched by its own loader.
test('only small loaders are server-rendered as page scripts', async ({ page }) => {
  await page.goto(selfPage('/docs/build/diagrams'));
  const scripts = await page.evaluate(() =>
    [...document.querySelectorAll('script[src]')].map((s) => s.getAttribute('src')).filter((s) => s.includes('__fastr-docs')));
  expect(scripts).toContain('/__fastr-docs/katex/katex-loader.js');
  expect(scripts).toContain('/__fastr-docs/mermaid/adapter.js');
  expect(scripts).not.toContain('/__fastr-docs/mermaid/diagram.js');
  expect(scripts).not.toContain('/__fastr-docs/katex/katex.js');
  expect(scripts.filter((s) => s.endsWith('.css'))).toHaveLength(0);
});

// The renderer is 266KB and most pages have no math. Carrying it everywhere
// would make it the largest thing on a page that never uses it.
test('a page without math never downloads the renderer', async ({ page }) => {
  const fetched = [];
  page.on('response', (r) => { if (r.url().includes('/katex/')) fetched.push(r.url().split('/katex/')[1]); });

  await page.goto(selfPage('/docs/build/diagrams'), { waitUntil: 'networkidle' });
  await page.waitForTimeout(1000);

  expect(fetched).toContain('katex-loader.js');
  expect(fetched).not.toContain('katex.js');
  expect(fetched).not.toContain('katex.css');
  expect(fetched.filter((f) => f.startsWith('fonts/'))).toHaveLength(0);
});

// The docs shell swaps pages without a reload, so a reader who arrives on a
// page with no math and navigates to one that has it still gets the renderer.
test('the renderer arrives when math does, including after a client-side navigation', async ({ page }) => {
  const fetched = [];
  page.on('response', (r) => { if (r.url().includes('/katex/')) fetched.push(r.url().split('/katex/')[1]); });

  await page.goto(selfPage('/docs/build/diagrams'), { waitUntil: 'networkidle' });
  expect(fetched).not.toContain('katex.js');

  await page.goto(selfPage('/docs/build/math'));
  await expect(page.locator('.fastr-docs-math .katex').first()).toBeVisible({ timeout: 20_000 });
  expect(fetched).toContain('katex.js');
  expect(fetched).toContain('katex.css');

  const rendered = await page.evaluate(() => document.querySelectorAll('.fastr-docs-math .katex').length);
  expect(rendered).toBe(6);
});

// A components reference that shows only rendered output teaches nothing: a
// reader cannot see how to write the thing they are looking at.
test('every Markdown component is shown with its source', async ({ page }) => {
  await page.goto(selfPage('/docs/build/components'));
  const state = await page.evaluate(() => {
    const blocks = [...document.querySelectorAll('pre.ui-code-block__body')].map((p) => p.textContent);
    return {
      withSource: blocks.filter((b) => b.includes('{{<')).length,
      // GoFastr's parser reads only three fence characters, so an example that
      // shows a fenced block inside a shortcode used to break into three
      // pieces, the closing tag stranded in its own block.
      filetreeWhole: blocks.filter((b) => b.includes('{{< filetree >}}') && b.includes('{{< /filetree >}}')).length,
      diffWhole: blocks.filter((b) => b.includes('{{< diff') && b.includes('{{< /diff >}}')).length,
      orphans: blocks.filter((b) => b.trim() === '{{< /filetree >}}' || b.trim() === '{{< /diff >}}').length,
    };
  });
  // Admonitions, callout, tabs, cards, steps, filetree, details, diff.
  expect(state.withSource).toBeGreaterThanOrEqual(8);
  expect(state.filetreeWhole).toBe(1);
  expect(state.diffWhole).toBe(1);
  expect(state.orphans).toBe(0);
  await expect(page.locator('body')).not.toContainText('FASTRDOCS');
});

// One build serves both languages, so the chrome has to follow the page. A
// Spanish page wrapped in English furniture is the bug this guards.
test('the chrome is translated on a translated page', async ({ page }) => {
  await page.goto(selfPage('/es/docs/getting-started'));
  const es = await page.evaluate(() => ({
    contents: document.querySelector('.ui-sidebar__title')?.textContent?.trim(),
    search: document.querySelector('.fastr-docs-command-trigger__label')?.textContent?.trim(),
    switcherLabel: document.querySelector('.fastr-docs-variant-select__label')?.textContent?.trim(),
    switcher: [...(document.querySelector('[data-docs-variant-select=locale]')?.options || [])].map((o) => o.getAttribute('data-docs-variant-label') || o.textContent),
    navTabs: [...document.querySelectorAll('.ui-site-header__links a')].map((a) => a.textContent.trim()),
  }));
  expect(es.contents).toBe('Contenido');
  expect(es.search).toBe('Buscar');
  expect(es.switcherLabel).toBe('Idioma');
  // The selector names languages. "es" is a worse label than "Español" for
  // exactly the reader who needs it.
  expect(es.switcher).toEqual(['English', 'Español']);
  // A translated section belongs behind that selector, not beside the original
  // as a tab of its own.
  expect(es.navTabs).not.toContain('Español');
  await expect(page.locator('body')).not.toContainText('On this page');

  await page.goto(selfPage('/docs/getting-started'));
  const en = await page.evaluate(() => ({
    contents: document.querySelector('.ui-sidebar__title')?.textContent?.trim(),
    switcherLabel: document.querySelector('.fastr-docs-variant-select__label')?.textContent?.trim(),
  }));
  expect(en.contents).toBe('Contents');
  expect(en.switcherLabel).toBe('Language');
});

test('the language selector moves between translations of the same page', async ({ page }) => {
  await page.goto(selfPage('/docs/getting-started'));
  await page.selectOption('[data-docs-variant-select=locale]', { value: '/es/docs/getting-started' });
  await page.waitForURL('**/es/docs/getting-started');
  await expect(page.locator('h1')).toContainText('Primeros pasos');
  await expect(page.locator('.ui-sidebar__title')).toContainText('Contenido');
});

// The Spanish tree is the worked example for translating a fastr-docs site, so
// it has to stay coherent: pages present, chrome translated, selector paired.
test('the Spanish tree is a complete worked example', async ({ page }) => {
  for (const path of ['/es', '/es/docs', '/es/docs/getting-started', '/es/docs/concepts/router',
    '/es/docs/concepts/content', '/es/docs/concepts/layouts', '/es/docs/build/screens',
    '/es/docs/build/framework-ui', '/es/docs/build/openapi', '/es/docs/build/plugins',
    '/es/docs/build/blog', '/es/docs/build/themes', '/es/docs/build/components',
    '/es/docs/build/diagrams', '/es/docs/build/math', '/es/docs/operate/search',
    '/es/docs/operate/offline', '/es/docs/operate/testing', '/es/docs/operate/deploy',
    '/es/docs/operate/i18n', '/es/docs/operate/assets', '/es/docs/operate/feature-coverage',
    '/es/docs/collaborate/ai-authoring', '/es/ejemplos/arbol-de-rutas', '/es/ejemplos/superficies-propias']) {
    const response = await page.goto(selfPage(path));
    expect(response.status(), path).toBe(200);
    const state = await page.evaluate(() => ({
      sidebar: document.querySelector('.ui-sidebar__title')?.textContent?.trim(),
      selector: [...(document.querySelector('[data-docs-variant-select=locale]')?.options || [])].map((o) => o.getAttribute('data-docs-variant-label') || o.textContent),
    }));
    expect(state.sidebar, path).toBe('Contenido');
    // Every Spanish page has an English counterpart, so every one offers the
    // selector. A page whose original omits `locale:` silently loses it.
    expect(state.selector, path).toEqual(['English', 'Español']);
  }
});

// The examples are translated slug and all, so their paths do not mirror the
// English ones. translation_of pairs them anyway: the selector moves between
// the two, and each page tells search engines about the other.
test('a translated slug pairs with its original through translation_of', async ({ page }) => {
  await page.goto(selfPage('/es/ejemplos/arbol-de-rutas'));
  await expect(page.locator('h1')).toContainText('Ejemplo de árbol de rutas');
  const spanish = await page.evaluate(() => ({
    options: [...document.querySelector('[data-docs-variant-select=locale]').options].map((o) => o.getAttribute('data-docs-variant-label') || o.textContent),
    hreflang: document.querySelector('link[rel="alternate"][hreflang="en"]')?.getAttribute('href'),
    nav: [...document.querySelectorAll('.ui-site-header__links a[aria-current="page"]')].map((a) => a.textContent.trim()),
  }));
  expect(spanish.options).toEqual(['English', 'Español']);
  expect(spanish.hreflang).toBe('/examples/route-tree');
  expect(spanish.nav).toEqual(['Ejemplos']);

  await page.selectOption('[data-docs-variant-select=locale]', { value: '/examples/route-tree' });
  await page.waitForURL('**/examples/route-tree');
  await expect(page.locator('h1')).toContainText('Route tree');
  const english = await page.evaluate(() => ({
    hreflang: document.querySelector('link[rel="alternate"][hreflang="es"]')?.getAttribute('href'),
  }));
  expect(english.hreflang).toBe('/es/ejemplos/arbol-de-rutas');

  await page.selectOption('[data-docs-variant-select=locale]', { value: '/es/ejemplos/arbol-de-rutas' });
  await page.waitForURL('**/es/ejemplos/arbol-de-rutas');
});

// The blog is a second collection in Spanish: its own landing, sidebar,
// archive, feed and posts, in its own strings, paired with the English one.
test('the Spanish blog is a blog in Spanish', async ({ page }, testInfo) => {
  await page.goto(selfPage('/es/blog'));
  await expect(page.locator('h1').filter({ hasText: 'Blog' })).toBeVisible();
  const state = await page.evaluate(() => ({
    lang: document.documentElement.lang,
    layout: document.querySelector('[data-fui-layout^="blog"]')?.getAttribute('data-fui-layout'),
    feed: document.querySelector('.fastr-docs-blog__feed-link')?.getAttribute('href'),
    hreflang: document.querySelector('link[rel="alternate"][hreflang="en"]')?.getAttribute('href'),
    selector: [...(document.querySelector('[data-docs-variant-select=locale]')?.options || [])].map((o) => o.value),
    latest: document.querySelector('.fastr-docs-blog__section-title, h2')?.textContent.trim(),
  }));
  expect(state.lang).toBe('es');
  expect(state.layout).toBe('blog-es');
  expect(state.feed).toBe('/es/blog/feed.xml');
  expect(state.hreflang).toBe('/blog');
  expect(state.selector).toEqual(['/blog', '/es/blog']);
  await expect(page.getByText('Últimas entradas', { exact: true })).toBeVisible();

  const blogNav = await openBlogNav(page, testInfo, 'fastr-docs-blog-es-blog-sections');
  for (const label of ['Todas las entradas', 'Archivo', 'Etiquetas', 'Autores', 'Buscar']) {
    await expect(blogNav.getByRole('link', { name: label, exact: true })).toBeVisible();
  }
  await expect(blogNav.getByRole('link', { name: 'All posts', exact: true })).toHaveCount(0);
  if (isMobileProject(testInfo)) await page.keyboard.press('Escape');

  // A post with a translated slug pairs with its original.
  await page.goto(selfPage('/es/blog/un-arbol-para-docs-y-publicacion'));
  await expect(page.locator('h1')).toContainText('Un árbol para docs y publicación');
  await expect(page.locator('link[rel="alternate"][hreflang="en"]')).toHaveAttribute('href', '/blog/route-tree');
  await page.selectOption('[data-docs-variant-select=locale]', { value: '/blog/route-tree' });
  await page.waitForURL('**/blog/route-tree');
  await expect(page.locator('h1')).toContainText('One tree for docs and publishing');

  const feed = await page.request.get(selfPage('/es/blog/feed.xml'));
  expect(feed.ok()).toBeTruthy();
  expect(await feed.text()).toContain('Un árbol para docs y publicación');
});

// The header lives in the layout layer GoFastr keeps across client-side
// navigations. Without a refresh, a Spanish reader who clicked into an
// untranslated section kept Spanish tabs, a Spanish search label, a selector
// still aimed at the previous page, and a Spanish <html lang> over English
// content.
test('the header follows the page across client-side navigations', async ({ page }, testInfo) => {
  test.skip(isMobileProject(testInfo), 'the desktop header is the surface under test');
  const chrome = () => page.evaluate(() => ({
    path: location.pathname,
    lang: document.documentElement.lang,
    tabs: [...document.querySelectorAll('.ui-site-header__links a')].map((a) => a.textContent.trim()),
    selector: [...(document.querySelector('[data-docs-variant-select=locale]')?.options || [])].map((o) => o.value),
    search: document.querySelector('.fastr-docs-command-trigger__label')?.textContent.trim(),
    loads: performance.getEntriesByType('navigation').length,
  }));

  await page.goto(selfPage('/es/docs/getting-started'));
  expect((await chrome()).tabs).toContain('Documentación');

  // Same language: the selector must aim at the new page, not the old one.
  await page.locator('.ui-site-header__links a', { hasText: 'Blog' }).first().click();
  await page.waitForURL(/\/es\/blog\/?$/);
  await expect.poll(async () => (await chrome()).selector).toEqual(['/blog', '/es/blog']);

  // Across languages: an untranslated section is English, and so is the
  // header around it.
  await page.locator('.ui-site-header__links a', { hasText: 'Example API reference' }).first().click();
  await page.waitForURL(/\/api-reference\/?$/);
  await expect.poll(async () => (await chrome()).lang).toBe('en');
  const english = await chrome();
  expect(english.tabs).toContain('Documentation');
  expect(english.tabs).not.toContain('Documentación');
  expect(english.search).toBe('Search');
  expect(english.selector).toEqual([]);
  expect(english.loads).toBe(1);

  // And back into Spanish through the tab, still without a full load.
  await page.goBack();
  await page.waitForURL(/\/es\/blog\/?$/);
  await expect.poll(async () => (await chrome()).lang).toBe('es');
  const spanish = await chrome();
  expect(spanish.tabs).toContain('Documentación');
  expect(spanish.search).toBe('Buscar');
  expect(spanish.loads).toBe(1);
});

// A phone-width header has room for a code, not a language name: the
// selector reads "ES" there and "Español" on a desktop.
test('the language selector shortens to the code on a phone', async ({ page }, testInfo) => {
  await page.goto(selfPage('/es/docs/getting-started'));
  const shown = () => page.evaluate(() => {
    const select = document.querySelector('[data-docs-variant-select=locale]');
    return { selected: select.options[select.selectedIndex].textContent, all: [...select.options].map((o) => o.textContent) };
  });
  if (isMobileProject(testInfo)) {
    await expect.poll(async () => (await shown()).selected).toBe('ES');
    expect((await shown()).all).toEqual(['EN', 'ES']);
    // Wide enough again, the names come back.
    await page.setViewportSize({ width: 1024, height: 844 });
    await expect.poll(async () => (await shown()).selected).toBe('Español');
    return;
  }
  expect((await shown()).all).toEqual(['English', 'Español']);
});

// The page about translation, translated, is the demonstration that matters.
test('the translated i18n guide reads as Spanish throughout', async ({ page }) => {
  await page.goto(selfPage('/es/docs/operate/i18n'));
  await expect(page.locator('h1')).toHaveText('Idiomas y traducción');
  await expect(page.locator('.tabs-summary').first()).toHaveText('El contenido');
  const nav = await page.evaluate(() =>
    [...document.querySelectorAll('.fastr-docs-toc a, [class*=toc] a')].map((a) => a.textContent.trim()));
  expect(nav.join(' ')).toContain('Familias de rutas');
  await expect(page.locator('body')).not.toContainText('On this page');
});

// Switching language should replace the site, not open a second one beside it.
// The nav, the sidebar and the section groups all have to come across.
test('choosing Spanish replaces the whole site, not just the article', async ({ page }) => {
  await page.goto(selfPage('/es/docs/build/math'));
  const es = await page.evaluate(() => ({
    nav: [...document.querySelectorAll('.ui-site-header__links a')].map((a) => a.textContent.trim()),
    navHrefs: [...document.querySelectorAll('.ui-site-header__links a')].map((a) => a.getAttribute('href')),
    groups: [...document.querySelectorAll('.ui-sidebar li')].map((li) => li.firstElementChild?.textContent?.trim()),
  }));
  // A translated section is swapped for its translation, and points into it.
  expect(es.nav).toContain('Documentación');
  expect(es.nav).toContain('Ejemplos');
  expect(es.navHrefs).toContain('/es/docs');
  // Sections with no translation keep their original label rather than vanishing.
  expect(es.nav).toContain('Blog');
  // The section groups are Spanish too, which needs GroupConfig.Locale: a group
  // has no front matter to declare one.
  expect(es.groups.join(' ')).toContain('Conceptos');
  expect(es.groups.join(' ')).toContain('Construir');
  expect(es.groups.join(' ')).toContain('Operar');

  await page.goto(selfPage('/docs/build/math'));
  const en = await page.evaluate(() =>
    [...document.querySelectorAll('.ui-site-header__links a')].map((a) => a.textContent.trim()));
  expect(en).toContain('Documentation');
  expect(en).toContain('Examples');
});

// English pages carry no `locale:` in front matter, as a monolingual site is
// written. They still have to pair with their translation, or adding a language
// would mean annotating every existing page.
test('the selector appears without annotating the original pages', async ({ page }) => {
  for (const path of ['/docs/build/themes', '/docs/operate/deploy', '/examples/route-tree']) {
    await page.goto(selfPage(path));
    const options = await page.evaluate(() =>
      [...(document.querySelector('[data-docs-variant-select=locale]')?.options || [])].map((o) => o.getAttribute('data-docs-variant-label') || o.textContent));
    expect(options, path).toEqual(['English', 'Español']);
  }
});

// The translated sidebar has to mirror the original, not nest it one level
// deeper under the language's own name.
test('the Spanish sidebar mirrors the English one', async ({ page }) => {
  const read = async (path) => {
    await page.goto(selfPage(path));
    return page.evaluate(() => {
      const links = [...document.querySelectorAll('.ui-sidebar a')];
      return { count: links.length, first: { label: links[0]?.textContent.trim(), href: links[0]?.getAttribute('href') } };
    });
  };
  const es = await read('/es/docs/build/math');
  const en = await read('/docs/build/math');

  expect(es.count).toBe(en.count);
  // The home is first in both, and each one stays inside its own language.
  expect(en.first).toEqual({ label: 'Home', href: '/' });
  expect(es.first).toEqual({ label: 'Inicio', href: '/es' });

  await page.goto(selfPage('/es/docs/build/math'));
  const labels = await page.evaluate(() => [...document.querySelectorAll('.ui-sidebar a')].map((a) => a.textContent.trim()));
  // "Español" was the locale home showing up as a section of the site.
  expect(labels).not.toContain('Español');
  expect(labels).toContain('Primeros pasos');
  expect(labels.every((l) => l !== 'Getting started')).toBe(true);
});

// Following the home link from a Spanish page must not drop the reader into the
// English site. Asserted by href rather than by clicking, because at mobile
// widths the sidebar lives inside a drawer.
test('the home link keeps a Spanish reader in Spanish', async ({ page }) => {
  await page.goto(selfPage('/es/docs/build/math'));
  const href = await page.locator('.ui-sidebar a').first().getAttribute('href');
  expect(href).toBe('/es');

  await page.goto(selfPage(href));
  await expect(page.locator('.ui-sidebar__title')).toContainText('Contenido');
  await expect(page.locator('h1')).toContainText('árbol de rutas');
});

// The pager under a document has to be in the reader's language, and has to
// stay inside it. ui.DocPrevNext writes "Previous" and "Next" as literals, so
// fastr-docs renders its own.
test('the pager is translated and does not cross languages', async ({ page }) => {
  await page.goto(selfPage('/es/docs/build/diagrams'));
  const es = await page.evaluate(() => ({
    dirs: [...document.querySelectorAll('.ui-doc-layout__pager-dir')].map((e) => e.textContent.trim()),
    hrefs: [...document.querySelectorAll('.ui-doc-layout__prev, .ui-doc-layout__next')].map((a) => a.getAttribute('href')),
  }));
  expect(es.dirs).toEqual(['← Anterior', 'Siguiente →']);
  expect(es.hrefs.every((h) => h.startsWith('/es/'))).toBe(true);

  await page.goto(selfPage('/docs/build/diagrams'));
  const en = await page.evaluate(() => ({
    dirs: [...document.querySelectorAll('.ui-doc-layout__pager-dir')].map((e) => e.textContent.trim()),
    hrefs: [...document.querySelectorAll('.ui-doc-layout__prev, .ui-doc-layout__next')].map((a) => a.getAttribute('href')),
  }));
  expect(en.dirs).toEqual(['← Previous', 'Next →']);
  expect(en.hrefs.some((h) => h.startsWith('/es/'))).toBe(false);
});

// One index holds every language, so results are narrowed to the page being
// read. Otherwise a Spanish query answers with English pages, and following one
// silently leaves the translation.
test('search results stay in the language of the page', async ({ page }) => {
  const search = async (path) => {
    await page.goto(selfPage(path), { waitUntil: 'networkidle' });
    await page.locator('.fastr-docs-command-trigger:visible').first().click();
    await page.locator('#fastr-docs-command-palette-input:visible').first().fill('router');
    await expect.poll(async () => page.evaluate(() =>
      document.querySelectorAll('[role="option"][data-fui-push-state]').length), { timeout: 15_000 }).toBeGreaterThan(0);
    return page.evaluate(() =>
      [...document.querySelectorAll('[role="option"][data-fui-push-state]')].map((o) => o.getAttribute('data-fui-push-state')));
  };

  const spanish = await search('/es/docs/build/math');
  expect(spanish.length).toBeGreaterThan(0);
  expect(spanish.every((u) => u === '/es' || u.startsWith('/es/'))).toBe(true);

  const english = await search('/docs/build/math');
  expect(english.length).toBeGreaterThan(0);
  expect(english.some((u) => u.startsWith('/es'))).toBe(false);
});

// The 404 is the one surface with no route to read a language from, so it goes
// by the URL that was missed.
test('a missing Spanish URL gets a Spanish 404', async ({ page }) => {
  await page.goto(selfPage('/es/no-existe'));
  await expect(page.locator('h1')).toContainText('no encontrada');
  await expect(page.locator('.fastr-docs-not-found__message')).toContainText('/es/no-existe');
  // The recovery link must not drop the reader into the English site.
  await expect(page.locator('.fastr-docs-not-found__link')).toHaveAttribute('href', '/es');

  await page.goto(selfPage('/no-such-page'));
  await expect(page.locator('h1')).toContainText('Page not found');
  await expect(page.locator('.fastr-docs-not-found__link')).toHaveAttribute('href', '/');
});

// The palette modal is mounted once site-wide, so its strings ride on the
// per-page trigger and are applied when it opens.
test('the search modal opens in the language of the page', async ({ page }) => {
  const openPalette = async (path) => {
    await page.goto(selfPage(path), { waitUntil: 'networkidle' });
    await page.locator('.fastr-docs-command-trigger:visible').first().click();
    await expect(page.locator('#fastr-docs-command-palette-input')).toBeVisible();
    return page.evaluate(() => ({
      placeholder: document.getElementById('fastr-docs-command-palette-input')?.getAttribute('placeholder'),
      close: document.querySelector('.fastr-docs-command-palette__close')?.getAttribute('aria-label'),
    }));
  };
  const es = await openPalette('/es/docs/build/math');
  expect(es.placeholder).toContain('Buscar');
  expect(es.close).toContain('Cerrar');

  const en = await openPalette('/docs/build/math');
  expect(en.placeholder).toContain('Search');
  expect(en.close).toContain('Close');
});
