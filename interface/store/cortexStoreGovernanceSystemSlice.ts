import { extractApiData } from '@/lib/apiContracts';
import type { CortexState } from '@/store/cortexStoreState';
import type {
    AuditLogEntry,
    PolicyConfig,
    ServiceHealthStatus,
    TeamAgent,
    TeamDetail,
} from '@/store/cortexStoreTypes';
import type { CortexGet, CortexSet, CortexSlice } from '@/store/cortexStoreSliceTypes';

export function createCortexGovernanceSystemSlice(
    set: CortexSet,
    _get: CortexGet,
): CortexSlice<
    | 'fetchTeamDetails'
    | 'fetchPolicy'
    | 'updatePolicy'
    | 'fetchPendingApprovals'
    | 'fetchAuditLog'
    | 'resolveApproval'
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

                const teams = teamsRes.ok ? await teamsRes.json() : [];
                const agentsData = agentsRes.ok ? await agentsRes.json() : { agents: [] };
                const agents: TeamAgent[] = Array.isArray(agentsData.agents) ? agentsData.agents : [];

                const roster: TeamDetail[] = (Array.isArray(teams) ? teams : []).map((team: any) => ({
                    id: team.id,
                    name: team.name,
                    role: team.role || 'observer',
                    agents: agents.filter((agent) => agent.team_id === team.id),
                }));

                set({ teamRoster: roster, isFetchingTeamRoster: false });
            } catch {
                set({ teamRoster: [], isFetchingTeamRoster: false });
            }
        },

        fetchPolicy: async () => {
            set({ isFetchingPolicy: true });
            try {
                const res = await fetch('/api/v1/governance/policy');
                if (res.ok) {
                    const data = await res.json();
                    set({ policyConfig: data, isFetchingPolicy: false });
                } else {
                    set({ isFetchingPolicy: false });
                }
            } catch {
                set({ isFetchingPolicy: false });
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
                    set({ policyConfig: config });
                }
            } catch (err) {
                console.error('[Governance] Update failed:', err);
            }
        },

        fetchPendingApprovals: async () => {
            set({ isFetchingApprovals: true });
            try {
                const res = await fetch('/api/v1/governance/pending');
                if (res.ok) {
                    const data = await res.json();
                    set({ pendingApprovals: Array.isArray(data) ? data : [], isFetchingApprovals: false });
                } else {
                    set({ pendingApprovals: [], isFetchingApprovals: false });
                }
            } catch {
                set({ pendingApprovals: [], isFetchingApprovals: false });
            }
        },

        fetchAuditLog: async () => {
            set({ isFetchingAuditLog: true });
            try {
                const res = await fetch('/api/v1/audit?limit=20');
                if (res.ok) {
                    const payload = await res.json();
                    const data = extractApiData<AuditLogEntry[] | unknown>(payload);
                    set({ auditLog: Array.isArray(data) ? data : [], isFetchingAuditLog: false });
                } else {
                    set({ auditLog: [], isFetchingAuditLog: false });
                }
            } catch {
                set({ auditLog: [], isFetchingAuditLog: false });
            }
        },

        resolveApproval: async (id: string, approved: boolean) => {
            try {
                const res = await fetch(`/api/v1/governance/resolve/${id}`, {
                    method: 'POST',
                    headers: { 'Content-Type': 'application/json' },
                    body: JSON.stringify({ action: approved ? 'APPROVE' : 'REJECT' }),
                });
                if (res.ok) {
                    set((s) => ({
                        pendingApprovals: s.pendingApprovals.filter((item) => item.id !== id),
                    }));
                }
            } catch (err) {
                console.error('[Governance] Resolve failed:', err);
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
