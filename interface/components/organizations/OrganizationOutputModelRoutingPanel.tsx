"use client";

import { Loader2, Sparkles } from "lucide-react";
import type {
    OrganizationOutputModelRoutingMode,
    OrganizationOutputModelRoutingPayload,
    OrganizationOutputTypeId,
} from "@/lib/organizations";
import {
    OUTPUT_TYPE_OPTIONS,
    outputModelOptions,
} from "@/components/organizations/organizationProfileOptions";

export function OutputModelRoutingPanel({
    routing,
    loading,
    error,
    selectedRoutingMode,
    selectedDefaultModelId,
    selectedBindings,
    updatePending,
    updateError,
    onRoutingModeChange,
    onDefaultModelChange,
    onBindingChange,
    onSubmit,
}: {
    routing: OrganizationOutputModelRoutingPayload | null;
    loading: boolean;
    error: string | null;
    selectedRoutingMode: OrganizationOutputModelRoutingMode;
    selectedDefaultModelId: string;
    selectedBindings: Record<string, string>;
    updatePending: boolean;
    updateError: string | null;
    onRoutingModeChange: (mode: OrganizationOutputModelRoutingMode) => void;
    onDefaultModelChange: (modelId: string) => void;
    onBindingChange: (outputTypeId: OrganizationOutputTypeId, modelId: string) => void;
    onSubmit: () => void;
}) {
    const modelOptions = outputModelOptions(routing);

    return (
        <div className="mt-5 rounded-3xl border border-cortex-primary/20 bg-cortex-bg p-5">
            <div className="flex flex-col gap-3 lg:flex-row lg:items-start lg:justify-between">
                <div>
                    <h4 className="text-lg font-semibold text-cortex-text-main">Output model routing</h4>
                    <p className="mt-2 max-w-3xl text-sm leading-7 text-cortex-text-muted">
                        Admins can keep every Team Lead and Specialist on one concrete model by default, or let Soma route detected output types like research, code, and vision to more appropriate self-hosted models.
                    </p>
                </div>
                <div className="rounded-2xl border border-cortex-border bg-cortex-surface px-4 py-3 text-sm text-cortex-text-muted">
                    <p className="font-medium text-cortex-text-main">Local-first routing</p>
                    <p className="mt-1">{routing?.hardware_summary ?? "Model routing data is still loading."}</p>
                </div>
            </div>

            {loading ? (
                <div className="mt-5 rounded-2xl border border-cortex-border bg-cortex-surface px-4 py-4 text-sm text-cortex-text-muted">
                    Loading local model inventory...
                </div>
            ) : error ? (
                <div className="mt-5 rounded-2xl border border-cortex-danger/30 bg-cortex-surface px-4 py-4 text-sm text-cortex-text-muted">
                    <p className="font-medium text-cortex-text-main">Unable to load output model routing</p>
                    <p className="mt-2 leading-6">{error}</p>
                </div>
            ) : (
                <>
                    <div className="mt-5 grid gap-3 lg:grid-cols-2">
                        <button
                            type="button"
                            onClick={() => onRoutingModeChange("single_model")}
                            disabled={updatePending}
                            className={`rounded-2xl border px-4 py-4 text-left transition-colors ${
                                selectedRoutingMode === "single_model"
                                    ? "border-cortex-primary/40 bg-cortex-primary/10"
                                    : "border-cortex-border bg-cortex-surface hover:border-cortex-primary/20"
                            }`}
                        >
                            <p className="text-sm font-semibold text-cortex-text-main">Use one model for everyone</p>
                            <p className="mt-2 text-sm leading-6 text-cortex-text-muted">
                                Keeps every Team Lead and Specialist on the same concrete model until you intentionally split by detected output type.
                            </p>
                        </button>
                        <button
                            type="button"
                            onClick={() => onRoutingModeChange("detected_output_types")}
                            disabled={updatePending}
                            className={`rounded-2xl border px-4 py-4 text-left transition-colors ${
                                selectedRoutingMode === "detected_output_types"
                                    ? "border-cortex-primary/40 bg-cortex-primary/10"
                                    : "border-cortex-border bg-cortex-surface hover:border-cortex-primary/20"
                            }`}
                        >
                            <p className="text-sm font-semibold text-cortex-text-main">Use detected models by output type</p>
                            <p className="mt-2 text-sm leading-6 text-cortex-text-muted">
                                Lets Soma keep the team simple while routing research, code, and vision-heavy work to more appropriate local models.
                            </p>
                        </button>
                    </div>

                    <div className="mt-5 rounded-2xl border border-cortex-border bg-cortex-surface px-4 py-4">
                        <label className="text-sm font-semibold text-cortex-text-main" htmlFor="organization-default-output-model">
                            Organization default model
                        </label>
                        <p className="mt-2 text-sm leading-6 text-cortex-text-muted">
                            This is the fallback model all roles use by default. In detected mode, output types only diverge where you explicitly bind a different model.
                        </p>
                        <select
                            id="organization-default-output-model"
                            value={selectedDefaultModelId}
                            onChange={(event) => onDefaultModelChange(event.target.value)}
                            disabled={updatePending}
                            className="mt-3 w-full rounded-2xl border border-cortex-border bg-cortex-bg px-4 py-3 text-sm text-cortex-text-main"
                        >
                            {modelOptions.map((option) => (
                                <option key={option.model_id} value={option.model_id}>
                                    {option.label} ({option.model_id})
                                </option>
                            ))}
                        </select>
                    </div>

                    {selectedRoutingMode === "detected_output_types" && (
                        <div className="mt-5 grid gap-3">
                            {OUTPUT_TYPE_OPTIONS.map((option) => (
                                <div key={option.id} className="rounded-2xl border border-cortex-border bg-cortex-surface px-4 py-4">
                                    <p className="text-sm font-semibold text-cortex-text-main">{option.label}</p>
                                    <p className="mt-2 text-sm leading-6 text-cortex-text-muted">{option.description}</p>
                                    <select
                                        value={selectedBindings[option.id] ?? selectedDefaultModelId}
                                        onChange={(event) => onBindingChange(option.id, event.target.value)}
                                        disabled={updatePending}
                                        className="mt-3 w-full rounded-2xl border border-cortex-border bg-cortex-bg px-4 py-3 text-sm text-cortex-text-main"
                                    >
                                        {modelOptions.map((entry) => (
                                            <option key={`${option.id}-${entry.model_id}`} value={entry.model_id}>
                                                {entry.label} ({entry.model_id})
                                            </option>
                                        ))}
                                    </select>
                                </div>
                            ))}
                        </div>
                    )}

                    {(routing?.recommended_models?.length ?? 0) > 0 && (
                        <div className="mt-5">
                            <p className="text-sm font-semibold text-cortex-text-main">Popular self-hosted starting points</p>
                            <div className="mt-3 grid gap-3 lg:grid-cols-2">
                                {(routing?.recommended_models ?? []).map((entry) => (
                                    <div key={entry.model_id} className="rounded-2xl border border-cortex-border bg-cortex-surface px-4 py-4">
                                        <div className="flex flex-wrap items-center gap-2">
                                            <p className="text-sm font-semibold text-cortex-text-main">{entry.label}</p>
                                            {entry.installed && (
                                                <span className="inline-flex rounded-full border border-cortex-primary/30 bg-cortex-primary/10 px-2 py-0.5 text-[10px] font-semibold uppercase tracking-[0.16em] text-cortex-primary">
                                                    Installed
                                                </span>
                                            )}
                                        </div>
                                        <p className="mt-2 text-sm leading-6 text-cortex-text-muted">{entry.summary}</p>
                                        {entry.hosting_fit && <p className="mt-2 text-sm leading-6 text-cortex-text-muted">{entry.hosting_fit}</p>}
                                    </div>
                                ))}
                            </div>
                        </div>
                    )}

                    <div className="mt-5 rounded-2xl border border-cortex-border bg-cortex-surface px-4 py-4">
                        <p className="text-sm font-semibold text-cortex-text-main">Soma model review guardrail</p>
                        <p className="mt-2 text-sm leading-6 text-cortex-text-muted">
                            {routing?.review_permission_prompt ??
                                "Ask the owner/admin before Soma reviews potential model behavior for a requested output or changes saved routing."}
                        </p>
                        {(routing?.automatic_selection_criteria?.length ?? 0) > 0 && (
                            <div className="mt-3 grid gap-2">
                                {(routing?.automatic_selection_criteria ?? []).map((criterion) => (
                                    <p key={criterion} className="rounded-xl border border-cortex-border bg-cortex-bg px-3 py-2 text-sm leading-6 text-cortex-text-muted">
                                        {criterion}
                                    </p>
                                ))}
                            </div>
                        )}
                    </div>

                    {(routing?.review_candidates?.length ?? 0) > 0 && (
                        <div className="mt-5">
                            <p className="text-sm font-semibold text-cortex-text-main">Behavior review candidates</p>
                            <p className="mt-2 text-sm leading-6 text-cortex-text-muted">
                                These are Soma&apos;s first-pass local candidates when the admin has not pinned a model for a specific output type yet.
                            </p>
                            <div className="mt-3 grid gap-3 lg:grid-cols-2">
                                {(routing?.review_candidates ?? []).map((candidate) => (
                                    <div key={`${candidate.output_type_id}-${candidate.model_id}`} className="rounded-2xl border border-cortex-border bg-cortex-surface px-4 py-4">
                                        <div className="flex flex-wrap items-center gap-2">
                                            <p className="text-sm font-semibold text-cortex-text-main">
                                                {candidate.output_type_label}: {candidate.model_summary}
                                            </p>
                                            {candidate.installed && (
                                                <span className="inline-flex rounded-full border border-cortex-primary/30 bg-cortex-primary/10 px-2 py-0.5 text-[10px] font-semibold uppercase tracking-[0.16em] text-cortex-primary">
                                                    Installed
                                                </span>
                                            )}
                                        </div>
                                        <p className="mt-2 text-xs font-mono uppercase tracking-[0.16em] text-cortex-text-muted">{candidate.model_id}</p>
                                        <div className="mt-3 grid gap-2">
                                            {candidate.review_criteria.map((criterion) => (
                                                <p key={criterion} className="rounded-xl border border-cortex-border bg-cortex-bg px-3 py-2 text-sm leading-6 text-cortex-text-muted">
                                                    {criterion}
                                                </p>
                                            ))}
                                        </div>
                                    </div>
                                ))}
                            </div>
                        </div>
                    )}

                    {updateError && (
                        <div className="mt-5 rounded-2xl border border-cortex-danger/30 bg-cortex-surface px-4 py-4 text-sm text-cortex-text-muted">
                            <p className="font-medium text-cortex-text-main">Unable to update output model routing</p>
                            <p className="mt-2 leading-6">{updateError}</p>
                        </div>
                    )}

                    <div className="mt-5 flex flex-wrap gap-3">
                        <button
                            type="button"
                            onClick={onSubmit}
                            disabled={!selectedDefaultModelId || updatePending}
                            className="inline-flex items-center gap-2 rounded-xl bg-cortex-primary px-4 py-2.5 text-sm font-semibold text-cortex-bg transition hover:brightness-110 disabled:cursor-not-allowed disabled:opacity-60"
                        >
                            {updatePending ? <Loader2 className="h-4 w-4 animate-spin" /> : <Sparkles className="h-4 w-4" />}
                            {updatePending ? "Updating output model routing..." : "Use output model routing"}
                        </button>
                    </div>
                </>
            )}
        </div>
    );
}
