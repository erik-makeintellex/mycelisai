import { NextResponse } from "next/server";
import { cookies } from "next/headers";
import { WEB_SESSION_COOKIE, getWebAuthConfig, googleConfigured, verifySessionToken } from "@/lib/webAuth";

export async function GET() {
    const config = getWebAuthConfig();
    const cookieStore = await cookies();
    const session = await verifySessionToken(cookieStore.get(WEB_SESSION_COOKIE)?.value, config.sessionSecret);
    return NextResponse.json({
        ok: true,
        data: {
            authenticated: Boolean(session),
            user: session ? { email: session.email, name: session.name, role: session.role, provider: session.provider, hd: session.hd } : null,
            providers: {
                local: Boolean(config.sessionSecret && config.localPassword),
                google_workspace: googleConfigured(config),
            },
        },
    });
}
