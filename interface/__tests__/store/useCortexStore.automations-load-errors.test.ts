import { beforeEach, describe, expect, it } from 'vitest';
import { useCortexStore } from '@/store/useCortexStore';
import { mockFetch } from '../setup';
import { resetCortexStore } from './useCortexStoreTestSupport';

describe('useCortexStore automations load errors', () => {
    beforeEach(() => {
        resetCortexStore();
    });

    describe('fetchTriggerRules', () => {
        it('sets triggerRulesLoadError on network failure and clears triggerRules-affecting state on retry', async () => {
            mockFetch.mockRejectedValueOnce(new Error('network down'));

            await useCortexStore.getState().fetchTriggerRules();

            expect(useCortexStore.getState().triggerRulesLoadError).toMatch(/did not respond/);

            mockFetch.mockResolvedValueOnce({ ok: true, json: async () => ({ data: [] }) });

            await useCortexStore.getState().fetchTriggerRules();

            expect(useCortexStore.getState().triggerRulesLoadError).toBeNull();
            expect(useCortexStore.getState().triggerRules).toEqual([]);
        });

        it('sets triggerRulesLoadError on a non-OK response', async () => {
            mockFetch.mockResolvedValueOnce({ ok: false, status: 500 });

            await useCortexStore.getState().fetchTriggerRules();

            expect(useCortexStore.getState().triggerRulesLoadError).toMatch(/HTTP 500/);
        });

        it('clears a prior error once the fetch succeeds', async () => {
            useCortexStore.setState({ triggerRulesLoadError: 'stale error' });
            mockFetch.mockResolvedValueOnce({ ok: true, json: async () => ({ data: [] }) });

            await useCortexStore.getState().fetchTriggerRules();

            expect(useCortexStore.getState().triggerRulesLoadError).toBeNull();
        });
    });

    describe('fetchPendingApprovals', () => {
        it('sets approvalsLoadError on network failure then clears it on success', async () => {
            mockFetch.mockRejectedValueOnce(new Error('network down'));

            await useCortexStore.getState().fetchPendingApprovals();

            expect(useCortexStore.getState().approvalsLoadError).toMatch(/did not respond/);

            mockFetch.mockResolvedValueOnce({ ok: true, json: async () => [] });

            await useCortexStore.getState().fetchPendingApprovals();

            expect(useCortexStore.getState().approvalsLoadError).toBeNull();
            expect(useCortexStore.getState().pendingApprovals).toEqual([]);
        });

        it('sets approvalsLoadError on a non-OK response', async () => {
            mockFetch.mockResolvedValueOnce({ ok: false, status: 503 });

            await useCortexStore.getState().fetchPendingApprovals();

            expect(useCortexStore.getState().approvalsLoadError).toMatch(/HTTP 503/);
        });
    });

    describe('fetchAuditLog', () => {
        it('sets auditLogLoadError on network failure then clears it on success', async () => {
            mockFetch.mockRejectedValueOnce(new Error('network down'));

            await useCortexStore.getState().fetchAuditLog();

            expect(useCortexStore.getState().auditLogLoadError).toMatch(/did not respond/);

            mockFetch.mockResolvedValueOnce({ ok: true, json: async () => ({ data: [] }) });

            await useCortexStore.getState().fetchAuditLog();

            expect(useCortexStore.getState().auditLogLoadError).toBeNull();
            expect(useCortexStore.getState().auditLog).toEqual([]);
        });
    });

    describe('fetchPolicy', () => {
        it('sets policyLoadError on network failure then clears it on success', async () => {
            mockFetch.mockRejectedValueOnce(new Error('network down'));

            await useCortexStore.getState().fetchPolicy();

            expect(useCortexStore.getState().policyLoadError).toMatch(/did not respond/);

            const config = { groups: [], defaults: { default_action: 'DENY' } };
            mockFetch.mockResolvedValueOnce({ ok: true, json: async () => config });

            await useCortexStore.getState().fetchPolicy();

            expect(useCortexStore.getState().policyLoadError).toBeNull();
            expect(useCortexStore.getState().policyConfig).toEqual(config);
        });
    });
});
