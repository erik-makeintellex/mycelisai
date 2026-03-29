import { extractApiData } from '@/lib/apiContracts';
import type { CortexState } from '@/store/cortexStoreState';
import type { CortexGet, CortexSet } from '@/store/cortexStoreSliceTypes';
import type { CouncilMember } from '@/store/cortexStoreTypes';

export function createCortexUserSettingsSlice(
    set: CortexSet,
    _get: CortexGet,
): Pick<CortexState, 'fetchUserSettings' | 'updateAssistantName' | 'setCouncilTarget' | 'fetchCouncilMembers'> {
    return {
        fetchUserSettings: async () => {
            try {
                const res = await fetch('/api/v1/settings/user');
                if (!res.ok) return;
                const payload = await res.json();
                const data = extractApiData<Record<string, unknown> | unknown>(payload);
                const assistantName = typeof (data as Record<string, unknown>)?.assistant_name === 'string'
                    ? ((data as Record<string, unknown>).assistant_name as string).trim()
                    : '';
                if (assistantName) {
                    set({ assistantName });
                }
            } catch {
                // degraded mode — keep local defaults
            }
        },

        updateAssistantName: async (name: string) => {
            const trimmed = name.trim();
            if (!trimmed) return false;
            try {
                const res = await fetch('/api/v1/settings/user', {
                    method: 'PUT',
                    headers: { 'Content-Type': 'application/json' },
                    body: JSON.stringify({ assistant_name: trimmed }),
                });
                const persisted = res.ok ? await res.json().then((body) => extractApiData<Record<string, unknown> | unknown>(body)) : null;
                const assistantName = typeof (persisted as Record<string, unknown> | null)?.assistant_name === 'string'
                    ? (((persisted as Record<string, unknown>).assistant_name as string).trim() || trimmed)
                    : trimmed;
                set({ assistantName });
                return res.ok;
            } catch {
                set({ assistantName: trimmed });
                return false;
            }
        },

        setCouncilTarget: (id: string) => {
            set({ councilTarget: id });
        },

        fetchCouncilMembers: async () => {
            try {
                const res = await fetch('/api/v1/council/members');
                if (!res.ok) return;
                const body = await res.json();
                if (Array.isArray(body.data)) {
                    set({ councilMembers: body.data as CouncilMember[] });
                }
            } catch {
                // degraded mode — keep local state unchanged
            }
        },
    };
}
