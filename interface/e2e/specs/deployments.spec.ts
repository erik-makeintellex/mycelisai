import { test, expect, type Page, type Route } from '@playwright/test';
import { createdTemplateOrganization, mockOrganizationEntryApis, openTeamDesignLane } from '../support/organization-entry';

type DeploymentKind = 'standing_team' | 'mission_team' | 'group' | 'planned_department';

type DeploymentRecord = {
    id: string;
    kind: DeploymentKind;
    name: string;
    purpose?: string;
    what_runs: string;
    scope: { kind: string; id?: string; name?: string };
    origin: string;
    runtime_id?: string;
    lifecycle_state: string;
    since?: string | null;
    expires_at?: string | null;
    last_activity?: string | null;
    members: Array<{
        id: string;
        role: string;
        provider?: string;
        model?: string;
        provider_origin?: string;
        status?: string;
        tools: Array<{ name: string; source: string }>;
    }>;
    tools: Array<{ name: string; source: string }>;
    work_mode?: string;
    approval_policy_ref?: string;
    approval_posture?: string;
    outputs_summary?: { retained: number; by_status: Record<string, number>; proof_status: string; last_output_at?: string | null } | null;
    config_ref: { kind: string; id: string; path: string };
};

type SourceStatus = { source: string; status: 'ok' | 'degraded' | 'unavailable' | 'restricted'; detail?: string; count: number };

async function fulfillJSON(route: Route, status: number, body: unknown) {
    await route.fulfill({
        status,
        contentType: 'application/json',
        body: JSON.stringify(body),
    });
}

function baseRecords(): DeploymentRecord[] {
    return [
        {
            id: 'launch-lead', kind: 'standing_team', name: 'Launch Delivery Team',
            purpose: 'Coordinate launch delivery work.',
            what_runs: 'Standing team coordinating campaign launch delivery.',
            scope: { kind: 'global' }, origin: 'bundle', runtime_id: 'launch-lead',
            lifecycle_state: 'running', since: '2026-08-01T10:00:00Z', expires_at: null,
            last_activity: '2026-09-24T09:00:00Z',
            members: [{
                id: 'launch-lead', role: 'Team Lead', provider: 'ollama', model: 'qwen2.5-coder:7b',
                provider_origin: 'manifest', status: 'online',
                tools: [{ name: 'store_inception_recipe', source: 'internal' }],
            }],
            tools: [{ name: 'store_inception_recipe', source: 'internal' }],
            work_mode: 'propose_only', approval_policy_ref: 'standard',
            approval_posture: 'Proposals require operator approval before executing.',
            outputs_summary: { retained: 3, by_status: { approved: 3 }, proof_status: 'not_recorded', last_output_at: '2026-09-24T08:55:00Z' },
            config_ref: { kind: 'teams', id: 'launch-lead', path: '/api/v1/config/teams/launch-lead' },
        },
        {
            id: 'mission-team-1', kind: 'mission_team', name: 'Q3 Rollout Mission Team',
            purpose: 'Deliver the Q3 rollout mission.',
            what_runs: 'Mission-scoped team executing the Q3 rollout.',
            scope: { kind: 'mission', id: 'mission-1', name: 'Q3 Rollout' },
            origin: 'mission:mission-1', runtime_id: 'mission-team-1', lifecycle_state: 'idle',
            since: '2026-09-10T12:00:00Z', expires_at: '2026-10-10T12:00:00Z',
            last_activity: '2026-09-20T15:30:00Z',
            members: [{
                id: 'mission-team-1-lead', role: 'Coordinator', provider: 'ollama',
                model: 'qwen2.5-coder:7b', provider_origin: 'profile', status: 'idle', tools: [],
            }],
            tools: [], work_mode: 'execute_with_approval', approval_policy_ref: 'high-impact',
            approval_posture: undefined, outputs_summary: null,
            config_ref: { kind: 'teams', id: 'mission-team-1', path: '/api/v1/config/teams/mission-team-1' },
        },
        {
            id: 'group-standing', kind: 'group', name: 'Standing Revenue Ops',
            purpose: 'Keep revenue operations visible.',
            what_runs: 'Standing collaboration group for revenue operations.',
            scope: { kind: 'global' }, origin: 'operator', lifecycle_state: 'running',
            since: '2026-06-01T09:00:00Z', expires_at: null, last_activity: '2026-09-23T11:00:00Z',
            members: [{
                id: 'revops-lead', role: 'Coordinator', provider: 'ollama',
                model: 'qwen2.5-coder:7b', status: 'online', tools: [],
            }],
            tools: [], work_mode: 'propose_only', approval_policy_ref: 'standard',
            approval_posture: 'Proposals require operator approval before executing.',
            outputs_summary: { retained: 0, by_status: {}, proof_status: 'not_recorded', last_output_at: null },
            config_ref: { kind: 'groups', id: 'group-standing', path: '/api/v1/config/groups/group-standing' },
        },
        {
            id: 'org-123:platform', kind: 'planned_department', name: 'Platform Department',
            purpose: 'Planned platform department.',
            what_runs: 'Nothing runs until the operator approves a launch.',
            scope: { kind: 'organization', id: 'org-123', name: 'Northstar Labs' },
            origin: 'template:engineering-starter', lifecycle_state: 'planned',
            since: null, expires_at: null, last_activity: null, members: [], tools: [],
            work_mode: undefined, approval_policy_ref: undefined, approval_posture: undefined,
            outputs_summary: undefined,
            config_ref: { kind: 'organizations', id: 'org-123', path: '/api/v1/config/organizations/org-123' },
        },
    ];
}

function baseSources(): SourceStatus[] {
    return [
        { source: 'standing_teams', status: 'ok', count: 1 },
        { source: 'mission_teams', status: 'ok', count: 1 },
        { source: 'groups', status: 'ok', count: 1 },
        { source: 'organizations', status: 'ok', count: 1 },
        { source: 'outputs', status: 'ok', count: 2 },
    ];
}

async function mockDeployments(
    page: Page,
    options?: { records?: DeploymentRecord[]; sources?: SourceStatus[]; failStatus?: number },
) {
    const records = options?.records ?? baseRecords();
    const sources = options?.sources ?? baseSources();

    await page.route(/\/api\/v1\/deployments(?:\?.*)?$/, async (route) => {
        if (route.request().method() !== 'GET') {
            await route.fallback();
            return;
        }
        if (options?.failStatus) {
            await fulfillJSON(route, options.failStatus, { ok: false, error: 'Deployments request failed.' });
            return;
        }
        const url = new URL(route.request().url());
        const kind = url.searchParams.get('kind');
        const state = url.searchParams.get('state');
        let filtered = records;
        if (kind) filtered = filtered.filter((r) => r.kind === kind);
        if (state) filtered = filtered.filter((r) => r.lifecycle_state === state);
        await fulfillJSON(route, 200, {
            ok: true,
            data: { deployments: filtered, sources, generated_at: '2026-09-25T12:00:00Z' },
        });
    });
}

async function enableAdvancedMode(page: Page) {
    await page.addInitScript(() => {
        window.localStorage.setItem('mycelis-advanced-mode', 'true');
    });
}

test.describe('Deployments (/deployments)', () => {
    test('lists deployments across all kinds and filters by kind and state', async ({ page }) => {
        await mockDeployments(page);
        await page.goto('/deployments', { waitUntil: 'domcontentloaded' });

        await expect(page.getByRole('heading', { name: 'Deployments' })).toBeVisible();
        await expect(page.getByTestId('deployments-list')).toBeVisible();
        await expect(page.getByTestId('deployments-list-item-standing_team:launch-lead')).toBeVisible();
        await expect(page.getByTestId('deployments-list-item-mission_team:mission-team-1')).toBeVisible();
        await expect(page.getByTestId('deployments-list-item-group:group-standing')).toBeVisible();
        await expect(page.getByTestId('deployments-list-item-planned_department:org-123:platform')).toBeVisible();

        await page.getByTestId('deployments-filter-kind-group').click();
        await expect(page.getByTestId('deployments-list-item-group:group-standing')).toBeVisible();
        await expect(page.getByTestId('deployments-list-item-standing_team:launch-lead')).toHaveCount(0);

        await page.getByTestId('deployments-filter-kind-all').click();
        await page.getByTestId('deployments-filter-state-planned').click();
        await expect(page.getByTestId('deployments-list-item-planned_department:org-123:platform')).toBeVisible();
        await expect(page.getByTestId('deployments-list-item-standing_team:launch-lead')).toHaveCount(0);
    });

    test('selecting a row shows the detail pane', async ({ page }) => {
        await mockDeployments(page);
        await page.goto('/deployments', { waitUntil: 'domcontentloaded' });

        await page.getByTestId('deployments-list-item-mission_team:mission-team-1').click();
        await expect(page.getByTestId('deployments-detail')).toBeVisible();
        await expect(page.getByTestId('deployments-detail').getByRole('heading', { name: 'Q3 Rollout Mission Team' })).toBeVisible();
        await expect(page.getByTestId('deployments-members')).toBeVisible();
    });

    test('outputs summary shows "unknown" when absent', async ({ page }) => {
        await mockDeployments(page);
        await page.goto('/deployments', { waitUntil: 'domcontentloaded' });

        await page.getByTestId('deployments-list-item-planned_department:org-123:platform').click();
        await expect(page.getByTestId('deployments-outputs-summary')).toContainText('unknown');
    });

    test('shows a restricted-source notice', async ({ page }) => {
        await mockDeployments(page, {
            sources: [
                { source: 'standing_teams', status: 'ok', count: 1 },
                { source: 'mission_teams', status: 'ok', count: 1 },
                { source: 'groups', status: 'restricted', count: 0 },
                { source: 'organizations', status: 'ok', count: 1 },
                { source: 'outputs', status: 'ok', count: 2 },
            ],
        });
        await page.goto('/deployments', { waitUntil: 'domcontentloaded' });

        await expect(page.getByTestId('deployments-source-notices')).toBeVisible();
        await expect(page.getByTestId('deployments-source-notice-groups')).toContainText(/admin/i);
    });

    test('shows a Core-down failure state', async ({ page }) => {
        await mockDeployments(page, { failStatus: 503 });
        await page.goto('/deployments', { waitUntil: 'domcontentloaded' });

        await expect(page.getByTestId('deployments-load-failure')).toBeVisible();
        await expect(page.getByTestId('deployments-load-failure')).toContainText('Could not load deployments');
        await expect(page.getByTestId('deployments-load-failure').getByRole('button', { name: 'Retry' })).toBeVisible();
    });

    test('View config is hidden when Advanced mode is off and visible when it is on', async ({ page }) => {
        await mockDeployments(page);
        await page.goto('/deployments', { waitUntil: 'domcontentloaded' });

        await page.getByTestId('deployments-list-item-standing_team:launch-lead').click();
        await expect(page.getByTestId('deployments-detail')).toBeVisible();
        await expect(page.getByTestId('deployments-view-config')).toHaveCount(0);

        await enableAdvancedMode(page);
        await page.goto('/deployments', { waitUntil: 'domcontentloaded' });
        await page.getByTestId('deployments-list-item-standing_team:launch-lead').click();
        await expect(page.getByTestId('deployments-view-config')).toBeVisible();
    });

    test('/teams lands on /deployments', async ({ page }) => {
        await mockDeployments(page);
        await page.goto('/teams', { waitUntil: 'domcontentloaded' });

        await expect(page).toHaveURL(/\/deployments$/);
        await expect(page.getByRole('heading', { name: 'Deployments' })).toBeVisible();
    });

    // Retired teams.spec.ts asserted the old /teams workspace linked to guided
    // team creation. That workspace is gone (/teams now redirects here), but
    // /teams/create remains a directly reachable route in its own right.
    test('guided team creation stays directly reachable at /teams/create', async ({ page }) => {
        await mockOrganizationEntryApis(page, {
            homeResponsesById: { [createdTemplateOrganization.id]: createdTemplateOrganization },
        });
        await page.addInitScript(
            ({ id, name }) => {
                window.localStorage.setItem('mycelis-last-organization-id', id);
                window.localStorage.setItem('mycelis-last-organization-name', name);
            },
            { id: createdTemplateOrganization.id, name: createdTemplateOrganization.name },
        );
        await openTeamDesignLane(page);
    });
});
