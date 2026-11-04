"use client";

import { Loader2, RefreshCcw, Sparkles } from "lucide-react";
import type { ResponseContractProfileId } from "@/lib/organizations";
import { RESPONSE_CONTRACT_OPTIONS } from "@/components/organizations/organizationProfileOptions";

export function AgentTypeResponseContractSelectionPanel({
    selectedAgentTypeResponseContractProfile,
    agentTypeResponseContractUpdatePending,
    agentTypeResponseContractUpdateError,
    onAgentTypeResponseContractProfileSelect,
    onClose,
    onSubmit,
}: {
    selectedAgentTypeResponseContractProfile: ResponseContractProfileId | null;
    agentTypeResponseContractUpdatePending: boolean;
    agentTypeResponseContractUpdateError: string | null;
    onAgentTypeResponseContractProfileSelect: (profile: ResponseContractProfileId) => void;
    onClose: () => void;
    onSubmit: () => void;
}) {
    return (
        <div className="mt-4 rounded-3xl border border-cortex-primary/20 bg-cortex-surface p-5">
            <div className="flex flex-col gap-3 lg:flex-row lg:items-start lg:justify-between">
                <div>
                    <h4 className="text-lg font-semibold text-cortex-text-main">Choose a Response Style for this Agent Type</h4>
                    <p className="mt-2 max-w-3xl text-sm leading-7 text-cortex-text-muted">
                        Pick one safe Response Style only for this Agent Type when this role needs a consistent specialist voice that should stay steady even if the wider default changes.
                    </p>
                </div>
                <div className="rounded-2xl border border-cortex-border bg-cortex-bg px-4 py-3 text-sm text-cortex-text-muted">
                    <p className="font-medium text-cortex-text-main">Agent Type level only</p>
                    <p className="mt-1">The Soma workspace stays visible, and Organization / Team default behavior remains in place for every other role type.</p>
                </div>
            </div>

            <div className="mt-5 grid gap-3">
                {RESPONSE_CONTRACT_OPTIONS.map((option) => {
                    const isSelected = selectedAgentTypeResponseContractProfile === option.id;
                    return (
                        <button
                            key={option.id}
                            type="button"
                            onClick={() => onAgentTypeResponseContractProfileSelect(option.id)}
                            disabled={agentTypeResponseContractUpdatePending}
                            className={`rounded-2xl border px-4 py-4 text-left transition-colors ${
                                isSelected
                                    ? "border-cortex-primary/40 bg-cortex-primary/10"
                                    : "border-cortex-border bg-cortex-bg hover:border-cortex-primary/20"
                            } ${agentTypeResponseContractUpdatePending ? "opacity-75" : ""}`}
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
                                        Selected
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

            {agentTypeResponseContractUpdateError && (
                <div className="mt-5 rounded-2xl border border-cortex-danger/30 bg-cortex-bg px-4 py-4 text-sm text-cortex-text-muted">
                    <p className="font-medium text-cortex-text-main">Unable to update this Agent Type Response Style</p>
                    <p className="mt-2 leading-6">{agentTypeResponseContractUpdateError}</p>
                    <p className="mt-2 leading-6">Try again to keep this specialist role aligned with the Soma workspace while the wider organization stays visible.</p>
                </div>
            )}

            <div className="mt-5 flex flex-wrap gap-3">
                <button
                    type="button"
                    onClick={onSubmit}
                    disabled={!selectedAgentTypeResponseContractProfile || agentTypeResponseContractUpdatePending}
                    className="inline-flex items-center gap-2 rounded-xl bg-cortex-primary px-4 py-2.5 text-sm font-semibold text-cortex-bg transition hover:brightness-110 disabled:cursor-not-allowed disabled:opacity-60"
                >
                    {agentTypeResponseContractUpdatePending ? <Loader2 className="h-4 w-4 animate-spin" /> : <Sparkles className="h-4 w-4" />}
                    {agentTypeResponseContractUpdatePending ? "Updating Agent Type Response Style..." : "Use selected Response Style"}
                </button>
                {agentTypeResponseContractUpdateError && (
                    <button
                        type="button"
                        onClick={onSubmit}
                        disabled={!selectedAgentTypeResponseContractProfile || agentTypeResponseContractUpdatePending}
                        className="inline-flex items-center gap-2 rounded-xl border border-cortex-border bg-cortex-bg px-4 py-2.5 text-sm font-medium text-cortex-text-main transition-colors hover:border-cortex-primary/20 disabled:cursor-not-allowed disabled:opacity-60"
                    >
                        <RefreshCcw className="h-4 w-4" />
                        Retry Response Style change
                    </button>
                )}
                <button
                    type="button"
                    onClick={onClose}
                    disabled={agentTypeResponseContractUpdatePending}
                    className="inline-flex items-center gap-2 rounded-xl border border-cortex-border bg-cortex-bg px-4 py-2.5 text-sm font-medium text-cortex-text-main transition-colors hover:border-cortex-primary/20 disabled:cursor-not-allowed disabled:opacity-60"
                >
                    Cancel
                </button>
            </div>
        </div>
    );
}
