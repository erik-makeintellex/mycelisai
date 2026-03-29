import type { CortexState } from '@/store/cortexStoreState';
import type {
    ArtifactFilters,
    CatalogueAgent,
    TeamProposal,
} from '@/store/cortexStoreTypes';
import type { CortexGet, CortexSet } from '@/store/cortexStoreSliceTypes';

export function createCortexResourceCatalogSlice(
    set: CortexSet,
    get: CortexGet,
): Pick<
    CortexState,
    | 'fetchSensors'
    | 'toggleSensorGroup'
    | 'fetchProposals'
    | 'approveProposal'
    | 'rejectProposal'
    | 'fetchMissions'
    | 'fetchCatalogue'
    | 'createCatalogueAgent'
    | 'updateCatalogueAgent'
    | 'deleteCatalogueAgent'
    | 'selectCatalogueAgent'
    | 'fetchArtifacts'
    | 'getArtifactDetail'
    | 'updateArtifactStatus'
> {
    return {
        fetchSensors: async () => {
            set({ isFetchingSensors: true });
            try {
                const res = await fetch('/api/v1/sensors');
                if (res.ok) {
                    const data = await res.json();
                    set({ sensorFeeds: data.sensors ?? [], isFetchingSensors: false });
                } else {
                    set({ sensorFeeds: [], isFetchingSensors: false });
                }
            } catch {
                set({ sensorFeeds: [], isFetchingSensors: false });
            }
        },

        toggleSensorGroup: (group: string) => {
            set((s) => {
                const current = s.subscribedSensorGroups;
                const next = current.includes(group)
                    ? current.filter((item) => item !== group)
                    : [...current, group];
                return { subscribedSensorGroups: next };
            });
        },

        fetchProposals: async () => {
            set({ isFetchingProposals: true });
            try {
                const res = await fetch('/api/v1/proposals');
                if (res.ok) {
                    const data = await res.json();
                    set({ teamProposals: data.proposals ?? [], isFetchingProposals: false });
                } else {
                    set({ teamProposals: [], isFetchingProposals: false });
                }
            } catch {
                set({ teamProposals: [], isFetchingProposals: false });
            }
        },

        approveProposal: async (id: string) => {
            try {
                const res = await fetch(`/api/v1/proposals/${id}/approve`, { method: 'POST' });
                if (res.ok) {
                    set((s) => ({
                        teamProposals: s.teamProposals.map((proposal: TeamProposal) =>
                            proposal.id === id ? { ...proposal, status: 'approved' } : proposal
                        ),
                    }));
                }
            } catch {
                // no-op in degraded mode
            }
        },

        rejectProposal: async (id: string) => {
            try {
                const res = await fetch(`/api/v1/proposals/${id}/reject`, { method: 'POST' });
                if (res.ok) {
                    set((s) => ({
                        teamProposals: s.teamProposals.map((proposal: TeamProposal) =>
                            proposal.id === id ? { ...proposal, status: 'rejected' } : proposal
                        ),
                    }));
                }
            } catch {
                // no-op in degraded mode
            }
        },

        fetchMissions: async () => {
            set({ isFetchingMissions: true });
            try {
                const res = await fetch('/api/v1/missions');
                if (!res.ok) {
                    set({ missions: [], isFetchingMissions: false });
                    return;
                }
                const data = await res.json();
                set({ missions: Array.isArray(data) ? data : [], isFetchingMissions: false });
            } catch {
                set({ missions: [], isFetchingMissions: false });
            }
        },

        fetchCatalogue: async () => {
            set({ isFetchingCatalogue: true });
            try {
                const res = await fetch('/api/v1/catalogue/agents');
                if (res.ok) {
                    const data = await res.json();
                    set({ catalogueAgents: Array.isArray(data) ? data : [], isFetchingCatalogue: false });
                } else {
                    set({ catalogueAgents: [], isFetchingCatalogue: false });
                }
            } catch {
                set({ catalogueAgents: [], isFetchingCatalogue: false });
            }
        },

        createCatalogueAgent: async (agent: Partial<CatalogueAgent>) => {
            try {
                const res = await fetch('/api/v1/catalogue/agents', {
                    method: 'POST',
                    headers: { 'Content-Type': 'application/json' },
                    body: JSON.stringify(agent),
                });
                if (res.ok) {
                    const created = await res.json();
                    set((s) => ({
                        catalogueAgents: [created, ...s.catalogueAgents],
                        selectedCatalogueAgent: created,
                    }));
                }
            } catch {
                // degraded mode — keep local state unchanged
            }
        },

        updateCatalogueAgent: async (id: string, agent: Partial<CatalogueAgent>) => {
            try {
                const res = await fetch(`/api/v1/catalogue/agents/${id}`, {
                    method: 'PUT',
                    headers: { 'Content-Type': 'application/json' },
                    body: JSON.stringify(agent),
                });
                if (res.ok) {
                    const updated = await res.json();
                    set((s) => ({
                        catalogueAgents: s.catalogueAgents.map((item) => item.id === id ? updated : item),
                        selectedCatalogueAgent: s.selectedCatalogueAgent?.id === id ? updated : s.selectedCatalogueAgent,
                    }));
                }
            } catch {
                // degraded mode — keep local state unchanged
            }
        },

        deleteCatalogueAgent: async (id: string) => {
            try {
                const res = await fetch(`/api/v1/catalogue/agents/${id}`, { method: 'DELETE' });
                if (res.ok) {
                    set((s) => ({
                        catalogueAgents: s.catalogueAgents.filter((item) => item.id !== id),
                        selectedCatalogueAgent: s.selectedCatalogueAgent?.id === id ? null : s.selectedCatalogueAgent,
                    }));
                }
            } catch {
                // degraded mode — keep local state unchanged
            }
        },

        selectCatalogueAgent: (agent: CatalogueAgent | null) => {
            set({ selectedCatalogueAgent: agent });
        },

        fetchArtifacts: async (filters?: ArtifactFilters) => {
            set({ isFetchingArtifacts: true });
            try {
                const params = new URLSearchParams();
                if (filters?.mission_id) params.set('mission_id', filters.mission_id);
                if (filters?.team_id) params.set('team_id', filters.team_id);
                if (filters?.agent_id) params.set('agent_id', filters.agent_id);
                if (filters?.limit) params.set('limit', String(filters.limit));
                const query = params.toString();
                const res = await fetch(query ? `/api/v1/artifacts?${query}` : '/api/v1/artifacts');
                if (res.ok) {
                    const data = await res.json();
                    set({ artifacts: Array.isArray(data) ? data : [], isFetchingArtifacts: false });
                } else {
                    set({ artifacts: [], isFetchingArtifacts: false });
                }
            } catch {
                set({ artifacts: [], isFetchingArtifacts: false });
            }
        },

        getArtifactDetail: async (id: string) => {
            try {
                const res = await fetch(`/api/v1/artifacts/${id}`);
                if (res.ok) {
                    const data = await res.json();
                    set({ selectedArtifactDetail: data });
                }
            } catch {
                // degraded mode — keep local state unchanged
            }
        },

        updateArtifactStatus: async (id: string, status: string) => {
            try {
                const res = await fetch(`/api/v1/artifacts/${id}/status`, {
                    method: 'POST',
                    headers: { 'Content-Type': 'application/json' },
                    body: JSON.stringify({ status }),
                });
                if (res.ok) {
                    set((s) => ({
                        artifacts: s.artifacts.map((artifact) =>
                            artifact.id === id ? { ...artifact, status: status as any } : artifact
                        ),
                        selectedArtifactDetail: s.selectedArtifactDetail?.id === id
                            ? { ...s.selectedArtifactDetail, status: status as any }
                            : s.selectedArtifactDetail,
                    }));
                }
            } catch {
                // degraded mode — keep local state unchanged
            }
        },
    };
}
