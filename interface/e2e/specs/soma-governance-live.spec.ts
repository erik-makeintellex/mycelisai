import fs from 'node:fs';
import path from 'node:path';
import { expect, test, type Page } from '@playwright/test';

function resolveBackendWorkspaceRoot() {
    // Live backend proof can run from a different checkout/worktree than the
    // Core server it is validating, so filesystem assertions need an explicit
    // override instead of assuming the backend workspace lives under this repo.
    const configuredRoot =
        process.env.PLAYWRIGHT_BACKEND_WORKSPACE_ROOT ?? process.env.MYCELIS_BACKEND_WORKSPACE_ROOT;
    if (configuredRoot && configuredRoot.trim().length > 0) {
        return path.resolve(configuredRoot);
    }
    return path.resolve(__dirname, '../../..', 'core', 'workspace');
}

const workspaceLogsDir = path.join(resolveBackendWorkspaceRoot(), 'workspace', 'logs');

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

type OrganizationEnvelope = {
    data?: {
        id?: string;
    };
};

type ConfirmEnvelope = {
    data?: {
        run_id?: string;
        verified?: boolean;
        execution_state?: string;
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
            purpose: 'Live governance verification',
            start_mode: 'empty',
        },
    });
    const parsed = await parseJSONIfPossible<OrganizationEnvelope>(response);
    expect(response.ok(), parsed.body ? JSON.stringify(parsed.body) : parsed.raw).toBeTruthy();
    const body = parsed.body;
    expect(body?.data?.id).toBeTruthy();
    return body?.data?.id as string;
}

async function openWorkspace(page: Page, organizationId: string) {
    await page.goto(`/organizations/${organizationId}`, { waitUntil: 'domcontentloaded' });
    await page.getByPlaceholder(/Tell .* what you want to create, review, or improve/i).waitFor({
        timeout: 30_000,
    });
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
    const parsed = await parseJSONIfPossible<ChatEnvelope>(response);
    return {
        response,
        raw: parsed.raw,
        body: parsed.body,
    };
}

async function waitForConfirmAction(page: Page) {
    const responsePromise = page.waitForResponse(
        (response) => response.url().includes('/api/v1/intent/confirm-action') && response.request().method() === 'POST',
        { timeout: 60_000 },
    );
    await page.getByRole('button', { name: /Approve & Execute|Execute/i }).click();
    const response = await responsePromise;
    const parsed = await parseJSONIfPossible<ConfirmEnvelope>(response);
    return {
        response,
        raw: parsed.raw,
        body: parsed.body,
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

        expect(response.ok(), body ? JSON.stringify(body) : await responseText(response)).toBeTruthy();
        expect(body?.data?.mode).toBe('answer');
        expect((body?.data?.payload?.text ?? '').trim().length).toBeGreaterThan(0);
        await expect(page.getByText(/could not produce a readable reply/i)).toHaveCount(0);
    });

    test('Scenario B: mutation requests still route to proposal mode after prior answer-mode chat', async ({ page }) => {
        test.slow();
        const organizationId = await createOrganization(page, `QA Scenario B ${Date.now()}`);
        await openWorkspace(page, organizationId);

        const direct = await submitWorkspaceChat(page, 'Summarize the current Workspace V8 design objectives.');
        expect(direct.response.ok(), direct.body ? JSON.stringify(direct.body) : direct.raw).toBeTruthy();
        expect(direct.body?.data?.mode).toBe('answer');

        const mutation = await submitWorkspaceChat(
            page,
            `Create a simple python file named workspace/logs/qa_browser_mixed_${Date.now()}.py that prints hello world.`,
        );

        expect(mutation.response.ok(), mutation.body ? JSON.stringify(mutation.body) : mutation.raw).toBeTruthy();
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

        expect(mutation.response.ok(), mutation.body ? JSON.stringify(mutation.body) : mutation.raw).toBeTruthy();
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

        expect(mutation.response.ok(), mutation.body ? JSON.stringify(mutation.body) : mutation.raw).toBeTruthy();
        expect(mutation.body?.data?.mode).toBe('proposal');
        await expect(page.getByText('PROPOSED ACTION')).toBeVisible({ timeout: 30_000 });
        expect(fs.existsSync(targetFile)).toBeFalsy();

        const confirmed = await waitForConfirmAction(page);

        expect(confirmed.response.ok(), confirmed.body ? JSON.stringify(confirmed.body) : confirmed.raw).toBeTruthy();
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
