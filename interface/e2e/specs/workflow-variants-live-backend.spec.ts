import { expect, test, type Page } from '@playwright/test';
import {
    type APIEnvelope,
    type ArtifactRecord,
    createOrganization,
    expectExecutionContractOutputs,
    gotoWithColdStartRetry,
    type GroupRecord,
    type MissionDetail,
    openTeamCreation,
    openWorkspace,
    parseJSONIfPossible,
    type SeedResponse,
    submitTeamDesign,
    submitWorkspaceChat,
    type TeamLeadWorkflowGroupDraft,
} from '../support/workflow-variants-helpers';

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

async function createLiveTeamIDs(page: Page, count: number) {
    const ids: string[] = [];
    for (let index = 0; index < count; index += 1) {
        const liveTeam = await createLiveMissionTeam(page);
        ids.push(liveTeam.teamID);
    }
    return ids;
}

async function createLiveGroup(page: Page, draft: TeamLeadWorkflowGroupDraft, teamIDs: string[]) {
    const uniqueSuffix = `${Date.now()}-${Math.floor(Math.random() * 10_000)}`;
    const expiry = typeof draft.expiry_hours === 'number' && draft.expiry_hours > 0
        ? new Date(Date.now() + draft.expiry_hours * 60 * 60 * 1000).toISOString()
        : new Date(Date.now() + 72 * 60 * 60 * 1000).toISOString();
    const response = await page.request.post('/api/v1/groups', {
        data: {
            name: `${draft.name} ${uniqueSuffix}`,
            goal_statement: draft.goal_statement,
            work_mode: draft.work_mode,
            allowed_capabilities: draft.allowed_capabilities ?? [],
            member_user_ids: ['owner'],
            team_ids: teamIDs,
            coordinator_profile: draft.coordinator_profile,
            approval_policy_ref: 'browser-live-proof',
            expiry,
        },
    });
    const parsed = await parseJSONIfPossible<APIEnvelope<GroupRecord>>(response);
    expect(response.status(), parsed.body ? JSON.stringify(parsed.body) : parsed.raw).toBe(201);
    expect(response.ok(), parsed.body ? JSON.stringify(parsed.body) : parsed.raw).toBeTruthy();
    expect(parsed.body?.ok, parsed.body ? JSON.stringify(parsed.body) : parsed.raw).toBe(true);
    expect(parsed.body?.data?.group_id).toBeTruthy();
    return parsed.body!.data!;
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
                source: 'workflow-variants-live-backend.spec.ts',
            },
            status: 'approved',
        },
    });
    const parsed = await parseJSONIfPossible<ArtifactRecord>(response);
    expect(response.ok(), parsed.body ? JSON.stringify(parsed.body) : parsed.raw).toBeTruthy();
    expect(parsed.body?.id).toBeTruthy();
    return parsed.body!;
}

async function waitForBrowserGroupList(page: Page, expectedGroupID: string) {
    const response = await page.waitForResponse((candidate) => {
        const url = new URL(candidate.url());
        return candidate.request().method() === 'GET' && url.pathname === '/api/v1/groups';
    }, { timeout: 30_000 });
    const parsed = await parseJSONIfPossible<APIEnvelope<GroupRecord[]>>(response);
    expect(response.ok(), parsed.body ? JSON.stringify(parsed.body) : parsed.raw).toBeTruthy();
    const groups = parsed.body?.data ?? [];
    expect(groups.some((candidate) => candidate.group_id === expectedGroupID)).toBeTruthy();
}

async function reviewArchivedOutputs(page: Page, group: GroupRecord, artifactTitles: string[], contributingLeads: number) {
    const groupsResponse = waitForBrowserGroupList(page, group.group_id);
    await gotoWithColdStartRetry(page, `/groups?group_id=${encodeURIComponent(group.group_id)}`);
    await groupsResponse;

    await expect(page.getByRole('heading', { name: 'Manage focused collaboration lanes.' })).toBeVisible();
    await expect(page.getByRole('heading', { name: group.name })).toBeVisible();
    await expect(page.getByText('Temporary group', { exact: true })).toBeVisible();
    await expect(page.getByTestId('groups-output-summary')).toContainText(`${artifactTitles.length} outputs`);
    await expect(page.getByTestId('groups-output-summary')).toContainText(`${contributingLeads} contributing leads`);
    for (const artifactTitle of artifactTitles) {
        await expect(page.getByText(artifactTitle, { exact: true })).toBeVisible();
    }

    await page.getByRole('button', { name: 'Archive temporary group' }).click();

    await expect(page.getByTestId('groups-notice')).toContainText('Temporary group archived.');
    await expect(page.getByText('Archived temporary group', { exact: true })).toBeVisible();
    await expect(page.getByTestId('groups-archived-readonly-note')).toContainText('retained output review');
    await expect(page.getByTestId('groups-output-summary')).toContainText(`${artifactTitles.length} outputs`);
    await expect(page.getByTestId('groups-output-summary')).toContainText(`${contributingLeads} contributing leads`);

    await page.reload({ waitUntil: 'domcontentloaded' });
    await expect(page.getByText('Archived temporary group', { exact: true })).toBeVisible();
    for (const artifactTitle of artifactTitles) {
        await expect(page.getByText(artifactTitle, { exact: true })).toBeVisible();
    }
}

test.describe('Workflow variants live backend contract', () => {
    test.skip(!process.env.PLAYWRIGHT_LIVE_BACKEND, 'requires a live Core backend');
    test.setTimeout(180_000);

    test('proves live direct answer plus compact and multi-team retained-output review', async ({ page }) => {
        test.slow();
        const stamp = Date.now();
        const organization = await createOrganization(page, `QA Workflow Variants ${stamp}`);

        await openWorkspace(page, organization.id);

        const direct = await submitWorkspaceChat(page, 'Summarize the current Workspace V8 design objectives.');
        expect(direct.response.ok(), direct.body ? JSON.stringify(direct.body) : direct.raw).toBeTruthy();
        expect(direct.body?.data?.mode).toBe('answer');
        expect(direct.body?.data?.payload?.ask_class).toBe('direct_answer');
        expect((direct.body?.data?.payload?.text ?? '').trim().length).toBeGreaterThan(0);
        await expect(page.getByText(/could not produce a readable reply/i)).toHaveCount(0);

        await openTeamCreation(page, organization.id);
        await expect(page.getByText('Current organization')).toBeVisible();
        await expect(page.getByText(organization.name, { exact: true }).last()).toBeVisible();

        const compactPrompt = 'Create a temporary marketing launch team for a new product rollout.';
        const compactGuidance = await submitTeamDesign(page, organization.id, compactPrompt);
        const compactContract = compactGuidance.execution_contract!;
        expect(compactContract.execution_mode).toBe('native_team');
        expect(compactContract.coordination_model).toBe('compact_team');
        expect(compactContract.team_name).toBeTruthy();
        expect(compactContract.target_outputs?.length).toBe(3);
        expect(compactContract.workflow_group?.work_mode).toBe('propose_only');
        expect((compactGuidance.headline ?? '').trim().length).toBeGreaterThan(0);
        await expect(page.getByText(compactContract.team_name!, { exact: true })).toBeVisible();
        await expectExecutionContractOutputs(page, compactContract.target_outputs ?? []);

        const compactTeamIDs = await createLiveTeamIDs(page, compactContract.target_outputs?.length ?? 3);
        const compactGroup = await createLiveGroup(page, compactContract.workflow_group!, compactTeamIDs);
        for (const [index, output] of (compactContract.target_outputs ?? []).entries()) {
            await storeLiveArtifact(
                page,
                compactTeamIDs[index]!,
                output,
                `compact-lead-${index + 1}`,
                `# ${output}\n\n- Compact workflow proof\n- Live retained output`,
            );
        }
        await reviewArchivedOutputs(page, compactGroup, compactContract.target_outputs ?? [], compactTeamIDs.length);

        await openTeamCreation(page, organization.id);
        const multiPrompt = 'Create a company-wide product launch program across marketing, sales, support, docs, and engineering so the organization can coordinate several workstreams at once.';
        const multiGuidance = await submitTeamDesign(page, organization.id, multiPrompt);
        const multiContract = multiGuidance.execution_contract!;
        expect(multiContract.execution_mode).toBe('native_team');
        expect(multiContract.coordination_model).toBe('multi_team_orchestration');
        expect(multiContract.recommended_team_count).toBe(3);
        expect(multiContract.recommended_team_member_limit).toBe(5);
        expect(multiContract.target_outputs?.length).toBeGreaterThan(0);
        expect(multiContract.workflow_group?.work_mode).toBe('propose_only');
        expect((multiGuidance.headline ?? '').trim().length).toBeGreaterThan(0);
        if (multiContract.team_name) {
            await expect(page.getByText(multiContract.team_name, { exact: true })).toBeVisible();
        }
        await expectExecutionContractOutputs(page, multiContract.target_outputs ?? []);

        const multiOutputs = (multiContract.target_outputs ?? []).slice(0, 3);
        const multiTeamIDs = await createLiveTeamIDs(page, multiOutputs.length || 3);
        const multiGroup = await createLiveGroup(page, multiContract.workflow_group!, multiTeamIDs);
        for (const [index, output] of multiOutputs.entries()) {
            await storeLiveArtifact(
                page,
                multiTeamIDs[index]!,
                output,
                `lane-lead-${index + 1}`,
                `# ${output}\n\n- Multi-team workflow proof\n- Live retained output`,
            );
        }
        await reviewArchivedOutputs(page, multiGroup, multiOutputs, multiTeamIDs.length);
    });
});
