import { test, expect } from '@playwright/test';
import { runtime } from '../support/runtime.mjs';

test('static export keeps the navigable docs and PWA surfaces', async ({ request }) => {
  const { staticURL } = runtime();
  const pages = [
    ['/', 'fastr-docs-home__hero'],
    ['/docs/', 'Documentation'],
    ['/docs/getting-started/', 'Getting started'],
    ['/api-reference/', 'data-openapi-try'],
  ];
  for (const [path, marker] of pages) {
    const response = await request.get(staticURL + path);
    expect(response.ok(), `${path} should be served`).toBeTruthy();
    const html = await response.text();
    expect(html).toContain(marker);
    expect(html).toContain('lang="pt-BR"');
  }

  const manifest = await request.get(staticURL + '/manifest.webmanifest');
  expect(manifest.ok()).toBeTruthy();
  expect(await manifest.text()).toContain('icon-192.png');

  const serviceWorker = await request.get(staticURL + '/service-worker.js');
  expect(serviceWorker.ok()).toBeTruthy();
  const serviceWorkerText = await serviceWorker.text();
  expect(serviceWorkerText).toContain('/__fastr-docs/search.json');
  expect(serviceWorkerText).toContain('/pagefind/pagefind.js');
  expect(serviceWorkerText).toContain('fastr-docs-sections/chrome');

  const exportManifest = await request.get(staticURL + '/__fastr-docs/manifest.json');
  expect(exportManifest.ok()).toBeTruthy();
  const exportManifestText = await exportManifest.text();
  expect(exportManifestText).toContain('fastr-docs/v1');
  expect(exportManifestText).toContain('"searchBackend": "pagefind"');
  const pagefindRuntime = await request.get(staticURL + '/pagefind/pagefind.js');
  expect(pagefindRuntime.ok()).toBeTruthy();

  const llms = await request.get(staticURL + '/llms.txt');
  expect(llms.ok()).toBeTruthy();
  expect(await llms.text()).toContain('## When to use');
  const agentCard = await request.get(staticURL + '/.well-known/agent-card.json');
  expect(agentCard.ok()).toBeTruthy();
  expect(await agentCard.text()).toContain('E2E Docs');

  const apiPage = await request.get(staticURL + '/api-reference/');
  expect(apiPage.ok()).toBeTruthy();
  expect(await apiPage.text()).toContain("connect-src 'self' http://127.0.0.1:4176");
});

test('static Pagefind search feeds the native command palette', async ({ page }) => {
  const { staticURL } = runtime();
  await page.goto(staticURL + '/docs/');
  await page.keyboard.press('Control+K');
  const input = page.locator('#fastr-docs-command-palette-input:visible').first();
  await expect(input).toBeVisible();
  await input.fill('getting started');
  const result = page.locator('[role="option"][data-fui-push-state*="/docs/getting-started"]:visible').first();
  await expect(result).toContainText('Getting started');
  await result.click();
  await expect(page).toHaveURL(/\/docs\/getting-started\/?$/);
});

test('static content remains navigable after the network is unavailable', async ({ page }) => {
  const { staticURL } = runtime();
  await page.goto(staticURL + '/docs/getting-started/');
  await expect(page.getByRole('heading', { name: 'Getting started', exact: true })).toBeVisible();
  await page.evaluate(async () => {
    if (!('serviceWorker' in navigator)) throw new Error('service workers are unavailable');
    await navigator.serviceWorker.ready;
  });
  await page.context().setOffline(true);
  try {
    await page.reload();
    await expect(page.getByRole('heading', { name: 'Getting started', exact: true })).toBeVisible();
    await expect(page.locator('.fastr-docs-site-header')).toBeVisible();
  } finally {
    await page.context().setOffline(false);
  }
});

test('static export renders every generated route family in a real browser', async ({ page }) => {
  const { staticURL } = runtime();
  const routes = [
    ['/', 'A docs framework that starts as a router.'],
    ['/docs/', 'E2E Docs'],
    ['/examples/playground/', 'Playground'],
    ['/examples/custom-surface/', 'Custom surfaces'],
    ['/api-reference/', 'API reference'],
  ];
  for (const [path, heading] of routes) {
    await page.goto(staticURL + path);
    await expect(page.getByRole('heading', { name: heading, exact: true })).toBeVisible();
    await expect(page.locator('html')).toHaveAttribute('lang', 'pt-BR');
    await expect.poll(() => page.evaluate(() => document.documentElement.scrollWidth <= window.innerWidth)).toBe(true);
  }
});
