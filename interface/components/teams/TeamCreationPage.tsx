"use client";

import { useEffect, useState } from "react";
import Link from "next/link";
import { useSearchParams } from "next/navigation";
import { ArrowRight, Bot, Building2, Loader2, Sparkles, Users } from "lucide-react";
import TeamLeadInteractionPanel, { type TeamLeadPromptSuggestion } from "@/components/organizations/TeamLeadInteractionPanel";
import { extractApiData, extractApiError } from "@/lib/apiContracts";
import { readLastOrganization, subscribeLastOrganizationChange, type LastOrganizationRef } from "@/lib/lastOrganization";
import type { OrganizationHomePayload } from "@/lib/organizations";

type LoadState = "loading" | "ready" | "error";

const PROMPT_SUGGESTIONS: TeamLeadPromptSuggestion[] = [
    { label: "Marketing launch team", prompt: "Create a temporary marketing launch team that can produce campaign copy, a landing page brief, and social rollout assets for a new product launch." },
    { label: "Customer research team", prompt: "Create a research team that can analyze customer interviews, summarize patterns, and produce an executive-ready recommendation brief." },
    { label: "RevOps workflow team", prompt: "Create an operations team that can design a lead intake workflow, define the automation contract, and return the implementation checklist." },
    { label: "Security review team", prompt: "Create a security review team that can assess a planned release, identify risks, and deliver an approval-ready review summary." },
];

const WORKFLOW_HIGHLIGHTS = [
    "Start from the organization context Soma already knows instead of building a team from scratch.",
    "Ask for the team outcome, expected outputs, and delivery style before member details.",
    "Keep the first team compact by default; if the ask is broad, let Soma split it into several small teams or lane bundles instead of one giant roster.",
];

export default function TeamCreationPage() {
    const searchParams = useSearchParams();
    const [lastOrganization, setLastOrganization] = useState<LastOrganizationRef | null>(null);
    const [organization, setOrganization] = useState<OrganizationHomePayload | null>(null);
    const [loadState, setLoadState] = useState<LoadState>("loading");
    const [error, setError] = useState<string | null>(null);

    useEffect(() => {
        setLastOrganization(readLastOrganization());
        return subscribeLastOrganizationChange((next) => setLastOrganization(next));
    }, []);

    const effectiveOrganizationId = searchParams?.get("organization_id") || lastOrganization?.id || null;

    useEffect(() => {
        if (!effectiveOrganizationId) {
            setOrganization(null);
            setLoadState("ready");
            setError(null);
            return;
        }

        let cancelled = false;
        const loadOrganization = async () => {
            setLoadState("loading");
            setError(null);
            try {
                const response = await fetch(`/api/v1/organizations/${effectiveOrganizationId}/home`, { cache: "no-store" });
                const payload = await response.json().catch(() => null);
                if (!response.ok) {
                    throw new Error(extractApiError(payload) || "Unable to load the current AI Organization.");
                }
                if (cancelled) return;
                setOrganization(extractApiData<OrganizationHomePayload>(payload));
                setLoadState("ready");
            } catch (err) {
                if (cancelled) return;
                setOrganization(null);
                setLoadState("error");
                setError(err instanceof Error ? err.message : "Unable to load the current AI Organization.");
            }
        };

        void loadOrganization();
        return () => {
            cancelled = true;
        };
    }, [effectiveOrganizationId]);

    return (
        <div className="h-full overflow-y-auto bg-cortex-bg px-6 py-6">
            <div className="mx-auto flex max-w-7xl flex-col gap-6">
                <section className="rounded-3xl border border-cortex-border bg-cortex-surface px-6 py-6 shadow-[0_18px_40px_rgba(148,163,184,0.16)]">
                    <div className="flex flex-col gap-5 lg:flex-row lg:items-start lg:justify-between">
                        <div className="max-w-3xl space-y-4">
                            <div className="inline-flex items-center gap-2 rounded-full border border-cortex-primary/25 bg-cortex-primary/10 px-3 py-1 text-[11px] font-mono uppercase tracking-[0.24em] text-cortex-primary">
                                <Sparkles className="h-3.5 w-3.5" />
                                Guided Team Creation
                            </div>
                            <div className="space-y-3">
                                <h1 className="text-4xl font-semibold tracking-tight text-cortex-text-main">
                                    Create a team through Soma
                                </h1>
                                <p className="max-w-2xl text-base leading-7 text-cortex-text-muted">
                                    Use a guided workflow when you want to create a new team or delivery lane. Soma will shape the team around the organization context, expected outputs, and the kind of work the team needs to deliver.
                                </p>
                            </div>
                            <div className="flex flex-wrap gap-3">
                                <Link
                                    href="/teams"
                                    className="inline-flex items-center gap-2 rounded-xl border border-cortex-border bg-cortex-bg px-4 py-2.5 text-sm font-semibold text-cortex-text-main transition-colors hover:border-cortex-primary/20"
                                >
                                    <Users className="h-4 w-4" />
                                    Back to Teams
                                </Link>
                                <Link
                                    href="/dashboard"
                                    className="inline-flex items-center gap-2 rounded-xl border border-cortex-primary/35 bg-cortex-primary px-4 py-2.5 text-sm font-semibold text-cortex-bg transition-colors hover:bg-cortex-primary/90"
                                >
                                    <Bot className="h-4 w-4" />
                                    Open Soma workspace
                                </Link>
                            </div>
                        </div>
                        <div className="rounded-2xl border border-cortex-border bg-cortex-bg px-4 py-4 text-sm text-cortex-text-muted lg:max-w-md">
                            <p className="font-medium text-cortex-text-main">What this workflow does</p>
                            <ul className="mt-3 space-y-2 leading-6">
                                {WORKFLOW_HIGHLIGHTS.map((item) => (
                                    <li key={item}>{item}</li>
                                ))}
                            </ul>
                            <div className="mt-4 rounded-2xl border border-cortex-border bg-cortex-surface px-4 py-3">
                                <p className="text-xs font-mono uppercase tracking-[0.18em] text-cortex-primary">Default team shape</p>
                                <p className="mt-2 leading-6 text-cortex-text-muted">
                                    Keep rosters small and focused first. When the request spans marketing, product, ops, or media in the same breath, Soma should create several compact teams or lane bundles and coordinate the handoffs with Council over NATS instead of inflating one team past a dozen members.
                                </p>
                            </div>
                        </div>
                    </div>
                </section>

                <div className="grid gap-6 xl:grid-cols-[0.95fr_1.05fr]">
                    <section className="space-y-4 rounded-3xl border border-cortex-border bg-cortex-surface p-6">
                        <div>
                            <p className="text-[11px] font-mono uppercase tracking-[0.2em] text-cortex-primary">Step 1</p>
                            <h2 className="mt-2 text-xl font-semibold text-cortex-text-main">Confirm the organization context</h2>
                            <p className="mt-1 text-sm leading-6 text-cortex-text-muted">
                                Team creation is anchored to an AI Organization so Soma can use the right memory, priorities, and team defaults.
                            </p>
                        </div>

                        {loadState === "loading" ? (
                            <div className="flex items-center gap-3 rounded-2xl border border-cortex-border bg-cortex-bg px-4 py-4 text-sm text-cortex-text-muted">
                                <Loader2 className="h-4 w-4 animate-spin" />
                                Loading the current AI Organization...
                            </div>
                        ) : null}

                        {loadState === "error" && error ? (
                            <div className="rounded-2xl border border-cortex-danger/30 bg-cortex-danger/10 px-4 py-4">
                                <p className="text-sm font-semibold text-cortex-text-main">The current AI Organization could not be loaded</p>
                                <p className="mt-2 text-sm leading-6 text-cortex-text-muted">{error}</p>
                                <div className="mt-4 flex flex-wrap gap-2">
                                    <Link
                                        href="/dashboard"
                                        className="inline-flex items-center gap-2 rounded-xl border border-cortex-primary/35 bg-cortex-primary px-4 py-2.5 text-sm font-semibold text-cortex-bg transition-colors hover:bg-cortex-primary/90"
                                    >
                                        Return to Soma
                                    </Link>
                                    <Link
                                        href="/teams"
                                        className="inline-flex items-center gap-2 rounded-xl border border-cortex-border bg-cortex-bg px-4 py-2.5 text-sm font-semibold text-cortex-text-main transition-colors hover:border-cortex-primary/20"
                                    >
                                        Review current teams
                                    </Link>
                                </div>
                            </div>
                        ) : null}

                        {loadState === "ready" && !organization ? (
                            <div className="rounded-2xl border border-dashed border-cortex-border bg-cortex-bg px-4 py-4">
                                <p className="text-sm font-semibold text-cortex-text-main">Choose or create an AI Organization first</p>
                                <p className="mt-2 text-sm leading-6 text-cortex-text-muted">
                                    Open Soma from the dashboard, enter an organization workspace, and then come back here to create a team with the right organization context already attached.
                                </p>
                                <div className="mt-4 flex flex-wrap gap-2">
                                    <Link
                                        href="/dashboard"
                                        className="inline-flex items-center gap-2 rounded-xl border border-cortex-primary/35 bg-cortex-primary px-4 py-2.5 text-sm font-semibold text-cortex-bg transition-colors hover:bg-cortex-primary/90"
                                    >
                                        Open Soma workspace
                                    </Link>
                                </div>
                            </div>
                        ) : null}

                        {organization ? (
                            <div className="rounded-2xl border border-cortex-border bg-cortex-bg p-4">
                                <div className="flex items-start justify-between gap-3">
                                    <div>
                                        <p className="text-[11px] font-mono uppercase tracking-[0.2em] text-cortex-primary">Current organization</p>
                                        <p className="mt-2 text-lg font-semibold text-cortex-text-main">{organization.name}</p>
                                        <p className="mt-1 text-sm leading-6 text-cortex-text-muted">
                                            {organization.purpose}
                                        </p>
                                    </div>
                                    <Building2 className="mt-1 h-5 w-5 text-cortex-primary" />
                                </div>
                                <div className="mt-4 grid gap-3 md:grid-cols-3">
                                    <ContextMetric label="Team lead" value={organization.team_lead_label} />
                                    <ContextMetric label="Departments" value={String(organization.department_count)} />
                                    <ContextMetric label="Specialists" value={String(organization.specialist_count)} />
                                </div>
                                <div className="mt-4 flex flex-wrap gap-2">
                                    <Link
                                        href={`/organizations/${organization.id}`}
                                        className="inline-flex items-center gap-2 rounded-xl border border-cortex-border bg-cortex-surface px-4 py-2.5 text-sm font-semibold text-cortex-text-main transition-colors hover:border-cortex-primary/20"
                                    >
                                        Review organization
                                    </Link>
                                    <Link
                                        href="/teams"
                                        className="inline-flex items-center gap-2 rounded-xl border border-cortex-border bg-cortex-surface px-4 py-2.5 text-sm font-semibold text-cortex-text-main transition-colors hover:border-cortex-primary/20"
                                    >
                                        Review existing teams
                                    </Link>
                                </div>
                            </div>
                        ) : null}

                        <div className="rounded-2xl border border-cortex-border bg-cortex-bg p-4">
                            <p className="text-[11px] font-mono uppercase tracking-[0.2em] text-cortex-primary">Step 2</p>
                            <h3 className="mt-2 text-base font-semibold text-cortex-text-main">Tell Soma what success looks like</h3>
                            <div className="mt-3 space-y-2 text-sm leading-6 text-cortex-text-muted">
                                <p>Start with the outcome you need, not a list of members.</p>
                                <p>Name the deliverables the team should return, such as review notes, campaign assets, workflow contracts, or approval-ready recommendations.</p>
                                <p>Use the prompt chips in the creation panel if you want a quick enterprise-oriented starting point, then let Soma keep the default team compact unless the ask clearly needs several small coordinated teams.</p>
                            </div>
                        </div>
                    </section>

                    <section className="space-y-4 rounded-3xl border border-cortex-border bg-cortex-surface p-6">
                        <div>
                            <p className="text-[11px] font-mono uppercase tracking-[0.2em] text-cortex-primary">Step 3</p>
                            <h2 className="mt-2 text-xl font-semibold text-cortex-text-main">Run guided team design with Soma</h2>
                            <p className="mt-1 text-sm leading-6 text-cortex-text-muted">
                                This is the focused workflow for creating a team. Soma can plan the lane, decide whether the work should become a native team or external workflow contract, and return the outputs it expects the team to produce.
                            </p>
                        </div>

                        {organization ? (
                            <TeamLeadInteractionPanel
                                organizationId={organization.id}
                                organizationName={organization.name}
                                somaName={`Soma for ${organization.name}`}
                                teamLeadName={organization.team_lead_label}
                                autoFocusOnLoad
                                promptSuggestions={PROMPT_SUGGESTIONS}
                            />
                        ) : loadState === "ready" ? (
                            <div className="rounded-2xl border border-dashed border-cortex-border bg-cortex-bg px-4 py-6">
                                <p className="text-sm font-semibold text-cortex-text-main">The guided creation lane appears once an AI Organization is available</p>
                                <p className="mt-2 text-sm leading-6 text-cortex-text-muted">
                                    After you return to an organization workspace, this page will open directly into Soma&apos;s team-design flow.
                                </p>
                            </div>
                        ) : null}

                        <div className="rounded-2xl border border-cortex-border bg-cortex-bg p-4">
                            <p className="text-[11px] font-mono uppercase tracking-[0.2em] text-cortex-primary">What happens next</p>
                            <div className="mt-3 grid gap-3 md:grid-cols-3">
                                <NextStepCard
                                    title="Review the proposed lane"
                                    detail="Soma should return the team posture, first move, and expected outputs before deeper execution begins."
                                />
                                <NextStepCard
                                    title="Open the lead workspace"
                                    detail="Once the team exists, the Teams page becomes the place to work through that focused lead and inspect the roster."
                                />
                                <NextStepCard
                                    title="Keep templates clean"
                                    detail="Use Teams to tune reusable member templates after the workflow clarifies what kinds of specialists Soma should prefer."
                                />
                            </div>
                        </div>
                    </section>
                </div>
            </div>
        </div>
    );
}

function ContextMetric({ label, value }: { label: string; value: string }) {
    return (
        <div className="rounded-2xl border border-cortex-border bg-cortex-surface px-4 py-3">
            <p className="text-[10px] font-mono uppercase tracking-[0.18em] text-cortex-text-muted">{label}</p>
            <p className="mt-2 text-sm font-medium text-cortex-text-main">{value}</p>
        </div>
    );
}

function NextStepCard({ title, detail }: { title: string; detail: string }) {
    return (
        <div className="rounded-2xl border border-cortex-border bg-cortex-surface px-4 py-3">
            <div className="flex items-start justify-between gap-3">
                <div>
                    <p className="text-sm font-semibold text-cortex-text-main">{title}</p>
                    <p className="mt-2 text-sm leading-6 text-cortex-text-muted">{detail}</p>
                </div>
                <ArrowRight className="mt-1 h-4 w-4 text-cortex-text-muted" />
            </div>
        </div>
    );
}
