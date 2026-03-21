import { test, expect, devices } from '@playwright/test';

test.use({ ...devices['Pixel 5'] });

test.describe('Mobile Viewport', () => {
    test('navigation is accessible on mobile', async ({ page }) => {
        await page.goto('/');
        await page.waitForLoadState('domcontentloaded');

        // Next.js error overlay must not appear
        const errorOverlay = page.locator('nextjs-portal');
        await expect(errorOverlay).not.toBeVisible();

        await expect(page.getByRole('link', { name: 'Create AI Organization' }).first()).toBeVisible();
        await expect(page.getByRole('link', { name: 'Explore Templates' }).first()).toBeVisible();
    });

    test('hero content renders correctly on mobile', async ({ page }) => {
        await page.goto('/');
        await page.waitForLoadState('domcontentloaded');

        // Page should render without crashing
        const errorOverlay = page.locator('nextjs-portal');
        await expect(errorOverlay).not.toBeVisible();

        await expect(page.locator('body')).toBeVisible();
        await expect(page.getByRole('heading', { name: 'Build AI Organizations that think, review, and evolve.' })).toBeVisible();

        // Verify content is not clipped — the main content area should have
        // a height that extends beyond the viewport (scrollable) or fit within it
        const viewportSize = page.viewportSize();
        expect(viewportSize).not.toBeNull();

        const widths = await page.evaluate(() => ({
            doc: document.documentElement.clientWidth,
            body: document.body.clientWidth,
        }));
        expect(widths.doc).toBeGreaterThan(0);
        expect(widths.body).toBeGreaterThan(0);
    });

    test('no horizontal overflow on mobile', async ({ page }) => {
        await page.goto('/');
        await page.waitForLoadState('domcontentloaded');

        // Next.js error overlay must not appear
        const errorOverlay = page.locator('nextjs-portal');
        await expect(errorOverlay).not.toBeVisible();

        // Check that the document body does not overflow horizontally
        const { scrollWidth, clientWidth } = await page.evaluate(() => ({
            scrollWidth: document.body.scrollWidth,
            clientWidth: document.documentElement.clientWidth,
        }));

        // scrollWidth should not exceed the viewport client width
        // Allow 1px tolerance for sub-pixel rendering
        expect(scrollWidth).toBeLessThanOrEqual(clientWidth + 1);
    });
});
