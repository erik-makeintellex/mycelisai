import { expect, test } from '@playwright/test';

test.describe('Live documentation authority', () => {
    test('serves the canonical PRD and current acceptance guide from the real manifest', async ({ page }) => {
        await page.goto('/docs?doc=mycelis-canonical-prd', { waitUntil: 'domcontentloaded' });

        await expect(page.getByRole('heading', { name: 'Mycelis Canonical PRD' })).toBeVisible();
        await expect(page.getByText('Mycelis Canonical PRD').first()).toBeVisible();
        await expect(page.getByText('Architecture Overview')).toHaveCount(0);

        await page.getByText('User Acceptance', { exact: true }).click();
        await expect(page).toHaveURL(/\/docs\?doc=user-acceptance$/);
        await expect(page.getByRole('heading', { name: 'User Acceptance Runbook' })).toBeVisible();
        await expect(page.getByRole('heading', { name: 'Trusted Outcome Walkthrough' })).toBeVisible();
    });
});
