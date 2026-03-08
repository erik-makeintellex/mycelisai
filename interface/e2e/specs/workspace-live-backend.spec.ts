import { test, expect } from '@playwright/test';

function expectedCouncilLabel(member: { id: string; role: string }) {
    const labels: Record<string, string> = {
        admin: 'Soma',
        'council-architect': 'Architect',
        'council-coder': 'Coder',
        'council-creative': 'Creative',
        'council-sentry': 'Sentry',
    };
    return labels[member.id] ?? member.role;
}

test.describe('Workspace live backend contract', () => {
    test.skip(!process.env.PLAYWRIGHT_LIVE_BACKEND, 'requires a live Core backend');

    test('dashboard binds live status and council data through the UI proxy', async ({ page }) => {
        const statusPromise = page.waitForResponse(
            (response) =>
                response.url().includes('/api/v1/services/status') &&
                response.request().method() === 'GET',
        );
        const membersPromise = page.waitForResponse(
            (response) =>
                response.url().includes('/api/v1/council/members') &&
                response.request().method() === 'GET',
        );

        await page.goto('/dashboard');
        await page.waitForLoadState('domcontentloaded');

        const statusResponse = await statusPromise;
        const membersResponse = await membersPromise;

        expect(statusResponse.ok()).toBeTruthy();
        expect(membersResponse.ok()).toBeTruthy();

        const statusBody = await statusResponse.json();
        const services = Array.isArray(statusBody?.data) ? statusBody.data : [];
        expect(services.length).toBeGreaterThan(0);
        expect(services.some((service: { name?: string }) => service.name === 'nats')).toBeTruthy();
        expect(services.some((service: { name?: string }) => service.name === 'postgres')).toBeTruthy();

        const membersBody = await membersResponse.json();
        const members = Array.isArray(membersBody?.data) ? membersBody.data : [];
        expect(members.length).toBeGreaterThan(0);

        const directButton = page.getByRole('button', { name: /^Direct$/ });
        await expect(directButton).toBeVisible();
        await directButton.click();

        const nonAdminMember = members.find((member: { id: string }) => member.id !== 'admin');
        if (nonAdminMember) {
            await expect(
                page.getByRole('button', { name: expectedCouncilLabel(nonAdminMember) }),
            ).toBeVisible();
        }

        await expect(page.getByTestId('mission-chat')).toBeVisible();
        await expect(page.getByText(/isn't running yet/i)).toHaveCount(0);
    });
});
