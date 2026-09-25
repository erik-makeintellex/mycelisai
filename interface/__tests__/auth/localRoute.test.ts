import { createHash } from "node:crypto";
import { NextRequest } from "next/server";
import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";
import { POST as localLogin } from "@/app/api/auth/local/route";
import { POST as logout } from "@/app/api/auth/logout/route";
import { GET as getSession } from "@/app/api/auth/session/route";
import { createSessionToken } from "@/lib/webAuth";

const { cookieJar } = vi.hoisted(() => ({ cookieJar: new Map<string, string>() }));
vi.mock("next/headers", () => ({
    cookies: vi.fn(async () => ({ get: (name: string) => (cookieJar.has(name) ? { value: cookieJar.get(name) } : undefined) })),
}));

const API_KEY = "api-key-0123456789abcdef0123456789abcdef";
const AUTH_ENV = [
    "MYCELIS_API_KEY",
    "MYCELIS_BREAK_GLASS_API_KEY",
    "MYCELIS_WEB_SESSION_SECRET",
    "MYCELIS_WEB_IDENTITY_FORWARD_SECRET",
    "MYCELIS_LOCAL_ADMIN_USERNAME",
    "MYCELIS_LOCAL_ADMIN_PASSWORD",
    "MYCELIS_LOCAL_ADMIN_PASSWORD_SHA256",
    "MYCELIS_PUBLIC_ORIGIN",
];
const OPEN_REDIRECT_NEXTS = ["/\\evil.example", "/%5Cevil.example", "/\t/evil.example", "//evil.example", "/%2F%2Fevil.example", "https://evil.example", "javascript:alert(1)"];
const saved = new Map<string, string | undefined>();

function configure(values: Record<string, string>) {
    for (const key of AUTH_ENV) process.env[key] = values[key] ?? "";
}

function loginRequest(password: string, headers: Record<string, string> = {}, url = "http://127.0.0.1:3000/auth/local?next=%2Fdashboard") {
    const body = new URLSearchParams({ username: "admin", password });
    return new NextRequest(url, {
        method: "POST",
        body,
        headers: { "content-type": "application/x-www-form-urlencoded", host: "127.0.0.1:3000", ...headers },
    });
}

const ready = {
    MYCELIS_API_KEY: API_KEY,
    MYCELIS_WEB_SESSION_SECRET: "session-secret-0123456789abcdef01234567",
    MYCELIS_WEB_IDENTITY_FORWARD_SECRET: "forward-secret-0123456789abcdef01234567",
    MYCELIS_LOCAL_ADMIN_PASSWORD_SHA256: createHash("sha256").update("correct horse battery staple").digest("hex"),
};

describe("local login route", () => {
    beforeEach(() => {
        for (const key of AUTH_ENV) saved.set(key, process.env[key]);
    });
    afterEach(() => {
        for (const key of AUTH_ENV) {
            const previous = saved.get(key);
            if (previous === undefined) delete process.env[key];
            else process.env[key] = previous;
        }
    });

    it("does not accept the API key as password when no local password is configured", async () => {
        configure({ ...ready, MYCELIS_LOCAL_ADMIN_PASSWORD_SHA256: "" });
        const response = await localLogin(loginRequest(API_KEY, { origin: "http://127.0.0.1:3000" }));
        expect(response.headers.get("set-cookie") ?? "").not.toContain("mycelis_web_session=");
        expect(new URL(response.headers.get("location") ?? "").searchParams.get("error")).toBe("config");
    });

    it("signs in with the hashed local password and redirects to the server origin", async () => {
        configure(ready);
        const response = await localLogin(loginRequest("correct horse battery staple", { origin: "http://127.0.0.1:3000" }));
        expect(response.status).toBe(303);
        expect(response.headers.get("location")).toBe("http://127.0.0.1:3000/dashboard");
        expect(response.headers.get("set-cookie") ?? "").toContain("mycelis_web_session=");
    });

    it.each(OPEN_REDIRECT_NEXTS)("never redirects off-origin after sign-in for next=%j", async (next) => {
        configure(ready);
        const url = `http://127.0.0.1:3000/auth/local?next=${encodeURIComponent(next)}`;
        const response = await localLogin(loginRequest("correct horse battery staple", { origin: "http://127.0.0.1:3000" }, url));
        expect(response.headers.get("set-cookie") ?? "").toContain("mycelis_web_session=");
        expect(response.headers.get("location")).toBe("http://127.0.0.1:3000/dashboard");
    });

    it("keeps a same-origin next path with query and fragment", async () => {
        configure(ready);
        const url = `http://127.0.0.1:3000/auth/local?next=${encodeURIComponent("/dashboard?x=1#y")}`;
        const response = await localLogin(loginRequest("correct horse battery staple", { origin: "http://127.0.0.1:3000" }, url));
        expect(response.headers.get("location")).toBe("http://127.0.0.1:3000/dashboard?x=1#y");
    });

    it("rejects a wrong password without setting a cookie", async () => {
        configure(ready);
        const response = await localLogin(loginRequest("wrong", { origin: "http://127.0.0.1:3000" }));
        expect(response.headers.get("set-cookie") ?? "").not.toContain("mycelis_web_session=");
        expect(new URL(response.headers.get("location") ?? "").searchParams.get("error")).toBe("invalid");
    });

    it.each([
        ["foreign Origin", { origin: "https://evil.example" }],
        ["cross-site fetch metadata", { "sec-fetch-site": "cross-site" }],
    ])("rejects a cross-site login POST (%s) with 403 and no cookie", async (_label, headers) => {
        configure(ready);
        const response = await localLogin(loginRequest("correct horse battery staple", headers));
        expect(response.status).toBe(403);
        expect(response.headers.get("set-cookie") ?? "").not.toContain("mycelis_web_session=");
        const location = response.headers.get("location") ?? "";
        expect(location).not.toContain("evil.example");
        expect(response.headers.get("refresh") ?? "").toContain("/login?error=origin");
    });

    it("never builds redirects from the Origin header", async () => {
        configure({ ...ready, MYCELIS_PUBLIC_ORIGIN: "https://mycelis.example" });
        const response = await localLogin(loginRequest("wrong", { origin: "https://mycelis.example" }));
        expect(new URL(response.headers.get("location") ?? "").origin).toBe("https://mycelis.example");

        configure(ready);
        const noOrigin = await localLogin(loginRequest("wrong"));
        expect(new URL(noOrigin.headers.get("location") ?? "").origin).toBe("http://127.0.0.1:3000");
    });
});

describe("session route", () => {
    beforeEach(() => {
        for (const key of AUTH_ENV) saved.set(key, process.env[key]);
    });
    afterEach(() => {
        cookieJar.clear();
        for (const key of AUTH_ENV) {
            const previous = saved.get(key);
            if (previous === undefined) delete process.env[key];
            else process.env[key] = previous;
        }
    });

    it("rejects a cookie signed with the API key and reports local sign-in unavailable", async () => {
        configure({ ...ready, MYCELIS_WEB_SESSION_SECRET: "" });
        const now = Math.floor(Date.now() / 1000);
        cookieJar.set("mycelis_web_session", await createSessionToken({ sub: "x", email: "x@example.com", name: "X", role: "admin", provider: "local", iat: now, exp: now + 60 }, API_KEY));
        const body = await (await getSession()).json();
        expect(body.data.authenticated).toBe(false);
        expect(body.data.providers.local).toBe(false);
    });

    it("reports local sign-in available with a hashed password only", async () => {
        configure(ready);
        const body = await (await getSession()).json();
        expect(body.data.providers.local).toBe(true);
    });
});

describe("logout route", () => {
    it("rejects a cross-site logout and keeps redirects on the server origin", async () => {
        const forged = await logout(new NextRequest("http://127.0.0.1:3000/auth/logout", {
            method: "POST",
            headers: { origin: "https://evil.example", host: "127.0.0.1:3000" },
        }));
        expect(forged.status).toBe(403);
        expect(forged.headers.get("location") ?? "").not.toContain("evil.example");

        const response = await logout(new NextRequest("http://127.0.0.1:3000/auth/logout", {
            method: "POST",
            headers: { origin: "http://127.0.0.1:3000", host: "127.0.0.1:3000" },
        }));
        expect(response.status).toBe(303);
        expect(new URL(response.headers.get("location") ?? "").origin).toBe("http://127.0.0.1:3000");
    });
});
