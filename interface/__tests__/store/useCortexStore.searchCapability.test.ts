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
            }],
        });
        expect(useCortexStore.getState().searchCapabilityError).toBeNull();
    });
});
