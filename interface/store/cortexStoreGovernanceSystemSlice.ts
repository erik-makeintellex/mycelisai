import { extractApiData } from '@/lib/apiContracts';
import type {
    AuditLogEntry,
    PolicyConfig,
    ServiceHealthStatus,
    TeamAgent,
    TeamDetail,
} from '@/store/cortexStoreTypes';
import type { CortexSet, CortexSlice } from '@/store/cortexStoreSliceTypes';

type TeamRosterResponse = Pick<TeamDetail, 'id' | 'name'> & { role?: string };

// Same {code, httpStatus} shape as policyError/auditLogError, so the UI can
// feed it to blockerCopy. Never invents a code: a body without one leaves it
// undefined and the generic "request failed" copy applies.
async function loadErrorFromResponse(res: Response) {
    const body = await res.json().catch(() => ({}));
    return { code: body?.data?.code as string | undefined, httpStatus: res.status };
}

export function createCortexGovernanceSystemSlice(
    set: CortexSet,
): CortexSlice<
    | 'fetchTeamDetails'
    | 'fetchPolicy'
    | 'updatePolicy'
    | 'fetchAuditLog'
    | 'fetchCognitiveStatus'
    | 'fetchServicesStatus'
> {
    return {
        fetchTeamDetails: async () => {
            set({ isFetchingTeamRoster: true });
            try {
                const [teamsRes, agentsRes] = await Promise.all([
                    fetch('/api/v1/teams'),
                    fetch('/agents'),
                ]);

                // A failed teams/agents request is a load failure, not an
                // empty roster: keep the last roster and record the error.
                const failed = !teamsRes.ok ? teamsRes : !agentsRes.ok ? agentsRes : null;
                if (failed) {
                    set({ teamRosterError: await loadErrorFromResponse(failed), isFetchingTeamRoster: false });
                    return;
                }
                const teams = await teamsRes.json();
                const agentsData = await agentsRes.json();
                const agents: TeamAgent[] = Array.isArray(agentsData.agents) ? agentsData.agents : [];

                const teamRecords = (Array.isArray(teams) ? teams : []) as TeamRosterResponse[];
                const roster: TeamDetail[] = teamRecords.map((team) => ({
                    id: team.id,
                    name: team.name,
                    role: team.role || 'observer',
                    agents: agents.filter((agent) => agent.team_id === team.id),
                }));

                set({ teamRoster: roster, teamRosterError: null, isFetchingTeamRoster: false });
            } catch {
                set({ teamRosterError: { httpStatus: undefined }, isFetchingTeamRoster: false });
            }
        },

        fetchPolicy: async () => {
            set({ isFetchingPolicy: true });
            try {
                const res = await fetch('/api/v1/governance/policy');
                if (res.ok) {
                    const data = await res.json();
                    set({ policyConfig: data, policyError: null, isFetchingPolicy: false });
                } else {
                    // Never collapse a non-OK response to a fake empty/DENY
                    // policy: keep policyConfig null and surface the code so
                    // the UI can show a real blocker instead of an editor
                    // that looks like the real (empty) policy.
                    const body = await res.json().catch(() => ({}));
                    set({
                        policyConfig: null,
                        policyError: { code: body?.data?.code, httpStatus: res.status },
                        isFetchingPolicy: false,
                    });
                }
            } catch {
                set({ policyConfig: null, policyError: { httpStatus: undefined }, isFetchingPolicy: false });
            }
        },

        updatePolicy: async (config: PolicyConfig) => {
            try {
                const res = await fetch('/api/v1/governance/policy', {
                    method: 'PUT',
                    headers: { 'Content-Type': 'application/json' },
                    body: JSON.stringify(config),
                });
                if (res.ok) {
                    set({ policyConfig: config, policyError: null });
                } else {
                    const body = await res.json().catch(() => ({}));
                    set({ policyError: { code: body?.data?.code, httpStatus: res.status } });
                }
            } catch (err) {
                console.error('[Governance] Update failed:', err);
                set({ policyError: { httpStatus: undefined } });
            }
        },

        fetchAuditLog: async () => {
            set({ isFetchingAuditLog: true });
            try {
                const res = await fetch('/api/v1/audit?limit=20');
                if (res.ok) {
                    const payload = await res.json();
                    const data = extractApiData<AuditLogEntry[] | unknown>(payload);
                    set({ auditLog: Array.isArray(data) ? data : [], auditLogError: null, isFetchingAuditLog: false });
                } else {
                    // A 403/5xx is not "no audit activity": keep the list
                    // untouched and surface the error so the UI shows a
                    // blocker instead of the empty state.
                    set({ auditLogError: await loadErrorFromResponse(res), isFetchingAuditLog: false });
                }
            } catch {
                set({ auditLogError: { httpStatus: undefined }, isFetchingAuditLog: false });
            }
        },

        fetchCognitiveStatus: async () => {
            try {
                const res = await fetch('/api/v1/cognitive/status');
                if (res.ok) {
                    const data = await res.json();
                    set({ cognitiveStatus: data });
                }
            } catch {
                // silently fail — dashboard gauge will show offline
            }
        },

        fetchServicesStatus: async () => {
            set({ isFetchingServicesStatus: true });
            try {
                const res = await fetch('/api/v1/services/status');
                if (!res.ok) {
                    set({ isFetchingServicesStatus: false });
                    return [];
                }
                const payload = await res.json();
                const data = extractApiData<ServiceHealthStatus[] | unknown>(payload);
                const next = Array.isArray(data) ? data : [];
                set({
                    servicesStatus: next,
                    isFetchingServicesStatus: false,
                    servicesStatusUpdatedAt: new Date().toISOString(),
                });
                return next;
            } catch {
                set({ isFetchingServicesStatus: false });
                return [];
            }
        },
    };
}
