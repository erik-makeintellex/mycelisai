import { test, expect } from '@playwright/test';

test.describe('Mission Control Dashboard (/dashboard)', () => {

    test.beforeEach(async ({ page }) => {
        await page.goto('/dashboard');
        await page.waitForLoadState('domcontentloaded');
    });

    test('dashboard loads without errors', async ({ page }) => {
        const errorOverlay = page.locator('nextjs-portal');
        await expect(errorOverlay).not.toBeVisible();

        await expect(page.getByRole('link', { name: 'Workspace' })).toBeVisible();
        await expect(page.locator('text=SOMA')).toBeVisible();
    });

    test('navigation rail is visible', async ({ page }) => {
        await expect(page.getByRole('link', { name: 'Automations' })).toBeVisible();
    });

    test('workspace command input renders', async ({ page }) => {
        await expect(page.locator('h1:has-text("Workspace")')).toBeVisible();
        await expect(page.locator('body')).toContainText(/SOMA|Workspace/i);
    });

    test('workspace system cards render', async ({ page }) => {
        await expect(page.locator('text=SYSTEM').first()).toBeVisible();
    });

    test('no bg-white leak on dashboard', async ({ page }) => {
        const body = await page.content();
        expect(body).not.toContain('bg-white');
    });
});
