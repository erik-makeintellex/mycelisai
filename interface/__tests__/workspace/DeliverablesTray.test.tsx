import { describe, it, expect, vi, beforeEach } from 'vitest';
import { render, screen } from '@testing-library/react';

vi.mock('lucide-react', () => ({
    Package: (props: any) => <svg data-testid="package-icon" {...props} />,
    Clock: (props: any) => <svg data-testid="clock-icon" {...props} />,
}));

import DeliverablesTray from '@/components/workspace/DeliverablesTray';
import { useCortexStore, type CTSEnvelope } from '@/store/useCortexStore';

const makeEnvelope = (id: string, overrides?: Partial<CTSEnvelope>): CTSEnvelope => ({
    id,
    source: `agent-${id}`,
    signal: 'artifact',
    timestamp: new Date().toISOString(),
    payload: {
        content: `Deliverable content for ${id}`,
        content_type: 'text',
        title: `Deliverable ${id}`,
    },
    ...overrides,
});

describe('DeliverablesTray', () => {
    beforeEach(() => {
        useCortexStore.setState({
            pendingArtifacts: [],
            selectedArtifact: null,
        });
    });

    it('renders deliverables list from store', () => {
        useCortexStore.setState({
            pendingArtifacts: [
                makeEnvelope('d-001'),
                makeEnvelope('d-002'),
            ],
        });

        render(<DeliverablesTray />);

        // The header shows "Pending Deliverables"
        expect(screen.getByText('Pending Deliverables')).toBeDefined();

        // Should show artifact titles
        expect(screen.getByText('Deliverable d-001')).toBeDefined();
        expect(screen.getByText('Deliverable d-002')).toBeDefined();

        // The count badge should show 2
        expect(screen.getByText('2')).toBeDefined();
    });

    it('shows governance halt state for governance_halt signals', () => {
        useCortexStore.setState({
            pendingArtifacts: [
                makeEnvelope('gov-001', {
                    signal: 'governance_halt',
                    trust_score: 0.3,
                    payload: {
                        content: 'Trust score below threshold. Awaiting human approval.',
                        content_type: 'text',
                        title: 'Governance Halt: agent-alpha',
                    },
                }),
            ],
        });

        render(<DeliverablesTray />);

        expect(screen.getByText('Governance Halt: agent-alpha')).toBeDefined();
        expect(screen.getByText('Pending Deliverables')).toBeDefined();
    });

    it('returns null (renders nothing) when no deliverables exist', () => {
        useCortexStore.setState({ pendingArtifacts: [] });

        const { container } = render(<DeliverablesTray />);

        // Component returns null when pendingArtifacts is empty
        expect(container.innerHTML).toBe('');
    });
});
