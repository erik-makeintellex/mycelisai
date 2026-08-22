import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';
import { POST as postWorkspaceChat } from '@/app/api/v1/chat/route';
import { POST as postCouncilChat } from '@/app/api/v1/council/[member]/chat/route';
import { POST as postLegacyChat } from '@/app/(app)/api/chat/route';
import { GET as getSearchSources, POST as postSearchSource } from '@/app/api/v1/search/sources/route';
import { DELETE as deleteSearchSource, PATCH as patchSearchSource } from '@/app/api/v1/search/sources/[id]/route';

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
            'http://127.0.0.1:8081/api/v1/chat',
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

    it('forwards signed web identity headers to Core for audit context', async () => {
        vi.stubGlobal('fetch', vi.fn().mockResolvedValue(new Response(JSON.stringify({ ok: true, data: { ok: true } }), {
            status: 200,
            headers: { 'content-type': 'application/json' },
        })));

        await postWorkspaceChat(new Request('http://localhost/api/v1/chat', {
            method: 'POST',
            body: JSON.stringify({ messages: [{ role: 'user', content: 'hello' }] }),
            headers: {
                'Content-Type': 'application/json',
                'x-mycelis-web-identity': 'signed-payload',
                'x-mycelis-web-identity-signature': 'signed-proof',
            },
        }));

        const init = vi.mocked(fetch).mock.calls[0]?.[1] as RequestInit;
        const headers = init.headers as Headers;
        expect(headers.get('X-Mycelis-Web-Identity')).toBe('signed-payload');
        expect(headers.get('X-Mycelis-Web-Identity-Signature')).toBe('signed-proof');
    });

	it('forwards the opt-in QA fixture scope without forwarding arbitrary headers', async () => {
		vi.stubGlobal('fetch', vi.fn().mockResolvedValue(new Response(JSON.stringify({ ok: true }), {
			status: 200,
			headers: { 'content-type': 'application/json' },
		})));

		await postWorkspaceChat(new Request('http://localhost/api/v1/chat', {
			method: 'POST',
			body: JSON.stringify({ messages: [] }),
			headers: {
				'Content-Type': 'application/json',
				'x-mycelis-qa-fixture-scope': '11111111-1111-1111-1111-111111111111',
				'x-untrusted-test-header': 'do-not-forward',
			},
		}));

		const init = vi.mocked(fetch).mock.calls[0]?.[1] as RequestInit;
		const headers = init.headers as Headers;
		expect(headers.get('X-Mycelis-QA-Fixture-Scope')).toBe('11111111-1111-1111-1111-111111111111');
		expect(headers.has('x-untrusted-test-header')).toBe(false);
	});

    it('proxies search-source reads and preserves query parameters', async () => {
        vi.stubGlobal('fetch', vi.fn().mockResolvedValue(new Response(JSON.stringify({ ok: true, data: [] }), {
            status: 200,
            headers: { 'content-type': 'application/json' },
        })));

        const response = await getSearchSources(new Request('http://localhost/api/v1/search/sources?scope_kind=group'));

        expect(response.status).toBe(200);
        expect(fetch).toHaveBeenCalledWith(
            'http://127.0.0.1:8081/api/v1/search/sources?scope_kind=group',
            expect.objectContaining({ method: 'GET', headers: expect.any(Headers) }),
        );
    });

    it('proxies search-source mutations through the shared authenticated transport', async () => {
        vi.stubGlobal('fetch', vi.fn().mockResolvedValue(new Response(JSON.stringify({ ok: true }), {
            status: 200,
            headers: { 'content-type': 'application/json' },
        })));
        const body = JSON.stringify({ name: 'Approved docs' });

        await postSearchSource(new Request('http://localhost/api/v1/search/sources', { method: 'POST', body }));
        await patchSearchSource(
            new Request('http://localhost/api/v1/search/sources/approved%20docs', { method: 'PATCH', body }),
            { params: Promise.resolve({ id: 'approved docs' }) },
        );
        await deleteSearchSource(
            new Request('http://localhost/api/v1/search/sources/approved%20docs', { method: 'DELETE' }),
            { params: Promise.resolve({ id: 'approved docs' }) },
        );

        expect(fetch).toHaveBeenNthCalledWith(1, 'http://127.0.0.1:8081/api/v1/search/sources', expect.objectContaining({ method: 'POST', body }));
        expect(fetch).toHaveBeenNthCalledWith(2, 'http://127.0.0.1:8081/api/v1/search/sources/approved%20docs', expect.objectContaining({ method: 'PATCH', body }));
        expect(fetch).toHaveBeenNthCalledWith(3, 'http://127.0.0.1:8081/api/v1/search/sources/approved%20docs', expect.objectContaining({ method: 'DELETE', body: undefined }));
    });
});
