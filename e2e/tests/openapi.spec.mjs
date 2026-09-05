import { test, expect } from '@playwright/test';
import { runtime } from '../support/runtime.mjs';

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

test('OpenAPI reference collects path parameters and JSON request bodies', async ({ page }) => {
  await page.goto(`${runtime().manualURL}/api-reference`);
  const select = page.locator('[data-openapi-operation-select]');
  const response = page.locator('[data-openapi-response]');

  await select.selectOption({ label: 'GET · /projects/{id}' });
  const pathInput = page.locator('[data-openapi-inputs-for] [data-openapi-param-name="id"]');
  await expect(pathInput).toBeVisible();
  await pathInput.fill('prj_e2e');
  await page.locator('[data-openapi-try]').click();
  await expect(response).toContainText('200');
  await expect(response).toContainText('E2E project detail');
  await expect(response).toContainText('/v1/projects/prj_e2e');

  await select.selectOption({ label: 'GET · /projects' });
  const limitInput = page.locator('[data-openapi-inputs-for="fastr-openapi-operation-1"] [data-openapi-param-name="limit"]');
  await limitInput.fill('5');
  await page.locator('[data-openapi-try]').click();
  await expect(response).toContainText('200');
  await expect(response).toContainText('/v1/projects?limit=5');

  await select.selectOption({ label: 'POST · /projects' });
  const body = page.locator('[data-openapi-inputs-for] [data-openapi-body]');
  await expect(body).toBeVisible();
  await body.fill('{"name":"Created from docs"}');
  await page.locator('[data-openapi-try]').click();
  await expect(response).toContainText('201');
  await expect(response).toContainText('Created from docs');
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
