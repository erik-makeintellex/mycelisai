import { type Edge, type Node, Position } from 'reactflow';
import type { AgentNodeData } from '@/components/wiring/AgentNode';
import type {
    ChatMessage,
    MissionBlueprint,
    ModuleBindingData,
    ProposalData,
    StreamSignal,
    TeamExpressionData,
} from '@/store/useCortexStore';

const TEAM_WIDTH = 280;
const TEAM_GAP = 60;
const AGENT_SPACING_Y = 130;
const AGENT_OFFSET_X = 60;
const TEAM_HEADER_Y = 80;

export const CHAT_STORAGE_KEY = 'mycelis-workspace-chat';
const CHAT_STORAGE_KEY_LEGACY = 'mycelis-mission-chat'; // migrate old key
const CHAT_MAX_PERSISTED = 200; // cap to avoid localStorage quota issues

export function blueprintToGraph(bp: MissionBlueprint): { nodes: Node[]; edges: Edge[] } {
    const nodes: Node[] = [];
    const edges: Edge[] = [];
    const outputMap = new Map<string, string>(); // topic -> agentNodeId

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

    // Wire edges by matching output->input topics
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

export function solidifyNodes(nodes: Node[]): Node[] {
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

/** Dispatch an SSE signal to matching ReactFlow nodes. */
export function dispatchSignalToNodes(signal: StreamSignal, nodes: Node[]): Node[] | null {
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

function uniqueStrings(values: string[]): string[] {
    const seen = new Set<string>();
    const out: string[] = [];
    for (const value of values) {
        const v = value.trim();
        if (!v || seen.has(v)) continue;
        seen.add(v);
        out.push(v);
    }
    return out;
}

function normalizeModuleBindings(raw: unknown): ModuleBindingData[] {
    if (!Array.isArray(raw)) return [];
    const bindings: ModuleBindingData[] = [];
    for (const item of raw) {
        if (!item || typeof item !== 'object') continue;
        const rec = item as Record<string, unknown>;
        const moduleID = typeof rec.module_id === 'string'
            ? rec.module_id
            : typeof rec.moduleId === 'string'
                ? rec.moduleId
                : '';
        if (!moduleID.trim()) continue;
        bindings.push({
            binding_id: typeof rec.binding_id === 'string'
                ? rec.binding_id
                : typeof rec.bindingId === 'string'
                    ? rec.bindingId
                    : undefined,
            module_id: moduleID.trim(),
            adapter_kind: typeof rec.adapter_kind === 'string'
                ? rec.adapter_kind
                : typeof rec.adapterKind === 'string'
                    ? rec.adapterKind
                    : undefined,
            operation: typeof rec.operation === 'string' ? rec.operation : undefined,
        });
    }
    return bindings;
}

function normalizeTeamExpressions(raw: unknown): TeamExpressionData[] {
    if (!Array.isArray(raw)) return [];
    const expressions: TeamExpressionData[] = [];
    for (const item of raw) {
        if (!item || typeof item !== 'object') continue;
        const rec = item as Record<string, unknown>;
        const objective = typeof rec.objective === 'string' ? rec.objective.trim() : '';
        const rolePlanRaw = Array.isArray(rec.role_plan) ? rec.role_plan : Array.isArray(rec.rolePlan) ? rec.rolePlan : [];
        const rolePlan = rolePlanRaw.filter((v): v is string => typeof v === 'string').map((v) => v.trim()).filter(Boolean);
        const moduleBindings = normalizeModuleBindings(rec.module_bindings ?? rec.moduleBindings);
        if (!objective && moduleBindings.length === 0) continue;
        expressions.push({
            expression_id: typeof rec.expression_id === 'string'
                ? rec.expression_id
                : typeof rec.expressionId === 'string'
                    ? rec.expressionId
                    : undefined,
            team_id: typeof rec.team_id === 'string'
                ? rec.team_id
                : typeof rec.teamId === 'string'
                    ? rec.teamId
                    : undefined,
            objective: objective || `Execute ${moduleBindings[0]?.module_id ?? 'operation'}`,
            role_plan: rolePlan,
            module_bindings: moduleBindings,
        });
    }
    return expressions;
}

export function normalizeProposalData(raw: unknown): ProposalData | undefined {
    if (!raw || typeof raw !== 'object') return undefined;
    const rec = raw as Record<string, unknown>;
    const teamExpressions = normalizeTeamExpressions(rec.team_expressions ?? rec.teamExpressions);
    const derivedTools = uniqueStrings(
        teamExpressions.flatMap((expr) => (expr.module_bindings ?? []).map((b) => b.module_id)),
    );
    const derivedTeams = uniqueStrings(teamExpressions.map((expr) => expr.team_id ?? '')).length;
    const derivedAgents = teamExpressions.reduce((sum, expr) => sum + (expr.role_plan?.length ?? 0), 0);
    const rawTools = Array.isArray(rec.tools) ? rec.tools.filter((v): v is string => typeof v === 'string') : [];

    return {
        intent: typeof rec.intent === 'string' && rec.intent.trim() ? rec.intent : 'chat-action',
        teams: typeof rec.teams === 'number' ? rec.teams : derivedTeams,
        agents: typeof rec.agents === 'number' ? rec.agents : derivedAgents,
        tools: uniqueStrings(rawTools.length > 0 ? rawTools : derivedTools),
        risk_level: typeof rec.risk_level === 'string' && rec.risk_level.trim() ? rec.risk_level : 'medium',
        confirm_token: typeof rec.confirm_token === 'string' ? rec.confirm_token : '',
        intent_proof_id: typeof rec.intent_proof_id === 'string' ? rec.intent_proof_id : '',
        team_expressions: teamExpressions.length > 0 ? teamExpressions : undefined,
    };
}

// Soma's memory: chat survives page refreshes. Use clearMissionChat to reset.
export function loadPersistedChat(): ChatMessage[] {
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

export function persistChat(messages: ChatMessage[]) {
    if (typeof window === 'undefined') return;
    try {
        // Only persist the last N messages to respect localStorage limits
        const toStore = messages.slice(-CHAT_MAX_PERSISTED);
        localStorage.setItem(CHAT_STORAGE_KEY, JSON.stringify(toStore));
    } catch {
        // quota exceeded - silently drop
    }
}
