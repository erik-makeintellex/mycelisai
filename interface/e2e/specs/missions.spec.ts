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
        await expect(page.getByTestId('soma-outcome-vault')).toHaveCount(0);
        await expect(page.getByRole('button', { name: /Open Outcome Vault/i }).first()).toBeVisible();
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

    test('Outcome Vault opens as an overlay without squeezing Soma', async ({ page }) => {
        await expect(page.getByRole('heading', { name: /Talk to Soma/i })).toBeVisible();
        const chatBox = page.getByTestId('central-soma-chat-frame');
        const widthBefore = await chatBox.boundingBox().then((box) => box?.width ?? 0);

        await expect(page.getByTestId('soma-outcome-vault')).toHaveCount(0);
        await page.getByRole('button', { name: /Open Outcome Vault/i }).first().click();

        await expect(page.getByTestId('soma-outcome-vault-overlay')).toBeVisible();
        await expect(page.getByRole('heading', { name: /Outcome Vault/i })).toBeVisible();
        await expect(page.getByRole('heading', { name: /Talk to Soma/i })).toBeVisible();
        const widthDuring = await chatBox.boundingBox().then((box) => box?.width ?? 0);
        expect(widthDuring).toBeGreaterThanOrEqual(widthBefore - 8);

        await page.getByRole('button', { name: /Close Outcome Vault/i }).click();

        await expect(page.getByTestId('soma-outcome-vault')).toHaveCount(0);
    });

    test('quick action studio saves a reusable Soma ask', async ({ page }) => {
        await page.evaluate(() => window.localStorage.removeItem('mycelis-soma-saved-actions'));
        await page.reload({ waitUntil: 'domcontentloaded' });
        const actionShelf = page.getByTestId('soma-action-shelf');
        await expect(actionShelf).toBeVisible();
        await expect(actionShelf).toHaveAttribute('data-hydrated', 'true');

        await page.getByRole('button', { name: /Create new quick action/i }).click();
        const studio = page.getByRole('dialog', { name: /Save quick action/i });
        await expect(studio).toBeVisible();
        await studio.getByLabel('Button label').fill('Client risk brief');
        await studio.getByLabel('Outcome', { exact: true }).fill('Create a retained brief with risks and next steps');
        await studio.getByLabel('Output format').fill('Markdown');
        await studio.getByRole('button', { name: /Save action/i }).click();

        await expect(page.getByRole('dialog', { name: /Save quick action/i })).toHaveCount(0);
        await expect(page.getByRole('button', { name: 'Client risk brief' })).toBeVisible();
        await expect(page.getByText(/Run Expense Audit/i)).toBeVisible();
    });

    test('dashboard keeps secondary setup chrome out of the Soma workspace', async ({ page }) => {
        const surface = page.getByTestId('soma-operating-surface');
        await expect(surface).toBeVisible();
        await expect(page.getByTestId('soma-environment-entry')).toHaveCount(0);
        await expect(page.getByText(/Create or open AI Organizations/i)).toHaveCount(0);
    });

    test('no bg-white leak on dashboard', async ({ page }) => {
        const body = await page.content();
        expect(body).not.toContain('bg-white');
    });
});
