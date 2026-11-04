import type { ReactNode } from "react";
import type { DeploymentContextEntry } from "./deploymentContextTypes";

export function DeploymentContextEntryList({
    entries,
    loading,
}: {
    entries: DeploymentContextEntry[];
    loading: boolean;
}) {
    return (
        <section className="rounded-2xl border border-cortex-border bg-cortex-surface p-5">
            <div className="flex items-center justify-between gap-3">
                <div>
                    <h2 className="text-sm font-semibold text-cortex-text-main">Loaded Governed Context</h2>
                    <p className="text-xs text-cortex-text-muted mt-1">
                        Recent governed context entries Soma can recall separately from Soma memory during planning and answer generation.
                    </p>
                </div>
                <span className="text-[11px] font-mono text-cortex-text-muted">{entries.length} entries</span>
            </div>

            <div className="mt-4 space-y-3">
                {loading ? (
                    <p className="text-xs font-mono text-cortex-text-muted animate-pulse">Loading knowledge…</p>
                ) : entries.length === 0 ? (
                    <div className="rounded-xl border border-dashed border-cortex-border p-4 text-xs text-cortex-text-muted">
                        No knowledge loaded yet.
                    </div>
                ) : (
                    entries.map((entry) => (
                        <article key={entry.artifact_id} className="rounded-xl border border-cortex-border bg-cortex-bg/60 p-4 space-y-2">
                            <div className="flex items-start justify-between gap-3">
                                <div>
                                    <h3 className="text-sm font-semibold text-cortex-text-main">{entry.title}</h3>
                                    <p className="text-[11px] text-cortex-text-muted mt-1">
                                        {entry.source_label} · {entry.source_kind.replaceAll("_", " ")}
                                    </p>
                                </div>
                                <span className="text-[10px] font-mono uppercase px-2 py-1 rounded bg-cortex-primary/10 text-cortex-primary">
                                    {entry.vector_count} vectors
                                </span>
                            </div>
                            <p className="text-sm text-cortex-text-main leading-relaxed">{entry.content_preview}</p>
                            <div className="flex flex-wrap gap-2 text-[10px] font-mono text-cortex-text-muted">
                                <Badge>{entry.knowledge_class.replaceAll("_", " ")}</Badge>
                                <Badge>{entry.visibility}</Badge>
                                <Badge>{entry.sensitivity_class}</Badge>
                                <Badge>{entry.trust_class}</Badge>
                                {entry.content_domain ? <Badge>{entry.content_domain.replaceAll("_", " ")}</Badge> : null}
                                {(entry.target_goal_sets ?? []).map((goal) => (
                                    <Badge key={`${entry.artifact_id}-${goal}`}>goal: {goal}</Badge>
                                ))}
                                <Badge>{entry.chunk_count} chunks</Badge>
                            </div>
                        </article>
                    ))
                )}
            </div>
        </section>
    );
}

function Badge({ children }: { children: ReactNode }) {
    return (
        <span className="px-2 py-1 rounded border border-cortex-border bg-cortex-surface/70">
            {children}
        </span>
    );
}
