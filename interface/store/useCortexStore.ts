import { create } from 'zustand';
import {
    type Node,
    type Edge,
    type OnNodesChange,
    type OnEdgesChange,
    applyNodeChanges,
    applyEdgeChanges,
    Position,
} from 'reactflow';
import type { AgentNodeData } from '@/components/wiring/AgentNode';
import type { ConversationTurn } from '@/types/conversations';

// ── Domain Types ──────────────────────────────────────────────

export interface ChatConsultation {
    member: string;
    summary: string;
}

export interface ChatArtifactRef {
    id?: string;              // artifact table ID (for stored artifacts)
    type: string;             // code | document | image | audio | data | chart | file
    title: string;
    content_type?: string;    // MIME type
    content?: string;         // inline content (text, JSON, base64 for images)
    url?: string;             // external URL (for links, images)
}

// CE-1: Orchestration Template types
export type TemplateID = 'chat-to-answer' | 'chat-to-proposal';
export type ExecutionMode = 'answer' | 'proposal';

export interface AnswerProvenance {
    resolved_intent: string;
    permission_check: string;
    policy_decision: string;
    audit_event_id: string;
    consult_chain?: string[];
}

export interface ProposalData {
    intent: string;
    teams: number;
    agents: number;
    tools: string[];
    risk_level: string;
    confirm_token: string;
    intent_proof_id: string;
}

// Mission Profiles — named provider-routing configs with optional reactive NATS watches
export interface MissionProfile {
    id: string;
    name: string;
    description?: string;
    role_providers: Record<string, string>;          // {"architect":"vllm","coder":"ollama"}
    subscriptions: { topic: string; condition?: string }[];
    context_strategy: 'fresh' | 'warm' | string;    // "fresh"|"warm"|"snapshot:<uuid>"
    auto_start: boolean;
    is_active: boolean;
    tenant_id: string;
    created_at: string;
    updated_at: string;
}

export interface MissionProfileCreate {
    name: string;
    description?: string;
    role_providers: Record<string, string>;
    subscriptions: { topic: string; condition?: string }[];
    context_strategy: string;
    auto_start: boolean;
}

export interface ContextSnapshot {
    id: string;
    name: string;
    description?: string;
    source_profile?: string;
    tenant_id: string;
    created_at: string;
}

// Phase 19: Brain provenance — which provider/model executed a response
export interface BrainProvenance {
    provider_id: string;
    provider_name?: string;
    model_id: string;
    location: 'local' | 'remote';
    data_boundary: 'local_only' | 'leaves_org';
    tokens_used?: number;
}

export interface ChatMessage {
    role: 'user' | 'architect' | 'admin' | 'council' | 'system';
    content: string;
    consultations?: ChatConsultation[];
    tools_used?: string[];
    source_node?: string;    // e.g. "council-architect"
    trust_score?: number;    // 0.0-1.0
    timestamp?: string;      // ISO 8601
    artifacts?: ChatArtifactRef[];
    // CE-1: Template metadata
    template_id?: TemplateID;
    mode?: ExecutionMode;
    provenance?: AnswerProvenance;
    proposal?: ProposalData;
    // Phase 19: Brain provenance
    brain?: BrainProvenance;
    // V7: run link for system messages
    run_id?: string;
}

export interface StreamSignal {
    type: string;
    source?: string;
    level?: string;
    message?: string;
    timestamp?: string;
    payload?: any;
    topic?: string;
}

export interface AgentManifest {
    id: string;
    role: string;
    system_prompt?: string;
    model?: string;
    inputs?: string[];
    outputs?: string[];
    tools?: string[];
}

interface BlueprintTeam {
    name: string;
    role: string;
    agents: AgentManifest[];
}

interface Constraint {
    constraint_id?: string;
    description: string;
}

export interface MissionBlueprint {
    mission_id: string;
    intent: string;
    teams: BlueprintTeam[];
    constraints?: Constraint[];
}

export type MissionStatus = 'idle' | 'draft' | 'active';

// ── Missions (Dashboard) ─────────────────────────────────────

export interface Mission {
    id: string;
    intent: string;
    status: 'active' | 'completed' | 'failed';
    teams: number;
    agents: number;
    created_at?: string;
}

// ── Deliverables / Governance ────────────────────────────────

export interface CTSEnvelope {
    id: string;
    source: string;
    signal: 'artifact' | 'governance_halt';
    timestamp: string;
    trust_score?: number;
    payload: {
        content: string;
        content_type: 'markdown' | 'json' | 'text' | 'image';
        title?: string;
    };
    proof?: {
        method: 'semantic' | 'empirical';
        logs: string;
        rubric_score: string;
        pass: boolean;
    };
}

// ── Sensory Periphery ────────────────────────────────────────

export interface SensorNode {
    id: string;
    type: string;
    status: 'online' | 'offline' | 'degraded';
    last_seen: string;
    label: string;
}

// ── Signal Detail (clickable event drawer) ──────────────────

export interface SignalDetail {
    type: string;
    source: string;
    level?: string;
    message: string;
    timestamp: string;
    topic?: string;
    payload?: any;
    id?: string;
    trace_id?: string;
    intent?: string;
    context?: Record<string, unknown>;
    trust_score?: number;
}

export interface LogEntry {
    id: string;
    trace_id: string;
    timestamp: string;
    level: string;
    source: string;
    intent: string;
    message: string;
    context: Record<string, unknown>;
}

// ── Council Chat API (Standardized CTS) ─────────────────────

export interface CouncilMember {
    id: string;
    role: string;
    team: string;
}

export interface APIResponse<T = unknown> {
    ok: boolean;
    data?: T;
    error?: string;
}

export interface CTSChatEnvelope {
    meta: { source_node: string; timestamp: string; trace_id?: string };
    signal_type: string;
    trust_score: number;
    template_id?: TemplateID;
    mode?: ExecutionMode;
    payload: {
        text: string;
        consultations?: ChatConsultation[];
        tools_used?: string[];
        artifacts?: ChatArtifactRef[];
        provenance?: AnswerProvenance;
        brain?: BrainProvenance;
        proposal?: {
            intent: string;
            tools: string[];
            risk_level: string;
            confirm_token: string;
            intent_proof_id: string;
        };
    };
}

// ── Team Manifestation Proposals ─────────────────────────────

export interface ProposedAgent {
    id: string;
    role: string;
    system_prompt?: string;
    model?: string;
}

export interface TeamProposal {
    id: string;
    name: string;
    role: string;
    agents: ProposedAgent[];
    reason: string;
    status: 'pending' | 'approved' | 'rejected';
    created_at: string;
}

// ── Agent Catalogue (Phase 7.5) ──────────────────────────────

export interface CatalogueAgent {
    id: string;
    name: string;
    role: string;
    system_prompt?: string;
    model?: string;
    tools: string[];
    inputs: string[];
    outputs: string[];
    verification_strategy?: string;
    verification_rubric: string[];
    validation_command?: string;
    created_at: string;
    updated_at: string;
}

// ── Artifacts (Phase 7.5) ────────────────────────────────────

export type ArtifactType = 'code' | 'document' | 'image' | 'audio' | 'data' | 'file' | 'chart';
export type ArtifactStatus = 'pending' | 'approved' | 'rejected' | 'archived';

export interface Artifact {
    id: string;
    mission_id?: string;
    team_id?: string;
    agent_id: string;
    trace_id?: string;
    artifact_type: ArtifactType;
    title: string;
    content_type: string;
    content?: string;
    file_path?: string;
    file_size_bytes?: number;
    metadata: Record<string, any>;
    trust_score?: number;
    status: ArtifactStatus;
    created_at: string;
}

export interface ArtifactFilters {
    mission_id?: string;
    team_id?: string;
    agent_id?: string;
    limit?: number;
}

// ── Team Explorer (Phase 7.6) ────────────────────────────────

export interface TeamAgent {
    id: string;
    name: string;
    team_id: string;
    status: number; // 0=offline, 1=idle, 2=busy, 3=error
    last_heartbeat: string;
}

export interface TeamDetail {
    id: string;
    name: string;
    role: string;
    agents: TeamAgent[];
}

// ── Team Management (Phase 11) ──────────────────────────────

export interface TeamDetailAgentEntry {
    id: string;
    role: string;
    status: number; // 0=offline, 1=idle, 2=busy, 3=error
    last_heartbeat: string;
    tools: string[];
    model: string;
    system_prompt?: string;
}

export interface TeamDetailEntry {
    id: string;
    name: string;
    role: string;
    type: 'standing' | 'mission';
    mission_id: string | null;
    mission_intent: string | null;
    inputs: string[];
    deliveries: string[];
    agents: TeamDetailAgentEntry[];
}

export type TeamsFilter = 'all' | 'standing' | 'mission';

// ── MCP Servers (Phase 7.5) ──────────────────────────────────

export interface MCPServer {
    id: string;
    name: string;
    transport: 'stdio' | 'sse';
    command?: string;
    args?: string[];
    env?: Record<string, string>;
    url?: string;
    headers?: Record<string, string>;
    status: string;
    error?: string;
    created_at: string;
}

export interface MCPTool {
    id: string;
    server_id: string;
    name: string;
    description?: string;
    input_schema: Record<string, any>;
}

export interface MCPServerWithTools extends MCPServer {
    tools: MCPTool[];
}

export interface MCPLibraryEntry {
    name: string;
    description: string;
    transport: string;
    command: string;
    args: string[];
    env?: Record<string, string>;
    tags: string[];
}

export interface MCPLibraryCategory {
    name: string;
    servers: MCPLibraryEntry[];
}

// ── V7: Mission Runs ─────────────────────────────────────────

export interface MissionRun {
    id: string;
    mission_id: string;
    tenant_id: string;
    status: 'pending' | 'running' | 'completed' | 'failed';
    run_depth: number;
    parent_run_id?: string;
    started_at: string;
    completed_at?: string;
    metadata?: Record<string, unknown>;
}

// ── V7: Mission Events (Run Timeline) ───────────────────────

export interface MissionEvent {
    id: string;
    run_id: string;
    tenant_id: string;
    event_type: string;
    severity: string;
    source_agent?: string;
    source_team?: string;
    payload?: Record<string, unknown>;
    audit_event_id?: string;
    emitted_at: string;
}

// ── V7: Trigger Rules (Team B) ───────────────────────────────

export interface TriggerRule {
    id: string;
    tenant_id: string;
    name: string;
    description?: string;
    event_pattern: string;
    condition: Record<string, unknown>;
    target_mission_id: string;
    mode: 'propose' | 'auto_execute';
    cooldown_seconds: number;
    max_depth: number;
    max_active_runs: number;
    is_active: boolean;
    last_fired_at?: string;
    created_at: string;
    updated_at: string;
}

export interface TriggerRuleCreate {
    name: string;
    description?: string;
    event_pattern: string;
    condition?: Record<string, unknown>;
    target_mission_id: string;
    mode?: 'propose' | 'auto_execute';
    cooldown_seconds?: number;
    max_depth?: number;
    max_active_runs?: number;
    is_active?: boolean;
}

export interface TriggerExecution {
    id: string;
    rule_id: string;
    event_id: string;
    run_id?: string;
    status: 'fired' | 'skipped' | 'proposed';
    skip_reason?: string;
    executed_at: string;
}

// ── Governance Policy (Phase 7.7) ────────────────────────────

export interface PolicyRule {
    intent: string;
    condition: string;
    action: 'ALLOW' | 'DENY' | 'REQUIRE_APPROVAL';
}

export interface PolicyGroup {
    name: string;
    description: string;
    targets: string[];
    rules: PolicyRule[];
}

export interface PolicyConfig {
    groups: PolicyGroup[];
    defaults: { default_action: string };
}

export interface PendingApproval {
    id: string;
    reason: string;
    source_agent: string;
    team_id: string;
    intent: string;
    timestamp: string;
    expires_at: string;
}

export interface CognitiveEngineStatus {
    status: "online" | "offline";
    endpoint?: string;
    model?: string;
}

export interface CognitiveStatus {
    text: CognitiveEngineStatus;
    media: CognitiveEngineStatus;
}

// ── Store Contract ────────────────────────────────────────────

export interface CortexState {
    chatHistory: ChatMessage[];
    nodes: Node[];
    edges: Edge[];
    isDrafting: boolean;
    isCommitting: boolean;
    error: string | null;
    blueprint: MissionBlueprint | null;
    missionStatus: MissionStatus;
    activeMissionId: string | null;

    // SSE stream state
    streamLogs: StreamSignal[];
    isStreamConnected: boolean;

    // Fractal navigation (Squad Room drill-down)
    activeSquadRoomId: string | null;

    // Governance / Deliverables
    pendingArtifacts: CTSEnvelope[];
    selectedArtifact: CTSEnvelope | null;

    // Missions (Dashboard)
    missions: Mission[];
    isFetchingMissions: boolean;

    // Trust Economy (Phase 5.2)
    trustThreshold: number;
    isSyncingThreshold: boolean;

    // Blueprint Library (Phase 5.2)
    savedBlueprints: MissionBlueprint[];
    isBlueprintDrawerOpen: boolean;

    // V7: Advanced Mode toggle — controls System tab + advanced surfaces visibility
    advancedMode: boolean;
    toggleAdvancedMode: () => void;

    // Tools Palette (Phase 7.7) — workspace tool browser
    isToolsPaletteOpen: boolean;

    // Sensor Library (Phase 5.3) — grouped subscriptions
    sensorFeeds: SensorNode[];
    isFetchingSensors: boolean;
    subscribedSensorGroups: string[];

    // Team Manifestation Proposals (Phase 5.3)
    teamProposals: TeamProposal[];
    isFetchingProposals: boolean;

    // Agent Catalogue (Phase 7.5)
    catalogueAgents: CatalogueAgent[];
    isFetchingCatalogue: boolean;
    selectedCatalogueAgent: CatalogueAgent | null;

    // Artifacts (Phase 7.5)
    artifacts: Artifact[];
    isFetchingArtifacts: boolean;
    selectedArtifactDetail: Artifact | null;

    // MCP Servers (Phase 7.5)
    mcpServers: MCPServerWithTools[];
    isFetchingMCPServers: boolean;
    mcpTools: MCPTool[];

    // MCP Library (Phase 7.7)
    mcpLibrary: MCPLibraryCategory[];
    isFetchingMCPLibrary: boolean;

    // Mission Control Chat (Phase 7.6) + Council API
    missionChat: ChatMessage[];
    isMissionChatting: boolean;
    missionChatError: string | null;
    councilTarget: string;              // active council member ID ("admin" default)
    councilMembers: CouncilMember[];    // populated from GET /council/members

    // Broadcast (Phase 8.0)
    isBroadcasting: boolean;
    lastBroadcastResult: { teams_hit: number } | null;

    // Team Explorer (Phase 7.6)
    teamRoster: TeamDetail[];
    isFetchingTeamRoster: boolean;

    // Governance (Phase 7.7)
    policyConfig: PolicyConfig | null;
    pendingApprovals: PendingApproval[];
    isFetchingPolicy: boolean;
    isFetchingApprovals: boolean;

    // Cognitive Engine Status (Phase 7.7 vLLM)
    cognitiveStatus: CognitiveStatus | null;

    // Team Management (Phase 11)
    teamsDetail: TeamDetailEntry[];
    isFetchingTeamsDetail: boolean;
    selectedTeamId: string | null;
    isTeamDrawerOpen: boolean;
    teamsFilter: TeamsFilter;

    // Wiring Edit/Delete (Phase 9)
    selectedAgentNodeId: string | null;
    isAgentEditorOpen: boolean;

    // CE-1: Orchestration Templates
    pendingProposal: ProposalData | null;
    activeConfirmToken: string | null;
    lastCommitProof: { intent_proof_id: string; audit_event_id: string } | null;

    // V7: Run Timeline
    activeRunId: string | null;
    runTimeline: MissionEvent[] | null;
    isFetchingTimeline: boolean;

    // V7: Recent Runs (dashboard widget)
    recentRuns: MissionRun[];
    isFetchingRuns: boolean;

    // V7: Trigger Rules (Team B)
    triggerRules: TriggerRule[];
    isFetchingTriggers: boolean;

    // V7: Conversation Log
    conversationTurns: ConversationTurn[] | null;
    isFetchingConversation: boolean;

    // Signal Detail Drawer
    selectedSignalDetail: SignalDetail | null;

    // Phase 19: Agent & Provider Orchestration
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
    initializeStream: () => void;
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

    // Agent Catalogue (Phase 7.5)
    fetchCatalogue: () => Promise<void>;
    createCatalogueAgent: (agent: Partial<CatalogueAgent>) => Promise<void>;
    updateCatalogueAgent: (id: string, agent: Partial<CatalogueAgent>) => Promise<void>;
    deleteCatalogueAgent: (id: string) => Promise<void>;
    selectCatalogueAgent: (agent: CatalogueAgent | null) => void;

    // Artifacts (Phase 7.5)
    fetchArtifacts: (filters?: ArtifactFilters) => Promise<void>;
    getArtifactDetail: (id: string) => Promise<void>;
    updateArtifactStatus: (id: string, status: string) => Promise<void>;

    // MCP Servers (Phase 7.5)
    fetchMCPServers: () => Promise<void>;
    installMCPServer: (config: Partial<MCPServer>) => Promise<void>;
    deleteMCPServer: (id: string) => Promise<void>;
    fetchMCPTools: () => Promise<void>;

    // MCP Library (Phase 7.7)
    fetchMCPLibrary: () => Promise<void>;
    installFromLibrary: (name: string, env?: Record<string, string>) => Promise<void>;

    // Mission Control Chat (Phase 7.6) + Council API
    sendMissionChat: (message: string) => Promise<void>;
    clearMissionChat: () => void;
    setCouncilTarget: (id: string) => void;
    fetchCouncilMembers: () => Promise<void>;

    // Broadcast (Phase 8.0)
    broadcastToSwarm: (message: string) => Promise<void>;

    // CE-1: Template-aware actions
    confirmProposal: () => Promise<void>;
    cancelProposal: () => void;
    fetchRunTimeline: (runId: string) => Promise<void>;
    fetchRecentRuns: () => Promise<void>;

    // Team Explorer (Phase 7.6)
    fetchTeamDetails: () => Promise<void>;

    // Governance (Phase 7.7)
    fetchPolicy: () => Promise<void>;
    updatePolicy: (config: PolicyConfig) => Promise<void>;
    fetchPendingApprovals: () => Promise<void>;
    resolveApproval: (id: string, approved: boolean) => Promise<void>;

    // Cognitive Engine Status (Phase 7.7 vLLM)
    fetchCognitiveStatus: () => Promise<void>;

    // Team Management (Phase 11)
    fetchTeamsDetail: () => Promise<void>;
    selectTeam: (teamId: string | null) => void;
    setTeamsFilter: (filter: TeamsFilter) => void;

    // Wiring Edit/Delete (Phase 9)
    selectAgentNode: (nodeId: string | null) => void;
    updateAgentInDraft: (teamIdx: number, agentIdx: number, updates: Partial<AgentManifest>) => void;
    deleteAgentFromDraft: (teamIdx: number, agentIdx: number) => void;
    discardDraft: () => void;
    updateAgentInMission: (agentName: string, manifest: Partial<AgentManifest>) => Promise<void>;
    deleteAgentFromMission: (agentName: string) => Promise<void>;
    deleteMission: (missionId: string) => Promise<void>;

    // Signal Detail Drawer
    selectSignalDetail: (detail: SignalDetail | null) => void;

    // Phase 19: Agent & Provider Orchestration
    setInspectedMessage: (msg: ChatMessage | null) => void;

    // Mission Profiles & Context Snapshots
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

    // V7 Trigger Rules (Team B)
    fetchTriggerRules: () => Promise<void>;
    createTriggerRule: (r: TriggerRuleCreate) => Promise<TriggerRule | null>;
    updateTriggerRule: (id: string, r: TriggerRuleCreate) => Promise<void>;
    deleteTriggerRule: (id: string) => Promise<void>;
    toggleTriggerRule: (id: string, isActive: boolean) => Promise<void>;

    // V7: Conversation Log
    fetchRunConversation: (runId: string, agentFilter?: string) => Promise<void>;
    interjectInRun: (runId: string, message: string, agentId?: string) => Promise<void>;
}

// ── Layout Constants ──────────────────────────────────────────

const TEAM_WIDTH = 280;
const TEAM_GAP = 60;
const AGENT_SPACING_Y = 130;
const AGENT_OFFSET_X = 60;
const TEAM_HEADER_Y = 80;

// ── Blueprint → ReactFlow Graph ──────────────────────────────

function blueprintToGraph(bp: MissionBlueprint): { nodes: Node[]; edges: Edge[] } {
    const nodes: Node[] = [];
    const edges: Edge[] = [];
    const outputMap = new Map<string, string>(); // topic → agentNodeId

    let teamX = 80;

    bp.teams.forEach((team, tIdx) => {
        const teamId = `team-${tIdx}`;
        const teamHeight = TEAM_HEADER_Y + team.agents.length * AGENT_SPACING_Y + 40;

        // Team group node
        nodes.push({
            id: teamId,
            type: 'group',
            position: { x: teamX, y: 40 },
            data: { label: '' },
            className: 'ghost-draft',
            style: {
                width: TEAM_WIDTH,
                height: teamHeight,
                background: 'rgba(30, 41, 59, 0.4)',
                border: '1px dashed rgba(6, 182, 212, 0.4)',
                borderRadius: '12px',
                padding: '8px',
            },
        });

        // Team label
        nodes.push({
            id: `${teamId}-label`,
            type: 'default',
            position: { x: 12, y: 8 },
            parentNode: teamId,
            extent: 'parent' as const,
            draggable: false,
            data: { label: team.name },
            style: {
                background: 'transparent',
                border: 'none',
                color: '#94a3b8',
                fontSize: '11px',
                fontWeight: 700,
                textTransform: 'uppercase' as const,
                letterSpacing: '0.1em',
                width: TEAM_WIDTH - 24,
                pointerEvents: 'none' as const,
            },
        });

        // Agent nodes
        team.agents.forEach((agent, aIdx) => {
            const agentId = `agent-${tIdx}-${aIdx}`;

            nodes.push({
                id: agentId,
                type: 'agentNode',
                position: {
                    x: AGENT_OFFSET_X,
                    y: TEAM_HEADER_Y + aIdx * AGENT_SPACING_Y,
                },
                parentNode: teamId,
                extent: 'parent' as const,
                className: 'ghost-draft',
                data: {
                    label: agent.id,
                    role: agent.role,
                    status: 'offline',
                    lastThought: agent.system_prompt
                        ? agent.system_prompt.slice(0, 60)
                        : undefined,
                    teamIdx: tIdx,
                    agentIdx: aIdx,
                } as AgentNodeData,
                sourcePosition: Position.Right,
                targetPosition: Position.Left,
            });

            agent.outputs?.forEach((topic) => {
                outputMap.set(topic, agentId);
            });
        });

        teamX += TEAM_WIDTH + TEAM_GAP;
    });

    // Wire edges by matching output→input topics
    bp.teams.forEach((team, tIdx) => {
        team.agents.forEach((agent, aIdx) => {
            const targetId = `agent-${tIdx}-${aIdx}`;
            agent.inputs?.forEach((topic) => {
                const sourceId = outputMap.get(topic);
                if (sourceId && sourceId !== targetId) {
                    edges.push({
                        id: `edge-${sourceId}-${targetId}-${topic}`,
                        source: sourceId,
                        target: targetId,
                        type: 'dataWire',
                        data: { type: 'output' },
                        animated: true,
                    });
                }
            });
        });
    });

    return { nodes, edges };
}

// ── Solidify: ghost-draft → active ───────────────────────────

function solidifyNodes(nodes: Node[]): Node[] {
    return nodes.map((node) => {
        if (node.className?.includes('ghost-draft')) {
            const solidNode = { ...node, className: '' };

            if (node.type === 'group') {
                solidNode.style = {
                    ...node.style,
                    border: '1px solid rgba(71, 85, 105, 0.6)',
                    boxShadow: '0 0 12px rgba(6, 182, 212, 0.15)',
                };
            }

            if (node.type === 'agentNode') {
                solidNode.data = {
                    ...node.data,
                    status: 'online',
                };
            }

            return solidNode;
        }
        return node;
    });
}

// ── SSE Connection (module-level ref) ─────────────────────────

let _eventSource: EventSource | null = null;

/** Dispatch an SSE signal to matching ReactFlow nodes */
function dispatchSignalToNodes(
    signal: StreamSignal,
    nodes: Node[],
): Node[] | null {
    const src = signal.source;
    if (!src) return null;

    let changed = false;
    const updated = nodes.map((node) => {
        // Match by node ID or agent label
        if (node.id !== src && node.data?.label !== src) return node;
        changed = true;

        if (signal.type === 'thought' || signal.type === 'cognitive') {
            return {
                ...node,
                data: { ...node.data, isThinking: true, lastThought: signal.message },
            };
        }
        if (signal.type === 'artifact' || signal.type === 'output') {
            return {
                ...node,
                data: {
                    ...node.data,
                    isThinking: false,
                    lastThought: signal.message ?? node.data.lastThought,
                },
            };
        }
        if (signal.type === 'error') {
            return {
                ...node,
                data: {
                    ...node.data,
                    status: 'error',
                    isThinking: false,
                    lastThought: signal.message,
                },
            };
        }
        return node;
    });

    return changed ? updated : null;
}

// ── Chat Persistence (localStorage) ──────────────────────────
// Soma's memory: chat survives page refreshes. Use clearMissionChat to reset.

const CHAT_STORAGE_KEY = 'mycelis-workspace-chat';
const CHAT_STORAGE_KEY_LEGACY = 'mycelis-mission-chat'; // migrate old key
const CHAT_MAX_PERSISTED = 200; // cap to avoid localStorage quota issues

function loadPersistedChat(): ChatMessage[] {
    if (typeof window === 'undefined') return [];
    try {
        // Try new key first, fall back to legacy key on first load (migration)
        const raw = localStorage.getItem(CHAT_STORAGE_KEY)
            ?? localStorage.getItem(CHAT_STORAGE_KEY_LEGACY);
        if (!raw) return [];
        const msgs: ChatMessage[] = JSON.parse(raw);
        return Array.isArray(msgs) ? msgs.slice(-CHAT_MAX_PERSISTED) : [];
    } catch {
        return [];
    }
}

function persistChat(messages: ChatMessage[]) {
    if (typeof window === 'undefined') return;
    try {
        // Only persist the last N messages to respect localStorage limits
        const toStore = messages.slice(-CHAT_MAX_PERSISTED);
        localStorage.setItem(CHAT_STORAGE_KEY, JSON.stringify(toStore));
    } catch { /* quota exceeded — silently drop */ }
}

// ── Zustand Store ─────────────────────────────────────────────

export const useCortexStore = create<CortexState>((set, get) => ({
    chatHistory: [],
    nodes: [],
    edges: [],
    isDrafting: false,
    isCommitting: false,
    error: null,
    blueprint: null,
    missionStatus: 'idle',
    activeMissionId: null,
    streamLogs: [],
    isStreamConnected: false,
    activeSquadRoomId: null,
    pendingArtifacts: [],
    selectedArtifact: null,
    missions: [],
    isFetchingMissions: false,
    trustThreshold: 0.7,
    isSyncingThreshold: false,
    savedBlueprints: [],
    isBlueprintDrawerOpen: false,
    advancedMode: typeof window !== 'undefined' && typeof localStorage !== 'undefined' && typeof localStorage.getItem === 'function' ? localStorage.getItem('mycelis-advanced-mode') === 'true' : false,
    isToolsPaletteOpen: false,
    sensorFeeds: [],
    isFetchingSensors: false,
    subscribedSensorGroups: [],
    teamProposals: [],
    isFetchingProposals: false,

    // Agent Catalogue (Phase 7.5)
    catalogueAgents: [],
    isFetchingCatalogue: false,
    selectedCatalogueAgent: null,

    // Artifacts (Phase 7.5)
    artifacts: [],
    isFetchingArtifacts: false,
    selectedArtifactDetail: null,

    // MCP Servers (Phase 7.5)
    mcpServers: [],
    isFetchingMCPServers: false,
    mcpTools: [],
    mcpLibrary: [],
    isFetchingMCPLibrary: false,

    // Mission Control Chat (Phase 7.6) — rehydrated from localStorage
    missionChat: loadPersistedChat(),
    isMissionChatting: false,
    missionChatError: null,
    councilTarget: 'admin',
    councilMembers: [],

    // Broadcast (Phase 8.0)
    isBroadcasting: false,
    lastBroadcastResult: null,

    // Team Explorer (Phase 7.6)
    teamRoster: [],
    isFetchingTeamRoster: false,

    // Governance (Phase 7.7)
    policyConfig: null,
    pendingApprovals: [],
    isFetchingPolicy: false,
    isFetchingApprovals: false,
    cognitiveStatus: null,

    // Team Management (Phase 11)
    teamsDetail: [],
    isFetchingTeamsDetail: false,
    selectedTeamId: null,
    isTeamDrawerOpen: false,
    teamsFilter: 'all' as TeamsFilter,

    // Wiring Edit/Delete (Phase 9)
    selectedAgentNodeId: null,
    isAgentEditorOpen: false,

    // CE-1: Orchestration Templates
    pendingProposal: null,
    activeConfirmToken: null,
    lastCommitProof: null,

    // V7: Run Timeline
    activeRunId: null,
    runTimeline: null,
    isFetchingTimeline: false,

    // V7: Recent Runs
    recentRuns: [],
    isFetchingRuns: false,

    // V7: Trigger Rules (Team B)
    triggerRules: [],
    isFetchingTriggers: false,

    // V7: Conversation Log
    conversationTurns: null,
    isFetchingConversation: false,

    // Signal Detail Drawer
    selectedSignalDetail: null,

    // Phase 19: Agent & Provider Orchestration
    activeBrain: null,
    activeMode: 'answer' as ExecutionMode,
    activeRole: '',
    governanceMode: 'passive' as const,
    inspectedMessage: null,
    isInspectorOpen: false,
    isStatusDrawerOpen: false,

    // Mission Profiles & Context Snapshots
    missionProfiles: [],
    activeProfileId: null,
    contextSnapshots: [],

    onNodesChange: (changes) => {
        set({ nodes: applyNodeChanges(changes, get().nodes) });
    },

    onEdgesChange: (changes) => {
        set({ edges: applyEdgeChanges(changes, get().edges) });
    },

    submitIntent: async (text: string) => {
        const trimmed = text.trim();
        if (!trimmed) return;

        // Append user message
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

            // CE-1: Negotiate response is now { blueprint, intent_proof, confirm_token, template }
            const bp = (data.blueprint ?? data) as MissionBlueprint;
            const { nodes, edges } = blueprintToGraph(bp);

            const agentCount = bp.teams.reduce((sum, t) => sum + t.agents.length, 0);
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

            // CE-1: Extract proposal data from enriched response
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

    initializeStream: () => {
        if (_eventSource) return; // Already connected

        const es = new EventSource('/api/v1/stream');

        es.onopen = () => {
            set({ isStreamConnected: true });
        };

        es.onmessage = (event) => {
            try {
                const signal: StreamSignal = JSON.parse(event.data);
                const { nodes } = get();

                // Push to stream log (capped at 100)
                const nextLogs = [signal, ...get().streamLogs].slice(0, 100);

                // Dispatch node-specific updates
                const updatedNodes = dispatchSignalToNodes(signal, nodes);

                const patch: Partial<CortexState> = updatedNodes
                    ? { streamLogs: nextLogs, nodes: updatedNodes }
                    : { streamLogs: nextLogs };

                // Intercept artifact signals → push to pending deliverables
                if (signal.type === 'artifact' && signal.source) {
                    const envelope: CTSEnvelope = {
                        id: `${signal.source}-${signal.timestamp ?? Date.now()}`,
                        source: signal.source,
                        signal: 'artifact',
                        timestamp: signal.timestamp ?? new Date().toISOString(),
                        trust_score: signal.payload?.trust_score,
                        payload: {
                            content: signal.message ?? JSON.stringify(signal.payload ?? {}),
                            content_type: signal.payload?.content_type ?? 'text',
                            title: signal.payload?.title,
                        },
                        proof: signal.payload?.proof,
                    };
                    patch.pendingArtifacts = [envelope, ...get().pendingArtifacts];
                }

                // Phase 5.2: Intercept governance_halt signals from Overseer
                // Low-trust envelopes halted by the Governance Valve
                if (signal.type === 'governance_halt' && signal.source) {
                    const envelope: CTSEnvelope = {
                        id: `gov-${signal.source}-${signal.timestamp ?? Date.now()}`,
                        source: signal.source,
                        signal: 'governance_halt',
                        timestamp: signal.timestamp ?? new Date().toISOString(),
                        trust_score: signal.payload?.trust_score ?? (signal as any).trust_score,
                        payload: {
                            content: `Trust score below threshold. Awaiting human approval.`,
                            content_type: 'text',
                            title: `Governance Halt: ${signal.source}`,
                        },
                    };
                    patch.pendingArtifacts = [envelope, ...get().pendingArtifacts];
                }

                set(patch);
            } catch (e) {
                console.error('Stream parse error', e);
            }
        };

        es.onerror = () => {
            set({ isStreamConnected: false });
            _eventSource = null;
            es.close();
        };

        _eventSource = es;
    },

    disconnectStream: () => {
        if (_eventSource) {
            _eventSource.close();
            _eventSource = null;
            set({ isStreamConnected: false });
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

    // Phase 19: Agent & Provider Orchestration
    setInspectedMessage: (msg: ChatMessage | null) => {
        set({ inspectedMessage: msg, isInspectorOpen: msg !== null });
    },

    approveArtifact: (id: string) => {
        const artifact = get().pendingArtifacts.find((a) => a.id === id);
        if (artifact) {
            console.log('[GOVERNANCE] APPROVED:', id, artifact);
        }
        set((s) => ({
            pendingArtifacts: s.pendingArtifacts.filter((a) => a.id !== id),
            selectedArtifact: s.selectedArtifact?.id === id ? null : s.selectedArtifact,
        }));
    },

    rejectArtifact: (id: string, reason: string) => {
        const artifact = get().pendingArtifacts.find((a) => a.id === id);
        if (artifact) {
            console.log('[GOVERNANCE] REJECTED:', id, reason, artifact);
        }
        set((s) => ({
            pendingArtifacts: s.pendingArtifacts.filter((a) => a.id !== id),
            selectedArtifact: s.selectedArtifact?.id === id ? null : s.selectedArtifact,
        }));
    },

    // ── Trust Economy (Phase 5.2) ──────────────────────────────

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
        } catch { /* degraded mode — use local default */ }
    },

    // ── Blueprint Library (Phase 5.2) ────────────────────────

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
            savedBlueprints: [bp, ...s.savedBlueprints.filter((b) => b.mission_id !== bp.mission_id)],
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

    // ── Sensory Periphery (Phase 5.3) ──────────────────────────

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
                ? current.filter((g) => g !== group)
                : [...current, group];
            return { subscribedSensorGroups: next };
        });
    },

    // ── Team Manifestation Proposals (Phase 5.3) ────────────

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
                    teamProposals: s.teamProposals.map((p) =>
                        p.id === id ? { ...p, status: 'approved' as const } : p
                    ),
                }));
            }
        } catch (err) {
            console.error('[PROPOSALS] Approve failed:', err);
        }
    },

    rejectProposal: async (id: string) => {
        try {
            const res = await fetch(`/api/v1/proposals/${id}/reject`, { method: 'POST' });
            if (res.ok) {
                set((s) => ({
                    teamProposals: s.teamProposals.map((p) =>
                        p.id === id ? { ...p, status: 'rejected' as const } : p
                    ),
                }));
            }
        } catch (err) {
            console.error('[PROPOSALS] Reject failed:', err);
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

    // ── Agent Catalogue (Phase 7.5) ──────────────────────────────

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
                    selectedCatalogueAgent: null,
                }));
            }
        } catch (err) {
            console.error('[CATALOGUE] Create failed:', err);
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
                    catalogueAgents: s.catalogueAgents.map((a) => (a.id === id ? updated : a)),
                    selectedCatalogueAgent: null,
                }));
            }
        } catch (err) {
            console.error('[CATALOGUE] Update failed:', err);
        }
    },

    deleteCatalogueAgent: async (id: string) => {
        try {
            const res = await fetch(`/api/v1/catalogue/agents/${id}`, { method: 'DELETE' });
            if (res.ok) {
                set((s) => ({
                    catalogueAgents: s.catalogueAgents.filter((a) => a.id !== id),
                    selectedCatalogueAgent: s.selectedCatalogueAgent?.id === id ? null : s.selectedCatalogueAgent,
                }));
            }
        } catch (err) {
            console.error('[CATALOGUE] Delete failed:', err);
        }
    },

    selectCatalogueAgent: (agent: CatalogueAgent | null) => {
        set({ selectedCatalogueAgent: agent });
    },

    // ── Artifacts (Phase 7.5) ────────────────────────────────────

    fetchArtifacts: async (filters?: ArtifactFilters) => {
        set({ isFetchingArtifacts: true });
        try {
            const params = new URLSearchParams();
            if (filters?.mission_id) params.set('mission_id', filters.mission_id);
            if (filters?.team_id) params.set('team_id', filters.team_id);
            if (filters?.agent_id) params.set('agent_id', filters.agent_id);
            if (filters?.limit) params.set('limit', String(filters.limit));
            const qs = params.toString();
            const res = await fetch(`/api/v1/artifacts${qs ? `?${qs}` : ''}`);
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
        } catch (err) {
            console.error('[ARTIFACTS] Get detail failed:', err);
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
                    artifacts: s.artifacts.map((a) => (a.id === id ? { ...a, status: status as ArtifactStatus } : a)),
                    selectedArtifactDetail: s.selectedArtifactDetail?.id === id
                        ? { ...s.selectedArtifactDetail, status: status as ArtifactStatus }
                        : s.selectedArtifactDetail,
                }));
            }
        } catch (err) {
            console.error('[ARTIFACTS] Status update failed:', err);
        }
    },

    // ── MCP Servers (Phase 7.5) ──────────────────────────────────

    fetchMCPServers: async () => {
        set({ isFetchingMCPServers: true });
        try {
            const res = await fetch('/api/v1/mcp/servers');
            if (res.ok) {
                const data = await res.json();
                set({ mcpServers: Array.isArray(data) ? data : [], isFetchingMCPServers: false });
            } else {
                set({ mcpServers: [], isFetchingMCPServers: false });
            }
        } catch {
            set({ mcpServers: [], isFetchingMCPServers: false });
        }
    },

    installMCPServer: async (config: Partial<MCPServer>) => {
        try {
            const res = await fetch('/api/v1/mcp/install', {
                method: 'POST',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify(config),
            });
            if (res.ok) {
                // Re-fetch to get full server state with tools
                get().fetchMCPServers();
            }
        } catch (err) {
            console.error('[MCP] Install failed:', err);
        }
    },

    deleteMCPServer: async (id: string) => {
        try {
            const res = await fetch(`/api/v1/mcp/servers/${id}`, { method: 'DELETE' });
            if (res.ok) {
                set((s) => ({
                    mcpServers: s.mcpServers.filter((srv) => srv.id !== id),
                }));
            }
        } catch (err) {
            console.error('[MCP] Delete failed:', err);
        }
    },

    fetchMCPTools: async () => {
        try {
            const res = await fetch('/api/v1/mcp/tools');
            if (res.ok) {
                const data = await res.json();
                set({ mcpTools: Array.isArray(data) ? data : [] });
            }
        } catch {
            set({ mcpTools: [] });
        }
    },

    // ── MCP Library (Phase 7.7) ─────────────────────────────────

    fetchMCPLibrary: async () => {
        set({ isFetchingMCPLibrary: true });
        try {
            const res = await fetch('/api/v1/mcp/library');
            if (res.ok) {
                const data = await res.json();
                set({ mcpLibrary: Array.isArray(data) ? data : [], isFetchingMCPLibrary: false });
            } else {
                set({ mcpLibrary: [], isFetchingMCPLibrary: false });
            }
        } catch {
            set({ mcpLibrary: [], isFetchingMCPLibrary: false });
        }
    },

    installFromLibrary: async (name: string, env?: Record<string, string>) => {
        try {
            const res = await fetch('/api/v1/mcp/library/install', {
                method: 'POST',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify({ name, env }),
            });
            if (res.ok) {
                // Refresh installed servers list
                get().fetchMCPServers();
            } else {
                console.error('[MCP Library] Install failed:', await res.text());
            }
        } catch (err) {
            console.error('[MCP Library] Install failed:', err);
        }
    },

    // ── Mission Control Chat + Council API ───────────────────────

    setCouncilTarget: (id: string) => {
        set({ councilTarget: id });
    },

    fetchCouncilMembers: async () => {
        try {
            const res = await fetch('/api/v1/council/members');
            if (!res.ok) return;
            const body: APIResponse<CouncilMember[]> = await res.json();
            if (body.ok && body.data) {
                set({ councilMembers: body.data });
            }
        } catch {
            // Swarm offline — leave members empty, selector falls back to "Admin"
        }
    },

    sendMissionChat: async (message: string) => {
        const trimmed = message.trim();
        if (!trimmed) return;

        const { councilTarget } = get();

        // Append user message and set loading
        set((s) => ({
            missionChat: [...s.missionChat, { role: 'user', content: trimmed }],
            isMissionChatting: true,
            missionChatError: null,
        }));

        try {
            // Smart windowing: send only the last 20 messages to the LLM.
            // Older context is auto-recalled from pgvector by the backend (Phase 19).
            const messages = [...get().missionChat]
                .slice(-20)
                .map((m) => ({
                    role: m.role === 'user' ? 'user' : 'assistant',
                    content: m.content,
                }));

            const res = await fetch(`/api/v1/council/${councilTarget}/chat`, {
                method: 'POST',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify({ messages }),
            });

            if (!res.ok) {
                const text = await res.text();
                let errMsg: string;
                try {
                    const parsed = JSON.parse(text);
                    errMsg = parsed.error || `Council agent error (${res.status})`;
                } catch {
                    errMsg = `Council agent unreachable (${res.status})`;
                }
                set((s) => ({
                    isMissionChatting: false,
                    missionChatError: errMsg,
                    missionChat: [...s.missionChat, { role: 'council', content: errMsg, source_node: councilTarget }],
                }));
                return;
            }

            const body: APIResponse<CTSChatEnvelope> = await res.json();

            if (!body.ok || !body.data) {
                const errText = body.error || `Chat request failed (${res.status})`;
                set((s) => ({
                    isMissionChatting: false,
                    missionChatError: errText,
                    missionChat: [...s.missionChat, { role: 'council', content: errText, source_node: councilTarget }],
                }));
                return;
            }

            const envelope = body.data;
            const chatMsg: ChatMessage = {
                role: 'council',
                content: envelope.payload.text,
                consultations: envelope.payload.consultations,
                tools_used: envelope.payload.tools_used,
                source_node: envelope.meta.source_node,
                trust_score: envelope.trust_score,
                timestamp: envelope.meta.timestamp,
                artifacts: envelope.payload.artifacts,
                // CE-1: Template metadata
                template_id: envelope.template_id || 'chat-to-answer',
                mode: envelope.mode || 'answer',
                provenance: envelope.payload?.provenance,
                // Phase 19: Brain provenance
                brain: envelope.payload?.brain,
                // Phase 19-B: Extract proposal from chat response
                proposal: envelope.payload?.proposal ? {
                    intent: envelope.payload.proposal.intent,
                    teams: 0,
                    agents: 0,
                    tools: envelope.payload.proposal.tools || [],
                    risk_level: envelope.payload.proposal.risk_level || 'medium',
                    confirm_token: envelope.payload.proposal.confirm_token,
                    intent_proof_id: envelope.payload.proposal.intent_proof_id,
                } : undefined,
            };

            // Phase 19: Derive governance mode from trust threshold
            const trust = get().trustThreshold;
            const govMode = trust >= 0.8 ? 'strict' as const : trust >= 0.5 ? 'active' as const : 'passive' as const;

            set((s) => ({
                isMissionChatting: false,
                missionChat: [...s.missionChat, chatMsg],
                // Phase 19: Update active orchestration state
                activeBrain: chatMsg.brain ?? null,
                activeMode: chatMsg.mode || 'answer',
                activeRole: chatMsg.source_node || '',
                governanceMode: govMode,
                // Phase 19-B: Set pending proposal if present
                ...(chatMsg.proposal ? {
                    pendingProposal: chatMsg.proposal,
                    activeConfirmToken: chatMsg.proposal.confirm_token,
                } : {}),
            }));
        } catch (err) {
            const msg = err instanceof Error ? err.message : 'Chat failed';
            set((s) => ({
                isMissionChatting: false,
                missionChatError: msg,
                missionChat: [...s.missionChat, { role: 'council', content: `Error: ${msg}`, source_node: councilTarget }],
            }));
        }
    },

    clearMissionChat: () => {
        set({ missionChat: [], missionChatError: null });
        if (typeof window !== 'undefined') {
            localStorage.removeItem(CHAT_STORAGE_KEY);
        }
    },

    // ── Broadcast (Phase 8.0) ─────────────────────────────────────

    broadcastToSwarm: async (message: string) => {
        const trimmed = message.trim();
        if (!trimmed) return;

        set((s) => ({
            missionChat: [...s.missionChat, { role: 'user', content: `[BROADCAST] ${trimmed}` }],
            isBroadcasting: true,
            missionChatError: null,
        }));

        try {
            const res = await fetch('/api/v1/swarm/broadcast', {
                method: 'POST',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify({ content: trimmed, source: 'mission-control' }),
            });

            if (!res.ok) {
                const errText = `Broadcast failed (${res.status})`;
                set((s) => ({
                    isBroadcasting: false,
                    missionChatError: errText,
                    missionChat: [...s.missionChat, { role: 'architect', content: errText }],
                }));
                return;
            }

            const data = await res.json();
            // Build a chat message per team reply
            const replyMessages: ChatMessage[] = [];
            if (Array.isArray(data.replies)) {
                for (const reply of data.replies) {
                    if (reply.error) {
                        replyMessages.push({
                            role: 'architect',
                            content: `**${reply.team_id}**: _timed out or unavailable_`,
                            source_node: reply.team_id,
                        });
                    } else if (reply.content) {
                        replyMessages.push({
                            role: 'council',
                            content: reply.content,
                            source_node: reply.team_id,
                        });
                    }
                }
            }
            // Fallback if no replies came back
            if (replyMessages.length === 0) {
                replyMessages.push({
                    role: 'architect',
                    content: `Broadcast sent to ${data.teams_hit} team(s) — no replies received.`,
                });
            }
            set((s) => ({
                isBroadcasting: false,
                lastBroadcastResult: { teams_hit: data.teams_hit },
                missionChat: [...s.missionChat, ...replyMessages],
            }));
        } catch (err) {
            const msg = err instanceof Error ? err.message : 'Broadcast failed';
            set((s) => ({
                isBroadcasting: false,
                missionChatError: msg,
                missionChat: [...s.missionChat, { role: 'architect', content: `Error: ${msg}` }],
            }));
        }
    },

    // ── Team Explorer (Phase 7.6) ────────────────────────────────

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

            // Merge agents into their teams
            const roster: TeamDetail[] = (Array.isArray(teams) ? teams : []).map((t: any) => ({
                id: t.id,
                name: t.name,
                role: t.role || 'observer',
                agents: agents.filter((a) => a.team_id === t.id),
            }));

            set({ teamRoster: roster, isFetchingTeamRoster: false });
        } catch {
            set({ teamRoster: [], isFetchingTeamRoster: false });
        }
    },

    // ── Governance (Phase 7.7) ────────────────────────────────────

    fetchPolicy: async () => {
        set({ isFetchingPolicy: true });
        try {
            const res = await fetch('/api/v1/governance/policy');
            if (res.ok) {
                const data = await res.json();
                set({ policyConfig: data, isFetchingPolicy: false });
            } else {
                set({ isFetchingPolicy: false });
            }
        } catch { set({ isFetchingPolicy: false }); }
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
        } catch (err) { console.error('[Governance] Update failed:', err); }
    },

    fetchPendingApprovals: async () => {
        set({ isFetchingApprovals: true });
        try {
            const res = await fetch('/api/v1/governance/pending');
            if (res.ok) {
                const data = await res.json();
                set({ pendingApprovals: Array.isArray(data) ? data : [], isFetchingApprovals: false });
            } else {
                set({ pendingApprovals: [], isFetchingApprovals: false });
            }
        } catch { set({ pendingApprovals: [], isFetchingApprovals: false }); }
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
                    pendingApprovals: s.pendingApprovals.filter((a) => a.id !== id),
                }));
            }
        } catch (err) { console.error('[Governance] Resolve failed:', err); }
    },

    fetchCognitiveStatus: async () => {
        try {
            const res = await fetch('/api/v1/cognitive/status');
            if (res.ok) {
                const data = await res.json();
                set({ cognitiveStatus: data });
            }
        } catch { /* silently fail — dashboard gauge will show offline */ }
    },

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

            // Solidify: ghost → active
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

    // ── Team Management (Phase 11) ──────────────────────────────

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

    // ── Wiring Edit/Delete (Phase 9) ──────────────────────────

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
        console.log('[DEBUG] deleteAgentFromDraft called', { teamIdx, agentIdx });
        const bp = get().blueprint;
        if (!bp) {
            console.error('[DEBUG] No blueprint found');
            return;
        }
        const newBp: MissionBlueprint = structuredClone(bp);
        console.log('[DEBUG] Blueprint cloned', newBp);
        const team = newBp.teams[teamIdx];
        if (!team) {
            console.error('[DEBUG] Team not found', teamIdx);
            return;
        }
        team.agents.splice(agentIdx, 1);
        console.log('[DEBUG] Agent spliced. Remaining in team:', team.agents.length);
        // Remove team if empty
        if (team.agents.length === 0) {
            newBp.teams.splice(teamIdx, 1);
            console.log('[DEBUG] Team removed (empty). Remaining teams:', newBp.teams.length);
        }
        const { nodes, edges } = blueprintToGraph(newBp);
        console.log('[DEBUG] Graph regenerated', { nodes: nodes.length, edges: edges.length });
        set({ blueprint: newBp, nodes, edges, selectedAgentNodeId: null, isAgentEditorOpen: false });
        console.log('[DEBUG] State updated');
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

            // Update local blueprint to match
            if (blueprint) {
                const newBp: MissionBlueprint = structuredClone(blueprint);
                for (const team of newBp.teams) {
                    const agent = team.agents.find((a) => a.id === agentName);
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
        console.log('[DEBUG] deleteAgentFromMission called', { agentName, activeMissionId });
        if (!activeMissionId) return;

        try {
            const res = await fetch(`/api/v1/missions/${activeMissionId}/agents/${agentName}`, {
                method: 'DELETE',
            });
            console.log('[DEBUG] DELETE API response:', res.status);
            if (!res.ok) {
                const data = await res.json().catch(() => ({}));
                const errMsg = data.error ?? res.statusText;
                console.error('deleteAgentFromMission error:', errMsg);

                const div = document.createElement('div');
                div.id = 'debug-result-error';
                div.innerText = 'DELETE ERROR: ' + errMsg;
                div.style.cssText = 'position:fixed;top:50%;left:50%;transform:translate(-50%,-50%);z-index:9999;background:red;color:white;font-size:32px;padding:20px;border:4px solid black;';
                document.body.appendChild(div);
                return;
            }

            const div = document.createElement('div');
            div.id = 'debug-result-success';
            div.innerText = 'DELETE SUCCESS: ' + agentName;
            div.style.cssText = 'position:fixed;top:50%;left:50%;transform:translate(-50%,-50%);z-index:9999;background:green;color:white;font-size:32px;padding:20px;border:4px solid black;';
            document.body.appendChild(div);

            // Update local blueprint
            if (blueprint) {
                const newBp: MissionBlueprint = structuredClone(blueprint);
                let spliced = false;
                for (let tIdx = 0; tIdx < newBp.teams.length; tIdx++) {
                    const aIdx = newBp.teams[tIdx].agents.findIndex((a) => a.id === agentName);
                    if (aIdx !== -1) {
                        console.log('[DEBUG] Splicing active agent at', { tIdx, aIdx });
                        newBp.teams[tIdx].agents.splice(aIdx, 1);
                        spliced = true;
                        if (newBp.teams[tIdx].agents.length === 0) {
                            newBp.teams.splice(tIdx, 1);
                        }
                        break;
                    }
                }
                if (!spliced) console.warn('[DEBUG] Active agent NOT found in blueprint:', agentName);

                const { nodes, edges } = blueprintToGraph(newBp);
                set({ blueprint: newBp, nodes: solidifyNodes(nodes), edges, selectedAgentNodeId: null, isAgentEditorOpen: false });
                console.log('[DEBUG] Active state updated');
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

            // Clear canvas
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

    // CE-1: Proposal confirmation — calls backend confirm-action endpoint
    confirmProposal: async () => {
        const { activeConfirmToken, pendingProposal } = get();
        if (!activeConfirmToken || !pendingProposal) return;
        try {
            const res = await fetch('/api/v1/intent/confirm-action', {
                method: 'POST',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify({ confirm_token: activeConfirmToken }),
            });
            if (res.ok) {
                const body = await res.json();
                const runId: string | null = body?.data?.run_id ?? body?.run_id ?? null;
                const systemMsg: ChatMessage = {
                    role: 'system',
                    content: 'Mission activated',
                    run_id: runId ?? undefined,
                    timestamp: new Date().toISOString(),
                };
                set((s) => ({
                    activeRunId: runId,
                    missionChat: [...s.missionChat, systemMsg],
                }));
            } else {
                const text = await res.text();
                console.error('[CE-1] Confirm action failed:', text);
            }
        } catch (err) {
            console.error('[CE-1] confirmProposal error:', err);
        }
        set({ pendingProposal: null, activeConfirmToken: null });
    },

    cancelProposal: () => {
        set({ pendingProposal: null, activeConfirmToken: null });
    },

    // V7: Fetch run timeline from event store
    fetchRunTimeline: async (runId: string) => {
        set({ isFetchingTimeline: true });
        try {
            const res = await fetch(`/api/v1/runs/${runId}/events`);
            if (!res.ok) return;
            const body = await res.json();
            set({ runTimeline: body.data ?? body ?? [] });
        } catch { /* offline — silent */ }
        finally { set({ isFetchingTimeline: false }); }
    },

    // V7: Fetch recent runs for the dashboard widget
    fetchRecentRuns: async () => {
        set({ isFetchingRuns: true });
        try {
            const res = await fetch('/api/v1/runs');
            if (!res.ok) return;
            const body = await res.json();
            set({ recentRuns: body.data ?? body ?? [] });
        } catch { /* offline — silent */ }
        finally { set({ isFetchingRuns: false }); }
    },

    // ── Mission Profiles ─────────────────────────────────────────

    fetchMissionProfiles: async () => {
        try {
            const res = await fetch('/api/v1/mission-profiles');
            if (!res.ok) return;
            const body = await res.json();
            const profiles: MissionProfile[] = body.data ?? [];
            const active = profiles.find((p) => p.is_active);
            set({ missionProfiles: profiles, activeProfileId: active?.id ?? null });
        } catch { /* degraded — silent */ }
    },

    createMissionProfile: async (p: MissionProfileCreate) => {
        try {
            const res = await fetch('/api/v1/mission-profiles', {
                method: 'POST',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify(p),
            });
            if (!res.ok) {
                console.error('[PROFILES] Create failed:', await res.text());
                return null;
            }
            const body = await res.json();
            const created: MissionProfile = body.data;
            set((s) => ({ missionProfiles: [...s.missionProfiles, created] }));
            return created;
        } catch (err) {
            console.error('[PROFILES] Create error:', err);
            return null;
        }
    },

    updateMissionProfile: async (id: string, p: MissionProfileCreate) => {
        try {
            const res = await fetch(`/api/v1/mission-profiles/${id}`, {
                method: 'PUT',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify(p),
            });
            if (!res.ok) {
                console.error('[PROFILES] Update failed:', await res.text());
                return;
            }
            const body = await res.json();
            const updated: MissionProfile = body.data;
            set((s) => ({
                missionProfiles: s.missionProfiles.map((mp) => mp.id === id ? updated : mp),
            }));
        } catch (err) {
            console.error('[PROFILES] Update error:', err);
        }
    },

    deleteMissionProfile: async (id: string) => {
        try {
            const res = await fetch(`/api/v1/mission-profiles/${id}`, { method: 'DELETE' });
            if (!res.ok) {
                console.error('[PROFILES] Delete failed:', await res.text());
                return;
            }
            set((s) => ({
                missionProfiles: s.missionProfiles.filter((mp) => mp.id !== id),
                activeProfileId: s.activeProfileId === id ? null : s.activeProfileId,
            }));
        } catch (err) {
            console.error('[PROFILES] Delete error:', err);
        }
    },

    activateMissionProfile: async (id: string) => {
        try {
            const res = await fetch(`/api/v1/mission-profiles/${id}/activate`, { method: 'POST' });
            if (!res.ok) {
                console.error('[PROFILES] Activate failed:', await res.text());
                return;
            }
            const body = await res.json();
            const activated: MissionProfile = body.data;
            set((s) => ({
                activeProfileId: id,
                missionProfiles: s.missionProfiles.map((mp) =>
                    mp.id === id
                        ? activated
                        : mp.auto_start ? mp : { ...mp, is_active: false }
                ),
            }));
        } catch (err) {
            console.error('[PROFILES] Activate error:', err);
        }
    },

    // ── Context Snapshots ────────────────────────────────────────

    fetchContextSnapshots: async () => {
        try {
            const res = await fetch('/api/v1/context/snapshots');
            if (!res.ok) return;
            const body = await res.json();
            set({ contextSnapshots: body.data ?? [] });
        } catch { /* degraded — silent */ }
    },

    createContextSnapshot: async (name: string) => {
        const { missionChat, activeRunId, missionProfiles, activeProfileId } = get();
        // Build role_providers map from active profile
        const activeProfile = missionProfiles.find((p) => p.id === activeProfileId);
        const roleProviders = activeProfile?.role_providers ?? {};
        try {
            const res = await fetch('/api/v1/context/snapshot', {
                method: 'POST',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify({
                    name,
                    messages: missionChat,
                    run_state: { run_id: activeRunId },
                    role_providers: roleProviders,
                    source_profile: activeProfileId ?? undefined,
                }),
            });
            if (!res.ok) {
                console.error('[SNAPSHOTS] Create failed:', await res.text());
                return null;
            }
            const body = await res.json();
            const snap: ContextSnapshot = body.data;
            set((s) => ({ contextSnapshots: [snap, ...s.contextSnapshots] }));
            return snap;
        } catch (err) {
            console.error('[SNAPSHOTS] Create error:', err);
            return null;
        }
    },

    // ── V7: Trigger Rules (Team B) ──────────────────────────────

    fetchTriggerRules: async () => {
        set({ isFetchingTriggers: true });
        try {
            const res = await fetch('/api/v1/triggers');
            if (!res.ok) return;
            const body = await res.json();
            set({ triggerRules: body.data ?? [] });
        } catch { /* degraded — silent */ }
        finally { set({ isFetchingTriggers: false }); }
    },

    createTriggerRule: async (r: TriggerRuleCreate) => {
        try {
            const res = await fetch('/api/v1/triggers', {
                method: 'POST',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify(r),
            });
            if (!res.ok) {
                console.error('[TRIGGERS] Create failed:', await res.text());
                return null;
            }
            const body = await res.json();
            const created: TriggerRule = body.data;
            set((s) => ({ triggerRules: [created, ...s.triggerRules] }));
            return created;
        } catch (err) {
            console.error('[TRIGGERS] Create error:', err);
            return null;
        }
    },

    updateTriggerRule: async (id: string, r: TriggerRuleCreate) => {
        try {
            const res = await fetch(`/api/v1/triggers/${id}`, {
                method: 'PUT',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify(r),
            });
            if (!res.ok) {
                console.error('[TRIGGERS] Update failed:', await res.text());
                return;
            }
            // Refetch to get fresh data
            get().fetchTriggerRules();
        } catch (err) {
            console.error('[TRIGGERS] Update error:', err);
        }
    },

    deleteTriggerRule: async (id: string) => {
        try {
            const res = await fetch(`/api/v1/triggers/${id}`, { method: 'DELETE' });
            if (!res.ok) {
                console.error('[TRIGGERS] Delete failed:', await res.text());
                return;
            }
            set((s) => ({
                triggerRules: s.triggerRules.filter((t) => t.id !== id),
            }));
        } catch (err) {
            console.error('[TRIGGERS] Delete error:', err);
        }
    },

    toggleTriggerRule: async (id: string, isActive: boolean) => {
        try {
            const res = await fetch(`/api/v1/triggers/${id}/toggle`, {
                method: 'POST',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify({ is_active: isActive }),
            });
            if (!res.ok) {
                console.error('[TRIGGERS] Toggle failed:', await res.text());
                return;
            }
            set((s) => ({
                triggerRules: s.triggerRules.map((t) =>
                    t.id === id ? { ...t, is_active: isActive } : t
                ),
            }));
        } catch (err) {
            console.error('[TRIGGERS] Toggle error:', err);
        }
    },

    // ── V7: Conversation Log ─────────────────────────────────────

    fetchRunConversation: async (runId: string, agentFilter?: string) => {
        set({ isFetchingConversation: true });
        try {
            const params = new URLSearchParams();
            if (agentFilter) params.set('agent', agentFilter);
            const qs = params.toString();
            const url = `/api/v1/runs/${runId}/conversation${qs ? `?${qs}` : ''}`;
            const res = await fetch(url);
            if (!res.ok) {
                console.error('[CONVERSATION] Fetch failed:', res.status);
                return;
            }
            const body = await res.json();
            const turns = body.data?.turns ?? body.data ?? [];
            set({ conversationTurns: Array.isArray(turns) ? turns : [] });
        } catch (err) {
            console.error('[CONVERSATION] Fetch error:', err);
        } finally {
            set({ isFetchingConversation: false });
        }
    },

    interjectInRun: async (runId: string, message: string, agentId?: string) => {
        try {
            const payload: Record<string, string> = { message };
            if (agentId) payload.agent_id = agentId;
            const res = await fetch(`/api/v1/runs/${runId}/interject`, {
                method: 'POST',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify(payload),
            });
            if (!res.ok) {
                console.error('[CONVERSATION] Interject failed:', await res.text());
                return;
            }
            // Re-fetch conversation to include the new interjection
            get().fetchRunConversation(runId);
        } catch (err) {
            console.error('[CONVERSATION] Interject error:', err);
        }
    },
}));

// ── Auto-sync missionChat → localStorage ─────────────────────
useCortexStore.subscribe((state, prevState) => {
    if (state.missionChat !== prevState.missionChat) {
        persistChat(state.missionChat);
    }
});
