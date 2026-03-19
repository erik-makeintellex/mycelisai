"use client";

import Link from "next/link";
import { useEffect, useState } from "react";
import { ArrowLeft, ArrowRight, Blocks, Bot, BrainCircuit, Building2, Compass, Loader2, RefreshCcw, Sparkles, Users2 } from "lucide-react";
import { extractApiData, extractApiError } from "@/lib/apiContracts";
import type { OrganizationHomePayload } from "@/lib/organizations";

type WorkspaceView = "plan" | "advisors" | "departments" | "engine" | "memory";

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
    const [activeView, setActiveView] = useState<WorkspaceView>("plan");

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
    const workspaceContent = getWorkspaceContent(organization, activeView);
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
                                    <HelpPill icon={<Compass className="h-4 w-4" />} label="Plan the next steps" />
                                    <HelpPill icon={<Users2 className="h-4 w-4" />} label="Review Advisors and Specialists" />
                                    <HelpPill icon={<Building2 className="h-4 w-4" />} label="Open Departments and structure" />
                                </div>
                            </div>
                        </div>

                        <div className="rounded-3xl border border-cortex-border bg-cortex-surface p-6">
                            <div className="flex flex-col gap-3 lg:flex-row lg:items-end lg:justify-between">
                                <div>
                                    <h2 className="text-xl font-semibold text-cortex-text-main">Team Lead workspace</h2>
                                    <p className="mt-1 text-sm text-cortex-text-muted">
                                        Work with the Team Lead inside {organization.name}. Organization context stays visible while you focus the next action.
                                    </p>
                                </div>
                                <div className="rounded-2xl border border-cortex-border bg-cortex-bg px-4 py-3 text-sm text-cortex-text-muted">
                                    <p className="font-medium text-cortex-text-main">Current focus</p>
                                    <p className="mt-1">{workspaceContent.focusLabel}</p>
                                </div>
                            </div>

                            <div className="mt-5 grid gap-4 lg:grid-cols-[0.9fr_1.1fr]">
                                <div className="rounded-2xl border border-cortex-border bg-cortex-bg px-4 py-4">
                                    <p className="text-[10px] font-mono uppercase tracking-[0.18em] text-cortex-text-muted">Operator request</p>
                                    <p className="mt-2 text-base font-semibold text-cortex-text-main">{workspaceContent.requestTitle}</p>
                                    <p className="mt-2 text-sm leading-6 text-cortex-text-muted">{workspaceContent.requestDetail}</p>
                                </div>

                                <div className="rounded-2xl border border-cortex-border bg-cortex-bg px-4 py-4">
                                    <p className="text-[10px] font-mono uppercase tracking-[0.18em] text-cortex-text-muted">Team Lead guidance</p>
                                    <p className="mt-2 text-base font-semibold text-cortex-text-main">{workspaceContent.guidanceTitle}</p>
                                    <p className="mt-2 text-sm leading-6 text-cortex-text-muted">{workspaceContent.guidanceDetail}</p>
                                    <div className="mt-4 space-y-2">
                                        {workspaceContent.points.map((point) => (
                                            <GuidanceRow key={point}>{point}</GuidanceRow>
                                        ))}
                                    </div>
                                </div>
                            </div>
                        </div>
                    </div>

                    <div className="space-y-4">
                        <div className="rounded-3xl border border-cortex-border bg-cortex-surface p-6">
                            <div>
                                <h2 className="text-xl font-semibold text-cortex-text-main">Next actions</h2>
                                <p className="mt-1 text-sm text-cortex-text-muted">
                                    Choose what the Team Lead should focus on next. Advanced controls stay hidden until you intentionally open them later.
                                </p>
                            </div>
                            <div className="mt-4 space-y-3">
                                <ActionTile
                                    label="Review Advisors"
                                    detail="See whether advisors are ready to guide the organization."
                                    active={activeView === "advisors"}
                                    onClick={() => setActiveView("advisors")}
                                />
                                <ActionTile
                                    label="Open Departments"
                                    detail="Inspect how work is grouped across departments."
                                    active={activeView === "departments"}
                                    onClick={() => setActiveView("departments")}
                                />
                                <ActionTile
                                    label="Ask Team Lead to plan next steps"
                                    detail="Focus the Team Lead on the next coordinated moves for this AI Organization."
                                    active={activeView === "plan"}
                                    onClick={() => setActiveView("plan")}
                                />
                                <ActionTile
                                    label="Review AI Engine Settings"
                                    detail="Review the current AI Engine Settings summary without opening advanced controls."
                                    active={activeView === "engine"}
                                    onClick={() => setActiveView("engine")}
                                />
                                <ActionTile
                                    label="Review Memory & Personality"
                                    detail="Review the memory and personality summary in the organization frame."
                                    active={activeView === "memory"}
                                    onClick={() => setActiveView("memory")}
                                />
                            </div>
                        </div>

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

                        <div className="rounded-3xl border border-cortex-border bg-cortex-surface p-6">
                            <div className="grid gap-3">
                                <SummaryRow
                                    icon={<Bot className="h-4 w-4" />}
                                    label="AI Engine Settings"
                                    value={organization.ai_engine_settings_summary}
                                />
                                <SummaryRow
                                    icon={<BrainCircuit className="h-4 w-4" />}
                                    label="Memory & Personality"
                                    value={organization.memory_personality_summary}
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

function getWorkspaceContent(organization: OrganizationHomePayload, activeView: WorkspaceView) {
    switch (activeView) {
        case "advisors":
            return {
                focusLabel: "Review Advisors",
                requestTitle: "Review the advisor picture",
                requestDetail: `Look at how advisory guidance supports ${organization.name} and whether the Team Lead needs more decision support.`,
                guidanceTitle:
                    organization.advisor_count > 0
                        ? `${formatConfiguredCount(organization.advisor_count, "Advisor")} ready to guide`
                        : "Advisors not configured yet",
                guidanceDetail:
                    organization.advisor_count > 0
                        ? "The Team Lead can pull advisor perspective into planning, review, and escalation decisions."
                        : "The Team Lead can still move the organization forward, but advisor coverage should be added when you want more domain guidance.",
                points:
                    organization.advisor_count > 0
                        ? [
                              "Review which advisory perspectives matter most for the current organization purpose.",
                              "Use advisor guidance to shape priorities before work fans out to Departments and Specialists.",
                              "Keep advisor involvement visible so the operator understands where direction is coming from.",
                          ]
                        : [
                              "Decide whether this organization needs advisory coverage before the next planning cycle.",
                              "Keep the first delivery loop simple while the Team Lead coordinates structure directly.",
                              "Add advisors when the organization needs stronger review, strategy, or specialist guidance.",
                          ],
            };
        case "departments":
            return {
                focusLabel: "Open Departments",
                requestTitle: "Inspect department structure",
                requestDetail: `Open the department view for ${organization.name} so the Team Lead can show how work is grouped and where the next handoff should happen.`,
                guidanceTitle:
                    organization.department_count > 0
                        ? `${formatConfiguredCount(organization.department_count, "Department")} available`
                        : "Departments not configured yet",
                guidanceDetail:
                    organization.department_count > 0
                        ? "Departments organize how the Team Lead will route work across the AI Organization."
                        : "Start by defining the first Department when you want the Team Lead to separate execution areas.",
                points:
                    organization.department_count > 0
                        ? [
                              "Review whether the current department groups match the organization purpose.",
                              "Confirm where the next delivery step should live before assigning specialist work.",
                              "Keep the Team Lead as the routing layer so department changes stay coordinated.",
                          ]
                        : [
                              "Define the first Department around the clearest execution area for this organization.",
                              "Use the Team Lead to keep routing simple until departments are in place.",
                              "Add more Departments only when the organization needs clearer separation of work.",
                          ],
            };
        case "engine":
            return {
                focusLabel: "Review AI Engine Settings",
                requestTitle: "Review the AI Engine Settings summary",
                requestDetail: `Check the current AI Engine Settings posture for ${organization.name} without opening deeper controls.`,
                guidanceTitle: "AI Engine Settings stay summarized here",
                guidanceDetail: organization.ai_engine_settings_summary,
                points: [
                    "Use this summary to confirm that the Team Lead has the expected operating posture.",
                    "Keep deeper tuning hidden until you intentionally open advanced settings later.",
                    "Treat this review as a readiness check, not a detour away from the Team Lead workflow.",
                ],
            };
        case "memory":
            return {
                focusLabel: "Review Memory & Personality",
                requestTitle: "Review memory and personality posture",
                requestDetail: `Check how ${organization.name} should remember context, preserve continuity, and present itself to the operator.`,
                guidanceTitle: "Memory & Personality summary",
                guidanceDetail: organization.memory_personality_summary,
                points: [
                    "Confirm that the Team Lead tone fits the organization purpose and expected operator relationship.",
                    "Use the summary to understand continuity posture without exposing deeper configuration by default.",
                    "Keep identity and memory choices visible as part of the AI Organization, not as disconnected system settings.",
                ],
            };
        case "plan":
        default:
            return {
                focusLabel: "Ask Team Lead to plan next steps",
                requestTitle: `Plan the next steps for ${organization.name}`,
                requestDetail: `Ask the Team Lead to turn the organization purpose into the next coordinated actions without dropping into a generic assistant surface.`,
                guidanceTitle: "Recommended next steps",
                guidanceDetail: `${organization.team_lead_label} is ready to move ${organization.name} from setup into focused execution.`,
                points: buildNextSteps(organization),
            };
    }
}

function buildNextSteps(organization: OrganizationHomePayload) {
    const steps = [
        `Start with ${organization.purpose.toLowerCase()} and confirm the first outcome the organization should deliver.`,
        organization.department_count > 0
            ? `Open ${formatConfiguredCount(organization.department_count, "Department")} and confirm which area should take the first lead.`
            : "Define the first Department so the Team Lead has a clear place to route the first body of work.",
        organization.specialist_count > 0
            ? `Use ${formatConfiguredCount(organization.specialist_count, "Specialist")} as execution capacity once the Team Lead sets the plan.`
            : "Add Specialists after the Team Lead has clarified the first department-level plan.",
    ];

    if (organization.advisor_count === 0) {
        steps.push("Decide whether advisor coverage is needed before the next planning cycle.");
    } else {
        steps.push(`Pull in ${formatConfiguredCount(organization.advisor_count, "Advisor")} when the Team Lead needs review or strategic guidance.`);
    }

    return steps;
}

function formatConfiguredCount(count: number, label: string) {
    if (count === 0) {
        return "Not configured yet";
    }
    return `${count} ${label}${count === 1 ? "" : "s"}`;
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

function ActionTile({
    label,
    detail,
    active,
    onClick,
}: {
    label: string;
    detail: string;
    active: boolean;
    onClick: () => void;
}) {
    return (
        <button
            onClick={onClick}
            className={`w-full rounded-2xl border p-4 text-left transition-colors ${
                active ? "border-cortex-primary/40 bg-cortex-primary/10" : "border-cortex-border bg-cortex-bg hover:border-cortex-primary/20"
            }`}
        >
            <div className="flex items-start justify-between gap-3">
                <div>
                    <p className="text-sm font-semibold text-cortex-text-main">{label}</p>
                    <p className="mt-1 text-sm leading-6 text-cortex-text-muted">{detail}</p>
                </div>
                <ArrowRight className={`mt-0.5 h-4 w-4 ${active ? "text-cortex-primary" : "text-cortex-text-muted"}`} />
            </div>
        </button>
    );
}

function HelpPill({ icon, label }: { icon: React.ReactNode; label: string }) {
    return (
        <div className="inline-flex items-center gap-2 rounded-full border border-cortex-border bg-cortex-bg px-3 py-2 text-sm text-cortex-text-main">
            <span className="text-cortex-primary">{icon}</span>
            <span>{label}</span>
        </div>
    );
}

function GuidanceRow({ children }: { children: React.ReactNode }) {
    return (
        <div className="flex items-start gap-3 rounded-2xl border border-cortex-border bg-cortex-surface/60 px-3 py-3 text-sm text-cortex-text-muted">
            <span className="mt-1 h-2 w-2 rounded-full bg-cortex-primary" />
            <span className="leading-6">{children}</span>
        </div>
    );
}

function SummaryRow({
    icon,
    label,
    value,
}: {
    icon: React.ReactNode;
    label: string;
    value: string;
}) {
    return (
        <div className="rounded-2xl border border-cortex-border bg-cortex-bg px-4 py-4">
            <div className="flex items-center gap-2 text-cortex-primary">
                {icon}
                <p className="text-sm font-medium text-cortex-text-main">{label}</p>
            </div>
            <p className="mt-2 text-sm leading-6 text-cortex-text-muted">{value}</p>
        </div>
    );
}
