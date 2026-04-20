import type { CortexState } from '@/store/cortexStoreState';
import { initialCortexState } from '@/store/cortexStoreInitialState';
import { blueprintToGraph } from '@/store/cortexStoreUtils';
import { useCortexStore } from '@/store/useCortexStore';

export const baseBlueprint = {
    mission_id: 'mission-1',
    intent: 'Test mission',
    teams: [
        {
            name: 'Ops',
            role: 'operators',
            agents: [
                {
                    id: 'alpha',
                    role: 'cognitive',
                    system_prompt: 'First agent',
                    model: 'model-a',
                    inputs: ['source.topic'],
                    outputs: ['ops.alpha'],
                },
                {
                    id: 'beta',
                    role: 'sensory',
                    system_prompt: 'Second agent',
                    model: 'model-b',
                    inputs: ['ops.alpha'],
                    outputs: ['ops.beta'],
                },
            ],
        },
    ],
};

export function resetCortexStore(overrides: Partial<CortexState> = {}) {
    const { nodes, edges } = blueprintToGraph(baseBlueprint);

    localStorage.clear();
    useCortexStore.setState({
        ...initialCortexState,
        missions: [],
        isFetchingMissions: false,
        artifacts: [],
        isFetchingArtifacts: false,
        sensorFeeds: [],
        isFetchingSensors: false,
        teamProposals: [],
        isFetchingProposals: false,
        catalogueAgents: [],
        isFetchingCatalogue: false,
        missionChat: [],
        workspaceChatScope: null,
        missionChatError: null,
        missionChatFailure: null,
        workspaceChatPrimed: false,
        councilTarget: 'admin',
        assistantName: 'Soma',
        theme: 'aero-light',
        mcpServers: [],
        isFetchingMCPServers: false,
        mcpServersError: null,
        mcpActivity: [],
        isFetchingMCPActivity: false,
        mcpTools: [],
        trustThreshold: 0.7,
        isSyncingThreshold: false,
        blueprint: null,
        nodes,
        edges,
        missionStatus: 'idle',
        activeMissionId: null,
        selectedAgentNodeId: null,
        isAgentEditorOpen: false,
        selectedArtifact: null,
        selectedArtifactDetail: null,
        selectedCatalogueAgent: null,
        pendingArtifacts: [],
        subscribedSensorGroups: [],
        auditLog: [],
        pendingApprovals: [],
        activeMode: 'answer',
        activeConfirmToken: null,
        activeRunId: null,
        pendingProposal: null,
        lastCommitProof: null,
        lastBroadcastResult: null,
        isBroadcasting: false,
        activeBrain: null,
        activeRole: '',
        ...overrides,
    });
}
