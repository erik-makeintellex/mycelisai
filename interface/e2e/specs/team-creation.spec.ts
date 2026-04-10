import { test, expect, type Route } from '@playwright/test';

async function fulfillJSON(route: Route, status: number, body: unknown) {
    await route.fulfill({
        status,
        contentType: 'application/json',
        body: JSON.stringify(body),
    });
}

test.describe('Guided Team Creation (/teams/create)', () => {
    test.beforeEach(async ({ page }) => {
        await page.addInitScript(() => {
            window.localStorage.setItem('mycelis-last-organization-id', 'org-123');
            window.localStorage.setItem('mycelis-last-organization-name', 'Northstar Labs');
        });

        await page.route('**/api/v1/organizations/org-123/home', async (route) => {
            await fulfillJSON(route, 200, {
                ok: true,
                data: {
                    id: 'org-123',
                    name: 'Northstar Labs',
                    purpose: 'Ship a focused AI engineering organization for product delivery.',
                    start_mode: 'template',
                    team_lead_label: 'Launch Lead',
                    advisor_count: 1,
                    department_count: 2,
                    specialist_count: 4,
                    ai_engine_settings_summary: 'Balanced',
                    response_contract_summary: 'Clear and structured',
                    memory_personality_summary: 'Stable continuity',
                    status: 'active',
                },
            });
        });

        await page.route('**/api/v1/organizations/org-123/workspace/actions', async (route) => {
            const body = route.request().postDataJSON();
            await fulfillJSON(route, 200, {
                ok: true,
                data: {
                    action: body.action,
                    request_label: 'Run a quick strategy check',
                    headline: 'Launch team plan for Northstar Labs',
                    summary: 'Soma recommends a launch-focused team with clear outputs.',
                    priority_steps: [
                        'Confirm the campaign outcomes the team owns.',
                        'Choose the first output to make visible this week.',
                    ],
                    suggested_follow_ups: ['Choose the first priority'],
                    execution_contract: {
                        execution_mode: 'native_team',
                        owner_label: 'Native Mycelis team',
                        team_name: 'Launch Delivery Team',
                        summary: 'Use a focused launch delivery team inside Northstar Labs.',
                        target_outputs: ['Campaign brief', 'Landing page draft'],
                    },
                },
            });
        });
    });

    test('guides the operator through current organization context and Soma team design', async ({ page }) => {
        await page.goto('/teams/create', { waitUntil: 'domcontentloaded' });

        await expect(page.getByRole('heading', { name: 'Create a team through Soma' })).toBeVisible();
        await expect(page.getByText('Current organization')).toBeVisible();
        await expect(
            page.locator('p').filter({ hasText: /^Northstar Labs$/ }).first(),
        ).toBeVisible();
        await expect(page.getByRole('button', { name: 'Marketing launch team' })).toBeVisible();

        await page.getByRole('button', { name: 'Marketing launch team' }).click();
        await expect(page.getByLabel('Tell Soma what team or delivery lane you want to create')).toHaveValue(/marketing launch team/i);

        await page.getByRole('button', { name: 'Start team design' }).click();

        await expect(page.getByText('Launch team plan for Northstar Labs')).toBeVisible();
        await expect(page.getByText('Launch Delivery Team', { exact: true })).toBeVisible();
        await expect(page.getByText('Campaign brief')).toBeVisible();
    });
});
