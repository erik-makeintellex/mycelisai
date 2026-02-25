import { describe, it, expect, vi } from 'vitest';

// The approvals page now redirects to /automations?tab=approvals
const mockRedirect = vi.fn();
vi.mock('next/navigation', () => ({
    redirect: (url: string) => { mockRedirect(url); throw new Error('NEXT_REDIRECT'); },
    useRouter: () => ({ push: vi.fn(), replace: vi.fn(), back: vi.fn(), prefetch: vi.fn() }),
    usePathname: () => '/approvals',
}));

import ApprovalsPage from '../app/(app)/approvals/page';

describe('Approvals Page (V7 redirect)', () => {
    it('redirects to /automations?tab=approvals', () => {
        try { ApprovalsPage(); } catch {}
        expect(mockRedirect).toHaveBeenCalledWith('/automations?tab=approvals');
    });
});
