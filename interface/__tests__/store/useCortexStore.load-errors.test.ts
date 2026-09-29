import { beforeEach, describe, expect, it } from 'vitest';
import { useCortexStore } from '@/store/useCortexStore';
import { mockFetch } from '../setup';
import { resetCortexStore } from './useCortexStoreTestSupport';

// Honest load state: a failed load records {code, httpStatus} and never
// presents an empty list as data. A successful load clears the error.

const rule = { id: 'r1', name: 'Rule' } as never;
const entry = { id: 'a1', actor: 'Soma', action: 'x', result_status: 'ok' } as never;

const failures = [
    { label: '500', res: { ok: false, status: 500, json: async () => ({ data: { code: 'core_unavailable' } }) }, code: 'core_unavailable', httpStatus: 500 },
    { label: '403', res: { ok: false, status: 403, json: async () => ({ data: { code: 'admin_required' } }) }, code: 'admin_required', httpStatus: 403 },
    { label: '403 with unparseable body', res: { ok: false, status: 403, json: async () => { throw new Error('bad'); } }, code: undefined, httpStatus: 403 },
];

type Load = {
    name: string;
    run: () => Promise<void>;
    listKey: 'triggerRules' | 'auditLog' | 'teamRoster';
    errorKey: 'triggerRulesError' | 'auditLogError' | 'teamRosterError';
    stale: unknown;
    ok: () => void;
    pendingKey: 'isFetchingTriggers' | 'isFetchingAuditLog' | 'isFetchingTeamRoster';
};

const loads: Load[] = [
    {
        name: 'fetchTriggerRules',
        run: () => useCortexStore.getState().fetchTriggerRules(),
        listKey: 'triggerRules', errorKey: 'triggerRulesError', stale: [rule], pendingKey: 'isFetchingTriggers',
        ok: () => mockFetch.mockResolvedValue({ ok: true, json: async () => ({ data: [rule] }) }),
    },
    {
        name: 'fetchAuditLog',
        run: () => useCortexStore.getState().fetchAuditLog(),
        listKey: 'auditLog', errorKey: 'auditLogError', stale: [entry], pendingKey: 'isFetchingAuditLog',
        ok: () => mockFetch.mockResolvedValue({ ok: true, json: async () => ({ ok: true, data: [entry] }) }),
    },
    {
        name: 'fetchTeamDetails',
        run: () => useCortexStore.getState().fetchTeamDetails(),
        listKey: 'teamRoster', errorKey: 'teamRosterError', stale: [{ id: 't1', name: 'T', role: 'observer', agents: [] }], pendingKey: 'isFetchingTeamRoster',
        ok: () => mockFetch.mockImplementation(async (url: string) =>
            url === '/agents'
                ? { ok: true, json: async () => ({ agents: [] }) }
                : { ok: true, json: async () => [{ id: 't1', name: 'T' }] }),
    },
];

describe.each(loads)('$name honest load state', (load) => {
    beforeEach(() => {
        resetCortexStore();
    });

    it.each(failures)('records an error for HTTP $label and does not present an empty list as data', async ({ res, code, httpStatus }) => {
        useCortexStore.setState({ [load.listKey]: load.stale } as never);
        mockFetch.mockResolvedValue(res);

        await load.run();

        const state = useCortexStore.getState();
        expect(state[load.errorKey]).toEqual({ code, httpStatus });
        expect(state[load.listKey]).toEqual(load.stale);
        expect(state[load.pendingKey]).toBe(false);
    });

    it('records an error for a network failure', async () => {
        mockFetch.mockRejectedValue(new Error('network down'));

        await load.run();

        const state = useCortexStore.getState();
        expect(state[load.errorKey]).toEqual({ httpStatus: undefined });
        expect(state[load.pendingKey]).toBe(false);
    });

    it('clears the error on a later success', async () => {
        mockFetch.mockResolvedValue(failures[0].res);
        await load.run();
        expect(useCortexStore.getState()[load.errorKey]).not.toBeNull();

        load.ok();
        await load.run();

        const state = useCortexStore.getState();
        expect(state[load.errorKey]).toBeNull();
        expect((state[load.listKey] as unknown[]).length).toBe(1);
    });
});

describe('fetchTeamDetails partial failure', () => {
    beforeEach(() => resetCortexStore());

    it('records an error when only the agents request fails', async () => {
        mockFetch.mockImplementation(async (url: string) =>
            url === '/agents'
                ? { ok: false, status: 503, json: async () => ({}) }
                : { ok: true, json: async () => [{ id: 't1', name: 'T' }] });

        await useCortexStore.getState().fetchTeamDetails();

        expect(useCortexStore.getState().teamRosterError).toEqual({ code: undefined, httpStatus: 503 });
    });
});
