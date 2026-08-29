import { test, expect } from '@playwright/test';

test('landing surface exposes the route-first product and navigates into docs', async ({ page }) => {
  await page.goto('/');
  await expect(page.getByRole('heading', { name: 'A docs framework that starts as a router.' })).toBeVisible();
  await expect(page.locator('.fastr-docs-home__route-row')).toHaveCount(6);
  await expect(page.getByText('18 routes', { exact: true })).toBeVisible();
  await expect(page.getByText('AI-ready from the first commit')).toBeVisible();

  await page.getByRole('link', { name: 'Open the docs' }).click();
  await expect(page).toHaveURL(/\/docs\/?$/);
  await expect(page.getByRole('heading', { name: 'E2E Docs', exact: true })).toBeVisible();

  await page.goto('/docs/getting-started');
  await expect(page).toHaveURL(/\/docs\/getting-started\/?$/);
  await expect(page.getByRole('heading', { name: 'Getting started', exact: true })).toBeVisible();
});

test('local search returns a route and opens it', async ({ page }) => {
  await page.goto('/docs');
  await page.keyboard.press('Control+K');
  const input = page.locator('#fastr-docs-command-palette-input:visible').first();
  await expect(input).toBeVisible();
  await input.fill('getting started');
  const result = page.locator('[role="option"][data-fui-push-state="/docs/getting-started"]:visible').first();
  await expect(result).toContainText('Getting started');
  await expect(result).toHaveAttribute('data-fui-push-state', '/docs/getting-started');
  await result.click();
  await expect(page).toHaveURL(/\/docs\/getting-started\/?$/);
});

test('landing actions expose both primary paths into the handbook', async ({ page }) => {
  await page.goto('/');
  await page.getByRole('link', { name: 'Read the guide' }).click();
  await expect(page).toHaveURL(/\/docs\/getting-started\/?$/);
  await expect(page.getByRole('heading', { name: 'Getting started', exact: true })).toBeVisible();

  await page.goto('/');
  await page.getByRole('link', { name: 'Open the docs' }).click();
  await expect(page).toHaveURL(/\/docs\/?$/);
  await expect(page.getByRole('heading', { name: 'E2E Docs', exact: true })).toBeVisible();
});

test('top navigation exposes the top-level documentation groups', async ({ page }, testInfo) => {
  test.skip(testInfo.project.name === 'mobile-chromium', 'primary nav is collapsed into the mobile drawer');
  await page.goto('/docs');
  const nav = page.locator('nav.ui-site-header__links');
  await expect(nav.getByRole('link', { name: 'Documentation', exact: true })).toBeVisible();
  await expect(nav.getByRole('link', { name: 'API reference', exact: true })).toBeVisible();
  await expect(nav.getByRole('link', { name: 'Examples', exact: true })).toBeVisible();
  await expect(page.locator('.fastr-docs-header__context')).toHaveCount(0);
  await expect(page.locator('.fastr-docs-command-trigger:visible')).toHaveCount(1);
  await expect(page.locator('.fastr-docs-command-trigger__hint:visible')).toContainText('K');

  await nav.getByRole('link', { name: 'API reference', exact: true }).click();
  await expect(page).toHaveURL(/\/api-reference\/?$/);
  await expect(page.getByRole('heading', { name: 'API reference', exact: true })).toBeVisible();
  await expect(page.locator('.layout-docs-api')).toBeVisible();
  await expect(page.locator('[data-fui-screen-group="/api-reference/"]')).toBeVisible();

  await nav.getByRole('link', { name: 'Examples', exact: true }).click();
  await expect(page).toHaveURL(/\/examples\/playground\/?$/);
  await expect(nav.getByRole('link', { name: 'Examples', exact: true })).toHaveAttribute('aria-current', 'page');
  await expect(page.getByRole('heading', { name: 'Playground', exact: true })).toBeVisible();
  await expect(page.locator('[data-testid="examples-playground"]')).toBeVisible();
  await expect(page.locator('.layout-docs-section')).toBeVisible();
});

test('the handbook mounts every documented route', async ({ page }) => {
  const routes = [
    ['/docs', 'E2E Docs'],
    ['/docs/getting-started', 'Getting started'],
    ['/docs/concepts/router', 'The router'],
    ['/docs/concepts/content', 'Content authoring'],
    ['/docs/concepts/layouts', 'Layouts and navigation'],
    ['/docs/build/screens', 'Screens and components'],
    ['/docs/build/openapi', 'OpenAPI reference'],
    ['/docs/build/plugins', 'Plugins and extensions'],
    ['/docs/operate/search', 'Search'],
    ['/docs/operate/offline', 'Offline and PWA'],
    ['/docs/operate/testing', 'Testing'],
    ['/docs/operate/deploy', 'Deploy and customize'],
    ['/docs/collaborate/ai-authoring', 'AI authoring'],
    ['/examples/playground', 'Playground'],
    ['/examples/route-tree', 'Route tree'],
    ['/examples/custom-surface', 'Custom surfaces'],
    ['/api-reference', 'API reference'],
  ];
  for (const [path, title] of routes) {
    await page.goto(path);
    await expect(page).toHaveURL(new RegExp(`${path.replaceAll('/', '\\/')}/?$`));
    await expect(page.getByRole('heading', { name: title, exact: true })).toBeVisible();
    await expect(page.locator('main')).toHaveCount(1);
    await expect(page.getByRole('heading', { level: 1 })).toHaveCount(1);
    await expect.poll(() => page.evaluate(() => document.documentElement.scrollWidth <= window.innerWidth)).toBe(true);
  }
});

test('the generated playground exercises typed framework controls', async ({ page }) => {
  await page.goto('/examples/playground');
  await expect(page.getByTestId('examples-playground')).toBeVisible();

  const implementationTab = page.getByRole('tab', { name: 'Implementation', exact: true });
  await expect(implementationTab).toBeVisible();
  await implementationTab.click();
  await expect(page.getByRole('tabpanel')).toContainText('Move the component into your product package');

  const counter = page.getByRole('group', { name: 'Counter' });
  await counter.getByRole('button', { name: 'Increment', exact: true }).click();
  await expect(counter.getByRole('status')).toHaveText('1');

  const notifications = page.locator('#playground-notifications');
  await expect(notifications).toBeChecked();
  await page.getByText('Enable notifications', { exact: true }).click();
  await expect(notifications).not.toBeChecked();

  const spacious = page.locator('#playground-density--spacious');
  await page.locator('label[for="playground-density--spacious"]').click();
  await expect(spacious).toBeChecked();

  await page.getByText('Show the design rule', { exact: true }).click();
  await expect(page.getByText('Start with a route, give it an explicit order', { exact: false })).toBeVisible();
});

test('documentation article chrome supports breadcrumbs and route paging', async ({ page }) => {
  await page.goto('/docs/getting-started');
  const crumbs = page.locator('nav[aria-label="Breadcrumb"]');
  await expect(crumbs.getByRole('link', { name: 'E2E Docs', exact: true })).toBeVisible();
  await expect(crumbs.getByRole('link', { name: 'Documentation', exact: true })).toBeVisible();
  await expect(crumbs.getByText('Getting started', { exact: true })).toBeVisible();

  const next = page.locator('.ui-doc-layout__next');
  await expect(next).toBeVisible();
  await expect(next).toContainText('The router');
  await next.click();
  await expect(page).toHaveURL(/\/docs\/concepts\/router\/?$/);
  await expect(page.getByRole('heading', { name: 'The router', exact: true })).toBeVisible();

  const previous = page.locator('.ui-doc-layout__prev');
  await expect(previous).toContainText('Getting started');
  await previous.click();
  await expect(page).toHaveURL(/\/docs\/getting-started\/?$/);
});

test('generated Markdown can render a registered typed content component', async ({ page }) => {
  await page.goto('/docs/concepts/content');
  await expect(page.getByText('A typed authoring escape hatch', { exact: true })).toBeVisible();
  await expect(page.getByText('The generated project registers this shared component vocabulary while keeping the page body Markdown.', { exact: true })).toBeVisible();
});

test('top navigation keeps its section active on nested documentation pages', async ({ page }, testInfo) => {
  test.skip(testInfo.project.name === 'mobile-chromium', 'primary nav is collapsed into the mobile drawer');
  await page.goto('/docs/getting-started');
  const nav = page.locator('nav.ui-site-header__links');
  const documentation = nav.getByRole('link', { name: 'Documentation', exact: true });
  const apiReference = nav.getByRole('link', { name: 'API reference', exact: true });

  await expect(documentation).toHaveAttribute('aria-current', 'page');
  await expect(documentation).toHaveClass(/active/);
  await expect(apiReference).not.toHaveAttribute('aria-current', 'page');
});

test('nested documentation is represented once and opens for the active route', async ({ page }, testInfo) => {
  test.skip(testInfo.project.name === 'mobile-chromium', 'desktop sidebar assertion');
  await page.goto('/docs');
  const indexGroup = page.locator('.ui-sidebar__inline details.ui-sidebar__group').filter({ hasText: 'Documentation' });
  await expect(indexGroup).toHaveAttribute('open', '');
  await expect(indexGroup.getByRole('link', { name: 'Getting started', exact: true })).toBeVisible();

  await page.goto('/docs/getting-started');

  const groups = page.locator('.ui-sidebar__inline details.ui-sidebar__group').filter({ hasText: 'Documentation' });
  await expect(groups).toHaveCount(1);
  await expect(groups.locator(':scope > .ui-sidebar__sublist > .ui-sidebar__item')).toHaveCount(5);
  await expect(groups.getByRole('link', { name: 'Getting started', exact: true })).toBeVisible();
  for (const label of ['Concepts', 'Build', 'Operate', 'Collaborate']) {
    await expect(groups.getByText(label, { exact: true })).toBeVisible();
  }
  await expect(groups).toHaveAttribute('open', '');
  await expect(page.locator('.ui-sidebar__inline a.ui-sidebar__link[href="/api-reference"]')).toHaveCount(0);

  await page.goto('/examples/route-tree');
  await expect(page.locator('.ui-sidebar__inline a.ui-sidebar__link[href="/examples/playground"]')).toBeVisible();
  await expect(page.locator('.ui-sidebar__inline details.ui-sidebar__group').filter({ hasText: 'Documentation' })).toHaveCount(0);

  await page.goto('/api-reference');
  await expect(page.locator('.ui-sidebar__inline a.ui-sidebar__link[href="/api-reference"]')).toBeVisible();
  await expect(page.locator('.ui-sidebar__inline details.ui-sidebar__group').filter({ hasText: 'Documentation' })).toHaveCount(0);
});

test('theme choice changes the document and persists across navigation', async ({ page }) => {
  await page.goto('/');
  await page.evaluate(() => localStorage.removeItem('gofastr.colorScheme'));
  await page.reload();
  const root = page.locator('html');
  const before = await root.getAttribute('data-color-scheme');
  expect(['light', 'dark']).toContain(before);
  const mobileHeader = page.locator('summary[aria-label="Toggle navigation"]:visible');
  if (await mobileHeader.count()) {
    await mobileHeader.click();
  }
  await page.locator('[data-fui-theme-toggle]:visible').first().click();
  await expect.poll(() => root.getAttribute('data-color-scheme')).not.toBe(before);
  const after = await root.getAttribute('data-color-scheme');
  await page.goto('/docs');
  await expect(root).toHaveAttribute('data-color-scheme', after);
  await page.reload();
  await expect(root).toHaveAttribute('data-color-scheme', after);
});

test('docs table of contents navigates to a heading', async ({ page }) => {
  await page.goto('/docs/getting-started');
  const tocSelect = page.locator('[data-docs-toc-select]:visible');
  if (await tocSelect.count()) {
    await tocSelect.selectOption('#add-a-page');
    await expect(tocSelect).toHaveValue('#add-a-page');
  } else {
    const tocLink = page.locator('.fastr-docs-toc--rail a[href="#add-a-page"]');
    await expect(tocLink).toBeVisible();
    await tocLink.click();
  }
  await expect(page).toHaveURL(/#add-a-page$/);
  await expect(page.locator('#add-a-page')).toBeInViewport();
});

test('responsive in-page navigation uses a dropdown selector', async ({ page }, testInfo) => {
  test.skip(testInfo.project.name === 'mobile-chromium', 'mobile selector behavior is covered separately');
  await page.setViewportSize({ width: 910, height: 812 });
  await page.goto('/docs/getting-started');
  await expect(page.locator('[data-docs-toc-select]')).toBeVisible();
  await expect(page.locator('.fastr-docs-toc--rail')).toBeHidden();
  const select = page.locator('[data-docs-toc-select]');
  await select.selectOption('#add-a-screen');
  await expect(select).toHaveValue('#add-a-screen');
  await expect(page).toHaveURL(/#add-a-screen$/);
  const selectorBox = await page.locator('.fastr-docs-toc-select').boundingBox();
  const targetBox = await page.locator('#add-a-screen').boundingBox();
  expect((targetBox?.y ?? 0)).toBeGreaterThan((selectorBox?.y ?? 0) + (selectorBox?.height ?? 0));
  await expect(page.locator('#add-a-screen')).toBeInViewport();

  await page.setViewportSize({ width: 740, height: 780 });
  await page.goto('/docs/getting-started');
  await expect(page.locator('.layout-docs .ui-sidebar__inline')).toBeHidden();
  await expect(page.locator('.fastr-docs-mobile-nav-trigger').first()).toBeVisible();
  await expect(page.locator('h1')).toBeInViewport();
  await expect.poll(() => page.evaluate(() => document.documentElement.scrollWidth <= window.innerWidth)).toBe(true);

  await page.setViewportSize({ width: 320, height: 844 });
  await page.goto('/docs/getting-started');
  await expect(page.locator('[data-docs-toc-select]')).toBeVisible();
  await expect.poll(() => page.evaluate(() => document.documentElement.scrollWidth <= window.innerWidth)).toBe(true);
  await expect(page.locator('.fastr-docs-toc-select .ui-select__label')).toHaveCSS('display', 'block');
  await expect(page.locator('.ui-doc-layout__crumbs')).toHaveCSS('padding-top', '0px');
  const narrowSelect = await page.locator('.fastr-docs-toc-select .ui-select__input').boundingBox();
  const narrowCard = await page.locator('.fastr-docs-toc-select').boundingBox();
  expect(narrowSelect?.width ?? 0).toBeGreaterThan((narrowCard?.width ?? 0) - 32);
});

test('white-label docs chrome keeps the sidebar and in-page rail readable', async ({ page }, testInfo) => {
  test.skip(testInfo.project.name === 'mobile-chromium', 'desktop geometry is covered separately on mobile');
  await page.goto('/docs/getting-started');

  const sidebar = page.locator('.layout-docs .layout-body > nav').first();
  const sidebarTitle = page.locator('.layout-docs .ui-sidebar__title').first();
  const sidebarNav = page.locator('.layout-docs .ui-sidebar__nav').first();
  const documentLayout = page.locator('.fastr-docs-doc-layout');
  const toc = page.locator('.fastr-docs-toc');
  const sidebarBox = await sidebar.boundingBox();
  const sidebarTitleBox = await sidebarTitle.boundingBox();
  const sidebarNavBox = await sidebarNav.boundingBox();
  const documentBox = await documentLayout.boundingBox();
  const tocBox = await toc.boundingBox();

  expect(sidebarBox?.width).toBeGreaterThanOrEqual(240);
  expect(sidebarBox?.width).toBeLessThanOrEqual(280);
  expect((sidebarNavBox?.y ?? 0) - (sidebarTitleBox?.y ?? 0)).toBeLessThan(140);
  expect(documentBox?.width).toBeGreaterThan(700);
  const articleBox = await page.locator('.ui-markdown').boundingBox();
  expect(articleBox?.width).toBeGreaterThan(600);
  expect(tocBox?.width).toBeGreaterThanOrEqual(150);
  expect((tocBox?.y ?? 0) - (documentBox?.y ?? 0)).toBeGreaterThanOrEqual(0);
  expect((tocBox?.y ?? 0) - (documentBox?.y ?? 0)).toBeLessThanOrEqual(100);
  await expect(page.getByText('Extensions', { exact: true })).toHaveCount(0);
  await expect(page.getByText('Agent-ready project', { exact: true })).toHaveCount(0);
  await expect(page.getByText('offline shell ready', { exact: true })).toHaveCount(0);
});

test('documentation surfaces keep semantic landmarks and keyboard access', async ({ page }, testInfo) => {
  await page.goto('/docs/getting-started');
  await expect(page.locator('html')).toHaveAttribute('lang', 'pt-BR');
  await expect(page.locator('main')).toHaveCount(1);
  await expect(page.locator('header')).toHaveCount(1);
  await expect(page.locator('nav[aria-label="Breadcrumb"]')).toHaveCount(1);
  await expect(page.getByRole('heading', { level: 1 })).toHaveCount(1);
  await expect(page.locator('img:not([alt])')).toHaveCount(0);
  const unnamedButtons = await page.locator('button').evaluateAll((buttons) => buttons.filter((button) => {
    const label = button.getAttribute('aria-label') || button.textContent || '';
    return !label.trim();
  }).length);
  expect(unnamedButtons).toBe(0);
  if (testInfo.project.name === 'mobile-chromium') {
    await page.locator('[data-fui-open="fastr-docs-sections"]:visible').first().click();
    await expect(page.locator('[data-fui-widget="fastr-docs-sections"]')).toBeVisible();
  } else {
    await page.keyboard.press('Control+K');
    await expect(page.locator('#fastr-docs-command-palette-input:visible')).toBeVisible();
    await page.locator('[data-fui-backdrop="fastr-docs-command-palette"]:visible').click({ position: { x: 4, y: 4 } });
    await expect(page.locator('#fastr-docs-command-palette-input:visible')).toBeHidden();
  }
});
