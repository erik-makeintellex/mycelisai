import fs from 'node:fs';
import path from 'node:path';
import { execFileSync } from 'node:child_process';
import { expect, type Page } from '@playwright/test';
import { organizationChatInput } from '../support/live-organization-workspace';
import { liveAPIHeaders, liveAPIURL } from '../support/live-api-auth';
import { clickVisibleControl } from '../support/click-visible-control';

const repoRoot = path.resolve(__dirname, '../../..');
export const LIVE_CHAT_RESPONSE_TIMEOUT_MS = 120_000;

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

export function resolveBackendLogTargets(filename: string) {
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

export function anyTargetExists(paths: string[]) {
    return paths.some((candidate) => fs.existsSync(candidate) || k8sTargetExists(candidate));
}

export function removeExistingTargets(paths: string[]) {
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
                // Cleanup is best-effort so the original browser/API failure stays visible.
            }
        }
    }
}

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

export async function createOrganization(page: Page, name: string) {
    const response = await page.request.post(liveAPIURL('/api/v1/organizations'), {
        headers: liveAPIHeaders(),
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

export async function submitWorkspaceChat(page: Page, content: string) {
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

export async function waitForConfirmAction(page: Page) {
    const responsePromise = page.waitForResponse(
        (response) => response.url().includes('/api/v1/intent/confirm-action') && response.request().method() === 'POST',
        { timeout: LIVE_CHAT_RESPONSE_TIMEOUT_MS },
    );
    const executeButton = page.getByRole('button', { name: /Approve & Execute|Execute/i }).last();
    await clickVisibleControl(page, executeButton, { timeout: 20_000 });
    const response = await responsePromise;
    const parsed = await parseJSONIfPossible<ConfirmEnvelope>(response);
    return {
        response,
        raw: parsed.raw,
        body: parsed.body,
    };
}

export async function responseText(response: { text(): Promise<string> }) {
    return response.text();
}
