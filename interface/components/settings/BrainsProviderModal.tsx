"use client";

import React from "react";
import { CheckCircle, Wifi, WifiOff, Loader2, X } from "lucide-react";
import {
    ProviderFormData,
    PROVIDER_PRESETS,
    PRESET_LABELS,
    COUNCIL_ROLES,
    TOKEN_BUDGET_PRESETS,
    blankForm,
} from "./brainsTypes";

export function ProviderForm({
    form,
    onChange,
    isEdit,
    probeResult,
    probing,
    onProbe,
    showPresets,
}: {
    form: ProviderFormData;
    onChange: (f: ProviderFormData) => void;
    isEdit: boolean;
    probeResult: "alive" | "dead" | null;
    probing: boolean;
    onProbe?: () => void;
    showPresets: boolean;
}) {
    const field = (key: keyof ProviderFormData, label: string, node: React.ReactNode) => (
        <div className="space-y-1">
            <label className="text-[10px] uppercase tracking-wider text-cortex-text-muted">{label}</label>
            {node}
        </div>
    );

    const inputCls = "w-full bg-cortex-bg border border-cortex-border rounded px-2.5 py-1.5 text-xs text-cortex-text-main focus:outline-none focus:ring-1 focus:ring-cortex-primary placeholder:text-cortex-text-muted/50";

    const toggleRole = (role: string) => {
        const cur = form.roles_allowed;
        if (role === "all") {
            onChange({ ...form, roles_allowed: ["all"] });
            return;
        }
        const without = cur.filter((r) => r !== "all" && r !== role);
        const next = cur.includes(role) ? without : [...without, role];
        onChange({ ...form, roles_allowed: next.length ? next : ["all"] });
    };

    return (
        <div className="space-y-3">
            {showPresets && (
                <div className="space-y-1" data-testid="brains-web-service-presets">
                    <label className="text-[10px] uppercase tracking-wider text-cortex-text-muted">Web &amp; Local Engine Presets</label>
                    <div className="flex flex-wrap gap-1.5">
                        {Object.entries(PRESET_LABELS).map(([key, label]) => (
                            <button
                                key={key}
                                type="button"
                                data-testid={`brain-preset-${key}`}
                                onClick={() => {
                                    const preset = PROVIDER_PRESETS[key];
                                    const suggestedId = key === "custom" ? "" : key;
                                    onChange({ ...blankForm(), ...preset, id: suggestedId || form.id });
                                }}
                                className="px-2 py-0.5 rounded text-[10px] border border-cortex-border text-cortex-text-muted hover:border-cortex-primary hover:text-cortex-primary transition-colors"
                            >
                                {label}
                            </button>
                        ))}
                    </div>
                </div>
            )}

            <div className="grid grid-cols-2 gap-3">
                {field("id", "Provider ID", (
                    <input
                        value={form.id}
                        onChange={(e) => onChange({ ...form, id: e.target.value.toLowerCase().replace(/[^a-z0-9_-]/g, "") })}
                        placeholder="e.g. my-ollama"
                        disabled={isEdit}
                        className={`${inputCls} ${isEdit ? "opacity-50 cursor-not-allowed" : ""}`}
                    />
                ))}
                {field("type", "Provider Type", (
                    <input
                        value={form.type}
                        onChange={(e) => onChange({ ...form, type: e.target.value })}
                        placeholder="openai_compatible"
                        className={inputCls}
                    />
                ))}
                {field("endpoint", "Endpoint URL", (
                    <input
                        value={form.endpoint}
                        onChange={(e) => onChange({ ...form, endpoint: e.target.value })}
                        placeholder="http://localhost:11434/v1"
                        className={inputCls}
                    />
                ))}
                {field("model_id", "Model ID", (
                    <input
                        value={form.model_id}
                        onChange={(e) => onChange({ ...form, model_id: e.target.value })}
                        placeholder="llama3:8b"
                        className={inputCls}
                    />
                ))}
                {field("api_key", isEdit ? "API Key (blank = keep existing)" : "API Key", (
                    <input
                        type="password"
                        value={form.api_key}
                        onChange={(e) => onChange({ ...form, api_key: e.target.value })}
                        placeholder={isEdit ? "leave blank to keep existing" : "optional if using API Key Env Var"}
                        className={inputCls}
                    />
                ))}
                {field("api_key_env", "API Key Env Var (Secret Ref)", (
                    <input
                        data-testid="brain-form-api-key-env"
                        value={form.api_key_env}
                        onChange={(e) => onChange({ ...form, api_key_env: e.target.value.toUpperCase().replace(/[^A-Z0-9_]/g, "") })}
                        placeholder="e.g. OPENAI_API_KEY / ANTHROPIC_API_KEY"
                        className={inputCls}
                    />
                ))}
                {field("usage_policy", "Usage Policy", (
                    <select
                        value={form.usage_policy}
                        onChange={(e) => onChange({ ...form, usage_policy: e.target.value })}
                        className={inputCls}
                    >
                        <option value="local_first">Local First</option>
                        <option value="allow_escalation">Allow Escalation</option>
                        <option value="require_approval">Require Approval</option>
                        <option value="disallowed">Disallowed</option>
                    </select>
                ))}
                {field("token_budget_profile", "Token Budget", (
                    <select
                        value={form.token_budget_profile}
                        onChange={(e) => {
                            const profile = e.target.value;
                            const preset = TOKEN_BUDGET_PRESETS[profile];
                            onChange({
                                ...form,
                                token_budget_profile: profile,
                                max_output_tokens: preset ? preset.tokens : form.max_output_tokens,
                            });
                        }}
                        className={inputCls}
                    >
                        {Object.entries(TOKEN_BUDGET_PRESETS).map(([value, preset]) => (
                            <option key={value} value={value}>{preset.label}</option>
                        ))}
                    </select>
                ))}
                {field("max_output_tokens", "Max Output Tokens", (
                    <input
                        type="number"
                        min={128}
                        step={128}
                        value={form.max_output_tokens}
                        onChange={(e) => onChange({ ...form, max_output_tokens: Number(e.target.value) || 0 })}
                        className={inputCls}
                    />
                ))}
                {field("location", "Location", (
                    <select
                        value={form.location}
                        onChange={(e) => {
                            const loc = e.target.value;
                            onChange({
                                ...form,
                                location: loc,
                                data_boundary: loc === "remote" ? "leaves_org" : "local_only",
                            });
                        }}
                        className={inputCls}
                    >
                        <option value="local">Local</option>
                        <option value="remote">Remote</option>
                    </select>
                ))}
                {field("data_boundary", "Data Boundary", (
                    <select
                        value={form.data_boundary}
                        onChange={(e) => onChange({ ...form, data_boundary: e.target.value })}
                        className={inputCls}
                    >
                        <option value="local_only">Local Only</option>
                        <option value="leaves_org">Leaves Org</option>
                    </select>
                ))}
            </div>

            <p className="text-[11px] leading-5 text-cortex-text-muted">
                Safe defaults follow common operational ranges: 512 for conservative, 1024 for standard, 2048 for extended, and 4096 for deep output budgets.
            </p>

            <div className="space-y-1">
                <label className="text-[10px] uppercase tracking-wider text-cortex-text-muted">Roles Allowed</label>
                <div className="flex flex-wrap gap-1.5">
                    {COUNCIL_ROLES.map((role) => {
                        const active = form.roles_allowed.includes(role);
                        return (
                            <button
                                key={role}
                                type="button"
                                onClick={() => toggleRole(role)}
                                className={`px-2 py-0.5 rounded text-[10px] border transition-colors ${
                                    active
                                        ? "border-cortex-primary text-cortex-primary bg-cortex-primary/10"
                                        : "border-cortex-border text-cortex-text-muted hover:border-cortex-primary/50"
                                }`}
                            >
                                {role}
                            </button>
                        );
                    })}
                </div>
            </div>

            <label className="flex items-center gap-2 cursor-pointer text-xs text-cortex-text-muted select-none">
                <div
                    onClick={() => onChange({ ...form, enabled: !form.enabled })}
                    className={`w-8 h-4 rounded-full relative transition-colors border cursor-pointer ${
                        form.enabled
                            ? "bg-cortex-success/20 border-cortex-success/40"
                            : "bg-cortex-bg border-cortex-border"
                    }`}
                >
                    <div className={`absolute top-0 w-4 h-4 rounded-full shadow-sm transition-all ${
                        form.enabled ? "right-0 bg-cortex-success" : "left-0 bg-cortex-text-muted"
                    }`} />
                </div>
                Enable on save
            </label>

            {isEdit && onProbe && (
                <div className="flex items-center gap-2">
                    <button
                        type="button"
                        onClick={onProbe}
                        disabled={probing}
                        className="flex items-center gap-1.5 px-3 py-1.5 rounded border border-cortex-border text-xs text-cortex-text-muted hover:border-cortex-primary hover:text-cortex-primary transition-colors disabled:opacity-50"
                    >
                        {probing ? <Loader2 className="w-3 h-3 animate-spin" /> : <Wifi className="w-3 h-3" />}
                        Test Connection
                    </button>
                    {probeResult === "alive" && (
                        <span className="flex items-center gap-1 text-cortex-success text-xs">
                            <CheckCircle className="w-3 h-3" /> Online
                        </span>
                    )}
                    {probeResult === "dead" && (
                        <span className="flex items-center gap-1 text-red-400 text-xs">
                            <WifiOff className="w-3 h-3" /> Unreachable
                        </span>
                    )}
                </div>
            )}
        </div>
    );
}

export function Modal({ title, onClose, children }: { title: string; onClose: () => void; children: React.ReactNode }) {
    return (
        <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/60 backdrop-blur-sm">
            <div className="bg-cortex-surface border border-cortex-border rounded-xl w-full max-w-2xl max-h-[90vh] overflow-y-auto shadow-2xl">
                <div className="flex items-center justify-between px-5 py-3 border-b border-cortex-border">
                    <h3 className="text-sm font-semibold text-cortex-text-main">{title}</h3>
                    <button onClick={onClose} className="p-1 rounded hover:bg-cortex-border text-cortex-text-muted hover:text-cortex-text-main transition-colors">
                        <X className="w-4 h-4" />
                    </button>
                </div>
                <div className="p-5">{children}</div>
            </div>
        </div>
    );
}
