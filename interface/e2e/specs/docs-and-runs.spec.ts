import { test, expect } from '@playwright/test';

test.describe('Docs and Runs Route Coverage', () => {
    test('docs route renders manifest entry and markdown content', async ({ page }) => {
        await page.route('**/docs-api', async (route) => {
            await route.fulfill({
                status: 200,
                contentType: 'application/json',
                body: JSON.stringify({
                    sections: [
                        {
                            section: 'Architecture',
                            docs: [
                                {
                                    slug: 'ui-test-doc',
                                    label: 'UI Test Doc',
                                    path: 'docs/architecture-library/UI_AND_OPERATOR_EXPERIENCE_V7.md',
                                    description: 'test doc',
                                },
                                {
                                    slug: 'workflow-variants-doc',
                                    label: 'Workflow Variants',
                                    path: 'docs/user/workflow-variants-and-plan-memory.md',
                                    description: 'linked doc',
                                },
                            ],
                        },
                    ],
                }),
            });
        });

        await page.route('**/docs-api/ui-test-doc', async (route) => {
            await route.fulfill({
                status: 200,
                contentType: 'application/json',
                body: JSON.stringify({
                    slug: 'ui-test-doc',
                    label: 'UI Test Doc',
                    content: '# UI Test Heading\n\nTesting docs route rendering.\n\nSee [Workflow Variants](workflow-variants-and-plan-memory.md).',
                }),
            });
        });

        await page.route('**/docs-api/workflow-variants-doc', async (route) => {
            await route.fulfill({
                status: 200,
                contentType: 'application/json',
                body: JSON.stringify({
                    slug: 'workflow-variants-doc',
                    label: 'Workflow Variants',
                    content: '# Workflow Variants\n\nCompact lanes stay visible.',
                }),
            });
        });

        await page.goto('/docs?doc=ui-test-doc', { waitUntil: 'domcontentloaded' });

        await expect(page.locator('body')).toContainText('Documentation', { timeout: 20_000 });
        await expect(page.getByText('UI Test Doc').first()).toBeVisible();
        await expect(page.getByRole('heading', { name: 'UI Test Heading' })).toBeVisible();

        await page.locator('.max-w-3xl').getByRole('button', { name: 'Workflow Variants' }).click();
        await expect(page).toHaveURL(/\/docs\?doc=workflow-variants-doc$/);
        await expect(page.getByRole('heading', { name: 'Workflow Variants' })).toBeVisible();

        await page.getByPlaceholder('Filter docs...').fill('missing-doc');
        await expect(page.getByText('No docs match "missing-doc"')).toBeVisible();
    });

    test('runs browser proof covers list to detail review, interjection, failure evidence, and chain navigation', async ({ page }) => {
        const runId = 'run-ui-active-1234';
        const failedRunId = 'run-ui-failed-9999';
        const childRunId = 'run-ui-child-0001';
        const missionId = 'mission-ui-7777';
        const now = new Date().toISOString();
        let showTerminalEvents = false;
        let failNextEventRequest = false;
        let conversationTurns = [
            {
                id: 'turn-ui-1',
                run_id: runId,
                session_id: 'session-ui-1',
                agent_id: 'admin',
                team_id: 'admin-core',
                turn_index: 1,
                role: 'assistant',
                content: 'Soma is coordinating the active run.',
                provider_id: 'local',
                model_used: 'operator-proof',
                created_at: now,
            },
            {
                id: 'turn-ui-2',
                run_id: runId,
                session_id: 'session-ui-1',
                agent_id: 'planner',
                team_id: 'delivery',
                turn_index: 2,
                role: 'assistant',
                content: 'Planner is preparing validation steps.',
                created_at: now,
            },
        ];

        await page.route(/\/api\/v1\/runs(?:\?.*)?$/, async (route) => {
            await route.fulfill({
                status: 200,
                contentType: 'application/json',
                body: JSON.stringify({
                    ok: true,
                    data: [
                        {
                            id: runId,
                            mission_id: missionId,
                            status: 'running',
                            started_at: now,
                            tenant_id: 'default',
                            run_depth: 0,
                        },
                        {
                            id: failedRunId,
                            mission_id: 'mission-ui-failed',
                            status: 'failed',
                            started_at: now,
                            completed_at: now,
                            tenant_id: 'default',
                            run_depth: 0,
                        },
                    ],
                }),
            });
        });

        await page.route(`**/api/v1/runs/${runId}/events**`, async (route) => {
            if (failNextEventRequest) {
                failNextEventRequest = false;
                showTerminalEvents = true;
                await route.fulfill({
                    status: 503,
                    contentType: 'application/json',
                    body: JSON.stringify({ ok: false, error: 'run timeline temporarily unavailable' }),
                });
                return;
            }

            await route.fulfill({
                status: 200,
                contentType: 'application/json',
                body: JSON.stringify({
                    ok: true,
                    data: showTerminalEvents
                        ? [
                              {
                                  id: 'evt-ui-1',
                                  run_id: runId,
                                  tenant_id: 'default',
                                  event_type: 'mission.started',
                                  severity: 'info',
                                  source_agent: 'admin',
                                  source_team: 'admin-core',
                                  payload: { mission_id: missionId },
                                  emitted_at: now,
                              },
                              {
                                  id: 'evt-ui-2',
                                  run_id: runId,
                                  tenant_id: 'default',
                                  event_type: 'tool.failed',
                                  severity: 'error',
                                  source_agent: 'planner',
                                  source_team: 'delivery',
                                  payload: { error: 'Planner validation provider timed out; operator retry is available.' },
                                  emitted_at: now,
                              },
                              {
                                  id: 'evt-ui-3',
                                  run_id: runId,
                                  tenant_id: 'default',
                                  event_type: 'mission.failed',
                                  severity: 'error',
                                  source_agent: 'admin',
                                  source_team: 'admin-core',
                                  payload: { error: 'Mission stopped after retry budget was exhausted.' },
                                  emitted_at: now,
                              },
                          ]
                        : [
                              {
                                  id: 'evt-ui-1',
                                  run_id: runId,
                                  tenant_id: 'default',
                                  event_type: 'mission.started',
                                  severity: 'info',
                                  source_agent: 'admin',
                                  source_team: 'admin-core',
                                  payload: { mission_id: missionId },
                                  emitted_at: now,
                              },
                          ],
                }),
            });
        });

        await page.route(`**/api/v1/runs/${runId}/conversation**`, async (route) => {
            const url = new URL(route.request().url());
            const agentFilter = url.searchParams.get('agent');
            const turns = agentFilter
                ? conversationTurns.filter((turn) => turn.agent_id === agentFilter)
                : conversationTurns;

            await route.fulfill({
                status: 200,
                contentType: 'application/json',
                body: JSON.stringify({
                    ok: true,
                    data: { turns },
                }),
            });
        });

        await page.route(`**/api/v1/runs/${runId}/interject`, async (route) => {
            const requestBody = route.request().postDataJSON() as { message?: string };
            conversationTurns = [
                ...conversationTurns,
                {
                    id: 'turn-ui-interjection',
                    run_id: runId,
                    session_id: 'session-ui-1',
                    agent_id: 'operator',
                    team_id: 'admin-core',
                    turn_index: 3,
                    role: 'interjection',
                    content: requestBody.message ?? '',
                    created_at: now,
                },
            ];

            await route.fulfill({
                status: 200,
                contentType: 'application/json',
                body: JSON.stringify({ ok: true }),
            });
        });

        await page.route(`**/api/v1/runs/${runId}/chain**`, async (route) => {
            await route.fulfill({
                status: 200,
                contentType: 'application/json',
                body: JSON.stringify({
                    run_id: runId,
                    mission_id: missionId,
                    chain: [
                        {
                            id: runId,
                            mission_id: missionId,
                            tenant_id: 'default',
                            status: 'failed',
                            run_depth: 0,
                            started_at: now,
                            completed_at: now,
                            metadata: {
                                source_kind: 'workspace_ui',
                                team_id: 'admin-core',
                            },
                        },
                        {
                            id: childRunId,
                            mission_id: missionId,
                            tenant_id: 'default',
                            status: 'running',
                            run_depth: 1,
                            parent_run_id: runId,
                            started_at: now,
                            metadata: {
                                source_kind: 'automation_trigger',
                                team_id: 'delivery',
                            },
                        },
                    ],
                }),
            });
        });

        await page.goto('/runs', { waitUntil: 'domcontentloaded' });

        await expect(page.locator('span:has-text("Runs")').first()).toBeVisible();
        await expect(page.getByText('1 active')).toBeVisible();
        await expect(page.getByText(runId, { exact: true })).toBeVisible();
        await expect(page.getByText(failedRunId, { exact: true })).toBeVisible();
        await expect(page.getByText('running').first()).toBeVisible();
        await expect(page.getByText('failed').first()).toBeVisible();

        const runRow = page.locator(`button:has-text("${runId}")`).first();
        await expect(runRow).toBeVisible();
        await runRow.click();

        await expect(page).toHaveURL(new RegExp(`/runs/${runId}$`));
        await expect(page.getByText('Soma is coordinating the active run.')).toBeVisible();
        await expect(page.getByText('Planner is preparing validation steps.')).toBeVisible();
        await expect(page.getByText('running').first()).toBeVisible();
        await expect(page.getByPlaceholder('Interject in this run...')).toBeVisible();

        await page.getByRole('button', { name: 'planner' }).click();
        await expect(page.getByText('Planner is preparing validation steps.')).toBeVisible();
        await expect(page.getByText('Soma is coordinating the active run.')).not.toBeVisible();

        await page.getByPlaceholder('Interject in this run...').fill('Pause this run and retry with a smaller validation step.');
        await page.getByRole('button', { name: 'Interject' }).click();
        await expect(page.getByText('Operator Interjection')).toBeVisible();
        await expect(page.getByText('Pause this run and retry with a smaller validation step.')).toBeVisible();

        failNextEventRequest = true;
        await page.getByRole('button', { name: 'Events' }).first().click();
        await expect(page.getByText('Failed to load events (503)')).toBeVisible();
        await page.getByRole('button', { name: 'Retry' }).click();
        await expect(page.getByText('mission.failed')).toBeVisible();
        await expect(page.getByText('tool.failed')).toBeVisible();
        await expect(page.getByText('planner').first()).toBeVisible();
        await expect(page.getByText('Planner validation provider timed out; operator retry is available.')).toBeVisible();
        await expect(page.getByText('Mission stopped after retry budget was exhausted.')).toBeVisible();
        await expect(page.getByText('failed').first()).toBeVisible();
        await expect(page.getByText('mission.started')).toBeVisible();

        await page.getByRole('link', { name: 'Chain' }).click();
        await expect(page).toHaveURL(new RegExp(`/runs/${runId}/chain$`));
        await expect(page.getByRole('heading', { name: 'Causal Chain' })).toBeVisible();
        await expect(page.getByText('2 runs')).toBeVisible();
        await expect(page.getByText(runId, { exact: true })).toBeVisible();
        await expect(page.getByText(childRunId, { exact: true })).toBeVisible();
        await expect(page.getByText('depth 0')).toBeVisible();
        await expect(page.getByText('depth 1')).toBeVisible();
        await expect(page.getByText(`parent ${runId}`)).toBeVisible();
        await expect(page.getByText('source_kind: workspace_ui')).toBeVisible();
        await expect(page.getByText('source_kind: automation_trigger')).toBeVisible();
        await expect(page.getByText('team_id: delivery')).toBeVisible();
        await expect(page.getByRole('link', { name: 'Run' })).toHaveAttribute('href', `/runs/${runId}`);
    });

    test('run chain route renders lineage from API payloads', async ({ page }) => {
        const runId = 'run-chain-ui-5678';
        const now = new Date().toISOString();

        await page.route(`**/api/v1/runs/${runId}/chain**`, async (route) => {
            await route.fulfill({
                status: 200,
                contentType: 'application/json',
                body: JSON.stringify({
                    run_id: runId,
                    mission_id: 'mission-chain-1111',
                    chain: [
                        {
                            id: runId,
                            mission_id: 'mission-chain-1111',
                            tenant_id: 'default',
                            status: 'completed',
                            run_depth: 0,
                            started_at: now,
                            completed_at: now,
                            metadata: {
                                source_kind: 'workspace_ui',
                            },
                        },
                        {
                            id: 'run-chain-child-0001',
                            mission_id: 'mission-chain-1111',
                            tenant_id: 'default',
                            status: 'running',
                            run_depth: 1,
                            parent_run_id: runId,
                            started_at: now,
                            metadata: {
                                team_id: 'team-alpha',
                            },
                        },
                    ],
                }),
            });
        });

        await page.goto(`/runs/${runId}/chain`, { waitUntil: 'domcontentloaded' });

        await expect(page.getByRole('heading', { name: 'Causal Chain' })).toBeVisible();
        await expect(page.getByText(runId, { exact: true })).toBeVisible();
        await expect(page.getByText('run-chain-child-0001', { exact: true })).toBeVisible();
        await expect(page.getByText('source_kind: workspace_ui')).toBeVisible();
        await expect(page.getByText('team_id: team-alpha')).toBeVisible();
        await expect(page.getByRole('link', { name: 'Run' })).toHaveAttribute('href', `/runs/${runId}`);
    });
});
