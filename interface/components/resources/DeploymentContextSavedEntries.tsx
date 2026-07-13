import type { ReactNode } from "react";
import { MessageSquareText } from "lucide-react";
import { requestSomaOutputContinuation } from "@/components/soma/outputContinuation";
import type { DeploymentContextEntry } from "./DeploymentContextPanel";

export default function DeploymentContextSavedEntries({
    entries,
    loading,
}: {
    entries: DeploymentContextEntry[];
    loading: boolean;
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
                <span className="text-[11px] font-mono text-cortex-text-muted">{entries.length} entries</span>
            </div>

            <div className="min-h-0 flex-1 space-y-3 overflow-y-auto p-5" role="region" aria-label="Loaded governed context list">
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
                                        {entry.source_label} · {entry.source_kind.replaceAll("_", " ")}
                                    </p>
                                </div>
                                <span className="rounded bg-cortex-primary/10 px-2 py-1 text-[10px] font-mono uppercase text-cortex-primary">
                                    {entry.vector_count} vectors
                                </span>
                            </div>
                            <p className="text-sm leading-relaxed text-cortex-text-main">{entry.content_preview}</p>
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
                            <button
                                type="button"
                                onClick={() => askSomaWithContext(entry)}
                                className="inline-flex items-center gap-1.5 rounded border border-cortex-primary/40 px-2.5 py-1.5 text-xs font-semibold text-cortex-primary hover:bg-cortex-primary/10"
                            >
                                <MessageSquareText className="h-3.5 w-3.5" />
                                Ask Soma with this
                            </button>
                        </article>
                    ))
                )}
            </div>
        </section>
    );
}

function Badge({ children }: { children: ReactNode }) {
    return <span className="rounded border border-cortex-border bg-cortex-surface/70 px-2 py-1">{children}</span>;
}
