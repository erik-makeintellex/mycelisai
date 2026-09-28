"use client";

import { useCallback, useEffect, useState } from "react";
import InlineBlockerNotice from "@/components/shared/InlineBlockerNotice";
import {
    deleteTokenBudgetOverride,
    getTokenBudgets,
    putTokenBudgetOverride,
    validateOverrideDraft,
    type TokenBudgetOverrideBody,
    type TokenBudgetOverrideLevel,
    type TokenBudgetsResponse,
} from "@/lib/tokenBudgets";

// B1: the effective token-budget policy plus (admin-only) override controls.
// Mounted under Advanced in BrainsPage.tsx. See
// docs/architecture-library/B1_TOKEN_BUDGETS_CONTRACT.md (D8, D10).

const LIMIT_FIELDS: Array<{ key: keyof TokenBudgetOverrideBody; label: string }> = [
    { key: "per_execution", label: "Per execution" },
    { key: "per_run", label: "Per run" },
    { key: "per_team_day", label: "Per team / day" },
    { key: "per_agent_day", label: "Per agent / day" },
    { key: "warn_pct", label: "Warn %" },
];

const OVERRIDE_LEVELS: TokenBudgetOverrideLevel[] = ["team", "agent", "profile", "class"];

function fmt(n: number | undefined): string {
    return typeof n === "number" ? n.toLocaleString() : "—";
}

function emptyDraft(): TokenBudgetOverrideBody {
    return {};
}

export default function TokenBudgetsPanel() {
    const [policy, setPolicy] = useState<TokenBudgetsResponse | null>(null);
    const [loadError, setLoadError] = useState<{ code: string; httpStatus?: number } | null>(null);
    const [level, setLevel] = useState<TokenBudgetOverrideLevel>("team");
    const [ref, setRef] = useState("");
    const [draft, setDraft] = useState<TokenBudgetOverrideBody>(emptyDraft());
    const [formError, setFormError] = useState<string | null>(null);
    const [submitting, setSubmitting] = useState(false);
    const [submitBlocker, setSubmitBlocker] = useState<{ code: string; httpStatus?: number } | null>(null);

    const load = useCallback(async () => {
        const result = await getTokenBudgets();
        if (!result.ok) {
            setLoadError({ code: result.code, httpStatus: result.httpStatus || undefined });
            return;
        }
        const data = result.data;
        if (!data || !Array.isArray(data.class_order) || !data.classes || !data.global) {
            // A caller can only see this with an unexpected/mocked backend
            // shape; fail closed to the blocker rather than crash the panel.
            setLoadError({ code: "request_failed" });
            return;
        }
        setLoadError(null);
        setPolicy(data);
    }, []);

    useEffect(() => {
        load();
    }, [load]);

    const submitOverride = async () => {
        if (!policy) return;
        const message = validateOverrideDraft(draft, { min_limit: policy.min_limit, max_limit: policy.max_limit });
        if (message) {
            setFormError(message);
            return;
        }
        if (!ref.trim()) {
            setFormError("Give a ref to apply this override to.");
            return;
        }
        setFormError(null);
        setSubmitBlocker(null);
        setSubmitting(true);
        try {
            const result = await putTokenBudgetOverride(level, ref.trim(), draft);
            if (!result.ok) {
                setSubmitBlocker({ code: result.code, httpStatus: result.httpStatus || undefined });
                return;
            }
            setRef("");
            setDraft(emptyDraft());
            await load();
        } finally {
            setSubmitting(false);
        }
    };

    const removeOverride = async (removeLevel: TokenBudgetOverrideLevel, removeRef: string) => {
        setSubmitBlocker(null);
        const result = await deleteTokenBudgetOverride(removeLevel, removeRef);
        if (!result.ok) {
            setSubmitBlocker({ code: result.code, httpStatus: result.httpStatus || undefined });
            return;
        }
        await load();
    };

    if (loadError) {
        return <InlineBlockerNotice code={loadError.code} httpStatus={loadError.httpStatus} onRetry={load} />;
    }
    if (!policy) {
        return <p className="text-xs text-cortex-text-muted">Loading token budgets…</p>;
    }

    const classRows = policy.class_order.filter((cls) => policy.classes[cls]).map((cls) => ({ name: cls, limits: policy.classes[cls] }));
    const overrideRows = OVERRIDE_LEVELS.flatMap((lvl) =>
        Object.entries(policy.overrides?.[lvl] ?? {}).map(([overrideRef, limits]) => ({ level: lvl, ref: overrideRef, limits })),
    );

    return (
        <div className="space-y-4">
            <div className="rounded-lg border border-cortex-border overflow-hidden">
                <table className="w-full text-xs font-mono">
                    <thead>
                        <tr className="bg-cortex-surface/50 text-cortex-text-muted">
                            <th className="text-left px-3 py-2">Class</th>
                            {LIMIT_FIELDS.map((field) => (
                                <th key={field.key} className="text-left px-3 py-2">{field.label}</th>
                            ))}
                        </tr>
                    </thead>
                    <tbody>
                        {classRows.map((row) => (
                            <tr key={row.name} className="border-t border-cortex-border">
                                <td className="px-3 py-2 text-cortex-text-main">{row.name}</td>
                                {LIMIT_FIELDS.map((field) => (
                                    <td key={field.key} className="px-3 py-2 text-cortex-text-main">{fmt(row.limits[field.key])}</td>
                                ))}
                            </tr>
                        ))}
                        <tr className="border-t border-cortex-border bg-cortex-surface/30">
                            <td className="px-3 py-2 text-cortex-text-muted">global (fallback)</td>
                            {LIMIT_FIELDS.map((field) => (
                                <td key={field.key} className="px-3 py-2 text-cortex-text-muted">{fmt(policy.global[field.key])}</td>
                            ))}
                        </tr>
                    </tbody>
                </table>
            </div>
            <p className="text-[10px] text-cortex-text-muted">
                Day period: {policy.day_period === "since_restart" ? "since restart" : "resets daily (UTC)"}.
            </p>

            {policy.providers && policy.providers.length > 0 && (
                <div className="rounded-lg border border-cortex-border overflow-hidden">
                    <table className="w-full text-xs font-mono">
                        <thead>
                            <tr className="bg-cortex-surface/50 text-cortex-text-muted">
                                <th className="text-left px-3 py-2">Provider</th>
                                <th className="text-left px-3 py-2">Model</th>
                                <th className="text-left px-3 py-2">Class</th>
                            </tr>
                        </thead>
                        <tbody>
                            {policy.providers.map((provider) => (
                                <tr key={provider.provider_id} className="border-t border-cortex-border">
                                    <td className="px-3 py-2 text-cortex-text-main">{provider.provider_id}</td>
                                    <td className="px-3 py-2 text-cortex-text-main">{provider.model_id}</td>
                                    <td className="px-3 py-2 text-cortex-text-muted">{provider.budget_class}</td>
                                </tr>
                            ))}
                        </tbody>
                    </table>
                </div>
            )}

            {policy.can_edit && (
                <div className="space-y-3 rounded-lg border border-cortex-border p-3">
                    <h4 className="text-[10px] uppercase tracking-wider text-cortex-text-muted">Overrides</h4>

                    {overrideRows.length > 0 && (
                        <ul className="space-y-1">
                            {overrideRows.map((row) => (
                                <li key={`${row.level}-${row.ref}`} className="flex items-center justify-between text-xs text-cortex-text-main">
                                    <span>{row.level}: <strong>{row.ref}</strong></span>
                                    <button
                                        onClick={() => removeOverride(row.level, row.ref)}
                                        className="px-2 py-0.5 rounded border border-cortex-border text-cortex-text-muted hover:text-red-400"
                                    >
                                        Remove
                                    </button>
                                </li>
                            ))}
                        </ul>
                    )}

                    {submitBlocker && (
                        <InlineBlockerNotice code={submitBlocker.code} httpStatus={submitBlocker.httpStatus} onDismiss={() => setSubmitBlocker(null)} />
                    )}

                    <div className="grid grid-cols-2 gap-2 sm:grid-cols-3">
                        <label className="space-y-1 text-[10px] uppercase tracking-wider text-cortex-text-muted">
                            Level
                            <select
                                value={level}
                                onChange={(e) => setLevel(e.target.value as TokenBudgetOverrideLevel)}
                                className="block w-full bg-cortex-bg border border-cortex-border rounded px-2 py-1 text-xs text-cortex-text-main"
                            >
                                {OVERRIDE_LEVELS.map((lvl) => (
                                    <option key={lvl} value={lvl}>{lvl}</option>
                                ))}
                            </select>
                        </label>
                        <label className="space-y-1 text-[10px] uppercase tracking-wider text-cortex-text-muted">
                            Ref
                            <input
                                aria-label="Ref"
                                value={ref}
                                onChange={(e) => setRef(e.target.value)}
                                className="block w-full bg-cortex-bg border border-cortex-border rounded px-2 py-1 text-xs text-cortex-text-main"
                            />
                        </label>
                        {LIMIT_FIELDS.map((field) => (
                            <label key={field.key} className="space-y-1 text-[10px] uppercase tracking-wider text-cortex-text-muted">
                                {field.label}
                                <input
                                    aria-label={field.label}
                                    type="number"
                                    value={draft[field.key] ?? ""}
                                    onChange={(e) => {
                                        const value = e.target.value;
                                        setDraft((prev) => ({ ...prev, [field.key]: value === "" ? undefined : Number(value) }));
                                    }}
                                    className="block w-full bg-cortex-bg border border-cortex-border rounded px-2 py-1 text-xs text-cortex-text-main"
                                />
                            </label>
                        ))}
                    </div>

                    {formError && <p className="text-xs text-red-400">{formError}</p>}

                    <button
                        onClick={submitOverride}
                        disabled={submitting}
                        className="px-3 py-1.5 rounded bg-cortex-primary/10 border border-cortex-primary/30 text-cortex-primary text-xs hover:bg-cortex-primary/20 disabled:opacity-50"
                    >
                        Save override
                    </button>
                </div>
            )}
        </div>
    );
}
