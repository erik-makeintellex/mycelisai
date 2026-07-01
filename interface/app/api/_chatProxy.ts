import { NextResponse } from 'next/server';

type ProxyTarget = {
    targetLabel: string;
    path: string;
};

function backendURL(path: string): string {
    const host = process.env.MYCELIS_API_HOST ?? 'localhost';
    const port = process.env.MYCELIS_API_PORT ?? '8081';
    return `http://${host}:${port}${path}`;
}

function upstreamHeaders(req: Request): Headers {
    const headers = new Headers();
    const auth = req.headers.get('authorization') || req.headers.get('Authorization') || process.env.MYCELIS_API_KEY;
    if (auth) {
        headers.set('Authorization', auth.startsWith('Bearer ') ? auth : `Bearer ${auth}`);
    }
    const forwardedIdentity = req.headers.get('x-mycelis-web-identity');
    const forwardedIdentitySignature = req.headers.get('x-mycelis-web-identity-signature');
    if (forwardedIdentity && forwardedIdentitySignature) {
        headers.set('X-Mycelis-Web-Identity', forwardedIdentity);
        headers.set('X-Mycelis-Web-Identity-Signature', forwardedIdentitySignature);
    }
    headers.set('Content-Type', req.headers.get('content-type') || 'application/json');
    return headers;
}

function structuredTransportUnavailable(targetLabel: string) {
    return {
        ok: false,
        error: `${targetLabel} is currently unreachable from the workspace runtime.`,
        data: {
            available: false,
            code: 'transport_unavailable',
            summary: `${targetLabel} is currently unreachable from the workspace runtime.`,
            recommended_action: 'Inspect Core connectivity and retry once the local runtime path is healthy.',
        },
    };
}

export async function proxyBackendPostRequest(req: Request, target: ProxyTarget): Promise<Response> {
    return proxyBackendRequest(req, target, 'POST');
}

export async function proxyBackendGetRequest(req: Request, target: ProxyTarget): Promise<Response> {
    return proxyBackendRequest(req, target, 'GET');
}

export async function proxyBackendPatchRequest(req: Request, target: ProxyTarget): Promise<Response> {
    return proxyBackendRequest(req, target, 'PATCH');
}

export async function proxyBackendDeleteRequest(req: Request, target: ProxyTarget): Promise<Response> {
    return proxyBackendRequest(req, target, 'DELETE');
}

async function proxyBackendRequest(req: Request, target: ProxyTarget, method: 'GET' | 'POST' | 'PATCH' | 'DELETE'): Promise<Response> {
    const body = method === 'GET' || method === 'DELETE' ? undefined : await req.text();
    const sourceUrl = new URL(req.url);
    const path = `${target.path}${sourceUrl.search}`;

    try {
        const response = await fetch(backendURL(path), {
            method,
            headers: upstreamHeaders(req),
            body,
        });

        const headers = new Headers();
        const contentType = response.headers.get('content-type');
        if (contentType) {
            headers.set('content-type', contentType);
        }

        return new Response(response.body, {
            status: response.status,
            statusText: response.statusText,
            headers,
        });
    } catch {
        return NextResponse.json(structuredTransportUnavailable(target.targetLabel), {
            status: 503,
        });
    }
}

export async function proxyChatRequest(req: Request, target: ProxyTarget): Promise<Response> {
    return proxyBackendPostRequest(req, target);
}
