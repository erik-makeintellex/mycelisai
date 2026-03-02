import { test, expect } from '@playwright/test';

test.describe('Error Scenarios', () => {
    test('404 page renders gracefully', async ({ page }) => {
        const response = await page.goto('/nonexistent-route');
        await page.waitForLoadState('domcontentloaded');

        expect(response).not.toBeNull();
        expect(page.locator('body')).toBeVisible();
        await expect(page.locator('nextjs-portal')).not.toBeVisible();
    });

    test('API failures do not crash the dashboard shell', async ({ page }) => {
        await page.route('**/api/v1/**', (route) => {
            route.fulfill({
                status: 500,
                contentType: 'application/json',
                body: JSON.stringify({ error: 'Internal Server Error' }),
            });
        });

        await page.goto('/dashboard');
        await page.waitForLoadState('domcontentloaded');

        await expect(page.locator('body')).toBeVisible();
        await expect(page.locator('a[href="/dashboard"]').first()).toBeVisible();
        await expect(page.locator('nextjs-portal')).not.toBeVisible();
    });

    test('SSE disconnect does not crash workspace', async ({ page }) => {
        const consoleErrors: string[] = [];
        page.on('console', (msg) => {
            if (msg.type() === 'error') consoleErrors.push(msg.text());
        });

        await page.route('**/api/v1/stream**', (route) => {
            route.abort('connectionrefused');
        });

        await page.goto('/dashboard');
        await page.waitForLoadState('domcontentloaded');
        await page.waitForTimeout(2000);

        await expect(page.locator('body')).toBeVisible();
        await expect(page.locator('nextjs-portal')).not.toBeVisible();

        const fatalErrors = consoleErrors.filter(
            (msg) => msg.includes('Hydration') || msg.includes('Unhandled'),
        );
        expect(fatalErrors).toHaveLength(0);
    });
});
