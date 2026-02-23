import { describe, it, expect, vi } from 'vitest';

const mockRedirect = vi.fn();
vi.mock('next/navigation', () => ({
    redirect: (url: string) => { mockRedirect(url); throw new Error('NEXT_REDIRECT'); },
}));

import ApprovalsRedirect from '@/app/(app)/approvals/page';

describe('Approvals Page (redirect)', () => {
    it('redirects to /automations?tab=approvals', () => {
        try { ApprovalsRedirect(); } catch {}
        expect(mockRedirect).toHaveBeenCalledWith('/automations?tab=approvals');
    });
});
