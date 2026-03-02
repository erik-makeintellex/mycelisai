import { test, expect } from '@playwright/test';

test.describe('Mission Proposal Entry Points', () => {
    test('workspace exposes launch controls for mission planning', async ({ page }) => {
        await page.goto('/dashboard');
        await page.waitForLoadState('domcontentloaded');

        await expect(page.locator('h1:has-text("Workspace")')).toBeVisible();
        await expect(page.locator('button:has-text("Launch Crew"), button:has-text("Launch")').first()).toBeVisible();
    });
});
