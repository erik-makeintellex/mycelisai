import { test, expect, type Page, type Route } from '@playwright/test';

// RC-004G: governed clone (templates) and apply (templates/groups) config
// flows behind Advanced mode. Mirrors docs/API_REFERENCE.md#clone-and-apply-config.
// Mocked end to end; a "template" kind deployment record is synthesized here
// (Deployments today only surfaces team/group/organization config refs) to
// exercise the "Clone as custom template" entry point through the real page.

test.skip(({ browserName }) => browserName !== 'chromium', 'Config apply coverage is Chromium-only for the MVP audit.');

async function fulfillJSON(route: Route, status: number, body: unknown) {
    await route.fulfill({ status, contentType: 'application/json', body: JSON.stringify(body) });
}

async function enableAdvancedMode(page: Page) {
    await page.addInitScript(() => {
        window.localStorage.setItem('mycelis-advanced-mode', 'true');
    });
}

const groupRecord = {
    id: 'group-revops', kind: 'group', name: 'Revenue Ops', purpose: 'Keep revenue operations visible.',
    what_runs: 'Standing collaboration group for revenue operations.',
    scope: { kind: 'global' }, origin: 'operator', lifecycle_state: 'running',
    since: '2026-06-01T09:00:00Z', expires_at: null, last_activity: '2026-09-23T11:00:00Z',
    members: [], tools: [], work_mode: 'propose_only', approval_policy_ref: 'standard',
    approval_posture: 'Proposals require operator approval before executing.',
    outputs_summary: { retained: 0, by_status: {}, proof_status: 'not_recorded', last_output_at: null },
    config_ref: { kind: 'group', id: 'group-revops', path: '/api/v1/config/groups/group-revops' },
};

const templateAsDeployment = {
    id: 'starter-x', kind: 'standing_team', name: 'Engineering Starter', purpose: 'A file starter template.',
    what_runs: 'Starter template surfaced here to exercise Clone as custom template.',
    scope: { kind: 'global' }, origin: 'bundle', lifecycle_state: 'running',
    since: '2026-06-01T09:00:00Z', expires_at: null, last_activity: '2026-09-23T11:00:00Z',
    members: [], tools: [], work_mode: undefined, approval_policy_ref: undefined,
    approval_posture: undefined, outputs_summary: null,
    config_ref: { kind: 'template', id: 'starter-x', path: '/api/v1/config/templates/starter-x' },
};

async function mockDeployments(page: Page, records: unknown[]) {
    await page.route(/\/api\/v1\/deployments(?:\?.*)?$/, async (route) => {
        if (route.request().method() !== 'GET') return route.fallback();
        await fulfillJSON(route, 200, {
            ok: true,
            data: { deployments: records, sources: [], generated_at: '2026-09-25T12:00:00Z' },
        });
    });
}

test.describe('Config clone and apply (/deployments, Advanced mode)', () => {
    test('clones a file template as a custom operator template', async ({ page }) => {
        await enableAdvancedMode(page);
        await mockDeployments(page, [templateAsDeployment]);
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

        await page.goto('/deployments', { waitUntil: 'domcontentloaded' });
        await page.getByTestId('deployments-list-item-standing_team:starter-x').click();
        await page.getByTestId('deployments-view-config').click();
        await page.getByTestId('config-viewer-clone-template').click();
        await expect(page.getByTestId('clone-template-dialog')).toBeVisible();

        await page.getByTestId('clone-template-name').fill('Acme');
        await page.getByTestId('clone-template-preview').click();
        await expect(page.getByTestId('config-governance-preview-summary')).toBeVisible();
        await expect(page.getByTestId('config-governance-diff-added')).toContainText('id');

        await page.getByTestId('clone-template-approve').click();
        await expect(page.getByTestId('clone-template-success')).toContainText('custom-acme');
    });

    test('edits and applies a group config with a preview diff', async ({ page }) => {
        await enableAdvancedMode(page);
        await mockDeployments(page, [groupRecord]);
        await page.route('**/api/v1/config/groups/group-revops?*', (route) =>
            fulfillJSON(route, 200, {
                ok: true,
                data: {
                    kind: 'group', id: 'group-revops', format: 'yaml',
                    document: 'id: group-revops\nspec:\n  name: Revenue Ops\n',
                    source: 'database', redactions: [], generated_at: '2026-09-25T12:00:00Z',
                },
            }),
        );
        let applyCalls = 0;
        await page.route('**/api/v1/config/groups/group-revops/apply', async (route) => {
            applyCalls += 1;
            if (applyCalls === 1) {
                return fulfillJSON(route, 202, {
                    ok: true,
                    data: {
                        requires_approval: true, blocked: false,
                        confirm_token: { token: 'tok-2', intent_proof_id: 'ip-2', expires_at: '2026-09-25T18:30:00Z' },
                        intent_proof: { id: 'ip-2', resolved_intent: 'config.apply.group.group-revops@sha256:bbb', status: 'pending' },
                        audit_event_id: 'audit-2',
                        preview: { kind: 'group', id: 'group-revops', document_sha256: 'bbb', valid: true, validation_errors: [], diff: { changed: ['spec.name'], added: [], removed: [] } },
                    },
                });
            }
            return fulfillJSON(route, 200, {
                ok: true,
                data: { applied: true, diff: { changed: ['spec.name'], added: [], removed: [] }, intent_proof_id: 'ip-2', audit_event_id: 'audit-2' },
            });
        });

        await page.goto('/deployments', { waitUntil: 'domcontentloaded' });
        await page.getByTestId('deployments-list-item-group:group-revops').click();
        await page.getByTestId('deployments-view-config').click();
        await page.getByTestId('config-viewer-edit-apply').click();
        await expect(page.getByTestId('apply-config-dialog')).toBeVisible();

        await page.getByTestId('apply-config-editor').fill('id: group-revops\nspec:\n  name: Global Revenue Ops\n');
        await page.getByTestId('apply-config-preview').click();
        await expect(page.getByTestId('config-governance-diff-changed')).toContainText('spec.name');

        await page.getByTestId('apply-config-approve').click();
        await expect(page.getByTestId('apply-config-success')).toBeVisible();
    });

    test('a blocked preview disables Approve', async ({ page }) => {
        await enableAdvancedMode(page);
        await mockDeployments(page, [groupRecord]);
        await page.route('**/api/v1/config/groups/group-revops?*', (route) =>
            fulfillJSON(route, 200, {
                ok: true,
                data: {
                    kind: 'group', id: 'group-revops', format: 'yaml',
                    document: 'id: group-revops\nspec:\n  name: Revenue Ops\n',
                    source: 'database', redactions: [], generated_at: '2026-09-25T12:00:00Z',
                },
            }),
        );
        await page.route('**/api/v1/config/groups/group-revops/apply', (route) =>
            fulfillJSON(route, 202, {
                ok: true,
                data: {
                    requires_approval: true, blocked: true,
                    intent_proof: { id: 'ip-3', resolved_intent: 'config.apply.group.group-revops@sha256:ccc', status: 'pending' },
                    audit_event_id: 'audit-3',
                    preview: { kind: 'group', id: 'group-revops', document_sha256: 'ccc', valid: false, validation_errors: ['spec.name: required'], diff: { changed: [], added: [], removed: [] } },
                },
            }),
        );

        await page.goto('/deployments', { waitUntil: 'domcontentloaded' });
        await page.getByTestId('deployments-list-item-group:group-revops').click();
        await page.getByTestId('deployments-view-config').click();
        await page.getByTestId('config-viewer-edit-apply').click();

        await page.getByTestId('apply-config-editor').fill('id: group-revops\nspec: {}\n');
        await page.getByTestId('apply-config-preview').click();
        await expect(page.getByTestId('config-governance-blocked')).toBeVisible();
        await expect(page.getByTestId('apply-config-approve')).toBeDisabled();
    });
});
