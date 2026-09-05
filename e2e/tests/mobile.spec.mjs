import { test, expect } from '@playwright/test';

test.use({ viewport: { width: 390, height: 844 }, isMobile: true });

test('mobile navigation opens the full route drawer and closes after navigation', async ({ page }) => {
  await page.goto('/docs/getting-started');
  const trigger = page.locator('[data-fui-open="fastr-docs-sections"]:visible').first();
  await expect(trigger).toBeVisible();
  await trigger.click();

  const drawer = page.locator('[data-fui-widget="fastr-docs-sections"]');
  await expect(drawer).toBeVisible();
  await expect(drawer.getByRole('link', { name: 'API reference', exact: true })).toBeVisible();
  await drawer.getByRole('link', { name: 'API reference', exact: true }).click();
  await expect(page).toHaveURL(/\/api-reference\/?$/);
  await expect(page.getByRole('heading', { name: 'API reference', exact: true })).toBeVisible();
  await expect(drawer).toBeHidden();
});

test('mobile route paging opens the active nested section when the drawer opens', async ({ page }) => {
  await page.goto('/docs/getting-started');
  await page.locator('.ui-doc-layout__next').click();
  await expect(page).toHaveURL(/\/docs\/concepts\/router\/?$/);

  await page.locator('[data-fui-open="fastr-docs-sections"]:visible').first().click();
  const drawer = page.locator('[data-fui-widget="fastr-docs-sections"]');
  const conceptsSummary = drawer.locator('details.ui-sidebar__group > summary').filter({ hasText: /^Concepts$/ });
  const conceptsGroup = conceptsSummary.locator('..');
  await expect(conceptsGroup).toHaveAttribute('open', '');
  await expect(conceptsGroup.getByRole('link', { name: 'The router', exact: true })).toHaveAttribute('aria-current', 'page');
});

test('mobile keeps the in-page navigation available', async ({ page }) => {
  await page.goto('/docs/getting-started');
  const toc = page.locator('[data-docs-toc-select]');
  await expect(toc).toBeVisible();
  await toc.selectOption('#add-a-page');
  await expect(toc).toHaveValue('#add-a-page');
  await expect(page).toHaveURL(/#add-a-page$/);
  await expect(page.locator('#add-a-page')).toBeInViewport();
});

test('mobile command palette exposes a visible close control without overflowing', async ({ page }) => {
  await page.goto('/docs/build/plugins');
  await page.locator('.fastr-docs-command-trigger:visible').first().click();

  const palette = page.locator('[data-fui-widget="fastr-docs-command-palette"]:visible');
  await expect(palette).toBeVisible();
  const close = palette.getByRole('button', { name: 'Close search', exact: true });
  await expect(close).toBeVisible();
  await expect.poll(() => page.evaluate(() => document.documentElement.scrollWidth <= window.innerWidth)).toBe(true);

  await close.click();
  await expect(palette).toBeHidden();
});

test('mobile rebinds the in-page navigation after drawer navigation', async ({ page }) => {
  await page.goto('/docs/operate/offline');
  await page.locator('[data-fui-open="fastr-docs-sections"]:visible').first().click();
  const drawer = page.locator('[data-fui-widget="fastr-docs-sections"]');
  await expect(drawer).toBeVisible();
  await drawer.getByRole('link', { name: 'Search', exact: true }).click();
  await expect(page).toHaveURL(/\/docs\/operate\/search\/?$/);

  const toc = page.locator('select[data-docs-toc-select]:visible');
  await expect(toc).toBeVisible();
  await toc.selectOption('#improve-result-quality');
  await expect(page).toHaveURL(/#improve-result-quality$/);
  await expect(toc).toHaveValue('#improve-result-quality');
});

test('mobile keeps white-label chrome free of implementation metadata', async ({ page }) => {
  await page.goto('/docs/getting-started');
  const toc = page.locator('.fastr-docs-toc-select');
  const tocBox = await toc.boundingBox();
  expect(tocBox?.width).toBeGreaterThan(300);
  await expect(page.getByText('Extensions', { exact: true })).toHaveCount(0);
  await expect(page.getByText('Agent-ready project', { exact: true })).toHaveCount(0);
  await expect(page.getByText('offline shell ready', { exact: true })).toHaveCount(0);
});

test('mobile generated playground keeps every local control reachable', async ({ page }) => {
  await page.goto('/examples/playground');
  await expect(page.getByTestId('examples-playground')).toBeVisible();
  await page.getByRole('tab', { name: 'Implementation', exact: true }).click();
  await expect(page.getByRole('tabpanel')).toContainText('Move the component into your product package');

  const counter = page.getByRole('group', { name: 'Counter' });
  await counter.getByRole('button', { name: 'Increment', exact: true }).click();
  await expect(counter.getByRole('status')).toHaveText('1');
  await page.getByText('Enable notifications', { exact: true }).click();
  await expect(page.locator('#playground-notifications')).not.toBeChecked();
  await page.locator('label[for="playground-density--spacious"]').click();
  await expect(page.locator('#playground-density--spacious')).toBeChecked();
  await expect.poll(() => page.evaluate(() => document.documentElement.scrollWidth <= window.innerWidth)).toBe(true);
});
