import { beforeEach, describe, expect, it } from 'vitest';
import { useCortexStore } from '@/store/useCortexStore';
import { mockFetch } from '../setup';
import { resetCortexStore } from './useCortexStoreTestSupport';

describe('useCortexStore search capability', () => {
    beforeEach(() => {
        resetCortexStore();
    });

    it('stores Mycelis Search capability status from API', async () => {
        mockFetch.mockResolvedValue({
            ok: true,
            json: async () => ({ ok: true, data: {
                provider: 'searxng',
                enabled: true,
                configured: true,
                supports_local_sources: false,
                supports_public_web: true,
                soma_tool_name: 'web_search',
                direct_soma_interaction: true,
                requires_hosted_api_token: false,
                max_results: 8,
                sources: [{
                    id: 'searxng',
                    display_name: 'Self-hosted web',
                    type: 'public_web',
                    scope_kind: 'all',
                    description: 'Self-hosted SearXNG endpoint',
                    auth: 'none',
                    sensitivity_class: 'public',
                    trust_class: 'bounded_external',
                }, {
                    id: 'workspace-code',
                    display_name: 'Workspace code map',
                    provider: 'code_context',
                    type: 'code_context',
                    scope_kind: 'host',
                    scope_ref: 'dev-workstation',
                    description: 'Approved workspace repository source',
                    auth: 'none',
                    sensitivity_class: 'governed',
                    trust_class: 'trusted_internal',
                    code_context: {
                        scope: 'workspace repository',
                        snapshot: {
                            status: 'ready',
                            id: 'snapshot:workspace-code:abc123',
                            digest: 'sha256:abc123',
                            updated_at: '2026-08-22T10:00:00Z',
                        },
                        index: {
                            status: 'stale',
                            id: 'index:workspace-code:def456',
                            digest: 'sha256:def456',
                            updated_at: '2026-08-21T09:30:00Z',
                        },
                    },
                }],
            } }),
        });

        await useCortexStore.getState().fetchSearchCapability();

        expect(mockFetch).toHaveBeenCalledWith('/api/v1/search/status');
        expect(useCortexStore.getState().searchCapability).toMatchObject({
            provider: 'searxng',
            sources: [{
                id: 'searxng',
                name: 'Self-hosted web',
                source_type: 'public_web',
                boundary: 'Self-hosted SearXNG endpoint',
                auth_scheme: 'none',
                mode: 'live',
                status: 'available',
            }, {
                id: 'workspace-code',
                name: 'Workspace code map',
                source_type: 'code_context',
                provider: 'code_context',
                code_context: {
                    scope: 'workspace repository',
                    snapshot_status: 'ready',
                    snapshot_ref: 'snapshot:workspace-code:abc123',
                    snapshot_digest: 'sha256:abc123',
                    last_snapshot_at: '2026-08-22T10:00:00Z',
                    index_status: 'stale',
                    index_ref: 'index:workspace-code:def456',
                    index_digest: 'sha256:def456',
                    last_indexed_at: '2026-08-21T09:30:00Z',
                },
            }],
        });
        expect(useCortexStore.getState().searchCapabilityError).toBeNull();
    });
});
