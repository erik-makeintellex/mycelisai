import { NextRequest, NextResponse } from 'next/server';
import { WEB_SESSION_COOKIE, createForwardedWebIdentityHeaders, getWebAuthConfig, verifySessionToken } from '@/lib/webAuth';

// Browser-supplied authority headers are never trusted. The proxy removes them
// on every backend-bound request and sets only its own.
const INBOUND_AUTHORITY_HEADERS = [
    'authorization',
    'x-mycelis-web-identity',
    'x-mycelis-web-identity-signature',
    'x-mycelis-qa-fixture-scope',
];

const PUBLIC_PATH_PREFIXES = [
    '/login',
    '/auth',
    '/api/auth',
    '/_next',
    '/favicon.ico',
    '/grid.svg',
    '/healthz',
];

export async function proxy(request: NextRequest) {
    const config = getWebAuthConfig();
    const session = await verifySessionToken(
        request.cookies.get(WEB_SESSION_COOKIE)?.value,
        config.sessionSecret,
    );
    if (!session && !isPublicPath(request.nextUrl.pathname)) {
        if (request.nextUrl.pathname.startsWith('/api/')) {
            if (shouldRedirectUnauthenticatedApiRequest(request.headers)) {
                return redirectToLogin(request);
            }
            return NextResponse.json({ ok: false, error: 'authentication_required' }, { status: 401 });
        }
        return redirectToLogin(request);
    }
    if (session?.role === 'standard' && isAdminPath(request.nextUrl)) {
        return NextResponse.redirect(new URL('/access-denied', request.url));
    }

    if (!isBackendProxyPath(request.nextUrl.pathname)) return NextResponse.next();

    const headers = new Headers(request.headers);
    for (const name of INBOUND_AUTHORITY_HEADERS) headers.delete(name);
    const apiKey = process.env.MYCELIS_API_KEY || '';
    if (!apiKey) return NextResponse.next({ request: { headers } });

    headers.set('Authorization', `Bearer ${apiKey}`);
    if (session) {
        // Without a usable forward secret a session would reach Core as the
        // API-key owner. Fail closed instead of widening authority.
        if (!config.forwardSecret) {
            return NextResponse.json({ ok: false, error: 'auth_configuration', data: { missing: 'MYCELIS_WEB_IDENTITY_FORWARD_SECRET' } }, { status: 503 });
        }
        const identityHeaders = await createForwardedWebIdentityHeaders(session, config.forwardSecret);
        for (const [key, value] of Object.entries(identityHeaders)) {
            headers.set(key, value);
        }
    }
    return NextResponse.next({ request: { headers } });
}

function redirectToLogin(request: NextRequest): NextResponse {
    const login = new URL('/login', request.url);
    login.searchParams.set('next', '/dashboard');
    return NextResponse.redirect(login);
}

export function shouldRedirectUnauthenticatedApiRequest(headers: Headers): boolean {
    const accept = headers.get('accept') || '';
    const fetchMode = headers.get('sec-fetch-mode') || '';
    const fetchDest = headers.get('sec-fetch-dest') || '';
    return fetchMode === 'navigate' || fetchDest === 'document' || accept.includes('text/html');
}

function isPublicPath(pathname: string): boolean {
    return PUBLIC_PATH_PREFIXES.some((prefix) => pathname === prefix || pathname.startsWith(`${prefix}/`));
}

function isBackendProxyPath(pathname: string): boolean {
    if (pathname === '/api/auth' || pathname.startsWith('/api/auth/')) return false;
    return pathname.startsWith('/api/') || pathname.startsWith('/admin/') || pathname === '/agents' || pathname === '/healthz';
}

function isAdminPath(url: URL): boolean {
    if (url.pathname === '/system') return true;
    if (url.pathname !== '/settings') return false;
    const tab = url.searchParams.get('tab');
    return tab === 'users' || tab === 'auth' || tab === 'engines' || tab === 'tools';
}

export const config = {
    matcher: ['/', '/((?!_next/static|_next/image|favicon.ico).*)'],
};
