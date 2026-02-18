import { test, expect } from '@playwright/test';

test.describe('Error Scenarios', () => {

    test('404 page renders gracefully', async ({ page }) => {
        const response = await page.goto('/nonexistent-route');
        await page.waitForLoadState('networkidle');

        // Page should respond (not crash/timeout)
        expect(response).not.toBeNull();
        const status = response?.status();
        // Accept 404 or a redirect (3xx) to a valid page
        expect(status === 404 || (status !== undefined && status >= 200 && status < 400)).toBeTruthy();

        // Next.js error overlay must not appear
        const errorOverlay = page.locator('nextjs-portal');
        await expect(errorOverlay).not.toBeVisible();

        // Page should render some content (404 message or redirected page)
        const body = page.locator('body');
        await expect(body).toBeVisible();
    });

    test('shows error state on API failure', async ({ page }) => {
        // Intercept all API calls to the backend and return 500
        await page.route('**/api/v1/**', (route) => {
            route.fulfill({
                status: 500,
                contentType: 'application/json',
                body: JSON.stringify({ error: 'Internal Server Error' }),
            });
        });

        await page.goto('/');
        await page.waitForLoadState('networkidle');

        // Next.js error overlay must not appear — the page should handle
        // API failures gracefully (showing empty states or error messages)
        const errorOverlay = page.locator('nextjs-portal');
        await expect(errorOverlay).not.toBeVisible();

        // Body should still be visible (page didn't crash)
        const body = page.locator('body');
        await expect(body).toBeVisible();

        // Mycelis branding should still render (shell is intact)
        await expect(page.locator('text=Mycelis')).toBeVisible();
    });

    test('SSE reconnection — disconnect does not crash the page', async ({ page }) => {
        // Track console errors during the test
        const consoleErrors: string[] = [];
        page.on('console', (msg) => {
            if (msg.type() === 'error') consoleErrors.push(msg.text());
        });

        // Navigate to the wiring page which uses SSE via /api/v1/stream
        await page.goto('/wiring');
        await page.waitForLoadState('networkidle');

        // Verify the page loaded without crashing
        const errorOverlay = page.locator('nextjs-portal');
        await expect(errorOverlay).not.toBeVisible();

        // Abort all SSE connections by intercepting the stream endpoint
        await page.route('**/api/v1/stream', (route) => {
            route.abort('connectionrefused');
        });

        // Wait a moment for any reconnection attempt to fire
        await page.waitForTimeout(2000);

        // Page should still be intact after SSE disconnection
        const body = page.locator('body');
        await expect(body).toBeVisible();

        // ReactFlow canvas should still be mounted (not unmounted on error)
        const reactFlow = page.locator('.react-flow');
        await expect(reactFlow).toBeVisible();

        // No fatal unhandled errors (hydration, null reference, etc.)
        const fatalErrors = consoleErrors.filter(
            (msg) => msg.includes('Hydration') || msg.includes('Unhandled')
        );
        expect(fatalErrors).toHaveLength(0);
    });
});
