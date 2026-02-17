import { describe, it, expect, vi } from 'vitest';
import { render, screen, act } from '@testing-library/react';

// Mock next/dynamic — resolve the loader synchronously since we mock the component anyway
vi.mock('next/dynamic', () => ({
    __esModule: true,
    default: (loader: any, _opts?: any) => {
        // Eagerly load the mock and return a wrapper component
        let Resolved: any = null;
        loader().then((mod: any) => { Resolved = mod.default || mod; });
        return (props: any) => {
            const React = require('react');
            // Force a microtask resolve for the lazy import
            const [Comp, setComp] = React.useState(Resolved);
            React.useEffect(() => {
                if (!Comp) loader().then((mod: any) => setComp(() => mod.default || mod));
            }, []);
            return Comp ? React.createElement(Comp, props) : null;
        };
    },
}));

// Mock the Workspace component (heavy: ReactFlow, CircuitBoard, ArchitectChat, NatsWaterfall)
vi.mock('@/components/workspace/Workspace', () => ({
    __esModule: true,
    default: () => <div data-testid="workspace">Workspace</div>,
}));

import WiringPage from '@/app/wiring/page';

describe('Wiring Page (app/wiring/page.tsx)', () => {
    it('mounts without crashing', async () => {
        await act(async () => {
            render(<WiringPage />);
        });

        expect(screen.getByText('Neural Wiring')).toBeDefined();
    });

    it('renders subtitle', async () => {
        await act(async () => {
            render(<WiringPage />);
        });

        expect(screen.getByText('Negotiate intent into executable DAGs')).toBeDefined();
    });

    it('shows live signal indicator', async () => {
        await act(async () => {
            render(<WiringPage />);
        });

        expect(screen.getByText('Live Signal Graph')).toBeDefined();
    });
});
