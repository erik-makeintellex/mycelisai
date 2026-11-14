import { test, expect } from '@playwright/test';
import { mockDocsRouteApis, mockRunsBrowserProofApis } from '../support/docs-and-runs-fixtures';

test.describe('Docs and Runs Route Coverage', () => {
    test('docs route renders manifest entry and markdown content', async ({ page }) => {
        await mockDocsRouteApis(page);

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

        const runMocks = await mockRunsBrowserProofApis(page, {
            runId,
            failedRunId,
            missionId,
            now,
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

        await page.getByRole('button', { name: 'Events' }).first().click();
        await expect(page.getByText('mission.started')).toBeVisible();
        runMocks.armManualRefreshFailure();
        await page.route(`**/api/v1/runs/${runId}/events**`, async (route) => {
            await route.fallback({ headers: { ...route.request().headers(), 'x-mycelis-e2e-event-refresh': '1' } });
        }, { times: 1 });
        await page.getByRole('button', { name: 'Refresh events' }).click();
        await expect(page.getByText('Failed to load events (503)')).toBeVisible();
        await page.getByRole('button', { name: 'Retry' }).last().click();
        await expect(page.getByText('mission.failed')).toBeVisible();
        await expect(page.getByText('tool.failed')).toBeVisible();
        await expect(page.getByText('planner').first()).toBeVisible();
        await expect(page.getByText('Planner validation provider timed out; operator retry is available.')).toBeVisible();
        await expect(page.getByText('Mission stopped after retry budget was exhausted.')).toBeVisible();

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
        await expect(page.getByRole('link', { name: 'Run', exact: true })).toHaveAttribute('href', `/runs/${runId}`);
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
        await expect(page.getByRole('link', { name: 'Run', exact: true })).toHaveAttribute('href', `/runs/${runId}`);
    });
});
