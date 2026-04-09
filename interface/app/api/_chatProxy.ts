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

export async function proxyChatRequest(req: Request, target: ProxyTarget): Promise<Response> {
    const body = await req.text();

    try {
        const response = await fetch(backendURL(target.path), {
            method: 'POST',
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
