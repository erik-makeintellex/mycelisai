import { beforeEach, describe, expect, it, vi } from 'vitest';
import { useCortexStore } from '@/store/useCortexStore';
import { mockFetch } from '../setup';
import { resetCortexStore } from './useCortexStoreTestSupport';

// MCPA W2 (MCP configuration authority): store handling of Core blockers for
// install, delete and toolset writes. Split from the resource-registry suite
// to keep both files under the 385-line policy.
describe('useCortexStore MCP authority blockers', () => {
    beforeEach(() => {
        resetCortexStore();
    });

        it('deleteMCPServer keeps the server on a 403 and surfaces the blocker code', async () => {
            useCortexStore.setState({
                mcpServers: [
                    { id: 'srv1', name: 'fs', transport: 'stdio' as const, status: 'connected', created_at: '', tools: [] },
                ],
            });
            mockFetch.mockResolvedValue({
                ok: false,
                status: 403,
                json: async () => ({ ok: false, error: 'admin required', data: { code: 'admin_required' } }),
            });

            const result = await useCortexStore.getState().deleteMCPServer('srv1');

            expect(result).toEqual({ ok: false, code: 'admin_required', httpStatus: 403 });
            expect(useCortexStore.getState().mcpServers).toHaveLength(1);
        });

        it('deleteMCPServer surfaces mcp_server_not_found on a 404', async () => {
            mockFetch.mockResolvedValue({
                ok: false,
                status: 404,
                json: async () => ({ ok: false, error: '', data: { code: 'mcp_server_not_found' } }),
            });

            const result = await useCortexStore.getState().deleteMCPServer('srv-missing');

            expect(result).toEqual({ ok: false, code: 'mcp_server_not_found', httpStatus: 404 });
        });

        it('createMCPToolSet surfaces admin_required on a 403 without refreshing', async () => {
            mockFetch.mockResolvedValue({
                ok: false,
                status: 403,
                json: async () => ({ ok: false, error: 'admin required', data: { code: 'admin_required', required_scope: 'mcp_config:write' } }),
            });

            const result = await useCortexStore.getState().createMCPToolSet({
                name: 'deploy',
                tool_refs: ['mcp:ssh/*'],
                scope_kind: 'all',
            });

            expect(result).toEqual({ ok: false, code: 'admin_required', httpStatus: 403 });
            expect(mockFetch).toHaveBeenCalledTimes(1);
        });

        it('installFromLibrary proceeds to install even when inspect answers require_approval', async () => {
            mockFetch
                .mockResolvedValueOnce({ ok: true, json: async () => ({ decision: 'require_approval', governance: { decision: 'require_approval', approval_required: true } }) })
                .mockResolvedValueOnce({ ok: true, json: async () => ({ status: 'installed', self_approved: true, audit_event_id: 'audit-1' }) });

            const result = await useCortexStore.getState().installFromLibrary('github', { GITHUB_PERSONAL_ACCESS_TOKEN: 'x' });

            expect(result.ok).toBe(true);
            expect(mockFetch).toHaveBeenCalledTimes(4); // inspect, install, then fetchMCPServers + fetchMCPActivity
            expect(mockFetch).toHaveBeenNthCalledWith(2, '/api/v1/mcp/library/install', expect.objectContaining({ method: 'POST' }));
        });

        it('installFromLibrary falls back to request_failed for a respondAPIError envelope with no data.code', async () => {
            mockFetch
                .mockResolvedValueOnce({ ok: true, json: async () => ({ decision: 'allow' }) })
                .mockResolvedValueOnce({ ok: false, status: 503, json: async () => ({ ok: false, error: 'MCP subsystem not initialized' }) });

            const result = await useCortexStore.getState().installFromLibrary('fetch');

            expect(result).toEqual({ ok: false, code: 'request_failed', httpStatus: 503, governance: undefined });
        });

        it('installFromLibrary surfaces a blocker from a rejected inspect call', async () => {
            mockFetch.mockResolvedValueOnce({
                ok: false,
                status: 403,
                json: async () => ({ ok: false, error: 'x', data: { code: 'admin_required' } }),
            });

            const result = await useCortexStore.getState().installFromLibrary('fetch');

            expect(result).toEqual({ ok: false, code: 'admin_required', httpStatus: 403 });
        });
});
