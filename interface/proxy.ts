import { NextRequest, NextResponse } from 'next/server';
import { WEB_SESSION_COOKIE, createForwardedWebIdentityHeaders, getWebAuthConfig, verifySessionToken } from '@/lib/webAuth';

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
            return NextResponse.json({ ok: false, error: 'authentication_required' }, { status: 401 });
        }
        const login = new URL('/login', request.url);
        login.searchParams.set('next', request.nextUrl.pathname + request.nextUrl.search);
        return NextResponse.redirect(login);
    }
    if (session?.role === 'standard' && isAdminPath(request.nextUrl)) {
        return NextResponse.redirect(new URL('/access-denied', request.url));
    }

    const apiKey = process.env.MYCELIS_API_KEY || '';
    if (!apiKey || !isBackendProxyPath(request.nextUrl.pathname)) return NextResponse.next();

    const headers = new Headers(request.headers);
    headers.set('Authorization', `Bearer ${apiKey}`);
    if (session) {
        const identityHeaders = await createForwardedWebIdentityHeaders(
            session,
            process.env.MYCELIS_WEB_IDENTITY_FORWARD_SECRET || config.sessionSecret,
        );
        for (const [key, value] of Object.entries(identityHeaders)) {
            headers.set(key, value);
        }
    }
    return NextResponse.next({ request: { headers } });
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
