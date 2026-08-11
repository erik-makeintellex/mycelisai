import { test, expect } from '@playwright/test';
import { expectNoHorizontalOverflow } from '../support/finalization-proof';

test.describe('Soma Dashboard (/dashboard)', () => {

    test.beforeEach(async ({ page }) => {
        await page.goto('/dashboard');
        await page.waitForLoadState('domcontentloaded');
    });

    test('dashboard loads without errors', async ({ page }) => {
        const errorOverlay = page.locator('nextjs-portal');
        await expect(errorOverlay).not.toBeVisible();

        await expect(page.getByRole('heading', { name: /Talk to Soma/i })).toBeVisible();
        await expect(page.getByTestId('soma-action-shelf')).toHaveCount(0);
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

        await page.getByRole('button', { name: 'Close Outcome Vault', exact: true }).click();

        await expect(page.getByTestId('soma-outcome-vault')).toHaveCount(0);
    });

    test('Outcome Vault and Soma composer stay usable across layout modes', async ({ page }, testInfo) => {
        const layouts = [
            { name: 'compact', width: 390, height: 844 },
            { name: 'medium', width: 820, height: 1180 },
            { name: 'workspace', width: 1366, height: 768 },
            { name: 'wide', width: 1440, height: 900 },
        ];

        for (const layout of layouts) {
            await page.setViewportSize({ width: layout.width, height: layout.height });
            await page.goto('/dashboard', { waitUntil: 'domcontentloaded' });
            await page.waitForLoadState('load');
            const chat = page.getByTestId('central-soma-chat-frame');
            const composer = chat.locator('textarea').first();
            const opener = page.getByRole('button', { name: /Open Outcome Vault/i });
            await expect(chat).toBeVisible();
            await expect(composer).toBeVisible();
            await expect(opener).toBeVisible();
            await expect(opener).toHaveAttribute('aria-expanded', 'false');
            const widthBefore = await chat.boundingBox().then((box) => box?.width ?? 0);

            await opener.click();

            const dialog = page.getByRole('dialog', { name: 'Outcome Vault' });
            await expect(dialog).toBeVisible();
            await expect(opener).toHaveAttribute('aria-expanded', 'true');
            await expect(page.getByRole('button', { name: 'Close Outcome Vault', exact: true })).toBeFocused();
            const widthDuring = await chat.boundingBox().then((box) => box?.width ?? 0);
            expect(widthDuring).toBeGreaterThanOrEqual(widthBefore - 8);
            const dialogBounds = await dialog.boundingBox();
            expect(dialogBounds?.width ?? 0).toBeLessThanOrEqual(layout.width);
            expect(dialogBounds?.height ?? 0).toBeLessThanOrEqual(layout.height);
            await page.screenshot({
                path: testInfo.outputPath(`dashboard-${layout.name}-outcomes.png`),
                fullPage: false,
            });

            await page.keyboard.press('Escape');

            await expect(dialog).toHaveCount(0);
            await expect(opener).toBeFocused();
            await expect(composer).toBeVisible();
            await composer.fill(`Continue from the ${layout.name} layout.`);
            await expect(composer).toHaveValue(`Continue from the ${layout.name} layout.`);
            await composer.fill('');
            await expectNoHorizontalOverflow(page);
            await expect(page.locator('nextjs-portal')).not.toBeVisible();
        }

        await page.getByRole('button', { name: /Open Outcome Vault/i }).click();
        await page.getByRole('button', { name: /Close Outcome Vault backdrop/i }).click({ position: { x: 4, y: 4 } });
        await expect(page.getByRole('dialog', { name: 'Outcome Vault' })).toHaveCount(0);
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
