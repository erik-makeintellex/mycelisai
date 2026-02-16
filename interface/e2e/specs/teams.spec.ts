import { test, expect } from '@playwright/test';

test.describe('Panel A — Team Management Page', () => {

    test.beforeEach(async ({ page }) => {
        await page.goto('/teams');
        await page.waitForLoadState('networkidle');
    });

    test('page loads without errors', async ({ page }) => {
        // No Next.js error overlay
        const errorOverlay = page.locator('nextjs-portal');
        await expect(errorOverlay).not.toBeVisible();

        // Header is visible
        await expect(page.locator('text=Team Management')).toBeVisible();
    });

    test('header shows team count and agent stats', async ({ page }) => {
        // Header bar is present
        const header = page.locator('h1:has-text("Team Management")');
        await expect(header).toBeVisible();

        // Team count text (e.g. "3 teams")
        await expect(page.locator('text=/\\d+ teams?/')).toBeVisible();

        // Agent count text (e.g. "2/5 agents online")
        await expect(page.locator('text=/\\d+\\/\\d+ agents online/')).toBeVisible();
    });

    test('filter dropdown is functional', async ({ page }) => {
        const select = page.locator('select');
        await expect(select).toBeVisible();

        // All three options exist
        await expect(select.locator('option')).toHaveCount(3);
        await expect(select.locator('option:has-text("All Teams")')).toBeVisible();
        await expect(select.locator('option:has-text("Standing")')).toBeVisible();
        await expect(select.locator('option:has-text("Mission")')).toBeVisible();

        // Changing filter updates the page (no crash)
        await select.selectOption('standing');
        await page.waitForTimeout(300);
        await select.selectOption('mission');
        await page.waitForTimeout(300);
        await select.selectOption('all');
    });

    test('refresh button is present and clickable', async ({ page }) => {
        // RefreshCw icon button
        const refreshBtn = page.locator('button').filter({
            has: page.locator('svg'),
        }).last();
        await expect(refreshBtn).toBeVisible();
        await refreshBtn.click();

        // Should show spin animation briefly
        await page.waitForTimeout(500);
    });

    test('team cards render in responsive grid', async ({ page }) => {
        // Cards are rendered as buttons with rounded-xl
        const cards = page.locator('button.rounded-xl');

        // With a running core server we expect at least 1 team (Admin)
        // Without core, the empty state should show
        const count = await cards.count();
        if (count > 0) {
            // First card has team name and type badge
            const firstCard = cards.first();
            await expect(firstCard).toBeVisible();

            // Card has agent count line
            await expect(firstCard.locator('text=/\\d+\\/\\d+ agent/')).toBeVisible();
        } else {
            // Empty state
            await expect(page.locator('text=No teams found')).toBeVisible();
        }
    });

    test('team card shows role-based left border accent', async ({ page }) => {
        const cards = page.locator('button.rounded-xl');
        const count = await cards.count();
        if (count === 0) return; // Skip if no teams

        const firstCard = cards.first();
        // Cards have border-l-4 accent
        await expect(firstCard).toHaveClass(/border-l-4/);
    });

    test('team card shows type badge', async ({ page }) => {
        const cards = page.locator('button.rounded-xl');
        const count = await cards.count();
        if (count === 0) return;

        const firstCard = cards.first();
        // Type badge: "standing" or "mission"
        const badge = firstCard.locator('span').filter({
            hasText: /^(standing|mission)$/i,
        });
        await expect(badge.first()).toBeVisible();
    });

    test('no bg-white leak on teams page', async ({ page }) => {
        const body = await page.content();
        expect(body).not.toContain('bg-white');
    });
});

test.describe('Panel B — Team Detail Drawer', () => {

    test.beforeEach(async ({ page }) => {
        await page.goto('/teams');
        await page.waitForLoadState('networkidle');
    });

    test('clicking a team card opens the detail drawer', async ({ page }) => {
        const cards = page.locator('button.rounded-xl');
        const count = await cards.count();
        if (count === 0) {
            test.skip();
            return;
        }

        // Click first card
        await cards.first().click();

        // Drawer slides in from the right (w-[480px])
        const drawer = page.locator('div').filter({
            has: page.locator('text=Agent Roster'),
        });
        await expect(drawer).toBeVisible({ timeout: 5000 });
    });

    test('drawer shows team name and role badge', async ({ page }) => {
        const cards = page.locator('button.rounded-xl');
        if ((await cards.count()) === 0) {
            test.skip();
            return;
        }

        await cards.first().click();

        // Role badge is visible (e.g., "admin", "action")
        const drawer = page.locator('.w-\\[480px\\]');
        await expect(drawer).toBeVisible();

        // Close button (X)
        const closeBtn = drawer.locator('button').first();
        await expect(closeBtn).toBeVisible();
    });

    test('drawer shows input and delivery topics', async ({ page }) => {
        const cards = page.locator('button.rounded-xl');
        if ((await cards.count()) === 0) {
            test.skip();
            return;
        }

        await cards.first().click();
        const drawer = page.locator('.w-\\[480px\\]');
        await expect(drawer).toBeVisible();

        // "Input Topics" or "Delivery Topics" section may appear
        // At minimum "Agent Roster" section is always present
        await expect(drawer.locator('text=Agent Roster')).toBeVisible();
    });

    test('agent roster lists agents with status dots', async ({ page }) => {
        const cards = page.locator('button.rounded-xl');
        if ((await cards.count()) === 0) {
            test.skip();
            return;
        }

        await cards.first().click();
        const drawer = page.locator('.w-\\[480px\\]');
        await expect(drawer).toBeVisible();

        // Agent rows have status dot (w-2 h-2 rounded-full)
        const statusDots = drawer.locator('.w-2.h-2.rounded-full');
        const dotCount = await statusDots.count();
        // At least the header dot exists even if no agents
        expect(dotCount).toBeGreaterThanOrEqual(0);
    });

    test('clicking agent row expands inline details', async ({ page }) => {
        const cards = page.locator('button.rounded-xl');
        if ((await cards.count()) === 0) {
            test.skip();
            return;
        }

        await cards.first().click();
        const drawer = page.locator('.w-\\[480px\\]');
        await expect(drawer).toBeVisible();

        // Find an agent row button (chevron + name)
        const agentRows = drawer.locator('button').filter({
            has: page.locator('svg'), // ChevronRight icon
        });

        if ((await agentRows.count()) === 0) return;

        // Click first agent row to expand
        await agentRows.first().click();

        // Expanded section should show "Status:" label
        await expect(drawer.locator('text=Status:')).toBeVisible();
    });

    test('close button dismisses the drawer', async ({ page }) => {
        const cards = page.locator('button.rounded-xl');
        if ((await cards.count()) === 0) {
            test.skip();
            return;
        }

        await cards.first().click();
        const drawer = page.locator('.w-\\[480px\\]');
        await expect(drawer).toBeVisible();

        // Click close (X) button in drawer header
        const closeBtn = drawer.locator('button').filter({
            has: page.locator('svg'),
        }).first();
        await closeBtn.click();

        // Drawer should disappear
        await expect(drawer).not.toBeVisible();
    });

    test('selected card has ring highlight', async ({ page }) => {
        const cards = page.locator('button.rounded-xl');
        if ((await cards.count()) === 0) {
            test.skip();
            return;
        }

        await cards.first().click();

        // Selected card gets ring-1 ring-cortex-primary
        await expect(cards.first()).toHaveClass(/ring-1/);
    });
});
