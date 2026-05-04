"use client";

import Link from "next/link";
import { use, useEffect, useMemo, useState } from "react";
import { Building2, FolderPlus } from "lucide-react";
import { readLastOrganization, subscribeLastOrganizationChange } from "@/lib/lastOrganization";
import CentralActivityStream from "@/components/dashboard/CentralActivityStream";
import { SomaOperatingSurface } from "@/components/soma/SomaOperatingSurface";
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
            <div className="rounded-3xl border border-cortex-border bg-cortex-surface p-3">
                <div className="flex flex-wrap gap-2">
                        {lastOrganization ? (
                            <QuickLink href={`/organizations/${lastOrganization.id}`} icon={<Building2 className="h-4 w-4" />} label={`Return to ${lastOrganization.name}`} />
                        ) : null}
                        <QuickAction onClick={openOrganizationSetup} icon={<FolderPlus className="h-4 w-4" />} label="Create or open AI Organizations" />
                </div>
            </div>
            {focusedTeam ? (
                <div className="rounded-2xl border border-cortex-primary/20 bg-cortex-primary/10 px-4 py-3 text-sm text-cortex-text-main">
                    <p className="text-[11px] font-mono uppercase tracking-[0.18em] text-cortex-primary">Focused team context</p>
                    <p className="mt-2 font-semibold">{focusedTeam.name}</p>
                    <p className="mt-1 text-sm text-cortex-text-muted">
                        {assistantName} is treating this team as the active operating lane inside the Soma surface.
                    </p>
                </div>
            ) : null}
            <SomaOperatingSurface
                organizationName={lastOrganization?.name}
                activeMode={focusedTeam ? `Focused team: ${focusedTeam.name}` : null}
            />

            <CentralActivityStream />
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
