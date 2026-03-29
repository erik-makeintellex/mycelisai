import { type Edge, type Node, Position } from 'reactflow';
import type { AgentNodeData } from '@/components/wiring/AgentNode';
import type { MissionBlueprint } from '@/store/useCortexStore';

const TEAM_WIDTH = 280;
const TEAM_GAP = 60;
const AGENT_SPACING_Y = 130;
const AGENT_OFFSET_X = 60;
const TEAM_HEADER_Y = 80;

export function blueprintToGraph(bp: MissionBlueprint): { nodes: Node[]; edges: Edge[] } {
    const nodes: Node[] = [];
    const edges: Edge[] = [];
    const outputMap = new Map<string, string>();
    let teamX = 80;

    bp.teams.forEach((team, tIdx) => {
        const teamId = `team-${tIdx}`;
        const teamHeight = TEAM_HEADER_Y + team.agents.length * AGENT_SPACING_Y + 40;

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

        team.agents.forEach((agent, aIdx) => {
            const agentId = `agent-${tIdx}-${aIdx}`;
            nodes.push({
                id: agentId,
                type: 'agentNode',
                position: { x: AGENT_OFFSET_X, y: TEAM_HEADER_Y + aIdx * AGENT_SPACING_Y },
                parentNode: teamId,
                extent: 'parent' as const,
                className: 'ghost-draft',
                data: {
                    label: agent.id,
                    role: agent.role,
                    status: 'offline',
                    lastThought: agent.system_prompt ? agent.system_prompt.slice(0, 60) : undefined,
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
