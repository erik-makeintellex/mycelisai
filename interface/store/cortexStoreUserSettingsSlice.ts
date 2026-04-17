import { extractApiData } from '@/lib/apiContracts';
import type { CortexState } from '@/store/cortexStoreState';
import type { CortexGet, CortexSet, CortexSlice } from '@/store/cortexStoreSliceTypes';
import type { CouncilMember } from '@/store/cortexStoreTypes';

type ThemeSetting = CortexState['theme'];
type LocalSettings = {
    assistantName?: string;
    theme?: ThemeSetting;
};

const LOCAL_SETTINGS_KEY = 'mycelis-user-settings';

function normalizeTheme(value: unknown): ThemeSetting {
    if (value === 'midnight-cortex' || value === 'system') {
        return value;
    }
    return 'aero-light';
}

function readLocalSettings(): LocalSettings {
    if (typeof window === 'undefined') {
        return {};
    }
    try {
        const raw = window.localStorage.getItem(LOCAL_SETTINGS_KEY);
        if (!raw) {
            return {};
        }
        const parsed = JSON.parse(raw) as Record<string, unknown>;
        const assistantName = typeof parsed.assistantName === 'string' ? parsed.assistantName.trim() : '';
        return {
            ...(assistantName ? { assistantName } : {}),
            theme: normalizeTheme(parsed.theme),
        };
    } catch {
        return {};
    }
}

function persistLocalSettings(patch: LocalSettings) {
    if (typeof window === 'undefined') {
        return;
    }
    const current = readLocalSettings();
    const nextSettings: LocalSettings = {
        ...current,
        ...patch,
    };
    try {
        window.localStorage.setItem(LOCAL_SETTINGS_KEY, JSON.stringify(nextSettings));
    } catch {
        // Ignore local persistence failures and keep the in-memory state.
    }
}

export function createCortexUserSettingsSlice(
    set: CortexSet,
    _get: CortexGet,
): CortexSlice<'fetchUserSettings' | 'updateAssistantName' | 'updateTheme' | 'setCouncilTarget' | 'fetchCouncilMembers'> {
    const settingsEndpoint = '/api/v1/user/settings';

    return {
        fetchUserSettings: async () => {
            try {
                const res = await fetch(settingsEndpoint);
                if (!res.ok) {
                    const fallback = readLocalSettings();
                    if (fallback.assistantName || fallback.theme) {
                        set({
                            ...(fallback.assistantName ? { assistantName: fallback.assistantName } : {}),
                            ...(fallback.theme ? { theme: fallback.theme } : {}),
                        });
                    }
                    return;
                }
                const payload = await res.json();
                const data = extractApiData<Record<string, unknown> | unknown>(payload);
                const assistantName = typeof (data as Record<string, unknown>)?.assistant_name === 'string'
                    ? ((data as Record<string, unknown>).assistant_name as string).trim()
                    : '';
                const theme = normalizeTheme((data as Record<string, unknown>)?.theme);
                persistLocalSettings({
                    ...(assistantName ? { assistantName } : {}),
                    theme,
                });
                set({
                    theme,
                    ...(assistantName ? { assistantName } : {}),
                });
            } catch {
                const fallback = readLocalSettings();
                if (fallback.assistantName || fallback.theme) {
                    set({
                        ...(fallback.assistantName ? { assistantName: fallback.assistantName } : {}),
                        ...(fallback.theme ? { theme: fallback.theme } : {}),
                    });
                }
            }
        },

        updateAssistantName: async (name: string) => {
            const trimmed = name.trim();
            if (!trimmed) return false;
            const persistFallback = () => {
                persistLocalSettings({ assistantName: trimmed });
                set({ assistantName: trimmed });
                return true;
            };
            try {
                const res = await fetch(settingsEndpoint, {
                    method: 'PUT',
                    headers: { 'Content-Type': 'application/json' },
                    body: JSON.stringify({ assistant_name: trimmed }),
                });
                if (!res.ok) {
                    return persistFallback();
                }
                const persisted = await res.json().then((body) => extractApiData<Record<string, unknown> | unknown>(body));
                const assistantName = typeof (persisted as Record<string, unknown> | null)?.assistant_name === 'string'
                    ? (((persisted as Record<string, unknown>).assistant_name as string).trim() || trimmed)
                    : trimmed;
                persistLocalSettings({ assistantName });
                set({ assistantName });
                return true;
            } catch {
                return persistFallback();
            }
        },

        updateTheme: async (theme: ThemeSetting) => {
            const normalized = normalizeTheme(theme);
            const persistFallback = () => {
                persistLocalSettings({ theme: normalized });
                set({ theme: normalized });
                return true;
            };
            try {
                const res = await fetch(settingsEndpoint, {
                    method: 'PUT',
                    headers: { 'Content-Type': 'application/json' },
                    body: JSON.stringify({ theme: normalized }),
                });
                if (!res.ok) {
                    return persistFallback();
                }
                const persisted = await res.json().then((body) => extractApiData<Record<string, unknown> | unknown>(body));
                const persistedTheme = normalizeTheme((persisted as Record<string, unknown> | null)?.theme ?? normalized);
                persistLocalSettings({ theme: persistedTheme });
                set({ theme: persistedTheme });
                return true;
            } catch {
                return persistFallback();
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
