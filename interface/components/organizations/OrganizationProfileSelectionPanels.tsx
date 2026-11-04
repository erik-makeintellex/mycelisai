"use client";

import { Loader2, RefreshCcw, Sparkles } from "lucide-react";
import type {
    OrganizationAIEngineProfileId,
    ResponseContractProfileId,
} from "@/lib/organizations";
import {
    AI_ENGINE_OPTIONS,
    RESPONSE_CONTRACT_OPTIONS,
} from "@/components/organizations/organizationProfileOptions";

export function AIEngineSelectionPanel({
    selectedAIEngineProfile,
    aiEngineUpdatePending,
    aiEngineUpdateError,
    onAIEngineProfileSelect,
    onClose,
    onSubmit,
}: {
    selectedAIEngineProfile: OrganizationAIEngineProfileId | null;
    aiEngineUpdatePending: boolean;
    aiEngineUpdateError: string | null;
    onAIEngineProfileSelect: (profile: OrganizationAIEngineProfileId) => void;
    onClose: () => void;
    onSubmit: () => void;
}) {
    return (
        <div className="mt-5 rounded-3xl border border-cortex-primary/20 bg-cortex-bg p-5">
            <div className="flex flex-col gap-3 lg:flex-row lg:items-start lg:justify-between">
                <div>
                    <h4 className="text-lg font-semibold text-cortex-text-main">Choose an AI Engine profile</h4>
                    <p className="mt-2 max-w-3xl text-sm leading-7 text-cortex-text-muted">
                        Pick one guided AI Engine profile for the whole AI Organization. This tunes how Soma plans, responds, and carries work forward.
                    </p>
                </div>
                <div className="rounded-2xl border border-cortex-border bg-cortex-surface px-4 py-3 text-sm text-cortex-text-muted">
                    <p className="font-medium text-cortex-text-main">Organization level only</p>
                    <p className="mt-1">Department and role tuning stay in Department details so this organization-wide choice stays clear.</p>
                </div>
            </div>

            <div className="mt-5 grid gap-3">
                {AI_ENGINE_OPTIONS.map((option) => {
                    const isSelected = selectedAIEngineProfile === option.id;
                    return (
                        <button
                            key={option.id}
                            type="button"
                            onClick={() => onAIEngineProfileSelect(option.id)}
                            disabled={aiEngineUpdatePending}
                            className={`rounded-2xl border px-4 py-4 text-left transition-colors ${
                                isSelected
                                    ? "border-cortex-primary/40 bg-cortex-primary/10"
                                    : "border-cortex-border bg-cortex-surface hover:border-cortex-primary/20"
                            } ${aiEngineUpdatePending ? "opacity-75" : ""}`}
                        >
                            <div className="flex flex-col gap-2 lg:flex-row lg:items-start lg:justify-between">
                                <div>
                                    <p className="text-sm font-semibold text-cortex-text-main">{option.label}</p>
                                    <p className="mt-2 text-sm leading-6 text-cortex-text-muted">{option.description}</p>
                                </div>
                                {isSelected && (
                                    <span className="inline-flex w-fit rounded-full border border-cortex-primary/30 bg-cortex-primary/10 px-3 py-1 text-xs font-medium uppercase tracking-[0.16em] text-cortex-primary">
                                        Current choice
                                    </span>
                                )}
                            </div>
                            <p className="mt-3 text-sm leading-6 text-cortex-text-muted">
                                <span className="font-medium text-cortex-text-main">Good for:</span> {option.goodFor}
                            </p>
                        </button>
                    );
                })}
            </div>

            {aiEngineUpdateError && (
                <div className="mt-5 rounded-2xl border border-cortex-danger/30 bg-cortex-surface px-4 py-4 text-sm text-cortex-text-muted">
                    <p className="font-medium text-cortex-text-main">Unable to update AI Engine Settings</p>
                    <p className="mt-2 leading-6">{aiEngineUpdateError}</p>
                    <p className="mt-2 leading-6">Try again to apply the selected AI Engine profile while keeping this Soma workspace in place.</p>
                </div>
            )}

            <div className="mt-5 flex flex-wrap gap-3">
                <button
                    type="button"
                    onClick={onSubmit}
                    disabled={!selectedAIEngineProfile || aiEngineUpdatePending}
                    className="inline-flex items-center gap-2 rounded-xl bg-cortex-primary px-4 py-2.5 text-sm font-semibold text-cortex-bg transition hover:brightness-110 disabled:cursor-not-allowed disabled:opacity-60"
                >
                    {aiEngineUpdatePending ? <Loader2 className="h-4 w-4 animate-spin" /> : <Sparkles className="h-4 w-4" />}
                    {aiEngineUpdatePending ? "Updating AI Engine..." : "Use selected AI Engine"}
                </button>
                {aiEngineUpdateError && (
                    <button
                        type="button"
                        onClick={onSubmit}
                        disabled={!selectedAIEngineProfile || aiEngineUpdatePending}
                        className="inline-flex items-center gap-2 rounded-xl border border-cortex-border bg-cortex-surface px-4 py-2.5 text-sm font-medium text-cortex-text-main transition-colors hover:border-cortex-primary/20 disabled:cursor-not-allowed disabled:opacity-60"
                    >
                        <RefreshCcw className="h-4 w-4" />
                        Retry AI Engine change
                    </button>
                )}
                <button
                    type="button"
                    onClick={onClose}
                    disabled={aiEngineUpdatePending}
                    className="inline-flex items-center gap-2 rounded-xl border border-cortex-border bg-cortex-surface px-4 py-2.5 text-sm font-medium text-cortex-text-main transition-colors hover:border-cortex-primary/20 disabled:cursor-not-allowed disabled:opacity-60"
                >
                    Cancel
                </button>
            </div>
        </div>
    );
}

export function ResponseContractSelectionPanel({
    selectedResponseContractProfile,
    responseContractUpdatePending,
    responseContractUpdateError,
    onResponseContractProfileSelect,
    onClose,
    onSubmit,
}: {
    selectedResponseContractProfile: ResponseContractProfileId | null;
    responseContractUpdatePending: boolean;
    responseContractUpdateError: string | null;
    onResponseContractProfileSelect: (profile: ResponseContractProfileId) => void;
    onClose: () => void;
    onSubmit: () => void;
}) {
    return (
        <div className="mt-5 rounded-3xl border border-cortex-primary/20 bg-cortex-bg p-5">
            <div className="flex flex-col gap-3 lg:flex-row lg:items-start lg:justify-between">
                <div>
                    <h4 className="text-lg font-semibold text-cortex-text-main">Choose a Response Style</h4>
                    <p className="mt-2 max-w-3xl text-sm leading-7 text-cortex-text-muted">
                        Pick one safe Response Style for the whole AI Organization. This shapes tone, structure, and detail without exposing raw prompt or policy text.
                    </p>
                </div>
                <div className="rounded-2xl border border-cortex-border bg-cortex-surface px-4 py-3 text-sm text-cortex-text-muted">
                    <p className="font-medium text-cortex-text-main">Organization level only</p>
                    <p className="mt-1">Department and role Response Style tuning stays in Department details so this organization-wide default stays easy to understand.</p>
                </div>
            </div>

            <div className="mt-5 grid gap-3">
                {RESPONSE_CONTRACT_OPTIONS.map((option) => {
                    const isSelected = selectedResponseContractProfile === option.id;
                    return (
                        <button
                            key={option.id}
                            type="button"
                            onClick={() => onResponseContractProfileSelect(option.id)}
                            disabled={responseContractUpdatePending}
                            className={`rounded-2xl border px-4 py-4 text-left transition-colors ${
                                isSelected
                                    ? "border-cortex-primary/40 bg-cortex-primary/10"
                                    : "border-cortex-border bg-cortex-surface hover:border-cortex-primary/20"
                            } ${responseContractUpdatePending ? "opacity-75" : ""}`}
                        >
                            <div className="flex flex-col gap-2 lg:flex-row lg:items-start lg:justify-between">
                                <div>
                                    <p className="text-sm font-semibold text-cortex-text-main">{option.label}</p>
                                    <p className="mt-2 text-sm leading-6 text-cortex-text-muted">
                                        <span className="font-medium text-cortex-text-main">Tone:</span> {option.toneStyle}
                                    </p>
                                </div>
                                {isSelected && (
                                    <span className="inline-flex w-fit rounded-full border border-cortex-primary/30 bg-cortex-primary/10 px-3 py-1 text-xs font-medium uppercase tracking-[0.16em] text-cortex-primary">
                                        Current choice
                                    </span>
                                )}
                            </div>
                            <p className="mt-3 text-sm leading-6 text-cortex-text-muted">
                                <span className="font-medium text-cortex-text-main">Structure:</span> {option.structure}
                            </p>
                            <p className="mt-2 text-sm leading-6 text-cortex-text-muted">
                                <span className="font-medium text-cortex-text-main">Detail:</span> {option.verbosity}
                            </p>
                            <p className="mt-2 text-sm leading-6 text-cortex-text-muted">
                                <span className="font-medium text-cortex-text-main">Best for:</span> {option.bestFor}
                            </p>
                        </button>
                    );
                })}
            </div>

            {responseContractUpdateError && (
                <div className="mt-5 rounded-2xl border border-cortex-danger/30 bg-cortex-surface px-4 py-4 text-sm text-cortex-text-muted">
                    <p className="font-medium text-cortex-text-main">Unable to update Response Style</p>
                    <p className="mt-2 leading-6">{responseContractUpdateError}</p>
                    <p className="mt-2 leading-6">Try again to keep the Soma workspace visible while the organization-wide response style is updated.</p>
                </div>
            )}

            <div className="mt-5 flex flex-wrap gap-3">
                <button
                    type="button"
                    onClick={onSubmit}
                    disabled={!selectedResponseContractProfile || responseContractUpdatePending}
                    className="inline-flex items-center gap-2 rounded-xl bg-cortex-primary px-4 py-2.5 text-sm font-semibold text-cortex-bg transition hover:brightness-110 disabled:cursor-not-allowed disabled:opacity-60"
                >
                    {responseContractUpdatePending ? <Loader2 className="h-4 w-4 animate-spin" /> : <Sparkles className="h-4 w-4" />}
                    {responseContractUpdatePending ? "Updating Response Style..." : "Use selected Response Style"}
                </button>
                {responseContractUpdateError && (
                    <button
                        type="button"
                        onClick={onSubmit}
                        disabled={!selectedResponseContractProfile || responseContractUpdatePending}
                        className="inline-flex items-center gap-2 rounded-xl border border-cortex-border bg-cortex-surface px-4 py-2.5 text-sm font-medium text-cortex-text-main transition-colors hover:border-cortex-primary/20 disabled:cursor-not-allowed disabled:opacity-60"
                    >
                        <RefreshCcw className="h-4 w-4" />
                        Retry Response Style change
                    </button>
                )}
                <button
                    type="button"
                    onClick={onClose}
                    disabled={responseContractUpdatePending}
                    className="inline-flex items-center gap-2 rounded-xl border border-cortex-border bg-cortex-surface px-4 py-2.5 text-sm font-medium text-cortex-text-main transition-colors hover:border-cortex-primary/20 disabled:cursor-not-allowed disabled:opacity-60"
                >
                    Cancel
                </button>
            </div>
        </div>
    );
}
