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

test.describe('Wiring Active Flow Guardrails', () => {
    test('wiring route is reachable via automations tab', async ({ page }) => {
        await openWiringTab(page);
        await expect(page.getByRole('button', { name: 'Workflow Builder' })).toBeVisible();
    });

    test('wiring path does not crash on load', async ({ page }) => {
        await openWiringTab(page);
        await expect(page.locator('nextjs-portal')).not.toBeVisible();
        await expect(page.locator('body')).toBeVisible();
    });
});

