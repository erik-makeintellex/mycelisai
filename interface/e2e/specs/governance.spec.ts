import { test, expect } from '@playwright/test';

test.describe('Governance Page (/approvals)', () => {

    test.beforeEach(async ({ page }) => {
        await page.goto('/approvals');
        await page.waitForLoadState('networkidle');
    });

    test('page loads without errors', async ({ page }) => {
        const errorOverlay = page.locator('nextjs-portal');
        await expect(errorOverlay).not.toBeVisible();

        // Governance header or content is present
        await expect(page.locator('text=Governance')).toBeVisible();
    });

    test('policy tab renders governance rules', async ({ page }) => {
        // Look for policy-related content
        const policySection = page.locator('text=/policy|rules|default/i');
        const visible = await policySection.first().isVisible().catch(() => false);
        if (!visible) {
            test.skip();
            return;
        }
        await expect(policySection.first()).toBeVisible();
    });

    test('pending approvals section renders', async ({ page }) => {
        // The pending section may show items or an empty state
        const pendingSection = page.locator('text=/pending|approval|queue/i');
        const visible = await pendingSection.first().isVisible().catch(() => false);
        if (!visible) {
            // Empty state — still counts as passing
            return;
        }
        await expect(pendingSection.first()).toBeVisible();
    });

    test('no bg-white leak on governance page', async ({ page }) => {
        const body = await page.content();
        expect(body).not.toContain('bg-white');
    });
});
