import { test, expect } from '@playwright/test';

test.describe('Mission Control Dashboard (/dashboard)', () => {

    test.beforeEach(async ({ page }) => {
        await page.goto('/dashboard');
        await page.waitForLoadState('domcontentloaded');
    });

    test('dashboard loads without errors', async ({ page }) => {
        const errorOverlay = page.locator('nextjs-portal');
        await expect(errorOverlay).not.toBeVisible();

        await expect(page.getByRole('heading', { name: 'Create AI Organization' })).toBeVisible();
        await expect(page.getByText(/Use a starter that already defines a Team Lead/i)).toBeVisible();
    });

    test('navigation rail is visible', async ({ page }) => {
        await expect(page.locator('a[href="/dashboard"]').first()).toBeVisible();
        await expect(page.getByRole('link', { name: 'Docs' })).toBeVisible();
        await expect(page.getByRole('link', { name: 'Settings' })).toBeVisible();
    });

    test('organization entry actions render', async ({ page }) => {
        await expect(page.getByRole('button', { name: 'Start from template' })).toBeVisible();
        await expect(page.getByRole('button', { name: 'Start Empty', exact: true })).toBeVisible();
    });

    test('recent organization guidance renders', async ({ page }) => {
        await expect(page.getByRole('heading', { name: 'Recent AI Organizations' })).toBeVisible();
        await expect(
            page.getByText(/Create the AI Organization first so Mycelis opens with structure, not a one-off assistant session\./i),
        ).toBeVisible();
    });

    test('no bg-white leak on dashboard', async ({ page }) => {
        const body = await page.content();
        expect(body).not.toContain('bg-white');
    });
});
