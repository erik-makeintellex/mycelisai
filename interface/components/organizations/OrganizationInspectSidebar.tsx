"use client";

import Link from "next/link";
import { ArrowLeft, Blocks, Bot, BrainCircuit, Building2, Sparkles, Users } from "lucide-react";
import type {
    OrganizationAutomationItem,
    OrganizationHomePayload,
    OrganizationLearningInsightItem,
} from "@/lib/organizations";
import { CausalFact } from "@/components/organizations/organizationActivityPanels";
import {
    advisorSummary,
    advisorSupportItems,
    aiEngineSummary,
    aiEngineSupportItems,
    automationStatusLabel,
    automationSummary,
    automationSupportItems,
    departmentSummary,
    departmentSupportItems,
    formatAutomationCount,
    formatConfiguredCount,
    learningContextSummary,
    learningContextSupportItems,
    responseContractSummary,
    responseContractSupportItems,
} from "@/components/organizations/organizationSummaryHelpers";
import type { OrganizationDetailViewId } from "@/components/organizations/useOrganizationAIEngineAndRouting";

export function OrganizationInspectSidebar({
    organization,
    teamLeadName,
    automations,
    automationsLoading,
    automationsError,
    learningInsights,
    onSelectDetailView,
}: {
    organization: OrganizationHomePayload;
    teamLeadName: string;
    automations: OrganizationAutomationItem[];
    automationsLoading: boolean;
    automationsError: string | null;
    learningInsights: OrganizationLearningInsightItem[];
    onSelectDetailView: (view: OrganizationDetailViewId) => void;
}) {
    return (
        <div className="space-y-4">
            <InspectOnlySummary
                icon={<Users className="h-4 w-4" />}
                title="Advisors"
                countLabel={formatConfiguredCount(organization.advisor_count, "Advisor")}
                summary={advisorSummary(organization.advisor_count, teamLeadName)}
                supportLabel="Advisor support"
                items={advisorSupportItems(organization.advisor_count)}
                changeSummary={organization.advisor_count > 0 ? `${formatConfiguredCount(organization.advisor_count, "Advisor")} visible right now` : "No advisor support configured yet"}
                changeReason="Advisor support expands when Soma needs more review coverage around current work."
                somaConnection="Soma uses advisor support for decision review, priority checks, and quality assurance."
                inspectActionLabel="Review Advisors"
                onInspect={() => onSelectDetailView("advisors")}
            />

            <InspectOnlySummary
                icon={<Building2 className="h-4 w-4" />}
                title="Departments"
                countLabel={formatConfiguredCount(organization.department_count, "Department")}
                summary={departmentSummary(organization.department_count, organization.specialist_count, teamLeadName)}
                supportLabel="Visible specialist roles"
                items={departmentSupportItems(organization)}
                changeSummary={organization.department_count > 0 ? `${formatConfiguredCount(organization.department_count, "Department")} shaping delivery` : "No working lanes defined yet"}
                changeReason="Departments and specialist roles show the execution structure Soma can work through next."
                somaConnection="Soma uses Departments and Specialists to turn intent into concrete delivery lanes."
                inspectActionLabel="Open Departments"
                onInspect={() => onSelectDetailView("departments")}
            />

            <InspectOnlySummary
                icon={<Blocks className="h-4 w-4" />}
                title="Automations"
                countLabel={formatAutomationCount(automations.length, automationsLoading, automationsError)}
                statusLabel={automationStatusLabel(automationsLoading, automationsError)}
                summary={automationSummary(automations.length, teamLeadName)}
                supportLabel="What these cover"
                items={automationSupportItems(automations, automationsLoading, automationsError)}
                changeSummary={automationsError ? "Automation details unavailable right now" : formatAutomationCount(automations.length, automationsLoading, automationsError)}
                changeReason="Automations refresh as ongoing reviews and checks become visible around the organization."
                somaConnection="Soma uses Automations to explain what is watching, reviewing, and checking work in the background."
                inspectActionLabel="Review Automations"
                onInspect={() => onSelectDetailView("automations")}
            />

            <div className="rounded-3xl border border-cortex-border bg-cortex-surface p-6">
                <div className="grid gap-4">
                    <InspectOnlySummary
                        icon={<Bot className="h-4 w-4" />}
                        title="AI Engine Settings"
                        countLabel="Guided tuning"
                        statusLabel="Organization level"
                        summary={aiEngineSummary(organization.ai_engine_settings_summary)}
                        supportLabel="What this affects"
                        items={aiEngineSupportItems(organization.ai_engine_settings_summary)}
                        changeSummary={`Current profile: ${organization.ai_engine_settings_summary || "Starter Defaults"}`}
                        changeReason="This guided choice shapes how deeply Soma plans, responds, and carries work forward."
                        somaConnection="Soma uses these AI Engine settings as the organization-wide operating posture."
                        inspectActionLabel="Review AI Engine Settings"
                        onInspect={() => onSelectDetailView("aiEngine")}
                    />
                    <InspectOnlySummary
                        icon={<BrainCircuit className="h-4 w-4" />}
                        title="Response Style"
                        countLabel="Guided tuning"
                        statusLabel="Organization level"
                        summary={responseContractSummary(organization.response_contract_summary)}
                        supportLabel="What this shapes"
                        items={responseContractSupportItems(organization.response_contract_summary)}
                        changeSummary={`Current profile: ${organization.response_contract_summary || "Clear & Balanced"}`}
                        changeReason="This guided choice shapes how Soma sounds, structures responses, and presents detail."
                        somaConnection="Soma uses this Response Style to keep the organization consistent and readable."
                        inspectActionLabel="Review Response Style"
                        onInspect={() => onSelectDetailView("responseContract")}
                    />
                    <InspectOnlySummary
                        icon={<BrainCircuit className="h-4 w-4" />}
                        title="Memory & Continuity"
                        countLabel="Inspect only"
                        summary={learningContextSummary(organization.memory_personality_summary)}
                        supportLabel="What this affects"
                        items={learningContextSupportItems(organization.memory_personality_summary)}
                        changeSummary={learningInsights.length > 0 ? `${learningInsights.length} retained pattern${learningInsights.length === 1 ? "" : "s"} visible` : "No retained patterns visible yet"}
                        changeReason="Reusable patterns and continuity cues become visible as Soma and supporting reviews build repeatable signals."
                        somaConnection="Soma uses durable memory for reusable recall and temporary continuity to stay oriented across return visits and later work."
                    />
                </div>
                <div className="mt-5">
                    <Link href="/dashboard" className="inline-flex items-center gap-2 text-cortex-primary hover:underline">
                        <ArrowLeft className="h-4 w-4" />
                        Create another AI Organization
                    </Link>
                </div>
            </div>
        </div>
    );
}

function InspectOnlySummary({
    icon,
    title,
    countLabel,
    summary,
    supportLabel,
    items,
    changeSummary,
    changeReason,
    somaConnection,
    inspectActionLabel,
    onInspect,
    statusLabel = "Inspect only",
}: {
    icon: React.ReactNode;
    title: string;
    countLabel: string;
    summary: string;
    supportLabel: string;
    items: Array<string | { label: string; detail?: string }>;
    changeSummary: string;
    changeReason: string;
    somaConnection: string;
    inspectActionLabel?: string;
    onInspect?: () => void;
    statusLabel?: string;
}) {
    return (
        <div className="rounded-3xl border border-cortex-border bg-cortex-surface p-6">
            <div className="flex items-start justify-between gap-4">
                <div>
                    <div className="flex items-center gap-2 text-cortex-primary">
                        {icon}
                        <h2 className="text-xl font-semibold text-cortex-text-main">{title}</h2>
                    </div>
                    <p className="mt-1 text-sm text-cortex-text-muted">{summary}</p>
                </div>
                <div className="rounded-2xl border border-cortex-border bg-cortex-bg px-4 py-3 text-sm text-cortex-text-muted">
                    <p className="font-medium text-cortex-text-main">{countLabel}</p>
                    <p className="mt-1">{statusLabel}</p>
                </div>
            </div>
            <div className="mt-5">
                <div className="grid gap-3 lg:grid-cols-3">
                    <CausalFact label="What changed" value={changeSummary} />
                    <CausalFact label="Why it changed" value={changeReason} />
                    <CausalFact label="How Soma uses it" value={somaConnection} />
                </div>
            </div>
            <div className="mt-5">
                <p className="text-sm font-medium text-cortex-text-main">{supportLabel}</p>
                <div className="mt-3 grid gap-2">
                    {items.map((item) => (
                        <div
                            key={typeof item === "string" ? item : `${item.label}-${item.detail ?? ""}`}
                            className="rounded-2xl border border-cortex-border bg-cortex-bg px-3 py-3 text-sm text-cortex-text-main"
                        >
                            {typeof item === "string" ? (
                                item
                            ) : (
                                <>
                                    <p className="font-medium text-cortex-text-main">{item.label}</p>
                                    {item.detail ? <p className="mt-1 text-sm leading-6 text-cortex-text-muted">{item.detail}</p> : null}
                                </>
                            )}
                        </div>
                    ))}
                </div>
            </div>
            {onInspect && inspectActionLabel && (
                <div className="mt-5">
                    <button
                        type="button"
                        onClick={onInspect}
                        className="inline-flex items-center gap-2 rounded-xl border border-cortex-border bg-cortex-bg px-3 py-2 text-sm font-medium text-cortex-text-main transition-colors hover:border-cortex-primary/20"
                    >
                        <Sparkles className="h-4 w-4 text-cortex-primary" />
                        {inspectActionLabel}
                    </button>
                </div>
            )}
        </div>
    );
}
