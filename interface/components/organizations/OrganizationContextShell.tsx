"use client";

import Link from "next/link";
import { useEffect, useState } from "react";
import { Activity, ArrowLeft, Blocks, BrainCircuit, Loader2, RefreshCcw, Sparkles, Users } from "lucide-react";
import { extractApiData, extractApiError } from "@/lib/apiContracts";
import { rememberLastOrganization } from "@/lib/lastOrganization";
import type { OrganizationHomePayload } from "@/lib/organizations";
import { SomaOperatingSurface } from "@/components/soma/SomaOperatingSurface";
import { useCortexStore } from "@/store/useCortexStore";
import { LearningVisibilityPanel, RecentActivityPanel } from "@/components/organizations/organizationActivityPanels";
import { extractOutputsFromConversation, findLatestConversationOutcome } from "@/components/organizations/conversationOutcomeHelpers";
import { useOrganizationLivePanelData } from "@/components/organizations/useOrganizationLivePanelData";
import { ActionPill } from "@/components/organizations/organizationWorkspaceChrome";
import { OrganizationStarterIntro } from "@/components/organizations/OrganizationStarterIntro";
import { OrganizationInspectSidebar } from "@/components/organizations/OrganizationInspectSidebar";
import { WorkspaceDetailView } from "@/components/organizations/OrganizationWorkspaceDetailView";
import { useOrganizationAIEngineAndRouting, type OrganizationDetailViewId } from "@/components/organizations/useOrganizationAIEngineAndRouting";
import { useDepartmentAndAgentTypeOverrides } from "@/components/organizations/useDepartmentAndAgentTypeOverrides";

export { extractOutputsFromConversation, findLatestConversationOutcome };

async function readJson(response: Response) {
    try {
        return await response.json();
    } catch {
        return null;
    }
}

export default function OrganizationContextShell({ organizationId }: { organizationId: string }) {
    const [organization, setOrganization] = useState<OrganizationHomePayload | null>(null);
    const [loading, setLoading] = useState(true);
    const [error, setError] = useState<string | null>(null);
    const [retryToken, setRetryToken] = useState(0);
    const [activeDetailView, setActiveDetailView] = useState<OrganizationDetailViewId | null>(null);

    const {
        recentActivity,
        activityLoading,
        activityError,
        retryActivity,
        automations,
        automationsLoading,
        automationsError,
        retryAutomations,
        learningInsights,
        learningInsightsLoading,
        learningInsightsError,
        retryLearningInsights,
    } = useOrganizationLivePanelData(organizationId);
    const missionChat = useCortexStore((s) => s.missionChat);

    const aiEngineRouting = useOrganizationAIEngineAndRouting({
        organizationId,
        retryToken,
        organization,
        setOrganization,
        activeDetailView,
        setActiveDetailView,
    });

    const departmentOverrides = useDepartmentAndAgentTypeOverrides({
        organization,
        setOrganization,
        activeDetailView,
        setActiveDetailView,
    });

    useEffect(() => {
        let cancelled = false;
        const load = async () => {
            setLoading(true);
            setError(null);
            try {
                const response = await fetch(`/api/v1/organizations/${organizationId}/home`, { cache: "no-store" });
                const payload = await readJson(response);
                if (!response.ok) throw new Error(extractApiError(payload) || "Unable to load AI Organization.");
                if (!cancelled) setOrganization(extractApiData<OrganizationHomePayload>(payload));
            } catch (err) {
                if (!cancelled) setError(err instanceof Error ? err.message : "Unable to load AI Organization.");
            } finally {
                if (!cancelled) setLoading(false);
            }
        };
        void load();
        return () => {
            cancelled = true;
        };
    }, [organizationId, retryToken]);

    useEffect(() => {
        if (typeof window === "undefined" || !organization) return;
        rememberLastOrganization({ id: organization.id, name: organization.name });
    }, [organization]);

    if (loading) {
        return (
            <div className="flex h-full items-center justify-center bg-cortex-bg">
                <div className="flex items-center gap-3 rounded-2xl border border-cortex-border bg-cortex-surface px-5 py-4 text-sm text-cortex-text-muted">
                    <Loader2 className="h-4 w-4 animate-spin" />
                    Loading AI Organization...
                </div>
            </div>
        );
    }

    if (error || !organization) {
        return (
            <div className="flex h-full items-center justify-center bg-cortex-bg px-6">
                <div className="max-w-lg rounded-3xl border border-cortex-danger/30 bg-cortex-surface p-6">
                    <p className="text-lg font-semibold text-cortex-text-main">AI Organization unavailable</p>
                    <p className="mt-2 text-sm leading-7 text-cortex-text-muted">{error || "This AI Organization could not be loaded."}</p>
                    <p className="mt-3 text-sm leading-7 text-cortex-text-muted">
                        Try loading the organization again, or return to the setup screen to create a new AI Organization.
                    </p>
                    <div className="mt-5 flex flex-wrap gap-3">
                        <button
                            onClick={() => setRetryToken((value) => value + 1)}
                            className="inline-flex items-center gap-2 rounded-xl border border-cortex-border bg-cortex-bg px-3 py-2 text-sm font-medium text-cortex-text-main transition-colors hover:border-cortex-primary/20"
                        >
                            <RefreshCcw className="h-4 w-4" />
                            Retry
                        </button>
                        <Link href="/dashboard" className="inline-flex items-center gap-2 rounded-xl border border-transparent px-3 py-2 text-sm font-medium text-cortex-primary hover:bg-cortex-primary/10">
                            <ArrowLeft className="h-4 w-4" />
                            Return to Create AI Organization
                        </Link>
                    </div>
                </div>
            </div>
        );
    }

    const somaName = `Soma for ${organization.name}`;
    const teamLeadName = `${organization.team_lead_label} for ${organization.name}`;
    const lastConversationAction = [...missionChat].reverse().find((message) => message.role === "user" && !message.content.startsWith("[BROADCAST]"));
    const lastConversationOutcome = findLatestConversationOutcome(missionChat, teamLeadName);
    const causalAction = lastConversationOutcome
        ? lastConversationOutcome.actionLabel
        : lastConversationAction
          ? lastConversationAction.content
          : "Ready for your first Soma request";

    return (
        <div className="h-full overflow-auto bg-cortex-bg px-6 py-8">
            <div className="mx-auto max-w-6xl space-y-8">
                <section className="rounded-2xl border border-cortex-border bg-cortex-surface px-5 py-4">
                    <div className="flex flex-col gap-3 lg:flex-row lg:items-center lg:justify-between">
                        <nav aria-label="Organization breadcrumb" className="flex flex-wrap items-center gap-2 text-sm text-cortex-text-muted">
                            <Link href="/dashboard" className="font-medium text-cortex-primary transition-colors hover:text-cortex-primary/80">
                                AI Organizations
                            </Link>
                            <span>/</span>
                            <span className="font-medium text-cortex-text-main">{organization.name}</span>
                        </nav>
                        <Link
                            href="/dashboard"
                            className="inline-flex items-center gap-2 rounded-xl border border-cortex-border bg-cortex-bg px-3 py-2 text-sm font-medium text-cortex-text-main transition-colors hover:border-cortex-primary/20"
                        >
                            <ArrowLeft className="h-4 w-4" />
                            Return to Organization list
                        </Link>
                    </div>
                </section>

                <section className="rounded-3xl border border-cortex-border bg-cortex-surface px-6 py-8 shadow-[0_18px_40px_rgba(148,163,184,0.16)]">
                    <div className="flex flex-col gap-5 lg:flex-row lg:items-end lg:justify-between">
                        <div className="space-y-3">
                            <div className="inline-flex items-center gap-2 rounded-full border border-cortex-primary/25 bg-cortex-primary/10 px-3 py-1 text-[11px] font-mono uppercase tracking-[0.22em] text-cortex-primary">
                                <Blocks className="h-3.5 w-3.5" />
                                AI Organization Home
                            </div>
                            <div>
                                <h1 className="text-4xl font-semibold tracking-tight text-cortex-text-main">{organization.name}</h1>
                                <p className="mt-2 max-w-3xl text-base leading-7 text-cortex-text-muted">{organization.purpose}</p>
                            </div>
                        </div>
                        <div className="rounded-2xl border border-cortex-border bg-cortex-bg px-4 py-3 text-sm text-cortex-text-muted lg:max-w-sm">
                            <p className="font-medium text-cortex-text-main">Soma ready</p>
                            <p className="mt-1 leading-6">
                                {somaName} is ready to guide planning, structure review, and organization setup decisions while coordinating the right teams when needed.
                            </p>
                            <a
                                href="#soma-panel"
                                className="mt-4 inline-flex items-center gap-2 rounded-xl border border-cortex-primary/35 bg-cortex-primary/10 px-3 py-2 font-medium text-cortex-primary transition-colors hover:bg-cortex-primary/15"
                            >
                                Start with Soma
                                <Sparkles className="h-4 w-4" />
                            </a>
                        </div>
                    </div>
                </section>

                <OrganizationStarterIntro somaIntro={organization.soma_intro} deliveryDefaults={organization.delivery_defaults} />
                <section className="space-y-4">
                    <div className="space-y-4">
                        <div className="rounded-3xl border border-cortex-border bg-cortex-surface p-6">
                            <div className="flex flex-col gap-4 lg:flex-row lg:items-start lg:justify-between">
                                <div className="space-y-3">
                                    <div className="inline-flex items-center gap-2 rounded-full border border-cortex-primary/20 bg-cortex-primary/10 px-3 py-1 text-[11px] font-mono uppercase tracking-[0.18em] text-cortex-primary">
                                        <Sparkles className="h-3.5 w-3.5" />
                                        Soma
                                    </div>
                                    <div>
                                        <h2 className="text-2xl font-semibold text-cortex-text-main">{somaName}</h2>
                                        <p className="mt-2 max-w-2xl text-sm leading-7 text-cortex-text-muted">
                                            Soma is the primary counterpart for {organization.name}, coordinating the right teams, advisors, departments, and specialists around the organization purpose.
                                        </p>
                                    </div>
                                </div>
                                <div className="rounded-2xl border border-cortex-border bg-cortex-bg px-4 py-3 text-sm text-cortex-text-muted lg:max-w-xs">
                                    <p className="font-medium text-cortex-text-main">Operational layer</p>
                                    <p className="mt-1 leading-6">
                                        {teamLeadName} leads day-to-day operational follow-through while Soma keeps the organization focused and surfaces the next best actions.
                                    </p>
                                </div>
                            </div>

                            <div className="mt-6">
                                <p className="text-sm font-medium text-cortex-text-main">Inspect the current organization</p>
                                <div className="mt-3 flex flex-wrap gap-2">
                                    <ActionPill label="Review Advisors" isActive={activeDetailView === "advisors"} onClick={() => setActiveDetailView("advisors")} />
                                    <ActionPill label="Open Departments" isActive={activeDetailView === "departments"} onClick={() => setActiveDetailView("departments")} />
                                    <ActionPill label="Review Automations" isActive={activeDetailView === "automations"} onClick={() => setActiveDetailView("automations")} />
                                    <ActionPill label="Review AI Engine Settings" isActive={activeDetailView === "aiEngine"} onClick={() => setActiveDetailView("aiEngine")} />
                                    <ActionPill label="Review Response Style" isActive={activeDetailView === "responseContract"} onClick={() => setActiveDetailView("responseContract")} />
                                </div>
                            </div>
                        </div>

                        {activeDetailView && (
                            <WorkspaceDetailView
                                view={activeDetailView}
                                organization={organization}
                                teamLeadName={teamLeadName}
                                automations={automations}
                                automationsLoading={automationsLoading}
                                automationsError={automationsError}
                                onRetryAutomations={retryAutomations}
                                aiEngineRouting={aiEngineRouting}
                                departmentOverrides={departmentOverrides}
                                onBack={() => setActiveDetailView(null)}
                            />
                        )}

                        <SomaOperatingSurface
                            organizationId={organizationId}
                            organizationName={organization.name}
                            activeMode="Organization work"
                            evidenceItems={[
                                { title: "Activity", detail: "What changed, why it changed, and which run or signal created the visible outcome.", href: "/activity", icon: <Activity className="h-4 w-4" /> },
                                { title: "Learning and context", detail: "Reusable patterns and continuity cues Soma can use on later work.", href: "/memory", icon: <BrainCircuit className="h-4 w-4" /> },
                                { title: "Teams and specialists", detail: "Departments and specialist roles Soma can coordinate when intent needs structure.", icon: <Users className="h-4 w-4" /> },
                                { title: "Tools and capabilities", detail: "Connected tools, search readiness, and governed capability boundaries.", href: "/resources?tab=tools", icon: <Blocks className="h-4 w-4" /> },
                            ]}
                        />
                    </div>

                    <section className="grid gap-4 lg:grid-cols-[minmax(0,1.1fr)_minmax(320px,0.9fr)]">
                        <div className="space-y-4">
                            <RecentActivityPanel items={recentActivity} loading={activityLoading} error={activityError} onRetry={retryActivity} causalAction={causalAction} />
                            <LearningVisibilityPanel items={learningInsights} loading={learningInsightsLoading} error={learningInsightsError} onRetry={retryLearningInsights} causalAction={causalAction} />
                        </div>
                        <OrganizationInspectSidebar
                            organization={organization}
                            teamLeadName={teamLeadName}
                            automations={automations}
                            automationsLoading={automationsLoading}
                            automationsError={automationsError}
                            learningInsights={learningInsights}
                            onSelectDetailView={setActiveDetailView}
                        />
                    </section>
                </section>
            </div>
        </div>
    );
}
