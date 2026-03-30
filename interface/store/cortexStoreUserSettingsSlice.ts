import { extractApiData } from '@/lib/apiContracts';
import type { CortexState } from '@/store/cortexStoreState';
import type { CortexGet, CortexSet } from '@/store/cortexStoreSliceTypes';
import type { CouncilMember } from '@/store/cortexStoreTypes';

type ThemeSetting = CortexState['theme'];

function normalizeTheme(value: unknown): ThemeSetting {
    if (value === 'midnight-cortex' || value === 'system') {
        return value;
    }
    return 'aero-light';
}

export function createCortexUserSettingsSlice(
    set: CortexSet,
    _get: CortexGet,
): Pick<CortexState, 'fetchUserSettings' | 'updateAssistantName' | 'updateTheme' | 'setCouncilTarget' | 'fetchCouncilMembers'> {
    const settingsEndpoint = '/api/v1/user/settings';

    return {
        fetchUserSettings: async () => {
            try {
                const res = await fetch(settingsEndpoint);
                if (!res.ok) return;
                const payload = await res.json();
                const data = extractApiData<Record<string, unknown> | unknown>(payload);
                const assistantName = typeof (data as Record<string, unknown>)?.assistant_name === 'string'
                    ? ((data as Record<string, unknown>).assistant_name as string).trim()
                    : '';
                const theme = normalizeTheme((data as Record<string, unknown>)?.theme);
                set({
                    theme,
                    ...(assistantName ? { assistantName } : {}),
                });
            } catch {
                // degraded mode — keep local defaults
            }
        },

        updateAssistantName: async (name: string) => {
            const trimmed = name.trim();
            if (!trimmed) return false;
            try {
                const res = await fetch(settingsEndpoint, {
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

        updateTheme: async (theme: ThemeSetting) => {
            const normalized = normalizeTheme(theme);
            try {
                const res = await fetch(settingsEndpoint, {
                    method: 'PUT',
                    headers: { 'Content-Type': 'application/json' },
                    body: JSON.stringify({ theme: normalized }),
                });
                const persisted = res.ok ? await res.json().then((body) => extractApiData<Record<string, unknown> | unknown>(body)) : null;
                set({ theme: normalizeTheme((persisted as Record<string, unknown> | null)?.theme ?? normalized) });
                return res.ok;
            } catch {
                set({ theme: normalized });
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
