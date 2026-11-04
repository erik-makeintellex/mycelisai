"use client";

import { ArrowLeft, Sparkles } from "lucide-react";
import type {
    OrganizationAutomationItem,
    OrganizationHomePayload,
} from "@/lib/organizations";
import { AutomationDetailPanel } from "@/components/organizations/organizationAutomationPanel";
import {
    advisorDetailItems,
    aiEngineDetailItems,
} from "@/components/organizations/organizationDetailHelpers";
import { responseContractDetailItems } from "@/components/organizations/organizationProfileOptions";
import { OutputModelRoutingPanel } from "@/components/organizations/OrganizationOutputModelRoutingPanel";
import {
    AIEngineSelectionPanel,
    ResponseContractSelectionPanel,
} from "@/components/organizations/OrganizationProfileSelectionPanels";
import { OrganizationDepartmentDetailSection } from "@/components/organizations/OrganizationDepartmentDetailSection";
import type {
    OrganizationAIEngineAndRoutingControls,
    OrganizationDetailViewId,
} from "@/components/organizations/useOrganizationAIEngineAndRouting";
import type { DepartmentAndAgentTypeOverrideControls } from "@/components/organizations/useDepartmentAndAgentTypeOverrides";

export function WorkspaceDetailView({
    view,
    organization,
    teamLeadName,
    automations,
    automationsLoading,
    automationsError,
    onRetryAutomations,
    aiEngineRouting,
    departmentOverrides,
    onBack,
}: {
    view: OrganizationDetailViewId;
    organization: OrganizationHomePayload;
    teamLeadName: string;
    automations: OrganizationAutomationItem[];
    automationsLoading: boolean;
    automationsError: string | null;
    onRetryAutomations: () => void;
    aiEngineRouting: OrganizationAIEngineAndRoutingControls;
    departmentOverrides: DepartmentAndAgentTypeOverrideControls;
    onBack: () => void;
}) {
    const {
        isAIEngineSelectorOpen,
        selectedAIEngineProfile,
        aiEngineUpdatePending,
        aiEngineUpdateError,
        outputModelRouting,
        outputModelRoutingLoading,
        outputModelRoutingError,
        selectedOutputModelRoutingMode,
        selectedDefaultOutputModelId,
        selectedOutputModelBindings,
        outputModelRoutingUpdatePending,
        outputModelRoutingUpdateError,
        onAIEngineProfileSelect,
        onOutputModelRoutingModeChange,
        onDefaultOutputModelChange,
        onOutputModelBindingChange,
        onOpenAIEngineSelector,
        onCloseAIEngineSelector,
        onSubmitAIEngineSelection,
        onSubmitOutputModelRouting,
        isResponseContractSelectorOpen,
        selectedResponseContractProfile,
        responseContractUpdatePending,
        responseContractUpdateError,
        onResponseContractProfileSelect,
        onOpenResponseContractSelector,
        onCloseResponseContractSelector,
        onSubmitResponseContractSelection,
    } = aiEngineRouting;

    const title =
        view === "advisors"
            ? "Advisor details"
            : view === "departments"
              ? "Department details"
              : view === "automations"
                ? "Automation details"
              : view === "aiEngine"
                ? "AI Engine Settings details"
                : "Response Style details";
    const summary =
        view === "advisors"
            ? `Soma can review the current Advisor support in ${organization.name} here without leaving the workspace.`
            : view === "departments"
              ? `Soma can inspect the current Department structure in ${organization.name} here without leaving the workspace while keeping ${teamLeadName} in view.`
              : view === "automations"
                ? `Soma can inspect the ongoing reviews and checks that support ${organization.name} here without leaving the workspace.`
              : view === "aiEngine"
                ? `Soma can inspect the current AI Engine Settings in ${organization.name} here without leaving the workspace.`
                : `Soma can inspect the current Response Style in ${organization.name} here without leaving the workspace.`;
    const items =
        view === "advisors"
            ? advisorDetailItems(organization.advisor_count).map((item) => ({
                  key: item.name,
                  title: item.name,
                  detail: item.purpose,
                  support: item.supportCue,
              }))
            : view === "automations"
              ? []
            : (view === "aiEngine" ? aiEngineDetailItems(organization) : responseContractDetailItems(organization.response_contract_summary)).map((item) => ({
                  key: item.name,
                  title: item.name,
                  detail: item.purpose,
                  support: item.supportCue,
              }));
    const automationItems = view === "automations" ? automations : [];
    const emptyStateMessage =
        view === "advisors"
            ? "Advisor support will appear here when Soma starts working with additional review help. Try reviewing your organization setup to decide where that support is needed first."
            : view === "departments"
              ? "Departments are the working lanes for this organization. They will appear here once the first lane is defined. Try reviewing your organization setup or running a quick strategy check."
              : view === "automations"
                ? "Automations will appear here as ongoing reviews and checks become available for this AI Organization. Run a quick strategy check or review your organization setup to start creating visible signals."
              : view === "aiEngine"
                ? `Soma is still using the shared AI engine profile until more scoped settings are surfaced here.`
                : `Soma is still using the organization-wide Response Style until a different guided profile is selected here.`;

    return (
        <div className="rounded-3xl border border-cortex-border bg-cortex-surface p-6">
            <div className="flex flex-col gap-4 lg:flex-row lg:items-start lg:justify-between">
                <div>
                    <h3 className="text-xl font-semibold text-cortex-text-main">{title}</h3>
                    <p className="mt-2 text-sm leading-7 text-cortex-text-muted">{summary}</p>
                </div>
                <div className="flex flex-wrap gap-3">
                    {view === "aiEngine" && (
                        <button
                            type="button"
                            onClick={isAIEngineSelectorOpen ? onCloseAIEngineSelector : onOpenAIEngineSelector}
                            className="inline-flex items-center gap-2 rounded-xl border border-cortex-primary/30 bg-cortex-primary/10 px-3 py-2 text-sm font-medium text-cortex-text-main transition-colors hover:border-cortex-primary/40"
                        >
                            <Sparkles className="h-4 w-4 text-cortex-primary" />
                            {isAIEngineSelectorOpen ? "Close AI Engine choices" : "Change AI Engine"}
                        </button>
                    )}
                    {view === "responseContract" && (
                        <button
                            type="button"
                            onClick={isResponseContractSelectorOpen ? onCloseResponseContractSelector : onOpenResponseContractSelector}
                            className="inline-flex items-center gap-2 rounded-xl border border-cortex-primary/30 bg-cortex-primary/10 px-3 py-2 text-sm font-medium text-cortex-text-main transition-colors hover:border-cortex-primary/40"
                        >
                            <Sparkles className="h-4 w-4 text-cortex-primary" />
                            {isResponseContractSelectorOpen ? "Close Response Style choices" : "Change Response Style"}
                        </button>
                    )}
                    <button
                        type="button"
                        onClick={onBack}
                        className="inline-flex items-center gap-2 rounded-xl border border-cortex-border bg-cortex-bg px-3 py-2 text-sm font-medium text-cortex-text-main transition-colors hover:border-cortex-primary/20"
                    >
                        <ArrowLeft className="h-4 w-4" />
                        Back to Soma
                    </button>
                </div>
            </div>
            {view === "departments" ? (
                <OrganizationDepartmentDetailSection
                    organization={organization}
                    teamLeadName={teamLeadName}
                    departmentOverrides={departmentOverrides}
                />
            ) : view === "automations" ? (
                <AutomationDetailPanel
                    items={automationItems}
                    loading={automationsLoading}
                    error={automationsError}
                    onRetry={onRetryAutomations}
                />
            ) : items.length > 0 ? (
                <div className="mt-5 grid gap-3">
                    {items.map((item) => (
                        <div key={item.key} className="rounded-2xl border border-cortex-border bg-cortex-bg px-4 py-4">
                            <p className="text-sm font-semibold text-cortex-text-main">{item.title}</p>
                            <p className="mt-2 text-sm leading-6 text-cortex-text-muted">{item.detail}</p>
                            <p className="mt-2 text-sm leading-6 text-cortex-text-muted">{item.support}</p>
                        </div>
                    ))}
                </div>
            ) : (
                <div className="mt-5 rounded-2xl border border-cortex-border bg-cortex-bg px-4 py-4 text-sm leading-6 text-cortex-text-muted">
                    {emptyStateMessage}
                </div>
            )}

            {view === "aiEngine" && isAIEngineSelectorOpen && (
                <AIEngineSelectionPanel
                    selectedAIEngineProfile={selectedAIEngineProfile}
                    aiEngineUpdatePending={aiEngineUpdatePending}
                    aiEngineUpdateError={aiEngineUpdateError}
                    onAIEngineProfileSelect={onAIEngineProfileSelect}
                    onClose={onCloseAIEngineSelector}
                    onSubmit={onSubmitAIEngineSelection}
                />
            )}

            {view === "aiEngine" && (
                <OutputModelRoutingPanel
                    routing={outputModelRouting}
                    loading={outputModelRoutingLoading}
                    error={outputModelRoutingError}
                    selectedRoutingMode={selectedOutputModelRoutingMode}
                    selectedDefaultModelId={selectedDefaultOutputModelId}
                    selectedBindings={selectedOutputModelBindings}
                    updatePending={outputModelRoutingUpdatePending}
                    updateError={outputModelRoutingUpdateError}
                    onRoutingModeChange={onOutputModelRoutingModeChange}
                    onDefaultModelChange={onDefaultOutputModelChange}
                    onBindingChange={onOutputModelBindingChange}
                    onSubmit={onSubmitOutputModelRouting}
                />
            )}

            {view === "responseContract" && isResponseContractSelectorOpen && (
                <ResponseContractSelectionPanel
                    selectedResponseContractProfile={selectedResponseContractProfile}
                    responseContractUpdatePending={responseContractUpdatePending}
                    responseContractUpdateError={responseContractUpdateError}
                    onResponseContractProfileSelect={onResponseContractProfileSelect}
                    onClose={onCloseResponseContractSelector}
                    onSubmit={onSubmitResponseContractSelection}
                />
            )}
        </div>
    );
}
