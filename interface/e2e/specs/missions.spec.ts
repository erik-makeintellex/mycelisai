import { test, expect } from '@playwright/test';

test.describe('Soma Dashboard (/dashboard)', () => {

    test.beforeEach(async ({ page }) => {
        await page.goto('/dashboard');
        await page.waitForLoadState('domcontentloaded');
    });

    test('dashboard loads without errors', async ({ page }) => {
        const errorOverlay = page.locator('nextjs-portal');
        await expect(errorOverlay).not.toBeVisible();

        await expect(page.getByRole('heading', { name: /What do you want Soma to do/i })).toBeVisible();
        await expect(page.getByTestId('soma-action-shelf')).toBeVisible();
        await expect(page.getByRole('heading', { name: /Talk to Soma/i })).toBeVisible();
        await expect(page.getByTestId('soma-outcome-vault')).toBeVisible();
        await expect(page.getByText(/Outcomes & Vault/i)).toBeVisible();
    });

    test('navigation rail is visible', async ({ page }) => {
        await expect(page.locator('a[href="/dashboard"]').first()).toBeVisible();
        await expect(page.getByRole('link', { name: 'Docs' })).toBeVisible();
        await expect(page.getByRole('link', { name: 'Settings' })).toBeVisible();
    });

    test('Soma-first entry replaces legacy organization setup actions', async ({ page }) => {
        await expect(page.getByTestId('soma-operating-surface')).toBeVisible();
        await expect(page.getByRole('button', { name: 'Set up an AI Organization' })).toHaveCount(0);
        await expect(page.getByRole('button', { name: 'Start from template' })).toHaveCount(0);
        await expect(page.getByRole('button', { name: 'Start Empty', exact: true })).toHaveCount(0);
    });

    test('signed-in environment guidance renders below the Soma workspace', async ({ page }) => {
        const surface = page.getByTestId('soma-operating-surface');
        const environment = page.getByTestId('soma-environment-entry');
        await expect(surface).toBeVisible();
        await expect(environment).toBeVisible();
        await expect(environment.getByText(/Signed in/i)).toBeVisible();
    });

    test('no bg-white leak on dashboard', async ({ page }) => {
        const body = await page.content();
        expect(body).not.toContain('bg-white');
    });
});
