import { test, expect, devices } from '@playwright/test';

test.describe('Mobile Viewport', () => {
    test.use({ ...devices['Pixel 5'] });

    test('navigation is accessible on mobile', async ({ page }) => {
        await page.goto('/');
        await page.waitForLoadState('networkidle');

        // Next.js error overlay must not appear
        const errorOverlay = page.locator('nextjs-portal');
        await expect(errorOverlay).not.toBeVisible();

        // The Mycelis brand should be visible or accessible
        // On mobile the rail may collapse or become a hamburger menu
        const brand = page.locator('text=Mycelis');
        const navLinks = page.locator('a[href="/teams"], a[href="/wiring"]');

        // Either the brand is directly visible, or there is a menu toggle
        const brandVisible = await brand.isVisible().catch(() => false);
        const menuToggle = page.locator('button[aria-label*="menu" i], button[aria-label*="nav" i], [data-testid="mobile-menu"]');
        const toggleVisible = await menuToggle.first().isVisible().catch(() => false);

        // At least one navigation mechanism must be present
        expect(brandVisible || toggleVisible).toBeTruthy();

        // If there's a menu toggle, clicking it should reveal nav links
        if (toggleVisible) {
            await menuToggle.first().click();
            await page.waitForTimeout(500);
            await expect(navLinks.first()).toBeVisible();
        }
    });

    test('mission control stacks vertically on mobile', async ({ page }) => {
        await page.goto('/');
        await page.waitForLoadState('networkidle');

        // Page should render without crashing
        const errorOverlay = page.locator('nextjs-portal');
        await expect(errorOverlay).not.toBeVisible();

        // Body must be visible
        const body = page.locator('body');
        await expect(body).toBeVisible();

        // Content should be present — at least the Mission Control heading
        await expect(page.locator('text=Mission Control')).toBeVisible();

        // Verify content is not clipped — the main content area should have
        // a height that extends beyond the viewport (scrollable) or fit within it
        const viewportSize = page.viewportSize();
        expect(viewportSize).not.toBeNull();

        // The page content should be rendered within the viewport width
        const mainContent = page.locator('main, [role="main"], #__next > div').first();
        const box = await mainContent.boundingBox();
        if (box) {
            // Content width should not exceed viewport width
            expect(box.width).toBeLessThanOrEqual(viewportSize!.width + 1);
        }
    });

    test('no horizontal overflow on mobile', async ({ page }) => {
        await page.goto('/');
        await page.waitForLoadState('networkidle');

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
