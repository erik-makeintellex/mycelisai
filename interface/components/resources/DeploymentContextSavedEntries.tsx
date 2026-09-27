import type { ReactNode } from "react";
import { MessageSquareText } from "lucide-react";
import { requestSomaOutputContinuation } from "@/components/soma/outputContinuation";
import type { DeploymentContextEntry } from "./DeploymentContextPanel";
import DeploymentContextEntryActions from "./DeploymentContextEntryActions";
import { KNOWLEDGE_CLASS_OPTIONS, SOURCE_KIND_OPTIONS, VISIBILITY_OPTIONS, optionLabel } from "./DeploymentContextOptions";

// How Soma can find each saved source today, from the stored chunk rows.
const SEARCH_STATUS: Record<string, { label: string; tone: string }> = {
    embedded: { label: "Keyword + meaning search", tone: "bg-cortex-success/10 text-cortex-success" },
    pending: { label: "Keyword search", tone: "bg-cortex-warning/10 text-cortex-warning" },
    failed_dimension: { label: "Keyword search (engine mismatch)", tone: "bg-cortex-warning/10 text-cortex-warning" },
    not_indexed: { label: "Not searchable", tone: "bg-cortex-danger/10 text-cortex-danger" },
};

export default function DeploymentContextSavedEntries({
    entries,
    loading,
    error = null,
    viewerIsAdmin = false,
    showArchived = false,
    onShowArchivedChange,
    onChanged,
}: {
    entries: DeploymentContextEntry[];
    loading: boolean;
    error?: string | null;
    viewerIsAdmin?: boolean;
    showArchived?: boolean;
    onShowArchivedChange?: (next: boolean) => void;
    onChanged?: () => void | Promise<void>;
}) {
    const askSomaWithContext = (entry: DeploymentContextEntry) => {
        requestSomaOutputContinuation(
            {
                title: entry.title,
                reference: `memory/deployment-context/${entry.artifact_id}`,
                proof: entry.trust_class,
                sourceLabel: "saved context source",
            },
            { persist: true, openSoma: true },
        );
    };

    return (
        <section className="flex min-h-0 flex-col overflow-hidden rounded-lg border border-cortex-border bg-cortex-surface">
            <div className="flex flex-shrink-0 items-center justify-between gap-3 border-b border-cortex-border px-5 py-4">
                <div>
                    <h2 className="text-sm font-semibold text-cortex-text-main">Saved Context Sources</h2>
                    <p className="mt-1 text-xs text-cortex-text-muted">
                        Durable source material Soma can recall separately from chat memory and generated output files.
                    </p>
                </div>
                <div className="flex flex-shrink-0 items-center gap-3">
                    {onShowArchivedChange ? (
                        <label className="inline-flex items-center gap-1.5 text-xs text-cortex-text-muted">
                            <input type="checkbox" checked={showArchived} onChange={(e) => onShowArchivedChange(e.target.checked)} />
                            Show archived
                        </label>
                    ) : null}
                    <span className="text-[11px] font-mono text-cortex-text-muted">{entries.length} entries</span>
                </div>
            </div>

            <div className="min-h-0 flex-1 space-y-3 overflow-y-auto p-5" role="region" aria-label="Loaded governed context list">
                {error ? (
                    <p role="alert" className="rounded-lg border border-cortex-danger/40 bg-cortex-danger/10 px-3 py-2 text-xs text-cortex-danger">{error}</p>
                ) : null}
                {loading ? (
                    <p className="animate-pulse text-xs font-mono text-cortex-text-muted">Loading deployment context...</p>
                ) : entries.length === 0 ? (
                    <div className="rounded-xl border border-dashed border-cortex-border p-4 text-xs text-cortex-text-muted">
                        No long-term context sources saved yet.
                    </div>
                ) : (
                    entries.map((entry) => (
                        <article key={entry.artifact_id} className="space-y-2 rounded-xl border border-cortex-border bg-cortex-bg/60 p-4">
                            <div className="flex items-start justify-between gap-3">
                                <div>
                                    <h3 className="text-sm font-semibold text-cortex-text-main">{entry.title}</h3>
                                    <p className="mt-1 text-[11px] text-cortex-text-muted">
                                        {entry.source_label} · {optionLabel(SOURCE_KIND_OPTIONS, entry.source_kind)}
                                    </p>
                                </div>
                                {entry.lifecycle_state === "archived" ? (
                                    <span className="whitespace-nowrap rounded bg-cortex-bg px-2 py-1 text-[10px] font-semibold text-cortex-text-muted">Archived</span>
                                ) : (
                                    <SearchStatus status={entry.embedding_status} />
                                )}
                            </div>
                            <p className="text-sm leading-relaxed text-cortex-text-main">{entry.content_preview}</p>
                            <div className="flex flex-wrap gap-2 text-[10px] font-mono text-cortex-text-muted">
                                <Badge>{optionLabel(KNOWLEDGE_CLASS_OPTIONS, entry.knowledge_class)}</Badge>
                                <Badge>{optionLabel(VISIBILITY_OPTIONS, entry.visibility)}</Badge>
                                <Badge>{entry.sensitivity_class}</Badge>
                                <Badge>{entry.trust_class}</Badge>
                                {entry.content_domain ? <Badge>{entry.content_domain.replaceAll("_", " ")}</Badge> : null}
                                {(entry.target_goal_sets ?? []).map((goal) => (
                                    <Badge key={`${entry.artifact_id}-${goal}`}>goal: {goal}</Badge>
                                ))}
                                <Badge>{entry.chunk_count} chunks</Badge>
                            </div>
                            {entry.lifecycle_state === "archived" ? (
                                <p className="text-xs text-cortex-text-muted">Archived: Soma does not use this until you restore it.</p>
                            ) : (
                                <button
                                    type="button"
                                    onClick={() => askSomaWithContext(entry)}
                                    className="inline-flex items-center gap-1.5 rounded border border-cortex-primary/40 px-2.5 py-1.5 text-xs font-semibold text-cortex-primary hover:bg-cortex-primary/10"
                                >
                                    <MessageSquareText className="h-3.5 w-3.5" />
                                    Ask Soma with this
                                </button>
                            )}
                            {onChanged ? <DeploymentContextEntryActions entry={entry} viewerIsAdmin={viewerIsAdmin} onChanged={onChanged} /> : null}
                        </article>
                    ))
                )}
            </div>
        </section>
    );
}

function SearchStatus({ status }: { status?: string }) {
    const known = SEARCH_STATUS[status ?? ""] ?? { label: "Search status unknown", tone: "bg-cortex-bg text-cortex-text-muted" };
    return <span className={`whitespace-nowrap rounded px-2 py-1 text-[10px] font-semibold ${known.tone}`}>{known.label}</span>;
}

function Badge({ children }: { children: ReactNode }) {
    return <span className="rounded border border-cortex-border bg-cortex-surface/70 px-2 py-1">{children}</span>;
}
