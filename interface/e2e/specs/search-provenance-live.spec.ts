import { expect, test, type Page } from '@playwright/test';
import { openOrganizationWorkspace, organizationChatInput } from '../support/live-organization-workspace';

const LIVE_SEARCH_TIMEOUT_MS = 180_000;
const LIVE_CHAT_RESPONSE_TIMEOUT_MS = 120_000;

type ChatEnvelope = {
    ok?: boolean;
    data?: {
        mode?: string;
        payload?: {
            ask_class?: string;
            execution_summary?: {
                execution?: {
                    shape?: string;
                    status?: string;
                };
                capability_use?: Array<{
                    id?: string;
                    reason?: string;
                }>;
            };
        };
    };
};

type OrganizationEnvelope = {
    data?: {
        id?: string;
    };
};

async function parseJSONIfPossible<T>(response: { text(): Promise<string> }) {
    const raw = await response.text();
    try {
        return {
            raw,
            body: JSON.parse(raw) as T,
        };
    } catch {
        return {
            raw,
            body: null as T | null,
        };
    }
}

async function createOrganization(page: Page, name: string) {
    const response = await page.request.post('/api/v1/organizations', {
        data: {
            name,
            purpose: 'Live search provenance verification',
            start_mode: 'empty',
        },
    });
    const parsed = await parseJSONIfPossible<OrganizationEnvelope>(response);
    expect(response.ok(), parsed.body ? JSON.stringify(parsed.body) : parsed.raw).toBeTruthy();
    const organizationId = parsed.body?.data?.id;
    expect(organizationId).toBeTruthy();
    return organizationId as string;
}

async function submitWorkspaceChat(page: Page, content: string) {
    const input = organizationChatInput(page);
    await input.fill(content);
    const submit = page.getByRole('button', { name: /Submit Soma request/i });
    await expect(submit).toBeEnabled({ timeout: 10_000 });
    const responsePromise = page.waitForResponse(
        (response) => response.url().includes('/api/v1/chat') && response.request().method() === 'POST',
        { timeout: LIVE_CHAT_RESPONSE_TIMEOUT_MS },
    );
    await submit.click();
    const response = await responsePromise;
    const parsed = await parseJSONIfPossible<ChatEnvelope>(response);
    return {
        response,
        raw: parsed.raw,
        body: parsed.body,
    };
}

test.describe('Soma search provenance live contract', () => {
    test.skip(!process.env.PLAYWRIGHT_LIVE_BACKEND, 'requires a live Core backend');
    test.describe.configure({ timeout: LIVE_SEARCH_TIMEOUT_MS });

    test('direct web search shows governed source proof', async ({ page }) => {
        const organizationId = await createOrganization(page, `QA Search Provenance ${Date.now()}`);
        await openOrganizationWorkspace(page, organizationId);

        const { response, body, raw } = await submitWorkspaceChat(page, 'try doing a search on hermes-agent');

        expect(response.ok(), body ? JSON.stringify(body) : raw).toBeTruthy();
        expect(body?.data?.mode).toBe('answer');
        expect(body?.data?.payload?.ask_class).toBe('direct_answer');
        expect(body?.data?.payload?.execution_summary?.execution?.shape).toBe('tool_assisted_work');
        expect(body?.data?.payload?.execution_summary?.execution?.status).toBe('completed');
        const searchUse = body?.data?.payload?.execution_summary?.capability_use?.find((item) => item.id === 'web_search');
        expect(searchUse?.id).toBe('web_search');
        expect(searchUse?.reason).toBeTruthy();

        await expect(page.getByText('Result verified').last()).toBeVisible();
        await expect(page.getByText('Soma Search').last()).toBeVisible();
        await expect(page.getByText(/public web/i).last()).toBeVisible();
        await expect(page.getByText('web_search')).toHaveCount(0);
        await expect(page.getByText(/via builtin_web/i)).toHaveCount(0);
        await expect(page.getByText(/Workspace chat unreachable/i)).toHaveCount(0);
    });

    test('generic freshness wording uses configured web search', async ({ page }) => {
        const organizationId = await createOrganization(page, `QA Public Search Boundary ${Date.now()}`);
        await openOrganizationWorkspace(page, organizationId);

        const { response, body, raw } = await submitWorkspaceChat(page, 'can you search on whats the latest popular multi agent framework?');

        expect(response.ok(), body ? JSON.stringify(body) : raw).toBeTruthy();
        expect(body?.data?.mode).toBe('answer');
        expect(body?.data?.payload?.ask_class).toBe('direct_answer');
        expect(body?.data?.payload?.execution_summary?.execution?.shape).toBe('tool_assisted_work');
        expect(body?.data?.payload?.execution_summary?.execution?.status).toBe('completed');
        const searchUse = body?.data?.payload?.execution_summary?.capability_use?.find((item) => item.id === 'web_search');
        expect(searchUse?.reason).toContain('External or public web provider');

        await expect(page.getByText('Result verified').last()).toBeVisible();
        await expect(page.getByText(/public web research is not configured/i)).toHaveCount(0);
        await expect(page.getByText('Soma Search').last()).toBeVisible();
        await expect(page.getByText(/public web/i).last()).toBeVisible();
        await expect(page.getByText(/via builtin_web/i)).toHaveCount(0);
        await expect(page.getByText(/Workspace chat unreachable/i)).toHaveCount(0);
    });
});
