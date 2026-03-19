"use client";

import Link from "next/link";
import { useEffect, useState } from "react";
import { ArrowLeft, Blocks, Bot, BrainCircuit, Building2, Loader2, RefreshCcw, Sparkles, Users } from "lucide-react";
import { extractApiData, extractApiError } from "@/lib/apiContracts";
import type { OrganizationHomePayload } from "@/lib/organizations";
import TeamLeadInteractionPanel from "@/components/organizations/TeamLeadInteractionPanel";

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
    const [activeDetailView, setActiveDetailView] = useState<"advisors" | "departments" | null>(null);

    useEffect(() => {
        let cancelled = false;

        const load = async () => {
            setLoading(true);
            setError(null);
            try {
                const response = await fetch(`/api/v1/organizations/${organizationId}/home`, { cache: "no-store" });
                const payload = await readJson(response);
                if (!response.ok) {
                    throw new Error(extractApiError(payload) || "Unable to load AI Organization.");
                }
                if (cancelled) {
                    return;
                }
                setOrganization(extractApiData<OrganizationHomePayload>(payload));
            } catch (err) {
                if (cancelled) {
                    return;
                }
                setError(err instanceof Error ? err.message : "Unable to load AI Organization.");
            } finally {
                if (!cancelled) {
                    setLoading(false);
                }
            }
        };

        void load();
        return () => {
            cancelled = true;
        };
    }, [organizationId, retryToken]);

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

    const teamLeadName = `${organization.team_lead_label} for ${organization.name}`;
    const overviewItems = [
        { label: "Started from", value: organization.start_mode === "template" ? (organization.template_name || "Template") : "Empty" },
        { label: "Advisors", value: formatConfiguredCount(organization.advisor_count, "Advisor") },
        { label: "Departments", value: formatConfiguredCount(organization.department_count, "Department") },
        { label: "Specialists", value: formatConfiguredCount(organization.specialist_count, "Specialist") },
        { label: "AI Organization", value: toTitleCase(organization.status) },
    ];

    return (
        <div className="h-full overflow-auto bg-cortex-bg px-6 py-8">
            <div className="mx-auto max-w-6xl space-y-8">
                <section className="rounded-3xl border border-cortex-border bg-cortex-surface px-6 py-8 shadow-[0_24px_60px_rgba(0,0,0,0.18)]">
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
                            <p className="font-medium text-cortex-text-main">Team Lead ready</p>
                            <p className="mt-1 leading-6">
                                {teamLeadName} is ready to guide planning, structure review, and organization setup decisions without leaving the AI Organization frame.
                            </p>
                        </div>
                    </div>
                </section>

                <section className="grid gap-4 lg:grid-cols-[1.2fr_0.8fr]">
                    <div className="space-y-4">
                        <div className="rounded-3xl border border-cortex-border bg-cortex-surface p-6">
                            <div className="flex flex-col gap-4 lg:flex-row lg:items-start lg:justify-between">
                                <div className="space-y-3">
                                    <div className="inline-flex items-center gap-2 rounded-full border border-cortex-primary/20 bg-cortex-primary/10 px-3 py-1 text-[11px] font-mono uppercase tracking-[0.18em] text-cortex-primary">
                                        <Sparkles className="h-3.5 w-3.5" />
                                        Team Lead
                                    </div>
                                    <div>
                                        <h2 className="text-2xl font-semibold text-cortex-text-main">{teamLeadName}</h2>
                                        <p className="mt-2 max-w-2xl text-sm leading-7 text-cortex-text-muted">
                                            Your Team Lead is the working counterpart for {organization.name}, coordinating Advisors, Departments, and Specialists around the organization purpose.
                                        </p>
                                    </div>
                                </div>
                                <div className="rounded-2xl border border-cortex-border bg-cortex-bg px-4 py-3 text-sm text-cortex-text-muted lg:max-w-xs">
                                    <p className="font-medium text-cortex-text-main">Current role</p>
                                    <p className="mt-1 leading-6">
                                        Keep the organization focused, surface the next best actions, and help the operator move from setup into coordinated delivery.
                                    </p>
                                </div>
                            </div>

                            <div className="mt-6">
                                <p className="text-sm font-medium text-cortex-text-main">What I can help with</p>
                                <div className="mt-3 flex flex-wrap gap-2">
                                    <HelpPill label="Plan the next steps" />
                                    <ActionPill
                                        label="Review Advisors"
                                        isActive={activeDetailView === "advisors"}
                                        onClick={() => setActiveDetailView("advisors")}
                                    />
                                    <ActionPill
                                        label="Open Departments"
                                        isActive={activeDetailView === "departments"}
                                        onClick={() => setActiveDetailView("departments")}
                                    />
                                </div>
                            </div>
                        </div>

                        {activeDetailView && (
                            <WorkspaceDetailView
                                view={activeDetailView}
                                organization={organization}
                                teamLeadName={teamLeadName}
                                onBack={() => setActiveDetailView(null)}
                            />
                        )}

                        <TeamLeadInteractionPanel
                            organizationId={organization.id}
                            organizationName={organization.name}
                            teamLeadName={teamLeadName}
                        />
                    </div>

                    <div className="space-y-4">
                        <div className="rounded-3xl border border-cortex-border bg-cortex-surface p-6">
                            <div>
                                <h2 className="text-xl font-semibold text-cortex-text-main">Organization overview</h2>
                                <p className="mt-1 text-sm text-cortex-text-muted">See the Team Lead, structure, and starting point for this AI Organization at a glance.</p>
                            </div>
                            <div className="mt-4 grid gap-3 md:grid-cols-2">
                                {overviewItems.map((item) => (
                                    <Metric key={item.label} label={item.label} value={item.value} />
                                ))}
                            </div>
                        </div>

                        <InspectOnlySummary
                            icon={<Users className="h-4 w-4" />}
                            title="Advisors"
                            countLabel={formatConfiguredCount(organization.advisor_count, "Advisor")}
                            summary={advisorSummary(organization.advisor_count, teamLeadName)}
                            supportLabel="Advisor support"
                            items={advisorSupportItems(organization.advisor_count)}
                            inspectActionLabel="Review Advisors"
                            onInspect={() => setActiveDetailView("advisors")}
                        />

                        <InspectOnlySummary
                            icon={<Building2 className="h-4 w-4" />}
                            title="Departments"
                            countLabel={formatConfiguredCount(organization.department_count, "Department")}
                            summary={departmentSummary(organization.department_count, organization.specialist_count, teamLeadName)}
                            supportLabel="Department view"
                            items={departmentSupportItems(organization)}
                            inspectActionLabel="Open Departments"
                            onInspect={() => setActiveDetailView("departments")}
                        />

                        <div className="rounded-3xl border border-cortex-border bg-cortex-surface p-6">
                            <div className="grid gap-4">
                                <InspectOnlySummary
                                    icon={<Bot className="h-4 w-4" />}
                                    title="AI Engine Settings"
                                    countLabel="Inspect only"
                                    summary={aiEngineSummary(organization.ai_engine_settings_summary)}
                                    supportLabel="What this affects"
                                    items={aiEngineSupportItems(organization.ai_engine_settings_summary)}
                                />
                                <InspectOnlySummary
                                    icon={<BrainCircuit className="h-4 w-4" />}
                                    title="Memory & Personality"
                                    countLabel="Inspect only"
                                    summary={memoryPersonalitySummary(organization.memory_personality_summary)}
                                    supportLabel="What this affects"
                                    items={memoryPersonalitySupportItems(organization.memory_personality_summary)}
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
                </section>
            </div>
        </div>
    );
}

function formatConfiguredCount(count: number, label: string) {
    if (count === 0) {
        return "Not configured yet";
    }
    return `${count} ${label}${count === 1 ? "" : "s"}`;
}

function advisorSummary(count: number, teamLeadName: string) {
    if (count === 0) {
        return `${teamLeadName} is handling planning and review directly until Advisors are added.`;
    }
    if (count === 1) {
        return `1 Advisor is ready to help ${teamLeadName} with review, priorities, and decision support.`;
    }
    return `${count} Advisors are ready to help ${teamLeadName} review decisions and keep the organization aligned.`;
}

function advisorSupportItems(count: number) {
    if (count === 0) {
        return ["Inspect only for now", "Advisor roles appear here once they are added"];
    }
    return ["Planning review", "Decision support", "Priority checks"].slice(0, Math.max(2, Math.min(count + 1, 3)));
}

function departmentSummary(count: number, specialistCount: number, teamLeadName: string) {
    if (count === 0) {
        return `${teamLeadName} can still shape the first operating lane before Departments are configured.`;
    }
    return `${count} Departments and ${formatConfiguredCount(specialistCount, "Specialist").toLowerCase()} are visible here so ${teamLeadName} can work with a clear delivery structure.`;
}

function departmentSupportItems(organization: OrganizationHomePayload) {
    const items = [
        organization.start_mode === "template" && organization.template_name
            ? `Started from ${organization.template_name}`
            : "Started from Empty",
        formatConfiguredCount(organization.specialist_count, "Specialist"),
        organization.department_count > 0 ? "Inspect only for now" : "Add the first Department when ready",
    ];
    return items;
}

function advisorDetailItems(count: number) {
    if (count === 0) {
        return [];
    }

    return [
        {
            name: "Planning Advisor",
            purpose: "Helps the Team Lead test priorities, sequence work, and keep the first plan practical.",
            supportCue: "Best when the operator wants a second look at the next move.",
        },
        {
            name: "Delivery Advisor",
            purpose: "Keeps an eye on progress, blockers, and whether work is ready to move forward.",
            supportCue: "Useful when delivery momentum needs a quick review.",
        },
        {
            name: "Decision Advisor",
            purpose: "Supports tradeoffs, review points, and operator-facing choices before work expands.",
            supportCue: "Useful when the Team Lead needs a clear go/no-go perspective.",
        },
    ].slice(0, Math.min(count, 3));
}

function departmentDetailItems(organization: OrganizationHomePayload) {
    if (organization.department_count === 0) {
        return [];
    }

    const names =
        organization.department_count === 1
            ? ["Core Delivery Department"]
            : ["Planning Department", "Delivery Department", "Operations Department", "Support Department"];

    return names.slice(0, organization.department_count).map((name, index) => ({
        name,
        purpose: departmentPurpose(name),
        specialists: spreadSpecialists(organization.specialist_count, organization.department_count, index),
    }));
}

function departmentPurpose(name: string) {
    if (name.includes("Planning")) {
        return "Shapes the first approach, breaks work into practical steps, and keeps priorities aligned.";
    }
    if (name.includes("Operations")) {
        return "Keeps follow-through organized, reduces friction, and supports steady execution.";
    }
    if (name.includes("Support")) {
        return "Handles supporting work that helps the main delivery lane stay clear and focused.";
    }
    return "Carries the main delivery lane so the Team Lead can move from planning into execution.";
}

function spreadSpecialists(total: number, departmentCount: number, index: number) {
    if (departmentCount <= 0 || total <= 0) {
        return 0;
    }
    const base = Math.floor(total / departmentCount);
    const remainder = total % departmentCount;
    return base + (index < remainder ? 1 : 0);
}

function aiEngineSummary(summary: string) {
    const normalized = summary.trim();
    if (!normalized || normalized === "Set up later in Advanced mode") {
        return "The current AI Engine Settings keep the organization on a simple starter profile until deeper tuning is needed.";
    }
    return `The current AI Engine Settings profile is ${normalized.toLowerCase()} and shapes how the organization responds, plans, and carries work forward.`;
}

function aiEngineSupportItems(summary: string) {
    if (summary.trim() === "Set up later in Advanced mode") {
        return ["Response style", "Planning depth", "Inspect only for now"];
    }
    return [summary, "Response style", "Planning depth"];
}

function memoryPersonalitySummary(summary: string) {
    const normalized = summary.trim();
    if (!normalized || normalized === "Set up later in Advanced mode") {
        return "Memory & Personality stay on a simple starter posture so the Team Lead keeps a consistent tone and working style.";
    }
    return `Memory & Personality are currently ${normalized.toLowerCase()}, which shapes how the Team Lead remembers context and presents guidance.`;
}

function memoryPersonalitySupportItems(summary: string) {
    if (summary.trim() === "Set up later in Advanced mode") {
        return ["Working tone", "Context continuity", "Inspect only for now"];
    }
    return [summary, "Working tone", "Context continuity"];
}

function toTitleCase(value: string) {
    return value
        .split(/[\s_-]+/)
        .filter(Boolean)
        .map((segment) => segment.charAt(0).toUpperCase() + segment.slice(1))
        .join(" ");
}

function Metric({ label, value }: { label: string; value: string }) {
    return (
        <div className="rounded-2xl border border-cortex-border bg-cortex-bg px-4 py-3">
            <p className="text-[10px] font-mono uppercase tracking-[0.18em] text-cortex-text-muted">{label}</p>
            <p className="mt-1 text-sm font-medium text-cortex-text-main">{value}</p>
        </div>
    );
}

function HelpPill({ label }: { label: string }) {
    return (
        <div className="inline-flex items-center gap-2 rounded-full border border-cortex-border bg-cortex-bg px-3 py-2 text-sm text-cortex-text-main">
            <Sparkles className="h-4 w-4 text-cortex-primary" />
            <span>{label}</span>
        </div>
    );
}

function ActionPill({
    label,
    isActive,
    onClick,
}: {
    label: string;
    isActive?: boolean;
    onClick: () => void;
}) {
    return (
        <button
            type="button"
            onClick={onClick}
            className={`inline-flex items-center gap-2 rounded-full border px-3 py-2 text-sm transition-colors ${
                isActive
                    ? "border-cortex-primary/40 bg-cortex-primary/10 text-cortex-text-main"
                    : "border-cortex-border bg-cortex-bg text-cortex-text-main hover:border-cortex-primary/20"
            }`}
        >
            <Sparkles className="h-4 w-4 text-cortex-primary" />
            <span>{label}</span>
        </button>
    );
}

function InspectOnlySummary({
    icon,
    title,
    countLabel,
    summary,
    supportLabel,
    items,
    inspectActionLabel,
    onInspect,
}: {
    icon: React.ReactNode;
    title: string;
    countLabel: string;
    summary: string;
    supportLabel: string;
    items: string[];
    inspectActionLabel?: string;
    onInspect?: () => void;
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
                    <p className="mt-1">Inspect only</p>
                </div>
            </div>
            <div className="mt-5">
                <p className="text-sm font-medium text-cortex-text-main">{supportLabel}</p>
                <div className="mt-3 flex flex-wrap gap-2">
                    {items.map((item) => (
                        <div key={item} className="rounded-full border border-cortex-border bg-cortex-bg px-3 py-2 text-sm text-cortex-text-main">
                            {item}
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

function WorkspaceDetailView({
    view,
    organization,
    teamLeadName,
    onBack,
}: {
    view: "advisors" | "departments";
    organization: OrganizationHomePayload;
    teamLeadName: string;
    onBack: () => void;
}) {
    const title = view === "advisors" ? "Advisor details" : "Department details";
    const summary =
        view === "advisors"
            ? `${teamLeadName} can review the current Advisor support in ${organization.name} here without leaving the workspace.`
            : `${teamLeadName} can inspect the current Department structure in ${organization.name} here without leaving the workspace.`;
    const items =
        view === "advisors"
            ? advisorDetailItems(organization.advisor_count).map((item) => ({
                  key: item.name,
                  title: item.name,
                  detail: item.purpose,
                  support: item.supportCue,
              }))
            : departmentDetailItems(organization).map((item) => ({
                  key: item.name,
                  title: item.name,
                  detail: item.purpose,
                  support:
                      item.specialists > 0
                          ? `${item.specialists} ${item.specialists === 1 ? "Specialist" : "Specialists"} visible here.`
                          : "No Specialists assigned yet.",
              }));

    return (
        <div className="rounded-3xl border border-cortex-border bg-cortex-surface p-6">
            <div className="flex flex-col gap-4 lg:flex-row lg:items-start lg:justify-between">
                <div>
                    <h3 className="text-xl font-semibold text-cortex-text-main">{title}</h3>
                    <p className="mt-2 text-sm leading-7 text-cortex-text-muted">{summary}</p>
                </div>
                <button
                    type="button"
                    onClick={onBack}
                    className="inline-flex items-center gap-2 rounded-xl border border-cortex-border bg-cortex-bg px-3 py-2 text-sm font-medium text-cortex-text-main transition-colors hover:border-cortex-primary/20"
                >
                    <ArrowLeft className="h-4 w-4" />
                    Back to Team Lead
                </button>
            </div>

            {items.length > 0 ? (
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
                    {view === "advisors"
                        ? `${teamLeadName} is handling review directly until Advisors are added.`
                        : `${teamLeadName} can still shape the first operating lane before Departments are added.`}
                </div>
            )}
        </div>
    );
}
