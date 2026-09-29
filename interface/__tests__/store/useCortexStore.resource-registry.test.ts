import { beforeEach, describe, expect, it, vi } from 'vitest';
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

            const result = await useCortexStore.getState().deleteMCPServer('srv1');

            expect(result).toEqual({ ok: true });
            expect(useCortexStore.getState().mcpServers).toHaveLength(0);
        });

        // MCPA D10: honest failure -- a rejected delete surfaces the
        // blocker envelope's code and never touches local state, so a 403
        // can never be mistaken for a removed server.
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

        it('stores scoped MCP access layers from API', async () => {
            const toolSets = [
                {
                    id: 'set-workspace',
                    name: 'workspace',
                    tool_refs: ['mcp:filesystem/*'],
                    scope_kind: 'all',
                },
                {
                    id: 'set-host',
                    name: 'deploy',
                    tool_refs: ['mcp:ssh/*'],
                    scope_kind: 'host',
                    scope_ref: 'edge-node-1',
                },
                { id: 'set-group', name: 'research', tool_refs: ['mcp:fetch/fetch', 'tool:web_search'], scope_kind: 'group', scope_ref: 'market-research' },
            ];
            mockFetch.mockResolvedValue({ ok: true, json: async () => ({ ok: true, data: toolSets }) });

            await useCortexStore.getState().fetchMCPToolSets();

            expect(mockFetch).toHaveBeenCalledWith('/api/v1/mcp/toolsets');
            expect(useCortexStore.getState().mcpToolSets).toEqual(toolSets);
            expect(useCortexStore.getState().mcpToolSets.map((set) => set.scope_kind)).toEqual(['all', 'host', 'group']);
            expect(useCortexStore.getState().mcpToolSetsError).toBeNull();
        });

        it('creates a scoped MCP access layer and refreshes the list', async () => {
            mockFetch
                .mockResolvedValueOnce({ ok: true, json: async () => ({ ok: true, data: { id: 'set-host' } }) })
                .mockResolvedValueOnce({ ok: true, json: async () => ({ ok: true, data: [] }) });

            const result = await useCortexStore.getState().createMCPToolSet({
                name: 'deploy',
                description: 'Deployment tools',
                tool_refs: ['mcp:ssh/*'],
                scope_kind: 'host',
                scope_ref: 'edge-node-1',
            });

            expect(result).toEqual({ ok: true });
            expect(mockFetch).toHaveBeenNthCalledWith(1, '/api/v1/mcp/toolsets', expect.objectContaining({
                method: 'POST',
                body: JSON.stringify({
                    name: 'deploy',
                    description: 'Deployment tools',
                    tool_refs: ['mcp:ssh/*'],
                    scope_kind: 'host',
                    scope_ref: 'edge-node-1',
                }),
            }));
            expect(mockFetch).toHaveBeenNthCalledWith(2, '/api/v1/mcp/toolsets');
        });

        // MCPA D10: admin-only, fail-closed. A 403 never calls the refresh
        // fetch and surfaces the blocker code, not raw backend text.
        // MCPA D5/D10: no dead end. A require_approval inspect decision no
        // longer blocks the client; install still proceeds to /install and
        // succeeds (the server enforces approvals:decide, not the browser).
        // MCPA D10: honest failure on inspect/install non-2xx -- no raw
        // backend text, and no success toast.
        it.each([
            ['admin_required', 403],
            ['mcp_env_rejected', 400],
            ['mcp_connect_failed', 502],
            ['service_unavailable', 503],
        ])('installFromLibrary surfaces %s on a %i from install', async (code, status) => {
            mockFetch
                .mockResolvedValueOnce({ ok: true, json: async () => ({ decision: 'allow' }) })
                .mockResolvedValueOnce({ ok: false, status, json: async () => ({ ok: false, error: 'x', data: { code } }) });

            const result = await useCortexStore.getState().installFromLibrary('fetch');

            expect(result).toEqual({ ok: false, code, httpStatus: status, governance: undefined });
        });

        // Slice RED: library inspect/install/apply "not initialized" / bad
        // JSON / unknown-entry error sites moved from http.Error text/plain
        // to respondAPIError's application/json {ok:false,error} envelope
        // (no data.code). The store parses by body content, not
        // Content-Type, so both shapes land on the same honest fallback.
        it('records Mycelis Search capability status failures', async () => {
            mockFetch.mockResolvedValue({ ok: false, status: 503 });

            await useCortexStore.getState().fetchSearchCapability();

            expect(useCortexStore.getState().searchCapability).toBeNull();
            expect(useCortexStore.getState().searchCapabilityError).toContain('HTTP 503');
        });

        it('normalizes capability manifests without exposing tool refs as outputs', async () => {
            mockFetch.mockResolvedValue({
                ok: true,
                json: async () => ({
                    ok: true,
                    data: {
                        capabilities: [{
                            id: 'filesystem.read',
                            display_name: 'Read workspace files',
                            source: 'mcp',
                            kind: 'filesystem',
                            risk_class: 'low',
                            approval_required: false,
                            audit_required: true,
                            output_schema_ref: 'schemas/filesystem/FileReadResult',
                            tool_refs: ['mcp:filesystem/read_file'],
                            default_allowed_roles: ['soma'],
                            metadata: { server_name: 'filesystem', provider: 'mcp' },
                        }],
                    },
                }),
            });

            await useCortexStore.getState().fetchCapabilities();

            expect(mockFetch).toHaveBeenCalledWith('/api/v1/capabilities');
            expect(useCortexStore.getState().capabilities[0]).toMatchObject({
                name: 'Read workspace files',
                outputs: ['FileReadResult output'],
                bound_server_name: 'filesystem',
                bound_tool_name: 'read_file',
            });
        });
    });

    describe('trust and governance state', () => {
        it('carries no V7 Overseer trust-threshold state or actions (CONS-C3)', () => {
            const state = useCortexStore.getState() as unknown as Record<string, unknown>;
            for (const key of ['trustThreshold', 'isSyncingThreshold', 'setTrustThreshold', 'fetchTrustThreshold']) {
                expect(key in state, key).toBe(false);
            }
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

        it('fetchAuditLog records a 403 as an error and keeps the list, not a fake empty state', async () => {
            const consoleError = vi.spyOn(console, 'error').mockImplementation(() => {});
            useCortexStore.setState({ auditLog: [{ id: 'stale' } as never] });
            mockFetch.mockResolvedValue({
                ok: false,
                status: 403,
                json: async () => ({ ok: false, error: 'Root admin role required' }),
            });

            await useCortexStore.getState().fetchAuditLog();

            expect(useCortexStore.getState().auditLogError).toEqual({ code: undefined, httpStatus: 403 });
            expect(useCortexStore.getState().auditLog).toEqual([{ id: 'stale' }]);
            expect(useCortexStore.getState().isFetchingAuditLog).toBe(false);
            expect(consoleError).not.toHaveBeenCalled();
            consoleError.mockRestore();
        });
    });
});
