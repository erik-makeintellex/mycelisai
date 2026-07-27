import { test, expect } from '@playwright/test';

test.describe('Agent Catalogue Page (/catalogue)', () => {
    test.skip(({ browserName }) => browserName === 'firefox', 'Firefox catalogue coverage is currently unstable; keep Chromium as the baseline browser for now.');

    test.beforeEach(async ({ page }) => {
        await page.goto('/catalogue', { waitUntil: 'domcontentloaded' });
        await expect(page).toHaveURL(/\/resources\?tab=roles$/);
    });

    test('page loads without errors', async ({ page }) => {
        await expect(page.getByRole('heading', { name: 'Resources' })).toBeVisible();
        await expect(page.getByText('Agent Catalogue', { exact: true })).toBeVisible();
        const errorOverlay = page.locator('nextjs-portal');
        await expect(errorOverlay).not.toBeVisible();
    });

    test('role library exposes its current controls', async ({ page }) => {
        await expect(page.getByRole('combobox')).toBeVisible();
        await expect(page.getByRole('button', { name: 'New Agent' })).toBeVisible();
    });

    test('create agent button is visible', async ({ page }) => {
        const createButton = page.getByRole('button', { name: 'New Agent' });
        await expect(createButton).toBeVisible();
        await createButton.click();
        await expect(page.getByText('New Agent', { exact: true }).last()).toBeVisible();
        await expect(page.getByRole('button', { name: 'Create' })).toBeVisible();
    });

    test('no bg-white leak on catalogue page', async ({ page }) => {
        const body = await page.content();
        expect(body).not.toContain('bg-white');
    });
});
