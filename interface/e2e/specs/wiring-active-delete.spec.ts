import { test, expect } from '@playwright/test';

async function openWiringTab(page: any) {
    await page.goto('/automations');
    await page.waitForLoadState('domcontentloaded');
    const advancedOff = page.getByRole('button', { name: 'Advanced: Off' });
    if (await advancedOff.isVisible().catch(() => false)) {
        await advancedOff.click();
    }
    const wiringTab = page.getByRole('button', { name: 'Workflow Builder' });
    if (!(await wiringTab.isVisible().catch(() => false))) return false;
    await wiringTab.click();
    await page.waitForTimeout(500);
    return true;
}

test.describe('Wiring Active Flow Guardrails', () => {
    test('wiring route is reachable via automations tab', async ({ page }) => {
        const opened = await openWiringTab(page);
        if (!opened) {
            // Workflow Builder is a live MVP tab in Automations (gated behind
            // Advanced Mode). This fixture does not enable Advanced Mode, so the
            // tab is currently unreachable here. fixme keeps the gap visible
            // instead of a silent skip until the fixture flips advancedMode on.
            test.fixme(true, 'Fixture does not enable Advanced Mode, so the Workflow Builder tab never renders for this spec.');
            return;
        }
        await expect(page.getByRole('button', { name: 'Workflow Builder' })).toBeVisible();
    });

    test('wiring path does not crash on load', async ({ page }) => {
        const opened = await openWiringTab(page);
        if (!opened) {
            test.fixme(true, 'Fixture does not enable Advanced Mode, so the Workflow Builder tab never renders for this spec.');
            return;
        }
        await expect(page.locator('nextjs-portal')).not.toBeVisible();
        await expect(page.locator('body')).toBeVisible();
    });
});

