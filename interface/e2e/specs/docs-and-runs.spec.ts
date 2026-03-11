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
                    content: '# UI Test Heading\n\nTesting docs route rendering.',
                }),
            });
        });

        await page.goto('/docs?doc=ui-test-doc');
        await page.waitForLoadState('domcontentloaded');

        await expect(page.getByText('Documentation')).toBeVisible();
        await expect(page.getByRole('button', { name: 'UI Test Doc' })).toBeVisible();
        await expect(page.getByRole('heading', { name: 'UI Test Heading' })).toBeVisible();

        await page.getByPlaceholder('Filter docs...').fill('missing-doc');
        await expect(page.getByText('No docs match "missing-doc"')).toBeVisible();
    });

    test('runs list route and run detail tabs render from API payloads', async ({ page }) => {
        const runId = 'run-ui-1234';
        const now = new Date().toISOString();

        await page.route('**/api/v1/runs*', async (route) => {
            await route.fulfill({
                status: 200,
                contentType: 'application/json',
                body: JSON.stringify({
                    ok: true,
                    data: [
                        {
                            id: runId,
                            mission_id: 'mission-ui-7777',
                            status: 'running',
                            started_at: now,
                        },
                    ],
                }),
            });
        });

        await page.route(`**/api/v1/runs/${runId}/events**`, async (route) => {
            await route.fulfill({
                status: 200,
                contentType: 'application/json',
                body: JSON.stringify({
                    ok: true,
                    data: [
                        {
                            id: 'evt-ui-1',
                            run_id: runId,
                            event_type: 'mission.started',
                            severity: 'info',
                            source_agent: 'admin',
                            source_team: 'admin-core',
                            payload: { mission_id: 'mission-ui-7777' },
                            emitted_at: now,
                        },
                    ],
                }),
            });
        });

        await page.route(`**/api/v1/runs/${runId}/conversation**`, async (route) => {
            await route.fulfill({
                status: 200,
                contentType: 'application/json',
                body: JSON.stringify({
                    ok: true,
                    data: {
                        turns: [
                            {
                                id: 'turn-ui-1',
                                run_id: runId,
                                session_id: 'session-ui-1',
                                agent_id: 'admin',
                                team_id: 'admin-core',
                                turn_index: 1,
                                role: 'assistant',
                                content: 'assistant response from run',
                                created_at: now,
                            },
                        ],
                    },
                }),
            });
        });

        await page.goto('/runs');
        await page.waitForLoadState('domcontentloaded');

        await expect(page.locator('span:has-text("Runs")').first()).toBeVisible();
        const runRow = page.locator(`button:has-text("${runId}")`).first();
        await expect(runRow).toBeVisible();
        await runRow.click();

        await expect(page).toHaveURL(new RegExp(`/runs/${runId}$`));
        await expect(page.getByText('assistant response from run')).toBeVisible();

        await page.getByRole('button', { name: 'Events' }).first().click();
        await expect(page.getByText('mission.started')).toBeVisible();
    });
});
