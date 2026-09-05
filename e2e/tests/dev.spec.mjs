import fs from 'node:fs/promises';
import path from 'node:path';
import { test, expect } from '@playwright/test';
import { runtime } from '../support/runtime.mjs';

test('fastr-docs dev reloads OpenAPI contract changes through GoFastr', async ({ page, request }) => {
  const { devURL, target } = runtime();
  const contractPath = path.join(target, 'openapi.json');
  const original = await fs.readFile(contractPath, 'utf8');
  const spec = JSON.parse(original);
  spec.paths['/v1/dev-reload-check'] = {
    get: {
      summary: 'Development reload check',
      operationId: 'devReloadCheck',
      responses: { 200: { description: 'Reload check response.' } },
    },
  };

  try {
    await page.goto(`${devURL}/api-reference`);
    await expect(page.locator('[data-openapi-reference]')).toBeVisible();
    await expect(page.locator('[data-openapi-operation]').filter({ hasText: 'devReloadCheck' })).toHaveCount(0);

    await fs.writeFile(contractPath, JSON.stringify(spec, null, 2));
    await expect.poll(async () => {
      try {
        const response = await request.get(`${devURL}/api-reference`);
        return response.ok() && (await response.text()).includes('devReloadCheck');
      } catch {
        // GoFastr briefly closes the child server while rebuilding.
        return false;
      }
    }, { timeout: 20_000 }).toBe(true);

    await page.reload();
    await expect(page.locator('[data-openapi-operation]').filter({ hasText: 'devReloadCheck' })).toBeVisible();
  } finally {
    await fs.writeFile(contractPath, original);
  }
});

test('fastr-docs dev reloads Markdown collection files through GoFastr', async ({ page, request }) => {
  const { devURL, target } = runtime();
  const contentPath = path.join(target, 'content', 'getting-started.md');
  const original = await fs.readFile(contentPath, 'utf8');
  const marker = 'Markdown reload check';

  try {
    await page.goto(`${devURL}/docs/getting-started`);
    await expect(page.getByRole('heading', { name: 'Getting started', exact: true })).toBeVisible();

    await fs.writeFile(contentPath, `${original}\n\n## Reload check\n\n${marker}.`);
    await expect.poll(async () => {
      try {
        const response = await request.get(`${devURL}/docs/getting-started`);
        return response.ok() && (await response.text()).includes(marker);
      } catch {
        return false;
      }
    }, { timeout: 20_000 }).toBe(true);

    await page.reload();
    await expect(page.getByRole('heading', { name: 'Reload check', exact: true })).toBeVisible({ timeout: 20_000 });
  } finally {
    await fs.writeFile(contentPath, original);
  }
});
