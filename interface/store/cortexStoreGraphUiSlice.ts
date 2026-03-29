import { applyEdgeChanges, applyNodeChanges } from 'reactflow';
import type { CortexState } from '@/store/cortexStoreState';
import { blueprintToGraph } from '@/store/cortexStoreUtils';
import type { CortexGet, CortexSet } from '@/store/cortexStoreSliceTypes';
import type { ChatMessage, CTSEnvelope, MissionBlueprint, ProposalData, SignalDetail } from '@/store/cortexStoreTypes';

export function createCortexGraphUiSlice(
    set: CortexSet,
    get: CortexGet,
): Pick<
    CortexState,
    | 'onNodesChange'
    | 'onEdgesChange'
    | 'submitIntent'
    | 'enterSquadRoom'
    | 'exitSquadRoom'
    | 'selectArtifact'
    | 'selectSignalDetail'
    | 'setInspectedMessage'
    | 'approveArtifact'
    | 'rejectArtifact'
    | 'setTrustThreshold'
    | 'fetchTrustThreshold'
    | 'toggleBlueprintDrawer'
    | 'toggleAdvancedMode'
    | 'toggleToolsPalette'
    | 'setStatusDrawerOpen'
    | 'saveBlueprint'
    | 'loadBlueprint'
> {
    return {
        onNodesChange: (changes) => {
            set({ nodes: applyNodeChanges(changes, get().nodes) });
        },

        onEdgesChange: (changes) => {
            set({ edges: applyEdgeChanges(changes, get().edges) });
        },

        submitIntent: async (text: string) => {
            const trimmed = text.trim();
            if (!trimmed) return;

            set((s) => ({
                chatHistory: [...s.chatHistory, { role: 'user', content: trimmed }],
                isDrafting: true,
                error: null,
                missionStatus: 'idle',
                activeMissionId: null,
            }));

            try {
                const res = await fetch('/api/v1/intent/negotiate', {
                    method: 'POST',
                    headers: { 'Content-Type': 'application/json' },
                    body: JSON.stringify({ intent: trimmed }),
                });

                const data = await res.json();

                if (data.error) {
                    set((s) => ({
                        isDrafting: false,
                        error: data.error,
                        chatHistory: [
                            ...s.chatHistory,
                            { role: 'architect', content: `Error: ${data.error}` },
                        ],
                    }));
                    return;
                }

                const bp = (data.blueprint ?? data) as MissionBlueprint;
                const { nodes, edges } = blueprintToGraph(bp);

                const agentCount = bp.teams.reduce((sum, team) => sum + team.agents.length, 0);
                const summary = [
                    `Blueprint **${bp.mission_id}** generated.`,
                    `${bp.teams.length} team${bp.teams.length !== 1 ? 's' : ''}, ${agentCount} agent${agentCount !== 1 ? 's' : ''}.`,
                    bp.constraints && bp.constraints.length > 0
                        ? `${bp.constraints.length} constraint${bp.constraints.length !== 1 ? 's' : ''} applied.`
                        : '',
                    data.confirm_token ? 'Confirm token issued.' : '',
                ]
                    .filter(Boolean)
                    .join(' ');

                const proposalData: ProposalData | null = data.confirm_token ? {
                    intent: bp.intent,
                    teams: bp.teams.length,
                    agents: agentCount,
                    tools: data.intent_proof?.scope_validation?.tools || [],
                    risk_level: data.intent_proof?.scope_validation?.risk_level || 'medium',
                    confirm_token: data.confirm_token.token,
                    intent_proof_id: data.intent_proof?.id || '',
                } : null;

                set((s) => ({
                    isDrafting: false,
                    blueprint: bp,
                    nodes,
                    edges,
                    missionStatus: 'draft',
                    pendingProposal: proposalData,
                    activeConfirmToken: data.confirm_token?.token || null,
                    chatHistory: [...s.chatHistory, { role: 'architect', content: summary }],
                }));
            } catch (err) {
                const msg = err instanceof Error ? err.message : 'Request failed';
                set((s) => ({
                    isDrafting: false,
                    error: msg,
                    chatHistory: [
                        ...s.chatHistory,
                        { role: 'architect', content: `Error: ${msg}` },
                    ],
                }));
            }
        },

        enterSquadRoom: (teamId: string) => {
            set({ activeSquadRoomId: teamId });
        },

        exitSquadRoom: () => {
            set({ activeSquadRoomId: null });
        },

        selectArtifact: (artifact: CTSEnvelope | null) => {
            set({ selectedArtifact: artifact });
        },

        selectSignalDetail: (detail: SignalDetail | null) => {
            set({ selectedSignalDetail: detail });
        },

        setInspectedMessage: (msg: ChatMessage | null) => {
            set({ inspectedMessage: msg, isInspectorOpen: msg !== null });
        },

        approveArtifact: (id: string) => {
            const artifact = get().pendingArtifacts.find((item) => item.id === id);
            if (artifact) {
                console.log('[GOVERNANCE] APPROVED:', id, artifact);
            }
            set((s) => ({
                pendingArtifacts: s.pendingArtifacts.filter((item) => item.id !== id),
                selectedArtifact: s.selectedArtifact?.id === id ? null : s.selectedArtifact,
            }));
        },

        rejectArtifact: (id: string, reason: string) => {
            const artifact = get().pendingArtifacts.find((item) => item.id === id);
            if (artifact) {
                console.log('[GOVERNANCE] REJECTED:', id, reason, artifact);
            }
            set((s) => ({
                pendingArtifacts: s.pendingArtifacts.filter((item) => item.id !== id),
                selectedArtifact: s.selectedArtifact?.id === id ? null : s.selectedArtifact,
            }));
        },

        setTrustThreshold: (value: number) => {
            set({ trustThreshold: value, isSyncingThreshold: true });
            fetch('/api/v1/trust/threshold', {
                method: 'PUT',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify({ threshold: value }),
            })
                .catch((err) => console.error('[TRUST] Failed to sync threshold:', err))
                .finally(() => set({ isSyncingThreshold: false }));
        },

        fetchTrustThreshold: async () => {
            try {
                const res = await fetch('/api/v1/trust/threshold');
                if (res.ok) {
                    const data = await res.json();
                    set({ trustThreshold: data.threshold });
                }
            } catch {
                // degraded mode — use local default
            }
        },

        toggleBlueprintDrawer: () => {
            set((s) => ({ isBlueprintDrawerOpen: !s.isBlueprintDrawerOpen }));
        },

        toggleAdvancedMode: () => {
            set((s) => {
                const next = !s.advancedMode;
                if (typeof window !== 'undefined') {
                    localStorage.setItem('mycelis-advanced-mode', String(next));
                }
                return { advancedMode: next };
            });
        },

        toggleToolsPalette: () => {
            set((s) => ({ isToolsPaletteOpen: !s.isToolsPaletteOpen }));
        },

        setStatusDrawerOpen: (open: boolean) => {
            set({ isStatusDrawerOpen: open });
        },

        saveBlueprint: (bp: MissionBlueprint) => {
            set((s) => ({
                savedBlueprints: [bp, ...s.savedBlueprints.filter((item) => item.mission_id !== bp.mission_id)],
            }));
        },

        loadBlueprint: (bp: MissionBlueprint) => {
            const { nodes, edges } = blueprintToGraph(bp);
            set({
                blueprint: bp,
                nodes,
                edges,
                missionStatus: 'draft',
                isBlueprintDrawerOpen: false,
            });
        },
    };
}
