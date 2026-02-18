import { describe, it, expect, vi } from 'vitest';
import { render, screen, act } from '@testing-library/react';

// Mock reactflow (store imports it, even if this page doesn't directly use it)
vi.mock('reactflow', async () => {
    const mock = await import('../mocks/reactflow');
    return mock;
});

import TelemetryPage from '@/app/telemetry/page';

describe('Telemetry Page (app/telemetry/page.tsx)', () => {
    it('mounts without crashing', async () => {
        await act(async () => {
            render(<TelemetryPage />);
        });

        expect(screen.getByText('System Telemetry')).toBeDefined();
    });

    it('renders status cards', async () => {
        await act(async () => {
            render(<TelemetryPage />);
        });

        expect(screen.getByText('Core Nodes')).toBeDefined();
        expect(screen.getByText('Postgres')).toBeDefined();
        expect(screen.getByText('NATS Bus')).toBeDefined();
        expect(screen.getByText('Guard')).toBeDefined();
    });

    it('shows live indicator', async () => {
        await act(async () => {
            render(<TelemetryPage />);
        });

        expect(screen.getByText('LIVE')).toBeDefined();
    });
});
