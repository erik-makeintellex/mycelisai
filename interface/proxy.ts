// Phase 0 security: inject MYCELIS_API_KEY into all proxied requests.
// Next.js rewrites proxy /api/* to the Go backend server-to-server.
// This proxy adds the Authorization header so the Go auth middleware
// accepts the request. Zero changes needed to individual fetch() calls.
import { NextRequest, NextResponse } from 'next/server';

export function proxy(request: NextRequest) {
    const apiKey = process.env.MYCELIS_API_KEY || '';
    if (!apiKey) return NextResponse.next();

    const headers = new Headers(request.headers);
    headers.set('Authorization', `Bearer ${apiKey}`);
    return NextResponse.next({ request: { headers } });
}

export const config = {
    matcher: ['/api/:path*', '/admin/:path*', '/agents', '/healthz'],
};
