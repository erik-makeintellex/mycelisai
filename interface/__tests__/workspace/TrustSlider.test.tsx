import { describe, it, expect, vi, beforeEach } from 'vitest';
import { render, screen, fireEvent, waitFor } from '@testing-library/react';

vi.mock('lucide-react', () => ({
    Shield: (props: any) => <svg data-testid="shield-icon" {...props} />,
}));

import TrustSlider from '@/components/workspace/TrustSlider';
import { useCortexStore } from '@/store/useCortexStore';
import { mockFetch } from '../setup';

describe('TrustSlider', () => {
    beforeEach(() => {
        useCortexStore.setState({
            trustThreshold: 0.7,
            isSyncingThreshold: false,
        });
    });

    it('renders with initial trust value from store', () => {
        useCortexStore.setState({ trustThreshold: 0.65 });
        render(<TrustSlider />);

        // The displayed value text shows the threshold formatted to 2 decimal places
        expect(screen.getByText('0.65')).toBeDefined();

        // The range input should have the correct value
        const slider = screen.getByRole('slider') as HTMLInputElement;
        expect(slider.value).toBe('0.65');
    });

    it('calls setTrustThreshold on change which triggers PUT to /api/v1/trust/threshold', async () => {
        mockFetch.mockResolvedValueOnce({ ok: true });

        render(<TrustSlider />);

        const slider = screen.getByRole('slider');
        fireEvent.change(slider, { target: { value: '0.85' } });

        // setTrustThreshold in the store fires a PUT request
        await waitFor(() => {
            expect(mockFetch).toHaveBeenCalledWith(
                '/api/v1/trust/threshold',
                expect.objectContaining({
                    method: 'PUT',
                    headers: { 'Content-Type': 'application/json' },
                    body: JSON.stringify({ threshold: 0.85 }),
                })
            );
        });
    });

    it('displays value synced with store state', () => {
        // Set a specific trust value
        useCortexStore.setState({ trustThreshold: 0.30 });
        const { rerender } = render(<TrustSlider />);

        // Should show PERMISSIVE label (threshold < 0.5)
        expect(screen.getByText('PERMISSIVE')).toBeDefined();
        expect(screen.getByText('0.30')).toBeDefined();

        // Update store to a higher value
        useCortexStore.setState({ trustThreshold: 0.90 });
        rerender(<TrustSlider />);

        // Should now show STRICT label (threshold >= 0.8)
        expect(screen.getByText('STRICT')).toBeDefined();
        expect(screen.getByText('0.90')).toBeDefined();
    });
});
