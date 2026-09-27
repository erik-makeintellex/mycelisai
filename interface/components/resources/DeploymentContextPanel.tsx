"use client";

import type { ReactNode } from "react";
import { useEffect, useMemo, useRef, useState } from "react";
import { BookOpenText, ChevronDown, RefreshCw, Upload } from "lucide-react";
import DeploymentContextSavedEntries from "./DeploymentContextSavedEntries";
import DeploymentContextSaveStatus, { savedStateFrom, type SaveState } from "./DeploymentContextSaveStatus";
import {
    CONTENT_DOMAIN_OPTIONS,
    KNOWLEDGE_CLASS_OPTIONS,
    SENSITIVITY_OPTIONS,
    SOURCE_KIND_OPTIONS,
    TRUST_OPTIONS,
    VISIBILITY_OPTIONS,
    defaultForm,
    splitList,
    withKnowledgeClass,
    type DeploymentContextForm,
    type SelectOption,
} from "./DeploymentContextOptions";

export type DeploymentContextEntry = {
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
    content_domain?: string;
    target_goal_sets?: string[];
    created_at: string;
    embedding_status?: string;
    lifecycle_state?: string;
    archived_at?: string;
    can_manage?: boolean;
};

const INPUT_CLASS = "w-full rounded-lg border border-cortex-border bg-cortex-bg px-3 py-2 text-sm text-cortex-text-main placeholder:text-cortex-text-muted/60 focus:outline-none focus:border-cortex-primary";

export default function DeploymentContextPanel() {
    const [entries, setEntries] = useState<DeploymentContextEntry[]>([]);
    const [loading, setLoading] = useState(true);
    const [listError, setListError] = useState<string | null>(null);
    const [showArchived, setShowArchived] = useState(false);
    const [viewerIsAdmin, setViewerIsAdmin] = useState(false);
    const [saveState, setSaveState] = useState<SaveState>({ kind: "idle" });
    const [showMore, setShowMore] = useState(false);
    const [form, setForm] = useState<DeploymentContextForm>(() => defaultForm(false));
    const classTouched = useRef(false);

    const canSubmit = useMemo(
        () => form.title.trim().length > 0 && form.content.trim().length > 0 && saveState.kind !== "saving",
        [form.title, form.content, saveState.kind],
    );
    const update = (patch: Partial<DeploymentContextForm>) => setForm((current) => ({ ...current, ...patch }));

    const loadEntries = async (archived = showArchived) => {
        setLoading(true);
        try {
            const res = await fetch(`/api/v1/memory/deployment-context?limit=12${archived ? "&include_archived=true" : ""}`);
            const payload = await res.json().catch(() => ({}));
            if (!res.ok) throw new Error(typeof payload?.error === "string" ? payload.error : "Saved context is unavailable.");
            setEntries(Array.isArray(payload.entries) ? payload.entries : []);
            setListError(null);
        } catch (err) {
            setEntries([]);
            setListError(err instanceof Error ? err.message : "Saved context is unavailable.");
        } finally {
            setLoading(false);
        }
    };

    useEffect(() => {
        void loadEntries();
        let cancelled = false;
        fetch("/auth/session", { cache: "no-store" })
            .then((res) => (res.ok ? res.json() : null))
            .then((body) => {
                if (!cancelled && body?.data?.user?.role === "admin") setViewerIsAdmin(true);
                // Admin defaults apply only until the user picks a class.
                if (!cancelled && body?.data?.user?.role === "admin" && !classTouched.current) {
                    const admin = defaultForm(true);
                    setForm((current) => ({ ...current, knowledge_class: admin.knowledge_class, trust_class: admin.trust_class }));
                }
            })
            .catch(() => undefined);
        return () => {
            cancelled = true;
        };
    }, []);

    const submit = async () => {
        if (!canSubmit) return;
        setSaveState({ kind: "saving" });
        try {
            const res = await fetch("/api/v1/memory/deployment-context", {
                method: "POST",
                headers: { "Content-Type": "application/json" },
                body: JSON.stringify({ ...form, tags: splitList(form.tags), target_goal_sets: splitList(form.target_goal_sets) }),
            });
            const payload = await res.json().catch(() => ({}));
            if (!res.ok) {
                throw new Error(typeof payload?.error === "string" ? payload.error : `Saving failed (HTTP ${res.status}).`);
            }
            setSaveState(savedStateFrom(payload));
            setForm((current) => ({ ...current, title: "", source_label: "", content: "" }));
        } catch (err) {
            setSaveState({ kind: "error", message: err instanceof Error ? err.message : "Saving failed." });
        } finally {
            await loadEntries();
        }
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
        <div className="h-full min-h-0 overflow-hidden bg-cortex-bg">
            <div className="mx-auto grid h-full min-h-0 max-w-7xl grid-cols-1 gap-4 p-4 xl:grid-cols-[minmax(0,1.08fr)_minmax(20rem,0.92fr)]">
                <section className="flex min-h-0 flex-col overflow-hidden rounded-lg border border-cortex-border bg-cortex-surface">
                    <div className="flex items-start justify-between gap-4 px-5 pt-5">
                        <div className="min-w-0">
                            <div className="flex items-center gap-2">
                                <BookOpenText className="h-4 w-4 text-cortex-primary" />
                                <h2 className="text-sm font-semibold text-cortex-text-main">Long-term context for Soma</h2>
                            </div>
                            <p className="mt-2 max-w-2xl text-xs text-cortex-text-muted">
                                Save notes or files Soma should use in later chats. Deliverables hold finished results; this holds reusable facts.
                            </p>
                        </div>
                        <button onClick={() => void loadEntries()} className="inline-flex items-center gap-1.5 rounded-lg border border-cortex-border px-3 py-1.5 text-xs text-cortex-text-main hover:bg-cortex-bg">
                            <RefreshCw className={`h-3.5 w-3.5 ${loading ? "animate-spin" : ""}`} />
                            Refresh
                        </button>
                    </div>

                    <div className="min-h-0 flex-1 space-y-3 overflow-y-auto px-5 py-4">
                        <Field label="Title">
                            <input value={form.title} onChange={(e) => update({ title: e.target.value })} placeholder="Company profile" className={INPUT_CLASS} />
                        </Field>
                        <Field label="Content">
                            <textarea
                                value={form.content}
                                onChange={(e) => update({ content: e.target.value })}
                                placeholder="Paste the facts Soma should remember, such as products, prices, tone, or customer requirements."
                                className={`${INPUT_CLASS} min-h-[200px] resize-y`}
                            />
                        </Field>
                        <Field label="Or upload a text file">
                            <input type="file" accept=".txt,.md,.markdown,.csv,.json,text/*,application/json" onChange={(e) => void loadFile(e.target.files?.[0])} className={INPUT_CLASS} />
                        </Field>
                        <fieldset>
                            <legend className="mb-1.5 text-[11px] font-mono uppercase tracking-widest text-cortex-text-muted">Who can use this</legend>
                            <div role="radiogroup" aria-label="Who can use this" className="flex flex-wrap gap-2">
                                {VISIBILITY_OPTIONS.map((option) => (
                                    <label key={option.value} className={`cursor-pointer rounded-lg border px-3 py-1.5 text-xs ${form.visibility === option.value ? "border-cortex-primary bg-cortex-primary/10 text-cortex-text-main" : "border-cortex-border text-cortex-text-muted"}`}>
                                        <input type="radio" name="deployment-context-visibility" className="sr-only" checked={form.visibility === option.value} onChange={() => update({ visibility: option.value })} />
                                        {option.label}
                                    </label>
                                ))}
                            </div>
                        </fieldset>

                        <button type="button" aria-expanded={showMore} onClick={() => setShowMore((open) => !open)} className="inline-flex items-center gap-1 text-xs font-semibold text-cortex-primary">
                            <ChevronDown className={`h-3.5 w-3.5 transition-transform ${showMore ? "rotate-180" : ""}`} />
                            More options
                        </button>
                        {showMore ? (
                            <div className="grid grid-cols-1 gap-3 md:grid-cols-2">
                                <SelectField label="Use as" value={form.knowledge_class} options={KNOWLEDGE_CLASS_OPTIONS} onChange={(value) => { classTouched.current = true; setForm((current) => withKnowledgeClass(current, value)); }} />
                                <SelectField label="Source kind" value={form.source_kind} options={SOURCE_KIND_OPTIONS} onChange={(value) => update({ source_kind: value })} />
                                <SelectField label="Topic" value={form.content_domain} options={CONTENT_DOMAIN_OPTIONS} onChange={(value) => update({ content_domain: value })} />
                                <SelectField label="Sensitivity" value={form.sensitivity_class} options={SENSITIVITY_OPTIONS} onChange={(value) => update({ sensitivity_class: value })} />
                                <SelectField label="Trust" value={form.trust_class} options={TRUST_OPTIONS} onChange={(value) => update({ trust_class: value })} />
                                <Field label="Source label">
                                    <input value={form.source_label} onChange={(e) => update({ source_label: e.target.value })} placeholder="owner notes" className={INPUT_CLASS} />
                                </Field>
                                <Field label="Target goal sets">
                                    <input value={form.target_goal_sets} onChange={(e) => update({ target_goal_sets: e.target.value })} placeholder="spring launch, tax planning" className={INPUT_CLASS} />
                                </Field>
                                <Field label="Tags">
                                    <input value={form.tags} onChange={(e) => update({ tags: e.target.value })} placeholder="menu, pricing" className={INPUT_CLASS} />
                                </Field>
                            </div>
                        ) : null}
                    </div>

                    <div className="flex flex-shrink-0 flex-col gap-3 border-t border-cortex-border bg-cortex-bg/70 px-5 py-4">
                        <DeploymentContextSaveStatus state={saveState} />
                        <button
                            onClick={submit}
                            disabled={!canSubmit}
                            className="inline-flex items-center justify-center gap-2 self-end rounded-lg bg-cortex-primary px-4 py-2 text-sm font-semibold text-white hover:brightness-105 disabled:cursor-not-allowed disabled:opacity-50"
                        >
                            <Upload className="h-4 w-4" />
                            Save context
                        </button>
                    </div>
                </section>

                <DeploymentContextSavedEntries
                    entries={entries}
                    loading={loading}
                    error={listError}
                    viewerIsAdmin={viewerIsAdmin}
                    showArchived={showArchived}
                    onShowArchivedChange={(next) => {
                        setShowArchived(next);
                        void loadEntries(next);
                    }}
                    onChanged={() => loadEntries()}
                />
            </div>
        </div>
    );
}

function readFileAsText(file: File): Promise<string> {
    if (typeof file.text === "function") {
        return file.text();
    }
    return new Promise((resolve, reject) => {
        const reader = new FileReader();
        reader.onload = () => resolve(typeof reader.result === "string" ? reader.result : "");
        reader.onerror = () => reject(reader.error ?? new Error("Unable to read uploaded file."));
        reader.readAsText(file);
    });
}

function Field({ label, children }: { label: string; children: ReactNode }) {
    return (
        <label className="block">
            <span className="mb-1.5 block text-[11px] font-mono uppercase tracking-widest text-cortex-text-muted">{label}</span>
            {children}
        </label>
    );
}

function SelectField({ label, value, options, onChange }: { label: string; value: string; options: SelectOption[]; onChange: (value: string) => void }) {
    return (
        <Field label={label}>
            <select value={value} onChange={(e) => onChange(e.target.value)} className={INPUT_CLASS}>
                {options.map((option) => <option key={option.value} value={option.value}>{option.label}</option>)}
            </select>
        </Field>
    );
}
