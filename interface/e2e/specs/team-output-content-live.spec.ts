import { expect, test, type Page } from '@playwright/test';

type APIEnvelope<T> = {
    ok?: boolean;
    data?: T;
    error?: string;
};

type ChatEnvelope = {
    data?: {
        mode?: string;
        payload?: {
            tools_used?: string[];
        };
    };
};

type ConfirmEnvelope = {
    data?: {
        run_id?: string;
        verified?: boolean;
        execution_state?: string;
        execution_summary?: {
            outputs?: Array<{ kind?: string; title?: string; id?: string; href?: string; retained?: boolean }>;
        };
    };
};

type OrganizationEnvelope = {
    data?: { id?: string };
};

type GroupRecord = {
    group_id: string;
    name: string;
    team_ids?: string[];
};

type ConversationEnvelope = {
    data?: {
        turns?: Array<{ tool_name?: string }>;
    };
};

type TeamOutputRef = { storage_ref?: string; team_id?: string; run_id?: string; kind?: string };

type TeamWorkItem = { work_item_id?: string; team_id?: string; run_id?: string; state?: string; execution_shape?: string; output_refs?: TeamOutputRef[] };

type TeamStatusEvent = { state?: string; source_kind?: string; source_channel?: string; payload_kind?: string; run_id?: string };

type RevealEnvelope = {
    ok?: boolean;
    data?: {
        workspace_path?: string;
        folder_path?: string;
    };
};

type TeamOutputAsk = {
    teamID: string;
    filePath: string;
    content: string;
    visibleText: string[];
    playGame?: boolean;
};

async function parseJSONIfPossible<T>(response: { text(): Promise<string> }) {
    const raw = await response.text();
    try {
        return { raw, body: JSON.parse(raw) as T };
    } catch {
        return { raw, body: null as T | null };
    }
}

async function requestWithTransientRetry<T>(request: () => Promise<T>): Promise<T> {
    let lastError = '';
    for (let attempt = 1; attempt <= 3; attempt += 1) {
        try {
            return await request();
        } catch (error) {
            lastError = error instanceof Error ? error.message : String(error);
            if (!['EADDRINUSE', 'ECONNRESET', 'ECONNREFUSED', 'ERR_CONNECTION_FAILED'].some((fragment) => lastError.includes(fragment)) || attempt === 3) break;
            await new Promise((resolve) => setTimeout(resolve, 1500));
        }
    }
    throw new Error(lastError || 'request failed');
}

function escapeRegex(value: string) {
    return value.replace(/[.*+?^${}()|[\]\\]/g, '\\$&');
}

function expectedRevealWorkspacePath(path: string) {
    return path.replace(/^workspace[\\/]/i, '').replace(/\\/g, '/');
}

async function createOrganization(page: Page, name: string) {
    const response = await page.request.post('/api/v1/organizations', {
        data: { name, purpose: 'Live team output content proof', start_mode: 'empty' },
    });
    const parsed = await parseJSONIfPossible<OrganizationEnvelope>(response);
    expect(response.ok(), parsed.body ? JSON.stringify(parsed.body) : parsed.raw).toBeTruthy();
    expect(parsed.body?.data?.id).toBeTruthy();
    return parsed.body!.data!.id!;
}

async function openWorkspace(page: Page, organizationId: string) {
    await page.goto(`/organizations/${organizationId}`, { waitUntil: 'domcontentloaded' });
    await page.getByPlaceholder(/Tell Soma what you want to plan, review, create, or (?:execute|run)/i).waitFor({ timeout: 30_000 });
}

async function submitWorkspaceChat(page: Page, content: string) {
    const input = page.getByPlaceholder(/Tell Soma what you want to plan, review, create, or (?:execute|run)/i);
    await input.fill(content);
    const responsePromise = page.waitForResponse(
        (response) => response.url().includes('/api/v1/chat') && response.request().method() === 'POST',
        { timeout: 120_000 },
    );
    await input.press('Enter');
    const response = await responsePromise;
    const parsed = await parseJSONIfPossible<ChatEnvelope>(response);
    return { response, raw: parsed.raw, body: parsed.body };
}

async function confirmProposal(page: Page) {
    const responsePromise = page.waitForResponse(
        (response) => response.url().includes('/api/v1/intent/confirm-action') && response.request().method() === 'POST',
        { timeout: 120_000 },
    );
    await page.getByRole('button', { name: /Approve & Execute|Execute|Run/i }).last().click();
    const response = await responsePromise;
    const parsed = await parseJSONIfPossible<ConfirmEnvelope>(response);
    return { response, raw: parsed.raw, body: parsed.body };
}

async function executeTeamOutput(page: Page, ask: TeamOutputAsk) {
    const outputHref = `/api/v1/workspace/files/view?path=${encodeURIComponent(ask.filePath)}`;
    const prompt = `Create a compact team with team_id ${ask.teamID}. Then have that team create a reviewable browser output at path ${ask.filePath} containing '${ask.content}'`;
    const proposal = await submitWorkspaceChat(page, prompt);

    expect(proposal.response.ok(), proposal.body ? JSON.stringify(proposal.body) : proposal.raw).toBeTruthy();
    expect(proposal.body?.data?.mode).toBe('proposal');
    expect(proposal.body?.data?.payload?.tools_used).toEqual(['create_team', 'write_file']);
    await expect(page.getByText('RUN CONFIRMATION').last()).toBeVisible({ timeout: 30_000 });
    await expect(page.getByText(ask.teamID).last()).toBeVisible();
    await expect(page.getByText(ask.filePath).last()).toBeVisible();

    const confirmed = await confirmProposal(page);
    expect(confirmed.response.ok(), confirmed.body ? JSON.stringify(confirmed.body) : confirmed.raw).toBeTruthy();
    expect(confirmed.body?.data?.verified).toBeTruthy();
    expect(confirmed.body?.data?.execution_state).toBe('verified');
    expect(confirmed.body?.data?.run_id).toBeTruthy();
    const outputs = confirmed.body?.data?.execution_summary?.outputs ?? [];
    expect(outputs.some((output) => output.kind === 'team' && output.id === ask.teamID && output.retained)).toBeTruthy();
    expect(outputs.some((output) => output.id === ask.filePath && output.href === outputHref && output.retained)).toBeTruthy();
    const runID = confirmed.body!.data!.run_id!;
    await expectConfirmedTeamWorkReadback(page, ask, runID);

    await expect(page.getByText(/Action completed|Result saved|The produced output is available for review/i).last()).toBeVisible({ timeout: 30_000 });
    await expect(page.getByText('Latest output').last()).toBeVisible({ timeout: 30_000 });
    await expect(page.getByText(ask.filePath).last()).toBeVisible();
    const openButton = page.getByRole('button', { name: /Open(?: file)? .+ in a new browser window/i }).last();
    const outputPagePromise = page.context().waitForEvent('page');
    await openButton.click();
    const outputPage = await outputPagePromise;
    await outputPage.waitForLoadState('domcontentloaded');
    for (const text of ask.visibleText) {
        await expect(outputPage.locator('body')).toContainText(text);
    }
    if (ask.playGame) {
        const score = outputPage.locator('#score');
        await expect(score).toBeVisible({ timeout: 30_000 });
        const before = Number(await score.textContent());
        await outputPage.locator('#game').click();
        await expect.poll(async () => Number(await score.textContent()), {
            timeout: 10_000,
            message: 'expected playable game score to increase after interaction',
        }).toBeGreaterThan(before);
    }
    await outputPage.close();
    const revealResponsePromise = page.waitForResponse((response) => {
        return response.url().includes('/api/v1/workspace/files/reveal') && response.request().method() === 'POST';
    }, { timeout: 30_000 });
    await page.getByRole('button', { name: new RegExp(`Open local folder for .*${escapeRegex(ask.filePath)}`, 'i') }).last().click();
    const revealResponse = await revealResponsePromise;
    expect(revealResponse.ok()).toBeTruthy();
    expect(revealResponse.url()).toContain(encodeURIComponent(ask.filePath));
    const parsedReveal = await parseJSONIfPossible<RevealEnvelope>(revealResponse);
    expect(parsedReveal.body?.ok, parsedReveal.raw).toBeTruthy();
    expect(parsedReveal.body?.data?.workspace_path).toBe(expectedRevealWorkspacePath(ask.filePath));
    expect(parsedReveal.body?.data?.folder_path).toBeTruthy();

    const conversation = await page.request.get(`/api/v1/runs/${runID}/conversation`);
    const parsedConversation = await parseJSONIfPossible<ConversationEnvelope>(conversation);
    expect(conversation.ok(), parsedConversation.body ? JSON.stringify(parsedConversation.body) : parsedConversation.raw).toBeTruthy();
    const toolNames = (parsedConversation.body?.data?.turns ?? []).map((turn) => turn.tool_name).filter(Boolean);
    expect(toolNames).toContain('create_team');
    expect(toolNames).toContain('write_file');
}

async function expectConfirmedTeamWorkReadback(page: Page, ask: TeamOutputAsk, runID: string) {
    const response = await requestWithTransientRetry(() => page.request.get(`/api/v1/teams/${encodeURIComponent(ask.teamID)}/work?limit=25`));
    const parsed = await parseJSONIfPossible<APIEnvelope<TeamWorkItem[]>>(response);
    expect(response.ok(), parsed.body ? JSON.stringify(parsed.body) : parsed.raw).toBeTruthy();
    const items = parsed.body?.data ?? [];
    const item = items.find((candidate) => (
        candidate.run_id === runID
        && (candidate.output_refs ?? []).some((output) => output.storage_ref === ask.filePath)
    ));
    expect(item, JSON.stringify(items)).toBeTruthy();
    expect(item?.team_id).toBe(ask.teamID);
    expect(item?.execution_shape).toBe('deliverable');
    expect(item?.state).toBe('output_ready');
    const outputRef = item?.output_refs?.find((output) => output.storage_ref === ask.filePath);
    expect(outputRef, JSON.stringify(item?.output_refs ?? [])).toBeTruthy();
    expect(outputRef?.storage_ref?.startsWith('/api/v1/workspace/files/view')).toBe(false);
    expect(outputRef?.team_id).toBe(ask.teamID);
    expect(outputRef?.run_id).toBe(runID);
    expect(outputRef?.kind).toBeTruthy();
    expect(item?.work_item_id).toBeTruthy();

    const eventsResponse = await page.request.get(`/api/v1/teams/${encodeURIComponent(ask.teamID)}/work/${encodeURIComponent(item!.work_item_id!)}/status-events?limit=25`);
    const parsedEvents = await parseJSONIfPossible<APIEnvelope<TeamStatusEvent[]>>(eventsResponse);
    expect(eventsResponse.ok(), parsedEvents.body ? JSON.stringify(parsedEvents.body) : parsedEvents.raw).toBeTruthy();
    const events = parsedEvents.body?.data ?? [];
    const states = events.map((event) => event.state);
    expect(states).toContain('queued');
    expect(states).toContain('running');
    expect(states).toContain('output_ready');
    for (const event of events.filter((candidate) => candidate.run_id === runID)) {
        expect(event.source_kind).toBe('web_api');
        expect(event.source_channel).toBe('api.intent.confirm-action');
        expect(event.payload_kind).toBe('status');
    }
}

async function expectTeamVisibleInGroups(page: Page, ask: TeamOutputAsk) {
    const response = await requestWithTransientRetry(() => page.request.get('/api/v1/groups'));
    const parsed = await parseJSONIfPossible<APIEnvelope<GroupRecord[]>>(response);
    expect(response.ok(), parsed.body ? JSON.stringify(parsed.body) : parsed.raw).toBeTruthy();
    const group = (parsed.body?.data ?? []).find((candidate) => candidate.team_ids?.includes(ask.teamID));
    expect(group, JSON.stringify(parsed.body?.data ?? [])).toBeTruthy();
    await page.goto(`/groups?group_id=${encodeURIComponent(group!.group_id)}`, { waitUntil: 'domcontentloaded' });
    await expect(page.getByRole('heading', { name: /Manage focused collaboration lanes/i })).toBeVisible({ timeout: 30_000 });
    await expect(page.getByRole('heading', { name: group!.name })).toBeVisible({ timeout: 30_000 });
    await expect(page.getByText(ask.teamID).first()).toBeVisible();

    await page.getByRole('tab', { name: /Workflow Log/i }).click();
    const workflowLog = page.getByTestId('groups-workflow-log');
    await expect(workflowLog).toBeVisible({ timeout: 30_000 });
    await expect(workflowLog).toContainText(ask.teamID);
    await expect(workflowLog).toContainText(ask.filePath);
}

async function expectTeamOutputVisibleOnDashboard(page: Page, ask: TeamOutputAsk) {
    await page.goto(`/dashboard?team_id=${encodeURIComponent(ask.teamID)}`, { waitUntil: 'domcontentloaded' });
    await expect(page.getByTestId('soma-operating-surface')).toBeVisible({ timeout: 30_000 });
    await expect(page.getByTestId('focused-team-output-dock')).toHaveCount(0);
    await expect(page.getByTestId('soma-team-context-switcher')).toContainText('Working in');
    const digest = page.getByTestId('soma-workbench-output-digest');
    await expect(digest).toBeVisible({ timeout: 30_000 });
    await expect(digest.getByText(ask.filePath).first()).toBeVisible();
    await expect(digest.getByRole('button', { name: /Open file/i })).toBeVisible();
    await expect(digest.getByRole('button', { name: /Open local folder/i })).toBeVisible();
}

async function expectTeamOutputVisibleInResources(page: Page, ask: TeamOutputAsk) {
    const normalizedPath = ask.filePath.replace(/\\/g, '/');
    const folderPath = normalizedPath.replace(/\/[^/]+$/, '');
    const fileName = normalizedPath.split('/').pop()!;

    await page.goto(`/resources?tab=workspace&path=${encodeURIComponent(folderPath)}`, { waitUntil: 'domcontentloaded' });
    await expect(page.getByRole('heading', { name: 'Resources' })).toBeVisible({ timeout: 30_000 });
    await expect(page.getByText(folderPath).last()).toBeVisible({ timeout: 30_000 });
    await expect(page.getByText(fileName).first()).toBeVisible({ timeout: 30_000 });
}

test.describe('Live teams produce reviewable content outputs', () => {
    test.skip(!process.env.PLAYWRIGHT_LIVE_BACKEND, 'requires a live Core backend');
    test.setTimeout(180_000);

    test('creates a team through Soma and verifies its playable generated output in the GUI', async ({ page }) => {
        test.slow();
        const stamp = Date.now();
        const organizationId = await createOrganization(page, `QA Team Output Content ${stamp}`);
        const asks: TeamOutputAsk[] = [
            {
                teamID: `qa-game-studio-${stamp}`,
                filePath: `workspace/logs/qa_orbit_dash_${stamp}.html`,
                content: '<!doctype html><title>Dot Dodge</title><style>body{background:#111;color:white;font-family:sans-serif}#game{width:520px;height:260px;background:linear-gradient(#123,#012);position:relative}#p{position:absolute;left:70px;top:110px;width:34px;height:34px;border-radius:50%;background:#fc3}</style><h1>Dot Dodge</h1><p>Score <span id=score>0</span></p><button id=jump>Jump</button><div id=game><div id=p></div></div><script>game.onclick=jump.onclick=()=>score.textContent=+score.textContent+1</script>',
                visibleText: ['Dot Dodge', 'Score', 'Jump'],
                playGame: true,
            },
        ];

        await openWorkspace(page, organizationId);
        for (const ask of asks) {
            await executeTeamOutput(page, ask);
        }
        for (const ask of asks) {
            await expectTeamOutputVisibleOnDashboard(page, ask);
        }
        for (const ask of asks) {
            await expectTeamVisibleInGroups(page, ask);
        }
        for (const ask of asks) {
            await expectTeamOutputVisibleInResources(page, ask);
        }
    });
});
