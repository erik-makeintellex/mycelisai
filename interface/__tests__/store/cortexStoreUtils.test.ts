import { describe, it, expect, beforeEach, vi } from 'vitest';
import type { Node } from 'reactflow';

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
        applyNodeChanges: (_changes: unknown[], nodes: unknown[]) => nodes,
        applyEdgeChanges: (_changes: unknown[], edges: unknown[]) => edges,
    };
});

import type { ChatMessage, MissionBlueprint } from '@/store/useCortexStore';
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
        const nodes: Node[] = [
            {
                id: 'team-0',
                type: 'group',
                position: { x: 0, y: 0 },
                className: 'ghost-draft',
                style: { border: '1px dashed red' },
                data: {},
            },
            {
                id: 'agent-0-0',
                type: 'agentNode',
                position: { x: 0, y: 0 },
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
        const nodes: Node[] = [
            { id: 'agent-a', position: { x: 0, y: 0 }, data: { label: 'agent-a', status: 'online' } },
            { id: 'agent-b', position: { x: 0, y: 0 }, data: { label: 'agent-b', status: 'online' } },
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

    it('normalizes structured work intent into proposal execution posture', () => {
        const proposal = normalizeProposalData({
            intent: 'prepare weekly client brief',
            risk_level: 'medium',
            confirm_token: 'ct-brief',
            intent_proof_id: 'ip-brief',
            execution_mode: 'schedule_handoff',
            work_intent: {
                kind: 'scheduled_workflow',
                objective: 'Generate the client brief each Monday.',
                cadence: 'scheduled',
                schedule_summary: 'Every Monday at 8 AM.',
                runtime_posture: 'Wait for source updates, then run the brief workflow.',
                bus_scope: 'current_team',
                target_team_id: 'client-brief-team',
                nats_subjects: ['swarm.team.client-brief.signal.status'],
                service_refs: ['scheduler.weekly-client-brief'],
                project_ref: 'client-briefs',
            },
        });

        expect(proposal).toBeDefined();
        expect(proposal?.execution_mode).toBe('schedule_handoff');
        expect(proposal?.task_cadence).toBe('scheduled');
        expect(proposal?.schedule_summary).toBe('Every Monday at 8 AM.');
        expect(proposal?.runtime_posture).toBe('Wait for source updates, then run the brief workflow.');
        expect(proposal?.bus_scope).toBe('current_team');
        expect(proposal?.nats_subjects).toEqual(['swarm.team.client-brief.signal.status']);
        expect(proposal?.work_intent).toMatchObject({
            objective: 'Generate the client brief each Monday.',
            target_team_id: 'client-brief-team',
            project_ref: 'client-briefs',
        });
    });

    it('persists and reloads chat history from storage', () => {
        const messages: ChatMessage[] = [
            { role: 'user', content: 'hello' },
            { role: 'council', content: 'world' },
        ];
        persistChat(messages);

        const loaded = loadPersistedChat();
        expect(loaded).toHaveLength(2);
        expect(localStorage.getItem(CHAT_STORAGE_KEY)).toContain('hello');
        expect(loaded[1]).toMatchObject({ role: 'council', content: 'world' });
    });

    it('sanitizes malformed legacy chat history before rehydrating the dashboard', () => {
        localStorage.setItem(CHAT_STORAGE_KEY, JSON.stringify([
            { role: 'user', content: { text: 'legacy object prompt' } },
            {
                role: 'architect',
                content: null,
                tools_used: ['web_search', 'web_search', 12],
                response_depth: 'definitely_not_real',
                proposal: {
                    intent: 'legacy proposal',
                    risk_level: 'medium',
                    confirm_token: 123,
                    intent_proof_id: 'proof-1',
                    tools: 'delegate_task',
                },
                execution_summary: {
                    outputs: { label: 'not an array' },
                    understanding: { assumptions: 'not an array' },
                },
            },
            { role: 'unknown', content: 'drop me' },
        ]));

        const loaded = loadPersistedChat();

        expect(loaded).toHaveLength(2);
        expect(loaded[0]).toMatchObject({
            role: 'user',
            content: '{"text":"legacy object prompt"}',
        });
        expect(loaded[1].content).toBe('');
        expect(loaded[1].tools_used).toEqual(['web_search', '12']);
        expect(loaded[1].response_depth).toBeUndefined();
        expect(loaded[1].proposal).toMatchObject({
            intent: 'legacy proposal',
            risk_level: 'medium',
            confirm_token: '123',
            intent_proof_id: 'proof-1',
        });
        expect(loaded[1].proposal?.tools).toEqual(['delegate_task']);
    });

    it('scopes persisted chat by organization key', () => {
        const orgAMessages: ChatMessage[] = [{ role: 'user', content: 'org-a' }];
        const orgBMessages: ChatMessage[] = [{ role: 'user', content: 'org-b' }];
        persistChat(orgAMessages, 'org-a');
        persistChat(orgBMessages, 'org-b');

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

    it('repairs legacy non-UUID chat session ids before sending workspace chat', () => {
        localStorage.setItem(buildChatSessionStorageKey('org-a'), 'session-1');

        const repaired = loadOrCreateChatSessionId('org-a');

        expect(repaired).toMatch(/^[0-9a-f]{8}-[0-9a-f]{4}-[1-5][0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12}$/i);
        expect(repaired).not.toBe('session-1');
        expect(localStorage.getItem(buildChatSessionStorageKey('org-a'))).toBe(repaired);
    });

    it('clears only the requested scoped chat history', () => {
        const scopedMessages: ChatMessage[] = [{ role: 'user', content: 'scoped' }];
        const globalMessages: ChatMessage[] = [{ role: 'user', content: 'global' }];
        persistChat(scopedMessages, 'org-1');
        persistChat(globalMessages);
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
