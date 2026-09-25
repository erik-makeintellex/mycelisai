import { extractApiData } from '@/lib/apiContracts';
import { classifyServicesStatusFailure } from '@/lib/coreReachability';
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
                    set({ policyConfig: data, isFetchingPolicy: false, policyLoadError: null });
                } else {
                    set({
                        isFetchingPolicy: false,
                        policyLoadError: `Could not load policy (HTTP ${res.status}).`,
                    });
                }
            } catch {
                set({
                    isFetchingPolicy: false,
                    policyLoadError: 'Could not load policy: the Core API did not respond.',
                });
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
                    if (!Array.isArray(data)) {
                        set({
                            isFetchingApprovals: false,
                            approvalsLoadError: 'Could not load approvals: unexpected response.',
                        });
                        return;
                    }
                    set({ pendingApprovals: data, isFetchingApprovals: false, approvalsLoadError: null });
                } else {
                    set({
                        isFetchingApprovals: false,
                        approvalsLoadError: `Could not load approvals (HTTP ${res.status}).`,
                    });
                }
            } catch {
                set({
                    isFetchingApprovals: false,
                    approvalsLoadError: 'Could not load approvals: the Core API did not respond.',
                });
            }
        },

        fetchAuditLog: async () => {
            set({ isFetchingAuditLog: true });
            try {
                const res = await fetch('/api/v1/audit?limit=20');
                if (res.ok) {
                    const payload = await res.json();
                    const data = extractApiData<AuditLogEntry[] | unknown>(payload);
                    if (!Array.isArray(data)) {
                        set({
                            isFetchingAuditLog: false,
                            auditLogLoadError: 'Could not load the audit log: unexpected response.',
                        });
                        return;
                    }
                    set({ auditLog: data, isFetchingAuditLog: false, auditLogLoadError: null });
                } else {
                    set({
                        isFetchingAuditLog: false,
                        auditLogLoadError: `Could not load the audit log (HTTP ${res.status}).`,
                    });
                }
            } catch {
                set({
                    isFetchingAuditLog: false,
                    auditLogLoadError: 'Could not load the audit log: the Core API did not respond.',
                });
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
                    set({
                        isFetchingServicesStatus: false,
                        servicesStatusFailure: classifyServicesStatusFailure(res.status),
                    });
                    return [];
                }
                const payload = await res.json();
                const data = extractApiData<ServiceHealthStatus[] | unknown>(payload);
                const next = Array.isArray(data) ? data : [];
                set({
                    servicesStatus: next,
                    isFetchingServicesStatus: false,
                    servicesStatusUpdatedAt: new Date().toISOString(),
                    servicesStatusFailure: null,
                });
                return next;
            } catch {
                set({
                    isFetchingServicesStatus: false,
                    servicesStatusFailure: classifyServicesStatusFailure(),
                });
                return [];
            }
        },
    };
}
