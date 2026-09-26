import { NextRequest, NextResponse } from "next/server";
import { WEB_SESSION_COOKIE, requestOriginAllowed, serverOrigin, webAuthRedirectURL } from "@/lib/webAuth";

export async function POST(request: NextRequest) {
    if (!requestOriginAllowed(request.headers, serverOrigin(request))) {
        return new NextResponse("Cross-site sign-out request rejected.", {
            status: 403,
            headers: { "content-type": "text/plain; charset=utf-8", refresh: "0; url=/login?error=origin" },
        });
    }
    const response = NextResponse.redirect(webAuthRedirectURL("/login", serverOrigin(request)), 303);
    response.cookies.delete(WEB_SESSION_COOKIE);
    return response;
}
