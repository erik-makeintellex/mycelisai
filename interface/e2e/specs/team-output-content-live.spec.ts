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
    await page.getByPlaceholder(/Tell Soma what you want to plan, review, create, or execute/i).waitFor({ timeout: 30_000 });
}

async function submitWorkspaceChat(page: Page, content: string) {
    const input = page.getByPlaceholder(/Tell Soma what you want to plan, review, create, or execute/i);
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
    await page.getByRole('button', { name: /Approve & Execute|Execute/i }).last().click();
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
    await expect(page.getByText('PROPOSED ACTION').last()).toBeVisible({ timeout: 30_000 });
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

    await expect(page.getByText(/Execution verified/i).last()).toBeVisible({ timeout: 30_000 });
    const outputLink = page.getByRole('link', { name: ask.filePath }).last();
    await expect(outputLink).toHaveAttribute('href', outputHref);
    const openButton = page.getByRole('button', { name: `Open ${ask.filePath} in a new browser window` }).last();
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
    await page.getByRole('button', { name: `Open local folder for ${ask.filePath}` }).last().click();
    const revealResponse = await revealResponsePromise;
    expect(revealResponse.ok()).toBeTruthy();

    const conversation = await page.request.get(`/api/v1/runs/${confirmed.body!.data!.run_id}/conversation`);
    const parsedConversation = await parseJSONIfPossible<ConversationEnvelope>(conversation);
    expect(conversation.ok(), parsedConversation.body ? JSON.stringify(parsedConversation.body) : parsedConversation.raw).toBeTruthy();
    const toolNames = (parsedConversation.body?.data?.turns ?? []).map((turn) => turn.tool_name).filter(Boolean);
    expect(toolNames).toContain('create_team');
    expect(toolNames).toContain('write_file');
}

async function expectTeamVisibleInGroups(page: Page, teamID: string) {
    const response = await requestWithTransientRetry(() => page.request.get('/api/v1/groups'));
    const parsed = await parseJSONIfPossible<APIEnvelope<GroupRecord[]>>(response);
    expect(response.ok(), parsed.body ? JSON.stringify(parsed.body) : parsed.raw).toBeTruthy();
    const group = (parsed.body?.data ?? []).find((candidate) => candidate.team_ids?.includes(teamID));
    expect(group, JSON.stringify(parsed.body?.data ?? [])).toBeTruthy();
    await page.goto(`/groups?group_id=${encodeURIComponent(group!.group_id)}`, { waitUntil: 'domcontentloaded' });
    await expect(page.getByRole('heading', { name: group!.name })).toBeVisible({ timeout: 30_000 });
    await expect(page.getByText(teamID).first()).toBeVisible();
}

test.describe('Live teams produce reviewable content outputs', () => {
    test.skip(!process.env.PLAYWRIGHT_LIVE_BACKEND, 'requires a live Core backend');
    test.setTimeout(180_000);

    test('creates teams through Soma and verifies their generated outputs in the GUI', async ({ page }) => {
        test.slow();
        const stamp = Date.now();
        const organizationId = await createOrganization(page, `QA Team Output Content ${stamp}`);
        const asks: TeamOutputAsk[] = [
            {
                teamID: `qa-marketing-output-${stamp}`,
                filePath: `workspace/logs/qa_marketing_launch_brief_${stamp}.html`,
                content: '<!doctype html><title>Launch Brief</title><main><h1>Launch brief</h1><p>Audience: operator buyers.</p><p>Promise: visible execution and proof.</p><p>Deliverable: launch narrative, top objections, approval checklist.</p></main>',
                visibleText: ['Launch brief', 'operator buyers', 'visible execution and proof', 'approval checklist'],
            },
            {
                teamID: `qa-readiness-output-${stamp}`,
                filePath: `workspace/logs/qa_readiness_matrix_${stamp}.html`,
                content: '<!doctype html><title>Readiness Matrix</title><main><h1>Readiness matrix</h1><p>Teams: sales support docs engineering.</p><p>Runbook: intake, escalation, proof review.</p><p>Deliverable: operator handoff.</p></main>',
                visibleText: ['Readiness matrix', 'sales support docs engineering', 'proof review', 'operator handoff'],
            },
            {
                teamID: `qa-game-studio-${stamp}`,
                filePath: `workspace/logs/qa_orbit_dash_${stamp}.html`,
                content: '<!doctype html><title>Dot Dodge</title><style>body{margin:0;background:#111;color:white;font-family:sans-serif}#game{width:640px;height:360px;background:linear-gradient(#123,#012);position:relative;overflow:hidden}#p,#o{position:absolute;width:32px;height:32px;border-radius:50%}#p{left:60px;top:150px;background:#ffcc33}#o{left:520px;top:80px;background:#ff5577}button{font-size:20px}</style><h1>Dot Dodge</h1><p>Score <span id=score>0</span></p><div id=game><div id=p></div><div id=o></div></div><button id=jump>Jump</button><script>let s=0,y=150,v=0,p=document.getElementById(String.fromCharCode(112)),score=document.getElementById(String.fromCharCode(115,99,111,114,101));jump.onclick=()=>v=-8;game.onclick=()=>v=-8;setInterval(()=>{v+=1;y+=v;if(y>300){y=300;v=0}s++;p.style.top=y+String.fromCharCode(112,120);score.textContent=s},60)</script>',
                visibleText: ['Dot Dodge', 'Score', 'Jump'],
                playGame: true,
            },
        ];

        await openWorkspace(page, organizationId);
        for (const ask of asks) {
            await executeTeamOutput(page, ask);
        }
        for (const ask of asks) {
            await expectTeamVisibleInGroups(page, ask.teamID);
        }
    });
});
