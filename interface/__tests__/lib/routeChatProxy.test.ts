import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';
import { POST as postWorkspaceChat } from '@/app/api/v1/chat/route';
import { POST as postCouncilChat } from '@/app/api/v1/council/[member]/chat/route';
import { POST as postLegacyChat } from '@/app/(app)/api/chat/route';

describe('chat proxy routes', () => {
    const originalApiKey = process.env.MYCELIS_API_KEY;

    beforeEach(() => {
        process.env.MYCELIS_API_KEY = 'test-api-key';
    });

    afterEach(() => {
        process.env.MYCELIS_API_KEY = originalApiKey;
        vi.restoreAllMocks();
    });

    it('passes through upstream workspace chat blocker envelopes', async () => {
        vi.stubGlobal('fetch', vi.fn().mockResolvedValue(new Response(JSON.stringify({
            ok: false,
            error: 'Soma is currently unreachable from the workspace runtime.',
            data: {
                available: false,
                code: 'transport_unavailable',
                summary: 'Soma is currently unreachable from the workspace runtime.',
            },
        }), {
            status: 503,
            headers: { 'content-type': 'application/json' },
        })));

        const response = await postWorkspaceChat(new Request('http://localhost/api/v1/chat', {
            method: 'POST',
            body: JSON.stringify({ messages: [{ role: 'user', content: 'hello' }] }),
            headers: { 'Content-Type': 'application/json' },
        }));

        expect(response.status).toBe(503);
        expect(response.headers.get('content-type')).toContain('application/json');
        await expect(response.json()).resolves.toMatchObject({
            ok: false,
            data: { code: 'transport_unavailable' },
        });
        expect(fetch).toHaveBeenCalledWith(
            'http://localhost:8081/api/v1/chat',
            expect.objectContaining({
                method: 'POST',
                headers: expect.any(Headers),
                body: JSON.stringify({ messages: [{ role: 'user', content: 'hello' }] }),
            }),
        );
    });

    it('returns a structured council transport blocker when the upstream fetch fails', async () => {
        vi.stubGlobal('fetch', vi.fn().mockRejectedValue(new Error('socket hang up')));

        const response = await postCouncilChat(
            new Request('http://localhost/api/v1/council/council-sentry/chat', {
                method: 'POST',
                body: JSON.stringify({ messages: [{ role: 'user', content: 'hello' }] }),
                headers: { 'Content-Type': 'application/json' },
            }),
            { params: Promise.resolve({ member: 'council-sentry' }) },
        );

        expect(response.status).toBe(503);
        await expect(response.json()).resolves.toMatchObject({
            ok: false,
            data: {
                code: 'transport_unavailable',
                summary: 'Council member council-sentry is currently unreachable from the workspace runtime.',
            },
        });
    });

    it('reuses the same proxy contract for the legacy chat route', async () => {
        vi.stubGlobal('fetch', vi.fn().mockResolvedValue(new Response(JSON.stringify({ ok: true, data: { ok: true } }), {
            status: 200,
            headers: { 'content-type': 'application/json' },
        })));

        const response = await postLegacyChat(new Request('http://localhost/api/chat', {
            method: 'POST',
            body: JSON.stringify({ messages: [{ role: 'user', content: 'hello' }] }),
            headers: { 'Content-Type': 'application/json' },
        }));

        expect(response.status).toBe(200);
        await expect(response.json()).resolves.toMatchObject({ ok: true });
    });
});
