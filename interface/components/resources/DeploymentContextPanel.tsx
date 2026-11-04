"use client";

import type { ReactNode } from "react";
import { useEffect, useMemo, useState } from "react";
import { BookOpenText, RefreshCw, ShieldCheck, Upload } from "lucide-react";
import { DeploymentContextEntryList } from "./DeploymentContextEntryList";
import {
    CONTENT_DOMAIN_OPTIONS,
    type DeploymentContextEntry,
    INITIAL_DEPLOYMENT_CONTEXT_FORM,
    INPUT_CLASS,
    KNOWLEDGE_CLASS_OPTIONS,
    type LoadResponse,
    SENSITIVITY_OPTIONS,
    SOURCE_KIND_OPTIONS,
    TRUST_OPTIONS,
    VISIBILITY_OPTIONS,
    applyKnowledgeClassPreset,
    knowledgeClassLabel,
    readFileAsText,
} from "./deploymentContextTypes";

export default function DeploymentContextPanel() {
    const [entries, setEntries] = useState<DeploymentContextEntry[]>([]);
    const [loading, setLoading] = useState(true);
    const [submitting, setSubmitting] = useState(false);
    const [status, setStatus] = useState("Ready");
    const [error, setError] = useState<string | null>(null);
    const [form, setForm] = useState(INITIAL_DEPLOYMENT_CONTEXT_FORM);

    const canSubmit = useMemo(() => {
        return form.title.trim().length > 0 && form.content.trim().length > 0 && !submitting;
    }, [form.title, form.content, submitting]);

    const loadEntries = async () => {
        setLoading(true);
        try {
            const res = await fetch("/api/v1/memory/deployment-context?limit=12");
            const payload = await res.json().catch(async () => ({ error: await res.text() }));
            if (!res.ok) {
                throw new Error(typeof payload?.error === "string" ? payload.error : "Knowledge unavailable.");
            }
            setEntries(Array.isArray(payload.entries) ? payload.entries : []);
            setError(null);
        } catch (err) {
            setEntries([]);
            setError(err instanceof Error ? err.message : "Knowledge unavailable.");
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
        setStatus("Loading governed context into the private/customer/company/reflection knowledge store...");
        try {
            const res = await fetch("/api/v1/memory/deployment-context", {
                method: "POST",
                headers: { "Content-Type": "application/json" },
                body: JSON.stringify({
                    ...form,
                    tags: form.tags.split(",").map((tag) => tag.trim()).filter(Boolean),
                    target_goal_sets: form.target_goal_sets.split(",").map((goal) => goal.trim()).filter(Boolean),
                }),
            });
            const payload = await res.json().catch(async () => ({ error: await res.text() }));
            if (!res.ok) {
                throw new Error(typeof payload?.error === "string" ? payload.error : "Knowledge load failed.");
            }
            const result = payload as LoadResponse;
            const label = knowledgeClassLabel(result.knowledge_class);
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
            setError(err instanceof Error ? err.message : "Knowledge load failed.");
            setStatus("Load failed");
        } finally {
            setSubmitting(false);
        }
    };

    const applyKnowledgeClass = (knowledgeClass: string) => {
        setForm((current) => applyKnowledgeClassPreset(current, knowledgeClass));
    };

    const loadFile = async (file: File | undefined) => {
        if (!file) return;
        const text = await readFileAsText(file);
        setForm((current) => ({
            ...current,
            title: current.title || file.name,
            source_label: current.source_label || file.name,
            content_type: file.type || "text/plain",
            content: text,
        }));
    };

    return (
        <div className="h-full overflow-y-auto bg-cortex-bg">
            <div className="max-w-7xl mx-auto p-6 grid grid-cols-1 xl:grid-cols-[minmax(0,1.1fr)_minmax(0,0.9fr)] gap-6">
                <section className="rounded-2xl border border-cortex-border bg-cortex-surface p-5">
                    <div className="flex items-start justify-between gap-4">
                        <div>
                            <div className="flex items-center gap-2">
                                <BookOpenText className="w-4 h-4 text-cortex-primary" />
                                <h2 className="text-sm font-semibold text-cortex-text-main">Knowledge Intake</h2>
                            </div>
                            <p className="text-xs text-cortex-text-muted mt-2 max-w-2xl">
                                Load governed knowledge into the separate context store Soma uses for RAG. Private records, diary notes, finance references, customer context, company guidance, admin-shaped Soma context, and reflection/synthesis observations stay in distinct lanes with explicit visibility and goal-set scope.
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
                                onChange={(e) => applyKnowledgeClass(e.target.value)}
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
                        <Field label="Content Domain">
                            <select
                                value={form.content_domain}
                                onChange={(e) => setForm((current) => ({ ...current, content_domain: e.target.value }))}
                                className={INPUT_CLASS}
                            >
                                {CONTENT_DOMAIN_OPTIONS.map((option) => (
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

                    <div className="mt-3 grid grid-cols-1 md:grid-cols-2 gap-3">
                        <Field label="Target Goal Sets">
                            <input
                                value={form.target_goal_sets}
                                onChange={(e) => setForm((current) => ({ ...current, target_goal_sets: e.target.value }))}
                                placeholder="tax planning, investor prep, household budget"
                                className={INPUT_CLASS}
                            />
                        </Field>
                        <Field label="Tags">
                            <input
                                value={form.tags}
                                onChange={(e) => setForm((current) => ({ ...current, tags: e.target.value }))}
                                placeholder="finance,records,planning"
                                className={INPUT_CLASS}
                            />
                        </Field>
                    </div>

                    <Field label="Upload Text File" className="mt-3">
                        <input
                            type="file"
                            accept=".txt,.md,.markdown,.csv,.json,text/*,application/json"
                            onChange={(e) => void loadFile(e.target.files?.[0])}
                            className={INPUT_CLASS}
                        />
                    </Field>

                    <Field label="Content" className="mt-3">
                        <textarea
                            value={form.content}
                            onChange={(e) => setForm((current) => ({ ...current, content: e.target.value }))}
                            placeholder="Paste private records, diary notes, finance references, customer docs, approved company guidance, reflection/synthesis observations, security requirements, provider constraints, or other governed context here."
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

                <DeploymentContextEntryList entries={entries} loading={loading} />
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
