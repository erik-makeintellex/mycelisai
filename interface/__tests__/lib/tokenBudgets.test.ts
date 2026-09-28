import { beforeEach, describe, expect, it } from 'vitest';
import { mockFetch } from '../setup';
import {
    deleteTokenBudgetOverride,
    getTokenBudgets,
    getTokenBudgetUsage,
    putTokenBudgetOverride,
    validateOverrideDraft,
    type TokenBudgetsResponse,
} from '@/lib/tokenBudgets';

const jsonResponse = (body: unknown, ok = true, status = ok ? 200 : 400) => ({
    ok,
    status,
    json: async () => body,
});

const BUDGETS: TokenBudgetsResponse = {
    default_class: 'local_large',
    class_order: ['local_large', 'local_small', 'hosted_standard', 'hosted_premium'],
    global: { per_execution: 32000, per_run: 128000, per_team_day: 500000, per_agent_day: 250000, warn_pct: 80 },
    classes: {
        local_large: { per_execution: 64000, per_run: 256000, per_team_day: 2000000, per_agent_day: 1000000, warn_pct: 80 },
    },
    day_period: 'utc_day',
    min_limit: 1024,
    max_limit: 5000000,
    can_edit: false,
};

describe('tokenBudgets client', () => {
    beforeEach(() => {
        mockFetch.mockReset();
    });

    describe('getTokenBudgets', () => {
        it('returns the effective policy on 200', async () => {
            mockFetch.mockResolvedValueOnce(jsonResponse({ ok: true, data: BUDGETS }));
            const result = await getTokenBudgets();
            expect(result.ok).toBe(true);
            if (result.ok) {
                expect(result.data.classes.local_large.per_execution).toBe(64000);
                expect(result.data.providers).toBeUndefined();
            }
            expect(mockFetch).toHaveBeenCalledWith('/api/v1/cognitive/budgets', { cache: 'no-store' });
        });

        it('handles providers being absent (non-root-admin caller)', async () => {
            mockFetch.mockResolvedValueOnce(jsonResponse({ ok: true, data: BUDGETS }));
            const result = await getTokenBudgets();
            expect(result.ok && result.data.providers).toBeUndefined();
        });

        it('surfaces a normalized code on 401', async () => {
            mockFetch.mockResolvedValueOnce(jsonResponse({ ok: false, error: 'Authentication required', data: { code: 'request_failed' } }, false, 401));
            const result = await getTokenBudgets();
            expect(result.ok).toBe(false);
            if (!result.ok) {
                expect(result.httpStatus).toBe(401);
                expect(result.code).toBe('request_failed');
            }
        });

        it('never throws on a network failure; returns request_failed', async () => {
            mockFetch.mockRejectedValueOnce(new Error('network down'));
            const result = await getTokenBudgets();
            expect(result.ok).toBe(false);
            if (!result.ok) expect(result.code).toBe('request_failed');
        });
    });

    describe('getTokenBudgetUsage', () => {
        it('queries by team_id and returns usage', async () => {
            mockFetch.mockResolvedValueOnce(jsonResponse({
                ok: true,
                data: { scope: 'team_day', ref: 't1', used: 12000, limit: 64000, remaining: 52000, warn: false, warn_pct: 80, period: 'utc_day', usage_reported: true },
            }));
            const result = await getTokenBudgetUsage({ team_id: 't1' });
            expect(mockFetch).toHaveBeenCalledWith('/api/v1/cognitive/budgets/usage?team_id=t1', { cache: 'no-store' });
            expect(result.ok).toBe(true);
            if (result.ok) expect(result.data.used).toBe(12000);
        });

        it('maps a 403 to token_budget_usage_forbidden', async () => {
            mockFetch.mockResolvedValueOnce(jsonResponse({ ok: false, error: 'forbidden', data: { code: 'token_budget_usage_forbidden', scope: 'team_day' } }, false, 403));
            const result = await getTokenBudgetUsage({ team_id: 't1' });
            expect(result.ok).toBe(false);
            if (!result.ok) {
                expect(result.httpStatus).toBe(403);
                expect(result.code).toBe('token_budget_usage_forbidden');
            }
        });
    });

    describe('putTokenBudgetOverride / deleteTokenBudgetOverride', () => {
        it('PUTs a validated body to the level/ref path', async () => {
            mockFetch.mockResolvedValueOnce(jsonResponse({
                ok: true,
                data: { level: 'team', ref: 't1', changed: true, effective: { limits: {}, sources: {} }, audit_event_id: 'a1' },
            }));
            const result = await putTokenBudgetOverride('team', 't1', { per_execution: 2000 });
            expect(mockFetch).toHaveBeenCalledWith('/api/v1/cognitive/budgets/overrides/team/t1', {
                method: 'PUT',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify({ per_execution: 2000 }),
            });
            expect(result.ok).toBe(true);
        });

        it('DELETEs the override path and is idempotent', async () => {
            mockFetch.mockResolvedValueOnce(jsonResponse({
                ok: true,
                data: { level: 'team', ref: 't1', changed: false, effective: { limits: {}, sources: {} }, audit_event_id: 'a2' },
            }));
            const result = await deleteTokenBudgetOverride('team', 't1');
            expect(mockFetch).toHaveBeenCalledWith('/api/v1/cognitive/budgets/overrides/team/t1', { method: 'DELETE' });
            expect(result.ok).toBe(true);
        });

        it('surfaces a 403 admin_required without throwing', async () => {
            mockFetch.mockResolvedValueOnce(jsonResponse({ ok: false, error: 'forbidden', data: { code: 'admin_required' } }, false, 403));
            const result = await putTokenBudgetOverride('team', 't1', { per_execution: 2000 });
            expect(result.ok).toBe(false);
            if (!result.ok) expect(result.code).toBe('admin_required');
        });
    });

    describe('validateOverrideDraft (mirrors server bounds)', () => {
        const bounds = { min_limit: 1024, max_limit: 5000000 };

        it('rejects a value below min_limit', () => {
            expect(validateOverrideDraft({ per_execution: 100 }, bounds)).toMatch(/1,?024|min/i);
        });

        it('rejects a value above max_limit', () => {
            expect(validateOverrideDraft({ per_execution: 9000000 }, bounds)).toMatch(/5,?000,?000|max/i);
        });

        it('rejects per_execution > per_run', () => {
            expect(validateOverrideDraft({ per_execution: 5000, per_run: 4000 }, bounds)).toMatch(/per_execution.*per_run|order/i);
        });

        it('rejects per_run > per_team_day', () => {
            expect(validateOverrideDraft({ per_run: 5000, per_team_day: 4000 }, bounds)).toMatch(/per_run.*per_team_day|order/i);
        });

        it('rejects an empty draft (no fields set)', () => {
            expect(validateOverrideDraft({}, bounds)).toBeTruthy();
        });

        it('rejects warn_pct outside 1-99', () => {
            expect(validateOverrideDraft({ warn_pct: 0 }, bounds)).toBeTruthy();
            expect(validateOverrideDraft({ warn_pct: 100 }, bounds)).toBeTruthy();
        });

        it('accepts a valid, ordered draft', () => {
            expect(validateOverrideDraft({ per_execution: 2000, per_run: 8000, per_team_day: 16000, warn_pct: 80 }, bounds)).toBeNull();
        });
    });
});
