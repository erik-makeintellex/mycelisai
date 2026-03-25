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
        const openChatLinks = page.getByRole('link', { name: 'Open chat' });
        const viewRunsLinks = page.getByRole('link', { name: 'View runs' });
        const viewWiringLinks = page.getByRole('link', { name: 'View wiring' });
        const viewSystemLinks = page.getByRole('link', { name: 'View system' });
        const emptyState = page.locator('text=No teams found');
        await expect
            .poll(async () => (await openChatLinks.count()) > 0 || (await emptyState.count()) > 0)
            .toBeTruthy();
        const count = await openChatLinks.count();
        if (count === 0) {
            await expect(emptyState).toBeVisible();
            return;
        }

        await expect(viewRunsLinks).toHaveCount(count);
        await expect(viewWiringLinks).toHaveCount(count);
        await expect(viewSystemLinks).toHaveCount(count);

        await expect(openChatLinks.first()).toHaveAttribute('href', '/dashboard');
        await expect(viewRunsLinks.first()).toHaveAttribute('href', '/runs');
        await expect(viewWiringLinks.first()).toHaveAttribute('href', '/automations?tab=wiring');
        await expect(viewSystemLinks.first()).toHaveAttribute('href', '/system?tab=services');
    });

    test('clicking a team card opens and closes the detail drawer', async ({ page }) => {
        test.slow();
        const cards = page.locator('div[role="button"][tabindex="0"]');
        const emptyState = page.getByText('No teams found');
        await expect
            .poll(async () => {
                if ((await cards.count()) > 0) return 'cards';
                const bodyText = await page.locator('body').textContent();
                return bodyText?.includes('No teams found') ? 'empty' : 'pending';
            })
            .not.toBe('pending');
        const count = await cards.count();
        if (count === 0) {
            await expect(emptyState).toBeVisible({ timeout: 15000 });
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
