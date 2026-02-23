import { describe, it, expect, vi } from 'vitest';

const mockRedirect = vi.fn();
vi.mock('next/navigation', () => ({
    redirect: (url: string) => { mockRedirect(url); throw new Error('NEXT_REDIRECT'); },
}));

import MatrixRedirect from '@/app/(app)/matrix/page';

describe('Matrix Page (redirect)', () => {
    it('redirects to /system?tab=matrix', () => {
        try { MatrixRedirect(); } catch {}
        expect(mockRedirect).toHaveBeenCalledWith('/system?tab=matrix');
    });
});
