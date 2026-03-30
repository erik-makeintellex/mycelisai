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
    MCPServer,
    MCPServerWithTools,
    MCPLibraryCategory,
    MCPTool,
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

// Keep the runtime store focused on behavior; the full public state contract
// lives here so UI consumers and tests have one explicit source of truth.
export interface CortexState {
    chatHistory: ChatMessage[];
    nodes: import('reactflow').Node[];
    edges: import('reactflow').Edge[];
    isDrafting: boolean;
    isCommitting: boolean;
    error: string | null;
    blueprint: MissionBlueprint | null;
    missionStatus: MissionStatus;
    activeMissionId: string | null;
    streamLogs: StreamSignal[];
    isStreamConnected: boolean;
    streamConnectionState: 'idle' | 'connecting' | 'online' | 'offline';
    activeSquadRoomId: string | null;
    pendingArtifacts: CTSEnvelope[];
    selectedArtifact: CTSEnvelope | null;
    missions: import('@/store/cortexStoreTypes').Mission[];
    isFetchingMissions: boolean;
    trustThreshold: number;
    isSyncingThreshold: boolean;
    savedBlueprints: MissionBlueprint[];
    isBlueprintDrawerOpen: boolean;
    advancedMode: boolean;
    toggleAdvancedMode: () => void;
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
    mcpTools: MCPTool[];
    mcpLibrary: MCPLibraryCategory[];
    isFetchingMCPLibrary: boolean;
    workspaceChatScope: string | null;
    missionChat: ChatMessage[];
    isMissionChatting: boolean;
    missionChatError: string | null;
    missionChatFailure: MissionChatFailure | null;
    workspaceChatPrimed: boolean;
    assistantName: string;
    theme: 'aero-light' | 'midnight-cortex' | 'system';
    councilTarget: string;
    councilMembers: CouncilMember[];
    isBroadcasting: boolean;
    lastBroadcastResult: { teams_hit: number } | null;
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
    selectedAgentNodeId: string | null;
    isAgentEditorOpen: boolean;
    pendingProposal: ProposalData | null;
    activeConfirmToken: string | null;
    lastCommitProof: { intent_proof_id: string; audit_event_id: string } | null;
    activeRunId: string | null;
    runTimeline: MissionEvent[] | null;
    isFetchingTimeline: boolean;
    recentRuns: MissionRun[];
    isFetchingRuns: boolean;
    triggerRules: TriggerRule[];
    isFetchingTriggers: boolean;
    conversationTurns: ConversationTurn[] | null;
    isFetchingConversation: boolean;
    selectedSignalDetail: SignalDetail | null;
    activeBrain: BrainProvenance | null;
    activeMode: ExecutionMode;
    activeRole: string;
    governanceMode: 'passive' | 'active' | 'strict';
    inspectedMessage: ChatMessage | null;
    isInspectorOpen: boolean;
    isStatusDrawerOpen: boolean;
    onNodesChange: OnNodesChange;
    onEdgesChange: OnEdgesChange;
    submitIntent: (text: string) => Promise<void>;
    instantiateMission: () => Promise<void>;
    fetchMissions: () => Promise<void>;
    initializeStream: (force?: boolean) => void;
    disconnectStream: () => void;
    enterSquadRoom: (teamId: string) => void;
    exitSquadRoom: () => void;
    selectArtifact: (artifact: CTSEnvelope | null) => void;
    approveArtifact: (id: string) => void;
    rejectArtifact: (id: string, reason: string) => void;
    setTrustThreshold: (value: number) => void;
    fetchTrustThreshold: () => Promise<void>;
    toggleBlueprintDrawer: () => void;
    toggleToolsPalette: () => void;
    setStatusDrawerOpen: (open: boolean) => void;
    saveBlueprint: (bp: MissionBlueprint) => void;
    loadBlueprint: (bp: MissionBlueprint) => void;
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
    installMCPServer: (config: Partial<MCPServer>) => Promise<void>;
    deleteMCPServer: (id: string) => Promise<void>;
    fetchMCPTools: () => Promise<void>;
    fetchMCPLibrary: () => Promise<void>;
    installFromLibrary: (name: string, env?: Record<string, string>) => Promise<void>;
    sendMissionChat: (message: string) => Promise<void>;
    clearMissionChat: () => void;
    setMissionChatScope: (scope: string | null) => void;
    fetchUserSettings: () => Promise<void>;
    updateAssistantName: (name: string) => Promise<boolean>;
    updateTheme: (theme: 'aero-light' | 'midnight-cortex' | 'system') => Promise<boolean>;
    setCouncilTarget: (id: string) => void;
    fetchCouncilMembers: () => Promise<void>;
    broadcastToSwarm: (message: string) => Promise<void>;
    confirmProposal: () => Promise<ConfirmProposalResult>;
    cancelProposal: () => void;
    fetchRunTimeline: (runId: string) => Promise<void>;
    fetchRecentRuns: () => Promise<void>;
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
    selectAgentNode: (nodeId: string | null) => void;
    updateAgentInDraft: (teamIdx: number, agentIdx: number, updates: Partial<AgentManifest>) => void;
    deleteAgentFromDraft: (teamIdx: number, agentIdx: number) => void;
    discardDraft: () => void;
    updateAgentInMission: (agentName: string, manifest: Partial<AgentManifest>) => Promise<void>;
    deleteAgentFromMission: (agentName: string) => Promise<void>;
    deleteMission: (missionId: string) => Promise<void>;
    selectSignalDetail: (detail: SignalDetail | null) => void;
    setInspectedMessage: (msg: ChatMessage | null) => void;
    missionProfiles: MissionProfile[];
    activeProfileId: string | null;
    contextSnapshots: ContextSnapshot[];
    fetchMissionProfiles: () => Promise<void>;
    createMissionProfile: (p: MissionProfileCreate) => Promise<MissionProfile | null>;
    updateMissionProfile: (id: string, p: MissionProfileCreate) => Promise<void>;
    deleteMissionProfile: (id: string) => Promise<void>;
    activateMissionProfile: (id: string) => Promise<void>;
    fetchContextSnapshots: () => Promise<void>;
    createContextSnapshot: (name: string) => Promise<ContextSnapshot | null>;
    fetchTriggerRules: () => Promise<void>;
    createTriggerRule: (r: TriggerRuleCreate) => Promise<TriggerRule | null>;
    updateTriggerRule: (id: string, r: TriggerRuleCreate) => Promise<void>;
    deleteTriggerRule: (id: string) => Promise<void>;
    toggleTriggerRule: (id: string, isActive: boolean) => Promise<void>;
    fetchRunConversation: (runId: string, agentFilter?: string) => Promise<void>;
    interjectInRun: (runId: string, message: string, agentId?: string) => Promise<void>;
}
