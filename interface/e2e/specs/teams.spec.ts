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
        await expect(page.getByRole('heading', { name: 'Worker profiles' })).toBeVisible();
        await expect(page.getByText('Context Analyst', { exact: true })).toBeVisible();
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
        await expect(filter).toBeEnabled();
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

        await expect(page.getByRole('heading', { name: 'Recovery and Review', exact: true })).toBeVisible();
        await expect(page.getByLabel('Review queue summary')).toBeVisible();
        await expect(page.getByText('Needs decision', { exact: true })).toBeVisible();
        await expect(page.getByTestId('work-review-inbox')).toBeVisible();
        await expect(page.getByRole('list', { name: 'Review work items' })).toBeVisible();
        await expect(page.getByLabel('Review details for Recover failed release notes')).toBeVisible();
        await expect(page.getByText('Reason', { exact: true })).toBeVisible();
        await expect(page.getByText('Trust', { exact: true })).toBeVisible();
        await expect(page.getByText('Move', { exact: true })).toBeVisible();
        await expect(page.getByRole('button', { name: /Retry recovery/i }).first()).toBeVisible();
        const bodyText = await page.locator('body').innerText();
        expect(bodyText.indexOf('Recovery and Review')).toBeLessThan(bodyText.indexOf('Team context'));
    });

    test('unknown external mutation requires Soma verification before retry', async ({ page }) => {
        let submittedVerification: Record<string, unknown> | null = null;
        let verificationRecorded = false;
        await page.route('**/api/v1/teams/detail', async (route) => {
            await route.fulfill({
                status: 200,
                contentType: 'application/json',
                body: JSON.stringify([
                    {
                        id: 'team-external-review',
                        name: 'External Review Team',
                        role: 'action',
                        type: 'mission',
                        mission_id: 'mission-external-review',
                        mission_intent: 'Verify external account update',
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
        await page.route('**/api/v1/teams/team-external-review/work?*', async (route) => {
            await route.fulfill({
                status: 200,
                contentType: 'application/json',
                body: JSON.stringify({
                    data: [
                        {
                            work_item_id: 'work-external-review-1',
                            team_id: 'team-external-review',
                            objective: 'Verify external account update',
                            execution_shape: 'deliverable',
                            state: verificationRecorded ? 'output_ready' : 'degraded',
                            needs_operator: !verificationRecorded,
                            degradation_state: verificationRecorded ? undefined : 'external_mutation_outcome_unknown',
                            recovery_options: verificationRecorded ? [] : ['verify the external result through Soma'],
                            work_intent: {
                                side_effect: {
                                    effect_kind: 'external_mutation',
                                    retry_safety: 'unknown',
                                    side_effect_state: verificationRecorded ? 'committed' : 'unknown',
                                },
                            },
                            updated_at: '2026-08-12T10:00:00Z',
                        },
                    ],
                }),
            });
        });
        await page.route('**/api/v1/teams/team-external-review/work/work-external-review-1/actions', async (route) => {
            submittedVerification = route.request().postDataJSON() as Record<string, unknown>;
            verificationRecorded = true;
            await route.fulfill({
                status: 200,
                contentType: 'application/json',
                body: JSON.stringify({ data: { status: 'verified' } }),
            });
        });

        await page.goto('/teams?view=work', { waitUntil: 'domcontentloaded' });

        await expect(page.getByLabel('Review details for Verify external account update')).toBeVisible();
        await expect(page.getByRole('button', { name: /Retry recovery/i })).toHaveCount(0);
        await expect(page.getByLabel('Verify external result for Verify external account update')).toBeVisible();
        await expect(page.getByText(/could not confirm whether the change completed/i)).toBeVisible();
        await expect(page.getByText(/Do not retry until the result is verified/i)).toBeVisible();

        await page.getByRole('radio', { name: 'Confirmed applied' }).click();
        await page.getByRole('textbox', { name: /What did you observe/i }).fill('The updated email address is visible on the account.');
        await page.getByRole('textbox', { name: /Evidence references/i }).fill('receipt-account-42');
        await page.getByRole('button', { name: 'Submit verification' }).click();

        await expect.poll(() => submittedVerification).toMatchObject({
            action: 'verify_external_outcome',
            summary: 'The updated email address is visible on the account.',
            payload: {
                result: 'committed',
                evidence_refs: ['receipt-account-42'],
            },
        });
        await expect(page.getByText(
            'Verification recorded: the external change was observed. No retry was requested.',
            { exact: true },
        )).toBeVisible();
        await expect(page.getByLabel('Verify external result for Verify external account update')).toHaveCount(0);
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
        if (await emptyState.isVisible().catch(() => false)) {
            return;
        }

        await expect(cards.first()).toBeVisible();
        const count = await cards.count();
        await expect(openChatLinks).toHaveCount(count);
        await expect(viewRunsLinks).toHaveCount(count);
        await expect(page.locator('a[data-testid$="-view-wiring"]')).toHaveCount(0);
        await expect(page.locator('a[data-testid$="-view-logs"]')).toHaveCount(0);

        await expect(openChatLinks.first()).toHaveAttribute('href', /\/dashboard\?team_id=/);
        await expect(viewRunsLinks.first()).toHaveAttribute('href', '/runs');
    });

    test('clicking a team card opens and closes the detail drawer', async ({ page }) => {
        await page.route('**/api/v1/teams/detail', async (route) => {
            await route.fulfill({
                status: 200,
                contentType: 'application/json',
                body: JSON.stringify([{
                    id: 'team-drawer-proof',
                    name: 'Drawer Proof Team',
                    role: 'action',
                    type: 'mission',
                    mission_id: 'mission-drawer-proof',
                    mission_intent: 'Prove the team detail drawer',
                    inputs: ['swarm.team.team-drawer-proof.internal.command'],
                    deliveries: ['swarm.team.team-drawer-proof.signal.result'],
                    agents: [],
                }]),
            });
        });
        await page.reload({ waitUntil: 'domcontentloaded' });

        const cards = page.getByRole('button')
            .filter({ hasText: 'Open lead workspace' })
            .filter({ hasText: 'View runs' });
        const firstCard = cards.first();
        await expect(firstCard).toContainText('Drawer Proof Team');
        await firstCard.click();

        const drawer = page.locator('div.w-\\[480px\\]');
        await expect(drawer).toBeVisible();
        await expect(drawer.locator('text=Agent Roster')).toBeVisible();
        await expect(drawer.getByText('Operator controls')).toBeVisible();
        await expect(drawer.getByRole('link', { name: 'Open lead workspace' })).toHaveAttribute('href', /\/dashboard\?team_id=/);
        await expect(drawer.getByRole('link', { name: 'View runs' })).toHaveAttribute('href', '/runs');
        await expect(drawer.getByRole('link', { name: 'View outputs' })).toHaveAttribute('href', '/groups');
        await drawer.getByText('Advanced coordination topics').click();
        await expect(drawer.getByRole('link', { name: 'View wiring' })).toHaveAttribute('href', '/automations?tab=wiring');
        await expect(drawer.getByRole('link', { name: 'View system' })).toHaveAttribute('href', '/system?tab=services');
        await expect(firstCard).toHaveClass(/ring-1/);

        await drawer.locator('button').first().click();
        await expect(drawer).not.toBeVisible();
    });
});
