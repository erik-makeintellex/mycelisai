import { describe, it, expect, vi, beforeEach } from 'vitest';
import { render, screen } from '@testing-library/react';
import { mockFetch } from '../setup';

// Mock the three sub-panels to isolate MemoryExplorer layout tests
vi.mock('@/components/memory/HotMemoryPanel', () => ({
    __esModule: true,
    default: () => <div data-testid="hot-panel">Hot Panel</div>,
}));

vi.mock('@/components/memory/WarmMemoryPanel', () => ({
    __esModule: true,
    default: ({ onSearchRelated }: any) => (
        <div data-testid="warm-panel">
            Warm Panel
            <button data-testid="warm-search-btn" onClick={() => onSearchRelated('test query')}>
                Search Related
            </button>
        </div>
    ),
}));

vi.mock('@/components/memory/ColdMemoryPanel', () => ({
    __esModule: true,
    default: ({ searchQuery }: any) => (
        <div data-testid="cold-panel">
            Cold Panel
            {searchQuery && <span data-testid="cold-query">{searchQuery}</span>}
        </div>
    ),
}));

import MemoryExplorer from '@/components/memory/MemoryExplorer';

describe('MemoryExplorer', () => {
    it('renders the three-tier layout with Hot, Warm, and Cold panels', () => {
        render(<MemoryExplorer />);

        // Header should display the three-tier memory title
        expect(screen.getByText('THREE-TIER MEMORY')).toBeDefined();

        // Tier chips should be visible in the header
        expect(screen.getByText('Hot')).toBeDefined();
        expect(screen.getByText('Warm')).toBeDefined();
        expect(screen.getByText('Cold')).toBeDefined();

        // All three panels should be mounted
        expect(screen.getByTestId('hot-panel')).toBeDefined();
        expect(screen.getByTestId('warm-panel')).toBeDefined();
        expect(screen.getByTestId('cold-panel')).toBeDefined();
    });

    it('passes search query from WarmMemoryPanel to ColdMemoryPanel', async () => {
        const { fireEvent } = await import('@testing-library/react');

        render(<MemoryExplorer />);

        // Initially no search query passed to cold panel
        expect(screen.queryByTestId('cold-query')).toBeNull();

        // Trigger a "search related" action from the warm panel
        fireEvent.click(screen.getByTestId('warm-search-btn'));

        // The cold panel should now receive the search query
        expect(screen.getByTestId('cold-query')).toBeDefined();
        expect(screen.getByText('test query')).toBeDefined();
    });
});
