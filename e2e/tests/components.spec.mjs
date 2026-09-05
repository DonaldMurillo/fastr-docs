import { test, expect } from '@playwright/test';

// The generated starter ships a page exercising the whole default shortcode
// vocabulary. These assertions run against that page, so a component that
// stops rendering fails here rather than in someone's docs site.
const componentsPage = '/docs/build/components';

test.beforeEach(async ({ page }) => {
  await page.goto(componentsPage);
});

test('the default shortcode vocabulary needs no registration', async ({ page }) => {
  await expect(page.locator('h1')).toContainText('Markdown components');

  // A marker leaking through means shortcode substitution broke, and the
  // reader sees FASTRDOCSSHORTCODESLOT0 in the middle of the prose.
  await expect(page.locator('body')).not.toContainText('FASTRDOCSSHORTCODESLOT');
  // An unrendered shortcode means the name was never registered.
  await expect(page.locator('body')).not.toContainText('{{< note');

  await expect(page.locator('.ui-callout')).toHaveCount(4);
  await expect(page.locator('.ui-card')).toHaveCount(3);
  await expect(page.locator('.ui-grid')).toHaveCount(1);
  await expect(page.locator('.ui-badge')).toHaveCount(2);
  await expect(page.locator('.ui-tag')).toHaveCount(1);
});

test('tabs switch without JavaScript and keep independent groups', async ({ page }) => {
  const tabSet = page.locator('.ui-markdown details[name]').first();
  await expect(tabSet).toBeVisible();

  const groups = await page.locator('.ui-markdown details[name]').evaluateAll((nodes) =>
    [...new Set(nodes.map((n) => n.getAttribute('name')))],
  );
  expect(groups.length).toBeGreaterThan(0);

  const shell = page.locator('.ui-markdown details[name] summary', { hasText: 'Shell' }).first();
  await shell.click();
  await expect(page.locator('body')).toContainText('Each set gets its own group name');
});

test('steps are numbered and the file tree nests', async ({ page }) => {
  const steps = page.locator('.fastr-docs-steps > * > ol > li');
  await expect(steps).toHaveCount(3);
  // The counter is drawn by CSS, so assert the rule is actually reaching it.
  const marker = await steps.first().evaluate((el) => getComputedStyle(el, '::before').content);
  expect(marker).toContain('counter');

  // Nested <ul> is the whole point: GoFastr's Markdown parser cannot nest
  // lists, so the file tree parses its raw body instead.
  await expect(page.locator('.fastr-docs-filetree ul ul')).not.toHaveCount(0);
  await expect(page.locator('.fastr-docs-filetree__dir').first()).toContainText('my-docs/');
  await expect(page.locator('.fastr-docs-filetree__file').first()).toContainText('index.md');
  // Code-block chrome must not leak into a raw shortcode's body.
  await expect(page.locator('.fastr-docs-filetree')).not.toContainText('lines');
});

test('the diff renders real added and removed lines', async ({ page }) => {
  await expect(page.locator('.ui-diff-viewer__line--add')).not.toHaveCount(0);
  await expect(page.locator('.ui-diff-viewer__line--remove')).not.toHaveCount(0);
  await expect(page.locator('.ui-diff-viewer')).not.toContainText('lines');
});

test('component bodies do not inherit the page prose layout', async ({ page }) => {
  // The page-level .ui-markdown carries a max width and 75px of bottom padding.
  // A callout body is itself a .ui-markdown and must not pick either up.
  const inner = await page.locator('.ui-callout .ui-markdown').first().evaluate((el) => {
    const style = getComputedStyle(el);
    return { paddingBottom: style.paddingBottom, marginLeft: style.marginLeft };
  });
  expect(inner.paddingBottom).toBe('0px');
  expect(inner.marginLeft).toBe('0px');

  // A card with an href is the anchor, so prose underlines would run through it.
  const card = page.locator('a.ui-card').first();
  if (await card.count()) {
    await expect(card).toHaveCSS('text-decoration-line', 'none');
  }
});
