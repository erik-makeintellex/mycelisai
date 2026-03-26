import fs from 'node:fs';
import path from 'node:path';
import { expect, test, type Page } from '@playwright/test';

const workspaceLogsDir = path.resolve(__dirname, '../../..', 'core', 'workspace', 'workspace', 'logs');

type ChatEnvelope = {
    ok?: boolean;
    data?: {
        mode?: string;
        payload?: {
            text?: string;
            proposal?: {
                confirm_token?: string;
                intent_proof_id?: string;
            };
        };
    };
};

async function createOrganization(page: Page, name: string) {
    const response = await page.request.post('/api/v1/organizations', {
        data: {
            name,
            purpose: 'Live governance verification',
            start_mode: 'empty',
        },
    });
    expect(response.ok()).toBeTruthy();
    const body = await response.json();
    expect(body?.data?.id).toBeTruthy();
    return body.data.id as string;
}

async function openWorkspace(page: Page, organizationId: string) {
    await page.goto(`/organizations/${organizationId}`, { waitUntil: 'domcontentloaded' });
    await page.getByPlaceholder(/Tell .* what you want to create, review, or improve/i).waitFor({ timeout: 30_000 });
}

async function submitWorkspaceChat(page: Page, content: string) {
    const input = page.getByPlaceholder(/Tell .* what you want to create, review, or improve/i);
    await input.fill(content);
    const responsePromise = page.waitForResponse(
        (response) => response.url().includes('/api/v1/chat') && response.request().method() === 'POST',
        { timeout: 60_000 },
    );
    await input.press('Enter');
    const response = await responsePromise;
    return {
        response,
        body: (await response.json()) as ChatEnvelope,
    };
}

async function waitForConfirmAction(page: Page) {
    const responsePromise = page.waitForResponse(
        (response) => response.url().includes('/api/v1/intent/confirm-action') && response.request().method() === 'POST',
        { timeout: 60_000 },
    );
    await page.getByRole('button', { name: /Confirm & Execute/i }).click();
    const response = await responsePromise;
    return {
        response,
        body: await response.json(),
    };
}

async function responseText(response: { text(): Promise<string> }) {
	return response.text();
}

test.describe('Soma governed mutation live contract', () => {
    test.skip(!process.env.PLAYWRIGHT_LIVE_BACKEND, 'requires a live Core backend');

    test('Scenario A: direct Soma answer works in a fresh organization', async ({ page }) => {
        test.slow();
        const organizationId = await createOrganization(page, `QA Scenario A ${Date.now()}`);
        await openWorkspace(page, organizationId);

        const { response, body } = await submitWorkspaceChat(page, 'Summarize the current Workspace V8 design objectives.');

        expect(response.ok(), await responseText(response)).toBeTruthy();
        expect(body?.data?.mode).toBe('answer');
        expect((body?.data?.payload?.text ?? '').trim().length).toBeGreaterThan(0);
        await expect(page.getByText(/could not produce a readable reply/i)).toHaveCount(0);
    });

    test('Scenario B: mutation requests still route to proposal mode after prior answer-mode chat', async ({ page }) => {
        test.slow();
        const organizationId = await createOrganization(page, `QA Scenario B ${Date.now()}`);
        await openWorkspace(page, organizationId);

        const direct = await submitWorkspaceChat(page, 'Summarize the current Workspace V8 design objectives.');
        expect(direct.response.ok(), await responseText(direct.response)).toBeTruthy();
        expect(direct.body?.data?.mode).toBe('answer');

        const mutation = await submitWorkspaceChat(
            page,
            `Create a simple python file named workspace/logs/qa_browser_mixed_${Date.now()}.py that prints hello world.`,
        );

        expect(mutation.response.ok(), await responseText(mutation.response)).toBeTruthy();
        expect(mutation.body?.data?.mode).toBe('proposal');
        await expect(page.getByText('PROPOSED ACTION')).toBeVisible({ timeout: 30_000 });
        await expect(page.getByText(/Awaiting approval/i)).toBeVisible();
    });

    test('Scenario C+D: fresh mutation proposal stays side-effect free until confirm, and cancel remains safe + persistent', async ({ page }) => {
        test.slow();
        const stamp = Date.now();
        const targetFile = path.join(workspaceLogsDir, `qa_browser_cancel_${stamp}.py`);
        const organizationId = await createOrganization(page, `QA Scenario CD ${stamp}`);
        await openWorkspace(page, organizationId);

        const mutation = await submitWorkspaceChat(
            page,
            `Create a simple python file named workspace/logs/qa_browser_cancel_${stamp}.py that prints hello world.`,
        );

        expect(mutation.response.ok(), await responseText(mutation.response)).toBeTruthy();
        expect(mutation.body?.data?.mode).toBe('proposal');
        await expect(page.getByText('PROPOSED ACTION')).toBeVisible({ timeout: 30_000 });
        expect(fs.existsSync(targetFile)).toBeFalsy();

        await page.getByRole('button', { name: /^Cancel$/i }).click();
        await expect(page.getByText(/Proposal cancelled\. No action executed\./i)).toBeVisible({ timeout: 30_000 });
        expect(fs.existsSync(targetFile)).toBeFalsy();

        await page.reload({ waitUntil: 'domcontentloaded' });
        await page.getByPlaceholder(/Tell .* what you want to create, review, or improve/i).waitFor({ timeout: 30_000 });
        await expect(page.getByText(/Proposal cancelled\. No action executed\./i)).toBeVisible({ timeout: 30_000 });
        expect(fs.existsSync(targetFile)).toBeFalsy();
    });

    test('Scenario E: confirm yields durable proof, executes after approval, and persists on reload', async ({ page }) => {
        test.slow();
        const stamp = Date.now();
        const targetFile = path.join(workspaceLogsDir, `qa_browser_confirm_${stamp}.py`);
        const organizationId = await createOrganization(page, `QA Scenario E ${stamp}`);
        await openWorkspace(page, organizationId);

        const mutation = await submitWorkspaceChat(
            page,
            `Create a simple python file named workspace/logs/qa_browser_confirm_${stamp}.py that prints hello world.`,
        );

        expect(mutation.response.ok(), await responseText(mutation.response)).toBeTruthy();
        expect(mutation.body?.data?.mode).toBe('proposal');
        await expect(page.getByText('PROPOSED ACTION')).toBeVisible({ timeout: 30_000 });
        expect(fs.existsSync(targetFile)).toBeFalsy();

        const confirmed = await waitForConfirmAction(page);

        expect(confirmed.response.ok(), JSON.stringify(confirmed.body)).toBeTruthy();
        expect(typeof confirmed.body?.data?.run_id).toBe('string');
        expect((confirmed.body?.data?.run_id ?? '').trim().length).toBeGreaterThan(0);
        expect(confirmed.body?.data?.verified).toBeTruthy();
        expect(confirmed.body?.data?.execution_state).toBe('verified');

        await expect
            .poll(() => fs.existsSync(targetFile), {
                timeout: 30_000,
                message: `expected ${targetFile} to exist only after confirmation`,
            })
            .toBeTruthy();

        await expect(page.getByText(/Execution verified/i)).toBeVisible({ timeout: 30_000 });
        await expect(page.getByRole('link', { name: /Mission activated/i })).toBeVisible({ timeout: 30_000 });

        await page.reload({ waitUntil: 'domcontentloaded' });
        await page.getByPlaceholder(/Tell .* what you want to create, review, or improve/i).waitFor({ timeout: 30_000 });
        await expect(page.getByText(/Execution verified/i)).toBeVisible({ timeout: 30_000 });
        await expect(page.getByRole('link', { name: /Mission activated/i })).toBeVisible({ timeout: 30_000 });
    });
});
