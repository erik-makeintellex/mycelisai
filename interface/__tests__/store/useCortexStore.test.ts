import { describe, it, expect, vi, beforeEach } from 'vitest';
import { useCortexStore } from '@/store/useCortexStore';
import { blueprintToGraph } from '@/store/cortexStoreUtils';
import { mockFetch } from '../setup';

// Direct store testing — no React rendering needed
const store = useCortexStore;
const baseBlueprint = {
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

describe('useCortexStore', () => {
    beforeEach(() => {
        localStorage.clear();
        const { nodes, edges } = blueprintToGraph(baseBlueprint);
        // Reset relevant state between tests
        store.setState({
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
            mcpServers: [],
            isFetchingMCPServers: false,
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
        });
    });

    // ── fetchMissions ────────────────────────────────────────────

    describe('fetchMissions', () => {
        it('sets isFetchingMissions during fetch', async () => {
            mockFetch.mockResolvedValue({
                ok: true,
                json: async () => [],
            });

            const promise = store.getState().fetchMissions();
            expect(store.getState().isFetchingMissions).toBe(true);

            await promise;
            expect(store.getState().isFetchingMissions).toBe(false);
        });

        it('stores missions array from API', async () => {
            const missions = [
                { id: 'm1', intent: 'Scan', status: 'active', teams: 2, agents: 5 },
            ];
            mockFetch.mockResolvedValue({ ok: true, json: async () => missions });

            await store.getState().fetchMissions();

            expect(store.getState().missions).toEqual(missions);
        });

        it('sets empty array on non-ok response', async () => {
            mockFetch.mockResolvedValue({ ok: false });

            await store.getState().fetchMissions();

            expect(store.getState().missions).toEqual([]);
        });

        it('sets empty array on network error', async () => {
            mockFetch.mockRejectedValue(new Error('Network error'));

            await store.getState().fetchMissions();

            expect(store.getState().missions).toEqual([]);
        });

        it('handles non-array response gracefully', async () => {
            mockFetch.mockResolvedValue({
                ok: true,
                json: async () => ({ error: 'not array' }),
            });

            await store.getState().fetchMissions();

            expect(store.getState().missions).toEqual([]);
        });
    });

    // ── fetchArtifacts ───────────────────────────────────────────

    describe('fetchArtifacts', () => {
        it('fetches all artifacts without filters', async () => {
            const artifacts = [
                { id: 'a1', agent_id: 'ag1', artifact_type: 'code', title: 'Output', content_type: 'text', metadata: {}, status: 'pending', created_at: '' },
            ];
            mockFetch.mockResolvedValue({ ok: true, json: async () => artifacts });

            await store.getState().fetchArtifacts();

            expect(mockFetch).toHaveBeenCalledWith('/api/v1/artifacts');
            expect(store.getState().artifacts).toEqual(artifacts);
        });

        it('passes filters as query params', async () => {
            mockFetch.mockResolvedValue({ ok: true, json: async () => [] });

            await store.getState().fetchArtifacts({
                mission_id: 'm1',
                team_id: 't1',
                limit: 10,
            });

            const url = mockFetch.mock.calls[0][0] as string;
            expect(url).toContain('mission_id=m1');
            expect(url).toContain('team_id=t1');
            expect(url).toContain('limit=10');
        });

        it('sets empty array on failure', async () => {
            mockFetch.mockRejectedValue(new Error('fail'));

            await store.getState().fetchArtifacts();

            expect(store.getState().artifacts).toEqual([]);
        });

        it('updateArtifactStatus uses the PUT contract and syncs selected detail', async () => {
            store.setState({
                artifacts: [
                    { id: 'a1', agent_id: 'ag1', artifact_type: 'code', title: 'Output', content_type: 'text', metadata: {}, status: 'pending', created_at: '' },
                ],
                selectedArtifactDetail: {
                    id: 'a1',
                    agent_id: 'ag1',
                    artifact_type: 'code',
                    title: 'Output',
                    content_type: 'text',
                    metadata: {},
                    status: 'pending',
                    created_at: '',
                },
            });
            mockFetch.mockResolvedValue({ ok: true });

            await store.getState().updateArtifactStatus('a1', 'approved');

            expect(mockFetch).toHaveBeenCalledWith(
                '/api/v1/artifacts/a1/status',
                expect.objectContaining({ method: 'PUT' }),
            );
            expect(store.getState().artifacts[0].status).toBe('approved');
            expect(store.getState().selectedArtifactDetail?.status).toBe('approved');
        });
    });

    // ── fetchSensors ─────────────────────────────────────────────

    describe('fetchSensors', () => {
        it('stores sensors from API response', async () => {
            const sensors = [
                { id: 's1', type: 'email', status: 'online', last_seen: '', label: 'Gmail' },
            ];
            mockFetch.mockResolvedValue({
                ok: true,
                json: async () => ({ sensors }),
            });

            await store.getState().fetchSensors();

            expect(store.getState().sensorFeeds).toEqual(sensors);
        });

        it('handles missing sensors key gracefully', async () => {
            mockFetch.mockResolvedValue({
                ok: true,
                json: async () => ({}),
            });

            await store.getState().fetchSensors();

            expect(store.getState().sensorFeeds).toEqual([]);
        });
    });

    // ── fetchProposals ───────────────────────────────────────────

    describe('fetchProposals', () => {
        it('stores proposals from API response', async () => {
            const proposals = [
                { id: 'p1', name: 'Squad', role: 'analytics', agents: [], reason: 'Test', status: 'pending', created_at: '' },
            ];
            mockFetch.mockResolvedValue({
                ok: true,
                json: async () => ({ proposals }),
            });

            await store.getState().fetchProposals();

            expect(store.getState().teamProposals).toEqual(proposals);
        });

        it('approveProposal updates status in store', async () => {
            store.setState({
                teamProposals: [
                    { id: 'p1', name: 'Squad', role: 'test', agents: [], reason: 'r', status: 'pending', created_at: '' },
                ],
            });
            mockFetch.mockResolvedValue({ ok: true, json: async () => ({}) });

            await store.getState().approveProposal('p1');

            expect(store.getState().teamProposals[0].status).toBe('approved');
        });

        it('rejectProposal updates status in store', async () => {
            store.setState({
                teamProposals: [
                    { id: 'p1', name: 'Squad', role: 'test', agents: [], reason: 'r', status: 'pending', created_at: '' },
                ],
            });
            mockFetch.mockResolvedValue({ ok: true, json: async () => ({}) });

            await store.getState().rejectProposal('p1');

            expect(store.getState().teamProposals[0].status).toBe('rejected');
        });
    });

    // ── fetchCatalogue ───────────────────────────────────────────

    describe('fetchCatalogue', () => {
        it('stores catalogue agents from API', async () => {
            const agents = [
                { id: 'c1', name: 'Scanner', role: 'cognitive', tools: [], inputs: [], outputs: [], verification_rubric: [], created_at: '', updated_at: '' },
            ];
            mockFetch.mockResolvedValue({ ok: true, json: async () => agents });

            await store.getState().fetchCatalogue();

            expect(store.getState().catalogueAgents).toEqual(agents);
        });

        it('createCatalogueAgent adds to store', async () => {
            const created = { id: 'c1', name: 'New Agent', role: 'cognitive', tools: [], inputs: [], outputs: [], verification_rubric: [], created_at: '', updated_at: '' };
            mockFetch.mockResolvedValue({ ok: true, json: async () => created });

            await store.getState().createCatalogueAgent({ name: 'New Agent', role: 'cognitive' });

            expect(store.getState().catalogueAgents[0]).toEqual(created);
        });

        it('updateCatalogueAgent keeps the selected agent in sync', async () => {
            const existing = { id: 'c1', name: 'Scanner', role: 'cognitive', tools: [], inputs: [], outputs: [], verification_rubric: [], created_at: '', updated_at: '' };
            const updated = { ...existing, name: 'Scanner Prime' };
            store.setState({
                catalogueAgents: [existing],
                selectedCatalogueAgent: existing,
            });
            mockFetch.mockResolvedValue({ ok: true, json: async () => updated });

            await store.getState().updateCatalogueAgent('c1', { name: 'Scanner Prime' });

            expect(store.getState().catalogueAgents[0]).toEqual(updated);
            expect(store.getState().selectedCatalogueAgent).toEqual(updated);
        });

        it('deleteCatalogueAgent removes from store', async () => {
            store.setState({
                catalogueAgents: [
                    { id: 'c1', name: 'A1', role: 'cognitive', tools: [], inputs: [], outputs: [], verification_rubric: [], created_at: '', updated_at: '' },
                    { id: 'c2', name: 'A2', role: 'sensory', tools: [], inputs: [], outputs: [], verification_rubric: [], created_at: '', updated_at: '' },
                ],
            });
            mockFetch.mockResolvedValue({ ok: true });

            await store.getState().deleteCatalogueAgent('c1');

            expect(store.getState().catalogueAgents).toHaveLength(1);
            expect(store.getState().catalogueAgents[0].id).toBe('c2');
        });
    });

    // ── fetchMCPServers ──────────────────────────────────────────

    describe('fetchMCPServers', () => {
        it('stores MCP servers from API', async () => {
            const servers = [
                { id: 'srv1', name: 'filesystem', transport: 'stdio', status: 'connected', created_at: '', tools: [] },
            ];
            mockFetch.mockResolvedValue({ ok: true, json: async () => servers });

            await store.getState().fetchMCPServers();

            expect(store.getState().mcpServers).toEqual(servers);
        });

        it('deleteMCPServer removes from store', async () => {
            store.setState({
                mcpServers: [
                    { id: 'srv1', name: 'fs', transport: 'stdio' as const, status: 'connected', created_at: '', tools: [] },
                ],
            });
            mockFetch.mockResolvedValue({ ok: true });

            await store.getState().deleteMCPServer('srv1');

            expect(store.getState().mcpServers).toHaveLength(0);
        });

        it('stores persisted MCP activity from API', async () => {
            const activity = [
                {
                    id: 'mcp-1',
                    server_id: 'srv1',
                    server_name: 'filesystem',
                    tool_name: 'read_file',
                    state: 'completed',
                    summary: 'Read workspace brief successfully.',
                    message: 'Read workspace brief successfully.',
                    channel_name: 'browser.research.results',
                    timestamp: '2026-04-06T12:00:00Z',
                },
            ];
            mockFetch.mockResolvedValue({ ok: true, json: async () => ({ ok: true, data: activity }) });

            await store.getState().fetchMCPActivity();

            expect(mockFetch).toHaveBeenCalledWith('/api/v1/mcp/activity?limit=12');
            expect(store.getState().mcpActivity).toEqual(activity);
        });
    });

    // ── Trust Economy ────────────────────────────────────────────

    describe('trust', () => {
        it('fetchTrustThreshold stores threshold from API', async () => {
            mockFetch.mockResolvedValue({
                ok: true,
                json: async () => ({ threshold: 0.85 }),
            });

            await store.getState().fetchTrustThreshold();

            expect(store.getState().trustThreshold).toBe(0.85);
        });

        it('setTrustThreshold updates store and calls API', async () => {
            mockFetch.mockResolvedValue({ ok: true });

            store.getState().setTrustThreshold(0.9);

            expect(store.getState().trustThreshold).toBe(0.9);
            expect(mockFetch).toHaveBeenCalledWith('/api/v1/trust/threshold', expect.objectContaining({
                method: 'PUT',
            }));
        });
    });

    // ── toggleSensorGroup ────────────────────────────────────────

    describe('toggleSensorGroup', () => {
        it('adds group to subscribed list', () => {
            store.getState().toggleSensorGroup('email');
            expect(store.getState().subscribedSensorGroups).toContain('email');
        });

        it('removes group if already subscribed', () => {
            store.setState({ subscribedSensorGroups: ['email', 'weather'] });
            store.getState().toggleSensorGroup('email');
            expect(store.getState().subscribedSensorGroups).toEqual(['weather']);
        });
    });

    // ── Governance ───────────────────────────────────────────────

    describe('governance', () => {
        it('selectArtifact sets selected artifact', () => {
            const artifact = {
                id: 'a1', source: 'agent-1', signal: 'artifact' as const,
                timestamp: '', trust_score: 0.8,
                payload: { content: 'test', content_type: 'text' as const },
            };

            store.getState().selectArtifact(artifact);

            expect(store.getState().selectedArtifact).toEqual(artifact);
        });

        it('approveArtifact removes from pending list', () => {
            store.setState({
                pendingArtifacts: [
                    { id: 'a1', source: 's', signal: 'artifact' as const, timestamp: '', payload: { content: 'c', content_type: 'text' as const } },
                    { id: 'a2', source: 's', signal: 'artifact' as const, timestamp: '', payload: { content: 'c', content_type: 'text' as const } },
                ],
            });

            store.getState().approveArtifact('a1');

            expect(store.getState().pendingArtifacts).toHaveLength(1);
            expect(store.getState().pendingArtifacts[0].id).toBe('a2');
        });

        it('rejectArtifact removes from pending and clears selection', () => {
            const artifact = { id: 'a1', source: 's', signal: 'artifact' as const, timestamp: '', payload: { content: 'c', content_type: 'text' as const } };
            store.setState({
                pendingArtifacts: [artifact],
                selectedArtifact: artifact,
            });

            store.getState().rejectArtifact('a1', 'Not accurate');

            expect(store.getState().pendingArtifacts).toHaveLength(0);
            expect(store.getState().selectedArtifact).toBeNull();
        });

        it('fetchAuditLog stores recent audit entries from the API envelope', async () => {
            const auditLog = [
                {
                    id: 'audit-1',
                    actor: 'Soma',
                    user: 'local-user',
                    action: 'proposal_generated',
                    timestamp: '2026-03-26T12:00:00Z',
                    result_status: 'pending',
                    approval_status: 'approval_required',
                },
            ];
            mockFetch.mockResolvedValue({
                ok: true,
                json: async () => ({ ok: true, data: auditLog }),
            });

            await store.getState().fetchAuditLog();

            expect(mockFetch).toHaveBeenCalledWith('/api/v1/audit?limit=20');
            expect(store.getState().auditLog).toEqual(auditLog);
        });
    });

    // ── Mission draft / graph sync ──────────────────────────────

    describe('mission draft graph sync', () => {
        it('updateAgentInDraft updates the local blueprint and closes the editor', () => {
            store.setState({
                blueprint: structuredClone(baseBlueprint),
                selectedAgentNodeId: 'agent-0-0',
                isAgentEditorOpen: true,
            });

            store.getState().updateAgentInDraft(0, 0, { model: 'model-updated' });

            expect(store.getState().blueprint?.teams[0].agents[0].model).toBe('model-updated');
            expect(store.getState().selectedAgentNodeId).toBeNull();
            expect(store.getState().isAgentEditorOpen).toBe(false);
        });

        it('deleteAgentFromDraft removes the last agent team and rebuilds the graph', () => {
            store.setState({
                blueprint: {
                    mission_id: 'mission-2',
                    intent: 'Single agent mission',
                    teams: [
                        {
                            name: 'Solo',
                            role: 'operators',
                            agents: [
                                {
                                    id: 'solo-agent',
                                    role: 'cognitive',
                                    outputs: ['solo.out'],
                                },
                            ],
                        },
                    ],
                },
                selectedAgentNodeId: 'agent-0-0',
                isAgentEditorOpen: true,
            });

            store.getState().deleteAgentFromDraft(0, 0);

            expect(store.getState().blueprint?.teams).toEqual([]);
            expect(store.getState().nodes).toEqual([]);
            expect(store.getState().edges).toEqual([]);
            expect(store.getState().isAgentEditorOpen).toBe(false);
        });

        it('updateAgentInMission updates the active blueprint and solidifies draft nodes', async () => {
            const { nodes, edges } = blueprintToGraph(baseBlueprint);
            store.setState({
                blueprint: structuredClone(baseBlueprint),
                nodes,
                edges,
                activeMissionId: 'mission-1',
                selectedAgentNodeId: 'agent-0-0',
                isAgentEditorOpen: true,
            });
            mockFetch.mockResolvedValue({ ok: true });

            await store.getState().updateAgentInMission('alpha', { model: 'model-live' });

            expect(store.getState().blueprint?.teams[0].agents[0].model).toBe('model-live');
            expect(store.getState().nodes.some((node) => node.type === 'agentNode' && node.className === '')).toBe(true);
            expect(store.getState().selectedAgentNodeId).toBeNull();
            expect(store.getState().isAgentEditorOpen).toBe(false);
        });

        it('discardDraft clears the draft workspace state', () => {
            store.setState({
                blueprint: structuredClone(baseBlueprint),
                missionStatus: 'active',
                activeMissionId: 'mission-1',
                selectedAgentNodeId: 'agent-0-0',
                isAgentEditorOpen: true,
            });

            store.getState().discardDraft();

            expect(store.getState().blueprint).toBeNull();
            expect(store.getState().missionStatus).toBe('idle');
            expect(store.getState().activeMissionId).toBeNull();
            expect(store.getState().nodes).toEqual([]);
            expect(store.getState().edges).toEqual([]);
            expect(store.getState().isAgentEditorOpen).toBe(false);
        });

        it('deleteMission clears the draft workspace after a successful delete', async () => {
            store.setState({
                blueprint: structuredClone(baseBlueprint),
                missionStatus: 'active',
                activeMissionId: 'mission-1',
                selectedAgentNodeId: 'agent-0-0',
                isAgentEditorOpen: true,
            });
            mockFetch.mockResolvedValue({ ok: true });

            await store.getState().deleteMission('mission-1');

            expect(mockFetch).toHaveBeenCalledWith('/api/v1/missions/mission-1', { method: 'DELETE' });
            expect(store.getState().blueprint).toBeNull();
            expect(store.getState().missionStatus).toBe('idle');
            expect(store.getState().activeMissionId).toBeNull();
            expect(store.getState().isAgentEditorOpen).toBe(false);
        });
    });

    // ── Launch Crew / proposal confirmation ─────────────────────

    describe('sendMissionChat', () => {
        it('silently retries the first transient Soma failure and recovers on the second attempt', async () => {
            vi.useFakeTimers();
            try {
                mockFetch
                    .mockResolvedValueOnce({
                        ok: false,
                        status: 500,
                        text: async () => '{"error":"Soma chat blocked (500)"}',
                    })
                    .mockResolvedValueOnce({
                        ok: true,
                        json: async () => ({
                            ok: true,
                            data: {
                                meta: { source_node: 'admin', timestamp: new Date().toISOString() },
                                signal_type: 'chat_response',
                                trust_score: 0.5,
                                template_id: 'chat-to-answer',
                                mode: 'answer',
                                payload: {
                                    text: 'Recovered answer.',
                                    tools_used: [],
                                },
                            },
                        }),
                    });

                const sendPromise = store.getState().sendMissionChat('hello');
                await vi.advanceTimersByTimeAsync(400);
                await sendPromise;

                expect(mockFetch).toHaveBeenCalledTimes(2);
                expect(store.getState().missionChatFailure).toBeNull();
                expect(store.getState().missionChatError).toBeNull();
                expect(store.getState().workspaceChatPrimed).toBe(true);
                expect(store.getState().missionChat.at(-1)?.content).toBe('Recovered answer.');
            } finally {
                vi.useRealTimers();
            }
        });

        it('retries a cold-start Soma failure after a scope change even when a prior session was primed', async () => {
            vi.useFakeTimers();
            try {
                store.setState({
                    workspaceChatPrimed: true,
                    missionChat: [],
                });
                store.getState().setMissionChatScope('org-123');
                mockFetch
                    .mockResolvedValueOnce({
                        ok: false,
                        status: 500,
                        text: async () => '{"error":"Soma chat blocked (500)"}',
                    })
                    .mockResolvedValueOnce({
                        ok: true,
                        json: async () => ({
                            ok: true,
                            data: {
                                meta: { source_node: 'admin', timestamp: new Date().toISOString() },
                                signal_type: 'chat_response',
                                trust_score: 0.5,
                                template_id: 'chat-to-answer',
                                mode: 'answer',
                                payload: {
                                    text: 'Recovered answer after scoped reset.',
                                    tools_used: [],
                                },
                            },
                        }),
                    });

                const sendPromise = store.getState().sendMissionChat('hello');
                await vi.advanceTimersByTimeAsync(400);
                await sendPromise;

                expect(mockFetch).toHaveBeenCalledTimes(2);
                expect(store.getState().missionChatFailure).toBeNull();
                expect(store.getState().workspaceChatPrimed).toBe(true);
                expect(store.getState().missionChat.at(-1)?.content).toBe('Recovered answer after scoped reset.');
            } finally {
                vi.useRealTimers();
            }
        });

        it('normalizes team expressions and module bindings from proposal payload', async () => {
            mockFetch.mockResolvedValue({
                ok: true,
                json: async () => ({
                    ok: true,
                    data: {
                        meta: { source_node: 'admin', timestamp: new Date().toISOString() },
                        signal_type: 'chat_response',
                        trust_score: 0.5,
                        template_id: 'chat-to-proposal',
                        mode: 'proposal',
                        payload: {
                            text: 'I prepared a governed execution plan.',
                            tools_used: ['delegate'],
                            proposal: {
                                intent: 'chat-action',
                                tools: ['delegate'],
                                risk_level: 'medium',
                                confirm_token: 'ct-123',
                                intent_proof_id: 'ip-123',
                                team_expressions: [
                                    {
                                        expression_id: 'expr-1',
                                        team_id: 'admin-core',
                                        objective: 'Execute delegate through governed module binding',
                                        role_plan: ['admin'],
                                        module_bindings: [
                                            {
                                                binding_id: 'binding-1-delegate',
                                                module_id: 'delegate',
                                                adapter_kind: 'internal',
                                                operation: 'delegate',
                                            },
                                        ],
                                    },
                                ],
                            },
                        },
                    },
                }),
            });

            await store.getState().sendMissionChat('launch a team');

            expect(store.getState().activeMode).toBe('proposal');
            expect(store.getState().activeConfirmToken).toBe('ct-123');
            expect(store.getState().pendingProposal).toMatchObject({
                intent: 'chat-action',
                teams: 1,
                agents: 1,
                tools: ['delegate'],
                confirm_token: 'ct-123',
                intent_proof_id: 'ip-123',
            });
            expect(store.getState().pendingProposal?.team_expressions?.[0]).toMatchObject({
                expression_id: 'expr-1',
                team_id: 'admin-core',
                objective: 'Execute delegate through governed module binding',
            });
            expect(store.getState().pendingProposal?.team_expressions?.[0].module_bindings?.[0]).toMatchObject({
                module_id: 'delegate',
                adapter_kind: 'internal',
            });
            expect(store.getState().missionChat.at(-1)).toMatchObject({
                proposal_status: 'active',
            });
        });

        it('stores governed artifact ask-class metadata for artifact-bearing Soma answers', async () => {
            mockFetch.mockResolvedValue({
                ok: true,
                json: async () => ({
                    ok: true,
                    data: {
                        meta: { source_node: 'admin', timestamp: new Date().toISOString() },
                        signal_type: 'chat_response',
                        trust_score: 0.5,
                        template_id: 'chat-to-answer',
                        mode: 'answer',
                        payload: {
                            text: 'I prepared a brief for review.',
                            ask_class: 'governed_artifact',
                            artifacts: [
                                { id: 'doc-1', type: 'document', title: 'Creative Brief', content_type: 'text/markdown', content: '# Brief' },
                            ],
                        },
                    },
                }),
            });

            await store.getState().sendMissionChat('create a brief');

            expect(store.getState().missionChat.at(-1)).toMatchObject({
                ask_class: 'governed_artifact',
            });
        });

        it('stores specialist consultation ask-class metadata for consulted answers', async () => {
            mockFetch.mockResolvedValue({
                ok: true,
                json: async () => ({
                    ok: true,
                    data: {
                        meta: { source_node: 'admin', timestamp: new Date().toISOString() },
                        signal_type: 'chat_response',
                        trust_score: 0.5,
                        template_id: 'chat-to-answer',
                        mode: 'answer',
                        payload: {
                            text: 'The architect reviewed the tradeoffs.',
                            ask_class: 'specialist_consultation',
                            consultations: [
                                { member: 'council-architect', summary: 'Recommend the safer option.' },
                            ],
                        },
                    },
                }),
            });

            await store.getState().sendMissionChat('review the architecture');

            expect(store.getState().missionChat.at(-1)).toMatchObject({
                ask_class: 'specialist_consultation',
            });
        });

        it('stores a structured workspace failure when Soma chat returns 500', async () => {
            vi.useFakeTimers();
            try {
                mockFetch.mockResolvedValue({
                    ok: false,
                    status: 500,
                    text: async () => '{"error":"Soma chat blocked (500)"}',
                });

                const sendPromise = store.getState().sendMissionChat('hello');
                await vi.advanceTimersByTimeAsync(400);
                await sendPromise;

                expect(mockFetch).toHaveBeenCalledTimes(2);
                expect(store.getState().activeMode).toBe('blocker');
                expect(store.getState().missionChatError).toBe('Soma hit a server-side failure while handling the request.');
                expect(store.getState().missionChatFailure).toMatchObject({
                    routeKind: 'workspace',
                    type: 'server_error',
                    bannerLabel: 'Workspace chat server error',
                });
            } finally {
                vi.useRealTimers();
            }
        });

        it('stores a setup-required blocker when Soma has no bound AI engine', async () => {
            mockFetch.mockResolvedValue({
                ok: false,
                status: 503,
                text: async () => JSON.stringify({
                    ok: false,
                    error: 'Soma is routed to an AI Engine that is configured but disabled.',
                    data: {
                        code: 'provider_disabled',
                        summary: 'Soma is routed to an AI Engine that is configured but disabled.',
                        recommended_action: 'Open Settings and enable a reachable AI Engine for Soma.',
                        setup_required: true,
                        setup_path: '/settings',
                    },
                }),
            });

            await store.getState().sendMissionChat('hello');

            expect(store.getState().activeMode).toBe('blocker');
            expect(store.getState().missionChatFailure).toMatchObject({
                routeKind: 'workspace',
                type: 'setup_required',
                bannerLabel: 'AI engine setup required',
                setupPath: '/settings',
            });
        });

        it('routes Soma failures through the workspace contract when no council target is selected', async () => {
            store.setState({
                councilTarget: 'admin',
                councilMembers: [],
            });
            mockFetch.mockRejectedValue(new Error('deadline exceeded'));

            await store.getState().sendMissionChat('hello');

            expect(mockFetch).toHaveBeenCalledWith('/api/v1/chat', expect.objectContaining({
                method: 'POST',
            }));
            expect(store.getState().activeMode).toBe('blocker');
            expect(store.getState().missionChatError).toBe('Soma did not return a response before the request deadline.');
            expect(store.getState().missionChatFailure).toMatchObject({
                routeKind: 'workspace',
                targetId: 'admin',
                type: 'timeout',
                title: 'Soma Chat Blocked',
            });
        });

        it('stores a structured council timeout when a direct council request throws', async () => {
            store.setState({ councilTarget: 'council-architect' });
            mockFetch.mockRejectedValue(new Error('deadline exceeded'));

            await store.getState().sendMissionChat('hello');

            expect(store.getState().activeMode).toBe('blocker');
            expect(store.getState().missionChatFailure).toMatchObject({
                routeKind: 'council',
                targetId: 'council-architect',
                type: 'timeout',
            });
        });

        it('routes direct council 5xx failures through the council blocker contract', async () => {
            store.setState({ councilTarget: 'council-coder' });
            mockFetch.mockResolvedValue({
                ok: false,
                status: 503,
                text: async () => '',
            });

            await store.getState().sendMissionChat('hello');

            expect(mockFetch).toHaveBeenCalledWith('/api/v1/council/council-coder/chat', expect.objectContaining({
                method: 'POST',
            }));
            expect(store.getState().activeMode).toBe('blocker');
            expect(store.getState().missionChatError).toBe('The council member service returned an internal error while handling the request.');
            expect(store.getState().missionChatFailure).toMatchObject({
                routeKind: 'council',
                targetId: 'council-coder',
                type: 'server_error',
                title: 'Council Call Failed',
            });
        });

        it('does not store raw JSON response text as the final Soma answer', async () => {
            mockFetch.mockResolvedValue({
                ok: true,
                json: async () => ({
                    ok: true,
                    data: {
                        meta: { source_node: 'admin', timestamp: new Date().toISOString() },
                        signal_type: 'chat_response',
                        trust_score: 0.5,
                        template_id: 'chat-to-answer',
                        mode: 'answer',
                        payload: {
                            text: '{"tool":"consult_council","status":"ok"}',
                            ask_class: 'direct_answer',
                        },
                    },
                }),
            });

            await store.getState().sendMissionChat('hello');

            expect(store.getState().missionChat.at(-1)?.content).toBe(
                'Soma could not produce a readable reply for that request. Retry or ask Soma to summarize the result directly.',
            );
        });
    });

    describe('confirmProposal', () => {
        it('keeps the proposal in a pending-proof state when confirmation succeeds without a run id', async () => {
            store.setState({
                pendingProposal: {
                    intent: 'Launch a docs crew',
                    teams: 1,
                    agents: 2,
                    tools: ['delegate_task'],
                    risk_level: 'medium',
                    confirm_token: 'ct-123',
                    intent_proof_id: 'ip-123',
                },
                activeConfirmToken: 'ct-123',
                missionChat: [{
                    role: 'council',
                    content: 'Proposed execution path',
                    mode: 'proposal',
                    proposal: {
                        intent: 'Launch a docs crew',
                        teams: 1,
                        agents: 2,
                        tools: ['delegate_task'],
                        risk_level: 'medium',
                        confirm_token: 'ct-123',
                        intent_proof_id: 'ip-123',
                    },
                    proposal_status: 'active',
                }],
                missionChatError: null,
                activeMode: 'proposal',
                activeRunId: null,
            });
            mockFetch.mockResolvedValue({
                ok: true,
                json: async () => ({ data: { confirmed: true, run_id: null } }),
            });

            const result = await store.getState().confirmProposal();

            expect(result).toEqual({ ok: true, runId: null });
            expect(store.getState().activeMode).toBe('proposal');
            expect(store.getState().activeRunId).toBeNull();
            expect(store.getState().pendingProposal).toBeNull();
            expect(store.getState().missionChat[0]).toMatchObject({
                proposal_status: 'confirmed_pending_execution',
                mode: 'proposal',
            });
            expect(store.getState().missionChat.at(-1)).toMatchObject({
                role: 'system',
                mode: 'proposal',
                content: 'Proposal confirmed. Waiting for execution proof.',
            });
        });

        it('records an execution result and run id on successful confirmation', async () => {
            store.setState({
                pendingProposal: {
                    intent: 'Launch a docs crew',
                    teams: 1,
                    agents: 2,
                    tools: ['delegate_task'],
                    risk_level: 'medium',
                    confirm_token: 'ct-123',
                    intent_proof_id: 'ip-123',
                },
                activeConfirmToken: 'ct-123',
                missionChat: [{
                    role: 'council',
                    content: 'Proposed execution path',
                    mode: 'proposal',
                    proposal: {
                        intent: 'Launch a docs crew',
                        teams: 1,
                        agents: 2,
                        tools: ['delegate_task'],
                        risk_level: 'medium',
                        confirm_token: 'ct-123',
                        intent_proof_id: 'ip-123',
                    },
                    proposal_status: 'active',
                }],
                missionChatError: null,
                activeMode: 'proposal',
                activeRunId: null,
            });
            mockFetch.mockResolvedValue({
                ok: true,
                json: async () => ({ data: { run_id: 'run-123' } }),
            });

            const result = await store.getState().confirmProposal();

            expect(result).toEqual({ ok: true, runId: 'run-123' });
            expect(store.getState().activeMode).toBe('execution_result');
            expect(store.getState().activeRunId).toBe('run-123');
            expect(store.getState().pendingProposal).toBeNull();
            expect(store.getState().missionChat[0]).toMatchObject({
                proposal_status: 'executed',
                mode: 'execution_result',
                run_id: 'run-123',
            });
            expect(store.getState().missionChat.at(-1)).toMatchObject({
                role: 'system',
                mode: 'execution_result',
                run_id: 'run-123',
            });
        });

        it('returns a blocker contract when confirmation fails', async () => {
            store.setState({
                pendingProposal: {
                    intent: 'Launch a docs crew',
                    teams: 1,
                    agents: 2,
                    tools: ['delegate_task'],
                    risk_level: 'medium',
                    confirm_token: 'ct-123',
                    intent_proof_id: 'ip-123',
                },
                activeConfirmToken: 'ct-123',
                missionChat: [{
                    role: 'council',
                    content: 'Proposed execution path',
                    mode: 'proposal',
                    proposal: {
                        intent: 'Launch a docs crew',
                        teams: 1,
                        agents: 2,
                        tools: ['delegate_task'],
                        risk_level: 'medium',
                        confirm_token: 'ct-123',
                        intent_proof_id: 'ip-123',
                    },
                    proposal_status: 'active',
                }],
                missionChatError: null,
                activeMode: 'proposal',
            });
            mockFetch.mockResolvedValue({
                ok: false,
                status: 500,
                text: async () => JSON.stringify({ error: 'confirmation denied' }),
            });

            const result = await store.getState().confirmProposal();

            expect(result).toEqual({
                ok: false,
                runId: null,
                error: 'Soma hit a server-side failure while handling the request.',
            });
            expect(store.getState().activeMode).toBe('blocker');
            expect(store.getState().missionChatError).toBe('Soma hit a server-side failure while handling the request.');
            expect(store.getState().missionChatFailure).toMatchObject({
                routeKind: 'workspace',
                type: 'server_error',
            });
            expect(store.getState().pendingProposal).toBeNull();
            expect(store.getState().missionChat[0]).toMatchObject({
                proposal_status: 'failed',
                mode: 'blocker',
            });
            expect(store.getState().missionChat.at(-1)).toMatchObject({
                role: 'council',
                mode: 'blocker',
                content: 'Soma hit a server-side failure while handling the request.',
            });
        });

        it('marks the proposal cancelled and appends a no-op system message when cancelled', () => {
            store.setState({
                pendingProposal: {
                    intent: 'Launch a docs crew',
                    teams: 1,
                    agents: 2,
                    tools: ['delegate_task'],
                    risk_level: 'medium',
                    confirm_token: 'ct-123',
                    intent_proof_id: 'ip-123',
                },
                activeConfirmToken: 'ct-123',
                missionChat: [{
                    role: 'council',
                    content: 'Proposed execution path',
                    mode: 'proposal',
                    proposal: {
                        intent: 'Launch a docs crew',
                        teams: 1,
                        agents: 2,
                        tools: ['delegate_task'],
                        risk_level: 'medium',
                        confirm_token: 'ct-123',
                        intent_proof_id: 'ip-123',
                    },
                    proposal_status: 'active',
                }],
                activeMode: 'proposal',
            });

            store.getState().cancelProposal();

            expect(store.getState().activeMode).toBe('answer');
            expect(store.getState().pendingProposal).toBeNull();
            expect(store.getState().activeConfirmToken).toBeNull();
            expect(store.getState().missionChat[0]).toMatchObject({
                proposal_status: 'cancelled',
            });
            expect(store.getState().missionChat.at(-1)).toMatchObject({
                role: 'system',
                content: 'Proposal cancelled. No action executed.',
            });
        });
    });

    describe('setMissionChatScope', () => {
        it('rehydrates scoped chat history when the organization scope changes', () => {
            localStorage.setItem('mycelis-workspace-chat:org-1', JSON.stringify([{ role: 'user', content: 'org-1 history' }]));
            localStorage.setItem('mycelis-workspace-chat:org-2', JSON.stringify([{ role: 'user', content: 'org-2 history' }]));

            store.getState().setMissionChatScope('org-1');
            expect(store.getState().workspaceChatScope).toBe('org-1');
            expect(store.getState().missionChat).toMatchObject([{ role: 'user', content: 'org-1 history' }]);

            store.getState().setMissionChatScope('org-2');
            expect(store.getState().workspaceChatScope).toBe('org-2');
            expect(store.getState().missionChat).toMatchObject([{ role: 'user', content: 'org-2 history' }]);
        });

        it('rehydrates a proof-pending proposal without promoting it to verified execution', () => {
            localStorage.setItem('mycelis-workspace-chat:org-proof', JSON.stringify([
                {
                    role: 'council',
                    content: 'Proposal confirmed. Waiting for execution proof.',
                    mode: 'proposal',
                    proposal: {
                        intent: 'Launch a docs crew',
                        teams: 1,
                        agents: 2,
                        tools: ['delegate_task'],
                        risk_level: 'medium',
                        confirm_token: 'ct-123',
                        intent_proof_id: 'ip-123',
                    },
                    proposal_status: 'executed',
                },
            ]));

            store.getState().setMissionChatScope('org-proof');

            expect(store.getState().workspaceChatScope).toBe('org-proof');
            expect(store.getState().activeMode).toBe('proposal');
            expect(store.getState().activeRunId).toBeNull();
            expect(store.getState().pendingProposal).toBeNull();
            expect(store.getState().missionChat[0]).toMatchObject({
                proposal_status: 'executed',
            });
        });
    });
});
