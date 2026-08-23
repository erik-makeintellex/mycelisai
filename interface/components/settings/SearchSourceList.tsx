import React from "react";
import { Code2, Pencil, RefreshCw, Trash2, Wrench } from "lucide-react";
import type { SearchCapabilitySource } from "@/store/useCortexStore";

export function SearchSourceList({
    sources,
    compact = false,
    onDeleteSource,
    onEditSource,
    title = "Places Soma Can Search",
}: {
    sources: SearchCapabilitySource[];
    compact?: boolean;
    onDeleteSource?: (sourceId: string, sourceName: string) => Promise<boolean>;
    onEditSource?: (source: SearchCapabilitySource) => void;
    title?: string;
}) {
    if (sources.length === 0) return null;
    return (
        <div className="mt-3 rounded-lg border border-cortex-border bg-cortex-bg/60 p-3">
            <p className="text-[10px] font-mono font-bold uppercase tracking-wider text-cortex-text-muted">
                {title}
            </p>
            <p className="mt-1 text-[11px] leading-4 text-cortex-text-muted">
                Approved places Soma may search: public web, approved local or mounted data, code context, and private APIs.
            </p>
            <div className="mt-2 grid gap-2">
                {sources.slice(0, compact ? 2 : undefined).map((source) => (
                    <div key={source.id} className="rounded-lg border border-cortex-border bg-cortex-surface px-3 py-2">
                        <div className="flex flex-wrap items-center justify-between gap-2">
                            <span className="text-xs font-semibold text-cortex-text-main">{source.name}</span>
                            <div className="flex flex-wrap items-center gap-1.5">
                                {source.managed && !compact && onEditSource && (
                                    <button
                                        type="button"
                                        onClick={() => onEditSource(source)}
                                        className="inline-flex items-center gap-1 rounded-full border border-cortex-border bg-cortex-bg px-2 py-0.5 text-[10px] font-semibold text-cortex-text-muted transition hover:text-cortex-primary"
                                    >
                                        <Pencil className="h-3 w-3" />
                                        Edit
                                    </button>
                                )}
                                {source.managed && !compact && onDeleteSource && (
                                    <button
                                        type="button"
                                        onClick={() => void onDeleteSource(source.id, source.name)}
                                        className="inline-flex items-center gap-1 rounded-full border border-cortex-warning/30 bg-cortex-warning/10 px-2 py-0.5 text-[10px] font-semibold text-cortex-warning transition hover:bg-cortex-warning/20"
                                    >
                                        <Trash2 className="h-3 w-3" />
                                        Remove
                                    </button>
                                )}
                                <span className={sourceStatusClass(source)}>
                                    {sourceStatusLabel(source.status)}
                                </span>
                            </div>
                        </div>
                        <p className={`mt-1 text-[11px] font-semibold leading-4 ${sourceReadinessClass(source)}`}>
                            {sourceReadinessLabel(source)}
                        </p>
                        <p className="mt-1 text-[11px] leading-4 text-cortex-text-muted">
                            {sourceTypeLabel(source.source_type)} · {scopeLabel(source)} · {authLabel(source)}
                        </p>
                        <p className="mt-1 text-[11px] leading-4 text-cortex-text-main">{source.boundary}</p>
                        {!compact && sourceIsCodeContext(source) && <CodeContextSourceSummary source={source} />}
                        {!compact && (
                            <details className="mt-2 rounded-lg border border-cortex-border bg-cortex-bg/60 px-2.5 py-2">
                                <summary className="cursor-pointer text-[9px] font-mono uppercase tracking-wider text-cortex-text-muted">
                                    Inspect technical refs
                                </summary>
                                <div className="mt-2 grid gap-2 sm:grid-cols-2">
                                    <SourceDetail label="Source ref" value={source.id} />
                                    <SourceDetail label="Type ref" value={source.source_type} />
                                    <SourceDetail label="Scope ref" value={source.scope_ref ? `${source.scope_kind}:${source.scope_ref}` : source.scope_kind} />
                                    {source.code_context?.scope && <SourceDetail label="Code scope ref" value={source.code_context.scope} />}
                                    {source.code_context?.snapshot_ref && <SourceDetail label="Snapshot ref" value={source.code_context.snapshot_ref} />}
                                    {source.code_context?.snapshot_digest && <SourceDetail label="Snapshot digest" value={source.code_context.snapshot_digest} />}
                                    {source.code_context?.index_ref && <SourceDetail label="Index ref" value={source.code_context.index_ref} />}
                                    {source.code_context?.index_digest && <SourceDetail label="Index digest" value={source.code_context.index_digest} />}
                                    <SourceDetail label="Auth ref" value={source.secret_ref ? `${source.auth_scheme}:${source.secret_ref}` : source.auth_scheme} />
                                    <SourceDetail label="Mode" value={source.mode} />
                                    <SourceDetail label="Trust" value={`${source.sensitivity_class} / ${source.trust_class}`} />
                                </div>
                            </details>
                        )}
                    </div>
                ))}
            </div>
        </div>
    );
}

function CodeContextSourceSummary({ source }: { source: SearchCapabilitySource }) {
    const codeContext = source.code_context;
    const snapshotStatus = codeContext?.snapshot_status ?? source.status;
    const indexStatus = codeContext?.index_status ?? "not reported";
    const needsRepair = sourceNeedsCodeContextRepair(source);
    const needsRefresh = !needsRepair && sourceNeedsCodeContextRefresh(source);
    const actionLabel = needsRepair
        ? (codeContext?.repair_action ?? source.recovery ?? "Repair code context map before Soma relies on impact refs.")
        : needsRefresh
        ? (codeContext?.refresh_action ?? "Refresh code context map after repository changes.")
        : (codeContext?.refresh_action ?? "Refresh code context map when the source changes.");

    return (
        <div className="mt-2 rounded-lg border border-cortex-border bg-cortex-bg/60 px-2.5 py-2">
            <div className="flex flex-wrap items-center gap-2">
                <span className="inline-flex items-center gap-1 rounded border border-cortex-border bg-cortex-surface px-1.5 py-0.5 text-[9px] font-mono uppercase text-cortex-text-muted">
                    <Code2 className="h-3 w-3" />
                    Native code context
                </span>
                <span className={codeContextStatusClass(snapshotStatus)}>Snapshot {statusText(snapshotStatus)}</span>
                <span className={codeContextStatusClass(indexStatus)}>Index {statusText(indexStatus)}</span>
            </div>
            <div className="mt-2 grid gap-2 sm:grid-cols-2">
                <CodeContextFact label="Scope" value={codeContext?.scope ?? scopeLabel(source)} />
                <CodeContextFact label="Last snapshot" value={codeContext?.last_snapshot_at ?? "Not reported"} />
                <CodeContextFact label="Last index" value={codeContext?.last_indexed_at ?? "Not reported"} />
                <div className={`rounded-lg border px-2.5 py-2 ${needsRepair ? "border-cortex-warning/25 bg-cortex-warning/10" : "border-cortex-border bg-cortex-surface"}`}>
                    <p className="flex items-center gap-1 text-[9px] font-mono uppercase tracking-wider text-cortex-text-muted">
                        {needsRepair ? <Wrench className="h-3 w-3 text-cortex-warning" /> : <RefreshCw className="h-3 w-3 text-cortex-info" />}
                        {needsRepair ? "Repair" : "Refresh"}
                    </p>
                    <p className="mt-1 text-[11px] leading-4 text-cortex-text-main">{actionLabel}</p>
                </div>
            </div>
        </div>
    );
}

function CodeContextFact({ label, value }: { label: string; value: string }) {
    return (
        <div className="rounded-lg border border-cortex-border bg-cortex-surface px-2.5 py-2">
            <p className="text-[9px] font-mono uppercase tracking-wider text-cortex-text-muted">{label}</p>
            <p className="mt-1 break-words text-[11px] leading-4 text-cortex-text-main">{value}</p>
        </div>
    );
}

function SourceDetail({ label, value }: { label: string; value: string }) {
    return (
        <div className="rounded-lg border border-cortex-border bg-cortex-surface px-2.5 py-2">
            <p className="text-[9px] font-mono uppercase tracking-wider text-cortex-text-muted">{label}</p>
            <p className="mt-1 break-words text-[11px] leading-4 text-cortex-text-main">{value}</p>
        </div>
    );
}

function sourceTypeLabel(type: string): string {
    if (type === "public_web") return "Public web";
    if (type === "local_sources") return "Approved Mycelis data";
    if (type === "local_api") return "Private API";
    if (type === "mounted_folder") return "Approved local or mounted data";
    if (type === "knowledge_collection") return "Approved knowledge collection";
    if (type === "code_context") return "Code context";
    return type.replace(/_/g, " ");
}

function sourceStatusLabel(status: string): string {
    if (status === "available" || status === "ready" || status === "online") return "Ready";
    if (status === "degraded") return "Needs repair";
    if (status === "disabled") return "Disabled";
    return status.replace(/_/g, " ");
}

function sourceStatusClass(source: SearchCapabilitySource): string {
    const base = "rounded-full border px-2 py-0.5 text-[10px] font-mono uppercase";
    if (!sourceIsReady(source.status) || sourceAuthNeedsAdapter(source)) {
        return `${base} border-cortex-warning/30 bg-cortex-warning/10 text-cortex-warning`;
    }
    return `${base} border-cortex-success/30 bg-cortex-success/10 text-cortex-success`;
}

function sourceReadinessLabel(source: SearchCapabilitySource): string {
    if (sourceIsCodeContext(source)) {
        if (sourceNeedsCodeContextRepair(source)) {
            return source.recovery || source.code_context?.repair_action || "Repair code context map before Soma relies on impact refs.";
        }
        if (sourceNeedsCodeContextRefresh(source)) {
            return source.code_context?.refresh_action || "Refresh code context map before relying on stale impact refs.";
        }
    }
    if (!sourceIsReady(source.status)) {
        return source.recovery || "Repair this source before Soma can use it.";
    }
    if (sourceUsesResolvableToken(source)) {
        return "Ready once saved access is available.";
    }
    if (sourceAuthNeedsAdapter(source)) {
        return "Registered safely. Soma needs a matching auth adapter before it can search this source.";
    }
    return "Ready for Soma to use when this scope is allowed.";
}

function sourceReadinessClass(source: SearchCapabilitySource): string {
    if (sourceIsCodeContext(source) && (sourceNeedsCodeContextRepair(source) || sourceNeedsCodeContextRefresh(source))) return "text-cortex-warning";
    if (!sourceIsReady(source.status) || sourceAuthNeedsAdapter(source)) return "text-cortex-warning";
    return "text-cortex-success";
}

function sourceIsReady(status: string): boolean {
    return ["available", "ready", "online"].includes(status);
}

function sourceAuthNeedsAdapter(source: SearchCapabilitySource): boolean {
    return source.auth_scheme !== "none" && source.auth_scheme !== "service_managed" && !sourceUsesResolvableToken(source);
}

function sourceUsesResolvableToken(source: SearchCapabilitySource): boolean {
    return source.auth_scheme === "api_token" || source.auth_scheme === "bearer_token";
}

function scopeLabel(source: SearchCapabilitySource): string {
    if (source.scope_kind === "all") return "Visible to everyone";
    if (source.scope_kind === "group") return "Visible to one group";
    if (source.scope_kind === "host") return "Visible to one host";
    return "Limited visibility";
}

function authLabel(source: SearchCapabilitySource): string {
    if (source.auth_scheme === "none") return "No secret needed";
    if (sourceUsesResolvableToken(source)) return "Uses a saved secret";
    return "Uses saved authentication";
}

function sourceIsCodeContext(source: SearchCapabilitySource): boolean {
    return source.source_type === "code_context" || source.provider === "code_context" || Boolean(source.code_context);
}

function sourceNeedsCodeContextRepair(source: SearchCapabilitySource): boolean {
    return codeContextStatusNeedsRepair(source.status)
        || codeContextStatusNeedsRepair(source.code_context?.snapshot_status)
        || codeContextStatusNeedsRepair(source.code_context?.index_status);
}

function sourceNeedsCodeContextRefresh(source: SearchCapabilitySource): boolean {
    return codeContextStatusNeedsRefresh(source.code_context?.snapshot_status)
        || codeContextStatusNeedsRefresh(source.code_context?.index_status)
        || (!source.code_context?.last_snapshot_at && !source.code_context?.last_indexed_at);
}

function codeContextStatusClass(status: string): string {
    const base = "rounded border px-1.5 py-0.5 text-[9px] font-mono uppercase";
    if (codeContextStatusNeedsRepair(status)) return `${base} border-cortex-warning/25 bg-cortex-warning/10 text-cortex-warning`;
    if (codeContextStatusNeedsRefresh(status)) return `${base} border-cortex-warning/25 bg-cortex-warning/10 text-cortex-warning`;
    if (sourceIsReady(status)) return `${base} border-cortex-success/25 bg-cortex-success/10 text-cortex-success`;
    return `${base} border-cortex-border bg-cortex-surface text-cortex-text-muted`;
}

function codeContextStatusNeedsRepair(status?: string): boolean {
    const value = status?.toLowerCase();
    return value === "degraded" || value === "failed" || value === "error" || value === "missing" || value === "unavailable" || value === "repair_required";
}

function codeContextStatusNeedsRefresh(status?: string): boolean {
    const value = status?.toLowerCase();
    return value === "stale" || value === "outdated" || value === "refresh_required";
}

function statusText(status: string): string {
    return status.replace(/_/g, " ");
}
