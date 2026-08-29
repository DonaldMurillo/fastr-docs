import { test, expect } from '@playwright/test';

test('desktop docs page stays within the approved visual surface', async ({ page }, testInfo) => {
  test.skip(testInfo.project.name !== 'chromium', 'the mobile project has its own responsive visual baseline');
  await page.goto('/docs/getting-started');
  await expect(page).toHaveScreenshot('docs-getting-started-desktop.png', {
    animations: 'disabled',
    caret: 'hide',
    fullPage: true,
  });
});

test('mobile docs page stays within the approved responsive visual surface', async ({ page }, testInfo) => {
  test.skip(testInfo.project.name !== 'mobile-chromium', 'the desktop project has its own visual baseline');
  await page.goto('/docs/getting-started');
  await expect(page).toHaveScreenshot('docs-getting-started-mobile.png', {
    animations: 'disabled',
    caret: 'hide',
    fullPage: true,
  });
});
