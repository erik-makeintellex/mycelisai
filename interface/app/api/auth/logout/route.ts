import { NextRequest, NextResponse } from "next/server";
import { WEB_SESSION_COOKIE } from "@/lib/webAuth";

export async function POST(request: NextRequest) {
    const response = NextResponse.redirect(new URL("/login", request.headers.get("origin") || request.nextUrl.origin), 303);
    response.cookies.delete(WEB_SESSION_COOKIE);
    return response;
}
