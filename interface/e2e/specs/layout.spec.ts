import { test, expect } from '@playwright/test';

test.describe('Automations Layout Geometry', () => {
    test.beforeEach(async ({ page }) => {
        await page.goto('/automations');
        await page.waitForLoadState('domcontentloaded');
    });

    test('no Next.js error overlay', async ({ page }) => {
        await expect(page.locator('nextjs-portal')).not.toBeVisible();
    });

    test('automations header and tabs render', async ({ page }) => {
        await expect(page.locator('h1:has-text("Automations")')).toBeVisible();
        await expect(page.getByRole('button', { name: 'Active Automations' })).toBeVisible();
        await expect(page.getByRole('button', { name: 'Draft Blueprints' })).toBeVisible();
        await expect(page.getByRole('button', { name: 'Trigger Rules' })).toBeVisible();
        await expect(page.getByRole('button', { name: 'Approvals' })).toBeVisible();
        await expect(page.getByRole('button', { name: 'Teams' })).toBeVisible();
    });

    test('automation hub baseline renders', async ({ page }) => {
        await expect(page.getByTestId('automations-hub-baseline')).toBeVisible();
        await expect(page.getByTestId('open-instantiation-wizard')).toBeVisible();
    });

    test('theme does not leak white utility classes', async ({ page }) => {
        const html = await page.content();
        expect(html).not.toContain('bg-white');
    });
});

