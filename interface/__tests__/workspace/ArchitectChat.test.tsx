import { describe, it, expect, vi, beforeEach } from 'vitest';
import { render, screen, fireEvent, waitFor } from '@testing-library/react';

// vi.mock calls BEFORE component imports
vi.mock('@/components/workspace/TrustSlider', () => ({
    default: () => <div data-testid="trust-slider">TrustSlider</div>,
}));
vi.mock('lucide-react', () => ({
    Send: (props: any) => <svg data-testid="send-icon" {...props} />,
    Loader2: (props: any) => <svg data-testid="loader-icon" {...props} />,
    Bot: (props: any) => <svg data-testid="bot-icon" {...props} />,
    User: (props: any) => <svg data-testid="user-icon" {...props} />,
    FileJson: (props: any) => <svg data-testid="file-json-icon" {...props} />,
}));

import ArchitectChat from '@/components/workspace/ArchitectChat';
import { useCortexStore } from '@/store/useCortexStore';

describe('ArchitectChat', () => {
    beforeEach(() => {
        useCortexStore.setState({
            chatHistory: [],
            isDrafting: false,
            error: null,
            blueprint: null,
            missionStatus: 'idle',
        });
    });

    it('renders the text input field', () => {
        render(<ArchitectChat />);
        const input = screen.getByPlaceholderText('Describe your mission intent...');
        expect(input).toBeDefined();
        expect(input.tagName.toLowerCase()).toBe('input');
    });

    it('calls submitIntent on submit', async () => {
        const submitIntentSpy = vi.fn();
        useCortexStore.setState({ submitIntent: submitIntentSpy });

        render(<ArchitectChat />);

        const input = screen.getByPlaceholderText('Describe your mission intent...');
        fireEvent.change(input, { target: { value: 'Build a research team' } });
        fireEvent.keyDown(input, { key: 'Enter' });

        expect(submitIntentSpy).toHaveBeenCalledWith('Build a research team');
    });

    it('displays message history from store', () => {
        useCortexStore.setState({
            chatHistory: [
                { role: 'user', content: 'Create a data pipeline' },
                { role: 'architect', content: 'Blueprint **pipeline-001** generated.' },
            ],
        });

        render(<ArchitectChat />);
        expect(screen.getByText('Create a data pipeline')).toBeDefined();
        expect(screen.getByText('Blueprint **pipeline-001** generated.')).toBeDefined();
    });

    it('shows loading state when isDrafting is true', () => {
        useCortexStore.setState({ isDrafting: true });

        render(<ArchitectChat />);
        // "drafting..." text appears in the header
        expect(screen.getByText('drafting...')).toBeDefined();
        // The input should be disabled
        const input = screen.getByPlaceholderText('Describe your mission intent...');
        expect(input).toHaveProperty('disabled', true);
    });
});
