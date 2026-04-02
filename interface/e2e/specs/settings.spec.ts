import { test, expect } from '@playwright/test';

test.describe('Settings Page (/settings)', () => {
    test.skip(({ browserName }) => browserName === 'webkit', 'WebKit currently crashes on this settings surface; keep Chromium/Firefox coverage stable for now.');

    test.beforeEach(async ({ page }) => {
        await page.goto('/settings');
        await page.waitForLoadState('domcontentloaded');
    });

    test('page loads without errors', async ({ page }) => {
        const errorOverlay = page.locator('nextjs-portal');
        await expect(errorOverlay).not.toBeVisible();
    });

    test('settings page leads with a guided setup path', async ({ page }) => {
        await expect(page.getByText('Guided setup path')).toBeVisible();
        await expect(page.getByText('Start with the controls most operators actually need.')).toBeVisible();
        await expect(page.getByRole('button', { name: 'Open Profile' })).toBeVisible();
        await expect(page.getByRole('button', { name: 'Open Mission Profiles' })).toBeVisible();
        await expect(page.getByRole('button', { name: 'Open People & Access' })).toBeVisible();
        await expect(page.getByText('Advanced setup stays intentional')).toBeVisible();
    });

    test('guided workflow controls can switch visible sections', async ({ page }) => {
        await page.getByRole('button', { name: 'Open Mission Profiles' }).click();
        await expect(page.getByRole('tab', { name: 'Mission Profiles', selected: true })).toBeVisible();
        await expect(page.getByText(/Mission Profiles/i).first()).toBeVisible();
    });

    test('no bg-white leak on settings page', async ({ page }) => {
        const body = await page.content();
        expect(body).not.toContain('bg-white');
    });
});
