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

// ── Domain Types ──────────────────────────────────────────────

export interface ChatMessage {
    role: 'user' | 'architect';
    content: string;
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

interface AgentManifest {
    id: string;
    role: string;
    system_prompt?: string;
    model?: string;
    inputs?: string[];
    outputs?: string[];
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

            const bp = data as MissionBlueprint;
            const { nodes, edges } = blueprintToGraph(bp);

            const agentCount = bp.teams.reduce((sum, t) => sum + t.agents.length, 0);
            const summary = [
                `Blueprint **${bp.mission_id}** generated.`,
                `${bp.teams.length} team${bp.teams.length !== 1 ? 's' : ''}, ${agentCount} agent${agentCount !== 1 ? 's' : ''}.`,
                bp.constraints && bp.constraints.length > 0
                    ? `${bp.constraints.length} constraint${bp.constraints.length !== 1 ? 's' : ''} applied.`
                    : '',
            ]
                .filter(Boolean)
                .join(' ');

            set((s) => ({
                isDrafting: false,
                blueprint: bp,
                nodes,
                edges,
                missionStatus: 'draft',
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
}));
