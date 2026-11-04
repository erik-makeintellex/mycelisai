"use client";

import { Loader2, Sparkles } from "lucide-react";
import type { OrganizationAIEngineProfileId } from "@/lib/organizations";
import { AI_ENGINE_OPTIONS } from "@/components/organizations/organizationProfileOptions";

export function DepartmentAIEngineSelectionPanel({
    selectedDepartmentAIEngineProfile,
    departmentAIEngineUpdatePending,
    departmentAIEngineUpdateError,
    onDepartmentAIEngineProfileSelect,
    onClose,
    onSubmit,
}: {
    selectedDepartmentAIEngineProfile: OrganizationAIEngineProfileId | null;
    departmentAIEngineUpdatePending: boolean;
    departmentAIEngineUpdateError: string | null;
    onDepartmentAIEngineProfileSelect: (profile: OrganizationAIEngineProfileId) => void;
    onClose: () => void;
    onSubmit: () => void;
}) {
    return (
        <div className="mt-4 rounded-3xl border border-cortex-primary/20 bg-cortex-surface p-5">
            <div className="flex flex-col gap-3 lg:flex-row lg:items-start lg:justify-between">
                <div>
                    <h4 className="text-lg font-semibold text-cortex-text-main">Choose an AI Engine for this Team</h4>
                    <p className="mt-2 max-w-3xl text-sm leading-7 text-cortex-text-muted">
                        Pick a guided AI Engine only for this Team when it needs a different planning or response style than the rest of the AI Organization.
                    </p>
                </div>
                <div className="rounded-2xl border border-cortex-border bg-cortex-bg px-4 py-3 text-sm text-cortex-text-muted">
                    <p className="font-medium text-cortex-text-main">Department-level only</p>
                    <p className="mt-1">Soma and the other Departments keep their current AI Engine behavior.</p>
                </div>
            </div>

            <div className="mt-5 grid gap-3">
                {AI_ENGINE_OPTIONS.map((option) => {
                    const isSelected = selectedDepartmentAIEngineProfile === option.id;
                    return (
                        <button
                            key={option.id}
                            type="button"
                            onClick={() => onDepartmentAIEngineProfileSelect(option.id)}
                            disabled={departmentAIEngineUpdatePending}
                            className={`rounded-2xl border px-4 py-4 text-left transition-colors ${
                                isSelected
                                    ? "border-cortex-primary/40 bg-cortex-primary/10"
                                    : "border-cortex-border bg-cortex-bg hover:border-cortex-primary/20"
                            } ${departmentAIEngineUpdatePending ? "opacity-75" : ""}`}
                        >
                            <div className="flex flex-col gap-2 lg:flex-row lg:items-start lg:justify-between">
                                <div>
                                    <p className="text-sm font-semibold text-cortex-text-main">{option.label}</p>
                                    <p className="mt-2 text-sm leading-6 text-cortex-text-muted">{option.description}</p>
                                </div>
                                {isSelected && (
                                    <span className="inline-flex w-fit rounded-full border border-cortex-primary/30 bg-cortex-primary/10 px-3 py-1 text-xs font-medium uppercase tracking-[0.16em] text-cortex-primary">
                                        Selected
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

            {departmentAIEngineUpdateError && (
                <div className="mt-5 rounded-2xl border border-cortex-danger/30 bg-cortex-bg px-4 py-4 text-sm text-cortex-text-muted">
                    <p className="font-medium text-cortex-text-main">Unable to update this Team AI Engine</p>
                    <p className="mt-2 leading-6">{departmentAIEngineUpdateError}</p>
                    <p className="mt-2 leading-6">Try again to keep this Department aligned with the Soma workspace while the rest of the organization stays visible.</p>
                </div>
            )}

            <div className="mt-5 flex flex-wrap gap-3">
                <button
                    type="button"
                    onClick={onSubmit}
                    disabled={!selectedDepartmentAIEngineProfile || departmentAIEngineUpdatePending}
                    className="inline-flex items-center gap-2 rounded-xl bg-cortex-primary px-4 py-2.5 text-sm font-semibold text-cortex-bg transition hover:brightness-110 disabled:cursor-not-allowed disabled:opacity-60"
                >
                    {departmentAIEngineUpdatePending ? <Loader2 className="h-4 w-4 animate-spin" /> : <Sparkles className="h-4 w-4" />}
                    {departmentAIEngineUpdatePending ? "Updating Team AI Engine..." : "Use selected AI Engine"}
                </button>
                <button
                    type="button"
                    onClick={onClose}
                    disabled={departmentAIEngineUpdatePending}
                    className="inline-flex items-center gap-2 rounded-xl border border-cortex-border bg-cortex-bg px-4 py-2.5 text-sm font-medium text-cortex-text-main transition-colors hover:border-cortex-primary/20 disabled:cursor-not-allowed disabled:opacity-60"
                >
                    Cancel
                </button>
            </div>
        </div>
    );
}

export function AgentTypeAIEngineSelectionPanel({
    selectedAgentTypeAIEngineProfile,
    agentTypeAIEngineUpdatePending,
    agentTypeAIEngineUpdateError,
    onAgentTypeAIEngineProfileSelect,
    onClose,
    onSubmit,
}: {
    selectedAgentTypeAIEngineProfile: OrganizationAIEngineProfileId | null;
    agentTypeAIEngineUpdatePending: boolean;
    agentTypeAIEngineUpdateError: string | null;
    onAgentTypeAIEngineProfileSelect: (profile: OrganizationAIEngineProfileId) => void;
    onClose: () => void;
    onSubmit: () => void;
}) {
    return (
        <div className="mt-4 rounded-3xl border border-cortex-primary/20 bg-cortex-surface p-5">
            <div className="flex flex-col gap-3 lg:flex-row lg:items-start lg:justify-between">
                <div>
                    <h4 className="text-lg font-semibold text-cortex-text-main">Choose an AI Engine for this Agent Type</h4>
                    <p className="mt-2 max-w-3xl text-sm leading-7 text-cortex-text-muted">
                        Pick a guided AI Engine only for this Agent Type when this role needs a consistent specialist behavior that should not drift with the Team default.
                    </p>
                </div>
                <div className="rounded-2xl border border-cortex-border bg-cortex-bg px-4 py-3 text-sm text-cortex-text-muted">
                    <p className="font-medium text-cortex-text-main">Agent Type level only</p>
                    <p className="mt-1">The Soma workspace stays visible, and Team-default inheritance remains in place for every other role type.</p>
                </div>
            </div>

            <div className="mt-5 grid gap-3">
                {AI_ENGINE_OPTIONS.map((option) => {
                    const isSelected = selectedAgentTypeAIEngineProfile === option.id;
                    return (
                        <button
                            key={option.id}
                            type="button"
                            onClick={() => onAgentTypeAIEngineProfileSelect(option.id)}
                            disabled={agentTypeAIEngineUpdatePending}
                            className={`rounded-2xl border px-4 py-4 text-left transition-colors ${
                                isSelected
                                    ? "border-cortex-primary/40 bg-cortex-primary/10"
                                    : "border-cortex-border bg-cortex-bg hover:border-cortex-primary/20"
                            } ${agentTypeAIEngineUpdatePending ? "opacity-75" : ""}`}
                        >
                            <div className="flex flex-col gap-2 lg:flex-row lg:items-start lg:justify-between">
                                <div>
                                    <p className="text-sm font-semibold text-cortex-text-main">{option.label}</p>
                                    <p className="mt-2 text-sm leading-6 text-cortex-text-muted">{option.description}</p>
                                </div>
                                {isSelected && (
                                    <span className="inline-flex w-fit rounded-full border border-cortex-primary/30 bg-cortex-primary/10 px-3 py-1 text-xs font-medium uppercase tracking-[0.16em] text-cortex-primary">
                                        Selected
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

            {agentTypeAIEngineUpdateError && (
                <div className="mt-5 rounded-2xl border border-cortex-danger/30 bg-cortex-bg px-4 py-4 text-sm text-cortex-text-muted">
                    <p className="font-medium text-cortex-text-main">Unable to update this Agent Type AI Engine</p>
                    <p className="mt-2 leading-6">{agentTypeAIEngineUpdateError}</p>
                    <p className="mt-2 leading-6">Try again to keep this specialist role aligned with the Department while the Soma workspace stays visible.</p>
                </div>
            )}

            <div className="mt-5 flex flex-wrap gap-3">
                <button
                    type="button"
                    onClick={onSubmit}
                    disabled={!selectedAgentTypeAIEngineProfile || agentTypeAIEngineUpdatePending}
                    className="inline-flex items-center gap-2 rounded-xl bg-cortex-primary px-4 py-2.5 text-sm font-semibold text-cortex-bg transition hover:brightness-110 disabled:cursor-not-allowed disabled:opacity-60"
                >
                    {agentTypeAIEngineUpdatePending ? <Loader2 className="h-4 w-4 animate-spin" /> : <Sparkles className="h-4 w-4" />}
                    {agentTypeAIEngineUpdatePending ? "Updating Agent Type AI Engine..." : "Use selected AI Engine"}
                </button>
                <button
                    type="button"
                    onClick={onClose}
                    disabled={agentTypeAIEngineUpdatePending}
                    className="inline-flex items-center gap-2 rounded-xl border border-cortex-border bg-cortex-bg px-4 py-2.5 text-sm font-medium text-cortex-text-main transition-colors hover:border-cortex-primary/20 disabled:cursor-not-allowed disabled:opacity-60"
                >
                    Cancel
                </button>
            </div>
        </div>
    );
}
