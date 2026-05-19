import { expect, test, type Page } from '@playwright/test';
import { openOrganizationWorkspace, waitForOrganizationWorkspaceReady } from '../support/live-organization-workspace';
import {
    anyTargetExists,
    createOrganization,
    removeExistingTargets,
    resolveBackendLogTargets,
    responseText,
    submitWorkspaceChat,
    waitForConfirmAction,
} from './soma-governance-live.helpers';

const LIVE_GOVERNANCE_TIMEOUT_MS = 180_000;

async function openWorkspace(page: Page, organizationId: string) {
    await openOrganizationWorkspace(page, organizationId);
}

test.describe('Soma governed mutation live contract', () => {
    test.skip(!process.env.PLAYWRIGHT_LIVE_BACKEND, 'requires a live Core backend');
    test.describe.configure({ timeout: LIVE_GOVERNANCE_TIMEOUT_MS, mode: 'serial' });

    test('Scenario A: direct Soma answer works in a fresh organization', async ({ page }) => {
        test.slow();
        const organizationId = await createOrganization(page, `QA Scenario A ${Date.now()}`);
        await openWorkspace(page, organizationId);

        const { response, body } = await submitWorkspaceChat(page, 'Summarize the current Workspace V8 design objectives.');

        expect(response.ok(), body ? JSON.stringify(body) : await responseText(response)).toBeTruthy();
        expect(body?.data?.mode).toBe('answer');
        expect(body?.data?.payload?.ask_class).toBe('direct_answer');
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
        expect(mutation.body?.data?.payload?.ask_class).toBe('governed_mutation');
        await expect(page.getByText('PROPOSED ACTION')).toBeVisible({ timeout: 30_000 });
        await expect(page.getByText(/Awaiting approval/i)).toBeVisible();
    });

    test('Scenario C+D: fresh mutation proposal stays side-effect free until confirm, and cancel remains safe + persistent', async ({ page }) => {
        test.slow();
        const stamp = Date.now();
        const targetPaths = resolveBackendLogTargets(`qa_browser_cancel_${stamp}.py`);
        const organizationId = await createOrganization(page, `QA Scenario CD ${stamp}`);
        await openWorkspace(page, organizationId);
        try {
            const mutation = await submitWorkspaceChat(
                page,
                `Create a simple python file named workspace/logs/qa_browser_cancel_${stamp}.py that prints hello world.`,
            );

            expect(mutation.response.ok(), mutation.body ? JSON.stringify(mutation.body) : mutation.raw).toBeTruthy();
            expect(mutation.body?.data?.mode).toBe('proposal');
            expect(mutation.body?.data?.payload?.ask_class).toBe('governed_mutation');
            await expect(page.getByText('PROPOSED ACTION')).toBeVisible({ timeout: 30_000 });
            expect(anyTargetExists(targetPaths)).toBeFalsy();

            await page.getByRole('button', { name: /^Cancel$/i }).click();
            await expect(page.getByText(/Proposal cancelled\. No action executed\./i)).toBeVisible({ timeout: 30_000 });
            expect(anyTargetExists(targetPaths)).toBeFalsy();

            await page.reload({ waitUntil: 'domcontentloaded' });
            await waitForOrganizationWorkspaceReady(page);
            await expect(page.getByText(/Proposal cancelled\. No action executed\./i)).toBeVisible({ timeout: 30_000 });
            expect(anyTargetExists(targetPaths)).toBeFalsy();
        } finally {
            removeExistingTargets(targetPaths);
        }
    });

    test('Scenario E: confirm yields durable proof, executes after approval, and persists on reload', async ({ page }) => {
        test.slow();
        const stamp = Date.now();
        const targetPaths = resolveBackendLogTargets(`qa_browser_confirm_${stamp}.py`);
        const organizationId = await createOrganization(page, `QA Scenario E ${stamp}`);
        await openWorkspace(page, organizationId);
        try {
            const mutation = await submitWorkspaceChat(
                page,
                `Create a simple python file named workspace/logs/qa_browser_confirm_${stamp}.py that prints hello world.`,
            );

            expect(mutation.response.ok(), mutation.body ? JSON.stringify(mutation.body) : mutation.raw).toBeTruthy();
            expect(mutation.body?.data?.mode).toBe('proposal');
            expect(mutation.body?.data?.payload?.ask_class).toBe('governed_mutation');
            await expect(page.getByText('PROPOSED ACTION')).toBeVisible({ timeout: 30_000 });
            expect(anyTargetExists(targetPaths)).toBeFalsy();

            const confirmed = await waitForConfirmAction(page);

            expect(confirmed.response.ok(), confirmed.body ? JSON.stringify(confirmed.body) : confirmed.raw).toBeTruthy();
            expect(typeof confirmed.body?.data?.run_id).toBe('string');
            expect((confirmed.body?.data?.run_id ?? '').trim().length).toBeGreaterThan(0);
            expect(confirmed.body?.data?.verified).toBeTruthy();
            expect(confirmed.body?.data?.execution_state).toBe('verified');

            await expect
                .poll(() => anyTargetExists(targetPaths), {
                    timeout: 30_000,
                    message: `expected one backend workspace target to exist after confirmation: ${targetPaths.join(', ')}`,
                })
                .toBeTruthy();

            await expect(page.getByText(/Execution verified/i)).toBeVisible({ timeout: 30_000 });
            await expect(page.getByRole('link', { name: /Mission activated/i })).toBeVisible({ timeout: 30_000 });

            await page.reload({ waitUntil: 'domcontentloaded' });
            await waitForOrganizationWorkspaceReady(page);
            await expect(page.getByText(/Execution verified/i)).toBeVisible({ timeout: 30_000 });
            await expect(page.getByRole('link', { name: /Mission activated/i })).toBeVisible({ timeout: 30_000 });
        } finally {
            removeExistingTargets(targetPaths);
        }
    });
});
