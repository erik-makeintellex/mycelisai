import { type Page, type Route } from '@playwright/test';

async function fulfillJson(route: Route, body: unknown, status = 200) {
    await route.fulfill({ status, contentType: 'application/json', body: JSON.stringify(body) });
}

export async function mockDocsRouteApis(page: Page) {
    await page.route('**/docs-api', async (route) => {
        await fulfillJson(route, {
            sections: [
                {
                    section: 'Architecture',
                    docs: [
                        { slug: 'ui-test-doc', label: 'UI Test Doc', path: 'docs/architecture-library/UI_AND_OPERATOR_EXPERIENCE_V7.md', description: 'test doc' },
                        { slug: 'workflow-variants-doc', label: 'Workflow Variants', path: 'docs/user/workflow-variants-and-plan-memory.md', description: 'linked doc' },
                    ],
                },
            ],
        });
    });

    await page.route('**/docs-api/ui-test-doc', async (route) => {
        await fulfillJson(route, {
            slug: 'ui-test-doc',
            label: 'UI Test Doc',
            content: '# UI Test Heading\n\nTesting docs route rendering.\n\nSee [Workflow Variants](workflow-variants-and-plan-memory.md).',
        });
    });

    await page.route('**/docs-api/workflow-variants-doc', async (route) => {
        await fulfillJson(route, {
            slug: 'workflow-variants-doc',
            label: 'Workflow Variants',
            content: '# Workflow Variants\n\nCompact lanes stay visible.',
        });
    });
}

type RunsBrowserProofOptions = {
    runId: string;
    failedRunId: string;
    missionId: string;
    now: string;
};

export async function mockRunsBrowserProofApis(page: Page, options: RunsBrowserProofOptions) {
    const { runId, failedRunId, missionId, now } = options;
    let showTerminalEvents = false;
    let failNextManualEventRequest = false;
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
        await fulfillJson(route, {
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
        });
    });

    await page.route(`**/api/v1/runs/${runId}/events**`, async (route) => {
        const isManualRefresh = route.request().headers()['x-mycelis-e2e-event-refresh'] === '1';
        if (failNextManualEventRequest && isManualRefresh) {
            failNextManualEventRequest = false;
            showTerminalEvents = true;
            await fulfillJson(route, { ok: false, error: 'run timeline temporarily unavailable' }, 503);
            return;
        }

        const baseEvent = {
            id: 'evt-ui-1',
            run_id: runId,
            tenant_id: 'default',
            event_type: 'mission.started',
            severity: 'info',
            source_agent: 'admin',
            source_team: 'admin-core',
            payload: { mission_id: missionId },
            emitted_at: now,
        };

        await fulfillJson(route, {
            ok: true,
            data: showTerminalEvents
                ? [
                      baseEvent,
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
                : [baseEvent],
        });
    });

    await page.route(`**/api/v1/runs/${runId}/conversation**`, async (route) => {
        const url = new URL(route.request().url());
        const agentFilter = url.searchParams.get('agent');
        const turns = agentFilter
            ? conversationTurns.filter((turn) => turn.agent_id === agentFilter)
            : conversationTurns;

        await fulfillJson(route, { ok: true, data: { turns } });
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

        await fulfillJson(route, { ok: true });
    });

    return {
        armManualRefreshFailure() {
            failNextManualEventRequest = true;
        },
    };
}
