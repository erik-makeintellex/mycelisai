"use client";

import { Loader2, RefreshCcw, Sparkles } from "lucide-react";
import type { OrganizationHomePayload } from "@/lib/organizations";
import {
    agentTypeAIEngineSourceLabel,
    agentTypeOutputModelSourceLabel,
    agentTypeResponseStyleSourceLabel,
    agentTypeSelectionKey,
    departmentDetailItems,
} from "@/components/organizations/organizationDetailHelpers";
import {
    AgentTypeAIEngineSelectionPanel,
    DepartmentAIEngineSelectionPanel,
} from "@/components/organizations/DepartmentOverrideSelectionPanels";
import { AgentTypeResponseContractSelectionPanel } from "@/components/organizations/AgentTypeResponseContractSelectionPanel";
import type { DepartmentAndAgentTypeOverrideControls } from "@/components/organizations/useDepartmentAndAgentTypeOverrides";

export function OrganizationDepartmentDetailSection({
    organization,
    teamLeadName,
    departmentOverrides,
}: {
    organization: OrganizationHomePayload;
    teamLeadName: string;
    departmentOverrides: DepartmentAndAgentTypeOverrideControls;
}) {
    const departmentItems = departmentDetailItems(organization);
    const {
        activeDepartmentAIEngineId,
        selectedDepartmentAIEngineProfile,
        departmentAIEngineUpdatePendingId,
        departmentAIEngineUpdateError,
        onDepartmentAIEngineProfileSelect,
        onOpenDepartmentAIEngineSelector,
        onCloseDepartmentAIEngineSelector,
        onSubmitDepartmentAIEngineSelection,
        onRevertDepartmentAIEngineSelection,
        activeAgentTypeAIEngineKey,
        selectedAgentTypeAIEngineProfile,
        agentTypeAIEngineUpdatePendingKey,
        agentTypeAIEngineUpdateError,
        onAgentTypeAIEngineProfileSelect,
        onOpenAgentTypeAIEngineSelector,
        onCloseAgentTypeAIEngineSelector,
        onSubmitAgentTypeAIEngineSelection,
        onRevertAgentTypeAIEngineSelection,
        activeAgentTypeResponseContractKey,
        selectedAgentTypeResponseContractProfile,
        agentTypeResponseContractUpdatePendingKey,
        agentTypeResponseContractUpdateError,
        onAgentTypeResponseContractProfileSelect,
        onOpenAgentTypeResponseContractSelector,
        onCloseAgentTypeResponseContractSelector,
        onSubmitAgentTypeResponseContractSelection,
        onRevertAgentTypeResponseContractSelection,
    } = departmentOverrides;

    if (departmentItems.length === 0) {
        return (
            <div className="mt-5 rounded-2xl border border-cortex-border bg-cortex-bg px-4 py-4 text-sm leading-6 text-cortex-text-muted">
                {`${teamLeadName} can still shape the first operating lane before Departments are added.`}
            </div>
        );
    }

    return (
        <div className="mt-5 grid gap-4">
            {departmentItems.map((item) => {
                const isEditing = activeDepartmentAIEngineId === item.id;
                const isPending = departmentAIEngineUpdatePendingId === item.id;
                const errorMessage = departmentAIEngineUpdateError?.departmentId === item.id ? departmentAIEngineUpdateError.message : null;

                return (
                    <div key={item.id} className="rounded-2xl border border-cortex-border bg-cortex-bg px-4 py-4">
                        <div className="flex flex-col gap-3 lg:flex-row lg:items-start lg:justify-between">
                            <div>
                                <p className="text-sm font-semibold text-cortex-text-main">{item.name}</p>
                                <p className="mt-2 text-sm leading-6 text-cortex-text-muted">{item.purpose}</p>
                                <p className="mt-2 text-sm leading-6 text-cortex-text-muted">
                                    {item.specialist_count > 0
                                        ? `${item.specialist_count} ${item.specialist_count === 1 ? "Specialist" : "Specialists"} visible here.`
                                        : "No Specialists assigned yet."}
                                </p>
                            </div>
                            <div className="rounded-2xl border border-cortex-border bg-cortex-surface px-4 py-3 text-sm text-cortex-text-muted lg:max-w-sm">
                                <p className="font-medium text-cortex-text-main">AI Engine status</p>
                                <p className="mt-1 leading-6">{item.aiEngineStateLabel}</p>
                                <p className="mt-2 leading-6">
                                    {item.inherits_organization_ai_engine
                                        ? "This Department follows the organization-wide AI Engine unless you set a Department-specific choice."
                                        : "This Department is using its own guided AI Engine choice and will keep it until you return it to the organization default."}
                                </p>
                            </div>
                        </div>

                        <div className="mt-4 flex flex-wrap gap-3">
                            <button
                                type="button"
                                onClick={() => onOpenDepartmentAIEngineSelector(item)}
                                disabled={Boolean(departmentAIEngineUpdatePendingId && departmentAIEngineUpdatePendingId !== item.id)}
                                className="inline-flex items-center gap-2 rounded-xl border border-cortex-primary/30 bg-cortex-primary/10 px-3 py-2 text-sm font-medium text-cortex-text-main transition-colors hover:border-cortex-primary/40 disabled:cursor-not-allowed disabled:opacity-60"
                            >
                                <Sparkles className="h-4 w-4 text-cortex-primary" />
                                Change for this Team
                            </button>
                            {!item.inherits_organization_ai_engine && (
                                <button
                                    type="button"
                                    onClick={() => onRevertDepartmentAIEngineSelection(item.id)}
                                    disabled={isPending}
                                    className="inline-flex items-center gap-2 rounded-xl border border-cortex-border bg-cortex-surface px-3 py-2 text-sm font-medium text-cortex-text-main transition-colors hover:border-cortex-primary/20 disabled:cursor-not-allowed disabled:opacity-60"
                                >
                                    {isPending ? <Loader2 className="h-4 w-4 animate-spin" /> : <RefreshCcw className="h-4 w-4" />}
                                    Revert to Organization Default
                                </button>
                            )}
                        </div>

                        {isEditing && (
                            <DepartmentAIEngineSelectionPanel
                                selectedDepartmentAIEngineProfile={selectedDepartmentAIEngineProfile}
                                departmentAIEngineUpdatePending={isPending}
                                departmentAIEngineUpdateError={errorMessage}
                                onDepartmentAIEngineProfileSelect={onDepartmentAIEngineProfileSelect}
                                onClose={onCloseDepartmentAIEngineSelector}
                                onSubmit={() => onSubmitDepartmentAIEngineSelection(item.id)}
                            />
                        )}
                        <div className="mt-5 rounded-2xl border border-cortex-border bg-cortex-surface px-4 py-4">
                            <div className="flex flex-col gap-2 lg:flex-row lg:items-start lg:justify-between">
                                <div>
                                    <p className="text-sm font-semibold text-cortex-text-main">Agent Type Profiles</p>
                                    <p className="mt-2 text-sm leading-6 text-cortex-text-muted">
                                        These guided profiles show how this Team&apos;s role types inherit the Team default or follow a type-specific AI Engine or Response Style.
                                    </p>
                                </div>
                                <div className="rounded-2xl border border-cortex-border bg-cortex-bg px-4 py-3 text-sm text-cortex-text-muted lg:max-w-sm">
                                    <p className="font-medium text-cortex-text-main">Guided role tuning</p>
                                    <p className="mt-1 leading-6">Use these profiles to review or safely tune role-type defaults while individual Specialist instances stay out of scope.</p>
                                </div>
                            </div>

                            {item.agentTypeProfiles.length > 0 ? (
                                <div className="mt-4 grid gap-3">
                                    {item.agentTypeProfiles.map((profile) => {
                                        const agentTypeKey = agentTypeSelectionKey(item.id, profile.id);
                                        const isAgentTypeEditing = activeAgentTypeAIEngineKey === agentTypeKey;
                                        const isAgentTypePending = agentTypeAIEngineUpdatePendingKey === agentTypeKey;
                                        const isAgentTypeResponseEditing = activeAgentTypeResponseContractKey === agentTypeKey;
                                        const isAgentTypeResponsePending = agentTypeResponseContractUpdatePendingKey === agentTypeKey;
                                        const agentTypeError =
                                            agentTypeAIEngineUpdateError?.departmentId === item.id && agentTypeAIEngineUpdateError.agentTypeId === profile.id
                                                ? agentTypeAIEngineUpdateError.message
                                                : null;
                                        const agentTypeResponseError =
                                            agentTypeResponseContractUpdateError?.departmentId === item.id &&
                                            agentTypeResponseContractUpdateError.agentTypeId === profile.id
                                                ? agentTypeResponseContractUpdateError.message
                                                : null;

                                        return (
                                            <div key={profile.id} className="rounded-2xl border border-cortex-border bg-cortex-bg px-4 py-4">
                                                <div className="flex flex-col gap-3 lg:flex-row lg:items-start lg:justify-between">
                                                    <div>
                                                        <p className="text-sm font-semibold text-cortex-text-main">{profile.name}</p>
                                                        <p className="mt-2 text-sm leading-6 text-cortex-text-muted">{profile.helps_with}</p>
                                                    </div>
                                                    <div className="rounded-2xl border border-cortex-border bg-cortex-surface px-4 py-3 text-sm text-cortex-text-muted lg:max-w-sm">
                                                        <p className="font-medium text-cortex-text-main">Inheritance clarity</p>
                                                        <p className="mt-1 leading-6">
                                                            {profile.inherits_department_ai_engine
                                                                ? "This agent type follows the Team AI Engine unless a type-specific binding is shown here."
                                                                : "This agent type keeps its own AI Engine binding instead of following the Team default."}
                                                        </p>
                                                    </div>
                                                </div>
                                                <div className="mt-4 grid gap-3 lg:grid-cols-3">
                                                    <div className="rounded-2xl border border-cortex-border bg-cortex-surface px-4 py-3">
                                                        <p className="text-[10px] font-mono uppercase tracking-[0.18em] text-cortex-text-muted">AI Engine</p>
                                                        <p className="mt-2 text-sm font-medium text-cortex-text-main">{agentTypeAIEngineSourceLabel(profile)}</p>
                                                    </div>
                                                    <div className="rounded-2xl border border-cortex-border bg-cortex-surface px-4 py-3">
                                                        <p className="text-[10px] font-mono uppercase tracking-[0.18em] text-cortex-text-muted">Response Style</p>
                                                        <p className="mt-2 text-sm font-medium text-cortex-text-main">{agentTypeResponseStyleSourceLabel(profile)}</p>
                                                    </div>
                                                    <div className="rounded-2xl border border-cortex-border bg-cortex-surface px-4 py-3">
                                                        <p className="text-[10px] font-mono uppercase tracking-[0.18em] text-cortex-text-muted">Detected output model</p>
                                                        <p className="mt-2 text-sm font-medium text-cortex-text-main">{agentTypeOutputModelSourceLabel(profile)}</p>
                                                        {profile.output_type_label && (
                                                            <p className="mt-2 text-sm leading-6 text-cortex-text-muted">
                                                                Output type: <span className="font-medium text-cortex-text-main">{profile.output_type_label}</span>
                                                            </p>
                                                        )}
                                                    </div>
                                                </div>
                                                <div className="mt-4 flex flex-wrap gap-3">
                                                    <button
                                                        type="button"
                                                        onClick={() => onOpenAgentTypeAIEngineSelector(item.id, profile)}
                                                        disabled={
                                                            Boolean(agentTypeAIEngineUpdatePendingKey && agentTypeAIEngineUpdatePendingKey !== agentTypeKey) ||
                                                            Boolean(agentTypeResponseContractUpdatePendingKey && agentTypeResponseContractUpdatePendingKey !== agentTypeKey)
                                                        }
                                                        className="inline-flex items-center gap-2 rounded-xl border border-cortex-primary/30 bg-cortex-primary/10 px-3 py-2 text-sm font-medium text-cortex-text-main transition-colors hover:border-cortex-primary/40 disabled:cursor-not-allowed disabled:opacity-60"
                                                    >
                                                        <Sparkles className="h-4 w-4 text-cortex-primary" />
                                                        Change for this Agent Type
                                                    </button>
                                                    {!profile.inherits_department_ai_engine && (
                                                        <button
                                                            type="button"
                                                            onClick={() => onRevertAgentTypeAIEngineSelection(item.id, profile.id)}
                                                            disabled={isAgentTypePending}
                                                            className="inline-flex items-center gap-2 rounded-xl border border-cortex-border bg-cortex-surface px-3 py-2 text-sm font-medium text-cortex-text-main transition-colors hover:border-cortex-primary/20 disabled:cursor-not-allowed disabled:opacity-60"
                                                        >
                                                            {isAgentTypePending ? <Loader2 className="h-4 w-4 animate-spin" /> : <RefreshCcw className="h-4 w-4" />}
                                                            Use Team Default
                                                        </button>
                                                    )}
                                                    <button
                                                        type="button"
                                                        onClick={() => onOpenAgentTypeResponseContractSelector(item.id, profile)}
                                                        disabled={
                                                            Boolean(agentTypeAIEngineUpdatePendingKey && agentTypeAIEngineUpdatePendingKey !== agentTypeKey) ||
                                                            Boolean(agentTypeResponseContractUpdatePendingKey && agentTypeResponseContractUpdatePendingKey !== agentTypeKey)
                                                        }
                                                        className="inline-flex items-center gap-2 rounded-xl border border-cortex-primary/30 bg-cortex-primary/10 px-3 py-2 text-sm font-medium text-cortex-text-main transition-colors hover:border-cortex-primary/40 disabled:cursor-not-allowed disabled:opacity-60"
                                                    >
                                                        <Sparkles className="h-4 w-4 text-cortex-primary" />
                                                        Change Response Style for this Agent Type
                                                    </button>
                                                    {!profile.inherits_default_response_contract && (
                                                        <button
                                                            type="button"
                                                            onClick={() => onRevertAgentTypeResponseContractSelection(item.id, profile.id)}
                                                            disabled={isAgentTypeResponsePending}
                                                            className="inline-flex items-center gap-2 rounded-xl border border-cortex-border bg-cortex-surface px-3 py-2 text-sm font-medium text-cortex-text-main transition-colors hover:border-cortex-primary/20 disabled:cursor-not-allowed disabled:opacity-60"
                                                        >
                                                            {isAgentTypeResponsePending ? <Loader2 className="h-4 w-4 animate-spin" /> : <RefreshCcw className="h-4 w-4" />}
                                                            Use Organization / Team Default
                                                        </button>
                                                    )}
                                                </div>

                                                {isAgentTypeEditing && (
                                                    <AgentTypeAIEngineSelectionPanel
                                                        selectedAgentTypeAIEngineProfile={selectedAgentTypeAIEngineProfile}
                                                        agentTypeAIEngineUpdatePending={isAgentTypePending}
                                                        agentTypeAIEngineUpdateError={agentTypeError}
                                                        onAgentTypeAIEngineProfileSelect={onAgentTypeAIEngineProfileSelect}
                                                        onClose={onCloseAgentTypeAIEngineSelector}
                                                        onSubmit={() => onSubmitAgentTypeAIEngineSelection(item.id, profile.id)}
                                                    />
                                                )}

                                                {isAgentTypeResponseEditing && (
                                                    <AgentTypeResponseContractSelectionPanel
                                                        selectedAgentTypeResponseContractProfile={selectedAgentTypeResponseContractProfile}
                                                        agentTypeResponseContractUpdatePending={isAgentTypeResponsePending}
                                                        agentTypeResponseContractUpdateError={agentTypeResponseError}
                                                        onAgentTypeResponseContractProfileSelect={onAgentTypeResponseContractProfileSelect}
                                                        onClose={onCloseAgentTypeResponseContractSelector}
                                                        onSubmit={() => onSubmitAgentTypeResponseContractSelection(item.id, profile.id)}
                                                    />
                                                )}
                                            </div>
                                        );
                                    })}
                                </div>
                            ) : (
                                <div className="mt-4 rounded-2xl border border-cortex-border bg-cortex-bg px-4 py-4 text-sm leading-6 text-cortex-text-muted">
                                    Agent Type Profiles will appear here as this Team starts assigning specialist role types.
                                </div>
                            )}
                        </div>
                    </div>
                );
            })}
        </div>
    );
}
