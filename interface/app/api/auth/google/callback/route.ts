import { NextRequest, NextResponse } from "next/server";
import {
    WEB_SESSION_COOKIE,
    createSessionToken,
    decodeOAuthStateCookie,
    getWebAuthConfig,
    googleWorkspacePolicy,
    roleForEmail,
    sessionCookieOptions,
    webAuthRedirectURL,
    type WebSession,
} from "@/lib/webAuth";

const STATE_COOKIE = "mycelis_google_state";

interface GoogleTokenInfo {
    sub?: string;
    email?: string;
    email_verified?: string | boolean;
    name?: string;
    hd?: string;
    aud?: string;
    exp?: string;
}

export async function GET(request: NextRequest) {
    const config = getWebAuthConfig();
    const workspacePolicy = googleWorkspacePolicy(config);
    const code = request.nextUrl.searchParams.get("code");
    const state = request.nextUrl.searchParams.get("state");
    const saved = decodeOAuthStateCookie(request.cookies.get(STATE_COOKIE)?.value);
    if (!config.sessionSecret) return redirectToLogin(request, "config");
    if (!code || !state || state !== saved.state) return redirectToLogin(request, "google_state");

    try {
        const tokenResponse = await fetch("https://oauth2.googleapis.com/token", {
            method: "POST",
            headers: { "content-type": "application/x-www-form-urlencoded" },
            body: new URLSearchParams({
                code,
                client_id: config.googleClientId,
                client_secret: config.googleClientSecret,
                redirect_uri: config.googleRedirectUri,
                grant_type: "authorization_code",
            }),
        });
        if (!tokenResponse.ok) {
            await logGoogleAuthFailure("token_exchange", tokenResponse);
            return redirectToLogin(request, "google_token");
        }
        const tokens = await tokenResponse.json() as { id_token?: string };
        if (!tokens.id_token) return redirectToLogin(request, "google_token");

        const infoResponse = await fetch(`https://oauth2.googleapis.com/tokeninfo?id_token=${encodeURIComponent(tokens.id_token)}`);
        if (!infoResponse.ok) {
            await logGoogleAuthFailure("tokeninfo", infoResponse);
            return redirectToLogin(request, "google_tokeninfo");
        }
        const info = await infoResponse.json() as GoogleTokenInfo;
        const emailVerified = info.email_verified === true || info.email_verified === "true";
        if (info.aud !== config.googleClientId || !info.email || !emailVerified) {
            return redirectToLogin(request, "google_identity");
        }
        const domain = (info.hd || info.email.split("@")[1] || "").toLowerCase();
        if (workspacePolicy.allowedDomains.length && !workspacePolicy.allowedDomains.includes(domain)) return redirectToLogin(request, "domain");

        const now = Math.floor(Date.now() / 1000);
        const session: WebSession = {
            sub: info.sub || info.email,
            email: info.email.toLowerCase(),
            name: info.name || info.email,
            role: roleForEmail(info.email, config),
            provider: "google",
            hd: domain,
            iat: now,
            exp: now + 60 * 60 * 8,
        };
        const response = NextResponse.redirect(webAuthRedirectURL(safeNext(saved.next) || "/dashboard", request.nextUrl.origin));
        response.cookies.delete(STATE_COOKIE);
        response.cookies.set(WEB_SESSION_COOKIE, await createSessionToken(session, config.sessionSecret), sessionCookieOptions());
        return response;
    } catch (error) {
        console.warn("[auth/google] callback failed", {
            phase: "exception",
            errorType: error instanceof Error ? error.name : "UnknownError",
        });
        return redirectToLogin(request, "google_exception");
    }
}

function redirectToLogin(request: NextRequest, error: string) {
    const url = webAuthRedirectURL("/login", request.nextUrl.origin);
    url.searchParams.set("error", error);
    const response = NextResponse.redirect(url);
    response.cookies.delete(STATE_COOKIE);
    return response;
}

function safeNext(value: string | null): string {
    return value && value.startsWith("/") && !value.startsWith("//") ? value : "";
}

async function logGoogleAuthFailure(phase: string, response: Response) {
    console.warn("[auth/google] callback failed", {
        phase,
        status: response.status,
    });
}
