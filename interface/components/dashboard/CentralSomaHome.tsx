"use client";

import Link from "next/link";
import { use, useEffect, useMemo, useState } from "react";
import { Building2, FolderPlus, MessageCircle, ShieldCheck, Sparkles } from "lucide-react";
import { readLastOrganization, subscribeLastOrganizationChange } from "@/lib/lastOrganization";
import CentralActivityStream from "@/components/dashboard/CentralActivityStream";
import { SomaOperatingSurface } from "@/components/soma/SomaOperatingSurface";
import { useCortexStore } from "@/store/useCortexStore";

type LastOrganization = {
    id: string;
    name: string;
};

type WebSessionUser = {
    email?: string;
    name?: string;
    role?: "admin" | "standard";
    provider?: "local" | "google";
    hd?: string;
};

export default function CentralSomaHome({
    requestedTeamIdPromise,
}: {
    requestedTeamIdPromise?: Promise<{ team_id?: string | string[] }>;
}) {
    const [lastOrganization, setLastOrganization] = useState<LastOrganization | null>(null);
    const [sessionUser, setSessionUser] = useState<WebSessionUser | null>(null);
    const fetchTeamsDetail = useCortexStore((s) => s.fetchTeamsDetail);
    const selectTeam = useCortexStore((s) => s.selectTeam);
    const selectedTeamId = useCortexStore((s) => s.selectedTeamId);
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
        let cancelled = false;
        const request = fetch("/auth/session", { cache: "no-store" });
        if (!request?.then) {
            return () => {
                cancelled = true;
            };
        }
        request
            .then((res) => (res.ok ? res.json() : null))
            .then((body) => {
                const user = body?.data?.user;
                if (!cancelled && user) {
                    setSessionUser({
                        email: typeof user.email === "string" ? user.email : undefined,
                        name: typeof user.name === "string" ? user.name : undefined,
                        role: user.role === "standard" ? "standard" : "admin",
                        provider: user.provider === "google" ? "google" : "local",
                        hd: typeof user.hd === "string" ? user.hd : undefined,
                    });
                }
            })
            .catch(() => undefined);
        return () => {
            cancelled = true;
        };
    }, []);

    useEffect(() => {
        void fetchTeamsDetail().finally(() => {
            if (requestedTeamId) {
                selectTeam(requestedTeamId);
            }
        });
    }, [fetchTeamsDetail, requestedTeamId, selectTeam]);

    const focusedTeam = useMemo(
        () => (selectedTeamId ? teamsDetail.find((team) => team.id === selectedTeamId) ?? null : null),
        [selectedTeamId, teamsDetail],
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
        <section className="space-y-3">
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
                focusedTeamId={selectedTeamId}
            />
            <EnvironmentEntryBar sessionUser={sessionUser} />

            <div className="rounded-2xl border border-cortex-border bg-cortex-surface px-3 py-2">
                <div className="flex flex-col gap-2 lg:flex-row lg:items-center lg:justify-between">
                    <p className="min-w-0 text-xs leading-5 text-cortex-text-muted">
                        <span className="mr-2 inline-flex items-center gap-1 font-mono uppercase tracking-[0.16em] text-cortex-primary">
                            <MessageCircle className="h-3.5 w-3.5" />
                            Start here
                        </span>
                        Ask Soma to plan, review, create, make changes, or follow up. Set up an AI Organization when you want durable workspace context.
                    </p>
                    <div className="flex shrink-0 flex-wrap gap-2">
                        {lastOrganization ? (
                            <QuickLink href={`/organizations/${lastOrganization.id}`} icon={<Building2 className="h-4 w-4" />} label={`Return to ${lastOrganization.name}`} />
                        ) : null}
                        <QuickAction onClick={openOrganizationSetup} icon={<FolderPlus className="h-4 w-4" />} label="Set up an AI Organization" />
                    </div>
                </div>
            </div>

            <details className="rounded-2xl border border-cortex-border bg-cortex-surface px-3 py-2">
                <summary className="cursor-pointer font-mono text-[11px] uppercase tracking-[0.16em] text-cortex-text-muted">
                    Recent team activity
                </summary>
                <div className="mt-3">
                    <CentralActivityStream />
                </div>
            </details>
        </section>
    );
}

function EnvironmentEntryBar({ sessionUser }: { sessionUser: WebSessionUser | null }) {
    const roleLabel = sessionUser?.role === "standard" ? "Standard user" : "Admin";
    const providerLabel = sessionUser?.provider === "google" ? "Google Workspace" : "Local owner";
    const identityLabel = sessionUser?.email || sessionUser?.name || "Signed-in operator";
    const workspaceScope = sessionUser?.hd ? sessionUser.hd : providerLabel;

    return (
        <section
            data-testid="soma-environment-entry"
            className="rounded-2xl border border-cortex-border bg-cortex-surface px-3 py-2"
            aria-label="Signed-in Soma environment"
        >
            <div className="flex flex-col gap-2 lg:flex-row lg:items-center lg:justify-between">
                <div className="min-w-0">
                    <div className="flex flex-wrap items-center gap-2">
                        <span className="inline-flex items-center gap-2 rounded-full border border-cortex-primary/25 bg-cortex-primary/10 px-2.5 py-1 text-[10px] font-mono uppercase tracking-[0.16em] text-cortex-primary">
                            <Sparkles className="h-3.5 w-3.5" />
                            Soma environment
                        </span>
                        <span className="rounded-full border border-cortex-success/25 bg-cortex-success/10 px-2.5 py-1 text-[10px] font-mono uppercase tracking-[0.14em] text-cortex-success">
                            Signed in
                        </span>
                        <span className="truncate text-sm font-medium text-cortex-text-main">
                            {identityLabel}
                        </span>
                    </div>
                </div>
                <div className="flex flex-wrap gap-2 text-xs text-cortex-text-muted">
                    <EntryFact label="Access" value={roleLabel} />
                    <EntryFact label="Identity" value={providerLabel} />
                    <EntryFact label="Scope" value={workspaceScope} />
                </div>
            </div>
        </section>
    );
}

function EntryFact({ label, value }: { label: string; value: string }) {
    return (
        <div className="inline-flex max-w-full items-center gap-2 rounded-full border border-cortex-border bg-cortex-bg px-2.5 py-1.5">
            <p className="flex items-center gap-1.5 text-[10px] font-mono uppercase tracking-[0.14em] text-cortex-text-muted">
                <ShieldCheck className="h-3.5 w-3.5 text-cortex-primary" />
                {label}
            </p>
            <p className="truncate font-medium text-cortex-text-main">{value}</p>
        </div>
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
