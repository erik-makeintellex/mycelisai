import { describe, it, expect, beforeEach, vi } from 'vitest';

vi.mock('reactflow', () => {
    const Position = {
        Left: 'left',
        Right: 'right',
        Top: 'top',
        Bottom: 'bottom',
    };
    return {
        __esModule: true,
        Position,
        applyNodeChanges: (_changes: any[], nodes: any[]) => nodes,
        applyEdgeChanges: (_changes: any[], edges: any[]) => edges,
    };
});

import type { MissionBlueprint } from '@/store/useCortexStore';
import {
    buildChatSessionStorageKey,
    CHAT_STORAGE_KEY,
    buildChatStorageKey,
    clearPersistedChat,
    blueprintToGraph,
    dispatchSignalToNodes,
    loadOrCreateChatSessionId,
    loadPersistedChat,
    normalizeProposalData,
    persistChat,
    solidifyNodes,
} from '@/store/cortexStoreUtils';

describe('cortexStoreUtils', () => {
    beforeEach(() => {
        localStorage.clear();
    });

    it('builds graph nodes and edges from blueprint io wiring', () => {
        const blueprint: MissionBlueprint = {
            mission_id: 'm-1',
            intent: 'wire teams',
            teams: [
                {
                    name: 'team-a',
                    role: 'producer',
                    agents: [{ id: 'agent-a', role: 'architect', outputs: ['topic.a'] }],
                },
                {
                    name: 'team-b',
                    role: 'consumer',
                    agents: [{ id: 'agent-b', role: 'coder', inputs: ['topic.a'] }],
                },
            ],
        };

        const { nodes, edges } = blueprintToGraph(blueprint);
        expect(nodes.some((n) => n.id === 'team-0')).toBe(true);
        expect(nodes.some((n) => n.id === 'team-0-label')).toBe(true);
        expect(nodes.some((n) => n.id === 'agent-0-0')).toBe(true);
        expect(nodes.some((n) => n.id === 'agent-1-0')).toBe(true);
        expect(edges).toHaveLength(1);
        expect(edges[0]).toMatchObject({
            source: 'agent-0-0',
            target: 'agent-1-0',
            type: 'dataWire',
        });
    });

    it('solidifies draft nodes by clearing draft class and marking agent online', () => {
        const nodes: any[] = [
            {
                id: 'team-0',
                type: 'group',
                className: 'ghost-draft',
                style: { border: '1px dashed red' },
                data: {},
            },
            {
                id: 'agent-0-0',
                type: 'agentNode',
                className: 'ghost-draft',
                data: { status: 'offline' },
            },
        ];

        const solid = solidifyNodes(nodes);
        expect(solid[0].className).toBe('');
        expect(String(solid[0].style?.border)).toContain('solid');
        expect(solid[1].className).toBe('');
        expect(solid[1].data?.status).toBe('online');
    });

    it('dispatches thought and error signals to matching nodes', () => {
        const nodes: any[] = [
            { id: 'agent-a', data: { label: 'agent-a', status: 'online' } },
            { id: 'agent-b', data: { label: 'agent-b', status: 'online' } },
        ];

        const thought = dispatchSignalToNodes(
            { type: 'thought', source: 'agent-a', message: 'reasoning' },
            nodes,
        );
        expect(thought).not.toBeNull();
        expect(thought?.[0].data?.isThinking).toBe(true);
        expect(thought?.[0].data?.lastThought).toBe('reasoning');

        const errored = dispatchSignalToNodes(
            { type: 'error', source: 'agent-a', message: 'failed' },
            thought ?? nodes,
        );
        expect(errored).not.toBeNull();
        expect(errored?.[0].data?.status).toBe('error');
        expect(errored?.[0].data?.isThinking).toBe(false);
    });

    it('normalizes proposal data and derives team/agent/tool counts from team expressions', () => {
        const proposal = normalizeProposalData({
            intent: 'ship feature',
            risk_level: 'medium',
            confirm_token: 'ct-1',
            intent_proof_id: 'ip-1',
            task_cadence: 'scheduled',
            schedule_summary: 'Every weekday at 9 AM.',
            bus_scope: 'multi_team',
            nats_subjects: ['swarm.team.a.signal.status', 'swarm.team.b.signal.result'],
            team_expressions: [
                {
                    team_id: 'admin-core',
                    objective: 'deliver',
                    role_plan: ['architect', 'coder'],
                    module_bindings: [
                        { module_id: 'delegate_task', adapter_kind: 'internal' },
                        { module_id: 'mcp:github/create_issue', adapter_kind: 'mcp' },
                    ],
                },
            ],
        });

        expect(proposal).toBeDefined();
        expect(proposal?.teams).toBe(1);
        expect(proposal?.agents).toBe(2);
        expect(proposal?.tools).toEqual(['delegate_task', 'mcp:github/create_issue']);
        expect(proposal?.task_cadence).toBe('scheduled');
        expect(proposal?.schedule_summary).toBe('Every weekday at 9 AM.');
        expect(proposal?.bus_scope).toBe('multi_team');
        expect(proposal?.nats_subjects).toEqual(['swarm.team.a.signal.status', 'swarm.team.b.signal.result']);
        expect(proposal?.team_expressions?.[0].module_bindings?.[0]).toMatchObject({
            module_id: 'delegate_task',
            adapter_kind: 'internal',
        });
    });

    it('persists and reloads chat history from storage', () => {
        persistChat([
            { role: 'user', content: 'hello' },
            { role: 'council', content: 'world' },
        ] as any);

        const loaded = loadPersistedChat();
        expect(loaded).toHaveLength(2);
        expect(localStorage.getItem(CHAT_STORAGE_KEY)).toContain('hello');
        expect(loaded[1]).toMatchObject({ role: 'council', content: 'world' });
    });

    it('scopes persisted chat by organization key', () => {
        persistChat([{ role: 'user', content: 'org-a' }] as any, 'org-a');
        persistChat([{ role: 'user', content: 'org-b' }] as any, 'org-b');

        expect(loadPersistedChat('org-a')).toMatchObject([{ role: 'user', content: 'org-a' }]);
        expect(loadPersistedChat('org-b')).toMatchObject([{ role: 'user', content: 'org-b' }]);
        expect(localStorage.getItem(buildChatStorageKey('org-a'))).toContain('org-a');
        expect(localStorage.getItem(buildChatStorageKey('org-b'))).toContain('org-b');
    });

    it('creates and reuses scoped chat session ids for server-side conversation continuity', () => {
        const first = loadOrCreateChatSessionId('org-a');
        const second = loadOrCreateChatSessionId('org-a');
        const otherScope = loadOrCreateChatSessionId('org-b');

        expect(first).toMatch(/^[0-9a-f-]{36}$/i);
        expect(second).toBe(first);
        expect(otherScope).not.toBe(first);
        expect(localStorage.getItem(buildChatSessionStorageKey('org-a'))).toBe(first);
        expect(localStorage.getItem(buildChatSessionStorageKey('org-b'))).toBe(otherScope);
    });

    it('clears only the requested scoped chat history', () => {
        persistChat([{ role: 'user', content: 'scoped' }] as any, 'org-1');
        persistChat([{ role: 'user', content: 'global' }] as any);
        const scopedSession = loadOrCreateChatSessionId('org-1');
        const globalSession = loadOrCreateChatSessionId();

        clearPersistedChat('org-1');

        expect(loadPersistedChat('org-1')).toEqual([]);
        expect(loadPersistedChat()).toMatchObject([{ role: 'user', content: 'global' }]);
        expect(localStorage.getItem(buildChatSessionStorageKey('org-1'))).toBeNull();
        expect(localStorage.getItem(buildChatSessionStorageKey())).toBe(globalSession);
        expect(scopedSession).toBeTruthy();
    });
});
