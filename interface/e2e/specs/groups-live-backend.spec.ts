import { expect, test, type Page } from '@playwright/test';
import { storeLiveArtifactWithRetry } from '../support/live-artifacts';
import { createLiveMissionTeam } from '../support/live-teams';

type APIEnvelope<T> = {
    ok?: boolean;
    data?: T;
    error?: string;
};

type GroupRecord = {
    group_id: string;
    name: string;
    goal_statement: string;
    status: string;
};

type ArtifactRecord = {
    id: string;
    title: string;
};

async function parseJSONIfPossible<T>(response: { text(): Promise<string> }) {
    const raw = await response.text();
    try {
        return {
            raw,
            body: JSON.parse(raw) as T,
        };
    } catch {
        return {
            raw,
            body: null as T | null,
        };
    }
}

async function createLiveGroup(page: Page, teamIDs: string[]) {
    const stamp = Date.now();
    const response = await page.request.post('/api/v1/groups', {
        data: {
            name: `Live Retained Output Sprint ${stamp}`,
            goal_statement: 'Prove live retained output review after temporary workflow closure.',
            work_mode: 'propose_only',
            allowed_capabilities: ['artifact.review'],
            member_user_ids: ['owner'],
            team_ids: teamIDs,
            coordinator_profile: 'Live retained-output lead',
            approval_policy_ref: 'browser-live-proof',
            expiry: new Date(Date.now() + 72 * 60 * 60 * 1000).toISOString(),
        },
    });
    const parsed = await parseJSONIfPossible<APIEnvelope<GroupRecord>>(response);
    expect(response.status(), parsed.body ? JSON.stringify(parsed.body) : parsed.raw).toBe(201);
    expect(response.ok(), parsed.body ? JSON.stringify(parsed.body) : parsed.raw).toBeTruthy();
    expect(parsed.body?.ok, parsed.body ? JSON.stringify(parsed.body) : parsed.raw).toBe(true);
    expect(parsed.body?.data?.group_id).toBeTruthy();
    return parsed.body?.data as GroupRecord;
}

async function listLiveGroups(page: Page) {
    const response = await page.request.get('/api/v1/groups');
    const parsed = await parseJSONIfPossible<APIEnvelope<GroupRecord[]>>(response);
    expect(response.ok(), parsed.body ? JSON.stringify(parsed.body) : parsed.raw).toBeTruthy();
    expect(Array.isArray(parsed.body?.data), parsed.body ? JSON.stringify(parsed.body) : parsed.raw).toBeTruthy();
    return parsed.body!.data!;
}

async function storeLiveArtifact(page: Page, teamID: string, title: string, agentID: string, content: string) {
    return await storeLiveArtifactWithRetry<ArtifactRecord>(page, {
        team_id: teamID,
        agent_id: agentID,
        artifact_type: 'document',
        title,
        content_type: 'text/markdown',
        content,
        metadata: {
            source: 'groups-live-backend.spec.ts',
        },
        status: 'approved',
    }, 'groups live backend');
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

async function openGroupsWorkspace(page: Page, path: string) {
    await gotoWithColdStartRetry(page, path);
    await expect(page.getByRole('heading', { name: /Manage focused collaboration lanes/i })).toBeVisible();
}

test.describe('Groups retained outputs live backend contract', () => {
    test.skip(!process.env.PLAYWRIGHT_LIVE_BACKEND, 'requires a live Core backend');
    test.setTimeout(120_000);

    test('archives a real temporary group while retaining backend-stored outputs', async ({ page }) => {
        test.slow();
        const liveTeam = await createLiveMissionTeam(page);
        const group = await createLiveGroup(page, [liveTeam.teamID]);
        await expect
            .poll(async () => (await listLiveGroups(page)).some((candidate) => candidate.group_id === group.group_id), {
                message: `created group ${group.group_id} should be visible through the Playwright baseURL proxy GET /api/v1/groups`,
            })
            .toBeTruthy();
        const brief = await storeLiveArtifact(
            page,
            liveTeam.teamID,
            `Live Launch Brief ${Date.now()}`,
            'launch-lead',
            '# Live launch brief\n\n- Message pillars\n- Delivery plan',
        );
        const checklist = await storeLiveArtifact(
            page,
            liveTeam.teamID,
            `Live Delivery Checklist ${Date.now()}`,
            'delivery-lead',
            '# Live delivery checklist\n\n- Owner review\n- Release criteria',
        );

        await openGroupsWorkspace(page, `/groups?group_id=${encodeURIComponent(group.group_id)}`);

        await expect(page.getByRole('heading', { name: 'Manage focused collaboration lanes.' })).toBeVisible();
        await expect(page.getByRole('link', { name: 'Open Soma' })).toHaveCount(2);
        await expect(page.getByTestId(`groups-list-item-${group.group_id}`)).toBeVisible();
        await page.getByTestId(`groups-list-item-${group.group_id}`).click();
        await expect(page.getByRole('heading', { name: group.name })).toBeVisible();
        await expect(page.getByText('Temporary group', { exact: true })).toBeVisible();
        await expect(page.getByTestId('groups-output-summary')).toContainText('2 outputs');
        await expect(page.getByTestId('groups-output-summary')).toContainText('2 contributing leads');
        await page.getByRole('tab', { name: /Outputs/i }).click();
        await expect(page.getByText(brief.title, { exact: true })).toBeVisible();
        await expect(page.getByText(checklist.title, { exact: true })).toBeVisible();
        await expect(page.getByRole('link', { name: 'Download' }).first()).toHaveAttribute('href', new RegExp('/api/v1/artifacts/.+/download'));

        await page.getByRole('tab', { name: /Overview/i }).click();
        await page.getByRole('button', { name: 'Clear group from active lanes' }).click();

        await expect(page.getByTestId('groups-notice')).toContainText('Group cleared from active lanes');
        await expect(page.getByText('Archived temporary group', { exact: true })).toBeVisible();
        await page.getByRole('tab', { name: /Message/i }).click();
        await expect(page.getByTestId('groups-archived-readonly-note')).toContainText('retained output review');
        await page.getByRole('tab', { name: /Outputs/i }).click();
        await expect(page.getByTestId('groups-retained-outputs-note')).toContainText('Downloads remain available');
        await expect(page.getByTestId('groups-output-summary')).toContainText('2 outputs');
        await expect(page.getByTestId('groups-output-summary')).toContainText('2 contributing leads');
        await expect(page.getByText(brief.title, { exact: true })).toBeVisible();
        await expect(page.getByText(checklist.title, { exact: true })).toBeVisible();
        await expect(page.getByRole('button', { name: 'Broadcast to group' })).toHaveCount(0);
    });
});
