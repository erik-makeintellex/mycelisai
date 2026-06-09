import { test, expect } from '@playwright/test';

test.describe('Teams Workspace (/teams)', () => {
    test.beforeEach(async ({ page }) => {
        await page.addInitScript(() => {
            window.localStorage.setItem('mycelis-advanced-mode', 'true');
        });
        await page.goto('/teams', { waitUntil: 'domcontentloaded' });
        await expect(page.locator('h1:has-text("Team Lead Workspaces")')).toBeVisible({ timeout: 15000 });
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

    test('output and actuation guidance points to dedicated workspaces', async ({ page }) => {
        await expect(page.getByText('Specialize new teams through Soma')).toBeVisible();
        await expect(page.getByRole('link', { name: 'Review group outputs' })).toHaveAttribute('href', '/groups');
        await expect(page.getByRole('link', { name: 'Configure event rules' })).toHaveAttribute('href', '/automations?tab=triggers');
    });

    test('guided team creation is reachable from the teams workspace', async ({ page }) => {
        await expect(page.getByRole('link', { name: 'Open guided team creation' })).toHaveAttribute('href', '/teams/create');
    });

    test('review route puts decision work before team context', async ({ page }) => {
        await page.route('**/api/v1/teams/detail', async (route) => {
            await route.fulfill({
                status: 200,
                contentType: 'application/json',
                body: JSON.stringify([
                    {
                        id: 'team-review',
                        name: 'Review Team',
                        role: 'action',
                        type: 'mission',
                        mission_id: 'mission-review',
                        mission_intent: 'Review blocked output',
                        inputs: [],
                        deliveries: [],
                        agents: [],
                    },
                ]),
            });
        });
        await page.route('**/api/v1/catalogue/agents**', async (route) => {
            await route.fulfill({
                status: 200,
                contentType: 'application/json',
                body: JSON.stringify({ data: [] }),
            });
        });
        await page.route('**/api/v1/teams/team-review/work?*', async (route) => {
            await route.fulfill({
                status: 200,
                contentType: 'application/json',
                body: JSON.stringify({
                    data: [
                        {
                            work_item_id: 'work-review-1',
                            team_id: 'team-review',
                            objective: 'Recover failed release notes',
                            execution_shape: 'deliverable',
                            state: 'degraded',
                            needs_operator: true,
                            degradation_state: 'provider_timeout',
                            recovery_options: ['retry with retained proof'],
                            updated_at: '2026-06-09T10:00:00Z',
                        },
                    ],
                }),
            });
        });

        await page.goto('/teams?view=work', { waitUntil: 'domcontentloaded' });

        await expect(page.getByRole('heading', { name: 'Work to Review', exact: true })).toBeVisible();
        await expect(page.getByLabel('Review queue summary')).toBeVisible();
        await expect(page.getByText('Needs decision', { exact: true })).toBeVisible();
        await expect(page.getByText('Reason', { exact: true })).toBeVisible();
        await expect(page.getByText('Trust', { exact: true })).toBeVisible();
        await expect(page.getByText('Move', { exact: true })).toBeVisible();
        const bodyText = await page.locator('body').innerText();
        expect(bodyText.indexOf('Work to Review')).toBeLessThan(bodyText.indexOf('Team context'));
    });

    test('team quick action links are wired', async ({ page }) => {
        const cards = page.getByRole('button')
            .filter({ hasText: 'Open lead workspace' })
            .filter({ hasText: 'View runs' });
        const openChatLinks = page.locator('a[data-testid$="-open-chat"]');
        const viewRunsLinks = page.locator('a[data-testid$="-view-runs"]');
        const emptyState = page.getByText('No teams found', { exact: true });
        await expect
            .poll(async () => {
                if ((await cards.count()) > 0) return 'cards';
                return await emptyState.isVisible().catch(() => false) ? 'empty' : 'pending';
            })
            .not.toBe('pending');
        const count = await cards.count();
        if (count === 0) {
            await expect(emptyState).toBeVisible();
            return;
        }

        await expect(openChatLinks).toHaveCount(count);
        await expect(viewRunsLinks).toHaveCount(count);
        await expect(page.locator('a[data-testid$="-view-wiring"]')).toHaveCount(0);
        await expect(page.locator('a[data-testid$="-view-logs"]')).toHaveCount(0);

        await expect(openChatLinks.first()).toHaveAttribute('href', /\/dashboard\?team_id=/);
        await expect(viewRunsLinks.first()).toHaveAttribute('href', '/runs');
    });

    test('clicking a team card opens and closes the detail drawer', async ({ page }) => {
        test.slow();
        const cards = page.getByRole('button')
            .filter({ hasText: 'Open lead workspace' })
            .filter({ hasText: 'View runs' });
        const emptyState = page.getByText('No teams found', { exact: true });
        await expect
            .poll(async () => {
                if ((await cards.count()) > 0) return 'cards';
                return await emptyState.isVisible().catch(() => false) ? 'empty' : 'pending';
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
        await expect(drawer.getByText('Operator controls')).toBeVisible();
        await expect(drawer.getByRole('link', { name: 'Open lead workspace' })).toHaveAttribute('href', /\/dashboard\?team_id=/);
        await expect(drawer.getByRole('link', { name: 'View runs' })).toHaveAttribute('href', '/runs');
        await expect(drawer.getByRole('link', { name: 'View outputs' })).toHaveAttribute('href', '/groups');
        await expect(drawer.getByText('Advanced coordination topics')).toBeVisible();
        await expect(drawer.getByRole('link', { name: 'View wiring' })).toHaveAttribute('href', '/automations?tab=wiring');
        await expect(drawer.getByRole('link', { name: 'View system' })).toHaveAttribute('href', '/system?tab=services');
        await expect(firstCard).toHaveClass(/ring-1/);

        await drawer.locator('button').first().click();
        await expect(drawer).not.toBeVisible();
    });
});
