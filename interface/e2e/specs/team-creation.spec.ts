import { test, expect, type Page, type Route } from '@playwright/test';

async function fulfillJSON(route: Route, status: number, body: unknown) {
    await route.fulfill({
        status,
        contentType: 'application/json',
        body: JSON.stringify(body),
    });
}

async function gotoWithColdStartRetry(page: Page, path: string) {
    try {
        await page.goto(path, { waitUntil: 'domcontentloaded' });
    } catch (error) {
        const message = error instanceof Error ? error.message : String(error);
        if (!message.includes('net::ERR_ABORTED') && !message.includes('frame was detached')) {
            throw error;
        }
        await page.goto(path, { waitUntil: 'domcontentloaded' });
    }
}

test.describe('Guided Team Creation (/teams/create)', () => {
    test.setTimeout(120_000);

    test.beforeEach(async ({ page }) => {
        const groups = [
            {
                group_id: 'group-temp-launch',
                name: 'Launch Delivery Team temporary workflow',
                goal_statement: 'Create a temporary marketing launch team for a new product rollout.',
                work_mode: 'propose_only',
                member_user_ids: ['owner', 'marketing-lead'],
                team_ids: ['launch-lead', 'design-lead'],
                coordinator_profile: 'Launch Delivery Team lead',
                approval_policy_ref: 'standard',
                status: 'active',
                expiry: '2026-04-13T18:00:00Z',
                created_by: 'owner',
                created_at: '2026-04-10T10:20:00Z',
            },
        ];

        await page.addInitScript(() => {
            window.localStorage.setItem('mycelis-last-organization-id', 'org-123');
            window.localStorage.setItem('mycelis-last-organization-name', 'Northstar Labs');
        });

        await page.route('**/api/v1/organizations/org-123/home', async (route) => {
            await fulfillJSON(route, 200, {
                ok: true,
                data: {
                    id: 'org-123',
                    name: 'Northstar Labs',
                    purpose: 'Ship a focused AI engineering organization for product delivery.',
                    start_mode: 'template',
                    team_lead_label: 'Launch Lead',
                    advisor_count: 1,
                    department_count: 2,
                    specialist_count: 4,
                    ai_engine_settings_summary: 'Balanced',
                    response_contract_summary: 'Clear and structured',
                    memory_personality_summary: 'Stable continuity',
                    status: 'active',
                },
            });
        });

        await page.route('**/api/v1/organizations/org-123/workspace/actions', async (route) => {
            const body = route.request().postDataJSON();
            await fulfillJSON(route, 200, {
                ok: true,
                data: {
                    action: body.action,
                    request_label: 'Run a quick strategy check',
                    headline: 'Launch team plan for Northstar Labs',
                    summary: 'Soma recommends a launch-focused team with clear outputs.',
                    priority_steps: [
                        'Confirm the campaign outcomes the team owns.',
                        'Choose the first output to make visible this week.',
                    ],
                    suggested_follow_ups: ['Choose the first priority'],
                    execution_contract: {
                        execution_mode: 'native_team',
                        owner_label: 'Native Mycelis team',
                        team_name: 'Launch Delivery Team',
                        summary: 'Use a focused launch delivery team inside Northstar Labs.',
                        target_outputs: ['Campaign brief', 'Landing page draft'],
                        workflow_group: {
                            name: 'Launch Delivery Team temporary workflow',
                            goal_statement: 'Create a temporary marketing launch team for a new product rollout.',
                            work_mode: 'propose_only',
                            coordinator_profile: 'Launch Delivery Team lead',
                            allowed_capabilities: ['team.coordinate', 'artifact.review'],
                            expiry_hours: 72,
                            summary: 'Launch a temporary workflow group for Launch Delivery Team.',
                        },
                    },
                },
            });
        });

        await page.route('**/api/v1/groups', async (route) => {
            if (route.request().method() === 'GET') {
                await fulfillJSON(route, 200, { ok: true, data: groups });
                return;
            }
            await fulfillJSON(route, 201, {
                ok: true,
                data: {
                    group_id: 'group-temp-launch',
                    name: 'Launch Delivery Team temporary workflow',
                },
            });
        });

        await page.route('**/api/v1/groups/monitor', async (route) => {
            await fulfillJSON(route, 200, {
                ok: true,
                data: {
                    status: 'online',
                    published_count: 2,
                    last_group_id: 'group-temp-launch',
                    last_message: 'Prepare launch brief and asset bundle',
                    last_published_at: '2026-04-10T10:30:00Z',
                },
            });
        });

        await page.route('**/api/v1/groups/group-temp-launch/status', async (route) => {
            if (route.request().method() !== 'PATCH') {
                await fulfillJSON(route, 405, { ok: false, error: 'method not allowed' });
                return;
            }
            groups[0] = {
                ...groups[0],
                status: 'archived',
            };
            await fulfillJSON(route, 200, { ok: true, data: groups[0] });
        });

        await page.route('**/api/v1/groups/group-temp-launch/outputs?limit=8', async (route) => {
            await fulfillJSON(route, 200, {
                ok: true,
                data: [
                    {
                        id: 'artifact-launch-brief',
                        team_id: 'launch-lead',
                        agent_id: 'launch-lead',
                        artifact_type: 'document',
                        title: 'Launch Brief',
                        content_type: 'text/markdown',
                        content: '# Launch brief\n\n- Message pillars\n- Delivery plan',
                        metadata: {},
                        status: 'approved',
                        created_at: '2026-04-10T10:31:00Z',
                    },
                    {
                        id: 'artifact-asset-bundle',
                        team_id: 'design-lead',
                        agent_id: 'design-lead',
                        artifact_type: 'file',
                        title: 'Asset Bundle',
                        content_type: 'application/zip',
                        file_path: 'workspace/groups/launch-delivery-team/assets.zip',
                        metadata: {},
                        status: 'approved',
                        created_at: '2026-04-10T10:32:00Z',
                    },
                ],
            });
        });
    });

    test('guides the operator through current organization context, launches a temporary workflow, archives it, and keeps retained outputs reviewable', async ({ page }) => {
        await gotoWithColdStartRetry(page, '/teams/create');

        await expect(page.getByRole('heading', { name: 'Create a team through Soma' })).toBeVisible();
        await expect(page.getByText('Current organization')).toBeVisible();
        await expect(
            page.locator('p').filter({ hasText: /^Northstar Labs$/ }).first(),
        ).toBeVisible();
        await expect(page.getByRole('button', { name: 'Marketing launch team' })).toBeVisible();

        await page.getByRole('button', { name: 'Marketing launch team' }).click();
        await expect(page.getByLabel('Tell Soma what team or delivery lane you want to create')).toHaveValue(/marketing launch team/i);

        await page.getByRole('button', { name: 'Start team design' }).click();

        await expect(page.getByText('Launch team plan for Northstar Labs')).toBeVisible();
        await expect(page.getByText('Launch Delivery Team', { exact: true })).toBeVisible();
        await expect(page.getByText('Campaign brief')).toBeVisible();
        await page.getByRole('button', { name: 'Create temporary workflow group' }).click();
        await expect(page.getByText(/Soma launched Launch Delivery Team temporary workflow/i)).toBeVisible();
        const openGroupLink = page.getByRole('link', { name: 'Open Launch Delivery Team temporary workflow' });
        await expect(openGroupLink).toHaveAttribute('href', '/groups?group_id=group-temp-launch');

        await openGroupLink.click();

        await expect(page.getByRole('heading', { name: 'Create, review, and coordinate focused groups.' })).toBeVisible();
        await expect(page.getByRole('heading', { name: 'Launch Delivery Team temporary workflow' })).toBeVisible();
        await expect(page.getByText('Temporary group', { exact: true })).toBeVisible();
        await expect(page.getByTestId('groups-output-summary')).toContainText('2 outputs');
        await expect(page.getByTestId('groups-output-summary')).toContainText('2 contributing leads');
        await expect(page.getByText('Launch Brief', { exact: true })).toBeVisible();
        await expect(page.getByText('Asset Bundle', { exact: true })).toBeVisible();
        await expect(page.getByRole('link', { name: 'Open Soma admin home' })).toHaveAttribute('href', '/dashboard');
        await expect(page.getByRole('link', { name: 'Open launch-lead lead' })).toHaveAttribute('href', '/dashboard?team_id=launch-lead');
        await expect(page.getByRole('link', { name: 'Open design-lead lead' })).toHaveAttribute('href', '/dashboard?team_id=design-lead');
        await expect(page.getByRole('link', { name: 'Download' }).last()).toHaveAttribute('href', '/api/v1/artifacts/artifact-asset-bundle/download');

        await page.getByRole('button', { name: 'Archive temporary group' }).click();

        await expect(page.getByTestId('groups-notice')).toContainText('Temporary group archived.');
        await expect(page.getByText('Archived temporary group', { exact: true })).toBeVisible();
        await expect(page.getByTestId('groups-archived-readonly-note')).toContainText('retained output review');
        await expect(page.getByTestId('groups-retained-outputs-note')).toContainText('Downloads remain available');
        await expect(page.getByTestId('groups-output-summary')).toContainText('2 outputs');
        await expect(page.getByTestId('groups-output-summary')).toContainText('2 contributing leads');
        await expect(page.getByText('Launch Brief', { exact: true })).toBeVisible();
        await expect(page.getByText('Asset Bundle', { exact: true })).toBeVisible();
        await expect(page.getByRole('button', { name: 'Broadcast to group' })).toHaveCount(0);
    });
});
