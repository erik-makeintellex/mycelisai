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
vi.mock('@/components/shell/ZoneD_Decision', () => ({
    ZoneD: () => <div data-testid="zone-d">ZoneD</div>,
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
});
