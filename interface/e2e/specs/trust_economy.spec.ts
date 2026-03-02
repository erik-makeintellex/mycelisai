import { test, expect } from '@playwright/test';

test.describe('Trust Economy & Governance Surfaces', () => {
    test('workspace renders governance/status controls', async ({ page }) => {
        await page.goto('/dashboard');
        await page.waitForLoadState('domcontentloaded');

        await expect(page.locator('text=/GOV:/')).toBeVisible();
        await expect(page.locator('text=/MODE:/')).toBeVisible();
    });

    test('approvals tab is reachable from automations', async ({ page }) => {
        await page.goto('/automations');
        await page.waitForLoadState('domcontentloaded');
        await page.getByRole('button', { name: 'Approvals' }).click();
        await expect(page.locator('body')).toContainText(/approval|governance|policy|degraded/i);
    });
});

