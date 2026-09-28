// Typed client for the B1 token-budgets API. See
// docs/architecture-library/B1_TOKEN_BUDGETS_CONTRACT.md (D8) and
// docs/API_REFERENCE.md (budgets rows) for the frozen shape.
//
// Every call goes through the Next.js route (the BFF/proxy already forwards
// /api/v1/* to Core with the session attached); this file never talks to
// Core directly and never decides authority — it only shapes requests and
// normalizes responses so callers get a typed result instead of raw fetch
// plumbing. A network failure or a non-JSON body is folded into
// `request_failed` rather than thrown, matching the rest of `lib/`.

export type TokenBudgetClass = 'local_large' | 'local_small' | 'hosted_standard' | 'hosted_premium';

export interface TokenBudgetLimits {
    per_execution?: number;
    per_run?: number;
    per_team_day?: number;
    per_agent_day?: number;
    warn_pct?: number;
}

export type TokenBudgetOverrideLevel = 'agent' | 'team' | 'profile' | 'class';

export interface TokenBudgetProvider {
    provider_id: string;
    model_id: string;
    budget_class: string;
    class_source: 'explicit' | 'data_boundary';
}

export interface TokenBudgetOverrides {
    agent?: Record<string, TokenBudgetLimits>;
    team?: Record<string, TokenBudgetLimits>;
    profile?: Record<string, TokenBudgetLimits>;
    class?: Record<string, TokenBudgetLimits>;
}

export interface TokenBudgetsResponse {
    default_class?: string;
    class_order: string[];
    global: TokenBudgetLimits;
    classes: Record<string, TokenBudgetLimits>;
    day_period: 'utc_day' | 'since_restart';
    min_limit: number;
    max_limit: number;
    can_edit: boolean;
    // Only present for a root admin with cognitive:read or cognitive:write.
    providers?: TokenBudgetProvider[];
    // Only present for an admin viewer.
    overrides?: TokenBudgetOverrides;
}

export type TokenBudgetUsageScope = 'team_day' | 'agent_day' | 'run';
export type TokenBudgetUsagePeriod = 'utc_day' | 'run' | 'since_restart';

export interface TokenBudgetUsage {
    scope: TokenBudgetUsageScope;
    ref: string;
    used: number;
    limit: number;
    remaining: number;
    warn: boolean;
    warn_pct: number;
    period: TokenBudgetUsagePeriod;
    resets_at?: string;
    // false means some calls were charged their output reservation, not a
    // provider-reported count; the UI must show "at least N".
    usage_reported: boolean;
    // B1R-A (parallel slice, merged to dev as 1e86f9d1 after this branch was
    // cut from feature/b1-route-scope): the output allowance held by
    // in-flight calls. remaining = limit - used - reserved. Optional because
    // it is not part of this branch's frozen base contract; treat it as
    // possibly absent or 0.
    reserved?: number;
}

export type TokenBudgetUsageQuery = { team_id: string } | { agent_id: string } | { run_id: string };

export interface TokenBudgetOverrideResult {
    level: TokenBudgetOverrideLevel;
    ref: string;
    changed: boolean;
    before?: TokenBudgetLimits;
    after?: TokenBudgetLimits;
    effective: { limits: TokenBudgetLimits; sources: Record<string, string> };
    audit_event_id: string;
    record_id?: string;
}

export interface TokenBudgetOverrideBody {
    per_execution?: number;
    per_run?: number;
    per_team_day?: number;
    per_agent_day?: number;
    warn_pct?: number;
}

/** Normalized failure: a blocker `code` plus the HTTP status, ready for
 * `blockerCopy()`. Never a raw Error — network failures fold into this too. */
export interface TokenBudgetErrorResult {
    ok: false;
    code: string;
    httpStatus: number;
    data?: Record<string, unknown>;
}

export type TokenBudgetResult<T> = { ok: true; data: T } | TokenBudgetErrorResult;

async function callTokenBudgetAPI<T>(url: string, init?: RequestInit): Promise<TokenBudgetResult<T>> {
    try {
        const res = await fetch(url, init);
        const body = await res.json().catch(() => null);
        if (!res.ok || !body?.ok) {
            const data = body?.data && typeof body.data === 'object' ? (body.data as Record<string, unknown>) : undefined;
            const code = typeof data?.code === 'string' ? data.code : 'request_failed';
            return { ok: false, code, httpStatus: res.status, data };
        }
        return { ok: true, data: body.data as T };
    } catch {
        return { ok: false, code: 'request_failed', httpStatus: 0 };
    }
}

/** GET /api/v1/cognitive/budgets — the effective policy. `providers` is only
 * present for a root admin with cognitive:read or cognitive:write; callers
 * must treat it as optional. */
export function getTokenBudgets(): Promise<TokenBudgetResult<TokenBudgetsResponse>> {
    return callTokenBudgetAPI<TokenBudgetsResponse>('/api/v1/cognitive/budgets', { cache: 'no-store' });
}

/** GET /api/v1/cognitive/budgets/usage?team_id=|agent_id=|run_id= — exactly
 * one ref. A non-member gets 403 `token_budget_usage_forbidden`; agent and
 * run usage are admin-only. */
export function getTokenBudgetUsage(query: TokenBudgetUsageQuery): Promise<TokenBudgetResult<TokenBudgetUsage>> {
    const [key, value] = Object.entries(query)[0];
    const params = new URLSearchParams({ [key]: value });
    return callTokenBudgetAPI<TokenBudgetUsage>(`/api/v1/cognitive/budgets/usage?${params.toString()}`, { cache: 'no-store' });
}

/** PUT /api/v1/cognitive/budgets/overrides/{level}/{ref} — root admin +
 * cognitive:write only; validate with `validateOverrideDraft` first. */
export function putTokenBudgetOverride(
    level: TokenBudgetOverrideLevel,
    ref: string,
    body: TokenBudgetOverrideBody,
): Promise<TokenBudgetResult<TokenBudgetOverrideResult>> {
    return callTokenBudgetAPI<TokenBudgetOverrideResult>(
        `/api/v1/cognitive/budgets/overrides/${encodeURIComponent(level)}/${encodeURIComponent(ref)}`,
        { method: 'PUT', headers: { 'Content-Type': 'application/json' }, body: JSON.stringify(body) },
    );
}

/** DELETE /api/v1/cognitive/budgets/overrides/{level}/{ref} — idempotent. */
export function deleteTokenBudgetOverride(
    level: TokenBudgetOverrideLevel,
    ref: string,
): Promise<TokenBudgetResult<TokenBudgetOverrideResult>> {
    return callTokenBudgetAPI<TokenBudgetOverrideResult>(
        `/api/v1/cognitive/budgets/overrides/${encodeURIComponent(level)}/${encodeURIComponent(ref)}`,
        { method: 'DELETE' },
    );
}

/**
 * Client-side mirror of the server's validation (server/token_budgets.go
 * `tokenBudgetOverrideBody.limits`): every set field must be a positive
 * integer within [min_limit, max_limit], warn_pct within [1, 99], at least
 * one field must be set, and per_execution ≤ per_run ≤ per_team_day among
 * the fields actually set. Returns a human-readable message, or null when
 * the draft is valid. This never replaces the server's own validation —
 * it only avoids a round trip for an obviously invalid draft.
 */
export function validateOverrideDraft(
    draft: TokenBudgetOverrideBody,
    bounds: { min_limit: number; max_limit: number },
): string | null {
    const fields: Array<keyof TokenBudgetOverrideBody> = ['per_execution', 'per_run', 'per_team_day', 'per_agent_day', 'warn_pct'];
    const set = fields.filter((field) => draft[field] !== undefined);
    if (set.length === 0) {
        return 'Set at least one of per-execution, per-run, per-team-day, per-agent-day or warn %.';
    }
    for (const field of set) {
        const value = draft[field] as number;
        if (!Number.isInteger(value)) {
            return `${field} must be a whole number.`;
        }
        if (field === 'warn_pct') {
            if (value < 1 || value > 99) {
                return 'Warn % must be between 1 and 99.';
            }
            continue;
        }
        if (value < bounds.min_limit || value > bounds.max_limit) {
            return `${field} must be between ${bounds.min_limit.toLocaleString()} and ${bounds.max_limit.toLocaleString()} (there is no unlimited value).`;
        }
    }
    if (draft.per_execution !== undefined && draft.per_run !== undefined && draft.per_execution > draft.per_run) {
        return 'per_execution must be less than or equal to per_run.';
    }
    if (draft.per_run !== undefined && draft.per_team_day !== undefined && draft.per_run > draft.per_team_day) {
        return 'per_run must be less than or equal to per_team_day.';
    }
    if (draft.per_execution !== undefined && draft.per_team_day !== undefined && draft.per_run === undefined && draft.per_execution > draft.per_team_day) {
        return 'per_execution must be less than or equal to per_team_day.';
    }
    return null;
}

/**
 * Human k/M token formatting for the D10 UI summary ("12k", "180k", "2M").
 * Shared by TeamTokenUsage.tsx and, once U1 merges, the same line in
 * TeamDetailDrawer.tsx and the partial-label in ExecutionSummaryCardModel.ts
 * (see the U1 mount spec in the B1-W2 close-out report).
 */
export function formatTokenCount(n: number): string {
    if (n >= 1_000_000) return `${Math.round(n / 1_000_000)}M`;
    if (n >= 1_000) return `${Math.round(n / 1_000)}k`;
    return `${n}`;
}
