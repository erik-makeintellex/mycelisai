import { beforeEach, describe, expect, it } from 'vitest';
import { useCortexStore } from '@/store/useCortexStore';
import { mockFetch } from '../setup';
import { resetCortexStore } from './useCortexStoreTestSupport';

describe('useCortexStore data fetch', () => {
    beforeEach(() => {
        resetCortexStore();
    });

    describe('fetchMissions', () => {
        it('sets isFetchingMissions during fetch', async () => {
            mockFetch.mockResolvedValue({
                ok: true,
                json: async () => [],
            });

            const promise = useCortexStore.getState().fetchMissions();
            expect(useCortexStore.getState().isFetchingMissions).toBe(true);

            await promise;
            expect(useCortexStore.getState().isFetchingMissions).toBe(false);
        });

        it('stores missions array from API', async () => {
            const missions = [
                { id: 'm1', intent: 'Scan', status: 'active', teams: 2, agents: 5 },
            ];
            mockFetch.mockResolvedValue({ ok: true, json: async () => missions });

            await useCortexStore.getState().fetchMissions();

            expect(useCortexStore.getState().missions).toEqual(missions);
        });

        it('sets empty array on non-ok response', async () => {
            mockFetch.mockResolvedValue({ ok: false });

            await useCortexStore.getState().fetchMissions();

            expect(useCortexStore.getState().missions).toEqual([]);
        });

        it('sets empty array on network error', async () => {
            mockFetch.mockRejectedValue(new Error('Network error'));

            await useCortexStore.getState().fetchMissions();

            expect(useCortexStore.getState().missions).toEqual([]);
        });

        it('handles non-array response gracefully', async () => {
            mockFetch.mockResolvedValue({
                ok: true,
                json: async () => ({ error: 'not array' }),
            });

            await useCortexStore.getState().fetchMissions();

            expect(useCortexStore.getState().missions).toEqual([]);
        });
    });

    describe('user settings fallback', () => {
        it('persists assistant name locally when the settings endpoint is unavailable', async () => {
            mockFetch.mockResolvedValue({ ok: false });

            await expect(useCortexStore.getState().updateAssistantName('Atlas')).resolves.toBe(true);

            expect(useCortexStore.getState().assistantName).toBe('Atlas');
            expect(localStorage.getItem('mycelis-user-settings')).toContain('"assistantName":"Atlas"');
        });

        it('hydrates assistant name and theme from local storage when settings fetch fails', async () => {
            localStorage.setItem('mycelis-user-settings', JSON.stringify({
                assistantName: 'Atlas',
                theme: 'midnight-cortex',
            }));
            mockFetch.mockResolvedValue({ ok: false });

            await useCortexStore.getState().fetchUserSettings();

            expect(useCortexStore.getState().assistantName).toBe('Atlas');
            expect(useCortexStore.getState().theme).toBe('midnight-cortex');
        });
    });

    describe('fetchArtifacts', () => {
        it('fetches all artifacts without filters', async () => {
            const artifacts = [
                { id: 'a1', agent_id: 'ag1', artifact_type: 'code', title: 'Output', content_type: 'text', metadata: {}, status: 'pending', created_at: '' },
            ];
            mockFetch.mockResolvedValue({ ok: true, json: async () => artifacts });

            await useCortexStore.getState().fetchArtifacts();

            expect(mockFetch).toHaveBeenCalledWith('/api/v1/artifacts');
            expect(useCortexStore.getState().artifacts).toEqual(artifacts);
        });

        it('passes filters as query params', async () => {
            mockFetch.mockResolvedValue({ ok: true, json: async () => [] });

            await useCortexStore.getState().fetchArtifacts({
                mission_id: 'm1',
                team_id: 't1',
                limit: 10,
            });

            const url = mockFetch.mock.calls[0][0] as string;
            expect(url).toContain('mission_id=m1');
            expect(url).toContain('team_id=t1');
            expect(url).toContain('limit=10');
        });

        it('sets empty array on failure', async () => {
            mockFetch.mockRejectedValue(new Error('fail'));

            await useCortexStore.getState().fetchArtifacts();

            expect(useCortexStore.getState().artifacts).toEqual([]);
        });

        it('updateArtifactStatus uses the PUT contract and syncs selected detail', async () => {
            useCortexStore.setState({
                artifacts: [
                    { id: 'a1', agent_id: 'ag1', artifact_type: 'code', title: 'Output', content_type: 'text', metadata: {}, status: 'pending', created_at: '' },
                ],
                selectedArtifactDetail: {
                    id: 'a1',
                    agent_id: 'ag1',
                    artifact_type: 'code',
                    title: 'Output',
                    content_type: 'text',
                    metadata: {},
                    status: 'pending',
                    created_at: '',
                },
            });
            mockFetch.mockResolvedValue({ ok: true });

            await useCortexStore.getState().updateArtifactStatus('a1', 'approved');

            expect(mockFetch).toHaveBeenCalledWith(
                '/api/v1/artifacts/a1/status',
                expect.objectContaining({ method: 'PUT' }),
            );
            expect(useCortexStore.getState().artifacts[0].status).toBe('approved');
            expect(useCortexStore.getState().selectedArtifactDetail?.status).toBe('approved');
        });
    });

    describe('fetchSensors', () => {
        it('stores sensors from API response', async () => {
            const sensors = [
                { id: 's1', type: 'email', status: 'online', last_seen: '', label: 'Gmail' },
            ];
            mockFetch.mockResolvedValue({
                ok: true,
                json: async () => ({ sensors }),
            });

            await useCortexStore.getState().fetchSensors();

            expect(useCortexStore.getState().sensorFeeds).toEqual(sensors);
        });

        it('handles missing sensors key gracefully', async () => {
            mockFetch.mockResolvedValue({
                ok: true,
                json: async () => ({}),
            });

            await useCortexStore.getState().fetchSensors();

            expect(useCortexStore.getState().sensorFeeds).toEqual([]);
        });
    });
});
