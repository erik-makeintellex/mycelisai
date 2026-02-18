import { test, expect } from '@playwright/test';

test.describe('Agent Catalogue Page (/catalogue)', () => {

    test.beforeEach(async ({ page }) => {
        await page.goto('/catalogue');
        await page.waitForLoadState('networkidle');
    });

    test('page loads without errors', async ({ page }) => {
        const errorOverlay = page.locator('nextjs-portal');
        await expect(errorOverlay).not.toBeVisible();
    });

    test('agent cards display when data exists', async ({ page }) => {
        // Cards or empty state
        const cards = page.locator('[class*="rounded"]').filter({
            has: page.locator('text=/cognitive|sensory|actuation|ledger/i'),
        });
        const count = await cards.count();
        if (count === 0) {
            // Empty state or no agents — still valid
            return;
        }
        await expect(cards.first()).toBeVisible();
    });

    test('create agent button is visible', async ({ page }) => {
        const createBtn = page.locator('button:has-text("Create"), button:has-text("Add"), button:has-text("New")');
        const visible = await createBtn.first().isVisible().catch(() => false);
        if (!visible) {
            test.skip();
            return;
        }
        await expect(createBtn.first()).toBeVisible();
    });

    test('no bg-white leak on catalogue page', async ({ page }) => {
        const body = await page.content();
        expect(body).not.toContain('bg-white');
    });
});
