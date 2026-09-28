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
