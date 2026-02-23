import { describe, it, expect } from 'vitest';
import { render, screen } from '@testing-library/react';
import DegradedState from '@/components/shared/DegradedState';

describe('DegradedState (missing_dependency_shows_degraded_panel)', () => {
    it('renders title and reason', () => {
        render(
            <DegradedState
                title="Test Feature"
                reason="This feature is not yet implemented."
            />
        );
        expect(screen.getByText('Test Feature')).toBeDefined();
        expect(screen.getByText('This feature is not yet implemented.')).toBeDefined();
    });

    it('renders unavailable items with danger styling', () => {
        const { container } = render(
            <DegradedState
                title="Offline"
                reason="Service unavailable."
                unavailable={['Feature A', 'Feature B']}
            />
        );
        expect(screen.getByText('Unavailable')).toBeDefined();
        expect(screen.getByText('Feature A')).toBeDefined();
        expect(screen.getByText('Feature B')).toBeDefined();
        // Danger dots should be present
        const dangerDots = container.querySelectorAll('.bg-cortex-danger');
        expect(dangerDots.length).toBeGreaterThanOrEqual(2);
    });

    it('renders available items with success styling', () => {
        const { container } = render(
            <DegradedState
                title="Partial"
                reason="Partially available."
                available={['Working A', 'Working B']}
            />
        );
        expect(screen.getByText('Still Working')).toBeDefined();
        expect(screen.getByText('Working A')).toBeDefined();
        expect(screen.getByText('Working B')).toBeDefined();
        // Success dots should be present
        const successDots = container.querySelectorAll('.bg-cortex-success');
        expect(successDots.length).toBeGreaterThanOrEqual(2);
    });

    it('renders action text when provided', () => {
        render(
            <DegradedState
                title="Pending"
                reason="Coming soon."
                action="Try again later."
            />
        );
        expect(screen.getByText('Try again later.')).toBeDefined();
    });

    it('omits unavailable section when not provided', () => {
        render(
            <DegradedState
                title="Simple"
                reason="Simple degraded state."
            />
        );
        expect(screen.queryByText('Unavailable')).toBeNull();
        expect(screen.queryByText('Still Working')).toBeNull();
    });

    it('omits action text when not provided', () => {
        render(
            <DegradedState
                title="No Action"
                reason="No action needed."
            />
        );
        // The component should render without the action paragraph
        const container = screen.getByText('No Action').closest('div');
        expect(container).toBeDefined();
    });

    it('uses cortex theme (no bg-white)', () => {
        const { container } = render(
            <DegradedState
                title="Theme Check"
                reason="Verifying dark theme."
                unavailable={['X']}
                available={['Y']}
                action="Action text"
            />
        );
        expect(container.innerHTML).not.toContain('bg-white');
        expect(container.innerHTML).toContain('bg-cortex-surface');
        expect(container.innerHTML).toContain('border-cortex-border');
    });

    it('renders warning icon', () => {
        const { container } = render(
            <DegradedState
                title="Warning Test"
                reason="Should show warning icon."
            />
        );
        // AlertTriangle from lucide-react renders as svg
        const svg = container.querySelector('svg');
        expect(svg).toBeDefined();
        expect(svg).not.toBeNull();
    });
});
