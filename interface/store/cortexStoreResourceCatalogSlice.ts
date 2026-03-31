import type { CortexState } from '@/store/cortexStoreState';
import type {
    Artifact,
    ArtifactFilters,
    ArtifactStatus,
    CatalogueAgent,
    TeamProposal,
} from '@/store/cortexStoreTypes';
import type { CortexGet, CortexSet, CortexSlice } from '@/store/cortexStoreSliceTypes';

function readArray<T>(value: unknown): T[] {
    return Array.isArray(value) ? value : [];
}

function toggleSubscribedGroup(groups: string[], group: string): string[] {
    return groups.includes(group)
        ? groups.filter((item) => item !== group)
        : [...groups, group];
}

function updateProposalStatus(
    proposals: TeamProposal[],
    id: string,
    status: TeamProposal['status'],
): TeamProposal[] {
    return proposals.map((proposal) => (
        proposal.id === id ? { ...proposal, status } : proposal
    ));
}

function syncSelectedCatalogueAgent(
    selected: CatalogueAgent | null,
    id: string,
    next: CatalogueAgent | null,
): CatalogueAgent | null {
    return selected?.id === id ? next : selected;
}

function buildArtifactQuery(filters?: ArtifactFilters): string {
    const params = new URLSearchParams();
    if (filters?.mission_id) params.set('mission_id', filters.mission_id);
    if (filters?.team_id) params.set('team_id', filters.team_id);
    if (filters?.agent_id) params.set('agent_id', filters.agent_id);
    if (filters?.limit) params.set('limit', String(filters.limit));
    const query = params.toString();
    return query ? `/api/v1/artifacts?${query}` : '/api/v1/artifacts';
}

function syncArtifactStatus(artifact: Artifact, id: string, status: ArtifactStatus): Artifact {
    return artifact.id === id ? { ...artifact, status } : artifact;
}

export function createCortexResourceCatalogSlice(
    set: CortexSet,
    get: CortexGet,
): CortexSlice<
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
                return { subscribedSensorGroups: toggleSubscribedGroup(s.subscribedSensorGroups, group) };
            });
        },

        fetchProposals: async () => {
            set({ isFetchingProposals: true });
            try {
                const res = await fetch('/api/v1/proposals');
                if (res.ok) {
                    const data = await res.json();
                    set({ teamProposals: readArray<TeamProposal>(data.proposals), isFetchingProposals: false });
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
                        teamProposals: updateProposalStatus(s.teamProposals, id, 'approved'),
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
                        teamProposals: updateProposalStatus(s.teamProposals, id, 'rejected'),
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
                set({ missions: readArray(data), isFetchingMissions: false });
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
                    set({ catalogueAgents: readArray(data), isFetchingCatalogue: false });
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
                        selectedCatalogueAgent: syncSelectedCatalogueAgent(s.selectedCatalogueAgent, id, updated),
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
                        selectedCatalogueAgent: syncSelectedCatalogueAgent(s.selectedCatalogueAgent, id, null),
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
                const res = await fetch(buildArtifactQuery(filters));
                if (res.ok) {
                    const data = await res.json();
                    set({ artifacts: readArray(data), isFetchingArtifacts: false });
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
                    method: 'PUT',
                    headers: { 'Content-Type': 'application/json' },
                    body: JSON.stringify({ status }),
                });
                if (res.ok) {
                    set((s) => ({
                        artifacts: s.artifacts.map((artifact) => syncArtifactStatus(artifact, id, status as ArtifactStatus)),
                        selectedArtifactDetail: s.selectedArtifactDetail?.id === id
                            ? { ...s.selectedArtifactDetail, status: status as ArtifactStatus }
                            : s.selectedArtifactDetail,
                    }));
                }
            } catch {
                // degraded mode — keep local state unchanged
            }
        },
    };
}
