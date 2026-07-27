import { test, expect } from '@playwright/test';
import type { Page } from '@playwright/test';

async function openWiringTab(page: Page) {
    await page.goto('/automations');
    await page.waitForLoadState('domcontentloaded');
    const advancedOff = page.getByRole('button', { name: 'Admin tools: Off' });
    if (await advancedOff.isVisible()) {
        await advancedOff.click();
    }
    const wiringTab = page.getByRole('button', { name: 'Workflow Builder' });
    await expect(wiringTab, 'Workflow Builder must be available when admin tools are enabled').toBeVisible();
    await wiringTab.click();
    await expect(page).toHaveURL(/\/automations/);
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
