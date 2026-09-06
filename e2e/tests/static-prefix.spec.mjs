import { test, expect } from '@playwright/test';
import { runtime } from '../support/runtime.mjs';

// The site exported below a prefix, the shape of a GitHub project page
// (user.github.io/repo/). The export rewrites the URL attributes; everything
// the runtime does with a path had to learn the base too: the language
// selector, search results, and the active state in the header and sidebar.

const prefixed = (path) => {
  const { selfPrefixURL, selfPrefix } = runtime();
  return selfPrefixURL + selfPrefix + path;
};

test('an export below a prefix serves every kind of page there', async ({ request }) => {
  for (const path of ['/', '/docs/getting-started/', '/es/docs/getting-started/', '/es/blog/', '/es/blog/un-arbol-para-docs-y-publicacion/', '/404.html']) {
    const response = await request.get(prefixed(path));
    expect(response.ok(), path).toBeTruthy();
    const html = await response.text();
    // The 404 is a bare screen with no search trigger, so it carries no base
    // attribute; every routed page does.
    if (path !== '/404.html') expect(html, path).toContain('data-fastr-docs-base="/prefix"');
    // No URL attribute escaped the rewrite.
    expect(html, path).not.toMatch(/(href|src)="\/(?!prefix\/|prefix")/);
  }
});

// GoFastr's runtime loads a component's CSS when no link carries the
// component's marker. The export wrote the links without it, so every page
// got a second copy of each component stylesheet after the site's own CSS,
// and the docs overrides lost the cascade: the sidebar title stretched to
// fill the column.
test('a static page does not load its component stylesheets twice', async ({ page }) => {
  await page.goto(prefixed('/docs/getting-started/'), { waitUntil: 'networkidle' });
  const sheets = await page.evaluate(() => [...document.querySelectorAll('link[rel="stylesheet"]')].map((l) => l.getAttribute('href').split('?')[0]));
  expect(new Set(sheets).size, sheets.join('\n')).toBe(sheets.length);
  // The persistent sidebar is hidden on the mobile project, where the
  // title measures 0; either way it must not be stretched.
  const titleHeight = await page.evaluate(() => document.querySelector('.ui-sidebar__title')?.offsetHeight ?? 0);
  expect(titleHeight).toBeLessThan(80);
});

test('the language selector navigates below the prefix', async ({ page }) => {
  await page.goto(prefixed('/es/docs/getting-started/'));
  await expect(page.locator('h1')).toContainText('Primeros pasos');
  // The option holds the route path; the runtime adds the base when it
  // navigates, and it also recognised this page as the selected option.
  await expect(page.locator('[data-docs-variant-select=locale]').first()).toHaveValue('/es/docs/getting-started');
  await page.selectOption('[data-docs-variant-select=locale]', { label: 'English' });
  await page.waitForURL('**/prefix/docs/getting-started');
  await expect(page.locator('h1')).toContainText('Getting started');
  await expect(page.locator('html')).toHaveAttribute('lang', 'en');
});

test('search results below the prefix link below the prefix', async ({ page }) => {
  await page.goto(prefixed('/es/docs/build/math/'), { waitUntil: 'networkidle' });
  await page.locator('.fastr-docs-command-trigger:visible').first().click();
  await page.locator('#fastr-docs-command-palette-input:visible').first().fill('router');
  // The palette lists every route until the index answers; wait for the
  // Pagefind results, which carry their own option ids.
  await expect.poll(async () => page.evaluate(() =>
    document.querySelectorAll('[role="option"][id^="fastr-docs-command-palette-list-opt-pagefind-"]').length), { timeout: 15_000 }).toBeGreaterThan(0);
  const targets = await page.evaluate(() =>
    [...document.querySelectorAll('[role="option"][id^="fastr-docs-command-palette-list-opt-pagefind-"]')].map((o) => o.getAttribute('data-fui-push-state')));
  expect(targets.every((u) => u.startsWith('/prefix/es/') || u === '/prefix/es')).toBe(true);
  // Following one lands on a page, not a 404.
  await page.locator('[role="option"][id^="fastr-docs-command-palette-list-opt-pagefind-"]').first().click();
  await expect.poll(() => page.url()).toMatch(/\/prefix\/es\//);
  await expect(page.locator('h1')).not.toHaveText(/no encontrada|not found/i);
});

test('the sidebar and header know which page is active below the prefix', async ({ page }) => {
  await page.goto(prefixed('/es/docs/getting-started/'));
  // The runtime applies both after load, so poll rather than read once.
  await expect.poll(() => page.evaluate(() =>
    [...document.querySelectorAll('.ui-sidebar__link[aria-current="page"]')].map((a) => a.getAttribute('href')))).toContain('/prefix/es/docs/getting-started');
  // The desktop nav and the mobile drawer nav are both in the page; the
  // project decides which one is visible.
  await expect.poll(() => page.evaluate(() =>
    [...document.querySelectorAll('.ui-site-header__links a[aria-current="page"], .ui-site-header__mobile-links a[aria-current="page"]')].map((a) => a.textContent.trim()))).toContain('Documentación');
});
