import { describe, it, expect, vi } from 'vitest';
import { render, screen, fireEvent } from '@testing-library/react';

// Mock child zones to isolate ShellLayout testing
vi.mock('@/components/shell/ZoneA_Rail', () => ({
    ZoneA: () => <div data-testid="zone-a">ZoneA</div>,
}));
vi.mock('@/components/shell/ZoneB_Workspace', () => ({
    ZoneB: ({ children }: { children: React.ReactNode }) => (
        <div data-testid="zone-b">{children}</div>
    ),
}));
vi.mock('@/components/shell/ZoneD_Decision', () => ({
    ZoneD: () => <div data-testid="zone-d">ZoneD</div>,
}));
vi.mock('@/components/dashboard/DegradedModeBanner', () => ({
    __esModule: true,
    default: () => <div data-testid="degraded-banner">DegradedBanner</div>,
}));
vi.mock('@/components/dashboard/StatusDrawer', () => ({
    __esModule: true,
    default: () => <div data-testid="status-drawer">StatusDrawer</div>,
}));

const setStatusDrawerOpen = vi.fn();
const fetchServicesStatus = vi.fn();
vi.mock('@/store/useCortexStore', () => ({
    useCortexStore: (selector: any) => selector({ setStatusDrawerOpen, fetchServicesStatus }),
}));

import { ShellLayout } from '@/components/shell/ShellLayout';

describe('ShellLayout', () => {
    it('renders all three zones', () => {
        render(<ShellLayout />);
        expect(screen.getByTestId('zone-a')).toBeDefined();
        expect(screen.getByTestId('zone-b')).toBeDefined();
        expect(screen.getByTestId('zone-d')).toBeDefined();
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

    it('opens status drawer from floating status button', () => {
        render(<ShellLayout />);
        fireEvent.click(screen.getByTitle('Open Status Drawer'));
        expect(setStatusDrawerOpen).toHaveBeenCalledWith(true);
    });
});
