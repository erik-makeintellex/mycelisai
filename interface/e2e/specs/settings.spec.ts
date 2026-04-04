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

    test('theme and assistant identity save and persist after reload', async ({ page }) => {
        const assistantInput = page.getByLabel('Assistant Name');
        const themeSelect = page.getByLabel('Theme');
        const saveButtons = page.getByRole('button', { name: 'Save' });
        const originalAssistantName = (await assistantInput.inputValue()).trim();
        const originalTheme = await themeSelect.inputValue();
        const updatedAssistantName = 'Atlas';
        const updatedTheme = 'midnight-cortex';

        await assistantInput.fill(updatedAssistantName);
        await saveButtons.nth(0).click();
        await expect(page.getByText('Saved').first()).toBeVisible();

        await themeSelect.selectOption(updatedTheme);
        await saveButtons.nth(1).click();
        await expect(page.locator('html')).toHaveAttribute('data-theme', updatedTheme);

        await page.reload({ waitUntil: 'domcontentloaded' });
        await expect(page.getByLabel('Assistant Name')).toHaveValue(updatedAssistantName);
        await expect(page.getByLabel('Theme')).toHaveValue(updatedTheme);
        await expect(page.locator('html')).toHaveAttribute('data-theme', updatedTheme);

        await page.getByLabel('Assistant Name').fill(originalAssistantName);
        await saveButtons.nth(0).click();
        await page.getByLabel('Theme').selectOption(originalTheme);
        await saveButtons.nth(1).click();
    });

    test('no bg-white leak on settings page', async ({ page }) => {
        const body = await page.content();
        expect(body).not.toContain('bg-white');
    });
});
