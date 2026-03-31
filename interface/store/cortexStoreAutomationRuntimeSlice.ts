import type { CortexState } from '@/store/cortexStoreState';
import type { CortexGet, CortexSet, CortexSlice } from '@/store/cortexStoreSliceTypes';
import type { TriggerRule, TriggerRuleCreate } from '@/store/cortexStoreTypes';

export function createCortexAutomationRuntimeSlice(
    set: CortexSet,
    get: CortexGet,
): CortexSlice<
    | 'fetchTriggerRules'
    | 'createTriggerRule'
    | 'updateTriggerRule'
    | 'deleteTriggerRule'
    | 'toggleTriggerRule'
    | 'fetchRunConversation'
    | 'interjectInRun'
> {
    return {
        fetchTriggerRules: async () => {
            set({ isFetchingTriggers: true });
            try {
                const res = await fetch('/api/v1/triggers');
                if (!res.ok) return;
                const body = await res.json();
                set({ triggerRules: body.data ?? [] });
            } catch {
                // degraded — silent
            } finally {
                set({ isFetchingTriggers: false });
            }
        },

        createTriggerRule: async (rule: TriggerRuleCreate) => {
            try {
                const res = await fetch('/api/v1/triggers', {
                    method: 'POST',
                    headers: { 'Content-Type': 'application/json' },
                    body: JSON.stringify(rule),
                });
                if (!res.ok) {
                    console.error('[TRIGGERS] Create failed:', await res.text());
                    return null;
                }
                const body = await res.json();
                const created: TriggerRule = body.data;
                set((s) => ({ triggerRules: [created, ...s.triggerRules] }));
                return created;
            } catch (err) {
                console.error('[TRIGGERS] Create error:', err);
                return null;
            }
        },

        updateTriggerRule: async (id: string, rule: TriggerRuleCreate) => {
            try {
                const res = await fetch(`/api/v1/triggers/${id}`, {
                    method: 'PUT',
                    headers: { 'Content-Type': 'application/json' },
                    body: JSON.stringify(rule),
                });
                if (!res.ok) {
                    console.error('[TRIGGERS] Update failed:', await res.text());
                    return;
                }
                get().fetchTriggerRules();
            } catch (err) {
                console.error('[TRIGGERS] Update error:', err);
            }
        },

        deleteTriggerRule: async (id: string) => {
            try {
                const res = await fetch(`/api/v1/triggers/${id}`, { method: 'DELETE' });
                if (!res.ok) {
                    console.error('[TRIGGERS] Delete failed:', await res.text());
                    return;
                }
                set((s) => ({
                    triggerRules: s.triggerRules.filter((item) => item.id !== id),
                }));
            } catch (err) {
                console.error('[TRIGGERS] Delete error:', err);
            }
        },

        toggleTriggerRule: async (id: string, isActive: boolean) => {
            try {
                const res = await fetch(`/api/v1/triggers/${id}/toggle`, {
                    method: 'POST',
                    headers: { 'Content-Type': 'application/json' },
                    body: JSON.stringify({ is_active: isActive }),
                });
                if (!res.ok) {
                    console.error('[TRIGGERS] Toggle failed:', await res.text());
                    return;
                }
                set((s) => ({
                    triggerRules: s.triggerRules.map((item) =>
                        item.id === id ? { ...item, is_active: isActive } : item
                    ),
                }));
            } catch (err) {
                console.error('[TRIGGERS] Toggle error:', err);
            }
        },

        fetchRunConversation: async (runId: string, agentFilter?: string) => {
            set({ isFetchingConversation: true });
            try {
                const params = new URLSearchParams();
                if (agentFilter) params.set('agent', agentFilter);
                const qs = params.toString();
                const url = `/api/v1/runs/${runId}/conversation${qs ? `?${qs}` : ''}`;
                const res = await fetch(url);
                if (!res.ok) {
                    console.error('[CONVERSATION] Fetch failed:', res.status);
                    return;
                }
                const body = await res.json();
                const turns = body.data?.turns ?? body.data ?? [];
                set({ conversationTurns: Array.isArray(turns) ? turns : [] });
            } catch (err) {
                console.error('[CONVERSATION] Fetch error:', err);
            } finally {
                set({ isFetchingConversation: false });
            }
        },

        interjectInRun: async (runId: string, message: string, agentId?: string) => {
            try {
                const payload: Record<string, string> = { message };
                if (agentId) payload.agent_id = agentId;
                const res = await fetch(`/api/v1/runs/${runId}/interject`, {
                    method: 'POST',
                    headers: { 'Content-Type': 'application/json' },
                    body: JSON.stringify(payload),
                });
                if (!res.ok) {
                    console.error('[CONVERSATION] Interject failed:', await res.text());
                    return;
                }
                get().fetchRunConversation(runId);
            } catch (err) {
                console.error('[CONVERSATION] Interject error:', err);
            }
        },
    };
}
