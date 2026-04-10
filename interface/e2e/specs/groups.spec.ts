import { expect, test, type Page } from '@playwright/test';

type GroupRecord = {
    group_id: string;
    name: string;
    goal_statement: string;
    work_mode: 'read_only' | 'propose_only' | 'execute_with_approval' | 'execute_bounded';
    member_user_ids: string[];
    team_ids: string[];
    coordinator_profile: string;
    approval_policy_ref: string;
    status: 'active' | 'paused' | 'archived';
    expiry?: string | null;
    created_by: string;
    created_at: string;
};

type ArtifactRecord = {
    id: string;
    team_id?: string;
    agent_id: string;
    artifact_type: 'document' | 'file' | 'code' | 'data' | 'chart' | 'image' | 'audio';
    title: string;
    content_type: string;
    content?: string;
    file_path?: string;
    metadata: Record<string, unknown>;
    status: 'pending' | 'approved' | 'rejected' | 'archived';
    created_at: string;
};

async function fulfillJson(route: { fulfill: (options: { status: number; contentType: string; body: string }) => Promise<void> }, status: number, body: unknown) {
    await route.fulfill({
        status,
        contentType: 'application/json',
        body: JSON.stringify(body),
    });
}

async function mockGroupsWorkspace(page: Page) {
    const createdAt = '2026-04-09T10:00:00Z';
    const groups: GroupRecord[] = [
        {
            group_id: 'group-standing',
            name: 'Standing Revenue Ops',
            goal_statement: 'Keep the revenue operations lane visible and coordinated.',
            work_mode: 'propose_only',
            member_user_ids: ['owner'],
            team_ids: ['revops-lead'],
            coordinator_profile: 'Revenue Ops lead',
            approval_policy_ref: 'standard',
            status: 'active',
            expiry: null,
            created_by: 'owner',
            created_at: createdAt,
        },
        {
            group_id: 'group-temp-launch',
            name: 'Temporary Launch Sprint',
            goal_statement: 'Drive a fast campaign launch across messaging, review, and delivery.',
            work_mode: 'execute_with_approval',
            member_user_ids: ['owner', 'marketing-lead'],
            team_ids: ['launch-lead', 'design-lead'],
            coordinator_profile: 'Launch Sprint lead',
            approval_policy_ref: 'high-impact',
            status: 'active',
            expiry: '2026-04-12T12:00:00Z',
            created_by: 'owner',
            created_at: createdAt,
        },
        {
            group_id: 'group-temp-archived',
            name: 'Archived Review Sprint',
            goal_statement: 'Keep the completed temporary collaboration available for retained output review.',
            work_mode: 'execute_with_approval',
            member_user_ids: ['owner', 'ops-lead'],
            team_ids: ['archive-review-lead'],
            coordinator_profile: 'Archive Review lead',
            approval_policy_ref: 'high-impact',
            status: 'archived',
            expiry: '2026-04-08T17:00:00Z',
            created_by: 'owner',
            created_at: createdAt,
        },
    ];

    const artifactsByGroup = new Map<string, ArtifactRecord[]>([
        ['group-temp-launch', [
            {
                id: 'artifact-brief',
                team_id: 'launch-lead',
                agent_id: 'launch-lead',
                artifact_type: 'document',
                title: 'Launch Brief',
                content_type: 'text/markdown',
                content: '# Launch brief\n\n- Message pillars\n- Campaign milestones',
                metadata: {},
                status: 'approved',
                created_at: '2026-04-09T10:10:00Z',
            },
        ]],
        ['design-lead', [
            {
                id: 'artifact-bundle',
                team_id: 'design-lead',
                agent_id: 'design-lead',
                artifact_type: 'file',
                title: 'Asset Bundle',
                content_type: 'application/zip',
                file_path: 'workspace/groups/temporary-launch-sprint/assets.zip',
                metadata: {},
                status: 'approved',
                created_at: '2026-04-09T10:12:00Z',
            },
        ]],
        ['archive-review-lead', [
            {
                id: 'artifact-archive-summary',
                team_id: 'archive-review-lead',
                agent_id: 'archive-review-lead',
                artifact_type: 'document',
                title: 'Retrospective Summary',
                content_type: 'text/markdown',
                content: '# Retained summary\n\n- Final decisions\n- Approved outputs',
                metadata: {},
                status: 'approved',
                created_at: '2026-04-08T16:55:00Z',
            },
        ]],
        ['group-temp-archived', [
            {
                id: 'artifact-retrospective',
                team_id: 'archive-review-lead',
                agent_id: 'archive-review-lead',
                artifact_type: 'document',
                title: 'Retrospective Summary',
                content_type: 'text/markdown',
                content: 'Approved outputs',
                metadata: {},
                status: 'approved',
                created_at: createdAt,
            },
        ]],
        ['group-standing', []],
    ]);

    const broadcastBodies: Array<Record<string, unknown>> = [];
    const statusBodies: Array<Record<string, unknown>> = [];

    await page.route('**/api/v1/groups/monitor', async (route) => {
        await fulfillJson(route, 200, {
            ok: true,
            data: {
                status: 'online',
                published_count: 4,
                last_group_id: 'group-temp-launch',
                last_message: 'Generate launch brief and asset bundle',
                last_published_at: '2026-04-09T10:13:00Z',
            },
        });
    });

    await page.route('**/api/v1/groups', async (route) => {
        if (route.request().method() === 'GET') {
            await fulfillJson(route, 200, { ok: true, data: groups });
            return;
        }
        const body = route.request().postDataJSON() as Record<string, unknown>;
        const created: GroupRecord = {
            group_id: 'group-temp-created',
            name: String(body.name ?? 'Created Group'),
            goal_statement: String(body.goal_statement ?? ''),
            work_mode: String(body.work_mode ?? 'propose_only') as GroupRecord['work_mode'],
            member_user_ids: Array.isArray(body.member_user_ids) ? (body.member_user_ids as string[]) : [],
            team_ids: Array.isArray(body.team_ids) ? (body.team_ids as string[]) : [],
            coordinator_profile: String(body.coordinator_profile ?? ''),
            approval_policy_ref: String(body.approval_policy_ref ?? ''),
            status: 'active',
            expiry: typeof body.expiry === 'string' && body.expiry.length > 0 ? body.expiry : null,
            created_by: 'owner',
            created_at: '2026-04-09T10:20:00Z',
        };
        groups.unshift(created);
        await fulfillJson(route, 200, { ok: true, data: created });
    });

    await page.route(/\/api\/v1\/groups\/[^/]+\/outputs\?limit=.*/, async (route) => {
        const url = new URL(route.request().url());
        const pathParts = url.pathname.split('/');
        const groupId = pathParts[pathParts.length - 2] ?? '';
        await fulfillJson(route, 200, { ok: true, data: artifactsByGroup.get(groupId) ?? [] });
    });

    await page.route('**/api/v1/groups/group-temp-launch/broadcast', async (route) => {
        broadcastBodies.push((route.request().postDataJSON() ?? {}) as Record<string, unknown>);
        await fulfillJson(route, 200, { ok: true, data: { queued: true } });
    });

    await page.route('**/api/v1/groups/group-temp-launch/status', async (route) => {
        const body = (route.request().postDataJSON() ?? {}) as Record<string, unknown>;
        statusBodies.push(body);
        groups[1] = {
            ...groups[1],
            status: String(body.status ?? 'archived') as GroupRecord['status'],
        };
        await fulfillJson(route, 200, { ok: true, data: groups[1] });
    });

    return {
        readBroadcastBodies: () => broadcastBodies,
        readStatusBodies: () => statusBodies,
    };
}

test.describe('Groups workspace (/groups)', () => {
    test('shows active and archived temporary groups with retained outputs review', async ({ page }) => {
        await mockGroupsWorkspace(page);

        await page.goto('/groups', { waitUntil: 'domcontentloaded' });

        await expect(page.getByRole('heading', { name: 'Create, review, and coordinate focused groups.' })).toBeVisible();
        await expect(page.getByRole('heading', { name: 'Standing groups', exact: true })).toBeVisible();
        await expect(page.getByRole('heading', { name: 'Temporary groups', exact: true })).toBeVisible();
        await expect(page.getByRole('heading', { name: 'Archived temporary groups', exact: true })).toBeVisible();

        await page.getByRole('button', { name: 'Temporary Launch Sprint Drive a fast campaign launch across messaging, review, and delivery.' }).click();

        await expect(page.getByText('Temporary group', { exact: true })).toBeVisible();
        await expect(page.getByText('Launch Brief', { exact: true })).toBeVisible();
        await expect(page.getByRole('link', { name: 'Open Soma admin home' })).toHaveAttribute('href', '/dashboard');
        await expect(page.getByRole('link', { name: 'Open launch-lead lead' })).toHaveAttribute('href', '/dashboard?team_id=launch-lead');
        await expect(page.getByRole('link', { name: 'Download' }).first()).toHaveAttribute('href', '/api/v1/artifacts/artifact-brief/download');

        await page.getByRole('button', { name: 'Archived Review Sprint Archived Keep the completed temporary collaboration available for retained output review.' }).click();
        await expect(page.getByText('Archived temporary group', { exact: true })).toBeVisible();
        await expect(page.getByTestId('groups-archived-readonly-note')).toContainText('retained output review');
        await expect(page.getByTestId('groups-retained-outputs-note')).toContainText('Downloads remain available');
        await expect(page.getByText('Retrospective Summary')).toBeVisible();
        await expect(page.getByText('Approved outputs')).toBeVisible();
        await expect(page.getByRole('button', { name: 'Broadcast to group' })).toHaveCount(0);
    });

    test('archives a temporary group and keeps retained outputs visible', async ({ page }) => {
        const harness = await mockGroupsWorkspace(page);

        await page.goto('/groups', { waitUntil: 'domcontentloaded' });

        await page.getByRole('button', { name: 'Temporary Launch Sprint Drive a fast campaign launch across messaging, review, and delivery.' }).click();
        await page.getByRole('button', { name: 'Archive temporary group' }).click();

        await expect(page.getByTestId('groups-notice')).toContainText('Temporary group archived.');
        await expect.poll(() => harness.readStatusBodies().length).toBe(1);
        await expect.poll(() => String(harness.readStatusBodies()[0]?.status ?? '')).toBe('archived');
        await expect(page.getByText('Archived temporary group', { exact: true })).toBeVisible();
        await expect(page.getByTestId('groups-archived-readonly-note')).toContainText('retained output review');
        await expect(page.getByTestId('groups-retained-outputs-note')).toContainText('Downloads remain available');
        await expect(page.getByText('Launch Brief', { exact: true })).toBeVisible();
        await expect(page.getByRole('link', { name: 'Download' }).first()).toHaveAttribute('href', '/api/v1/artifacts/artifact-brief/download');
        await expect(page.getByRole('button', { name: /Temporary Launch Sprint.*Archived.*Drive a fast campaign launch across messaging, review, and delivery\./ })).toBeVisible();
        await expect(page.getByRole('button', { name: 'Broadcast to group' })).toHaveCount(0);
    });

    test('creates a temporary group and supports broadcast workflow for business execution', async ({ page }) => {
        const harness = await mockGroupsWorkspace(page);

        await page.goto('/groups', { waitUntil: 'domcontentloaded' });

        await page.getByLabel('Name').fill('Regional Expansion Sprint');
        await page.getByLabel('Goal Statement').fill('Prepare outreach, pricing notes, and operator-ready launch assets for a new region.');
        await page.getByLabel('Work Mode').selectOption('execute_with_approval');
        await page.getByLabel('Expiry').fill('2026-04-15T09:30');
        await page.getByLabel('Team IDs').fill('launch-lead, design-lead');
        await page.getByLabel('Member IDs').fill('owner, marketing-lead');
        await page.getByLabel('Coordinator Profile').fill('Regional expansion lead');
        await page.getByLabel('Allowed Capabilities').fill('write_file, publish_signal');
        await page.getByLabel('Approval Policy Ref').fill('regional-expansion');
        await page.getByTestId('groups-create-button').click();

        await expect(page.getByTestId('groups-notice')).toContainText('Group created successfully.');

        await page.getByRole('button', { name: 'Temporary Launch Sprint Drive a fast campaign launch across messaging, review, and delivery.' }).click();
        await page.locator('textarea').nth(1).fill('Generate a launch brief, pricing checklist, and asset package for tomorrow morning review.');
        await page.getByRole('button', { name: 'Broadcast to group' }).click();

        await expect(page.getByTestId('groups-notice')).toContainText('Broadcast queued for the selected group.');
        await expect.poll(() => harness.readBroadcastBodies().length).toBe(1);
        await expect.poll(() => String(harness.readBroadcastBodies()[0]?.message ?? '')).toContain('pricing checklist');
    });
});
