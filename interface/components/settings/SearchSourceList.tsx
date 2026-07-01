import React from "react";
import { Pencil, Trash2 } from "lucide-react";
import type { SearchCapabilitySource } from "@/store/useCortexStore";

export function SearchSourceList({
    sources,
    compact = false,
    onDeleteSource,
    onEditSource,
    title = "Search sources Soma can use",
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
                                <span className="rounded-full border border-cortex-border bg-cortex-bg px-2 py-0.5 text-[10px] font-mono uppercase text-cortex-text-muted">
                                    {sourceStatusLabel(source.status)}
                                </span>
                            </div>
                        </div>
                        <p className="mt-1 text-[11px] leading-4 text-cortex-text-muted">
                            {sourceTypeLabel(source.source_type)} · {scopeLabel(source)} · {authLabel(source)}
                        </p>
                        <p className="mt-1 text-[11px] leading-4 text-cortex-text-main">{source.boundary}</p>
                        {!compact && (
                            <details className="mt-2 rounded-lg border border-cortex-border bg-cortex-bg/60 px-2.5 py-2">
                                <summary className="cursor-pointer text-[9px] font-mono uppercase tracking-wider text-cortex-text-muted">
                                    Inspect source refs
                                </summary>
                                <div className="mt-2 grid gap-2 sm:grid-cols-2">
                                    <SourceDetail label="Source ref" value={source.id} />
                                    <SourceDetail label="Type ref" value={source.source_type} />
                                    <SourceDetail label="Scope ref" value={source.scope_ref ? `${source.scope_kind}:${source.scope_ref}` : source.scope_kind} />
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
    if (type === "local_sources") return "Retained Mycelis context";
    if (type === "local_api") return "Private search API";
    if (type === "knowledge_collection") return "Knowledge collection";
    return type.replace(/_/g, " ");
}

function sourceStatusLabel(status: string): string {
    if (status === "available" || status === "ready" || status === "online") return "Ready";
    if (status === "degraded") return "Needs repair";
    if (status === "disabled") return "Disabled";
    return status.replace(/_/g, " ");
}

function scopeLabel(source: SearchCapabilitySource): string {
    if (source.scope_kind === "all") return "Visible to everyone";
    if (source.scope_kind === "group") return source.scope_ref ? `Group ${source.scope_ref}` : "Group-scoped";
    if (source.scope_kind === "host") return source.scope_ref ? `Host ${source.scope_ref}` : "Host-scoped";
    return source.scope_ref ? `${source.scope_kind} ${source.scope_ref}` : source.scope_kind;
}

function authLabel(source: SearchCapabilitySource): string {
    if (source.auth_scheme === "none") return "No secret needed";
    if (source.secret_ref) return `Uses secret reference ${source.secret_ref}`;
    return "Uses saved authentication";
}
