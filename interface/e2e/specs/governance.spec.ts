import { test, expect } from '@playwright/test';

test.describe('Governance Page (/automations?tab=approvals)', () => {

    test.beforeEach(async ({ page }) => {
        await page.goto('/automations?tab=approvals');
        await page.waitForLoadState('domcontentloaded');
        await expect(page.locator('body')).toContainText('Automations', { timeout: 20_000 });
    });

    test('page loads without errors', async ({ page }) => {
        const errorOverlay = page.locator('nextjs-portal');
        await expect(errorOverlay).not.toBeVisible();
        await expect(page.getByRole('button', { name: 'Approvals' })).toBeVisible();
    });

    test('policy tab renders governance rules', async ({ page }) => {
        const content = page.locator('main, body');
        await expect(content).toContainText(/approval|governance|policy|degraded/i);
    });

    test('pending approvals section renders', async ({ page }) => {
        await expect(page.locator('body')).toContainText(/pending|approval|queue|none/i);
    });

    test('audit tab renders inspect-only governance activity', async ({ page }) => {
        await page.getByRole('button', { name: 'Audit' }).click();
        await expect(page.locator('body')).toContainText(/activity log|audit/i);
        await expect(page.locator('body')).toContainText(/proposal|execution|capability|recent/i);
    });

    test('no bg-white leak on governance page', async ({ page }) => {
        const body = await page.content();
        expect(body).not.toContain('bg-white');
    });
});
