import { describe, it, expect, vi } from 'vitest';

const mockRedirect = vi.fn();
vi.mock('next/navigation', () => ({
    redirect: (url: string) => { mockRedirect(url); throw new Error('NEXT_REDIRECT'); },
}));

import WiringRedirect from '@/app/(app)/wiring/page';

describe('Wiring Page (redirect)', () => {
    it('redirects to /automations?tab=wiring', () => {
        try { WiringRedirect(); } catch {}
        expect(mockRedirect).toHaveBeenCalledWith('/automations?tab=wiring');
    });
});
