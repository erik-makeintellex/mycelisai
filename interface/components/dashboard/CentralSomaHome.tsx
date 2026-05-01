"use client";

import Link from "next/link";
import { use, useEffect, useMemo, useState } from "react";
import { Activity, Building2, CheckSquare, FolderPlus, ListChecks, Sparkles, Users, Wrench } from "lucide-react";
import { readLastOrganization, subscribeLastOrganizationChange } from "@/lib/lastOrganization";
import MissionControlChat from "@/components/dashboard/MissionControlChat";
import CentralActivityStream from "@/components/dashboard/CentralActivityStream";
import { SomaCapabilityGuide } from "@/components/dashboard/SomaCapabilityGuide";
import { useCortexStore } from "@/store/useCortexStore";

type LastOrganization = {
    id: string;
    name: string;
};

export default function CentralSomaHome({
    requestedTeamIdPromise,
}: {
    requestedTeamIdPromise?: Promise<{ team_id?: string | string[] }>;
}) {
    const [lastOrganization, setLastOrganization] = useState<LastOrganization | null>(null);
    const fetchTeamsDetail = useCortexStore((s) => s.fetchTeamsDetail);
    const selectTeam = useCortexStore((s) => s.selectTeam);
    const teamsDetail = useCortexStore((s) => s.teamsDetail);
    const assistantName = useCortexStore((s) => s.assistantName);
    const resolvedSearchParams = requestedTeamIdPromise ? use(requestedTeamIdPromise) : undefined;
    const requestedTeamIdValue = resolvedSearchParams?.team_id;
    const requestedTeamId = Array.isArray(requestedTeamIdValue)
        ? requestedTeamIdValue[0]?.trim() ?? ""
        : requestedTeamIdValue?.trim() ?? "";

    useEffect(() => {
        setLastOrganization(readLastOrganization());
        return subscribeLastOrganizationChange((organization) => {
            setLastOrganization(organization);
        });
    }, []);

    useEffect(() => {
        if (!requestedTeamId) {
            return;
        }
        void fetchTeamsDetail().finally(() => {
            selectTeam(requestedTeamId);
        });
    }, [fetchTeamsDetail, requestedTeamId, selectTeam]);

    const focusedTeam = useMemo(
        () => (requestedTeamId ? teamsDetail.find((team) => team.id === requestedTeamId) ?? null : null),
        [requestedTeamId, teamsDetail],
    );

    const openOrganizationSetup = () => {
        if (typeof document === "undefined") {
            return;
        }

        const details = document.getElementById("dashboard-organization-setup");
        if (!(details instanceof HTMLDetailsElement)) {
            return;
        }

        details.open = true;
        if (typeof details.scrollIntoView === "function") {
            details.scrollIntoView({ behavior: "smooth", block: "start" });
        }
        if (typeof window !== "undefined") {
            window.history.replaceState(null, "", "#dashboard-organization-setup");
        }
    };

    return (
        <section className="space-y-5">
            <div className="rounded-3xl border border-cortex-border bg-cortex-surface p-5 shadow-[0_18px_40px_rgba(148,163,184,0.16)]">
                <div className="mb-4 flex flex-col gap-3 border-b border-cortex-border pb-4">
                    <div className="inline-flex w-fit items-center gap-2 rounded-full border border-cortex-primary/25 bg-cortex-primary/10 px-3 py-1 text-[11px] font-mono uppercase tracking-[0.24em] text-cortex-primary">
                        <Sparkles className="h-3.5 w-3.5" />
                        Central Soma
                    </div>
                    <div className="space-y-2">
                        <h1 className="text-3xl font-semibold tracking-tight text-cortex-text-main">
                            Work directly with {assistantName} from the admin home.
                        </h1>
                        <p className="max-w-4xl text-sm leading-7 text-cortex-text-muted">
                            The root workspace should feel like a direct conversation with Soma. Ask for planning, team creation, reviews, repeated workflows, and output delivery here first; readable results stay inline, and saved or binary outputs should stay visible as clickable files.
                        </p>
                    </div>
                    <div className="flex flex-wrap gap-2">
                        <QuickLink href="/groups" icon={<Users className="h-4 w-4" />} label="Open groups workspace" />
                        {lastOrganization ? (
                            <QuickLink href={`/organizations/${lastOrganization.id}`} icon={<Building2 className="h-4 w-4" />} label={`Return to ${lastOrganization.name}`} />
                        ) : null}
                        <QuickAction onClick={openOrganizationSetup} icon={<FolderPlus className="h-4 w-4" />} label="Create or open AI Organizations" />
                        <QuickLink href="/activity" icon={<Activity className="h-4 w-4" />} label="Review workflow activity" />
                    </div>
                    {focusedTeam ? (
                        <div className="rounded-2xl border border-cortex-primary/20 bg-cortex-primary/10 px-4 py-3 text-sm text-cortex-text-main">
                            <p className="text-[11px] font-mono uppercase tracking-[0.18em] text-cortex-primary">Focused Team Context</p>
                            <p className="mt-2 font-semibold">{focusedTeam.name}</p>
                            <p className="mt-1 text-sm text-cortex-text-muted">
                                {assistantName} is currently treating this team as the active lead lane while keeping the conversation in the central admin workspace.
                            </p>
                        </div>
                    ) : null}
                </div>
                <div className="mb-4">
                    <SomaCapabilityGuide />
                </div>
                <div
                    data-testid="central-soma-chat-frame"
                    className="h-[72vh] min-h-[560px] max-h-[760px] overflow-hidden rounded-3xl border border-cortex-border bg-cortex-bg"
                >
                    <MissionControlChat simpleMode autoFocus />
                </div>
            </div>

            <div className="grid gap-4 xl:grid-cols-[1.2fr_0.8fr]">
                <CentralActivityStream />

                <div className="rounded-3xl border border-cortex-border bg-cortex-surface p-5">
                    <div className="flex items-center justify-between gap-3">
                        <div>
                            <p className="text-[11px] font-mono uppercase tracking-[0.2em] text-cortex-primary">Management Workbench</p>
                            <h2 className="mt-2 text-base font-semibold text-cortex-text-main">Move from conversation to action</h2>
                        </div>
                        <Sparkles className="h-5 w-5 text-cortex-primary" />
                    </div>
                    <div className="mt-4 divide-y divide-cortex-border overflow-hidden rounded-2xl border border-cortex-border bg-cortex-bg">
                        <WorkbenchLink
                            href="/approvals"
                            icon={<CheckSquare className="h-4 w-4" />}
                            title="Approval queue"
                            detail="Review gated actions, risk decisions, and pending confirmations."
                        />
                        <WorkbenchLink
                            href="/activity"
                            icon={<ListChecks className="h-4 w-4" />}
                            title="Workflow activity"
                            detail="Review current runs and readable message-bus activity."
                        />
                        <WorkbenchLink
                            href="/groups"
                            icon={<Users className="h-4 w-4" />}
                            title="Groups and teams"
                            detail="Select a group, inspect its teams, and manage launch lanes."
                        />
                        <WorkbenchLink
                            href="/settings?tab=tools"
                            icon={<Wrench className="h-4 w-4" />}
                            title="Tool readiness"
                            detail="Check connected tools, search configuration, and MCP capability status."
                        />
                    </div>
                </div>
            </div>
        </section>
    );
}

function QuickLink({
    href,
    icon,
    label,
}: {
    href: string;
    icon: React.ReactNode;
    label: string;
}) {
    return (
        <Link
            href={href}
            className="inline-flex items-center gap-2 rounded-2xl border border-cortex-border bg-cortex-bg px-3 py-2 text-sm font-medium text-cortex-text-main transition-colors hover:border-cortex-primary/25 hover:bg-cortex-surface"
        >
            <span className="text-cortex-primary">{icon}</span>
            {label}
        </Link>
    );
}

function QuickAction({
    onClick,
    icon,
    label,
}: {
    onClick: () => void;
    icon: React.ReactNode;
    label: string;
}) {
    return (
        <button
            type="button"
            onClick={onClick}
            className="inline-flex items-center gap-2 rounded-2xl border border-cortex-border bg-cortex-bg px-3 py-2 text-sm font-medium text-cortex-text-main transition-colors hover:border-cortex-primary/25 hover:bg-cortex-surface"
        >
            <span className="text-cortex-primary">{icon}</span>
            {label}
        </button>
    );
}

function WorkbenchLink({
    href,
    icon,
    title,
    detail,
}: {
    href: string;
    icon: React.ReactNode;
    title: string;
    detail: string;
}) {
    return (
        <Link
            href={href}
            className="grid grid-cols-[auto_1fr] gap-3 px-4 py-3 transition-colors hover:bg-cortex-surface"
        >
            <span className="mt-0.5 text-cortex-primary">{icon}</span>
            <span>
                <span className="block text-sm font-semibold text-cortex-text-main">{title}</span>
                <span className="mt-1 block text-sm leading-6 text-cortex-text-muted">{detail}</span>
            </span>
        </Link>
    );
}
