import { describe, it, expect, vi, beforeEach } from 'vitest';
import { render, screen, fireEvent, act } from '@testing-library/react';
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

        expect(await screen.findByText('Documentation and guidance')).toBeDefined();
        const labels = await screen.findAllByText('UI Generation Spec');
        expect(labels.length).toBeGreaterThan(0);
        expect(await screen.findByRole('heading', { name: 'Terminal States' })).toBeDefined();
        expect(routerReplace).toHaveBeenCalledWith('/docs?doc=ui-generation-spec', { scroll: false });
    });

    it('follows internal markdown links through the in-app docs manifest', async () => {
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
                                    {
                                        slug: 'workflow-variants-and-plan-memory',
                                        label: 'Workflow Variants And Plan Memory',
                                        path: 'docs/user/workflow-variants-and-plan-memory.md',
                                        description: 'Workflow variants doc',
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
                        content: '# Terminal States\n\nSee [Workflow variants](workflow-variants-and-plan-memory.md).',
                    }),
                } as Response;
            }
            if (url === '/docs-api/workflow-variants-and-plan-memory') {
                return {
                    ok: true,
                    json: async () => ({
                        slug: 'workflow-variants-and-plan-memory',
                        label: 'Workflow Variants And Plan Memory',
                        content: '# Workflow Variants\n\nCompact lanes stay visible.',
                    }),
                } as Response;
            }
            throw new Error(`unexpected fetch: ${url}`);
        });

        render(<DocsPage />);

        const internalLink = await screen.findByRole('link', { name: 'Workflow variants' });
        fireEvent.click(internalLink);

        expect(await screen.findByRole('heading', { name: 'Workflow Variants' })).toBeDefined();
        expect(routerReplace).toHaveBeenLastCalledWith('/docs?doc=workflow-variants-and-plan-memory', { scroll: false });
    });

    it('uses a list-to-article flow with an obvious return action on mobile', async () => {
        docSearchParams.delete('doc');
        mockFetch.mockImplementation(async (input) => {
            const url = String(input);
            if (url === '/docs-api') {
                return {
                    ok: true,
                    json: async () => ({
                        sections: [
                            {
                                section: 'Start here',
                                docs: [
                                    {
                                        slug: 'getting-started',
                                        label: 'Getting Started',
                                        path: 'docs/user/getting-started.md',
                                    },
                                    {
                                        slug: 'working-with-soma',
                                        label: 'Working With Soma',
                                        path: 'docs/user/working-with-soma.md',
                                    },
                                ],
                            },
                        ],
                    }),
                } as Response;
            }
            if (url.startsWith('/docs-api/')) {
                const slug = url.split('/').pop();
                return {
                    ok: true,
                    json: async () => ({
                        slug,
                        label: slug === 'working-with-soma' ? 'Working With Soma' : 'Getting Started',
                        content: slug === 'working-with-soma' ? '# Work With Soma' : '# Getting Started',
                    }),
                } as Response;
            }
            throw new Error(`unexpected fetch: ${url}`);
        });

        render(<DocsPage />);

        const navigation = await screen.findByTestId('docs-navigation-pane');
        const article = screen.getByTestId('docs-article-pane');
        expect(navigation.classList.contains('hidden')).toBe(false);
        expect(article.classList.contains('hidden')).toBe(true);

        fireEvent.click(screen.getByRole('button', { name: 'Working With Soma' }));
        expect(await screen.findByRole('heading', { name: 'Work With Soma' })).toBeDefined();
        expect(navigation.classList.contains('hidden')).toBe(true);
        expect(article.classList.contains('hidden')).toBe(false);

        fireEvent.click(screen.getByRole('button', { name: 'All docs' }));
        expect(navigation.classList.contains('hidden')).toBe(false);
        expect(article.classList.contains('hidden')).toBe(true);
    });

    it('shows manifest error state when docs manifest fetch fails', async () => {
        mockFetch.mockRejectedValueOnce(new Error('manifest-down'));

        render(<DocsPage />);

        expect(await screen.findByText(/Failed to load doc manifest/)).toBeDefined();
    });

    it('does not replace navigation after the user leaves while the manifest is loading', async () => {
        let resolveManifest: ((response: Response) => void) | undefined;
        mockFetch.mockImplementation(() => new Promise<Response>((resolve) => {
            resolveManifest = resolve;
        }));

        const { unmount } = render(<DocsPage />);
        unmount();

        await act(async () => {
            resolveManifest?.({
                ok: true,
                json: async () => ({
                    sections: [{
                        section: 'Start here',
                        docs: [{ slug: 'getting-started', label: 'Getting Started', path: 'docs/user/getting-started.md' }],
                    }],
                }),
            } as Response);
            await Promise.resolve();
        });

        expect(routerReplace).not.toHaveBeenCalled();
    });

    it('shows a readable doc-load error when a selected doc fetch fails', async () => {
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
                    ok: false,
                    status: 503,
                    json: async () => ({}),
                } as Response;
            }
            throw new Error(`unexpected fetch: ${url}`);
        });

        render(<DocsPage />);

        expect(await screen.findByText('Failed to load "UI Generation Spec": HTTP 503')).toBeDefined();
    });
});
