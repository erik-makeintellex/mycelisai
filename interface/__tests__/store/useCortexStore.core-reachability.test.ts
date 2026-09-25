import { beforeEach, describe, expect, it } from 'vitest';
import { selectCoreReachability } from '@/lib/coreReachability';
import { useCortexStore } from '@/store/useCortexStore';
import { mockFetch } from '../setup';
import { resetCortexStore } from './useCortexStoreTestSupport';

describe('useCortexStore core reachability', () => {
    beforeEach(() => {
        resetCortexStore();
        useCortexStore.setState({ servicesStatusFailure: null, servicesStatusUpdatedAt: null, runsLoadError: null });
    });

    it('marks Core unreachable when the services-status proxy returns 500, and clears it on recovery', async () => {
        mockFetch.mockResolvedValue({ ok: false, status: 500, json: async () => ({}) });
        await useCortexStore.getState().fetchServicesStatus();
        expect(selectCoreReachability(useCortexStore.getState())).toBe('unreachable');

        mockFetch.mockResolvedValue({
            ok: true,
            json: async () => ({ ok: true, data: [{ name: 'nats', status: 'online' }] }),
        });
        await useCortexStore.getState().fetchServicesStatus();
        expect(useCortexStore.getState().servicesStatusFailure).toBeNull();
        expect(selectCoreReachability(useCortexStore.getState())).not.toBe('unreachable');
    });

    it('marks Core unreachable on a network error', async () => {
        mockFetch.mockRejectedValue(new TypeError('Failed to fetch'));
        await useCortexStore.getState().fetchServicesStatus();
        expect(selectCoreReachability(useCortexStore.getState())).toBe('unreachable');
    });

    it('records runsLoadError on failed runs fetch and clears it on success', async () => {
        mockFetch.mockResolvedValue({ ok: false, status: 500, json: async () => ({}) });
        await useCortexStore.getState().fetchRecentRuns();
        expect(useCortexStore.getState().runsLoadError).toMatch(/HTTP 500/);

        mockFetch.mockRejectedValue(new TypeError('Failed to fetch'));
        await useCortexStore.getState().fetchRecentRuns();
        expect(useCortexStore.getState().runsLoadError).toMatch(/did not respond/);

        mockFetch.mockResolvedValue({ ok: true, json: async () => ({ data: [] }) });
        await useCortexStore.getState().fetchRecentRuns();
        expect(useCortexStore.getState().runsLoadError).toBeNull();
        expect(useCortexStore.getState().recentRuns).toEqual([]);
    });
});
