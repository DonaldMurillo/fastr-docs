import { test, expect } from '@playwright/test';

// The generated starter's home page is a splash page built entirely from front
// matter, which is the whole point: a landing page with no Go.
test('the generated home page is a front-matter splash page', async ({ page }) => {
  await page.goto('/');

  const hero = page.locator('.fastr-docs-page-hero');
  await expect(hero).toBeVisible();
  await expect(hero.locator('h1')).toContainText('E2E Docs');
  await expect(hero).toContainText('Documentation built on GoFastr');

  const actions = hero.locator('a');
  await expect(actions).toHaveCount(2);
  await expect(actions.first()).toHaveAttribute('href', '/docs/getting-started');

  // Front matter must never reach the reader as text.
  await expect(page.locator('body')).not.toContainText('template: splash');
  await expect(page.locator('body')).not.toContainText('FASTRDOCSSHORTCODESLOT');
});

test('a splash page drops the table of contents and breadcrumbs', async ({ page }, testInfo) => {
  await page.goto('/');
  await expect(page.locator('.fastr-docs-splash')).toHaveCount(1);
  await expect(page.locator('.fastr-docs-toc')).toHaveCount(0);
  await expect(page.locator('.ui-doc-layout__crumbs')).toHaveCount(0);

  // Reclaiming the reserved table-of-contents column is what makes the landing
  // column wider than a reference page's. Only meaningful on desktop.
  test.skip(testInfo.project.name === 'mobile-chromium', 'single column on mobile');
  const splashWidth = await page.locator('.ui-markdown').first().evaluate((el) => el.getBoundingClientRect().width);
  await page.goto('/docs/getting-started');
  const normalWidth = await page.locator('.ui-markdown').first().evaluate((el) => el.getBoundingClientRect().width);
  expect(splashWidth).toBeGreaterThan(normalWidth);
});

test('the body below the hero is ordinary Markdown with shortcodes', async ({ page }) => {
  await page.goto('/');
  await expect(page.locator('.ui-card')).toHaveCount(3);
  await expect(page.locator('.fastr-docs-steps > * > ol > li')).toHaveCount(3);
});

test('the typed landing page remains available as the Go alternative', async ({ page }) => {
  await page.goto('/examples/typed-landing');
  await expect(page.locator('h1').first()).toBeVisible();
  await expect(page.locator('.fastr-docs-home')).toHaveCount(1);
});
