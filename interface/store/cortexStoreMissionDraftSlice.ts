import type { CortexState } from '@/store/cortexStoreState';
import type { CortexGet, CortexSet } from '@/store/cortexStoreSliceTypes';
import type { AgentManifest, MissionBlueprint, TeamsFilter } from '@/store/cortexStoreTypes';
import { blueprintToGraph, solidifyNodes } from '@/store/cortexStoreUtils';

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
            const bp = get().blueprint;
            if (!bp) return;
            const newBp: MissionBlueprint = structuredClone(bp);
            const agent = newBp.teams[teamIdx]?.agents[agentIdx];
            if (!agent) return;
            Object.assign(agent, updates);
            const { nodes, edges } = blueprintToGraph(newBp);
            set({ blueprint: newBp, nodes, edges, selectedAgentNodeId: null, isAgentEditorOpen: false });
        },

        deleteAgentFromDraft: (teamIdx: number, agentIdx: number) => {
            const bp = get().blueprint;
            if (!bp) return;
            const newBp: MissionBlueprint = structuredClone(bp);
            const team = newBp.teams[teamIdx];
            if (!team) return;
            team.agents.splice(agentIdx, 1);
            if (team.agents.length === 0) {
                newBp.teams.splice(teamIdx, 1);
            }
            const { nodes, edges } = blueprintToGraph(newBp);
            set({ blueprint: newBp, nodes, edges, selectedAgentNodeId: null, isAgentEditorOpen: false });
        },

        discardDraft: () => {
            set({
                blueprint: null,
                nodes: [],
                edges: [],
                missionStatus: 'idle',
                activeMissionId: null,
                selectedAgentNodeId: null,
                isAgentEditorOpen: false,
            });
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
                    const newBp: MissionBlueprint = structuredClone(blueprint);
                    for (const team of newBp.teams) {
                        const agent = team.agents.find((item) => item.id === agentName);
                        if (agent) {
                            Object.assign(agent, manifest);
                            break;
                        }
                    }
                    const { nodes, edges } = blueprintToGraph(newBp);
                    set({ blueprint: newBp, nodes: solidifyNodes(nodes), edges, selectedAgentNodeId: null, isAgentEditorOpen: false });
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
                    const newBp: MissionBlueprint = structuredClone(blueprint);
                    let spliced = false;
                    for (let teamIdx = 0; teamIdx < newBp.teams.length; teamIdx += 1) {
                        const agentIdx = newBp.teams[teamIdx].agents.findIndex((item) => item.id === agentName);
                        if (agentIdx !== -1) {
                            newBp.teams[teamIdx].agents.splice(agentIdx, 1);
                            spliced = true;
                            if (newBp.teams[teamIdx].agents.length === 0) {
                                newBp.teams.splice(teamIdx, 1);
                            }
                            break;
                        }
                    }
                    if (!spliced) {
                        console.warn('deleteAgentFromMission: active agent not found in local blueprint:', agentName);
                    }

                    const { nodes, edges } = blueprintToGraph(newBp);
                    set({ blueprint: newBp, nodes: solidifyNodes(nodes), edges, selectedAgentNodeId: null, isAgentEditorOpen: false });
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

                set({
                    blueprint: null,
                    nodes: [],
                    edges: [],
                    missionStatus: 'idle',
                    activeMissionId: null,
                    selectedAgentNodeId: null,
                    isAgentEditorOpen: false,
                });
            } catch (err) {
                console.error('deleteMission:', err);
            }
        },
    };
}
