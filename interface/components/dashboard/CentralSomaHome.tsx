"use client";

import Link from "next/link";
import { use, useEffect, useMemo, useState } from "react";
import { Building2, FolderPlus, Layers3, Sparkles, Users } from "lucide-react";
import { readLastOrganization, subscribeLastOrganizationChange } from "@/lib/lastOrganization";
import MissionControlChat from "@/components/dashboard/MissionControlChat";
import CentralActivityStream from "@/components/dashboard/CentralActivityStream";
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
                        <QuickLink href="/docs?doc=v8-universal-soma-context-model" icon={<Layers3 className="h-4 w-4" />} label="Review Soma context model" />
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
                <div className="min-h-[680px] overflow-hidden rounded-3xl border border-cortex-border bg-cortex-bg">
                    <MissionControlChat simpleMode autoFocus />
                </div>
            </div>

            <div className="grid gap-4 xl:grid-cols-[1.2fr_0.8fr]">
                <CentralActivityStream />

                <div className="rounded-3xl border border-cortex-border bg-cortex-surface p-5">
                    <p className="text-[11px] font-mono uppercase tracking-[0.2em] text-cortex-primary">What this surface is for</p>
                    <div className="mt-3 space-y-3">
                        <CompactNote
                            title="Talk to Soma first"
                            detail="The dashboard should stay centered on direct Soma interaction, not a dense admin control board."
                        />
                        <CompactNote
                            title="Create teams through Soma"
                            detail="Use Soma to shape teams and recurring workflows, then move into a lead workspace only when a narrower lane is useful."
                        />
                        <CompactNote
                            title="Review outputs here"
                            detail="Soma should pull forward summaries, artifacts, download links, and team status without making the owner hunt through multiple pages first."
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

function CompactNote({
    title,
    detail,
}: {
    title: string;
    detail: string;
}) {
    return (
        <div className="rounded-2xl border border-cortex-border bg-cortex-bg px-4 py-3">
            <p className="text-sm font-semibold text-cortex-text-main">{title}</p>
            <p className="mt-2 text-sm leading-6 text-cortex-text-muted">{detail}</p>
        </div>
    );
}
