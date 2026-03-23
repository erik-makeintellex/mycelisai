import { test, expect } from '@playwright/test';

test.describe('Workspace live backend contract', () => {
    test.skip(!process.env.PLAYWRIGHT_LIVE_BACKEND, 'requires a live Core backend');

    test('dashboard status surfaces bind live status and council data through the UI proxy', async ({ page }) => {
        test.slow();
        await page.goto('/dashboard');
        await page.waitForLoadState('domcontentloaded');
        const browserData = await page.evaluate(async () => {
            const [statusResponse, membersResponse] = await Promise.all([
                fetch('/api/v1/services/status'),
                fetch('/api/v1/council/members'),
            ]);
            const statusBody = await statusResponse.json();
            const membersBody = await membersResponse.json();
            return {
                statusOk: statusResponse.ok,
                membersOk: membersResponse.ok,
                statusBody,
                membersBody,
            };
        });
        expect(browserData.statusOk).toBeTruthy();
        expect(browserData.membersOk).toBeTruthy();

        const services = Array.isArray(browserData.statusBody?.data) ? browserData.statusBody.data : [];
        expect(services.length).toBeGreaterThan(0);
        expect(services.some((service: { name?: string }) => service.name === 'nats')).toBeTruthy();
        expect(services.some((service: { name?: string }) => service.name === 'postgres')).toBeTruthy();

        const members = Array.isArray(browserData.membersBody?.data) ? browserData.membersBody.data : [];
        expect(members.length).toBeGreaterThan(0);

        await page.getByRole('button', { name: /^Status$/ }).click();
        await expect(page.getByText('Council Reachability')).toBeVisible({ timeout: 15000 });
        await expect(page.getByText(/isn't running yet/i)).toHaveCount(0);
    });
});
