import { describe, it, expect, vi, beforeEach } from 'vitest';
import { render, screen, act, waitFor } from '@testing-library/react';
import { mockFetch } from '../setup';

// Mock next/dynamic — resolve the loader and flush microtask queue
vi.mock('next/dynamic', () => ({
    __esModule: true,
    default: (loader: any, _opts?: any) => {
        const React = require('react');
        const DynamicComponent = (props: any) => {
            const [Comp, setComp] = React.useState<any>(null);
            React.useEffect(() => {
                loader().then((mod: any) => {
                    setComp(() => mod.default || mod);
                });
            }, []);
            return Comp ? React.createElement(Comp, props) : null;
        };
        return DynamicComponent;
    },
}));

// Mock lucide-react icons used by MemoryExplorer
vi.mock('lucide-react', () => ({
    Brain: (props: any) => <svg data-testid="brain-icon" {...props} />,
    Flame: (props: any) => <svg data-testid="flame-icon" {...props} />,
    Database: (props: any) => <svg data-testid="database-icon" {...props} />,
    Snowflake: (props: any) => <svg data-testid="snowflake-icon" {...props} />,
    Search: (props: any) => <svg data-testid="search-icon" {...props} />,
    RefreshCw: (props: any) => <svg data-testid="refresh-icon" {...props} />,
    Clock: (props: any) => <svg data-testid="clock-icon" {...props} />,
    FileText: (props: any) => <svg data-testid="filetext-icon" {...props} />,
    ChevronDown: (props: any) => <svg data-testid="chevdown-icon" {...props} />,
    ChevronRight: (props: any) => <svg data-testid="chevright-icon" {...props} />,
    Loader2: (props: any) => <svg data-testid="loader-icon" {...props} />,
    AlertCircle: (props: any) => <svg data-testid="alert-icon" {...props} />,
    Zap: (props: any) => <svg data-testid="zap-icon" {...props} />,
    MessageSquare: (props: any) => <svg data-testid="msg-icon" {...props} />,
}));

// Mock the three memory panels (they fetch data and have complex rendering)
vi.mock('@/components/memory/HotMemoryPanel', () => ({
    __esModule: true,
    default: () => <div data-testid="hot-memory-panel">Hot Memory</div>,
}));

vi.mock('@/components/memory/WarmMemoryPanel', () => ({
    __esModule: true,
    default: () => <div data-testid="warm-memory-panel">Warm Memory</div>,
}));

vi.mock('@/components/memory/ColdMemoryPanel', () => ({
    __esModule: true,
    default: () => <div data-testid="cold-memory-panel">Cold Memory</div>,
}));

import MemoryRoute from '@/app/memory/page';

describe('Memory Page (app/memory/page.tsx)', () => {
    beforeEach(() => {
        mockFetch.mockResolvedValue({
            ok: true,
            json: async () => ({}),
        });
    });

    it('mounts without crashing', async () => {
        await act(async () => {
            render(<MemoryRoute />);
        });
        // Flush microtask queue for dynamic import resolution
        await act(async () => {
            await new Promise((r) => setTimeout(r, 0));
        });

        expect(document.body.innerHTML.length).toBeGreaterThan(0);
    });

    it('renders the memory explorer header', async () => {
        await act(async () => {
            render(<MemoryRoute />);
        });
        await act(async () => {
            await new Promise((r) => setTimeout(r, 0));
        });

        await waitFor(() => {
            expect(screen.getByText('THREE-TIER MEMORY')).toBeDefined();
        });
    });

    it('renders tier legend chips', async () => {
        await act(async () => {
            render(<MemoryRoute />);
        });
        await act(async () => {
            await new Promise((r) => setTimeout(r, 0));
        });

        await waitFor(() => {
            expect(screen.getByText('Hot')).toBeDefined();
            expect(screen.getByText('Warm')).toBeDefined();
            expect(screen.getByText('Cold')).toBeDefined();
        });
    });
});
