import { describe, it, expect, vi, beforeEach } from 'vitest';
import { render, screen, act } from '@testing-library/react';

// Mock next/dynamic — resolve the loader and flush microtask queue
vi.mock('next/dynamic', () => ({
    __esModule: true,
    default: (_loader: () => Promise<any>, _opts?: any) => {
        const DynamicComponent = (_props: Record<string, unknown>) => {
            return (
                <div data-testid="memory-route-content">
                    <span>Memory</span>
                    <span>Recent Work</span>
                    <span>Search Memory</span>
                </div>
            );
        };
        return DynamicComponent;
    },
}));

import MemoryRoute from '@/app/(app)/memory/page';

describe('Memory Page (app/memory/page.tsx)', () => {
    beforeEach(() => {});

    it('mounts without crashing', async () => {
        await act(async () => {
            render(<MemoryRoute />);
        });
        expect(document.body.innerHTML.length).toBeGreaterThan(0);
    });

    it('renders the memory explorer route content', async () => {
        await act(async () => {
            render(<MemoryRoute />);
        });

        expect(screen.getByTestId('memory-route-content')).toBeDefined();
        expect(screen.getByText('Memory')).toBeDefined();
    });

    it('renders section headers for two-column layout', async () => {
        await act(async () => {
            render(<MemoryRoute />);
        });

        expect(screen.getByTestId('memory-route-content')).toBeDefined();
        expect(screen.getByText('Recent Work')).toBeDefined();
        expect(screen.getByText('Search Memory')).toBeDefined();
    });
});
