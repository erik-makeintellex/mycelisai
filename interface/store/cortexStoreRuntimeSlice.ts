import type { CortexState } from '@/store/cortexStoreState';
import type {
    ContextSnapshot,
    MissionProfile,
    MissionProfileCreate,
} from '@/store/cortexStoreTypes';
import type { CortexGet, CortexSet } from '@/store/cortexStoreSliceTypes';

export function createCortexRuntimeSlice(
    set: CortexSet,
    get: CortexGet,
): Pick<
    CortexState,
    | 'fetchRunTimeline'
    | 'fetchRecentRuns'
    | 'fetchMissionProfiles'
    | 'createMissionProfile'
    | 'updateMissionProfile'
    | 'deleteMissionProfile'
    | 'activateMissionProfile'
    | 'fetchContextSnapshots'
    | 'createContextSnapshot'
> {
    return {
        fetchRunTimeline: async (runId: string) => {
            set({ isFetchingTimeline: true });
            try {
                const res = await fetch(`/api/v1/runs/${runId}/events`);
                if (!res.ok) return;
                const body = await res.json();
                set({ runTimeline: body.data ?? body ?? [] });
            } catch {
                // offline — silent
            } finally {
                set({ isFetchingTimeline: false });
            }
        },

        fetchRecentRuns: async () => {
            set({ isFetchingRuns: true });
            try {
                const res = await fetch('/api/v1/runs');
                if (!res.ok) return;
                const body = await res.json();
                set({ recentRuns: body.data ?? body ?? [] });
            } catch {
                // offline — silent
            } finally {
                set({ isFetchingRuns: false });
            }
        },

        fetchMissionProfiles: async () => {
            try {
                const res = await fetch('/api/v1/mission-profiles');
                if (!res.ok) return;
                const body = await res.json();
                const profiles: MissionProfile[] = body.data ?? [];
                const active = profiles.find((profile) => profile.is_active);
                set({ missionProfiles: profiles, activeProfileId: active?.id ?? null });
            } catch {
                // degraded — silent
            }
        },

        createMissionProfile: async (profile: MissionProfileCreate) => {
            try {
                const res = await fetch('/api/v1/mission-profiles', {
                    method: 'POST',
                    headers: { 'Content-Type': 'application/json' },
                    body: JSON.stringify(profile),
                });
                if (!res.ok) {
                    console.error('[PROFILES] Create failed:', await res.text());
                    return null;
                }
                const body = await res.json();
                const created: MissionProfile = body.data;
                set((s) => ({ missionProfiles: [...s.missionProfiles, created] }));
                return created;
            } catch (err) {
                console.error('[PROFILES] Create error:', err);
                return null;
            }
        },

        updateMissionProfile: async (id: string, profile: MissionProfileCreate) => {
            try {
                const res = await fetch(`/api/v1/mission-profiles/${id}`, {
                    method: 'PUT',
                    headers: { 'Content-Type': 'application/json' },
                    body: JSON.stringify(profile),
                });
                if (!res.ok) {
                    console.error('[PROFILES] Update failed:', await res.text());
                    return;
                }
                const body = await res.json();
                const updated: MissionProfile = body.data;
                set((s) => ({
                    missionProfiles: s.missionProfiles.map((item) => (item.id === id ? updated : item)),
                }));
            } catch (err) {
                console.error('[PROFILES] Update error:', err);
            }
        },

        deleteMissionProfile: async (id: string) => {
            try {
                const res = await fetch(`/api/v1/mission-profiles/${id}`, { method: 'DELETE' });
                if (!res.ok) {
                    console.error('[PROFILES] Delete failed:', await res.text());
                    return;
                }
                set((s) => ({
                    missionProfiles: s.missionProfiles.filter((item) => item.id !== id),
                    activeProfileId: s.activeProfileId === id ? null : s.activeProfileId,
                }));
            } catch (err) {
                console.error('[PROFILES] Delete error:', err);
            }
        },

        activateMissionProfile: async (id: string) => {
            try {
                const res = await fetch(`/api/v1/mission-profiles/${id}/activate`, { method: 'POST' });
                if (!res.ok) {
                    console.error('[PROFILES] Activate failed:', await res.text());
                    return;
                }
                const body = await res.json();
                const activated: MissionProfile = body.data;
                set((s) => ({
                    activeProfileId: id,
                    missionProfiles: s.missionProfiles.map((item) =>
                        item.id === id
                            ? activated
                            : item.auto_start ? item : { ...item, is_active: false }
                    ),
                }));
            } catch (err) {
                console.error('[PROFILES] Activate error:', err);
            }
        },

        fetchContextSnapshots: async () => {
            try {
                const res = await fetch('/api/v1/context/snapshots');
                if (!res.ok) return;
                const body = await res.json();
                set({ contextSnapshots: body.data ?? [] });
            } catch {
                // degraded — silent
            }
        },

        createContextSnapshot: async (name: string) => {
            const { missionChat, activeRunId, missionProfiles, activeProfileId } = get();
            const activeProfile = missionProfiles.find((profile) => profile.id === activeProfileId);
            const roleProviders = activeProfile?.role_providers ?? {};
            try {
                const res = await fetch('/api/v1/context/snapshot', {
                    method: 'POST',
                    headers: { 'Content-Type': 'application/json' },
                    body: JSON.stringify({
                        name,
                        messages: missionChat,
                        run_state: { run_id: activeRunId },
                        role_providers: roleProviders,
                        source_profile: activeProfileId ?? undefined,
                    }),
                });
                if (!res.ok) {
                    console.error('[SNAPSHOTS] Create failed:', await res.text());
                    return null;
                }
                const body = await res.json();
                const snapshot: ContextSnapshot = body.data;
                set((s) => ({ contextSnapshots: [snapshot, ...s.contextSnapshots] }));
                return snapshot;
            } catch (err) {
                console.error('[SNAPSHOTS] Create error:', err);
                return null;
            }
        },
    };
}
