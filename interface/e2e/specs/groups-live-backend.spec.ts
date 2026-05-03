import { expect, test, type Page } from '@playwright/test';

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

type SeedResponse = {
    mission_id?: string;
};

type MissionDetail = {
    id?: string;
    teams?: Array<{
        id?: string;
        name?: string;
    }>;
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

function summarizeGroups(groups: GroupRecord[]) {
    return groups.map((group) => `${group.group_id}:${group.name}:${group.status}`).join(', ');
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

async function waitForBrowserGroupList(page: Page, expectedGroupID: string) {
    let response;
    try {
        response = await page.waitForResponse((candidate) => {
            const url = new URL(candidate.url());
            return candidate.request().method() === 'GET' && url.pathname === '/api/v1/groups';
        }, { timeout: 30_000 });
    } catch (error) {
        const message = error instanceof Error ? error.message : String(error);
        throw new Error(`Groups UI did not issue GET /api/v1/groups after navigation. Check Playwright baseURL, Next proxy rewrites, and built static JS hydration. ${message}`);
    }

    const parsed = await parseJSONIfPossible<APIEnvelope<GroupRecord[]>>(response);
    expect(response.ok(), parsed.body ? JSON.stringify(parsed.body) : parsed.raw).toBeTruthy();
    const groups = parsed.body?.data ?? [];
    expect(
        groups.some((candidate) => candidate.group_id === expectedGroupID),
        `Browser UI proxy GET /api/v1/groups did not include seeded group ${expectedGroupID}. Got: ${summarizeGroups(groups)}`,
    ).toBeTruthy();
}

async function createLiveMissionTeam(page: Page) {
    const seedResponse = await page.request.post('/api/v1/intent/seed/symbiotic');
    const parsedSeed = await parseJSONIfPossible<SeedResponse>(seedResponse);
    expect(seedResponse.ok(), parsedSeed.body ? JSON.stringify(parsedSeed.body) : parsedSeed.raw).toBeTruthy();
    expect(parsedSeed.body?.mission_id).toBeTruthy();

    const missionResponse = await page.request.get(`/api/v1/missions/${parsedSeed.body!.mission_id}`);
    const parsedMission = await parseJSONIfPossible<MissionDetail>(missionResponse);
    expect(missionResponse.ok(), parsedMission.body ? JSON.stringify(parsedMission.body) : parsedMission.raw).toBeTruthy();
    const team = parsedMission.body?.teams?.find((candidate) => candidate.id);
    expect(team?.id).toBeTruthy();
    return {
        missionID: parsedSeed.body!.mission_id!,
        teamID: team!.id!,
        teamName: team?.name ?? 'live mission team',
    };
}

async function storeLiveArtifact(page: Page, teamID: string, title: string, agentID: string, content: string) {
    const response = await page.request.post('/api/v1/artifacts', {
        data: {
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
        },
    });
    const parsed = await parseJSONIfPossible<ArtifactRecord>(response);
    expect(response.ok(), parsed.body ? JSON.stringify(parsed.body) : parsed.raw).toBeTruthy();
    expect(parsed.body?.id).toBeTruthy();
    return parsed.body as ArtifactRecord;
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

        const groupsResponse = waitForBrowserGroupList(page, group.group_id);
        await gotoWithColdStartRetry(page, `/groups?group_id=${encodeURIComponent(group.group_id)}`);
        await groupsResponse;

        await expect(page.getByRole('heading', { name: 'Manage focused collaboration lanes.' })).toBeVisible();
        await expect(page.getByRole('heading', { name: group.name })).toBeVisible();
        await expect(page.getByText('Temporary group', { exact: true })).toBeVisible();
        await expect(page.getByTestId('groups-output-summary')).toContainText('2 outputs');
        await expect(page.getByTestId('groups-output-summary')).toContainText('2 contributing leads');
        await expect(page.getByText(brief.title, { exact: true })).toBeVisible();
        await expect(page.getByText(checklist.title, { exact: true })).toBeVisible();
        await expect(page.getByRole('link', { name: 'Download' }).first()).toHaveAttribute('href', new RegExp('/api/v1/artifacts/.+/download'));

        await page.getByRole('button', { name: 'Archive temporary group' }).click();

        await expect(page.getByTestId('groups-notice')).toContainText('Temporary group archived.');
        await expect(page.getByText('Archived temporary group', { exact: true })).toBeVisible();
        await expect(page.getByTestId('groups-archived-readonly-note')).toContainText('retained output review');
        await expect(page.getByTestId('groups-retained-outputs-note')).toContainText('Downloads remain available');
        await expect(page.getByTestId('groups-output-summary')).toContainText('2 outputs');
        await expect(page.getByTestId('groups-output-summary')).toContainText('2 contributing leads');
        await expect(page.getByText(brief.title, { exact: true })).toBeVisible();
        await expect(page.getByText(checklist.title, { exact: true })).toBeVisible();
        await expect(page.getByRole('button', { name: 'Broadcast to group' })).toHaveCount(0);
    });
});
