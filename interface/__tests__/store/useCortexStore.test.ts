import { describe, it, expect, vi, beforeEach } from 'vitest';
import { useCortexStore } from '@/store/useCortexStore';
import { mockFetch } from '../setup';

// Direct store testing — no React rendering needed
const store = useCortexStore;

describe('useCortexStore', () => {
    beforeEach(() => {
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
            mcpServers: [],
            isFetchingMCPServers: false,
            mcpTools: [],
            trustThreshold: 0.7,
            isSyncingThreshold: false,
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
    });
});
