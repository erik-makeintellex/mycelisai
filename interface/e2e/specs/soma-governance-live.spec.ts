import fs from 'node:fs';
import path from 'node:path';
import { execFileSync } from 'node:child_process';
import { expect, test, type Page } from '@playwright/test';
import { openOrganizationWorkspace, organizationChatInput, waitForOrganizationWorkspaceReady } from '../support/live-organization-workspace';

const repoRoot = path.resolve(__dirname, '../../..');
const LIVE_GOVERNANCE_TIMEOUT_MS = 180_000;
const LIVE_CHAT_RESPONSE_TIMEOUT_MS = 120_000;

function resolveBackendWorkspaceRoots() {
    const configuredRoot =
        process.env.PLAYWRIGHT_BACKEND_WORKSPACE_ROOT ?? process.env.MYCELIS_BACKEND_WORKSPACE_ROOT;
    if (configuredRoot && configuredRoot.trim().length > 0) {
        return [path.isAbsolute(configuredRoot) ? configuredRoot : path.join(repoRoot, configuredRoot)];
    }

    return [
        path.join(repoRoot, 'workspace', 'docker-compose', 'data', 'workspace'),
        path.join(repoRoot, 'core', 'workspace'),
    ];
}

function resolveBackendLogTargets(filename: string) {
    return resolveBackendWorkspaceRoots().map((workspaceRoot) => path.join(workspaceRoot, 'logs', filename));
}

let cachedK8sPodName: string | null = null;

function k8sWorkspaceProbeEnabled() {
    return process.env.PLAYWRIGHT_BACKEND_WORKSPACE_PROBE === 'k8s';
}

function shellQuote(value: string) {
    return `'${value.replace(/'/g, `'\\''`)}'`;
}

function k8sWorkspaceRoot() {
    return (
        process.env.PLAYWRIGHT_K8S_BACKEND_WORKSPACE_ROOT ??
        process.env.MYCELIS_BACKEND_WORKSPACE_ROOT ??
        '/data/workspace'
    ).replace(/\\/g, '/');
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

function k8sLogPathForTarget(candidate: string) {
    return `${k8sWorkspaceRoot().replace(/\/$/, '')}/logs/${path.basename(candidate)}`;
}

function k8sTargetExists(candidate: string) {
    if (!k8sWorkspaceProbeEnabled()) return false;
    const namespace = process.env.PLAYWRIGHT_K8S_NAMESPACE ?? 'mycelis';
    const podName = k8sCorePodName();
    const k8sPath = k8sLogPathForTarget(candidate);
    try {
        execFileSync('kubectl', ['exec', '-n', namespace, podName, '--', 'sh', '-c', `test -f ${shellQuote(k8sPath)}`], {
            stdio: 'ignore',
            timeout: 10_000,
        });
        return true;
    } catch {
        return false;
    }
}

function anyTargetExists(paths: string[]) {
    return paths.some((candidate) => fs.existsSync(candidate) || k8sTargetExists(candidate));
}

function removeExistingTargets(paths: string[]) {
    for (const candidate of paths) {
        if (fs.existsSync(candidate)) {
            try {
                fs.rmSync(candidate, { force: true });
            } catch (error) {
                const code = (error as NodeJS.ErrnoException).code;
                if (code !== 'EACCES' && code !== 'EPERM') {
                    throw error;
                }
            }
        }
        if (k8sWorkspaceProbeEnabled()) {
            const namespace = process.env.PLAYWRIGHT_K8S_NAMESPACE ?? 'mycelis';
            const podName = k8sCorePodName();
            const k8sPath = k8sLogPathForTarget(candidate);
            try {
                execFileSync('kubectl', ['exec', '-n', namespace, podName, '--', 'sh', '-c', `rm -f ${shellQuote(k8sPath)}`], {
                    stdio: 'ignore',
                    timeout: 10_000,
                });
            } catch {
                // Cleanup is best-effort because failed assertions should preserve the
                // original browser/API failure instead of hiding it behind kubectl noise.
            }
        }
    }
}

type ChatEnvelope = {
    ok?: boolean;
    data?: {
        mode?: string;
        payload?: {
            text?: string;
            ask_class?: string;
            proposal?: {
                confirm_token?: string;
                intent_proof_id?: string;
            };
        };
    };
};

type OrganizationEnvelope = {
    data?: {
        id?: string;
    };
};

type ConfirmEnvelope = {
    data?: {
        run_id?: string;
        verified?: boolean;
        execution_state?: string;
    };
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

async function createOrganization(page: Page, name: string) {
    const response = await page.request.post('/api/v1/organizations', {
        data: {
            name,
            purpose: 'Live governance verification',
            start_mode: 'empty',
        },
    });
    const parsed = await parseJSONIfPossible<OrganizationEnvelope>(response);
    expect(response.ok(), parsed.body ? JSON.stringify(parsed.body) : parsed.raw).toBeTruthy();
    const body = parsed.body;
    expect(body?.data?.id).toBeTruthy();
    return body?.data?.id as string;
}

async function openWorkspace(page: Page, organizationId: string) {
    await openOrganizationWorkspace(page, organizationId);
}

async function submitWorkspaceChat(page: Page, content: string) {
    const input = organizationChatInput(page);
    await input.fill(content);
    const responsePromise = page.waitForResponse(
        (response) => response.url().includes('/api/v1/chat') && response.request().method() === 'POST',
        { timeout: LIVE_CHAT_RESPONSE_TIMEOUT_MS },
    );
    await input.press('Enter');
    const response = await responsePromise;
    const parsed = await parseJSONIfPossible<ChatEnvelope>(response);
    return {
        response,
        raw: parsed.raw,
        body: parsed.body,
    };
}

async function waitForConfirmAction(page: Page) {
    const responsePromise = page.waitForResponse(
        (response) => response.url().includes('/api/v1/intent/confirm-action') && response.request().method() === 'POST',
        { timeout: LIVE_CHAT_RESPONSE_TIMEOUT_MS },
    );
    await page.getByRole('button', { name: /Approve & Execute|Execute/i }).click();
    const response = await responsePromise;
    const parsed = await parseJSONIfPossible<ConfirmEnvelope>(response);
    return {
        response,
        raw: parsed.raw,
        body: parsed.body,
    };
}

async function responseText(response: { text(): Promise<string> }) {
    return response.text();
}

test.describe('Soma governed mutation live contract', () => {
    test.skip(!process.env.PLAYWRIGHT_LIVE_BACKEND, 'requires a live Core backend');
    test.describe.configure({ timeout: LIVE_GOVERNANCE_TIMEOUT_MS });

    test('Scenario A: direct Soma answer works in a fresh organization', async ({ page }) => {
        test.slow();
        const organizationId = await createOrganization(page, `QA Scenario A ${Date.now()}`);
        await openWorkspace(page, organizationId);

        const { response, body } = await submitWorkspaceChat(page, 'Summarize the current Workspace V8 design objectives.');

        expect(response.ok(), body ? JSON.stringify(body) : await responseText(response)).toBeTruthy();
        expect(body?.data?.mode).toBe('answer');
        expect(body?.data?.payload?.ask_class).toBe('direct_answer');
        expect((body?.data?.payload?.text ?? '').trim().length).toBeGreaterThan(0);
        await expect(page.getByText(/could not produce a readable reply/i)).toHaveCount(0);
    });

    test('Scenario B: mutation requests still route to proposal mode after prior answer-mode chat', async ({ page }) => {
        test.slow();
        const organizationId = await createOrganization(page, `QA Scenario B ${Date.now()}`);
        await openWorkspace(page, organizationId);

        const direct = await submitWorkspaceChat(page, 'Summarize the current Workspace V8 design objectives.');
        expect(direct.response.ok(), direct.body ? JSON.stringify(direct.body) : direct.raw).toBeTruthy();
        expect(direct.body?.data?.mode).toBe('answer');

        const mutation = await submitWorkspaceChat(
            page,
            `Create a simple python file named workspace/logs/qa_browser_mixed_${Date.now()}.py that prints hello world.`,
        );

        expect(mutation.response.ok(), mutation.body ? JSON.stringify(mutation.body) : mutation.raw).toBeTruthy();
        expect(mutation.body?.data?.mode).toBe('proposal');
        expect(mutation.body?.data?.payload?.ask_class).toBe('governed_mutation');
        await expect(page.getByText('PROPOSED ACTION')).toBeVisible({ timeout: 30_000 });
        await expect(page.getByText(/Awaiting approval/i)).toBeVisible();
    });

    test('Scenario C+D: fresh mutation proposal stays side-effect free until confirm, and cancel remains safe + persistent', async ({ page }) => {
        test.slow();
        const stamp = Date.now();
        const targetPaths = resolveBackendLogTargets(`qa_browser_cancel_${stamp}.py`);
        const organizationId = await createOrganization(page, `QA Scenario CD ${stamp}`);
        await openWorkspace(page, organizationId);
        try {
            const mutation = await submitWorkspaceChat(
                page,
                `Create a simple python file named workspace/logs/qa_browser_cancel_${stamp}.py that prints hello world.`,
            );

            expect(mutation.response.ok(), mutation.body ? JSON.stringify(mutation.body) : mutation.raw).toBeTruthy();
            expect(mutation.body?.data?.mode).toBe('proposal');
            expect(mutation.body?.data?.payload?.ask_class).toBe('governed_mutation');
            await expect(page.getByText('PROPOSED ACTION')).toBeVisible({ timeout: 30_000 });
            expect(anyTargetExists(targetPaths)).toBeFalsy();

            await page.getByRole('button', { name: /^Cancel$/i }).click();
            await expect(page.getByText(/Proposal cancelled\. No action executed\./i)).toBeVisible({ timeout: 30_000 });
            expect(anyTargetExists(targetPaths)).toBeFalsy();

            await page.reload({ waitUntil: 'domcontentloaded' });
            await waitForOrganizationWorkspaceReady(page);
            await expect(page.getByText(/Proposal cancelled\. No action executed\./i)).toBeVisible({ timeout: 30_000 });
            expect(anyTargetExists(targetPaths)).toBeFalsy();
        } finally {
            removeExistingTargets(targetPaths);
        }
    });

    test('Scenario E: confirm yields durable proof, executes after approval, and persists on reload', async ({ page }) => {
        test.slow();
        const stamp = Date.now();
        const targetPaths = resolveBackendLogTargets(`qa_browser_confirm_${stamp}.py`);
        const organizationId = await createOrganization(page, `QA Scenario E ${stamp}`);
        await openWorkspace(page, organizationId);
        try {
            const mutation = await submitWorkspaceChat(
                page,
                `Create a simple python file named workspace/logs/qa_browser_confirm_${stamp}.py that prints hello world.`,
            );

            expect(mutation.response.ok(), mutation.body ? JSON.stringify(mutation.body) : mutation.raw).toBeTruthy();
            expect(mutation.body?.data?.mode).toBe('proposal');
            expect(mutation.body?.data?.payload?.ask_class).toBe('governed_mutation');
            await expect(page.getByText('PROPOSED ACTION')).toBeVisible({ timeout: 30_000 });
            expect(anyTargetExists(targetPaths)).toBeFalsy();

            const confirmed = await waitForConfirmAction(page);

            expect(confirmed.response.ok(), confirmed.body ? JSON.stringify(confirmed.body) : confirmed.raw).toBeTruthy();
            expect(typeof confirmed.body?.data?.run_id).toBe('string');
            expect((confirmed.body?.data?.run_id ?? '').trim().length).toBeGreaterThan(0);
            expect(confirmed.body?.data?.verified).toBeTruthy();
            expect(confirmed.body?.data?.execution_state).toBe('verified');

            await expect
                .poll(() => anyTargetExists(targetPaths), {
                    timeout: 30_000,
                    message: `expected one backend workspace target to exist after confirmation: ${targetPaths.join(', ')}`,
                })
                .toBeTruthy();

            await expect(page.getByText(/Execution verified/i)).toBeVisible({ timeout: 30_000 });
            await expect(page.getByRole('link', { name: /Mission activated/i })).toBeVisible({ timeout: 30_000 });

            await page.reload({ waitUntil: 'domcontentloaded' });
            await waitForOrganizationWorkspaceReady(page);
            await expect(page.getByText(/Execution verified/i)).toBeVisible({ timeout: 30_000 });
            await expect(page.getByRole('link', { name: /Mission activated/i })).toBeVisible({ timeout: 30_000 });
        } finally {
            removeExistingTargets(targetPaths);
        }
    });
});
