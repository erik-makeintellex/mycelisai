import { test, expect } from '@playwright/test';

test.describe('Mission Control Dashboard (/)', () => {

    test.beforeEach(async ({ page }) => {
        await page.goto('/');
        await page.waitForLoadState('networkidle');
    });

    test('dashboard loads without errors', async ({ page }) => {
        const errorOverlay = page.locator('nextjs-portal');
        await expect(errorOverlay).not.toBeVisible();

        // Mission Control heading or branding
        await expect(page.locator('text=Mission Control')).toBeVisible();
    });

    test('navigation rail is visible', async ({ page }) => {
        await expect(page.locator('text=Mycelis')).toBeVisible();
    });

    test('telemetry row renders metric cards or skeleton', async ({ page }) => {
        // Telemetry section should show cards or loading state
        const telemetryCards = page.locator('[data-testid="telemetry-card"]');
        const skeletons = page.locator('.animate-pulse');
        const cardCount = await telemetryCards.count();
        const skeletonCount = await skeletons.count();

        // Either metrics loaded or showing skeleton
        expect(cardCount + skeletonCount).toBeGreaterThan(0);
    });

    test('mission cards render when data exists', async ({ page }) => {
        // Mission cards in the dashboard
        const missionCards = page.locator('[class*="rounded"]').filter({
            has: page.locator('text=/active|draft|mission/i'),
        });
        const count = await missionCards.count();
        if (count === 0) {
            // No missions — empty state is valid
            return;
        }
        await expect(missionCards.first()).toBeVisible();
    });

    test('no bg-white leak on dashboard', async ({ page }) => {
        const body = await page.content();
        expect(body).not.toContain('bg-white');
    });
});
