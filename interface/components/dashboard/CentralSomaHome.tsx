"use client";

import Link from "next/link";
import { useEffect, useMemo, useState } from "react";
import { useSearchParams } from "next/navigation";
import { ArrowRight, Bot, Building2, FolderPlus, Layers3, ShieldCheck, Sparkles, Users } from "lucide-react";
import { readLastOrganization, subscribeLastOrganizationChange } from "@/lib/lastOrganization";
import MissionControlChat from "@/components/dashboard/MissionControlChat";
import CentralActivityStream from "@/components/dashboard/CentralActivityStream";
import { useCortexStore } from "@/store/useCortexStore";

type LastOrganization = {
    id: string;
    name: string;
};

const CENTRAL_PROMPTS = [
    "Review all active teams and what they are producing",
    "Create or reshape teams from the root Soma workspace",
    "Create or review groups without leaving the main admin flow",
    "Pull forward outputs, summaries, and download links for binary artifacts",
];

export default function CentralSomaHome() {
    const searchParams = useSearchParams();
    const [lastOrganization, setLastOrganization] = useState<LastOrganization | null>(null);
    const fetchTeamsDetail = useCortexStore((s) => s.fetchTeamsDetail);
    const selectTeam = useCortexStore((s) => s.selectTeam);
    const teamsDetail = useCortexStore((s) => s.teamsDetail);
    const assistantName = useCortexStore((s) => s.assistantName);
    const requestedTeamId = searchParams?.get("team_id")?.trim() ?? "";

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

    return (
        <section className="space-y-5">
            <div className="grid gap-5 xl:grid-cols-[1.35fr_0.65fr]">
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
                            <p className="max-w-3xl text-sm leading-7 text-cortex-text-muted">
                                This is the root owner workspace for reviewing teams, shaping groups, pulling forward outputs, and asking Soma to coordinate the next move. Readable outputs should appear inline here, while saved or binary results should link back into downloadable objects.
                            </p>
                        </div>
                        {focusedTeam ? (
                            <div className="rounded-2xl border border-cortex-primary/20 bg-cortex-primary/10 px-4 py-3 text-sm text-cortex-text-main">
                                <p className="text-[11px] font-mono uppercase tracking-[0.18em] text-cortex-primary">Focused Team Context</p>
                                <p className="mt-2 font-semibold">{focusedTeam.name}</p>
                                <p className="mt-1 text-sm text-cortex-text-muted">
                                    {assistantName} is currently treating this team as the active lead lane while staying in the root admin workspace.
                                </p>
                            </div>
                        ) : null}
                    </div>
                    <div className="min-h-[560px] overflow-hidden rounded-3xl border border-cortex-border bg-cortex-bg">
                        <MissionControlChat simpleMode autoFocus />
                    </div>
                </div>

                <div className="space-y-4">
                    <div className="rounded-3xl border border-cortex-border bg-cortex-surface p-5">
                        <p className="text-[11px] font-mono uppercase tracking-[0.2em] text-cortex-primary">Admin actions</p>
                        <div className="mt-3 flex flex-col gap-3">
                            <ActionLink href="/groups" icon={<Users className="h-4 w-4" />} title="Open groups workspace" detail="Create, review, and broadcast collaboration groups in their own lane." />
                            {lastOrganization ? (
                                <ActionLink href={`/organizations/${lastOrganization.id}`} icon={<Building2 className="h-4 w-4" />} title={`Return to ${lastOrganization.name}`} detail="Re-enter the current AI Organization and continue scoped delivery work." />
                            ) : null}
                            <ActionLink href="#dashboard-organization-setup" icon={<FolderPlus className="h-4 w-4" />} title="Create or open AI Organizations" detail="Keep organization setup secondary until you intentionally open it." />
                            <ActionLink href="/docs?doc=v8-universal-soma-context-model" icon={<Layers3 className="h-4 w-4" />} title="Review the Soma context model" detail="Inspect the current product model for central Soma and governed contexts." />
                        </div>
                    </div>

                    <div className="rounded-3xl border border-cortex-border bg-cortex-surface p-5">
                        <p className="text-[11px] font-mono uppercase tracking-[0.2em] text-cortex-primary">How Soma should help</p>
                        <div className="mt-3 space-y-3">
                            {CENTRAL_PROMPTS.map((prompt) => (
                                <PrincipleRow
                                    key={prompt}
                                    icon={<Bot className="h-4 w-4" />}
                                    title={prompt}
                                    detail="Stay in the central Soma chat unless you intentionally need a narrower team or group lane."
                                />
                            ))}
                        </div>
                    </div>

                    <CentralActivityStream />

                    <div className="rounded-3xl border border-cortex-border bg-cortex-bg p-5">
                        <div className="flex items-center gap-2 text-cortex-primary">
                            <ShieldCheck className="h-4 w-4" />
                            <p className="text-sm font-semibold text-cortex-text-main">Output handling</p>
                        </div>
                        <p className="mt-3 text-sm leading-6 text-cortex-text-muted">
                            Non-binary outputs should render inline in the Soma conversation. Saved or binary outputs should stay visible through clickable saved paths or download links without forcing the user into a system page first.
                        </p>
                        <div className="mt-4 rounded-2xl border border-cortex-border bg-cortex-surface px-4 py-3 text-sm text-cortex-text-muted">
                            Team lanes stay focused around team leads. Groups live in their own workspace. Soma remains the root reviewer that can summarize status, pull outputs forward, and coordinate across all of them.
                        </div>
                    </div>
                </div>
            </div>
        </section>
    );
}

function PrincipleRow({
    icon,
    title,
    detail,
}: {
    icon: React.ReactNode;
    title: string;
    detail: string;
}) {
    return (
        <div className="rounded-2xl border border-cortex-border bg-cortex-surface px-4 py-3">
            <div className="flex items-center gap-2 text-cortex-primary">
                {icon}
                <p className="text-sm font-semibold text-cortex-text-main">{title}</p>
            </div>
            <p className="mt-2 text-sm leading-6 text-cortex-text-muted">{detail}</p>
        </div>
    );
}

function ActionLink({
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
            className="rounded-2xl border border-cortex-border bg-cortex-bg px-4 py-3 transition-colors hover:border-cortex-primary/25 hover:bg-cortex-surface"
        >
            <div className="flex items-center justify-between gap-3">
                <div className="flex min-w-0 items-start gap-2">
                    <div className="mt-0.5 text-cortex-primary">{icon}</div>
                    <div className="min-w-0">
                        <p className="text-sm font-semibold text-cortex-text-main">{title}</p>
                        <p className="mt-1 text-sm leading-6 text-cortex-text-muted">{detail}</p>
                    </div>
                </div>
                <ArrowRight className="h-4 w-4 flex-shrink-0 text-cortex-primary" />
            </div>
        </Link>
    );
}
