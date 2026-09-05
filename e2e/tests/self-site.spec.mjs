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

const openBlogNav = async (page, testInfo) => {
  const nav = page.locator(isMobileProject(testInfo) ? '[data-fui-widget="fastr-docs-blog-sections"]' : '.layout-blog .ui-sidebar__inline');
  if (isMobileProject(testInfo) && await nav.isHidden()) {
    await page.locator('[data-fui-open="fastr-docs-blog-sections"]:visible').first().click();
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
  await expect(page.locator('.layout-blog')).toBeVisible();
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
    switcher: [...(document.querySelector('[data-docs-variant-select=locale]')?.options || [])].map((o) => o.textContent),
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
  await page.selectOption('[data-docs-variant-select=locale]', { label: 'Español' });
  await page.waitForURL('**/es/docs/getting-started');
  await expect(page.locator('h1')).toContainText('Primeros pasos');
  await expect(page.locator('.ui-sidebar__title')).toContainText('Contenido');
});
