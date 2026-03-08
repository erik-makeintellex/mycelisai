import { test, expect } from '@playwright/test';

test.describe('Mission Proposal Entry Points', () => {
    test('workspace exposes launch controls for mission planning', async ({ page }) => {
        await page.goto('/dashboard');
        await page.waitForLoadState('domcontentloaded');

        await expect(page.locator('h1:has-text("Workspace")')).toBeVisible();
        await expect(page.locator('button:has-text("Launch Crew"), button:has-text("Launch")').first()).toBeVisible();
    });

    test('launch crew reaches a proposal outcome instead of a planning-only modal state', async ({ page }) => {
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
    });
});
