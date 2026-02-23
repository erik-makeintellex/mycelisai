import { describe, it, expect, vi } from 'vitest';

const mockRedirect = vi.fn();
vi.mock('next/navigation', () => ({
    redirect: (url: string) => { mockRedirect(url); throw new Error('NEXT_REDIRECT'); },
}));

import TelemetryRedirect from '@/app/(app)/telemetry/page';

describe('Telemetry Page (redirect)', () => {
    it('redirects to /system?tab=health', () => {
        try { TelemetryRedirect(); } catch {}
        expect(mockRedirect).toHaveBeenCalledWith('/system?tab=health');
    });
});
