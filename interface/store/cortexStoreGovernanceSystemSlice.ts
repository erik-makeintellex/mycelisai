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

export function createCortexGovernanceSystemSlice(
    set: CortexSet,
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

                const teamRecords = (Array.isArray(teams) ? teams : []) as TeamRosterResponse[];
                const roster: TeamDetail[] = teamRecords.map((team) => ({
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

        fetchPendingApprovals: async () => {
            set({ isFetchingApprovals: true });
            try {
                const res = await fetch('/api/v1/governance/pending');
                if (res.ok) {
                    const data = await res.json();
                    set({ pendingApprovals: Array.isArray(data) ? data : [], approvalsError: null, isFetchingApprovals: false });
                } else {
                    // A 403/503 is not "zero pending requests": don't render
                    // "All Clear" for a blocker. Keep the list untouched and
                    // surface the error instead.
                    const body = await res.json().catch(() => ({}));
                    set({ approvalsError: { code: body?.data?.code, httpStatus: res.status }, isFetchingApprovals: false });
                }
            } catch {
                set({ approvalsError: { httpStatus: undefined }, isFetchingApprovals: false });
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
                        resolveApprovalError: null,
                    }));
                } else {
                    // Nothing changed: keep the card in the queue and tell
                    // the admin why, instead of the card just staying put
                    // with no explanation.
                    const body = await res.json().catch(() => ({}));
                    set({ resolveApprovalError: { code: body?.data?.code, httpStatus: res.status } });
                }
            } catch (err) {
                console.error('[Governance] Resolve failed:', err);
                set({ resolveApprovalError: { httpStatus: undefined } });
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
