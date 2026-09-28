import { describe, expect, it } from 'vitest';
import { blockerCopy, resolveBlockerCopy } from '@/lib/blockerCopy';

const DECK_CODES = [
    'approver_required',
    'confirmer_not_proposer',
    'token_already_used',
    'token_purpose_unknown',
    'token_wrong_purpose',
    'outcome_template_unresolved',
    'blueprint_mismatch',
    'confirm_required',
    'invalid_confirm_token',
    'governance_policy_unavailable',
    'provider_bound',
    'model_missing',
    'service_unavailable',
    'transport_unavailable',
    'provider_timeout',
    'empty_provider_output',
    'connector_deployment_unavailable',
    'token_budget_exhausted',
    'token_budget_usage_forbidden',
    'mcp_tool_not_found',
    'mcp_call_forbidden',
    'mcp_env_rejected',
    'mcp_server_not_found',
    'mcp_connect_failed',
    'memory_entry_archived',
    'memory_entry_changed',
];

describe('blockerCopy', () => {
    it.each(DECK_CODES)('maps deck code %s to plain-language copy with no dead end', (code) => {
        const copy = blockerCopy({ code, viewerIsAdmin: false });
        expect(copy.code).toBe(code);
        expect(copy.title.length).toBeGreaterThan(0);
        expect(copy.title.split(' ').length).toBeLessThanOrEqual(8);
        expect(copy.whatHappened.length).toBeGreaterThan(0);
        expect(copy.nextAction.label.length).toBeGreaterThan(0);
        // Never leak raw backend vocabulary.
        expect(copy.title).not.toMatch(/token|confirm_token|swarm|mission|degraded/i);
        expect(copy.whatHappened).not.toMatch(/confirm_token|api\/|MYCELIS_/i);
    });

    it('maps admin_required with the generic default reason', () => {
        const copy = blockerCopy({ code: 'admin_required', httpStatus: 403, viewerIsAdmin: false });
        expect(copy.code).toBe('admin_required');
        expect(copy.title).toBe('This area is for admins');
        expect(copy.nextAction.label).toBeTruthy();
        expect(copy.whoCanHelp).toMatch(/admin/i);
    });

    it.each(['approvals', 'settings', 'groups', 'automation_profile'])(
        'maps admin_required with reason=%s to a distinct, non-dead-end variant',
        (reason) => {
            const copy = blockerCopy({ code: 'admin_required', httpStatus: 403, viewerIsAdmin: false, reason });
            expect(copy.code).toBe('admin_required');
            expect(copy.nextAction.label).toBeTruthy();
            expect(copy.nextAction.href).toBeTruthy();
        },
    );

    it.each(['output_readback_missing', 'output_readback_mismatch', 'output_missing'])(
        'maps the output_* prefix %s to a shared "Couldn\'t confirm the file" copy with a next action',
        (code) => {
            const copy = blockerCopy({ code, viewerIsAdmin: false });
            expect(copy.code).toBe(code);
            expect(copy.title).toBe("Couldn't confirm the file");
            expect(copy.nextAction.label).toBe('Open the file');
            expect(copy.whoCanHelp).toMatch(/ask soma to redo/i);
        },
    );

    it('falls back to generic request_failed copy for an unknown code, echoing the original code', () => {
        const copy = blockerCopy({ code: 'some_new_backend_code', viewerIsAdmin: false });
        expect(copy.code).toBe('some_new_backend_code');
        expect(copy.title).toBe('Something went wrong');
        expect(copy.nextAction.label).toBe('Try again');
    });

    it('falls back the same way for an admin viewer', () => {
        const copy = blockerCopy({ code: 'totally_unknown', viewerIsAdmin: true });
        expect(copy.code).toBe('totally_unknown');
        expect(copy.title).toBe('Something went wrong');
    });

    it('never returns Retry for token_already_used', () => {
        const copy = blockerCopy({ code: 'token_already_used', viewerIsAdmin: false });
        expect(copy.title).toBe('Already approved');
        expect(copy.nextAction.intent).not.toBe('retry');
        expect(copy.nextAction.label.toLowerCase()).not.toContain('retry');
    });

    it('keeps token-kept codes free of a false "All Clear" or discard tone', () => {
        for (const code of ['approver_required', 'confirmer_not_proposer', 'token_purpose_unknown', 'token_wrong_purpose']) {
            const copy = blockerCopy({ code, viewerIsAdmin: false });
            expect(copy.whatHappened.toLowerCase()).not.toContain('all clear');
        }
    });

    it('maps service_unavailable to the activity-log copy with an admin next step', () => {
        const copy = blockerCopy({ code: 'service_unavailable', httpStatus: 503, viewerIsAdmin: false });
        expect(copy.whatHappened).toContain('activity log is unavailable');
        expect(copy.adminVariant?.whatHappened).toContain('Restore it');
    });

    it('maps admin_required with reason=mcp_call to a Soma-proposal next step, not the generic default', () => {
        const copy = blockerCopy({ code: 'admin_required', httpStatus: 403, viewerIsAdmin: false, reason: 'mcp_call' });
        expect(copy.code).toBe('admin_required');
        expect(copy.title).not.toBe('This area is for admins');
        expect(copy.whatHappened).toMatch(/use this tool directly/i);
        expect(copy.nextAction.label).toMatch(/ask soma/i);
        const admin = resolveBlockerCopy(copy, true);
        expect(admin.whatHappened).toMatch(/missing a permission/i);
    });

    it('maps service_unavailable with reason=mcp_call to the MCP connection copy, not the activity-log copy', () => {
        const copy = blockerCopy({ code: 'service_unavailable', httpStatus: 503, viewerIsAdmin: false, reason: 'mcp_call' });
        expect(copy.code).toBe('service_unavailable');
        expect(copy.whatHappened).toMatch(/connection is offline/i);
        expect(copy.whatHappened).not.toContain('activity log is unavailable');
        const admin = resolveBlockerCopy(copy, true);
        expect(admin.nextAction.href).toBe('/resources?tab=tools');
    });

    it('maps mcp_tool_not_found to a plain-language copy with a retry and an admin detail', () => {
        const copy = blockerCopy({ code: 'mcp_tool_not_found', httpStatus: 404, viewerIsAdmin: false });
        expect(copy.whatHappened).toMatch(/isn't available/i);
        expect(copy.nextAction.intent).toBe('retry');
        const admin = resolveBlockerCopy(copy, true);
        expect(admin.whatHappened).toMatch(/exact name/i);
    });

    it('maps mcp_call_forbidden to a plain-language copy naming Soma or an admin', () => {
        const copy = blockerCopy({ code: 'mcp_call_forbidden', httpStatus: 403, viewerIsAdmin: false });
        expect(copy.whatHappened).toMatch(/can't use this tool directly/i);
        expect(copy.whoCanHelp).toMatch(/admin/i);
        const admin = resolveBlockerCopy(copy, true);
        expect(admin.whatHappened).toMatch(/outputs:read/i);
    });

    it('maps mcp_env_rejected to a plain-language copy naming the settings problem', () => {
        const copy = blockerCopy({ code: 'mcp_env_rejected', httpStatus: 400, viewerIsAdmin: false });
        expect(copy.whatHappened).toMatch(/can't be set up with those settings/i);
        const admin = resolveBlockerCopy(copy, true);
        expect(admin.whatHappened).toMatch(/environment variables/i);
    });

    // MCPA D10: delete resolves the server before auditing (D9); an
    // unknown id is 404 mcp_server_not_found with no admin-only detail.
    it('maps mcp_server_not_found to a plain-language copy with a retry', () => {
        const copy = blockerCopy({ code: 'mcp_server_not_found', httpStatus: 404, viewerIsAdmin: false });
        expect(copy.whatHappened).toMatch(/removed/i);
        expect(copy.nextAction.intent).toBe('retry');
    });

    // MCPA: install/apply's Connect failed after the server row was
    // created; the code mirrors codeMCPConnectFailed and is never reported
    // as a success.
    it('maps mcp_connect_failed to a plain-language copy with an admin detail path', () => {
        const copy = blockerCopy({ code: 'mcp_connect_failed', httpStatus: 502, viewerIsAdmin: false });
        expect(copy.whatHappened).toMatch(/could not connect/i);
        expect(copy.whoCanHelp).toMatch(/admin/i);
        const admin = resolveBlockerCopy(copy, true);
        expect(admin.whatHappened).toMatch(/Connect failed after install/i);
        expect(admin.nextAction.href).toBe('/resources?tab=tools');
    });

    // MEM W2: the entry moved (edited or archived elsewhere) or was
    // archived since the editor opened it (409s from PATCH
    // /api/v1/memory/deployment-context/{id}).
    it('maps memory_entry_archived to a restore-first copy with a dismiss action, no retry dead end', () => {
        const copy = blockerCopy({ code: 'memory_entry_archived', httpStatus: 409, viewerIsAdmin: false });
        expect(copy.whatHappened).toMatch(/archived/i);
        expect(copy.whatHappened).toMatch(/nothing was saved/i);
        expect(copy.nextAction.intent).toBe('dismiss');
    });

    it('maps memory_entry_changed to a reload copy, distinct from memory_entry_archived', () => {
        const copy = blockerCopy({ code: 'memory_entry_changed', httpStatus: 409, viewerIsAdmin: false });
        expect(copy.whatHappened).toMatch(/changed/i);
        expect(copy.nextAction.intent).toBe('retry');
        expect(copy.nextAction.label).toBe('Reload');
    });

    it.each(['provider_timeout', 'empty_provider_output'])(
        'maps %s to "Soma couldn\'t draft that file" with a try-again next action',
        (code) => {
            const copy = blockerCopy({ code, viewerIsAdmin: false });
            expect(copy.title).toBe("Soma couldn't draft that file");
            expect(copy.nextAction.label).toBe('Try again');
            expect(copy.whatHappened.toLowerCase()).toContain('add the content yourself');
        },
    );

    it('maps connector_deployment_unavailable to a plain-language copy with a next action', () => {
        const copy = blockerCopy({ code: 'connector_deployment_unavailable', httpStatus: 501, viewerIsAdmin: false });
        expect(copy.title).toBe("Connector installs aren't available yet");
        expect(copy.nextAction.label).toBeTruthy();
        expect(copy.nextAction.href).toBeTruthy();
        expect(copy.whatHappened).not.toMatch(/501|api\//i);
    });

    it('maps token_budget_exhausted with the exact D5 user copy and an admin override next step', () => {
        const user = blockerCopy({ code: 'token_budget_exhausted', httpStatus: 429, viewerIsAdmin: false });
        expect(user.whatHappened).toBe('This work stopped because it reached its token budget.');
        expect(resolveBlockerCopy(user, false).nextAction.intent).not.toBe('retry');

        const admin = resolveBlockerCopy(user, true);
        expect(admin.nextAction.href).toBe('/settings?tab=engines');
        expect(admin.nextAction.label.toLowerCase()).toMatch(/budget|override/);
    });

    it('maps token_budget_usage_forbidden to a member-only explanation with no dead end', () => {
        const copy = blockerCopy({ code: 'token_budget_usage_forbidden', httpStatus: 403, viewerIsAdmin: false });
        expect(copy.whatHappened.toLowerCase()).toContain('team');
        expect(copy.nextAction.label.length).toBeGreaterThan(0);
    });

    it('maps transport_unavailable to the team-service copy, not a generic/auth message', () => {
        const copy = blockerCopy({ code: 'transport_unavailable', httpStatus: 503, viewerIsAdmin: false });
        expect(copy.title).toBe("Soma's team service isn't running");
        expect(copy.whatHappened).not.toMatch(/unauthorized|not authorized/i);
    });

    it('maps the legacy alias policy_unavailable onto governance_policy_unavailable (the real backend code)', () => {
        const aliased = blockerCopy({ code: 'policy_unavailable', viewerIsAdmin: false });
        const canonical = blockerCopy({ code: 'governance_policy_unavailable', viewerIsAdmin: false });
        expect(aliased.title).toBe(canonical.title);
        expect(aliased.whatHappened).toBe(canonical.whatHappened);
        expect(aliased.code).toBe('governance_policy_unavailable');
    });

    describe('resolveBlockerCopy', () => {
        it('resolves the admin variant when the viewer is an admin and one exists', () => {
            const copy = blockerCopy({ code: 'governance_policy_unavailable', viewerIsAdmin: false });
            expect(copy.adminVariant).toBeDefined();
            const resolved = resolveBlockerCopy(copy, true);
            expect(resolved.title).toBe("Safety rules didn't load");
            expect((resolved as { adminVariant?: unknown }).adminVariant).toBeUndefined();
        });

        it('keeps the base copy for a non-admin viewer', () => {
            const copy = blockerCopy({ code: 'governance_policy_unavailable', viewerIsAdmin: false });
            const resolved = resolveBlockerCopy(copy, false);
            expect(resolved.title).toBe('Safety lock is on');
        });

        it('falls back to the base copy when no admin variant exists', () => {
            const copy = blockerCopy({ code: 'token_already_used', viewerIsAdmin: false });
            const resolved = resolveBlockerCopy(copy, true);
            expect(resolved.title).toBe('Already approved');
        });
    });
});
