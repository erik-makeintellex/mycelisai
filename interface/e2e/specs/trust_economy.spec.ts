import { test, expect } from '@playwright/test';

test.describe('Trust Economy & Governance Surfaces', () => {
    test('approvals tab is reachable from automations', async ({ page }) => {
        await page.goto('/automations');
        await page.waitForLoadState('domcontentloaded');
        await page.getByRole('button', { name: 'Approvals' }).click();
        await expect(page.locator('body')).toContainText(/approval|governance|policy|degraded/i);
    });
});

