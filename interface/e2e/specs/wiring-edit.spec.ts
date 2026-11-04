import { test, expect, type Page } from '@playwright/test';

async function openWiringTab(page: Page) {
    await page.addInitScript(() => {
        window.localStorage.setItem('mycelis-advanced-mode', 'true');
    });
    await page.goto('/automations', { waitUntil: 'domcontentloaded' });
    const wiringTab = page.getByRole('button', { name: 'Workflow Builder' });
    await expect(wiringTab).toBeVisible();
    await wiringTab.click();
}

test.describe('Wiring Editor Surface', () => {
    test('wiring tab is reachable in advanced mode', async ({ page }) => {
        await openWiringTab(page);
        await expect(page.getByRole('button', { name: 'Workflow Builder' })).toBeVisible();
        await expect(page.locator('nextjs-portal')).not.toBeVisible();
    });

    test('wiring surface does not crash when mounted', async ({ page }) => {
        await openWiringTab(page);
        await expect(page.locator('body')).toBeVisible();
        await expect(page.locator('nextjs-portal')).not.toBeVisible();
    });
});
