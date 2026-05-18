import { NextRequest, NextResponse } from "next/server";
import { getWebAuthConfig, googleConfigured, secureCookieEnabled, webAuthRedirectURL } from "@/lib/webAuth";

const STATE_COOKIE = "mycelis_google_state";

export async function GET(request: NextRequest) {
    const config = getWebAuthConfig();
    if (!config.sessionSecret || !googleConfigured(config)) return redirectToLogin(request, "config");

    const state = crypto.randomUUID();
    const authURL = new URL("https://accounts.google.com/o/oauth2/v2/auth");
    authURL.searchParams.set("client_id", config.googleClientId);
    authURL.searchParams.set("redirect_uri", config.googleRedirectUri);
    authURL.searchParams.set("response_type", "code");
    authURL.searchParams.set("scope", "openid email profile");
    authURL.searchParams.set("state", state);
    authURL.searchParams.set("access_type", "online");
    authURL.searchParams.set("prompt", "select_account");
    if (config.googleHostedDomain) authURL.searchParams.set("hd", config.googleHostedDomain);

    const response = NextResponse.redirect(authURL);
    response.cookies.set(STATE_COOKIE, `${state}:${safeNext(request.nextUrl.searchParams.get("next")) || "/dashboard"}`, {
        httpOnly: true,
        sameSite: "lax",
        secure: secureCookieEnabled(),
        path: "/",
        maxAge: 60 * 10,
    });
    return response;
}

function redirectToLogin(request: NextRequest, error: string) {
    const url = webAuthRedirectURL("/login", request.nextUrl.origin);
    url.searchParams.set("error", error);
    return NextResponse.redirect(url);
}

function safeNext(value: string | null): string {
    return value && value.startsWith("/") && !value.startsWith("//") ? value : "";
}
