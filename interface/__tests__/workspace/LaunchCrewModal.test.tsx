import { describe, it, expect, vi, beforeEach } from 'vitest';
import { act, fireEvent, render, screen, waitFor } from '@testing-library/react';
import LaunchCrewModal from '@/components/workspace/LaunchCrewModal';
import { useCortexStore, type ProposalData } from '@/store/useCortexStore';

function baseProposal(): ProposalData {
    return {
        intent: 'Launch a delivery crew',
        teams: 2,
        agents: 4,
        tools: ['delegate_task', 'write_file'],
        risk_level: 'medium',
        confirm_token: 'ct-123',
        intent_proof_id: 'ip-123',
    };
}

function resetStore(overrides?: Partial<ReturnType<typeof useCortexStore.getState>>) {
    useCortexStore.setState({
        assistantName: 'Soma',
        missionChat: [],
        isMissionChatting: false,
        missionChatError: null,
        pendingProposal: null,
        activeMode: 'answer',
        activeRunId: null,
        sendMissionChat: vi.fn(),
        confirmProposal: vi.fn(async () => ({ ok: true, runId: null })),
        cancelProposal: vi.fn(),
        setCouncilTarget: vi.fn(),
        ...overrides,
    });
}

describe('LaunchCrewModal', () => {
    beforeEach(() => {
        resetStore();
    });

    it('sends the intent through Soma and enters the evaluation state', async () => {
        const sendMissionChat = vi.fn();
        resetStore({ sendMissionChat });

        render(<LaunchCrewModal onClose={vi.fn()} />);

        fireEvent.change(screen.getByPlaceholderText(/Describe the outcome you need/i), {
            target: { value: 'Design a crew for workflow onboarding' },
        });
        fireEvent.click(screen.getByRole('button', { name: /Send to Soma/i }));

        expect(sendMissionChat).toHaveBeenCalledWith('Design a crew for workflow onboarding');
        expect(screen.getByText(/is evaluating your request/i)).toBeDefined();
    });

    it('renders the proposal outcome when Launch Crew receives a proposal', async () => {
        const proposal = baseProposal();
        render(<LaunchCrewModal onClose={vi.fn()} />);

        fireEvent.change(screen.getByPlaceholderText(/Describe the outcome you need/i), {
            target: { value: 'Assemble a team for docs cleanup' },
        });
        fireEvent.click(screen.getByRole('button', { name: /Send to Soma/i }));

        act(() => {
            useCortexStore.setState({
                isMissionChatting: false,
                activeMode: 'proposal',
                pendingProposal: proposal,
                missionChat: [
                    { role: 'user', content: 'Assemble a team for docs cleanup' },
                    {
                        role: 'council',
                        content: 'I have a proposal ready.',
                        source_node: 'admin',
                        mode: 'proposal',
                        proposal,
                    },
                ],
            });
        });

        await waitFor(() => {
            expect(screen.getByText(/prepared a crew proposal/i)).toBeDefined();
        });
        expect(screen.getByRole('button', { name: /^Launch Crew$/i })).toBeDefined();
    });

    it('renders a direct answer outcome when no crew launch is needed', async () => {
        render(<LaunchCrewModal onClose={vi.fn()} />);

        fireEvent.change(screen.getByPlaceholderText(/Describe the outcome you need/i), {
            target: { value: 'Summarize our current architecture goals' },
        });
        fireEvent.click(screen.getByRole('button', { name: /Send to Soma/i }));

        act(() => {
            useCortexStore.setState({
                isMissionChatting: false,
                activeMode: 'answer',
                missionChat: [
                    { role: 'user', content: 'Summarize our current architecture goals' },
                    {
                        role: 'council',
                        content: 'The next target is Launch Crew execution clarity.',
                        source_node: 'admin',
                        mode: 'answer',
                    },
                ],
            });
        });

        await waitFor(() => {
            expect(screen.getByText(/answered directly/i)).toBeDefined();
        });
        expect(screen.getByText(/No crew launch was required/i)).toBeDefined();
    });

    it('renders an actionable blocker outcome when the request fails', async () => {
        render(<LaunchCrewModal onClose={vi.fn()} />);

        fireEvent.change(screen.getByPlaceholderText(/Describe the outcome you need/i), {
            target: { value: 'Launch a crew for deployment' },
        });
        fireEvent.click(screen.getByRole('button', { name: /Send to Soma/i }));

        act(() => {
            useCortexStore.setState({
                isMissionChatting: false,
                missionChatError: 'Soma chat blocked (500)',
                activeMode: 'blocker',
                missionChat: [
                    { role: 'user', content: 'Launch a crew for deployment' },
                    {
                        role: 'council',
                        content: 'Soma chat blocked (500)',
                        source_node: 'admin',
                        mode: 'blocker',
                    },
                ],
            });
        });

        await waitFor(() => {
            expect(screen.getByText(/Launch Crew is blocked/i)).toBeDefined();
        });
        expect(screen.getByText(/Soma chat blocked \(500\)/i)).toBeDefined();
    });

    it('shows an execution result after confirmation succeeds', async () => {
        const proposal = baseProposal();
        const confirmProposal = vi.fn(async () => {
            useCortexStore.setState({ activeRunId: 'run-12345' });
            return { ok: true, runId: 'run-12345' };
        });
        resetStore({
            pendingProposal: proposal,
            activeMode: 'proposal',
            confirmProposal,
            missionChat: [
                { role: 'user', content: 'Launch a docs crew' },
                {
                    role: 'council',
                    content: 'I have a proposal ready.',
                    source_node: 'admin',
                    mode: 'proposal',
                    proposal,
                },
            ],
        });

        render(<LaunchCrewModal onClose={vi.fn()} />);

        fireEvent.change(screen.getByPlaceholderText(/Describe the outcome you need/i), {
            target: { value: 'Launch a docs crew' },
        });
        fireEvent.click(screen.getByRole('button', { name: /Send to Soma/i }));

        act(() => {
            useCortexStore.setState({
                isMissionChatting: false,
                pendingProposal: proposal,
                activeMode: 'proposal',
            });
        });

        await waitFor(() => {
            expect(screen.getByRole('button', { name: /^Launch Crew$/i })).toBeDefined();
        });

        fireEvent.click(screen.getByRole('button', { name: /^Launch Crew$/i }));

        await waitFor(() => {
            expect(screen.getByText(/Crew launch submitted/i)).toBeDefined();
        });
        expect(confirmProposal).toHaveBeenCalledTimes(1);
        expect(screen.getByRole('button', { name: /View Run/i })).toBeDefined();
    });

    it('shows an execution result without a run link when confirmation returns no run id', async () => {
        const proposal = baseProposal();
        const confirmProposal = vi.fn(async () => ({ ok: true, runId: null }));
        resetStore({
            pendingProposal: proposal,
            activeMode: 'proposal',
            confirmProposal,
            missionChat: [
                { role: 'user', content: 'Launch a docs crew' },
                {
                    role: 'council',
                    content: 'I have a proposal ready.',
                    source_node: 'admin',
                    mode: 'proposal',
                    proposal,
                },
            ],
        });

        render(<LaunchCrewModal onClose={vi.fn()} />);

        fireEvent.change(screen.getByPlaceholderText(/Describe the outcome you need/i), {
            target: { value: 'Launch a docs crew' },
        });
        fireEvent.click(screen.getByRole('button', { name: /Send to Soma/i }));

        act(() => {
            useCortexStore.setState({
                isMissionChatting: false,
                pendingProposal: proposal,
                activeMode: 'proposal',
                activeRunId: null,
            });
        });

        await waitFor(() => {
            expect(screen.getByRole('button', { name: /^Launch Crew$/i })).toBeDefined();
        });

        fireEvent.click(screen.getByRole('button', { name: /^Launch Crew$/i }));

        await waitFor(() => {
            expect(screen.getByText(/Crew launch submitted/i)).toBeDefined();
        });
        expect(confirmProposal).toHaveBeenCalledTimes(1);
        expect(screen.queryByRole('button', { name: /View Run/i })).toBeNull();
    });
});
