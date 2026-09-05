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
