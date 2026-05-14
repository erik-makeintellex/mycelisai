import fs from 'node:fs';
import path from 'node:path';
import { execFileSync } from 'node:child_process';
import { expect, test, type Page } from '@playwright/test';

const repoRoot = path.resolve(__dirname, '../../..');
const LIVE_TIMEOUT_MS = 180_000;
const CHAT_TIMEOUT_MS = 120_000;

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
            proposal?: {
                nats_subjects?: string[];
                bus_scope?: string;
            };
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
    work_mode?: string;
    status?: string;
    team_ids?: string[];
};

type ConversationEnvelope = {
    data?: {
        turns?: Array<{ role?: string; tool_name?: string; content?: string }>;
    };
};

async function parseJSONIfPossible<T>(response: { text(): Promise<string> }) {
    const raw = await response.text();
    try {
        return { raw, body: JSON.parse(raw) as T };
    } catch {
        return { raw, body: null as T | null };
    }
}

function backendWorkspaceRoot() {
    const configuredRoot = process.env.PLAYWRIGHT_BACKEND_WORKSPACE_ROOT ?? process.env.MYCELIS_BACKEND_WORKSPACE_ROOT;
    if (configuredRoot?.trim()) {
        return path.isAbsolute(configuredRoot) ? configuredRoot : path.join(repoRoot, configuredRoot);
    }
    return path.join(repoRoot, 'workspace', 'docker-compose', 'data', 'workspace');
}

function escapeRegExp(value: string) {
    return value.replace(/[.*+?^${}()|[\]\\]/g, '\\$&');
}

let cachedK8sPodName: string | null = null;

function k8sWorkspaceProbeEnabled() {
    return process.env.PLAYWRIGHT_BACKEND_WORKSPACE_PROBE === 'k8s';
}

function shellQuote(value: string) {
    return `'${value.replace(/'/g, `'\\''`)}'`;
}

function k8sCorePodName() {
    if (cachedK8sPodName) return cachedK8sPodName;
    const namespace = process.env.PLAYWRIGHT_K8S_NAMESPACE ?? 'mycelis';
    const selector = process.env.PLAYWRIGHT_K8S_CORE_SELECTOR ?? 'app=mycelis-core';
    cachedK8sPodName = execFileSync(
        'kubectl',
        ['get', 'pods', '-n', namespace, '-l', selector, '-o', 'jsonpath={.items[0].metadata.name}'],
        { encoding: 'utf8', stdio: ['ignore', 'pipe', 'ignore'], timeout: 10_000 },
    ).trim();
    return cachedK8sPodName;
}

function targetExists(relativePath: string) {
    const normalized = relativePath.replace(/^workspace[\\/]/, '').replace(/\\/g, '/');
    const localTarget = path.join(backendWorkspaceRoot(), normalized);
    if (fs.existsSync(localTarget)) return true;
    if (!k8sWorkspaceProbeEnabled()) return false;

    const namespace = process.env.PLAYWRIGHT_K8S_NAMESPACE ?? 'mycelis';
    const root = (process.env.PLAYWRIGHT_K8S_BACKEND_WORKSPACE_ROOT ?? '/data/workspace').replace(/\/$/, '');
    try {
        execFileSync(
            'kubectl',
            ['exec', '-n', namespace, k8sCorePodName(), '--', 'sh', '-c', `test -f ${shellQuote(`${root}/${normalized}`)}`],
            { stdio: 'ignore', timeout: 10_000 },
        );
        return true;
    } catch {
        return false;
    }
}

function removeTarget(relativePath: string) {
    const normalized = relativePath.replace(/^workspace[\\/]/, '').replace(/\\/g, '/');
    fs.rmSync(path.join(backendWorkspaceRoot(), normalized), { force: true });
    if (!k8sWorkspaceProbeEnabled()) return;

    const namespace = process.env.PLAYWRIGHT_K8S_NAMESPACE ?? 'mycelis';
    const root = (process.env.PLAYWRIGHT_K8S_BACKEND_WORKSPACE_ROOT ?? '/data/workspace').replace(/\/$/, '');
    try {
        execFileSync(
            'kubectl',
            ['exec', '-n', namespace, k8sCorePodName(), '--', 'sh', '-c', `rm -f ${shellQuote(`${root}/${normalized}`)}`],
            { stdio: 'ignore', timeout: 10_000 },
        );
    } catch {
        // Preserve the primary Playwright failure if cleanup cannot reach the cluster.
    }
}

async function createOrganization(page: Page, name: string) {
    const response = await page.request.post('/api/v1/organizations', {
        data: { name, purpose: 'Live team execution artifact proof', start_mode: 'empty' },
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
        { timeout: CHAT_TIMEOUT_MS },
    );
    await input.press('Enter');
    const response = await responsePromise;
    const parsed = await parseJSONIfPossible<ChatEnvelope>(response);
    return { response, raw: parsed.raw, body: parsed.body };
}

async function confirmProposal(page: Page) {
    const responsePromise = page.waitForResponse(
        (response) => response.url().includes('/api/v1/intent/confirm-action') && response.request().method() === 'POST',
        { timeout: CHAT_TIMEOUT_MS },
    );
    await page.getByRole('button', { name: /Approve & Execute|Execute/i }).click();
    const response = await responsePromise;
    const parsed = await parseJSONIfPossible<ConfirmEnvelope>(response);
    return { response, raw: parsed.raw, body: parsed.body };
}

test.describe('Live team execution produces retained code outputs', () => {
    test.skip(!process.env.PLAYWRIGHT_LIVE_BACKEND, 'requires a live Core backend');
    test.describe.configure({ timeout: LIVE_TIMEOUT_MS });

    test('creates a team, writes a small browser game, records run proof, and exposes the team in Groups', async ({ page }) => {
        test.slow();
        const stamp = Date.now();
        const teamID = `qa-browser-game-team-${stamp}`;
        const filePath = `workspace/logs/qa_team_click_game_${stamp}.html`;
        const outputHref = `/api/v1/workspace/files/view?path=${encodeURIComponent(filePath)}`;
        const gameCode = '<!doctype html><title>QA Click Game</title><button id=coin>Coin</button><p id=score>0</p><script>let s=0;coin.onclick=()=>score.textContent=++s</script>';
        const expectedOutput = 'Expected output: a playable retained HTML game link in chat, title QA Click Game, a Coin button, score starts at 0, and score becomes 2 after two clicks.';
        const organizationId = await createOrganization(page, `QA Team Execution ${stamp}`);

        await openWorkspace(page, organizationId);
        try {
            const proposal = await submitWorkspaceChat(
                page,
                `Create a compact team with team_id ${teamID}. Then have that team create a simple browser click game at path ${filePath} containing '${gameCode}'. ${expectedOutput}`,
            );

            expect(proposal.response.ok(), proposal.body ? JSON.stringify(proposal.body) : proposal.raw).toBeTruthy();
            expect(proposal.body?.data?.mode).toBe('proposal');
            expect(proposal.body?.data?.payload?.tools_used).toEqual(['create_team', 'write_file']);
            expect(proposal.body?.data?.payload?.proposal?.bus_scope).toBeTruthy();
            const natsSubjects = proposal.body?.data?.payload?.proposal?.nats_subjects ?? [];
            expect(natsSubjects).toEqual([
                `swarm.team.${teamID}.internal.command`,
                `swarm.team.${teamID}.signal.status`,
                `swarm.team.${teamID}.signal.result`,
            ]);
            expect(natsSubjects.some((subject) => subject.includes('admin-core'))).toBeFalsy();
            await expect(page.getByText('PROPOSED ACTION')).toBeVisible({ timeout: 30_000 });
            await expect(page.getByText(teamID).first()).toBeVisible();
            await expect(page.getByText(filePath).first()).toBeVisible();
            expect(targetExists(filePath)).toBeFalsy();

            const confirmed = await confirmProposal(page);

            expect(confirmed.response.ok(), confirmed.body ? JSON.stringify(confirmed.body) : confirmed.raw).toBeTruthy();
            expect(confirmed.body?.data?.verified).toBeTruthy();
            expect(confirmed.body?.data?.execution_state).toBe('verified');
            const runID = confirmed.body?.data?.run_id;
            expect(runID).toBeTruthy();
            const outputs = confirmed.body?.data?.execution_summary?.outputs ?? [];
            expect(outputs.some((output) => output.kind === 'team' && output.id === teamID && output.retained)).toBeTruthy();
            const codeOutput = outputs.find((output) => output.kind === 'code' && output.id === filePath);
            expect(codeOutput).toBeTruthy();
            expect(codeOutput?.retained).toBeTruthy();
            expect(codeOutput?.title).toBe(filePath);
            expect(codeOutput?.href).toBe(outputHref);

            await expect.poll(() => targetExists(filePath), {
                timeout: 30_000,
                message: `expected backend workspace file ${filePath} to exist after approval`,
            }).toBeTruthy();
            await expect(page.getByText(/Execution verified/i)).toBeVisible({ timeout: 30_000 });
            await expect(page.getByText(filePath).first()).toBeVisible({ timeout: 30_000 });
            const gameLink = page.getByRole('link', { name: new RegExp(escapeRegExp(filePath)) }).first();
            await expect(gameLink).toHaveAttribute('href', outputHref);
            await expect(gameLink).toHaveAttribute('target', '_blank');
            const gamePagePromise = page.context().waitForEvent('page');
            await gameLink.click();
            const gamePage = await gamePagePromise;
            await gamePage.waitForLoadState('domcontentloaded');
            await expect(gamePage).toHaveTitle('QA Click Game');
            await expect(gamePage.locator('#coin')).toBeVisible({ timeout: 30_000 });
            await expect(gamePage.locator('#score')).toHaveText('0');
            await gamePage.locator('#coin').click();
            await gamePage.locator('#coin').click();
            await expect(gamePage.locator('#score')).toHaveText('2');
            await gamePage.close();

            const conversation = await page.request.get(`/api/v1/runs/${runID}/conversation`);
            const parsedConversation = await parseJSONIfPossible<ConversationEnvelope>(conversation);
            expect(conversation.ok(), parsedConversation.body ? JSON.stringify(parsedConversation.body) : parsedConversation.raw).toBeTruthy();
            const toolNames = (parsedConversation.body?.data?.turns ?? []).map((turn) => turn.tool_name).filter(Boolean);
            expect(toolNames).toContain('create_team');
            expect(toolNames).toContain('write_file');

            const groupsResponse = await page.request.get('/api/v1/groups');
            const parsedGroups = await parseJSONIfPossible<APIEnvelope<GroupRecord[]>>(groupsResponse);
            expect(groupsResponse.ok(), parsedGroups.body ? JSON.stringify(parsedGroups.body) : parsedGroups.raw).toBeTruthy();
            const group = (parsedGroups.body?.data ?? []).find((candidate) => candidate.team_ids?.includes(teamID));
            expect(group, JSON.stringify(parsedGroups.body?.data ?? [])).toBeTruthy();
            expect(group!.work_mode).toBe('propose_only');
            expect(group!.status).toBe('active');

            await page.goto(`/groups?group_id=${encodeURIComponent(group!.group_id)}`, { waitUntil: 'domcontentloaded' });
            await expect(page.getByRole('heading', { name: group!.name })).toBeVisible({ timeout: 30_000 });
            await expect(page.getByText(teamID).first()).toBeVisible();
            await expect(page.getByText('propose_only').first()).toBeVisible();
        } finally {
            removeTarget(filePath);
        }
    });
});
