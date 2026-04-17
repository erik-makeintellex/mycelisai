import { test, expect } from '@playwright/test';
import fs from 'node:fs';
import path from 'node:path';

const repoRoot = path.resolve(__dirname, '../../..');

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
    return resolveBackendWorkspaceRoots().map((workspaceRoot) => path.join(workspaceRoot, 'workspace', 'logs', filename));
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
    }
}

async function issueLiveConfirmToken(page: import('@playwright/test').Page, targetPath: string) {
    const response = await page.request.post('/api/v1/chat', {
        data: {
            messages: [
                {
                    role: 'user',
                    content: `Create a simple python file named ${targetPath} that prints hello world.`,
                },
            ],
        },
    });

    const text = await response.text();
    let body: any = null;
    try {
        body = JSON.parse(text);
    } catch {
        body = null;
    }
    expect(response.ok(), body ? JSON.stringify(body) : text).toBeTruthy();
    expect(body?.data?.mode).toBe('proposal');

    const proposal = body?.data?.payload?.proposal ?? {};
    const confirmToken = proposal?.confirm_token;
    const intentProofId = proposal?.intent_proof_id;
    expect(typeof confirmToken).toBe('string');
    expect(confirmToken.length).toBeGreaterThan(0);
    expect(typeof intentProofId).toBe('string');
    return { confirmToken, intentProofId };
}

async function createLiveOrganization(page: import('@playwright/test').Page) {
    const response = await page.request.post('/api/v1/organizations', {
        data: {
            name: `Launch Crew E2E ${Date.now()}`,
            purpose: 'Live Launch Crew browser verification',
            start_mode: 'empty',
        },
    });
    expect(response.ok()).toBeTruthy();
    const body = await response.json();
    const organizationId = body?.data?.id;
    expect(typeof organizationId).toBe('string');
    return organizationId as string;
}

async function openCrewLauncher(page: import('@playwright/test').Page) {
    const organizationId = await createLiveOrganization(page);
    await page.goto(`/organizations/${organizationId}`);
    await page.waitForLoadState('domcontentloaded');
    await page.getByRole('button', { name: 'Create teams with Soma' }).click();
    await page.getByRole('button', { name: 'Open crew launcher' }).click();
    return page.locator('.fixed.inset-0.z-50').last();
}

test.describe('Mission Proposal Entry Points', () => {
    test.skip(!process.env.PLAYWRIGHT_LIVE_BACKEND, 'requires live AI Organization workspace');

    test('workspace exposes launch controls for mission planning', async ({ page }) => {
        const organizationId = await createLiveOrganization(page);
        await page.goto(`/organizations/${organizationId}`);
        await page.waitForLoadState('domcontentloaded');

        await expect(page.getByRole('heading', { name: /Soma for/i })).toBeVisible();
        await page.getByRole('button', { name: 'Create teams with Soma' }).click();
        await expect(page.getByRole('button', { name: 'Open crew launcher' })).toBeVisible();
    });

    test('launch crew reaches a proposal outcome and can be cancelled without executing', async ({ page }) => {
        await page.route('**/api/v1/council/members', async (route) => {
            await route.fulfill({
                json: {
                    ok: true,
                    data: [{ id: 'admin', role: 'admin', team: 'admin-core' }],
                },
            });
        });

        await page.route('**/api/v1/chat', async (route) => {
            await route.fulfill({
                json: {
                    ok: true,
                    data: {
                        meta: { source_node: 'admin', timestamp: '2026-03-07T10:00:00Z' },
                        signal_type: 'chat_response',
                        trust_score: 0.8,
                        mode: 'proposal',
                        payload: {
                            text: 'I can launch a crew for that request.',
                            proposal: {
                                intent: 'Create a documentation delivery crew',
                                tools: ['delegate_task', 'write_file'],
                                risk_level: 'medium',
                                confirm_token: 'ct-123',
                                intent_proof_id: 'ip-123',
                            },
                        },
                    },
                },
            });
        });

        const modal = await openCrewLauncher(page);
        await page.getByPlaceholder('Describe the outcome you need...').fill('Create a documentation delivery crew');
        await page.getByRole('button', { name: /Send to Soma/i }).click();

        await expect(modal.getByText(/prepared a crew proposal/i)).toBeVisible();
        await expect(modal.getByText(/Create a documentation delivery crew/i)).toBeVisible();
        await expect(modal.getByRole('button', { name: /^Launch Crew$/i })).toBeVisible();

        await modal.getByRole('button', { name: /Back to intent/i }).click();

        await expect(modal.getByPlaceholder('Describe the outcome you need...')).toBeVisible();
        await expect(modal.getByText(/prepared a crew proposal/i)).toHaveCount(0);
    });

    test('launch crew reaches a blocker outcome with recovery actions when Soma chat fails', async ({ page }) => {
        await page.route('**/api/v1/council/members', async (route) => {
            await route.fulfill({
                json: {
                    ok: true,
                    data: [{ id: 'admin', role: 'admin', team: 'admin-core' }],
                },
            });
        });

        await page.route('**/api/v1/chat', async (route) => {
            await route.fulfill({
                status: 500,
                contentType: 'application/json',
                body: JSON.stringify({ error: 'Soma chat blocked (500)' }),
            });
        });

        const modal = await openCrewLauncher(page);
        await page.getByPlaceholder('Describe the outcome you need...').fill('Launch a crew for deployment recovery');
        await page.getByRole('button', { name: /Send to Soma/i }).click();

        await expect(modal.getByText(/Launch Crew is blocked/i)).toBeVisible();
        await expect(modal.getByText(/Soma hit a server-side failure while handling the request/i)).toBeVisible();
        await expect(modal.getByRole('button', { name: /Revise request/i })).toBeVisible();
        await expect(modal.getByRole('button', { name: /Continue in chat/i })).toBeVisible();
    });

    test('launch crew confirm path returns durable proof when the live backend verifies execution', async ({ page }) => {
        const filename = `qa_launch_crew_confirm_${Date.now()}.py`;
        const targetPath = `workspace/logs/${filename}`;
        const targetPaths = resolveBackendLogTargets(filename);
        removeExistingTargets(targetPaths);
        const { confirmToken, intentProofId } = await issueLiveConfirmToken(page, targetPath);

        await page.route('**/api/v1/chat', async (route) => {
            await route.fulfill({
                json: {
                    ok: true,
                    data: {
                        meta: { source_node: 'admin', timestamp: '2026-03-07T10:00:00Z' },
                        signal_type: 'chat_response',
                        trust_score: 0.8,
                        mode: 'proposal',
                        payload: {
                            text: 'I can launch a crew for that request.',
                            proposal: {
                                intent: 'Create a live-backed workflow onboarding crew',
                                tools: ['delegate_task', 'write_file'],
                                risk_level: 'medium',
                                confirm_token: confirmToken,
                                intent_proof_id: intentProofId,
                            },
                        },
                    },
                },
            });
        });

        const modal = await openCrewLauncher(page);
        try {
            await page.getByPlaceholder('Describe the outcome you need...').fill('Create a live-backed workflow onboarding crew');
            await page.getByRole('button', { name: /Send to Soma/i }).click();

            await expect(modal.getByText(/prepared a crew proposal/i)).toBeVisible();

            const confirmResponsePromise = page.waitForResponse(
                (response) =>
                    response.url().includes('/api/v1/intent/confirm-action') &&
                    response.request().method() === 'POST',
            );

            await modal.getByRole('button', { name: /^Launch Crew$/i }).click();

            const confirmResponse = await confirmResponsePromise;
            const confirmText = await confirmResponse.text();
            let confirmBody: any = null;
            try {
                confirmBody = JSON.parse(confirmText);
            } catch {
                confirmBody = null;
            }
            expect(confirmResponse.ok(), confirmBody ? JSON.stringify(confirmBody) : confirmText).toBeTruthy();
            expect(confirmBody?.data?.confirmed).toBeTruthy();

            expect(typeof confirmBody?.data?.run_id).toBe('string');
            expect(confirmBody?.data?.run_id?.length ?? 0).toBeGreaterThan(0);
            expect(confirmBody?.data?.verified).toBeTruthy();
            expect(confirmBody?.data?.execution_state).toBe('verified');

            await expect(modal.getByText(/Mission activated/i)).toBeVisible();
        } finally {
            removeExistingTargets(targetPaths);
        }
    });
});
