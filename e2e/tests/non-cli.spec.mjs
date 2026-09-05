import { test, expect } from '@playwright/test';
import { runtime } from '../support/runtime.mjs';

const manualPage = (path = '/') => `${runtime().manualURL}${path}`;
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

test('hand-authored Router preserves nested navigation and typed screens', async ({ page }, testInfo) => {
  await page.goto(manualPage('/'));
  await expect(page.getByRole('heading', { name: 'Manual Docs', exact: true })).toBeVisible();
  await expect(page.locator('html')).toHaveAttribute('lang', 'pt-BR');
  await expect(page.getByRole('heading', { name: 'Start here', exact: true })).toBeVisible();
  await expect(page.getByRole('heading', { name: 'Next steps', exact: true })).toBeVisible();
  await expect(page.getByRole('link', { name: 'Getting started', exact: true }).first()).toBeVisible();
  if (testInfo.project.name === 'mobile-chromium') {
    const trigger = page.locator('[data-fui-open="fastr-docs-sections"]:visible').first();
    await expect(trigger).toBeVisible();
    await trigger.click();
    const drawer = page.locator('[data-fui-widget="fastr-docs-sections"]');
    await expect(drawer.getByRole('link', { name: 'Framework lab', exact: true })).toBeVisible();
    await drawer.getByRole('link', { name: 'Framework lab', exact: true }).click();
    await expect(page).toHaveURL(/\/framework-lab\/?$/);
    await expect(page.locator('[data-testid="framework-lab"]')).toBeVisible();
  } else {
    await expect(page.locator('nav.ui-site-header__links').getByRole('link', { name: 'Guides', exact: true })).toBeVisible();
    await expect(page.locator('.ui-sidebar__inline a.ui-sidebar__link[href="/framework-lab"]')).toBeVisible();
    await expect(page.locator('.ui-sidebar__inline a.ui-sidebar__link[href="/api-reference"]')).toBeVisible();
  }
  await page.keyboard.press('Control+K');
  const paletteInput = page.locator('#fastr-docs-command-palette-input:visible');
  await expect(paletteInput).toBeVisible();
  await paletteInput.fill('framework lab');
  await page.locator('[role="option"][data-fui-push-state="/framework-lab"]:visible').click();
  await expect(page).toHaveURL(/\/framework-lab\/?$/);
  await page.goto(manualPage('/'));
  if (testInfo.project.name === 'mobile-chromium') {
    await expect(page.locator('.ui-sidebar__inline')).toBeHidden();
  } else {
    await expect(page.locator('.ui-sidebar__inline a.ui-sidebar__link[href="/framework-lab"]')).toBeVisible();
    await expect(page.locator('.ui-sidebar__inline a.ui-sidebar__link[href="/api-reference"]')).toBeVisible();
  }

  await page.goto(manualPage('/guides/getting-started'));
  await expect(page.getByRole('heading', { name: 'Getting started', exact: true })).toBeVisible();
  await expect(page.getByText('Manual component', { exact: true })).toBeVisible();
  await expect(page.locator('.fastr-docs-toc')).toContainText('Compose a route tree');
  await expect(page.locator('.ui-sidebar__inline a.ui-sidebar__link[href="/framework-lab"]')).toHaveCount(0);
  await expect(page.locator('.ui-sidebar__inline a.ui-sidebar__link[href="/api-reference"]')).toHaveCount(0);

  await page.goto(manualPage('/framework-lab'));
  await expect(page.getByRole('heading', { name: 'Framework lab', exact: true })).toBeVisible();
  await expect(page.locator('[data-testid="framework-lab"]')).toBeVisible();
  if (testInfo.project.name === 'mobile-chromium') {
    await expect(page.locator('[data-fui-widget="fastr-docs-sections"]')).toBeHidden();
  } else {
    await expect(page.locator('.ui-sidebar__inline a.ui-sidebar__link[href="/framework-lab"]')).toBeVisible();
  }
  await expect(page.locator('.ui-sidebar__inline a.ui-sidebar__link[href="/api-reference"]')).toHaveCount(0);
});

test('hand-authored command palette reopens after filtering', async ({ page }) => {
  await page.goto(manualPage('/'));

  await page.keyboard.press('Control+K');
  const firstInput = page.locator('#fastr-docs-command-palette-input:visible').first();
  await expect(firstInput).toBeVisible();
  // Searching swaps the static route list for ranked index results, which
  // drops data-fui-static-options, so this matches the listbox itself.
  const options = page.locator('#fastr-docs-command-palette-input-listbox [role="option"]:visible');

  await firstInput.fill('framework');
  await expect(options.filter({ hasText: 'Framework lab' })).toHaveCount(1);
  const filteredCount = await options.count();

  await firstInput.fill('');
  const optionCount = await options.count();
  expect(optionCount).toBeGreaterThan(filteredCount);
  await firstInput.press('Escape');
  await expect(page.locator('[data-fui-widget="fastr-docs-command-palette"]:visible')).toHaveCount(0);

  await page.keyboard.press('Control+K');
  const secondInput = page.locator('#fastr-docs-command-palette-input:visible').first();
  await expect(secondInput).toBeVisible();
  await expect(options).toHaveCount(optionCount);
});

test('hand-authored navigation expands and selects the current nested page', async ({ page }, testInfo) => {
  await page.goto(manualPage('/guides/patterns'));

  const sidebar = await openSectionNav(page, testInfo);
  const guides = sidebar.locator('summary.ui-sidebar__link').filter({ hasText: /^Guides$/ }).locator('xpath=..');

  await expect(guides).toHaveAttribute('open', '');
  const patterns = guides.getByRole('link', { name: 'Patterns', exact: true });
  await expect(patterns).toHaveAttribute('aria-current', 'page');
  await expect(patterns).toHaveClass(/active/);
});

test('hand-authored route badges stay visible in desktop and mobile navigation', async ({ page }, testInfo) => {
  await page.goto(manualPage('/guides/patterns'));

  const inline = page.locator('.ui-sidebar__inline');
  if (testInfo.project.name === 'mobile-chromium') {
    await page.locator('[data-fui-open="fastr-docs-sections"]:visible').first().click();
    const drawer = page.locator('[data-fui-widget="fastr-docs-sections"]');
    await expect(drawer.locator('.fastr-docs-nav-badge[data-badge-label="Popular"]')).toBeVisible();
    await expect(drawer.locator('.fastr-docs-nav-badge[data-badge-label="New"]')).toBeVisible();
    await expect(drawer.getByRole('link', { name: 'Patterns', exact: true })).toHaveAttribute('aria-current', 'page');
    return;
  }

  await expect(inline.locator('.fastr-docs-nav-badge[data-badge-label="Popular"]')).toBeVisible();
  await expect(inline.locator('.fastr-docs-nav-badge[data-badge-label="New"]')).toBeVisible();
  await expect(inline.getByRole('link', { name: 'Patterns', exact: true })).toHaveAttribute('aria-current', 'page');
});

test('hand-authored framework lab exercises interactive component behavior', async ({ page }) => {
  await page.goto(manualPage('/framework-lab'));

  await expect(page.locator('[data-testid="stats-surface"]')).toBeVisible();
  await expect(page.locator('[data-testid="interactive-surface"]')).toBeVisible();
  await expect(page.locator('[data-testid="form-surface"]')).toBeVisible();
  await expect(page.locator('[data-testid="content-surface"]')).toBeVisible();
  await expect(page.locator('[data-testid="data-surface"]')).toBeVisible();

  await page.getByText('Runtime', { exact: true }).click();
  await expect(page.getByText('The signal runtime owns the selection.', { exact: true })).toBeVisible();
  await page.getByText('Show implementation note', { exact: true }).click();
  await expect(page.getByText('Disclosure uses native details semantics with framework behavior.', { exact: true })).toBeVisible();

  const counter = page.locator('.fui-counter').first();
  await expect(counter).toBeVisible();
  await expect(counter).toContainText('0');
  await counter.getByRole('button').last().click();
  await expect(counter).toContainText('1');

  const notifications = page.locator('#manual-notifications');
  await expect(notifications).toBeChecked();
  await page.locator('label[for="manual-notifications"]').click();
  await expect(notifications).not.toBeChecked();
  await page.locator('label[for="density--spacious"]').click();
  await expect(page.locator('#density--spacious')).toBeChecked();
});

test('hand-authored framework lab dismisses persistent feedback and preserves it on reload', async ({ page }) => {
  await page.goto(manualPage('/framework-lab'));
  const banner = page.locator('#manual-fixture-banner');
  await expect(banner).toBeVisible();
  await banner.locator('[data-fui-banner-dismiss]').click();
  await expect(banner).toBeHidden();

  await page.reload();
  await expect(page.locator('#manual-fixture-banner')).toHaveCount(0);
});

test('hand-authored localized routes switch locale and version without losing the page family', async ({ page }) => {
  await page.goto(manualPage('/locales/pt-BR/v1/guide'));
  await expect(page.getByRole('heading', { name: 'Variant guide', exact: true })).toBeVisible();
  const locale = page.locator('[data-docs-variant-select="locale"]:visible').first();
  const version = page.locator('[data-docs-variant-select="version"]:visible').first();
  await expect(locale).toBeVisible();
  await expect(version).toBeVisible();
  await locale.selectOption('/locales/fr/v1/guide');
  await expect(page).toHaveURL(/\/locales\/fr\/v1\/guide\/?$/);
  await version.selectOption('/locales/fr/v2/guide');
  await expect(page).toHaveURL(/\/locales\/fr\/v2\/guide\/?$/);
  await expect(page.getByRole('heading', { name: 'Variant guide', exact: true })).toBeVisible();
});

test('hand-authored localized route family keeps its top navigation active', async ({ page }, testInfo) => {
  await page.goto(manualPage('/locales/fr/v2/guide'));
  if (isMobileProject(testInfo)) {
    const sidebar = await openSectionNav(page, testInfo);
    const localized = sidebar.locator('summary.ui-sidebar__link').filter({ hasText: /^Localized guides$/ }).locator('xpath=..');
    await expect(localized).toHaveAttribute('open', '');
    await expect(localized.locator('a.ui-sidebar__link[href="/locales/fr/v2/guide"]')).toHaveAttribute('aria-current', 'page');
    return;
  }

  const primaryNav = await openPrimaryNav(page, testInfo);
  const localized = primaryNav.getByRole('link', { name: 'Localized guides', exact: true });
  await expect(localized).toHaveAttribute('aria-current', 'page');
  await expect(localized).toHaveClass(/active/);
});

test('hand-authored form posts through the server and re-renders context', async ({ page }) => {
  await page.goto(manualPage('/framework-lab'));
  await page.locator('#manual-name').fill('Grace Hopper');
  await page.locator('#manual-theme').selectOption('dark');
  await page.locator('#manual-notes').fill('Server-rendered preference flow.');
  await page.getByRole('button', { name: 'Save preferences', exact: true }).click();

  await expect(page).toHaveURL(/\/framework-lab\?saved=1$/);
  await expect(page.locator('#manual-saved')).toContainText('Preferences saved');
  await expect(page.locator('#manual-saved')).toContainText('native POST completed');
});

test('hand-authored content and data components expose usable controls', async ({ page }) => {
  await page.goto(manualPage('/framework-lab'));

  await page.getByText('JSON', { exact: true }).click();
  await expect(page.getByText('{ "route": "/framework-lab" }', { exact: true })).toBeVisible();
  await page.getByRole('button', { name: 'Copy route', exact: true }).click();
  const copyStatus = page.getByRole('button', { name: 'Copy route', exact: true }).locator('xpath=..').locator('[data-fui-copy-status]');
  await expect(copyStatus).toHaveText('Copied');

  await page.getByRole('button', { name: /next slide/i }).click();
  await expect(page.getByText('Framework-owned interaction.', { exact: true })).toBeVisible();

  await page.locator('[data-fui-carousel-dot="2"]').click();
  await expect(page.locator('[data-fui-carousel-dot="2"]')).toHaveAttribute('aria-current', 'true');
  await expect(page.getByText('Exportable PWA shell.', { exact: true })).toBeVisible();

  await expect(page.locator('#manual-table')).toBeVisible();
  await expect(page.locator('#manual-table')).toContainText('/framework-lab');
  await expect(page.locator('#manual-line-chart')).toBeVisible();
  await expect(page.locator('#manual-bar-chart')).toBeVisible();
  await expect(page.locator('#manual-pie-chart')).toBeVisible();
});

test('hand-authored form controls accept number, rating, and tag input interactions', async ({ page }) => {
  await page.goto(manualPage('/framework-lab'));
  await page.locator('#manual-sessions').fill('8');
  await expect(page.locator('#manual-sessions')).toHaveValue('8');

  const rating = page.locator('[name="rating"][value="5"]');
  await page.locator('label[for="rating-5"]').click();
  await expect(rating).toBeChecked();

  const tags = page.locator('#manual-tags');
  await tags.fill('e2e');
  await tags.press('Enter');
  await expect(page.locator('input[name="tags"][value="e2e"]')).toHaveCount(1);
  await expect(page.locator('.ui-tag-input__chip').filter({ hasText: 'e2e' })).toBeVisible();
});

test('hand-authored filter toolbar and sortable table preserve URL-driven state', async ({ page }, testInfo) => {
  await page.goto(manualPage('/framework-lab'));
  await page.locator('#filter-search-q').fill('route');
  await page.locator('[name="status"]').selectOption('ready');
  await page.locator('[name="sort"]').selectOption('recent');
  await page.getByRole('button', { name: 'Apply', exact: true }).click();
  await expect(page).toHaveURL(/\/framework-lab\?[^#]*q=route/);
  await expect(page).toHaveURL(/status=ready/);
  await expect(page).toHaveURL(/sort=recent/);

  await page.getByRole('link', { name: 'Reset', exact: true }).click();
  await expect(page).toHaveURL(/\/framework-lab\/?$/);

  if (testInfo.project.name === 'mobile-chromium') {
    await page.locator('[name="sort"]').selectOption('name');
    await page.getByRole('button', { name: 'Apply', exact: true }).click();
    await expect(page).toHaveURL(/\/framework-lab\?.*sort=name/);
    return;
  }
  await page.locator('#manual-table a.ui-data-table__sort').first().click();
  await expect(page).toHaveURL(/\/framework-lab\?sort=route&dir=asc$/);
});

test('hand-authored OpenAPI, local PWA, and static export surfaces work', async ({ page, request }) => {
  await page.goto(manualPage('/api-reference'));
  await expect(page.getByRole('heading', { name: 'Manual fixture API', exact: true })).toBeVisible();
  await expect(page.locator('[data-openapi-server-url]')).toHaveAttribute('data-openapi-server-url', /4176\/v1/);

  const pattern = await request.get(manualPage('/guides/patterns'));
  expect(pattern.ok()).toBeTruthy();
  const patternHTML = await pattern.text();
  expect(patternHTML).toContain('noindex,nofollow');
  expect(patternHTML).toContain('https://docs.example/patterns');
  expect(patternHTML).toContain('Edit this page');
  const redirect = await request.get(manualPage('/old-patterns'), { maxRedirects: 0 });
  expect(redirect.status()).toBe(308);
  expect(redirect.headers().location).toContain('/guides/patterns');
  const asset = await request.get(manualPage('/assets/fixture.txt'));
  expect(asset.ok()).toBeTruthy();
  expect(await asset.text()).toBe('manual asset');
  const liveManifest = await request.get(manualPage('/__manual/manifest.json'));
  expect(liveManifest.ok()).toBeTruthy();
  expect(await liveManifest.text()).toContain('fastr-docs/v1');
  const llms = await request.get(manualPage('/llms.txt'));
  expect(llms.ok()).toBeTruthy();
  expect(await llms.text()).toContain('## When to use');
  const agentCard = await request.get(manualPage('/.well-known/agent-card.json'));
  expect(agentCard.ok()).toBeTruthy();
  expect(await agentCard.text()).toContain('Manual Docs');
  const mcpManifest = await request.get(manualPage('/.well-known/mcp.json'));
  expect(mcpManifest.ok()).toBeTruthy();
  expect(await mcpManifest.text()).toContain('/mcp');
  const mcpServerCard = await request.get(manualPage('/.well-known/mcp/server-card.json'));
  expect(mcpServerCard.ok()).toBeTruthy();
  expect(await mcpServerCard.text()).toContain('Manual Docs');
  const mcpCatalog = await request.get(manualPage('/.well-known/mcp/catalog.json'));
  expect(mcpCatalog.ok()).toBeTruthy();
  expect(await mcpCatalog.text()).toContain('/mcp');
  const mcpReservedCard = await request.get(manualPage('/mcp/server-card'));
  expect(mcpReservedCard.ok()).toBeTruthy();
  expect(await mcpReservedCard.text()).toContain('Manual Docs');
  const mcpInit = await request.post(manualPage('/mcp'), {
    headers: { 'Content-Type': 'application/json' },
    data: {
      jsonrpc: '2.0',
      id: 1,
      method: 'initialize',
      params: {
        protocolVersion: '2025-06-18',
        capabilities: {},
        clientInfo: { name: 'fastr-docs-non-cli-e2e', version: '1' },
      },
    },
  });
  expect(mcpInit.ok()).toBeTruthy();
  const mcpBody = await mcpInit.text();
  expect(mcpBody).toContain('serverInfo');
  expect(mcpBody).not.toContain('"error"');
  const sitemap = await request.get(manualPage('/sitemap.xml'));
  expect(sitemap.ok()).toBeTruthy();
  expect(await sitemap.text()).toContain('/guides/getting-started');
  expect(await sitemap.text()).not.toContain('/guides/patterns');
  const robots = await request.get(manualPage('/robots.txt'));
  expect(robots.ok()).toBeTruthy();
  expect(await robots.text()).toContain('Disallow: /__manual/');

  const blog = await request.get(manualPage('/blog'));
  expect(blog.ok()).toBeTruthy();
  expect(await blog.text()).toContain('Fixture release');
  const feed = await request.get(manualPage('/blog/feed.xml'));
  expect(feed.ok()).toBeTruthy();
  expect(feed.headers()['content-type']).toContain('application/rss+xml');
  expect(await feed.text()).toContain('Fixture release');

  const filter = page.locator('[data-openapi-filter]');
  const operation = page.locator('[data-openapi-operation]').first();
  await filter.fill('listProjects');
  await expect(operation).toBeVisible();
  await page.locator('[data-openapi-operation-select]').selectOption({ index: 0 });
  await page.locator('[data-openapi-try]').click();
  await expect(page.locator('[data-openapi-response]')).toContainText('E2E project');

  const { manualStaticURL } = runtime();
  for (const [path, marker] of [
    ['/', 'Manual Docs'],
    ['/guides/getting-started/', 'Getting started'],
    ['/framework-lab/', 'Framework lab'],
    ['/api-reference/', 'data-openapi-try'],
    ['/blog/', 'Fixture release'],
    ['/blog/fixture-release/', 'Fixture release'],
    ['/assets/fixture.txt', 'manual asset'],
  ]) {
    const response = await request.get(manualStaticURL + path);
    expect(response.ok(), `${path} should be served by the manual export`).toBeTruthy();
    expect(await response.text()).toContain(marker);
  }

  const staticFeed = await request.get(manualStaticURL + '/blog/feed.xml');
  expect(staticFeed.ok()).toBeTruthy();
  expect(staticFeed.headers()['content-type']).toContain('application/rss+xml');
  expect(await staticFeed.text()).toContain('Fixture release');

  const staticSitemap = await request.get(manualStaticURL + '/sitemap.xml');
  expect(staticSitemap.ok()).toBeTruthy();
  expect(await staticSitemap.text()).toContain('/guides/getting-started');
  expect(await staticSitemap.text()).not.toContain('/guides/patterns');

  const manifest = await request.get(manualStaticURL + '/manifest.webmanifest');
  expect(manifest.ok()).toBeTruthy();
  expect(await manifest.text()).toContain('icon-192.png');
  const registration = await request.get(manualStaticURL + '/__gofastr/pwa/register.js');
  expect(registration.ok()).toBeTruthy();
  expect(await registration.text()).toContain('gofastr:pwa-update');
  const offlineScreen = await request.get(manualStaticURL + '/__gofastr/pwa/offline/');
  expect(offlineScreen.ok()).toBeTruthy();
  expect((await offlineScreen.text()).toLowerCase()).toContain('offline');
  const serviceWorker = await request.get(manualStaticURL + '/service-worker.js');
  expect(serviceWorker.ok()).toBeTruthy();
  expect(await serviceWorker.text()).toContain('__manual/search.json');
  expect(await serviceWorker.text()).toContain('__manual/manifest.json');
});

test('hand-authored framework surfaces remain usable at narrow widths', async ({ page }) => {
  await page.setViewportSize({ width: 320, height: 844 });
  await page.goto(manualPage('/framework-lab'));
  await expect(page.locator('[data-testid="framework-lab"]')).toBeVisible();
  await expect(page.getByRole('button', { name: 'Apply', exact: true })).toBeVisible();
  await expect(page.getByRole('link', { name: 'Reset', exact: true })).toBeVisible();
  await expect(page.locator('#manual-table')).toBeVisible();
  await expect.poll(() => page.evaluate(() => document.documentElement.scrollWidth <= window.innerWidth)).toBe(true);
});
