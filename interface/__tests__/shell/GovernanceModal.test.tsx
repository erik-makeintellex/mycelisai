import { describe, it, expect, vi, beforeEach } from 'vitest';
import { render, screen, fireEvent } from '@testing-library/react';
import GovernanceModal from '@/components/shell/GovernanceModal';
import { useCortexStore } from '@/store/useCortexStore';

const MOCK_ARTIFACT = {
    id: 'art-1',
    source: 'agent-alpha',
    timestamp: new Date().toISOString(),
    payload: {
        title: 'Test Output',
        content: 'Generated content here',
        content_type: 'text/plain',
    },
    proof: {
        method: 'semantic',
        pass: true,
        rubric_score: '4/5',
        logs: 'Verification passed',
    },
};

describe('GovernanceModal', () => {
    beforeEach(() => {
        useCortexStore.setState({
            selectedArtifact: null,
            pendingArtifacts: [],
        });
    });

    it('returns null when no artifact is selected', () => {
        const { container } = render(<GovernanceModal />);
        expect(container.innerHTML).toBe('');
    });

    it('renders when selectedArtifact is set', () => {
        useCortexStore.setState({ selectedArtifact: MOCK_ARTIFACT as any });
        render(<GovernanceModal />);
        expect(screen.getByText('Governance Review')).toBeDefined();
        expect(screen.getByText('agent-alpha')).toBeDefined();
    });

    it('shows agent output content', () => {
        useCortexStore.setState({ selectedArtifact: MOCK_ARTIFACT as any });
        render(<GovernanceModal />);
        expect(screen.getByText('Test Output')).toBeDefined();
        expect(screen.getByText('Generated content here')).toBeDefined();
    });

    it('shows proof of work details', () => {
        useCortexStore.setState({ selectedArtifact: MOCK_ARTIFACT as any });
        render(<GovernanceModal />);
        expect(screen.getByText('PASSED')).toBeDefined();
        expect(screen.getByText('semantic')).toBeDefined();
        expect(screen.getByText('Verification passed')).toBeDefined();
    });

    it('shows "No proof of work" when proof is null', () => {
        const noProof = { ...MOCK_ARTIFACT, proof: null };
        useCortexStore.setState({ selectedArtifact: noProof as any });
        render(<GovernanceModal />);
        expect(screen.getByText('No proof of work provided')).toBeDefined();
    });

    it('calls selectArtifact(null) on close button click', () => {
        useCortexStore.setState({
            selectedArtifact: MOCK_ARTIFACT as any,
        });
        render(<GovernanceModal />);

        // The close button (X) is in the header
        const buttons = screen.getAllByRole('button');
        const closeBtn = buttons[0]; // First button is X
        fireEvent.click(closeBtn);

        expect(useCortexStore.getState().selectedArtifact).toBeNull();
    });

    it('shows Approve & Dispatch button', () => {
        useCortexStore.setState({ selectedArtifact: MOCK_ARTIFACT as any });
        render(<GovernanceModal />);
        expect(screen.getByText('Approve & Dispatch')).toBeDefined();
    });

    it('shows Reject & Rework button', () => {
        useCortexStore.setState({ selectedArtifact: MOCK_ARTIFACT as any });
        render(<GovernanceModal />);
        expect(screen.getByText('Reject & Rework')).toBeDefined();
    });
});
