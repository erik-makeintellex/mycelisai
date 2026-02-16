import { describe, it, expect, vi, beforeEach } from 'vitest';
import { render, screen, waitFor } from '@testing-library/react';
import TeamsSummaryCard from '@/components/dashboard/TeamsSummaryCard';
import { useCortexStore } from '@/store/useCortexStore';
import { mockFetch } from '../setup';

describe('TeamsSummaryCard', () => {
    beforeEach(() => {
        useCortexStore.setState({
            missions: [],
            artifacts: [],
            isFetchingMissions: false,
            isFetchingArtifacts: false,
        });
    });

    it('renders with zero metrics when no data', async () => {
        mockFetch.mockResolvedValue({ ok: true, json: async () => [] });

        render(<TeamsSummaryCard />);

        await waitFor(() => {
            expect(screen.getByTestId('teams-summary')).toBeDefined();
        });

        // All values should be 0
        expect(screen.getAllByText('0')).toHaveLength(3);
    });

    it('calls fetchMissions and fetchArtifacts on mount', async () => {
        mockFetch.mockResolvedValue({ ok: true, json: async () => [] });

        render(<TeamsSummaryCard />);

        await waitFor(() => {
            const calls = mockFetch.mock.calls.map(c => c[0]);
            expect(calls).toContain('/api/v1/missions');
            expect(calls).toContain('/api/v1/artifacts');
        });
    });

    it('shows correct team/agent counts from active missions', async () => {
        useCortexStore.setState({
            missions: [
                { id: 'm1', intent: 'Scan', status: 'active', teams: 3, agents: 8 },
                { id: 'm2', intent: 'Build', status: 'active', teams: 2, agents: 5 },
                { id: 'm3', intent: 'Done', status: 'completed', teams: 1, agents: 2 },
            ],
            artifacts: [],
        });
        mockFetch.mockResolvedValue({ ok: true, json: async () => [] });

        render(<TeamsSummaryCard />);

        await waitFor(() => {
            expect(screen.getByText('5')).toBeDefined();  // 3+2 teams (active only)
            expect(screen.getByText('13')).toBeDefined(); // 8+5 agents (active only)
        });
    });

    it('shows artifact count as Outputs', async () => {
        useCortexStore.setState({
            missions: [
                { id: 'm1', intent: 'Scan', status: 'active', teams: 1, agents: 1 },
            ],
            artifacts: [
                { id: 'a1', agent_id: 'ag1', artifact_type: 'code', title: 'Output 1', content_type: 'text/plain', metadata: {}, status: 'pending', created_at: new Date().toISOString() },
                { id: 'a2', agent_id: 'ag1', artifact_type: 'document', title: 'Output 2', content_type: 'text/plain', metadata: {}, status: 'approved', created_at: new Date().toISOString() },
                { id: 'a3', agent_id: 'ag1', artifact_type: 'data', title: 'Output 3', content_type: 'text/plain', metadata: {}, status: 'pending', created_at: new Date().toISOString() },
            ],
        });
        mockFetch.mockResolvedValue({ ok: true, json: async () => [] });

        render(<TeamsSummaryCard />);

        // Outputs pill should show 3 (unique to the artifact count)
        await waitFor(() => {
            expect(screen.getByText('3')).toBeDefined();
        });
    });

    it('excludes completed/failed missions from team/agent counts', async () => {
        useCortexStore.setState({
            missions: [
                { id: 'm1', intent: 'Done', status: 'completed', teams: 10, agents: 50 },
                { id: 'm2', intent: 'Fail', status: 'failed', teams: 5, agents: 20 },
            ],
            artifacts: [],
        });
        mockFetch.mockResolvedValue({ ok: true, json: async () => [] });

        render(<TeamsSummaryCard />);

        await waitFor(() => {
            // All zeros — no active missions
            expect(screen.getAllByText('0')).toHaveLength(3);
        });
    });

    it('renders 3 metric pills (Teams, Agents, Outputs)', async () => {
        mockFetch.mockResolvedValue({ ok: true, json: async () => [] });

        render(<TeamsSummaryCard />);

        await waitFor(() => {
            expect(screen.getByText('Teams')).toBeDefined();
            expect(screen.getByText('Agents')).toBeDefined();
            expect(screen.getByText('Outputs')).toBeDefined();
        });
    });
});
