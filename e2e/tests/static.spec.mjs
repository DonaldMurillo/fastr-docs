import { test, expect } from '@playwright/test';
import { runtime } from '../support/runtime.mjs';

test('static export keeps the navigable docs and PWA surfaces', async ({ request }) => {
  const { staticURL } = runtime();
  const pages = [
    ['/', 'fastr-docs-home__hero'],
    ['/docs/', 'Documentation'],
    ['/docs/getting-started/', 'Getting started'],
    ['/docs/build/framework-ui/', 'Framework UI'],
    ['/docs/operate/feature-coverage/', 'Feature coverage'],
    ['/blog/', 'Your first post'],
    ['/blog/first-post/', 'Your first post'],
    ['/api-reference/', 'data-openapi-try'],
  ];
  for (const [path, marker] of pages) {
    const response = await request.get(staticURL + path);
    expect(response.ok(), `${path} should be served`).toBeTruthy();
    const html = await response.text();
    expect(html).toContain(marker);
    expect(html).toContain('lang="pt-BR"');
  }

  const notFound = await request.get(staticURL + '/404.html');
  expect(notFound.ok()).toBeTruthy();
  expect(await notFound.text()).toContain('Page not found');
  const notFoundStyles = await request.get(staticURL + '/404.css');
  expect(notFoundStyles.ok()).toBeTruthy();
  expect(notFoundStyles.headers()['content-type']).toContain('text/css');

  const missingRoute = await request.get(staticURL + '/does-not-exist/');
  expect(missingRoute.status()).toBe(404);
  const malformedRoute = await request.get(staticURL + '/%E0%A4%A');
  expect(malformedRoute.status()).toBe(400);
  expect((await request.get(staticURL + '/docs/')).ok()).toBeTruthy();

  const feed = await request.get(staticURL + '/blog/feed.xml');
  expect(feed.ok()).toBeTruthy();
  expect(feed.headers()['content-type']).toContain('application/rss+xml');
  expect(await feed.text()).toContain('Your first post');

  const manifest = await request.get(staticURL + '/manifest.webmanifest');
  expect(manifest.ok()).toBeTruthy();
  expect(await manifest.text()).toContain('icon-192.png');

  const registration = await request.get(staticURL + '/__gofastr/pwa/register.js');
  expect(registration.ok()).toBeTruthy();
  expect(await registration.text()).toContain('gofastr:pwa-update');
  const offlineScreen = await request.get(staticURL + '/__gofastr/pwa/offline/');
  expect(offlineScreen.ok()).toBeTruthy();
  expect((await offlineScreen.text()).toLowerCase()).toContain('offline');

  const serviceWorker = await request.get(staticURL + '/service-worker.js');
  expect(serviceWorker.ok()).toBeTruthy();
  const serviceWorkerText = await serviceWorker.text();
  expect(serviceWorkerText).toContain('gofastr-pwa-static-');
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
  // RewriteStaticCSP HTML-escapes the policy, so read the attribute and decode
  // it instead of matching the raw source.
  const cspTag = (await apiPage.text())
    .match(/<meta[^>]*>/gi)
    ?.find((tag) => /http-equiv\s*=\s*"Content-Security-Policy"/i.test(tag));
  expect(cspTag).toBeTruthy();
  const policy = /content\s*=\s*"([^"]*)"/i.exec(cspTag)[1].replace(/&#39;/g, "'").replace(/&amp;/g, '&');
  expect(policy).toContain("connect-src 'self' http://127.0.0.1:4176");
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

test('static blog search filters the publication without a server round trip', async ({ page }) => {
  const { staticURL } = runtime();
  await page.goto(staticURL + '/blog/search');
  const input = page.locator('[data-fastr-docs-blog-search-input]');
  await expect(input).toBeVisible();
  await input.fill('first');
  await page.getByRole('button', { name: 'Search', exact: true }).click();
  await expect(page).toHaveURL(/\/blog\/search\?q=first$/);
  await expect(page.locator('[data-fastr-docs-blog-search-item]:not([hidden])')).toHaveCount(1);
  await expect(page.locator('[data-fastr-docs-blog-search-item]:not([hidden])')).toContainText('Your first post');
  await expect(page.locator('[data-fastr-docs-blog-search-summary]')).toContainText('1 matches');
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
    ['/docs/build/framework-ui/', 'Framework UI'],
    ['/docs/operate/feature-coverage/', 'Feature coverage'],
    ['/blog/', 'Blog'],
    ['/blog/first-post/', 'Your first post'],
    ['/examples/playground/', 'Playground'],
    ['/examples/custom-surface/', 'Custom surfaces'],
    ['/api-reference/', 'API reference'],
  ];
  for (const [path, heading] of routes) {
    await page.goto(staticURL + path);
    await expect(page.locator('.layout-content h1', { hasText: heading })).toBeVisible();
    await expect(page.locator('html')).toHaveAttribute('lang', 'pt-BR');
    await expect.poll(() => page.evaluate(() => document.documentElement.scrollWidth <= window.innerWidth)).toBe(true);
  }
});
