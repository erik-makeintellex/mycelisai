import { test, expect } from '@playwright/test';

async function issueLiveConfirmToken(request: import('@playwright/test').APIRequestContext) {
    const apiKey = process.env.MYCELIS_API_KEY;
    if (!apiKey) {
        throw new Error('MYCELIS_API_KEY is required for live-backend Launch Crew confirmation coverage');
    }

    const response = await request.post('http://127.0.0.1:8081/api/v1/groups', {
        headers: {
            Authorization: `Bearer ${apiKey}`,
            'Content-Type': 'application/json',
        },
        data: {
            name: 'launch-crew-live-confirm',
            goal_statement: 'Generate a live confirm token for Launch Crew browser verification',
            work_mode: 'execute_with_approval',
            allowed_capabilities: ['groups:write'],
            member_user_ids: [],
            team_ids: [],
            coordinator_profile: 'admin',
            approval_policy_ref: 'live-backend-test',
            confirm_token: '',
        },
    });

    expect(response.status()).toBe(202);
    const body = await response.json();
    const data = body?.data ?? {};
    const confirmToken = data?.confirm_token?.token;
    const intentProofId = data?.intent_proof?.id;
    expect(typeof confirmToken).toBe('string');
    expect(confirmToken.length).toBeGreaterThan(0);
    expect(typeof intentProofId).toBe('string');
    return { confirmToken, intentProofId };
}

test.describe('Mission Proposal Entry Points', () => {
    test('workspace exposes launch controls for mission planning', async ({ page }) => {
        await page.goto('/dashboard');
        await page.waitForLoadState('domcontentloaded');

        await expect(page.locator('h1:has-text("Workspace")')).toBeVisible();
        await expect(page.locator('button:has-text("Launch Crew"), button:has-text("Launch")').first()).toBeVisible();
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

        await page.goto('/dashboard');
        await page.waitForLoadState('domcontentloaded');

        await page.getByRole('button', { name: /Launch Crew/i }).click();
        await page.getByPlaceholder('Describe the outcome you need...').fill('Create a documentation delivery crew');
        await page.getByRole('button', { name: /Send to Soma/i }).click();

        const modal = page.locator('.fixed.inset-0.z-50').last();
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

        await page.goto('/dashboard');
        await page.waitForLoadState('domcontentloaded');

        await page.getByRole('button', { name: /Launch Crew/i }).click();
        await page.getByPlaceholder('Describe the outcome you need...').fill('Launch a crew for deployment recovery');
        await page.getByRole('button', { name: /Send to Soma/i }).click();

        const modal = page.locator('.fixed.inset-0.z-50').last();
        await expect(modal.getByText(/Launch Crew is blocked/i)).toBeVisible();
        await expect(modal.getByText(/Soma chat blocked \(500\)/i)).toBeVisible();
        await expect(modal.getByRole('button', { name: /Revise request/i })).toBeVisible();
        await expect(modal.getByRole('button', { name: /Continue in chat/i })).toBeVisible();
    });

    test.skip(!process.env.PLAYWRIGHT_LIVE_BACKEND, 'requires a live Core backend');

    test('launch crew confirm path returns durable proof when the live backend verifies execution', async ({ page, request }) => {
        const { confirmToken, intentProofId } = await issueLiveConfirmToken(request);

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

        await page.goto('/dashboard');
        await page.waitForLoadState('domcontentloaded');

        await page.getByRole('button', { name: /Launch Crew/i }).click();
        await page.getByPlaceholder('Describe the outcome you need...').fill('Create a live-backed workflow onboarding crew');
        await page.getByRole('button', { name: /Send to Soma/i }).click();

        const modal = page.locator('.fixed.inset-0.z-50').last();
        await expect(modal.getByText(/prepared a crew proposal/i)).toBeVisible();

        const confirmResponsePromise = page.waitForResponse(
            (response) =>
                response.url().includes('/api/v1/intent/confirm-action') &&
                response.request().method() === 'POST',
        );

        await modal.getByRole('button', { name: /^Launch Crew$/i }).click();

        const confirmResponse = await confirmResponsePromise;
        expect(confirmResponse.ok()).toBeTruthy();
        const confirmBody = await confirmResponse.json();
        expect(confirmBody?.data?.confirmed).toBeTruthy();

        expect(typeof confirmBody?.data?.run_id).toBe('string');
        expect(confirmBody?.data?.run_id?.length ?? 0).toBeGreaterThan(0);
        expect(confirmBody?.data?.verified).toBeTruthy();

        await expect(modal.getByText(/Mission activated/i)).toBeVisible();
    });
});
