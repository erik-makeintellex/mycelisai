import type { Page } from '@playwright/test';

export function organizationChatInput(page: Page) {
    return page.getByPlaceholder(/Tell Soma what you want to plan, review, create, or (execute|run)/i);
}

export async function waitForOrganizationWorkspaceReady(page: Page) {
    await organizationChatInput(page).waitFor({ timeout: 30_000 });
}

export async function openOrganizationWorkspace(page: Page, organizationId: string) {
    let lastError: unknown;
    for (let attempt = 0; attempt < 3; attempt++) {
        await page.goto(`/organizations/${organizationId}`, { waitUntil: 'domcontentloaded' });
        try {
            await waitForOrganizationWorkspaceReady(page);
            return;
        } catch (error) {
            lastError = error;
        }
    }
    throw lastError;
}
