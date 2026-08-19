import { test, expect } from '@playwright/test';

test.describe('Worker Profiles Page (/catalogue)', () => {
    test.skip(({ browserName }) => browserName === 'firefox', 'Firefox catalogue coverage is currently unstable; keep Chromium as the baseline browser for now.');

    test.beforeEach(async ({ page }) => {
        await page.goto('/catalogue', { waitUntil: 'domcontentloaded' });
        await expect(page).toHaveURL(/\/resources\?tab=roles$/);
    });

    test('page loads without errors', async ({ page }) => {
        await expect(page.getByRole('heading', { name: 'Resources' })).toBeVisible();
        await expect(page.getByText('Worker profiles', { exact: true })).toBeVisible();
        const errorOverlay = page.locator('nextjs-portal');
        await expect(errorOverlay).not.toBeVisible();
    });

    test('worker profile library exposes its current controls', async ({ page }) => {
        await expect(page.getByText(/ready-made profiles/)).toBeVisible();
        await expect(page.getByRole('button', { name: 'Create with Soma' })).toBeVisible();
    });

    test('custom profile creation continues naturally in Soma', async ({ page }) => {
        await page.getByRole('button', { name: 'Create with Soma' }).click();
        await expect(page).toHaveURL(/\/dashboard$/);
        await expect(page.getByPlaceholder(/Tell Soma/)).toHaveValue(/create a reusable Worker Profile/);
    });

    test('ready-made profile can be inspected and customized through Soma', async ({ page }) => {
        await expect(page.getByText('Research Specialist', { exact: true })).toBeVisible();
        await page.getByText('Research Specialist', { exact: true }).click();
        const dialog = page.getByRole('dialog', { name: 'Research Specialist profile' });
        await expect(dialog).toBeVisible();
        await expect(dialog.getByRole('button', { name: 'Customize with Soma' })).toBeVisible();

        await dialog.getByRole('tab', { name: 'Access & context' }).click();
        await expect(dialog.getByText('web_search', { exact: true })).toBeVisible();
        await expect(dialog.getByLabel('Context type 1')).toHaveValue('public_web');
        await expect(dialog.getByLabel('Context access 1')).toHaveValue('search');

        await dialog.getByRole('button', { name: 'Customize with Soma' }).click();
        await expect(page).toHaveURL(/\/dashboard$/);
        await expect(page.getByPlaceholder(/Tell Soma/)).toHaveValue(/based on "Research Specialist"/);
    });

    test('no bg-white leak on catalogue page', async ({ page }) => {
        const body = await page.content();
        expect(body).not.toContain('bg-white');
    });
});
