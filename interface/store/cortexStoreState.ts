import type { OnEdgesChange, OnNodesChange } from 'reactflow';
import type { MissionChatFailure } from '@/lib/missionChatFailure';
import type { ConversationTurn } from '@/types/conversations';
import type {
    AgentManifest,
    Artifact,
    ArtifactFilters,
    AuditLogEntry,
    BrainProvenance,
    CatalogueAgent,
    ChatMessage,
    CognitiveStatus,
    ConfirmProposalResult,
    ContextSnapshot,
    CouncilMember,
    CTSEnvelope,
    ExecutionMode,
    MCPActivityEntry,
    MCPServerWithTools,
    MCPLibraryCategory,
    MCPInstallResult,
    MCPTool,
    Mission,
    MissionBlueprint,
    MissionEvent,
    MissionProfile,
    MissionProfileCreate,
    MissionRun,
    MissionStatus,
    PendingApproval,
    PolicyConfig,
    ProposalData,
    SensorNode,
    ServiceHealthStatus,
    SignalDetail,
    StreamSignal,
    TeamDetail,
    TeamDetailEntry,
    TeamProposal,
    TeamsFilter,
    TriggerRule,
    TriggerRuleCreate,
} from '@/store/cortexStoreTypes';

// Keep the runtime store focused on behavior while the public store contract
// stays grouped by product domain so UI callers can find the right boundary fast.

export interface CortexDraftGraphContract {
    chatHistory: ChatMessage[];
    nodes: import('reactflow').Node[];
    edges: import('reactflow').Edge[];
    isDrafting: boolean;
    isCommitting: boolean;
    error: string | null;
    blueprint: MissionBlueprint | null;
    missionStatus: MissionStatus;
    activeMissionId: string | null;
    savedBlueprints: MissionBlueprint[];
    isBlueprintDrawerOpen: boolean;
    selectedAgentNodeId: string | null;
    isAgentEditorOpen: boolean;
    onNodesChange: OnNodesChange;
    onEdgesChange: OnEdgesChange;
    submitIntent: (text: string) => Promise<void>;
    instantiateMission: () => Promise<void>;
    toggleBlueprintDrawer: () => void;
    saveBlueprint: (bp: MissionBlueprint) => void;
    loadBlueprint: (bp: MissionBlueprint) => void;
    selectAgentNode: (nodeId: string | null) => void;
    updateAgentInDraft: (teamIdx: number, agentIdx: number, updates: Partial<AgentManifest>) => void;
    deleteAgentFromDraft: (teamIdx: number, agentIdx: number) => void;
    discardDraft: () => void;
    updateAgentInMission: (agentName: string, manifest: Partial<AgentManifest>) => Promise<void>;
    deleteAgentFromMission: (agentName: string) => Promise<void>;
    deleteMission: (missionId: string) => Promise<void>;
}

export interface CortexResourcesContract {
    pendingArtifacts: CTSEnvelope[];
    selectedArtifact: CTSEnvelope | null;
    missions: Mission[];
    isFetchingMissions: boolean;
    trustThreshold: number;
    isSyncingThreshold: boolean;
    isToolsPaletteOpen: boolean;
    sensorFeeds: SensorNode[];
    isFetchingSensors: boolean;
    subscribedSensorGroups: string[];
    teamProposals: TeamProposal[];
    isFetchingProposals: boolean;
    catalogueAgents: CatalogueAgent[];
    isFetchingCatalogue: boolean;
    selectedCatalogueAgent: CatalogueAgent | null;
    artifacts: Artifact[];
    isFetchingArtifacts: boolean;
    selectedArtifactDetail: Artifact | null;
    mcpServers: MCPServerWithTools[];
    isFetchingMCPServers: boolean;
    mcpActivity: MCPActivityEntry[];
    isFetchingMCPActivity: boolean;
    mcpTools: MCPTool[];
    mcpLibrary: MCPLibraryCategory[];
    isFetchingMCPLibrary: boolean;
    fetchMissions: () => Promise<void>;
    selectArtifact: (artifact: CTSEnvelope | null) => void;
    approveArtifact: (id: string) => void;
    rejectArtifact: (id: string, reason: string) => void;
    setTrustThreshold: (value: number) => void;
    fetchTrustThreshold: () => Promise<void>;
    toggleToolsPalette: () => void;
    fetchSensors: () => Promise<void>;
    toggleSensorGroup: (group: string) => void;
    fetchProposals: () => Promise<void>;
    approveProposal: (id: string) => Promise<void>;
    rejectProposal: (id: string) => Promise<void>;
    fetchCatalogue: () => Promise<void>;
    createCatalogueAgent: (agent: Partial<CatalogueAgent>) => Promise<void>;
    updateCatalogueAgent: (id: string, agent: Partial<CatalogueAgent>) => Promise<void>;
    deleteCatalogueAgent: (id: string) => Promise<void>;
    selectCatalogueAgent: (agent: CatalogueAgent | null) => void;
    fetchArtifacts: (filters?: ArtifactFilters) => Promise<void>;
    getArtifactDetail: (id: string) => Promise<void>;
    updateArtifactStatus: (id: string, status: string) => Promise<void>;
    fetchMCPServers: () => Promise<void>;
    fetchMCPActivity: () => Promise<void>;
    deleteMCPServer: (id: string) => Promise<void>;
    fetchMCPTools: () => Promise<void>;
    fetchMCPLibrary: () => Promise<void>;
    installFromLibrary: (name: string, env?: Record<string, string>) => Promise<MCPInstallResult>;
}

export interface CortexMissionChatContract {
    workspaceChatScope: string | null;
    missionChat: ChatMessage[];
    isMissionChatting: boolean;
    missionChatError: string | null;
    missionChatFailure: MissionChatFailure | null;
    workspaceChatPrimed: boolean;
    pendingProposal: ProposalData | null;
    activeConfirmToken: string | null;
    lastCommitProof: { intent_proof_id: string; audit_event_id: string } | null;
    activeRunId: string | null;
    assistantName: string;
    councilTarget: string;
    councilMembers: CouncilMember[];
    isBroadcasting: boolean;
    lastBroadcastResult: { teams_hit: number } | null;
    activeBrain: BrainProvenance | null;
    activeMode: ExecutionMode;
    activeRole: string;
    governanceMode: 'passive' | 'active' | 'strict';
    sendMissionChat: (message: string) => Promise<void>;
    clearMissionChat: () => void;
    setMissionChatScope: (scope: string | null) => void;
    setCouncilTarget: (id: string) => void;
    fetchCouncilMembers: () => Promise<void>;
    broadcastToSwarm: (message: string) => Promise<void>;
    confirmProposal: () => Promise<ConfirmProposalResult>;
    cancelProposal: () => void;
}

export interface CortexGovernanceOpsContract {
    streamLogs: StreamSignal[];
    isStreamConnected: boolean;
    streamConnectionState: 'idle' | 'connecting' | 'online' | 'offline';
    activeSquadRoomId: string | null;
    teamRoster: TeamDetail[];
    isFetchingTeamRoster: boolean;
    policyConfig: PolicyConfig | null;
    pendingApprovals: PendingApproval[];
    isFetchingPolicy: boolean;
    isFetchingApprovals: boolean;
    auditLog: AuditLogEntry[];
    isFetchingAuditLog: boolean;
    cognitiveStatus: CognitiveStatus | null;
    servicesStatus: ServiceHealthStatus[];
    isFetchingServicesStatus: boolean;
    servicesStatusUpdatedAt: string | null;
    teamsDetail: TeamDetailEntry[];
    isFetchingTeamsDetail: boolean;
    selectedTeamId: string | null;
    isTeamDrawerOpen: boolean;
    teamsFilter: TeamsFilter;
    selectedSignalDetail: SignalDetail | null;
    inspectedMessage: ChatMessage | null;
    isInspectorOpen: boolean;
    isStatusDrawerOpen: boolean;
    initializeStream: (force?: boolean) => void;
    disconnectStream: () => void;
    enterSquadRoom: (teamId: string) => void;
    exitSquadRoom: () => void;
    setStatusDrawerOpen: (open: boolean) => void;
    fetchTeamDetails: () => Promise<void>;
    fetchPolicy: () => Promise<void>;
    updatePolicy: (config: PolicyConfig) => Promise<void>;
    fetchPendingApprovals: () => Promise<void>;
    resolveApproval: (id: string, approved: boolean) => Promise<void>;
    fetchAuditLog: () => Promise<void>;
    fetchCognitiveStatus: () => Promise<void>;
    fetchServicesStatus: () => Promise<ServiceHealthStatus[]>;
    fetchTeamsDetail: () => Promise<void>;
    selectTeam: (teamId: string | null) => void;
    setTeamsFilter: (filter: TeamsFilter) => void;
    selectSignalDetail: (detail: SignalDetail | null) => void;
    setInspectedMessage: (msg: ChatMessage | null) => void;
}

export interface CortexAutomationRunsContract {
    runTimeline: MissionEvent[] | null;
    isFetchingTimeline: boolean;
    recentRuns: MissionRun[];
    isFetchingRuns: boolean;
    triggerRules: TriggerRule[];
    isFetchingTriggers: boolean;
    conversationTurns: ConversationTurn[] | null;
    isFetchingConversation: boolean;
    fetchRunTimeline: (runId: string) => Promise<void>;
    fetchRecentRuns: () => Promise<void>;
    fetchTriggerRules: () => Promise<void>;
    createTriggerRule: (r: TriggerRuleCreate) => Promise<TriggerRule | null>;
    updateTriggerRule: (id: string, r: TriggerRuleCreate) => Promise<void>;
    deleteTriggerRule: (id: string) => Promise<void>;
    toggleTriggerRule: (id: string, isActive: boolean) => Promise<void>;
    fetchRunConversation: (runId: string, agentFilter?: string) => Promise<void>;
    interjectInRun: (runId: string, message: string, agentId?: string) => Promise<void>;
}

export interface CortexProfilesSettingsContract {
    advancedMode: boolean;
    theme: 'aero-light' | 'midnight-cortex' | 'system';
    missionProfiles: MissionProfile[];
    activeProfileId: string | null;
    contextSnapshots: ContextSnapshot[];
    toggleAdvancedMode: () => void;
    fetchUserSettings: () => Promise<void>;
    updateAssistantName: (name: string) => Promise<boolean>;
    updateTheme: (theme: 'aero-light' | 'midnight-cortex' | 'system') => Promise<boolean>;
    fetchMissionProfiles: () => Promise<void>;
    createMissionProfile: (p: MissionProfileCreate) => Promise<MissionProfile | null>;
    updateMissionProfile: (id: string, p: MissionProfileCreate) => Promise<void>;
    deleteMissionProfile: (id: string) => Promise<void>;
    activateMissionProfile: (id: string) => Promise<void>;
    fetchContextSnapshots: () => Promise<void>;
    createContextSnapshot: (name: string) => Promise<ContextSnapshot | null>;
}

export interface CortexState
    extends CortexDraftGraphContract,
    CortexResourcesContract,
    CortexMissionChatContract,
    CortexGovernanceOpsContract,
    CortexAutomationRunsContract,
    CortexProfilesSettingsContract {}
