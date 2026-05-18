import { NextRequest, NextResponse } from "next/server";
import {
    WEB_SESSION_COOKIE,
    createSessionToken,
    getWebAuthConfig,
    sessionCookieOptions,
    sha256Hex,
    webAuthRedirectURL,
    type WebSession,
} from "@/lib/webAuth";

export async function POST(request: NextRequest) {
    const config = getWebAuthConfig();
    if (!config.sessionSecret || !config.localPassword) return redirectToLogin(request, "config");

    const form = await request.formData();
    const username = String(form.get("username") || "").trim();
    const password = String(form.get("password") || "");
    const passwordOk = config.localPasswordSha256
        ? (await sha256Hex(password)) === config.localPasswordSha256.toLowerCase()
        : password === config.localPassword;
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
    const response = NextResponse.redirect(authRedirectURL(request, safeNext(request.nextUrl.searchParams.get("next")) || "/dashboard"), 303);
    response.cookies.set(WEB_SESSION_COOKIE, await createSessionToken(session, config.sessionSecret), sessionCookieOptions());
    return response;
}

function redirectToLogin(request: NextRequest, error: string) {
    const url = authRedirectURL(request, "/login");
    url.searchParams.set("error", error);
    const next = safeNext(request.nextUrl.searchParams.get("next"));
    if (next) url.searchParams.set("next", next);
    return NextResponse.redirect(url);
}

function authRedirectURL(request: NextRequest, path: string): URL {
    return webAuthRedirectURL(path, request.headers.get("origin") || request.nextUrl.origin);
}

function safeNext(value: string | null): string {
    return value && value.startsWith("/") && !value.startsWith("//") ? value : "";
}
