import { test, expect } from '@playwright/test';

test('OpenAPI reference filters operations and sends a request to the configured server', async ({ page }) => {
  await page.goto('/api-reference');
  await expect(page.getByRole('heading', { name: 'API reference', exact: true })).toBeVisible();
  await expect(page.locator('[data-openapi-server-url]')).toHaveAttribute('data-openapi-server-url', /4176\/v1/);

  const filter = page.locator('[data-openapi-filter]');
  const operation = page.locator('[data-openapi-operation]').first();
  await filter.fill('does-not-exist');
  await expect(operation).toBeHidden();
  await filter.fill('listProjects');
  await expect(operation).toBeVisible();

  await page.locator('[data-openapi-operation-select]').selectOption({ index: 0 });
  await page.locator('[data-openapi-try]').click();
  const response = page.locator('[data-openapi-response]');
  await expect(response).toContainText('200');
  await expect(response).toContainText('E2E project');
  await expect(response).not.toContainText('/v1/v1/projects');
});

test('OpenAPI reference keeps its index, console, and endpoint cards usable responsively', async ({ page }, testInfo) => {
  await page.goto('/api-reference');
  await expect(page.locator('.fastr-openapi-reference__index')).toBeVisible();
  await expect(page.locator('.fastr-openapi-reference__console')).toBeVisible();
  await expect(page.locator('[data-openapi-operation]')).not.toHaveCount(0);

  const layoutDisplay = await page.locator('.fastr-openapi-reference__layout').evaluate((node) => getComputedStyle(node).display);
  if (testInfo.project.name === 'mobile-chromium') {
    expect(layoutDisplay).toBe('block');
  } else {
    expect(layoutDisplay).toBe('grid');
  }
  await expect.poll(() => page.evaluate(() => document.documentElement.scrollWidth <= window.innerWidth)).toBe(true);
});
