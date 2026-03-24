import { test, expect } from '@playwright/test';

test.describe('Teams Tab (/automations?tab=teams)', () => {
    test.beforeEach(async ({ page }) => {
        await page.addInitScript(() => {
            window.localStorage.setItem('mycelis-advanced-mode', 'true');
        });
        await page.goto('/automations?tab=teams', { waitUntil: 'domcontentloaded' });
        await expect(page.getByRole('button', { name: /Advanced:/ })).toContainText('Advanced: On');
        await expect(page.locator('h1:has-text("Shared Teams")')).toBeVisible({ timeout: 15000 });
    });

    test('header and filter controls render', async ({ page }) => {
        await expect(page.locator('text=/\\d+ team/')).toBeVisible();
        await expect(
            page
                .locator('span')
                .filter({ hasText: /\d+\/\d+ agents online/ })
                .first(),
        ).toBeVisible();
        const filter = page
            .locator('select')
            .filter({ has: page.locator('option:has-text("All Teams")') })
            .first();
        await expect(filter).toBeVisible();
        await expect(filter.locator('option')).toHaveCount(3);
        await filter.selectOption('standing');
        await expect(filter).toHaveValue('standing');
        await filter.selectOption('mission');
        await expect(filter).toHaveValue('mission');
        await filter.selectOption('all');
        await expect(filter).toHaveValue('all');
    });

    test('group management panel renders', async ({ page }) => {
        await expect(page.locator('text=Collaboration Groups').first()).toBeVisible();
        await expect(page.getByTestId('groups-create-button')).toBeVisible();
    });

    test('team quick action links are wired', async ({ page }) => {
        const openChatLinks = page.locator('[data-testid$="-open-chat"]');
        const count = await openChatLinks.count();
        if (count === 0) {
            await expect(page.locator('text=No teams found')).toBeVisible();
            return;
        }

        await expect(page.locator('[data-testid$="-view-runs"]')).toHaveCount(count);
        await expect(page.locator('[data-testid$="-view-wiring"]')).toHaveCount(count);
        await expect(page.locator('[data-testid$="-view-logs"]')).toHaveCount(count);

        await expect(openChatLinks.first()).toHaveAttribute('href', '/dashboard');
        await expect(page.locator('[data-testid$="-view-runs"]').first()).toHaveAttribute('href', '/runs');
        await expect(page.locator('[data-testid$="-view-wiring"]').first()).toHaveAttribute('href', '/automations?tab=wiring');
        await expect(page.locator('[data-testid$="-view-logs"]').first()).toHaveAttribute('href', '/system?tab=services');
    });

    test('clicking a team card opens and closes the detail drawer', async ({ page }) => {
        test.slow();
        const cards = page.locator('div[role="button"][tabindex="0"]');
        const count = await cards.count();
        if (count === 0) {
            await expect(page.locator('text=No teams found')).toBeVisible({ timeout: 15000 });
            return;
        }

        const firstCard = cards.first();
        await firstCard.click();

        const drawer = page.locator('div.w-\\[480px\\]');
        await expect(drawer).toBeVisible();
        await expect(drawer.locator('text=Agent Roster')).toBeVisible();
        await expect(firstCard).toHaveClass(/ring-1/);

        await drawer.locator('button').first().click();
        await expect(drawer).not.toBeVisible();
    });
});
