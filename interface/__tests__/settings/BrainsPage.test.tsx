import { act, fireEvent, render, screen, waitFor } from '@testing-library/react';
import { beforeEach, describe, expect, it } from 'vitest';

import { mockFetch } from '../setup';
import BrainsPage from '@/components/settings/BrainsPage';

const jsonResponse = (body: unknown, ok = true) => ({
    ok,
    json: async () => body,
});

describe('BrainsPage', () => {
    beforeEach(() => {
        mockFetch.mockResolvedValue(jsonResponse({
            ok: true,
            data: [
                {
                    id: 'production_gpt4',
                    type: 'openai',
                    endpoint: 'https://api.openai.com/v1',
                    model_id: 'gpt-4-turbo',
                    location: 'remote',
                    data_boundary: 'leaves_org',
                    usage_policy: 'require_approval',
                    token_budget_profile: 'extended',
                    max_output_tokens: 2048,
                    roles_allowed: ['all'],
                    enabled: true,
                    status: 'online',
                },
            ],
        }));
    });

    it('shows token budget details in the provider table', async () => {
        await act(async () => {
            render(<BrainsPage />);
        });

        expect(await screen.findByText('Token Budget')).toBeDefined();
        expect(screen.getByText('Extended')).toBeDefined();
        expect(screen.getByText('2048 max')).toBeDefined();
    });

    it('applies hosted-provider preset token defaults in the add modal', async () => {
        await act(async () => {
            render(<BrainsPage />);
        });

        fireEvent.click(await screen.findByText('Add Provider'));
        fireEvent.click(screen.getByRole('button', { name: 'OpenAI' }));

        await waitFor(() => {
            expect(screen.getByDisplayValue('https://api.openai.com/v1')).toBeDefined();
            expect(screen.getByDisplayValue('2048')).toBeDefined();
        });
    });

    it('explains a provider_bound 409 instead of silently snapping the toggle back', async () => {
        await act(async () => {
            render(<BrainsPage />);
        });
        await screen.findByText('Token Budget');

        mockFetch.mockResolvedValueOnce(jsonResponse({
            data: { code: 'provider_bound', profiles: ['chat', 'coder'] },
        }, false));

        const toggles = screen.getAllByRole('button').filter((el) => el.className.includes('rounded-full'));
        fireEvent.click(toggles[0]);

        expect(await screen.findByRole('alert')).toBeDefined();
        expect(screen.getByText('This AI engine is in use')).toBeDefined();
        expect(screen.getByText('Use default engine for chat')).toBeDefined();
        expect(screen.getByText('Use default engine for coder')).toBeDefined();
    });

    it('resets a profile override and clears the blocker when "Use default engine" is clicked', async () => {
        await act(async () => {
            render(<BrainsPage />);
        });
        await screen.findByText('Token Budget');

        mockFetch.mockResolvedValueOnce(jsonResponse({
            data: { code: 'provider_bound', profiles: ['chat'] },
        }, false));
        const toggles = screen.getAllByRole('button').filter((el) => el.className.includes('rounded-full'));
        fireEvent.click(toggles[0]);
        await screen.findByRole('alert');

        mockFetch.mockResolvedValueOnce({ ok: true, json: async () => ({}) });
        mockFetch.mockResolvedValueOnce(jsonResponse({ ok: true, data: [] }));

        fireEvent.click(screen.getByText('Use default engine for chat'));

        await waitFor(() => {
            expect(screen.queryByRole('alert')).toBeNull();
        });
        expect(mockFetch).toHaveBeenCalledWith('/api/v1/cognitive/profiles/chat/override', { method: 'DELETE' });
    });
});
