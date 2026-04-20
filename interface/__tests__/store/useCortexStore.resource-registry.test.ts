import { beforeEach, describe, expect, it } from 'vitest';
import { useCortexStore } from '@/store/useCortexStore';
import { mockFetch } from '../setup';
import { resetCortexStore } from './useCortexStoreTestSupport';

describe('useCortexStore resource registry', () => {
    beforeEach(() => {
        resetCortexStore();
    });

    describe('fetchCatalogue', () => {
        it('stores catalogue agents from API', async () => {
            const agents = [
                { id: 'c1', name: 'Scanner', role: 'cognitive', tools: [], inputs: [], outputs: [], verification_rubric: [], created_at: '', updated_at: '' },
            ];
            mockFetch.mockResolvedValue({ ok: true, json: async () => agents });

            await useCortexStore.getState().fetchCatalogue();

            expect(useCortexStore.getState().catalogueAgents).toEqual(agents);
        });

        it('createCatalogueAgent adds to store', async () => {
            const created = { id: 'c1', name: 'New Agent', role: 'cognitive', tools: [], inputs: [], outputs: [], verification_rubric: [], created_at: '', updated_at: '' };
            mockFetch.mockResolvedValue({ ok: true, json: async () => created });

            await useCortexStore.getState().createCatalogueAgent({ name: 'New Agent', role: 'cognitive' });

            expect(useCortexStore.getState().catalogueAgents[0]).toEqual(created);
        });

        it('updateCatalogueAgent keeps the selected agent in sync', async () => {
            const existing = { id: 'c1', name: 'Scanner', role: 'cognitive', tools: [], inputs: [], outputs: [], verification_rubric: [], created_at: '', updated_at: '' };
            const updated = { ...existing, name: 'Scanner Prime' };
            useCortexStore.setState({
                catalogueAgents: [existing],
                selectedCatalogueAgent: existing,
            });
            mockFetch.mockResolvedValue({ ok: true, json: async () => updated });

            await useCortexStore.getState().updateCatalogueAgent('c1', { name: 'Scanner Prime' });

            expect(useCortexStore.getState().catalogueAgents[0]).toEqual(updated);
            expect(useCortexStore.getState().selectedCatalogueAgent).toEqual(updated);
        });

        it('deleteCatalogueAgent removes from store', async () => {
            useCortexStore.setState({
                catalogueAgents: [
                    { id: 'c1', name: 'A1', role: 'cognitive', tools: [], inputs: [], outputs: [], verification_rubric: [], created_at: '', updated_at: '' },
                    { id: 'c2', name: 'A2', role: 'sensory', tools: [], inputs: [], outputs: [], verification_rubric: [], created_at: '', updated_at: '' },
                ],
            });
            mockFetch.mockResolvedValue({ ok: true });

            await useCortexStore.getState().deleteCatalogueAgent('c1');

            expect(useCortexStore.getState().catalogueAgents).toHaveLength(1);
            expect(useCortexStore.getState().catalogueAgents[0].id).toBe('c2');
        });
    });

    describe('fetchMCPServers', () => {
        it('stores MCP servers from API', async () => {
            const servers = [
                { id: 'srv1', name: 'filesystem', transport: 'stdio', status: 'connected', created_at: '', tools: [] },
            ];
            mockFetch.mockResolvedValue({ ok: true, json: async () => servers });

            await useCortexStore.getState().fetchMCPServers();

            expect(useCortexStore.getState().mcpServers).toEqual(servers);
            expect(useCortexStore.getState().mcpServersError).toBeNull();
        });

        it('records MCP server fetch failures separately from empty registry state', async () => {
            mockFetch.mockResolvedValue({ ok: false, status: 500 });

            await useCortexStore.getState().fetchMCPServers();

            expect(useCortexStore.getState().mcpServers).toEqual([]);
            expect(useCortexStore.getState().mcpServersError).toContain('HTTP 500');
        });

        it('deleteMCPServer removes from store', async () => {
            useCortexStore.setState({
                mcpServers: [
                    { id: 'srv1', name: 'fs', transport: 'stdio' as const, status: 'connected', created_at: '', tools: [] },
                ],
            });
            mockFetch.mockResolvedValue({ ok: true });

            await useCortexStore.getState().deleteMCPServer('srv1');

            expect(useCortexStore.getState().mcpServers).toHaveLength(0);
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

            await useCortexStore.getState().fetchMCPActivity();

            expect(mockFetch).toHaveBeenCalledWith('/api/v1/mcp/activity?limit=12');
            expect(useCortexStore.getState().mcpActivity).toEqual(activity);
        });
    });

    describe('trust and governance state', () => {
        it('fetchTrustThreshold stores threshold from API', async () => {
            mockFetch.mockResolvedValue({
                ok: true,
                json: async () => ({ threshold: 0.85 }),
            });

            await useCortexStore.getState().fetchTrustThreshold();

            expect(useCortexStore.getState().trustThreshold).toBe(0.85);
        });

        it('setTrustThreshold updates store and calls API', async () => {
            mockFetch.mockResolvedValue({ ok: true });

            useCortexStore.getState().setTrustThreshold(0.9);

            expect(useCortexStore.getState().trustThreshold).toBe(0.9);
            expect(mockFetch).toHaveBeenCalledWith('/api/v1/trust/threshold', expect.objectContaining({
                method: 'PUT',
            }));
        });

        it('toggleSensorGroup adds group to subscribed list', () => {
            useCortexStore.getState().toggleSensorGroup('email');
            expect(useCortexStore.getState().subscribedSensorGroups).toContain('email');
        });

        it('toggleSensorGroup removes group if already subscribed', () => {
            useCortexStore.setState({ subscribedSensorGroups: ['email', 'weather'] });
            useCortexStore.getState().toggleSensorGroup('email');
            expect(useCortexStore.getState().subscribedSensorGroups).toEqual(['weather']);
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

            await useCortexStore.getState().fetchAuditLog();

            expect(mockFetch).toHaveBeenCalledWith('/api/v1/audit?limit=20');
            expect(useCortexStore.getState().auditLog).toEqual(auditLog);
        });
    });
});
