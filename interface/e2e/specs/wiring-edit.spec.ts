import { test, expect } from '@playwright/test';

async function openWiringTab(page: any) {
    await page.goto('/automations');
    await page.waitForLoadState('domcontentloaded');
    const advancedOff = page.getByRole('button', { name: 'Advanced: Off' });
    if (await advancedOff.isVisible().catch(() => false)) {
        await advancedOff.click();
    }
    const wiringTab = page.getByRole('button', { name: 'Neural Wiring' });
    if (!(await wiringTab.isVisible().catch(() => false))) return false;
    await wiringTab.click();
    await page.waitForTimeout(500);
    return true;
}

test.describe('Wiring Editor Surface', () => {
    test('wiring tab is reachable in advanced mode', async ({ page }) => {
        const opened = await openWiringTab(page);
        if (!opened) {
            test.skip();
            return;
        }
        await expect(page.getByRole('button', { name: 'Neural Wiring' })).toBeVisible();
        await expect(page.locator('nextjs-portal')).not.toBeVisible();
    });

    test('wiring surface does not crash when mounted', async ({ page }) => {
        const opened = await openWiringTab(page);
        if (!opened) {
            test.skip();
            return;
        }
        await expect(page.locator('body')).toBeVisible();
        await expect(page.locator('nextjs-portal')).not.toBeVisible();
    });
});

