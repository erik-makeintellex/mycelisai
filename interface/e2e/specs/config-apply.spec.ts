import { test, expect, type Page, type Route } from '@playwright/test';
import {
    clickStartMode,
    mockOrganizationEntryApis,
    openOrganizationSetup,
} from '../support/organization-entry';

// RC-004G / RC-004G-2: governed clone (templates) and apply (templates/groups)
// config flows behind Advanced mode. Mirrors docs/API_REFERENCE.md#clone-and-apply-config.
// Reached through the real entry point: organization setup -> a starter card's
// "View config" action (Advanced mode only) -> the governed ConfigViewer.

test.skip(({ browserName }) => browserName !== 'chromium', 'Config apply coverage is Chromium-only for the MVP audit.');

async function fulfillJSON(route: Route, status: number, body: unknown) {
    await route.fulfill({ status, contentType: 'application/json', body: JSON.stringify(body) });
}

async function enableAdvancedMode(page: Page) {
    await page.addInitScript(() => {
        window.localStorage.setItem('mycelis-advanced-mode', 'true');
    });
}

const fileStarter = {
    id: 'starter-x',
    name: 'Engineering Starter',
    description: 'A file starter template.',
    organization_type: 'AI Organization starter',
    team_lead_label: 'Team Lead',
    advisor_count: 1,
    department_count: 1,
    specialist_count: 2,
    ai_engine_settings_summary: 'Local engines',
    response_contract_summary: 'Balanced',
    memory_personality_summary: 'Starter defaults',
    source: 'starter',
    available: true,
};

const customStarter = {
    id: 'custom-x',
    name: 'Acme Studio',
    description: 'An operator-cloned custom template.',
    organization_type: 'AI Organization starter',
    team_lead_label: 'Team Lead',
    advisor_count: 1,
    department_count: 1,
    specialist_count: 2,
    ai_engine_settings_summary: 'Local engines',
    response_contract_summary: 'Balanced',
    memory_personality_summary: 'Starter defaults',
    source: 'operator',
    available: true,
};

async function openTemplateStarters(page: Page) {
    const templatesLoaded = page.waitForResponse(
        (response) => response.url().includes('/api/v1/templates') && response.request().method() === 'GET',
    );
    await openOrganizationSetup(page);
    await templatesLoaded;
    await clickStartMode(page, /Start from template/i);
}

test.describe('Config clone and apply (organization setup, Advanced mode)', () => {
    test('clones a file template as a custom operator template', async ({ page }) => {
        await enableAdvancedMode(page);
        await mockOrganizationEntryApis(page, { templates: [fileStarter, customStarter] });
        await page.route('**/api/v1/config/templates/starter-x?*', (route) =>
            fulfillJSON(route, 200, {
                ok: true,
                data: {
                    kind: 'template', id: 'starter-x', format: 'yaml',
                    document: 'id: starter-x\nname: Engineering Starter\n',
                    source: 'file', source_path: 'templates/starters/starter-x.yaml',
                    redactions: [], generated_at: '2026-09-25T12:00:00Z',
                },
            }),
        );
        let cloneCalls = 0;
        await page.route('**/api/v1/config/templates/starter-x/clone', async (route) => {
            cloneCalls += 1;
            if (cloneCalls === 1) {
                return fulfillJSON(route, 202, {
                    ok: true,
                    data: {
                        requires_approval: true, blocked: false,
                        confirm_token: { token: 'tok-1', intent_proof_id: 'ip-1', expires_at: '2026-09-25T18:30:00Z' },
                        intent_proof: { id: 'ip-1', resolved_intent: 'template.clone.template.custom-acme@sha256:aaa', status: 'pending' },
                        audit_event_id: 'audit-1',
                        preview: { kind: 'template', id: 'custom-acme', document_sha256: 'aaa', valid: true, validation_errors: [], diff: { changed: [], added: ['id', 'name'], removed: [] } },
                    },
                });
            }
            return fulfillJSON(route, 201, {
                ok: true,
                data: { template: { id: 'custom-acme', name: 'Acme', source: 'operator', document_sha256: 'aaa' }, intent_proof_id: 'ip-1', audit_event_id: 'audit-1' },
            });
        });

        await page.goto('/dashboard', { waitUntil: 'domcontentloaded' });
        await openTemplateStarters(page);

        await page.getByTestId('starter-card-starter-x-view-config').click();
        await expect(page.getByTestId('config-viewer')).toBeVisible();
        await page.getByTestId('config-viewer-clone-template').click();
        await expect(page.getByTestId('clone-template-dialog')).toBeVisible();

        await page.getByTestId('clone-template-name').fill('Acme');
        await page.getByTestId('clone-template-preview').click();
        await expect(page.getByTestId('config-governance-preview-summary')).toBeVisible();
        await expect(page.getByTestId('config-governance-diff-added')).toContainText('id');

        await page.getByTestId('clone-template-approve').click();
        await expect(page.getByTestId('clone-template-success')).toContainText('custom-acme');
    });

    test('edits and applies a custom operator template with a preview diff', async ({ page }) => {
        await enableAdvancedMode(page);
        await mockOrganizationEntryApis(page, { templates: [fileStarter, customStarter] });
        await page.route('**/api/v1/config/templates/custom-x?*', (route) =>
            fulfillJSON(route, 200, {
                ok: true,
                data: {
                    kind: 'template', id: 'custom-x', format: 'yaml',
                    document: 'id: custom-x\nname: Acme Studio\n',
                    source: 'operator', redactions: [], generated_at: '2026-09-25T12:00:00Z',
                },
            }),
        );
        let applyCalls = 0;
        await page.route('**/api/v1/config/templates/custom-x/apply', async (route) => {
            applyCalls += 1;
            if (applyCalls === 1) {
                return fulfillJSON(route, 202, {
                    ok: true,
                    data: {
                        requires_approval: true, blocked: false,
                        confirm_token: { token: 'tok-2', intent_proof_id: 'ip-2', expires_at: '2026-09-25T18:30:00Z' },
                        intent_proof: { id: 'ip-2', resolved_intent: 'config.apply.template.custom-x@sha256:bbb', status: 'pending' },
                        audit_event_id: 'audit-2',
                        preview: { kind: 'template', id: 'custom-x', document_sha256: 'bbb', valid: true, validation_errors: [], diff: { changed: ['name'], added: [], removed: [] } },
                    },
                });
            }
            return fulfillJSON(route, 200, {
                ok: true,
                data: { applied: true, diff: { changed: ['name'], added: [], removed: [] }, intent_proof_id: 'ip-2', audit_event_id: 'audit-2' },
            });
        });

        await page.goto('/dashboard', { waitUntil: 'domcontentloaded' });
        await openTemplateStarters(page);

        await page.getByTestId('starter-card-custom-x-view-config').click();
        await expect(page.getByTestId('config-viewer')).toBeVisible();
        await page.getByTestId('config-viewer-edit-apply').click();
        await expect(page.getByTestId('apply-config-dialog')).toBeVisible();

        await page.getByTestId('apply-config-editor').fill('id: custom-x\nname: Acme Studio Global\n');
        await page.getByTestId('apply-config-preview').click();
        await expect(page.getByTestId('config-governance-diff-changed')).toContainText('name');

        await page.getByTestId('apply-config-approve').click();
        await expect(page.getByTestId('apply-config-success')).toBeVisible();
    });

    test('a blocked preview disables Approve', async ({ page }) => {
        await enableAdvancedMode(page);
        await mockOrganizationEntryApis(page, { templates: [fileStarter, customStarter] });
        await page.route('**/api/v1/config/templates/custom-x?*', (route) =>
            fulfillJSON(route, 200, {
                ok: true,
                data: {
                    kind: 'template', id: 'custom-x', format: 'yaml',
                    document: 'id: custom-x\nname: Acme Studio\n',
                    source: 'operator', redactions: [], generated_at: '2026-09-25T12:00:00Z',
                },
            }),
        );
        await page.route('**/api/v1/config/templates/custom-x/apply', (route) =>
            fulfillJSON(route, 202, {
                ok: true,
                data: {
                    requires_approval: true, blocked: true,
                    intent_proof: { id: 'ip-3', resolved_intent: 'config.apply.template.custom-x@sha256:ccc', status: 'pending' },
                    audit_event_id: 'audit-3',
                    preview: { kind: 'template', id: 'custom-x', document_sha256: 'ccc', valid: false, validation_errors: ['name: required'], diff: { changed: [], added: [], removed: [] } },
                },
            }),
        );

        await page.goto('/dashboard', { waitUntil: 'domcontentloaded' });
        await openTemplateStarters(page);

        await page.getByTestId('starter-card-custom-x-view-config').click();
        await page.getByTestId('config-viewer-edit-apply').click();

        await page.getByTestId('apply-config-editor').fill('id: custom-x\nname: {}\n');
        await page.getByTestId('apply-config-preview').click();
        await expect(page.getByTestId('config-governance-blocked')).toBeVisible();
        await expect(page.getByTestId('apply-config-approve')).toBeDisabled();
    });
});
