import { test, expect } from '@playwright/test';

test('docs page stays within its approved responsive visual surface', async ({ page }, testInfo) => {
  await page.goto('/docs/getting-started');
  const snapshot = testInfo.project.name === 'mobile-chromium'
    ? 'docs-getting-started-mobile.png'
    : 'docs-getting-started-desktop.png';
  await expect(page).toHaveScreenshot(snapshot, {
    animations: 'disabled',
    caret: 'hide',
    fullPage: true,
  });
});
