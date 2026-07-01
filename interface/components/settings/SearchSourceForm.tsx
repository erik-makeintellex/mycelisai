"use client";

import React from "react";
import type { SearchCapabilitySource } from "@/store/useCortexStore";

export interface SearchSourceDraft {
    name: string;
    source_type: string;
    endpoint?: string;
    boundary: string;
    scope_kind: string;
    scope_ref?: string;
    auth_scheme: string;
    secret_ref?: string;
    mode: string;
    sensitivity_class: string;
    trust_class: string;
}

export function SearchSourceAddForm({
    initialDraft,
    isAdding,
    submitLabel = "Add search source",
    onCancel,
    onSubmit,
}: {
    initialDraft?: SearchSourceDraft;
    isAdding: boolean;
    submitLabel?: string;
    onCancel: () => void;
    onSubmit: (input: SearchSourceDraft) => Promise<boolean>;
}) {
    const [draft, setDraft] = React.useState<SearchSourceDraft>(initialDraft ?? {
        name: "",
        source_type: "public_web",
        endpoint: "",
        boundary: "",
        scope_kind: "all",
        auth_scheme: "none",
        mode: "live",
        sensitivity_class: "public",
        trust_class: "bounded_external",
    });
    const [formError, setFormError] = React.useState<string | null>(null);

    async function submit(event: React.FormEvent<HTMLFormElement>) {
        event.preventDefault();
        const validationError = validateSearchSourceDraft(draft);
        if (validationError) {
            setFormError(validationError);
            return;
        }
        setFormError(null);
        await onSubmit({
            ...draft,
            auth_scheme: draft.auth_scheme === "secret_ref" ? "api_token" : draft.auth_scheme,
            endpoint: draft.endpoint?.trim() || undefined,
            name: draft.name.trim(),
            boundary: draft.boundary.trim(),
            scope_ref: draft.scope_ref?.trim() || undefined,
            secret_ref: draft.secret_ref?.trim() || undefined,
        });
    }

    return (
        <form onSubmit={submit} className="mt-4 rounded-lg border border-cortex-border bg-cortex-bg/60 p-3">
            <div className="grid gap-3 md:grid-cols-2">
                <SourceField label="Source name">
                    <input value={draft.name} onChange={(event) => setDraft((current) => ({ ...current, name: event.target.value }))} className="w-full rounded-lg border border-cortex-border bg-cortex-surface px-3 py-2 text-xs text-cortex-text-main outline-none focus:border-cortex-primary/50" placeholder="Company knowledge search" />
                </SourceField>
                <SourceField label="Kind">
                    <select value={draft.source_type} onChange={(event) => setDraft((current) => ({ ...current, source_type: event.target.value }))} className="w-full rounded-lg border border-cortex-border bg-cortex-surface px-3 py-2 text-xs text-cortex-text-main outline-none focus:border-cortex-primary/50">
                        <option value="public_web">Public web</option>
                        <option value="local_sources">Retained Mycelis context</option>
                        <option value="local_api">Private search API</option>
                        <option value="knowledge_collection">Knowledge collection</option>
                    </select>
                </SourceField>
                <SourceField label="Endpoint for web/API">
                    <input value={draft.endpoint ?? ""} onChange={(event) => setDraft((current) => ({ ...current, endpoint: event.target.value }))} className="w-full rounded-lg border border-cortex-border bg-cortex-surface px-3 py-2 text-xs text-cortex-text-main outline-none focus:border-cortex-primary/50" placeholder="https://search.example.com/api" />
                </SourceField>
                <SourceField label="Boundary">
                    <input value={draft.boundary} onChange={(event) => setDraft((current) => ({ ...current, boundary: event.target.value }))} className="w-full rounded-lg border border-cortex-border bg-cortex-surface px-3 py-2 text-xs text-cortex-text-main outline-none focus:border-cortex-primary/50" placeholder="Approved company knowledge index" />
                </SourceField>
                <SourceField label="Visible to">
                    <select value={draft.scope_kind} onChange={(event) => setDraft((current) => ({ ...current, scope_kind: event.target.value, scope_ref: event.target.value === "all" ? "" : current.scope_ref }))} className="w-full rounded-lg border border-cortex-border bg-cortex-surface px-3 py-2 text-xs text-cortex-text-main outline-none focus:border-cortex-primary/50">
                        <option value="all">Everyone</option>
                        <option value="group">One group</option>
                        <option value="host">One host</option>
                    </select>
                </SourceField>
                {draft.scope_kind !== "all" && (
                    <SourceField label="Scope reference">
                        <input value={draft.scope_ref ?? ""} onChange={(event) => setDraft((current) => ({ ...current, scope_ref: event.target.value }))} className="w-full rounded-lg border border-cortex-border bg-cortex-surface px-3 py-2 text-xs text-cortex-text-main outline-none focus:border-cortex-primary/50" placeholder={draft.scope_kind === "group" ? "research-team" : "workstation-1"} />
                    </SourceField>
                )}
                <SourceField label="Authentication">
                    <select value={draft.auth_scheme} onChange={(event) => setDraft((current) => ({ ...current, auth_scheme: event.target.value, secret_ref: event.target.value === "none" ? "" : current.secret_ref }))} className="w-full rounded-lg border border-cortex-border bg-cortex-surface px-3 py-2 text-xs text-cortex-text-main outline-none focus:border-cortex-primary/50">
                        <option value="none">No secret needed</option>
                        <option value="secret_ref">Use a saved secret reference</option>
                    </select>
                </SourceField>
                {draft.auth_scheme !== "none" && (
                    <SourceField label="Secret reference">
                        <input value={draft.secret_ref ?? ""} onChange={(event) => setDraft((current) => ({ ...current, secret_ref: event.target.value }))} className="w-full rounded-lg border border-cortex-border bg-cortex-surface px-3 py-2 text-xs text-cortex-text-main outline-none focus:border-cortex-primary/50" placeholder="SEARCH_PROVIDER_API_KEY" />
                        <span className="text-[10px] leading-4 text-cortex-text-muted">Use a saved reference name, not the token value.</span>
                    </SourceField>
                )}
            </div>
            {formError && <p className="mt-3 text-[11px] text-cortex-warning">{formError}</p>}
            <div className="mt-3 flex justify-end gap-2">
                <button type="button" onClick={onCancel} className="rounded-lg border border-cortex-border bg-cortex-surface px-3 py-2 text-xs font-semibold text-cortex-text-muted transition hover:text-cortex-text-main">Cancel</button>
                <button type="submit" disabled={isAdding} className="rounded-lg border border-cortex-primary/30 bg-cortex-primary/10 px-3 py-2 text-xs font-semibold text-cortex-primary transition hover:bg-cortex-primary/20 disabled:cursor-not-allowed disabled:opacity-60">
                    {isAdding ? "Saving..." : submitLabel}
                </button>
            </div>
        </form>
    );
}

export function sourceToDraft(source: SearchCapabilitySource): SearchSourceDraft {
    return {
        name: source.name,
        source_type: source.source_type,
        endpoint: source.endpoint ?? source.base_url ?? "",
        boundary: source.boundary,
        scope_kind: source.scope_kind || "all",
        scope_ref: source.scope_ref,
        auth_scheme: source.auth_scheme === "api_token" ? "secret_ref" : source.auth_scheme,
        secret_ref: source.secret_ref,
        mode: source.mode || "live",
        sensitivity_class: source.sensitivity_class || "public",
        trust_class: source.trust_class || "bounded_external",
    };
}

function SourceField({ children, label }: { children: React.ReactNode; label: string }) {
    return (
        <label className="grid gap-1">
            <span className="text-[10px] font-mono uppercase tracking-wider text-cortex-text-muted">{label}</span>
            {children}
        </label>
    );
}

export function emptySourceLabel(addSupported: boolean): string {
    return addSupported ? "No configured sources reported" : "Source management unavailable";
}

function validateSearchSourceDraft(draft: SearchSourceDraft): string | null {
    if (!draft.name.trim()) return "Name the search source.";
    if (requiresEndpoint(draft.source_type) && !draft.endpoint?.trim()) return "Add the search endpoint Soma may use for this source.";
    if (draft.endpoint?.trim() && !/^https?:\/\/[^/\s]+/i.test(draft.endpoint.trim())) return "Use a full http(s) endpoint.";
    if (!draft.boundary.trim()) return "Describe the boundary Soma may search.";
    if (draft.scope_kind !== "all" && !draft.scope_ref?.trim()) return "Add the group or host reference for this source.";
    if (draft.auth_scheme !== "none") {
        const secretRef = draft.secret_ref?.trim() ?? "";
        if (!secretRef) return "Use a secret reference name, not a secret value.";
        if (secretRef.includes("=") || /\s/.test(secretRef) || /^sk-[A-Za-z0-9]/.test(secretRef)) return "Use a saved secret reference name, not a raw secret value.";
    }
    return null;
}

function requiresEndpoint(sourceType: string): boolean {
    return sourceType === "public_web" || sourceType === "local_api";
}
