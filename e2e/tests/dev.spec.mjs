import fs from 'node:fs/promises';
import path from 'node:path';
import { test, expect } from '@playwright/test';
import { runtime } from '../support/runtime.mjs';

// GoFastr's dev loop restarts the child server on every rebuild, including the
// one triggered when a test restores its fixture file in `finally`. The next
// test then starts against a server that is still coming back, and its very
// first navigation fails. Both helpers below exist for that, not for any
// product behaviour.
// These tests wait on real rebuilds of a growing site, so the default 30s
// budget is not enough: the wait alone could consume it.
test.describe.configure({ timeout: 120_000 });

const waitForDevServer = async (request, url) => {
  await expect.poll(async () => {
    try {
      return (await request.get(url)).ok();
    } catch {
      return false;
    }
  }, { timeout: 60_000, intervals: [250] }).toBe(true);
};

const reloadThroughDevLoop = async (page, url) => {
  for (let attempt = 0; attempt < 20; attempt++) {
    try {
      await page.goto(url, { waitUntil: 'load' });
      return;
    } catch (error) {
      if (attempt === 19) throw error;
      await page.waitForTimeout(500);
    }
  }
};

test('fastr-docs dev reloads OpenAPI spec changes through GoFastr', async ({ page, request }) => {
  const { devURL, target } = runtime();
  const specPath = path.join(target, 'openapi.json');
  const original = await fs.readFile(specPath, 'utf8');
  const spec = JSON.parse(original);
  spec.paths['/v1/dev-reload-check'] = {
    get: {
      summary: 'Development reload check',
      operationId: 'devReloadCheck',
      responses: { 200: { description: 'Reload check response.' } },
    },
  };

  await waitForDevServer(request, `${devURL}/api-reference`);

  try {
    await reloadThroughDevLoop(page, `${devURL}/api-reference`);
    await expect(page.locator('[data-openapi-reference]')).toBeVisible();
    await expect(page.locator('[data-openapi-operation]').filter({ hasText: 'devReloadCheck' })).toHaveCount(0);

    await fs.writeFile(specPath, JSON.stringify(spec, null, 2));
    await expect.poll(async () => {
      try {
        const response = await request.get(`${devURL}/api-reference`);
        return response.ok() && (await response.text()).includes('devReloadCheck');
      } catch {
        // GoFastr briefly closes the child server while rebuilding.
        return false;
      }
    }, { timeout: 60_000 }).toBe(true);

    await reloadThroughDevLoop(page, `${devURL}/api-reference`);
    await expect(page.locator('[data-openapi-operation]').filter({ hasText: 'devReloadCheck' })).toBeVisible({ timeout: 60_000 });
  } finally {
    await fs.writeFile(specPath, original);
  }
});

test('fastr-docs dev reloads Markdown collection files through GoFastr', async ({ page, request }) => {
  const { devURL, target } = runtime();
  const contentPath = path.join(target, 'content', 'getting-started.md');
  const original = await fs.readFile(contentPath, 'utf8');
  const marker = 'Markdown reload check';

  await waitForDevServer(request, `${devURL}/docs/getting-started`);

  try {
    await reloadThroughDevLoop(page, `${devURL}/docs/getting-started`);
    await expect(page.getByRole('heading', { name: 'Getting started', exact: true })).toBeVisible();

    await fs.writeFile(contentPath, `${original}\n\n## Reload check\n\n${marker}.`);
    await expect.poll(async () => {
      try {
        const response = await request.get(`${devURL}/docs/getting-started`);
        return response.ok() && (await response.text()).includes(marker);
      } catch {
        return false;
      }
    }, { timeout: 60_000 }).toBe(true);

    await reloadThroughDevLoop(page, `${devURL}/docs/getting-started`);
    await expect(page.getByRole('heading', { name: 'Reload check', exact: true })).toBeVisible({ timeout: 60_000 });
  } finally {
    await fs.writeFile(contentPath, original);
  }
});
