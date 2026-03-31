import type { CortexState } from '@/store/cortexStoreState';
import type { CortexGet, CortexSet } from '@/store/cortexStoreSliceTypes';
import type { AgentManifest, MissionBlueprint, TeamsFilter } from '@/store/cortexStoreTypes';
import { blueprintToGraph, solidifyNodes } from '@/store/cortexStoreUtils';

function buildClosedAgentEditorState() {
    return {
        selectedAgentNodeId: null,
        isAgentEditorOpen: false,
    } as const;
}

function buildResetDraftState() {
    return {
        blueprint: null,
        nodes: [],
        edges: [],
        missionStatus: 'idle' as const,
        activeMissionId: null,
        ...buildClosedAgentEditorState(),
    };
}

function applyBlueprintDraftState(set: CortexSet, blueprint: MissionBlueprint, solidify = false) {
    const { nodes, edges } = blueprintToGraph(blueprint);
    set({
        blueprint,
        nodes: solidify ? solidifyNodes(nodes) : nodes,
        edges,
        ...buildClosedAgentEditorState(),
    });
}

function mutateDraftBlueprint(
    get: CortexGet,
    set: CortexSet,
    mutate: (blueprint: MissionBlueprint) => boolean,
    solidify = false,
) {
    const blueprint = get().blueprint;
    if (!blueprint) return false;

    const nextBlueprint: MissionBlueprint = structuredClone(blueprint);
    if (!mutate(nextBlueprint)) {
        return false;
    }

    applyBlueprintDraftState(set, nextBlueprint, solidify);
    return true;
}

function removeAgentAt(blueprint: MissionBlueprint, teamIdx: number, agentIdx: number) {
    const team = blueprint.teams[teamIdx];
    if (!team || !team.agents[agentIdx]) {
        return false;
    }

    team.agents.splice(agentIdx, 1);
    if (team.agents.length === 0) {
        blueprint.teams.splice(teamIdx, 1);
    }
    return true;
}

function removeAgentById(blueprint: MissionBlueprint, agentId: string) {
    for (let teamIdx = 0; teamIdx < blueprint.teams.length; teamIdx += 1) {
        const agentIdx = blueprint.teams[teamIdx].agents.findIndex((item) => item.id === agentId);
        if (agentIdx !== -1) {
            removeAgentAt(blueprint, teamIdx, agentIdx);
            return true;
        }
    }
    return false;
}

export function createCortexMissionDraftSlice(
    set: CortexSet,
    get: CortexGet,
): Pick<
    CortexState,
    | 'instantiateMission'
    | 'fetchTeamsDetail'
    | 'selectTeam'
    | 'setTeamsFilter'
    | 'selectAgentNode'
    | 'updateAgentInDraft'
    | 'deleteAgentFromDraft'
    | 'discardDraft'
    | 'updateAgentInMission'
    | 'deleteAgentFromMission'
    | 'deleteMission'
> {
    return {
        instantiateMission: async () => {
            const { blueprint } = get();
            if (!blueprint) return;

            set({ isCommitting: true, error: null });

            try {
                const res = await fetch('/api/v1/intent/commit', {
                    method: 'POST',
                    headers: { 'Content-Type': 'application/json' },
                    body: JSON.stringify(blueprint),
                });

                const data = await res.json();

                if (data.error) {
                    set((s) => ({
                        isCommitting: false,
                        error: data.error,
                        chatHistory: [
                            ...s.chatHistory,
                            { role: 'architect', content: `Commit failed: ${data.error}` },
                        ],
                    }));
                    return;
                }

                set((s) => ({
                    isCommitting: false,
                    missionStatus: 'active',
                    activeMissionId: data.mission_id,
                    nodes: solidifyNodes(s.nodes),
                    chatHistory: [
                        ...s.chatHistory,
                        {
                            role: 'architect',
                            content: `Mission **${data.mission_id}** instantiated. ${data.teams} teams, ${data.agents} agents now ACTIVE.`,
                        },
                    ],
                }));
            } catch (err) {
                const msg = err instanceof Error ? err.message : 'Commit failed';
                set((s) => ({
                    isCommitting: false,
                    error: msg,
                    chatHistory: [
                        ...s.chatHistory,
                        { role: 'architect', content: `Error: ${msg}` },
                    ],
                }));
            }
        },

        fetchTeamsDetail: async () => {
            set({ isFetchingTeamsDetail: true });
            try {
                const res = await fetch('/api/v1/teams/detail');
                if (res.ok) {
                    const data = await res.json();
                    set({ teamsDetail: Array.isArray(data) ? data : [], isFetchingTeamsDetail: false });
                } else {
                    set({ teamsDetail: [], isFetchingTeamsDetail: false });
                }
            } catch {
                set({ teamsDetail: [], isFetchingTeamsDetail: false });
            }
        },

        selectTeam: (teamId: string | null) => {
            set({
                selectedTeamId: teamId,
                isTeamDrawerOpen: teamId !== null,
            });
        },

        setTeamsFilter: (filter: TeamsFilter) => {
            set({ teamsFilter: filter });
        },

        selectAgentNode: (nodeId: string | null) => {
            set({
                selectedAgentNodeId: nodeId,
                isAgentEditorOpen: nodeId !== null,
            });
        },

        updateAgentInDraft: (teamIdx: number, agentIdx: number, updates: Partial<AgentManifest>) => {
            mutateDraftBlueprint(get, set, (blueprint) => {
                const agent = blueprint.teams[teamIdx]?.agents[agentIdx];
                if (!agent) {
                    return false;
                }
                Object.assign(agent, updates);
                return true;
            });
        },

        deleteAgentFromDraft: (teamIdx: number, agentIdx: number) => {
            mutateDraftBlueprint(get, set, (blueprint) => removeAgentAt(blueprint, teamIdx, agentIdx));
        },

        discardDraft: () => {
            set(buildResetDraftState());
        },

        updateAgentInMission: async (agentName: string, manifest: Partial<AgentManifest>) => {
            const { activeMissionId, blueprint } = get();
            if (!activeMissionId) return;

            try {
                const res = await fetch(`/api/v1/missions/${activeMissionId}/agents/${agentName}`, {
                    method: 'PUT',
                    headers: { 'Content-Type': 'application/json' },
                    body: JSON.stringify(manifest),
                });
                if (!res.ok) {
                    const data = await res.json().catch(() => ({}));
                    console.error('updateAgentInMission:', data.error ?? res.statusText);
                    return;
                }

                if (blueprint) {
                    mutateDraftBlueprint(get, set, (draftBlueprint) => {
                        for (const team of draftBlueprint.teams) {
                            const agent = team.agents.find((item) => item.id === agentName);
                            if (agent) {
                                Object.assign(agent, manifest);
                                return true;
                            }
                        }
                        return false;
                    }, true);
                }
            } catch (err) {
                console.error('updateAgentInMission:', err);
            }
        },

        deleteAgentFromMission: async (agentName: string) => {
            const { activeMissionId, blueprint } = get();
            if (!activeMissionId) return;

            try {
                const res = await fetch(`/api/v1/missions/${activeMissionId}/agents/${agentName}`, {
                    method: 'DELETE',
                });
                if (!res.ok) {
                    const data = await res.json().catch(() => ({}));
                    console.error('deleteAgentFromMission error:', data.error ?? res.statusText);
                    return;
                }

                if (blueprint) {
                    const removed = mutateDraftBlueprint(get, set, (draftBlueprint) => removeAgentById(draftBlueprint, agentName), true);
                    if (!removed) {
                        console.warn('deleteAgentFromMission: active agent not found in local blueprint:', agentName);
                    }
                }
            } catch (err) {
                console.error('deleteAgentFromMission:', err);
            }
        },

        deleteMission: async (missionId: string) => {
            try {
                const res = await fetch(`/api/v1/missions/${missionId}`, {
                    method: 'DELETE',
                });
                if (!res.ok) {
                    const data = await res.json().catch(() => ({}));
                    console.error('deleteMission:', data.error ?? res.statusText);
                    return;
                }

                set(buildResetDraftState());
            } catch (err) {
                console.error('deleteMission:', err);
            }
        },
    };
}
