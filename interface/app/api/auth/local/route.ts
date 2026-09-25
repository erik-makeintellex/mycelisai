import { NextRequest, NextResponse } from "next/server";
import {
    WEB_SESSION_COOKIE,
    createSessionToken,
    getWebAuthConfig,
    localLoginConfigured,
    requestOriginAllowed,
    sessionCookieOptions,
    verifyLocalPassword,
    serverOrigin,
    webAuthRedirectURL,
    type WebSession,
    safeNextPath,
} from "@/lib/webAuth";

export async function POST(request: NextRequest) {
    if (!requestOriginAllowed(request.headers, serverOrigin(request))) return originRejected();
    const config = getWebAuthConfig();
    if (!localLoginConfigured(config)) return redirectToLogin(request, "config");

    const form = await request.formData();
    const username = String(form.get("username") || "").trim();
    const password = String(form.get("password") || "");
    const passwordOk = verifyLocalPassword(password, config);
    if (username !== config.localUsername || !passwordOk) return redirectToLogin(request, "invalid");

    const now = Math.floor(Date.now() / 1000);
    const session: WebSession = {
        sub: "local-owner",
        email: `${config.localUsername}@local.mycelis`,
        name: config.localUsername,
        role: "admin",
        provider: "local",
        iat: now,
        exp: now + 60 * 60 * 8,
    };
    const response = NextResponse.redirect(authRedirectURL(request, safeNextPath(request.nextUrl.searchParams.get("next")) || "/dashboard"), 303);
    response.cookies.set(WEB_SESSION_COOKIE, await createSessionToken(session, config.sessionSecret), sessionCookieOptions());
    return response;
}

function redirectToLogin(request: NextRequest, error: string) {
    const url = authRedirectURL(request, "/login");
    url.searchParams.set("error", error);
    const next = safeNextPath(request.nextUrl.searchParams.get("next"));
    if (next) url.searchParams.set("next", next);
    return NextResponse.redirect(url);
}

// Redirects use MYCELIS_PUBLIC_ORIGIN or the server's own origin, never the
// request Origin header.
function authRedirectURL(request: NextRequest, path: string): URL {
    return webAuthRedirectURL(path, serverOrigin(request));
}

function originRejected() {
    return new NextResponse("Cross-site sign-in request rejected.", {
        status: 403,
        headers: { "content-type": "text/plain; charset=utf-8", refresh: "0; url=/login?error=origin" },
    });
}
