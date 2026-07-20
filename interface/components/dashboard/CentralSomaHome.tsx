"use client";

import { use, useEffect, useMemo } from "react";
import { useLastOrganization } from "@/lib/lastOrganization";
import { SomaOperatingSurface } from "@/components/soma/SomaOperatingSurface";
import { useCortexStore } from "@/store/useCortexStore";

type LastOrganization = {
    id: string;
    name: string;
};

export function resolveDashboardRequestedTeamId(
    requestedTeamId: string,
    teams: Array<{ id: string }>,
) {
    if (!requestedTeamId) {
        return null;
    }
    return teams.some((team) => team.id === requestedTeamId) ? requestedTeamId : null;
}

export default function CentralSomaHome({
    requestedTeamIdPromise,
}: {
    requestedTeamIdPromise?: Promise<{ team_id?: string | string[] }>;
}) {
    const lastOrganization: LastOrganization | null = useLastOrganization();
    const fetchTeamsDetail = useCortexStore((s) => s.fetchTeamsDetail);
    const selectTeam = useCortexStore((s) => s.selectTeam);
    const selectedTeamId = useCortexStore((s) => s.selectedTeamId);
    const teamsDetail = useCortexStore((s) => s.teamsDetail);
    const resolvedSearchParams = requestedTeamIdPromise ? use(requestedTeamIdPromise) : undefined;
    const requestedTeamIdValue = resolvedSearchParams?.team_id;
    const requestedTeamId = Array.isArray(requestedTeamIdValue)
        ? requestedTeamIdValue[0]?.trim() ?? ""
        : requestedTeamIdValue?.trim() ?? "";

    useEffect(() => {
        void fetchTeamsDetail().finally(() => {
            const resolvedTeamId = resolveDashboardRequestedTeamId(
                requestedTeamId,
                useCortexStore.getState().teamsDetail,
            );
            selectTeam(resolvedTeamId);
        });
    }, [fetchTeamsDetail, requestedTeamId, selectTeam]);

    const focusedTeam = useMemo(
        () => (selectedTeamId ? teamsDetail.find((team) => team.id === selectedTeamId) ?? null : null),
        [selectedTeamId, teamsDetail],
    );

    return (
        <section className="h-full min-h-0">
            <SomaOperatingSurface
                organizationId={lastOrganization?.id}
                organizationName={lastOrganization?.name}
                activeMode={focusedTeam ? focusedTeam.name : null}
                focusedTeamId={focusedTeam?.id ?? null}
            />
        </section>
    );
}
