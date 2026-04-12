import type {
    CortexAutomationRunsContract,
    CortexDraftGraphContract,
    CortexGovernanceOpsContract,
    CortexMissionChatContract,
    CortexProfilesSettingsContract,
    CortexResourcesContract,
    CortexState,
} from '@/store/cortexStoreState';
import { deriveMissionChatState } from '@/store/cortexStoreChatWorkflow';
import { loadPersistedChat } from '@/store/cortexStoreUtils';

type StripActions<T> = {
    [K in keyof T as T[K] extends (...args: any[]) => any ? never : K]: T[K];
};

const initialMissionChat = loadPersistedChat(null);
const initialMissionChatState = deriveMissionChatState(initialMissionChat);

function readInitialAdvancedMode(): boolean {
    return typeof window !== 'undefined'
        && typeof localStorage !== 'undefined'
        && typeof localStorage.getItem === 'function'
        ? localStorage.getItem('mycelis-advanced-mode') === 'true'
        : false;
}

const initialDraftGraphState: StripActions<CortexDraftGraphContract> = {
    chatHistory: [],
    nodes: [],
    edges: [],
    isDrafting: false,
    isCommitting: false,
    error: null,
    blueprint: null,
    missionStatus: 'idle',
    activeMissionId: null,
    savedBlueprints: [],
    isBlueprintDrawerOpen: false,
    selectedAgentNodeId: null,
    isAgentEditorOpen: false,
};

const initialResourcesState: StripActions<CortexResourcesContract> = {
    pendingArtifacts: [],
    selectedArtifact: null,
    missions: [],
    isFetchingMissions: false,
    trustThreshold: 0.7,
    isSyncingThreshold: false,
    isToolsPaletteOpen: false,
    sensorFeeds: [],
    isFetchingSensors: false,
    subscribedSensorGroups: [],
    teamProposals: [],
    isFetchingProposals: false,
    catalogueAgents: [],
    isFetchingCatalogue: false,
    selectedCatalogueAgent: null,
    artifacts: [],
    isFetchingArtifacts: false,
    selectedArtifactDetail: null,
    mcpServers: [],
    isFetchingMCPServers: false,
    mcpServersError: null,
    mcpActivity: [],
    isFetchingMCPActivity: false,
    mcpTools: [],
    mcpLibrary: [],
    isFetchingMCPLibrary: false,
};

const initialMissionChatContractState: StripActions<CortexMissionChatContract> = {
    workspaceChatScope: null,
    missionChat: initialMissionChat,
    isMissionChatting: false,
    missionChatError: null,
    missionChatFailure: null,
    workspaceChatPrimed: false,
    pendingProposal: initialMissionChatState.pendingProposal,
    activeConfirmToken: initialMissionChatState.activeConfirmToken,
    lastCommitProof: null,
    activeRunId: initialMissionChatState.activeRunId,
    assistantName: 'Soma',
    councilTarget: 'admin',
    councilMembers: [],
    isBroadcasting: false,
    lastBroadcastResult: null,
    activeBrain: null,
    activeMode: initialMissionChatState.activeMode,
    activeRole: '',
    governanceMode: 'passive',
};

const initialGovernanceOpsState: StripActions<CortexGovernanceOpsContract> = {
    streamLogs: [],
    isStreamConnected: false,
    streamConnectionState: 'idle',
    activeSquadRoomId: null,
    teamRoster: [],
    isFetchingTeamRoster: false,
    policyConfig: null,
    pendingApprovals: [],
    isFetchingPolicy: false,
    isFetchingApprovals: false,
    auditLog: [],
    isFetchingAuditLog: false,
    cognitiveStatus: null,
    servicesStatus: [],
    isFetchingServicesStatus: false,
    servicesStatusUpdatedAt: null,
    teamsDetail: [],
    isFetchingTeamsDetail: false,
    selectedTeamId: null,
    isTeamDrawerOpen: false,
    teamsFilter: 'all',
    selectedSignalDetail: null,
    inspectedMessage: null,
    isInspectorOpen: false,
    isStatusDrawerOpen: false,
};

const initialAutomationRunsState: StripActions<CortexAutomationRunsContract> = {
    runTimeline: null,
    isFetchingTimeline: false,
    recentRuns: [],
    isFetchingRuns: false,
    triggerRules: [],
    isFetchingTriggers: false,
    conversationTurns: null,
    isFetchingConversation: false,
};

const initialProfilesSettingsState: StripActions<CortexProfilesSettingsContract> = {
    advancedMode: readInitialAdvancedMode(),
    theme: 'aero-light',
    missionProfiles: [],
    activeProfileId: null,
    contextSnapshots: [],
};

// Keep the default store state isolated from action logic so product behavior
// slices are easier to review and reuse without re-scanning the whole store.
export const initialCortexState: StripActions<CortexState> = {
    ...initialDraftGraphState,
    ...initialResourcesState,
    ...initialMissionChatContractState,
    ...initialGovernanceOpsState,
    ...initialAutomationRunsState,
    ...initialProfilesSettingsState,
};
