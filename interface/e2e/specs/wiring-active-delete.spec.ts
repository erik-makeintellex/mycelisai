import { test, expect } from '@playwright/test';

test.describe('Wiring Active Flow Guardrails', () => {
    test('wiring route is reachable via automations tab', async ({ page }) => {
        await page.goto('/automations?tab=wiring');
        await page.waitForLoadState('domcontentloaded');
        await expect(page.locator('h1:has-text("Automations")')).toBeVisible();
    });

    test('wiring path does not crash on load', async ({ page }) => {
        await page.goto('/automations?tab=wiring');
        await page.waitForLoadState('domcontentloaded');
        await expect(page.locator('nextjs-portal')).not.toBeVisible();
        await expect(page.locator('body')).toBeVisible();
    });
});

