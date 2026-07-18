"use client";

import { RadioTower, RefreshCw } from "lucide-react";
import type { InputSource, InputSourceBufferView } from "./InputSourceRegistry";

export function InputSourceRegistryCard({
    bufferError,
    bufferView,
    isFetchingBuffer,
    isLoading,
    registrySupported,
    selectedSourceId,
    sources,
    sourcesError,
    onRefresh,
    onSelectSource,
}: {
    bufferError: string | null;
    bufferView: InputSourceBufferView | null;
    isFetchingBuffer: boolean;
    isLoading: boolean;
    registrySupported: boolean;
    selectedSourceId: string | null;
    sources: InputSource[];
    sourcesError: string | null;
    onRefresh: () => void;
    onSelectSource: (sourceId: string, mode?: string) => void;
}) {
    const selected = sources.find((source) => source.id === selectedSourceId) ?? sources[0] ?? null;

    return (
        <section className="rounded-xl border border-cortex-border bg-cortex-surface px-4 py-4">
            <div className="flex flex-col gap-3 sm:flex-row sm:items-start sm:justify-between">
                <div className="flex items-start gap-3">
                    <div className="mt-0.5 rounded-lg border border-cortex-primary/25 bg-cortex-primary/10 p-2">
                        <RadioTower className="h-4 w-4 text-cortex-primary" />
                    </div>
                    <div>
                        <p className="text-[10px] font-mono font-bold uppercase tracking-wider text-cortex-text-muted">
                            Live inputs
                        </p>
                        <p className="mt-1 text-sm font-semibold text-cortex-text-main">
                            {sourceSummary(sources, registrySupported)}
                        </p>
                        <p className="mt-1 text-xs leading-5 text-cortex-text-muted">
                            Service, device, webhook, or private-system events that Soma can reference from a bounded buffer.
                        </p>
                    </div>
                </div>
                <button
                    type="button"
                    onClick={onRefresh}
                    className="inline-flex shrink-0 items-center justify-center gap-2 rounded-lg border border-cortex-border bg-cortex-bg px-3 py-2 text-xs font-semibold text-cortex-text-muted transition hover:text-cortex-text-main"
                >
                    <RefreshCw className={`h-3.5 w-3.5 ${isLoading ? "animate-spin" : ""}`} />
                    Refresh
                </button>
            </div>

            {sourcesError && <Notice tone="warning" message={sourcesError} />}
            {!registrySupported && !isLoading && sources.length === 0 && (
                <Notice
                    tone="muted"
                    message="Live input registry is not available from this interface yet."
                />
            )}

            {registrySupported && sources.length === 0 && (
                <div className="mt-4 rounded-lg border border-cortex-border bg-cortex-bg/70 px-3 py-3">
                    <p className="text-sm font-semibold text-cortex-text-main">No live inputs registered yet.</p>
                    <p className="mt-1 text-xs leading-5 text-cortex-text-muted">
                        Add one when an external service, private API, device, scheduler, or workflow feed should inform an Outcome. Soma reads the buffer, not raw traffic.
                    </p>
                </div>
            )}

            {sources.length > 0 && (
                <div className="mt-4 grid gap-3 lg:grid-cols-[minmax(0,1fr)_minmax(280px,0.9fr)]">
                    <div className="space-y-2">
                        {sources.slice(0, 5).map((source) => (
                            <button
                                key={source.id}
                                type="button"
                                onClick={() => onSelectSource(source.id, source.buffer_mode)}
                                className={`w-full rounded-lg border px-3 py-2 text-left transition ${
                                    selected?.id === source.id
                                        ? "border-cortex-primary/40 bg-cortex-primary/10"
                                        : "border-cortex-border bg-cortex-bg hover:border-cortex-primary/25"
                                }`}
                            >
                                <div className="flex items-center justify-between gap-2">
                                    <span className="truncate text-sm font-semibold text-cortex-text-main">{source.name}</span>
                                    <StatusPill status={source.status} />
                                </div>
                                <p className="mt-1 text-xs text-cortex-text-muted">
                                    {labelize(source.adapter_kind)} · {bufferLabel(source.buffer_mode)} · {scopeLabel(source)}
                                </p>
                            </button>
                        ))}
                        {sources.length > 5 && (
                            <p className="px-1 text-[11px] text-cortex-text-muted">
                                {sources.length - 5} more live inputs are available through Inspect.
                            </p>
                        )}
                    </div>

                    <div className="rounded-lg border border-cortex-border bg-cortex-bg/70 px-3 py-3">
                        <div className="flex items-center justify-between gap-2">
                            <p className="text-[10px] font-mono font-bold uppercase tracking-wider text-cortex-text-muted">
                                Buffered context
                            </p>
                            {selected && (
                                <button
                                    type="button"
                                    onClick={() => onSelectSource(selected.id, selected.buffer_mode)}
                                    className="text-[11px] font-semibold text-cortex-primary hover:text-cortex-primary/80"
                                >
                                    Preview
                                </button>
                            )}
                        </div>
                        {!selected && (
                            <p className="mt-2 text-xs leading-5 text-cortex-text-muted">
                                Select a live input to preview its latest buffered context.
                            </p>
                        )}
                        {selected && !bufferView && !isFetchingBuffer && !bufferError && (
                            <p className="mt-2 text-xs leading-5 text-cortex-text-muted">
                                Preview the latest buffered events before routing this source into Soma or a team.
                            </p>
                        )}
                        {isFetchingBuffer && <p className="mt-2 text-xs text-cortex-text-muted">Loading buffer preview...</p>}
                        {bufferError && <Notice tone="warning" message={bufferError} />}
                        {bufferView && <BufferPreview view={bufferView} />}
                        {selected && (
                            <details className="mt-3 rounded-lg border border-cortex-border bg-cortex-surface/60 px-3 py-2">
                                <summary className="cursor-pointer text-[10px] font-mono uppercase tracking-wider text-cortex-text-muted">
                                    Inspect source
                                </summary>
                                <dl className="mt-2 grid gap-2 text-[11px] text-cortex-text-muted">
                                    <div>
                                        <dt className="font-semibold text-cortex-text-main">Ingress ref</dt>
                                        <dd className="break-all">{selected.allowed_ingress_subject || "Not assigned"}</dd>
                                    </div>
                                    <div>
                                        <dt className="font-semibold text-cortex-text-main">Auth</dt>
                                        <dd>{selected.auth_scheme}{selected.secret_ref ? ` · ${selected.secret_ref}` : ""}</dd>
                                    </div>
                                </dl>
                            </details>
                        )}
                    </div>
                </div>
            )}
        </section>
    );
}

function BufferPreview({ view }: { view: InputSourceBufferView }) {
    const rows = view.latest?.length
        ? view.latest.map((item) => ({ key: item.event_id || item.channel_key, title: item.channel_key, detail: previewPayload(item.payload) }))
        : view.windows?.length
            ? view.windows.map((item) => ({ key: item.window_key, title: item.channel_key, detail: item.summary || `${item.count ?? 0} events` }))
            : (view.events ?? []).map((item) => ({ key: item.event_id || item.channel_key, title: item.channel_key, detail: previewPayload(item.payload) }));

    if (rows.length === 0) {
        return <p className="mt-2 text-xs leading-5 text-cortex-text-muted">No buffered events yet.</p>;
    }

    return (
        <div className="mt-2 space-y-2">
            {rows.slice(0, 3).map((row) => (
                <div key={row.key} className="rounded-lg border border-cortex-border bg-cortex-surface px-3 py-2">
                    <p className="truncate text-xs font-semibold text-cortex-text-main">{row.title}</p>
                    <p className="mt-1 line-clamp-2 text-[11px] leading-4 text-cortex-text-muted">{row.detail}</p>
                </div>
            ))}
        </div>
    );
}

function Notice({ message, tone }: { message: string; tone: "muted" | "warning" }) {
    const cls = tone === "warning"
        ? "border-cortex-warning/25 bg-cortex-warning/10 text-cortex-warning"
        : "border-cortex-border bg-cortex-bg/70 text-cortex-text-muted";
    return <p className={`mt-3 rounded-lg border px-3 py-2 text-xs ${cls}`}>{message}</p>;
}

function StatusPill({ status }: { status: string }) {
    const ready = status === "available";
    return (
        <span className={`rounded-full border px-2 py-0.5 text-[10px] font-mono uppercase ${
            ready ? "border-cortex-success/25 bg-cortex-success/10 text-cortex-success" : "border-cortex-warning/25 bg-cortex-warning/10 text-cortex-warning"
        }`}>
            {ready ? "Ready" : "Needs attention"}
        </span>
    );
}

function sourceSummary(sources: InputSource[], supported: boolean) {
    if (!supported && sources.length === 0) return "Live input registry unavailable";
    if (sources.length === 0) return "No event feeds yet";
    return `${sources.length} event feed${sources.length === 1 ? "" : "s"} Soma can reference`;
}

function scopeLabel(source: InputSource) {
    if (source.scope_kind === "all") return "All work";
    if (source.scope_kind === "group") return "Group scoped";
    if (source.scope_kind === "host") return "Host scoped";
    return labelize(source.scope_kind);
}

function bufferLabel(mode: string) {
    if (mode === "latest_state") return "Latest state";
    if (mode === "append_with_latest") return "Event log + latest";
    if (mode === "windowed_rollup") return "Windowed summary";
    return "Event log";
}

function labelize(value: string) {
    return value.replaceAll("_", " ").replace(/\b\w/g, (char) => char.toUpperCase());
}

function previewPayload(value: unknown) {
    if (value == null) return "No payload";
    if (typeof value === "string") return value;
    try {
        return JSON.stringify(value);
    } catch {
        return String(value);
    }
}
