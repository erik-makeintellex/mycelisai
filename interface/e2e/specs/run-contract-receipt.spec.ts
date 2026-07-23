import { expect, test } from '@playwright/test';

test('completed work without its approved deliverable is recoverable and inspectable', async ({ page }, testInfo) => {
    const runId = 'run-output-recovery-1001';
    const now = new Date().toISOString();
    const consoleErrors: string[] = [];
    const pageErrors: string[] = [];
    page.on('console', (message) => {
        if (message.type() === 'error') consoleErrors.push(message.text());
    });
    page.on('pageerror', (error) => pageErrors.push(error.message));

    await page.route(`**/api/v1/runs/${runId}/events**`, async (route) => {
        await route.fulfill({
            status: 200,
            contentType: 'application/json',
            body: JSON.stringify({
                ok: true,
                data: [
                    {
                        id: 'event-work-contract',
                        run_id: runId,
                        tenant_id: 'default',
                        event_type: 'team_work.status',
                        severity: 'info',
                        emitted_at: now,
                        payload: {
                            execution_mode: 'team_async',
                            work_intent: {
                                kind: 'project',
                                output_contract: {
                                    shape: 'app_package',
                                    primary_deliverable: 'generated/application/index.html',
                                    retention: 'user_deliverable',
                                },
                                lifecycle: {
                                    control_summary: 'Pause, resume, or restore retained work.',
                                },
                            },
                        },
                    },
                    {
                        id: 'event-completed',
                        run_id: runId,
                        tenant_id: 'default',
                        event_type: 'mission.completed',
                        severity: 'info',
                        emitted_at: now,
                        payload: { operator_summary: 'Work finished.', execution_mode: 'team_async' },
                    },
                ],
            }),
        });
    });

    await page.goto(`/runs/${runId}?tab=events`, { waitUntil: 'domcontentloaded' });
    await expect(page.getByLabel('Run receipt')).toBeVisible();
    await expect(page.getByLabel('Outcome health: Degraded')).toBeVisible();
    await expect(page.getByText('Run needs output recovery')).toBeVisible();
    await expect(page.getByText(/approved deliverable was not retained/i)).toBeVisible();
    await expect(page.getByText(/result is not ready to rely on/i)).toBeVisible();
    await expect(page.getByText('Approved work')).not.toBeVisible();

    await page.getByRole('button', { name: /Inspect receipt evidence/i }).click();
    await expect(page.getByText('Approved work')).toBeVisible();
    await expect(page.getByText('project · team_async')).toBeVisible();
    await expect(page.getByText('generated/application/index.html')).toBeVisible();
    await expect(page.getByText('Pause, resume, or restore retained work.')).toBeVisible();

    const overflow = await page.evaluate(() => document.documentElement.scrollWidth - document.documentElement.clientWidth);
    expect(overflow).toBeLessThanOrEqual(1);
    expect(consoleErrors).toEqual([]);
    expect(pageErrors).toEqual([]);
    await page.screenshot({ path: testInfo.outputPath('run-output-recovery.png'), fullPage: true });
});
