"use client";

import type { ReactNode } from "react";
import { useEffect, useMemo, useState } from "react";
import { BookOpenText, RefreshCw, ShieldCheck, Upload } from "lucide-react";

type DeploymentContextEntry = {
    artifact_id: string;
    knowledge_class: string;
    title: string;
    source_label: string;
    source_kind: string;
    visibility: string;
    sensitivity_class: string;
    trust_class: string;
    chunk_count: number;
    vector_count: number;
    content_preview: string;
    content_length: number;
    created_at: string;
};

type LoadResponse = {
    artifact_id: string;
    knowledge_class: string;
    title: string;
    chunk_count: number;
    vector_count: number;
};

const KNOWLEDGE_CLASS_OPTIONS = [
    { value: "customer_context", label: "Customer context" },
    { value: "company_knowledge", label: "Approved company knowledge" },
];

const SOURCE_KIND_OPTIONS = [
    { value: "user_document", label: "User document" },
    { value: "user_note", label: "User note" },
    { value: "workspace_file", label: "Workspace file" },
    { value: "web_research", label: "Web research" },
];

const VISIBILITY_OPTIONS = [
    { value: "global", label: "Global" },
    { value: "team", label: "Team" },
    { value: "private", label: "Private" },
];

const SENSITIVITY_OPTIONS = [
    { value: "role_scoped", label: "Role scoped" },
    { value: "team_scoped", label: "Team scoped" },
    { value: "restricted", label: "Restricted" },
];

const TRUST_OPTIONS = [
    { value: "user_provided", label: "User provided" },
    { value: "validated_external", label: "Validated external" },
    { value: "bounded_external", label: "Bounded external" },
    { value: "trusted_internal", label: "Trusted internal" },
];

const INPUT_CLASS = "w-full rounded-lg border border-cortex-border bg-cortex-bg px-3 py-2 text-sm text-cortex-text-main placeholder:text-cortex-text-muted/60 focus:outline-none focus:border-cortex-primary";

export default function DeploymentContextPanel() {
    const [entries, setEntries] = useState<DeploymentContextEntry[]>([]);
    const [loading, setLoading] = useState(true);
    const [submitting, setSubmitting] = useState(false);
    const [status, setStatus] = useState("Ready");
    const [error, setError] = useState<string | null>(null);
    const [form, setForm] = useState({
        knowledge_class: "customer_context",
        title: "",
        source_label: "",
        content: "",
        source_kind: "user_document",
        visibility: "global",
        sensitivity_class: "role_scoped",
        trust_class: "user_provided",
        tags: "deployment,context",
    });

    const canSubmit = useMemo(() => {
        return form.title.trim().length > 0 && form.content.trim().length > 0 && !submitting;
    }, [form.title, form.content, submitting]);

    const loadEntries = async () => {
        setLoading(true);
        try {
            const res = await fetch("/api/v1/memory/deployment-context?limit=12");
            const payload = await res.json().catch(async () => ({ error: await res.text() }));
            if (!res.ok) {
                throw new Error(typeof payload?.error === "string" ? payload.error : "Deployment context unavailable.");
            }
            setEntries(Array.isArray(payload.entries) ? payload.entries : []);
            setError(null);
        } catch (err) {
            setEntries([]);
            setError(err instanceof Error ? err.message : "Deployment context unavailable.");
        } finally {
            setLoading(false);
        }
    };

    useEffect(() => {
        loadEntries();
    }, []);

    const submit = async () => {
        if (!canSubmit) return;
        setSubmitting(true);
        setStatus("Loading governed context into the customer/company knowledge store...");
        try {
            const res = await fetch("/api/v1/memory/deployment-context", {
                method: "POST",
                headers: { "Content-Type": "application/json" },
                body: JSON.stringify({
                    ...form,
                    tags: form.tags.split(",").map((tag) => tag.trim()).filter(Boolean),
                }),
            });
            const payload = await res.json().catch(async () => ({ error: await res.text() }));
            if (!res.ok) {
                throw new Error(typeof payload?.error === "string" ? payload.error : "Deployment context load failed.");
            }
            const result = payload as LoadResponse;
            const label = result.knowledge_class === "company_knowledge" ? "approved company knowledge" : "customer context";
            setStatus(`Loaded ${result.title} as ${label} into ${result.vector_count} vectors across ${result.chunk_count} chunks.`);
            setForm((current) => ({
                ...current,
                title: "",
                source_label: "",
                content: "",
            }));
            setError(null);
            await loadEntries();
        } catch (err) {
            setError(err instanceof Error ? err.message : "Deployment context load failed.");
            setStatus("Load failed");
        } finally {
            setSubmitting(false);
        }
    };

    return (
        <div className="h-full overflow-y-auto bg-cortex-bg">
            <div className="max-w-7xl mx-auto p-6 grid grid-cols-1 xl:grid-cols-[minmax(0,1.1fr)_minmax(0,0.9fr)] gap-6">
                <section className="rounded-2xl border border-cortex-border bg-cortex-surface p-5">
                    <div className="flex items-start justify-between gap-4">
                        <div>
                            <div className="flex items-center gap-2">
                                <BookOpenText className="w-4 h-4 text-cortex-primary" />
                                <h2 className="text-sm font-semibold text-cortex-text-main">Deployment Context Intake</h2>
                            </div>
                            <p className="text-xs text-cortex-text-muted mt-2 max-w-2xl">
                                Load governed knowledge into the separate context store Soma uses for deployment RAG. Keep customer-provided context separate from approved company knowledge.
                            </p>
                        </div>
                        <button
                            onClick={loadEntries}
                            className="inline-flex items-center gap-1.5 px-3 py-1.5 rounded-lg border border-cortex-border text-xs font-mono text-cortex-text-main hover:bg-cortex-bg"
                        >
                            <RefreshCw className={`w-3.5 h-3.5 ${loading ? "animate-spin" : ""}`} />
                            Refresh
                        </button>
                    </div>

                    <div className="mt-5 grid grid-cols-1 md:grid-cols-2 gap-3">
                        <Field label="Title">
                            <input
                                value={form.title}
                                onChange={(e) => setForm((current) => ({ ...current, title: e.target.value }))}
                                placeholder="Deployment architecture brief"
                                className={INPUT_CLASS}
                            />
                        </Field>
                        <Field label="Source Label">
                            <input
                                value={form.source_label}
                                onChange={(e) => setForm((current) => ({ ...current, source_label: e.target.value }))}
                                placeholder="customer handoff doc"
                                className={INPUT_CLASS}
                            />
                        </Field>
                        <Field label="Knowledge Class">
                            <select
                                value={form.knowledge_class}
                                onChange={(e) => setForm((current) => ({ ...current, knowledge_class: e.target.value }))}
                                className={INPUT_CLASS}
                            >
                                {KNOWLEDGE_CLASS_OPTIONS.map((option) => (
                                    <option key={option.value} value={option.value}>{option.label}</option>
                                ))}
                            </select>
                        </Field>
                        <Field label="Source Kind">
                            <select
                                value={form.source_kind}
                                onChange={(e) => setForm((current) => ({ ...current, source_kind: e.target.value }))}
                                className={INPUT_CLASS}
                            >
                                {SOURCE_KIND_OPTIONS.map((option) => (
                                    <option key={option.value} value={option.value}>{option.label}</option>
                                ))}
                            </select>
                        </Field>
                        <Field label="Visibility">
                            <select
                                value={form.visibility}
                                onChange={(e) => setForm((current) => ({ ...current, visibility: e.target.value }))}
                                className={INPUT_CLASS}
                            >
                                {VISIBILITY_OPTIONS.map((option) => (
                                    <option key={option.value} value={option.value}>{option.label}</option>
                                ))}
                            </select>
                        </Field>
                        <Field label="Sensitivity">
                            <select
                                value={form.sensitivity_class}
                                onChange={(e) => setForm((current) => ({ ...current, sensitivity_class: e.target.value }))}
                                className={INPUT_CLASS}
                            >
                                {SENSITIVITY_OPTIONS.map((option) => (
                                    <option key={option.value} value={option.value}>{option.label}</option>
                                ))}
                            </select>
                        </Field>
                        <Field label="Trust Class">
                            <select
                                value={form.trust_class}
                                onChange={(e) => setForm((current) => ({ ...current, trust_class: e.target.value }))}
                                className={INPUT_CLASS}
                            >
                                {TRUST_OPTIONS.map((option) => (
                                    <option key={option.value} value={option.value}>{option.label}</option>
                                ))}
                            </select>
                        </Field>
                    </div>

                    <Field label="Tags" className="mt-3">
                        <input
                            value={form.tags}
                            onChange={(e) => setForm((current) => ({ ...current, tags: e.target.value }))}
                            placeholder="deployment,security,mcp"
                            className={INPUT_CLASS}
                        />
                    </Field>

                    <Field label="Content" className="mt-3">
                        <textarea
                            value={form.content}
                            onChange={(e) => setForm((current) => ({ ...current, content: e.target.value }))}
                            placeholder="Paste customer docs, deployment notes, approved company guidance, security requirements, provider constraints, or other governed context here."
                            className={`${INPUT_CLASS} min-h-[240px] resize-y`}
                        />
                    </Field>

                    <div className="mt-4 flex items-center justify-between gap-3 rounded-xl border border-cortex-border/70 bg-cortex-bg/70 px-4 py-3">
                        <div className="min-w-0">
                            <div className="flex items-center gap-2 text-cortex-text-main">
                                <ShieldCheck className="w-4 h-4 text-cortex-success" />
                                <span className="text-xs font-semibold">Governed knowledge store</span>
                            </div>
                            <p className="text-[11px] text-cortex-text-muted mt-1">
                                Stored as an approved document artifact plus governed pgvector chunks with visibility, sensitivity, trust, and knowledge-class metadata.
                            </p>
                            <p className="text-[11px] font-mono text-cortex-text-muted mt-1">{status}</p>
                        </div>
                        <button
                            onClick={submit}
                            disabled={!canSubmit}
                            className="inline-flex items-center gap-2 px-4 py-2 rounded-lg bg-cortex-primary text-white text-sm font-semibold disabled:opacity-50 disabled:cursor-not-allowed hover:brightness-105"
                        >
                            <Upload className="w-4 h-4" />
                            Load Context
                        </button>
                    </div>

                    {error ? (
                        <div className="mt-3 rounded-xl border border-cortex-danger/30 bg-cortex-danger/5 px-4 py-3">
                            <p className="text-sm text-cortex-danger">Context load error</p>
                            <p className="text-xs text-cortex-text-muted mt-1">{error}</p>
                        </div>
                    ) : null}
                </section>

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
                            <p className="text-xs font-mono text-cortex-text-muted animate-pulse">Loading deployment context…</p>
                        ) : entries.length === 0 ? (
                            <div className="rounded-xl border border-dashed border-cortex-border p-4 text-xs text-cortex-text-muted">
                                No deployment context loaded yet.
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
                                        <Badge>{entry.chunk_count} chunks</Badge>
                                    </div>
                                </article>
                            ))
                        )}
                    </div>
                </section>
            </div>
        </div>
    );
}

function Field({ label, children, className = "" }: { label: string; children: ReactNode; className?: string }) {
    return (
        <label className={`block ${className}`}>
            <span className="block text-[11px] font-mono uppercase tracking-widest text-cortex-text-muted mb-1.5">{label}</span>
            {children}
        </label>
    );
}

function Badge({ children }: { children: ReactNode }) {
    return (
        <span className="px-2 py-1 rounded border border-cortex-border bg-cortex-surface/70">
            {children}
        </span>
    );
}
