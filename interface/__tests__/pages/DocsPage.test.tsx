import { describe, it, expect, vi, beforeEach } from 'vitest';
import { render, screen } from '@testing-library/react';
import { mockFetch } from '../setup';

const routerReplace = vi.fn();
const docSearchParams = new URLSearchParams();

vi.mock('next/navigation', () => ({
    useRouter: () => ({
        push: vi.fn(),
        replace: routerReplace,
        back: vi.fn(),
        prefetch: vi.fn(),
    }),
    useSearchParams: () => docSearchParams,
}));

import DocsPage from '@/app/(app)/docs/page';

describe('DocsPage', () => {
    beforeEach(() => {
        routerReplace.mockReset();
        docSearchParams.set('doc', 'ui-generation-spec');
    });

    it('loads requested doc from manifest and renders markdown content', async () => {
        mockFetch.mockImplementation(async (input) => {
            const url = String(input);
            if (url === '/docs-api') {
                return {
                    ok: true,
                    json: async () => ({
                        sections: [
                            {
                                section: 'Architecture',
                                docs: [
                                    {
                                        slug: 'ui-generation-spec',
                                        label: 'UI Generation Spec',
                                        path: 'docs/architecture/UI_GENERATION.md',
                                        description: 'UI generation test doc',
                                    },
                                ],
                            },
                        ],
                    }),
                } as Response;
            }
            if (url === '/docs-api/ui-generation-spec') {
                return {
                    ok: true,
                    json: async () => ({
                        slug: 'ui-generation-spec',
                        label: 'UI Generation Spec',
                        content: '# Terminal States\n\n- answer\n- proposal',
                    }),
                } as Response;
            }
            throw new Error(`unexpected fetch: ${url}`);
        });

        render(<DocsPage />);

        expect(await screen.findByText('Documentation')).toBeDefined();
        const labels = await screen.findAllByText('UI Generation Spec');
        expect(labels.length).toBeGreaterThan(0);
        expect(await screen.findByRole('heading', { name: 'Terminal States' })).toBeDefined();
        expect(routerReplace).toHaveBeenCalledWith('/docs?doc=ui-generation-spec', { scroll: false });
    });

    it('shows manifest error state when docs manifest fetch fails', async () => {
        mockFetch.mockRejectedValueOnce(new Error('manifest-down'));

        render(<DocsPage />);

        expect(await screen.findByText('Failed to load doc manifest')).toBeDefined();
    });
});
