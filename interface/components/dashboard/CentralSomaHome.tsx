"use client";

import { use, useEffect, useMemo, useState } from "react";
import { ShieldCheck, Sparkles } from "lucide-react";
import { readLastOrganization, subscribeLastOrganizationChange } from "@/lib/lastOrganization";
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

    return (
        <section className="space-y-3">
            {focusedTeam ? (
                <div className="rounded-2xl border border-cortex-primary/20 bg-cortex-primary/10 px-4 py-3 text-sm text-cortex-text-main">
                    <p className="text-[11px] font-mono uppercase tracking-[0.18em] text-cortex-primary">Focused team</p>
                    <p className="mt-2 font-semibold">{focusedTeam.name}</p>
                    <p className="mt-1 text-sm text-cortex-text-muted">
                        {assistantName} is focused on this team. Ask for updates, decisions, or the next file from here.
                    </p>
                </div>
            ) : null}
            <SomaOperatingSurface
                organizationName={lastOrganization?.name}
                activeMode={focusedTeam ? focusedTeam.name : null}
                focusedTeamId={selectedTeamId}
            />
            <EnvironmentEntryBar sessionUser={sessionUser} />
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
                    <EntryFact label="Role" value={roleLabel} />
                    <EntryFact label="Sign-in" value={providerLabel} />
                    <EntryFact label="Workspace" value={workspaceScope} />
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
