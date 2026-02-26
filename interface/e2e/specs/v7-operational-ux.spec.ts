import { test, expect, type Page } from '@playwright/test';

async function installStableEventSource(page: Page) {
    await page.addInitScript(() => {
        class StableEventSource {
            url: string;
            onopen: ((ev: Event) => void) | null = null;
            onmessage: ((ev: MessageEvent) => void) | null = null;
            onerror: ((ev: Event) => void) | null = null;

            constructor(url: string) {
                this.url = url;
                setTimeout(() => {
                    if (this.onopen) this.onopen(new Event('open'));
                }, 0);
            }

            close() {
                // no-op for tests
            }
        }

        // @ts-expect-error test-time EventSource override
        window.EventSource = StableEventSource;
        localStorage.removeItem('workspace-focus-mode');
    });
}

test.describe('V7 Operational UX Gate A', () => {
    test('degraded banner appears and clears after recovery without reload', async ({ page }) => {
        await installStableEventSource(page);

        let degraded = true;

        await page.route('**/api/v1/services/status', async (route) => {
            const data = degraded
                ? [
                    { name: 'nats', status: 'offline' },
                    { name: 'postgres', status: 'online' },
                ]
                : [
                    { name: 'nats', status: 'online' },
                    { name: 'postgres', status: 'online' },
                ];
            await route.fulfill({
                status: 200,
                contentType: 'application/json',
                body: JSON.stringify({ data }),
            });
        });

        await page.route('**/api/v1/council/members', async (route) => {
            await route.fulfill({
                status: 200,
                contentType: 'application/json',
                body: JSON.stringify({ ok: true, data: [{ id: 'admin', role: 'admin', team: 'admin-core' }] }),
            });
        });

        await page.goto('/dashboard');
        await page.waitForLoadState('networkidle');

        const banner = page.getByText(/System in Degraded Mode/i);
        await expect(banner).toBeVisible();

        degraded = false;
        await page.getByRole('button', { name: 'Retry' }).first().click();
        await expect(banner).toBeHidden({ timeout: 5000 });
        expect(page.url()).toContain('/dashboard');
    });

    test('status drawer is globally accessible', async ({ page }) => {
        await installStableEventSource(page);

        await page.route('**/api/v1/services/status', async (route) => {
            await route.fulfill({
                status: 200,
                contentType: 'application/json',
                body: JSON.stringify({
                    data: [
                        { name: 'nats', status: 'online' },
                        { name: 'postgres', status: 'online' },
                    ],
                }),
            });
        });

        await page.route('**/api/v1/council/members', async (route) => {
            await route.fulfill({
                status: 200,
                contentType: 'application/json',
                body: JSON.stringify({ ok: true, data: [{ id: 'admin', role: 'admin', team: 'admin-core' }] }),
            });
        });

        await page.goto('/dashboard');
        await page.waitForLoadState('networkidle');

        await page.getByTitle('Open Status Drawer').first().click();
        await expect(page.getByRole('dialog', { name: 'System status drawer' })).toBeVisible();
        await expect(page.getByText('System Status')).toBeVisible();

        await page.getByLabel('Close status drawer').click();
        await expect(page.getByRole('dialog', { name: 'System status drawer' })).toBeHidden();
    });

    test('council failure reroutes via Soma in one click', async ({ page }) => {
        await installStableEventSource(page);

        let adminCalls = 0;

        await page.route('**/api/v1/services/status', async (route) => {
            await route.fulfill({
                status: 200,
                contentType: 'application/json',
                body: JSON.stringify({
                    data: [
                        { name: 'nats', status: 'online' },
                        { name: 'postgres', status: 'online' },
                    ],
                }),
            });
        });

        await page.route('**/api/v1/council/members', async (route) => {
            await route.fulfill({
                status: 200,
                contentType: 'application/json',
                body: JSON.stringify({
                    ok: true,
                    data: [
                        { id: 'admin', role: 'admin', team: 'admin-core' },
                        { id: 'council-sentry', role: 'sentry', team: 'council-core' },
                    ],
                }),
            });
        });

        await page.route('**/api/v1/council/**/chat', async (route) => {
            const url = route.request().url();
            if (url.includes('/api/v1/council/council-sentry/chat')) {
                await route.fulfill({
                    status: 500,
                    contentType: 'application/json',
                    body: JSON.stringify({ error: 'Council agent unreachable (500)' }),
                });
                return;
            }
            if (url.includes('/api/v1/council/admin/chat')) {
                adminCalls += 1;
                await route.fulfill({
                    status: 200,
                    contentType: 'application/json',
                    body: JSON.stringify({
                        ok: true,
                        data: {
                            meta: { source_node: 'admin', timestamp: new Date().toISOString() },
                            signal_type: 'chat_response',
                            trust_score: 0.8,
                            template_id: 'chat-to-answer',
                            mode: 'answer',
                            payload: { text: 'Recovered via Soma', consultations: [], tools_used: [] },
                        },
                    }),
                });
                return;
            }
            await route.continue();
        });

        await page.goto('/dashboard');
        await page.waitForLoadState('networkidle');

        await page.getByTitle('Direct message to a specific council member').click();
        await page.getByRole('button', { name: 'Sentry' }).click();

        const input = page.getByPlaceholder(/Ask Soma|Direct to/i);
        await input.fill('reroute this request');
        await input.press('Enter');

        await expect(page.getByText('Council Call Failed')).toBeVisible();
        await page.getByRole('button', { name: 'Switch to Soma' }).click();

        await expect(page.getByText('Recovered via Soma')).toBeVisible();
        expect(adminCalls).toBeGreaterThan(0);
    });

    test('automations landing is actionable when scheduler is not active', async ({ page }) => {
        await page.goto('/automations');
        await page.waitForLoadState('networkidle');

        await expect(page.getByText('Available Now')).toBeVisible();
        await expect(page.getByText('Coming Soon')).toBeVisible();
        await expect(page.getByText('Scheduler')).toBeVisible();
        await expect(page.getByRole('button', { name: 'Set Up Your First Automation Chain' })).toBeVisible();
    });

    test('system quick checks update timestamp when a check is run', async ({ page }) => {
        await page.route('**/api/v1/services/status', async (route) => {
            await route.fulfill({
                status: 200,
                contentType: 'application/json',
                body: JSON.stringify({
                    data: [
                        { name: 'nats', status: 'online' },
                        { name: 'postgres', status: 'online' },
                        { name: 'reactive', status: 'online' },
                    ],
                }),
            });
        });

        await page.route('**/api/v1/telemetry/compute', async (route) => {
            await route.fulfill({
                status: 200,
                contentType: 'application/json',
                body: JSON.stringify({
                    goroutines: 10,
                    heap_alloc_mb: 5,
                    sys_mem_mb: 12,
                    llm_tokens_sec: 1.2,
                    timestamp: new Date().toISOString(),
                }),
            });
        });

        await page.goto('/system?tab=health');
        await page.waitForLoadState('networkidle');

        await expect(page.getByText('Quick Checks')).toBeVisible();

        const natsRow = page.locator('div').filter({ hasText: 'NATS connected' }).first();
        await natsRow.getByRole('button', { name: 'Run Check' }).click();
        await expect(natsRow).toContainText('Last checked:');
        await expect(natsRow).not.toContainText('never');
    });

    test('focus mode toggles with F key without reload', async ({ page }) => {
        await installStableEventSource(page);

        await page.route('**/api/v1/services/status', async (route) => {
            await route.fulfill({
                status: 200,
                contentType: 'application/json',
                body: JSON.stringify({
                    data: [
                        { name: 'nats', status: 'online' },
                        { name: 'postgres', status: 'online' },
                    ],
                }),
            });
        });

        await page.goto('/dashboard');
        await page.waitForLoadState('networkidle');

        await expect(page.getByText('Press F to toggle')).toBeHidden();
        const before = page.url();

        await page.keyboard.press('f');
        await expect(page.getByText('Press F to toggle')).toBeVisible();
        expect(page.url()).toBe(before);

        await page.keyboard.press('f');
        await expect(page.getByText('Press F to toggle')).toBeHidden();
        expect(page.url()).toBe(before);
    });
});

