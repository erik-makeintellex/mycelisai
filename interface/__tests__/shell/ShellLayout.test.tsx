import { describe, it, expect, vi } from 'vitest';
import { render, screen } from '@testing-library/react';

// Mock child zones to isolate ShellLayout testing
vi.mock('@/components/shell/ZoneA_Rail', () => ({
    ZoneA: () => <div data-testid="zone-a">ZoneA</div>,
}));
vi.mock('@/components/shell/ZoneB_Workspace', () => ({
    ZoneB: ({ children }: { children: React.ReactNode }) => (
        <div data-testid="zone-b">{children}</div>
    ),
}));
vi.mock('@/components/dashboard/DegradedModeBanner', () => ({
    __esModule: true,
    default: () => <div data-testid="degraded-banner">DegradedBanner</div>,
}));
vi.mock('@/components/dashboard/StatusDrawer', () => ({
    __esModule: true,
    default: () => <div data-testid="status-drawer">StatusDrawer</div>,
}));
vi.mock('@/components/stream/SignalDetailDrawer', () => ({
    __esModule: true,
    default: () => <div data-testid="signal-detail-drawer">SignalDetailDrawer</div>,
}));

const setStatusDrawerOpen = vi.fn();
const fetchServicesStatus = vi.fn();
const fetchUserSettings = vi.fn();
const initializeStream = vi.fn();
vi.mock('@/store/useCortexStore', () => ({
    useCortexStore: (selector: any) => selector({ setStatusDrawerOpen, fetchServicesStatus, fetchUserSettings, initializeStream }),
}));

import { ShellLayout } from '@/components/shell/ShellLayout';

describe('ShellLayout', () => {
    it('loads Soma workspace startup dependencies on mount', () => {
        render(<ShellLayout />);
        expect(fetchUserSettings).toHaveBeenCalled();
        expect(fetchServicesStatus).toHaveBeenCalled();
        expect(initializeStream).toHaveBeenCalled();
    });

    it('renders the shell rail and workspace zones', () => {
        render(<ShellLayout />);
        expect(screen.getByTestId('zone-a')).toBeDefined();
        expect(screen.getByTestId('zone-b')).toBeDefined();
    });

    it('applies cortex-bg background class', () => {
        const { container } = render(<ShellLayout />);
        const root = container.firstChild as HTMLElement;
        expect(root.className).toContain('bg-cortex-bg');
    });

    it('renders children inside ZoneB', () => {
        render(
            <ShellLayout>
                <div data-testid="child-content">Hello</div>
            </ShellLayout>
        );
        const zoneB = screen.getByTestId('zone-b');
        expect(zoneB.querySelector('[data-testid="child-content"]')).toBeDefined();
    });

    it('mounts the signal detail drawer at shell level for clickable activity rows', () => {
        render(<ShellLayout />);
        expect(screen.getByTestId('signal-detail-drawer')).toBeDefined();
    });

    it('does not render a floating status button over workspace content', () => {
        render(<ShellLayout />);
        expect(screen.queryByTitle('Open Status Drawer')).toBeNull();
        expect(screen.getByTestId('zone-a')).toBeDefined();
    });
});
