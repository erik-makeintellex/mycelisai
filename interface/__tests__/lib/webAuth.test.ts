import { afterEach, describe, expect, it } from "vitest";
import { createForwardedWebIdentityHeaders, createSessionToken, decodeOAuthStateCookie, encodeOAuthStateCookie, getWebAuthConfig, safeNextPath, googleWorkspacePolicy, roleForEmail, sha256Hex, splitList, verifySessionToken, webAuthRedirectURL, type WebSession } from "@/lib/webAuth";

describe("webAuth", () => {
    it("signs and verifies web sessions", async () => {
        const now = Math.floor(Date.now() / 1000);
        const session: WebSession = {
            sub: "user-1",
            email: "admin@example.com",
            name: "Admin",
            role: "admin",
            provider: "local",
            iat: now,
            exp: now + 60,
        };
        const token = await createSessionToken(session, "secret");
        await expect(verifySessionToken(token, "secret")).resolves.toMatchObject({ email: "admin@example.com", role: "admin" });
        await expect(verifySessionToken(token, "wrong")).resolves.toBeNull();
    });

    it("maps explicit admin emails and list config", () => {
        const config = {
            sessionSecret: "secret",
            forwardSecret: "",
            setupIssues: [] as string[],
            localUsername: "admin",
            localPassword: "pw",
            localPasswordSha256: "",
            googleClientId: "",
            googleClientSecret: "",
            googleRedirectUri: "",
            googleHostedDomain: "example.com",
            allowedDomains: splitList("example.com, other.example"),
            adminEmails: splitList("owner@example.com"),
        };
        expect(roleForEmail("owner@example.com", config)).toBe("admin");
        expect(roleForEmail("user@example.com", config)).toBe("standard");
    });

    it("derives Google Workspace display and enforcement policy from one config object", () => {
        const config = {
            sessionSecret: "secret",
            forwardSecret: "",
            setupIssues: [] as string[],
            localUsername: "admin",
            localPassword: "pw",
            localPasswordSha256: "",
            googleClientId: "client",
            googleClientSecret: "secret",
            googleRedirectUri: "http://127.0.0.1:3000/auth/google/callback",
            googleHostedDomain: "Primary.Example",
            allowedDomains: splitList("primary.example,secondary.example"),
            adminEmails: [],
        };

        expect(googleWorkspacePolicy(config)).toEqual({
            hostedDomain: "primary.example",
            allowedDomains: ["primary.example", "secondary.example"],
            displayDomains: ["primary.example", "secondary.example"],
            domainLabel: "primary.example, secondary.example",
        });
    });

    it("supports hashed local admin password comparison material", async () => {
        await expect(sha256Hex("correct horse battery staple")).resolves.toMatch(/^[a-f0-9]{64}$/);
    });

    it("creates signed forwarded identity headers for Core audit propagation", async () => {
        const now = Math.floor(Date.now() / 1000);
        const session: WebSession = {
            sub: "google-123",
            email: "erik@mycelis.link",
            name: "Erik",
            role: "admin",
            provider: "google",
            hd: "mycelis.link",
            iat: now,
            exp: now + 60,
        };

        const headers = await createForwardedWebIdentityHeaders(session, "forward-secret");

        expect(headers["x-mycelis-web-identity"]).toMatch(/^[A-Za-z0-9_-]+$/);
        expect(headers["x-mycelis-web-identity-signature"]).toMatch(/^[A-Za-z0-9_-]+$/);
        expect(headers["x-mycelis-web-identity-signature"]).not.toBe(headers["x-mycelis-web-identity"]);
    });

    it("round-trips Google OAuth state cookie values without delimiter fragility", () => {
        const encoded = encodeOAuthStateCookie("state-123", "/api/v1/workspace/files/view?path=generated%2Fmedia%2Fproof%2Fimage.png");
        expect(decodeOAuthStateCookie(encoded)).toEqual({
            state: "state-123",
            next: "/api/v1/workspace/files/view?path=generated%2Fmedia%2Fproof%2Fimage.png",
        });
        expect(decodeOAuthStateCookie("state-123:/dashboard")).toEqual({ state: "", next: "/dashboard" });
    });

    it.each(["http://[::]:3000", "http://0.0.0.0:3000", "not-a-url", "ftp://example.com"])("keeps auth redirects off invalid browser origin %s", (origin) => {
        const previous = process.env.MYCELIS_PUBLIC_ORIGIN;
        process.env.MYCELIS_PUBLIC_ORIGIN = "http://127.0.0.1:3000";
        expect(webAuthRedirectURL("/login", origin).toString()).toBe("http://127.0.0.1:3000/login");
        if (previous === undefined) delete process.env.MYCELIS_PUBLIC_ORIGIN;
        else process.env.MYCELIS_PUBLIC_ORIGIN = previous;
    });

    it("uses the managed-server origin when no public origin is configured", () => {
        const previous = process.env.MYCELIS_PUBLIC_ORIGIN;
        process.env.MYCELIS_PUBLIC_ORIGIN = "";
        expect(webAuthRedirectURL("/dashboard", "http://127.0.0.1:3100").toString()).toBe("http://127.0.0.1:3100/dashboard");
        if (previous === undefined) delete process.env.MYCELIS_PUBLIC_ORIGIN;
        else process.env.MYCELIS_PUBLIC_ORIGIN = previous;
    });
    it.each(["https://attacker.example", "http://127.0.0.1:3100"])("pins configured public origin instead of request origin %s", (origin) => {
        const previous = process.env.MYCELIS_PUBLIC_ORIGIN;
        process.env.MYCELIS_PUBLIC_ORIGIN = "https://mycelis.example";
        try {
            expect(webAuthRedirectURL("/login", origin).toString()).toBe("https://mycelis.example/login");
        } finally {
            if (previous === undefined) delete process.env.MYCELIS_PUBLIC_ORIGIN;
            else process.env.MYCELIS_PUBLIC_ORIGIN = previous;
        }
    });

    it.each([
        ["tampered last character", (token: string) => token.slice(0, -1) + (token.endsWith("A") ? "B" : "A")],
        ["short signature", (token: string) => token.slice(0, token.indexOf(".") + 5)],
        ["long signature", (token: string) => `${token}AAAA`],
        ["payload only", (token: string) => `${token.split(".")[0]}.`],
        ["signature only", (token: string) => `.${token.split(".")[1]}`],
        ["extra segment", (token: string) => `${token}.extra`],
        ["non base64 signature", (token: string) => `${token.split(".")[0]}.!!!!`],
    ])("rejects a forged session token (%s) without throwing", async (_label, forge) => {
        const now = Math.floor(Date.now() / 1000);
        const token = await createSessionToken({ sub: "u", email: "u@example.com", name: "U", role: "admin", provider: "local", iat: now, exp: now + 60 }, SESSION_SECRET);
        await expect(verifySessionToken(forge(token), SESSION_SECRET)).resolves.toBeNull();
    });
});

const API_KEY = "api-key-0123456789abcdef0123456789abcdef";
const SESSION_SECRET = "session-secret-0123456789abcdef01234567";
const FORWARD_SECRET = "forward-secret-0123456789abcdef01234567";
const AUTH_ENV = ["MYCELIS_API_KEY", "MYCELIS_BREAK_GLASS_API_KEY", "MYCELIS_WEB_SESSION_SECRET", "MYCELIS_WEB_IDENTITY_FORWARD_SECRET", "MYCELIS_LOCAL_ADMIN_PASSWORD", "MYCELIS_LOCAL_ADMIN_PASSWORD_SHA256"];

describe("webAuth secret separation", () => {
    const saved = Object.fromEntries(AUTH_ENV.map((key) => [key, process.env[key]]));
    const setEnv = (values: Partial<Record<string, string>>) => {
        for (const key of AUTH_ENV) process.env[key] = values[key] ?? "";
    };
    afterEach(() => {
        for (const key of AUTH_ENV) {
            if (saved[key] === undefined) delete process.env[key];
            else process.env[key] = saved[key];
        }
    });

    it("never falls back to MYCELIS_API_KEY for the session secret", async () => {
        setEnv({ MYCELIS_API_KEY: API_KEY, MYCELIS_WEB_IDENTITY_FORWARD_SECRET: FORWARD_SECRET, MYCELIS_LOCAL_ADMIN_PASSWORD: "pw" });
        const config = getWebAuthConfig();
        expect(config.sessionSecret).toBe("");
        expect(config.setupIssues.join(" ")).toContain("MYCELIS_WEB_SESSION_SECRET");
        const now = Math.floor(Date.now() / 1000);
        const forged = await createSessionToken({ sub: "x", email: "x@example.com", name: "X", role: "admin", provider: "local", iat: now, exp: now + 60 }, API_KEY);
        await expect(verifySessionToken(forged, config.sessionSecret)).resolves.toBeNull();
    });

    it("never falls back to MYCELIS_API_KEY for the local password", () => {
        setEnv({ MYCELIS_API_KEY: API_KEY, MYCELIS_WEB_SESSION_SECRET: SESSION_SECRET, MYCELIS_WEB_IDENTITY_FORWARD_SECRET: FORWARD_SECRET });
        const config = getWebAuthConfig();
        expect(config.localPassword).toBe("");
        expect(config.localPasswordSha256).toBe("");
        expect(config.setupIssues.join(" ")).toContain("MYCELIS_LOCAL_ADMIN_PASSWORD_SHA256");
    });

    it.each([
        ["session secret equals API key", { MYCELIS_WEB_SESSION_SECRET: API_KEY }, "MYCELIS_WEB_SESSION_SECRET"],
        ["session secret too short", { MYCELIS_WEB_SESSION_SECRET: "short" }, "MYCELIS_WEB_SESSION_SECRET"],
        ["forward secret equals session secret", { MYCELIS_WEB_IDENTITY_FORWARD_SECRET: SESSION_SECRET }, "MYCELIS_WEB_IDENTITY_FORWARD_SECRET"],
        ["forward secret unset", { MYCELIS_WEB_IDENTITY_FORWARD_SECRET: "" }, "MYCELIS_WEB_IDENTITY_FORWARD_SECRET"],
    ])("fails closed when %s", (_label, override, variable) => {
        setEnv({ MYCELIS_API_KEY: API_KEY, MYCELIS_WEB_SESSION_SECRET: SESSION_SECRET, MYCELIS_WEB_IDENTITY_FORWARD_SECRET: FORWARD_SECRET, MYCELIS_LOCAL_ADMIN_PASSWORD: "pw", ...override });
        const config = getWebAuthConfig();
        expect(config.sessionSecret).toBe("");
        expect(config.setupIssues.join(" ")).toContain(variable);
        expect(config.setupIssues.join(" ")).not.toContain(API_KEY);
    });

    it("rejects a local password that reuses the API key", () => {
        setEnv({ MYCELIS_API_KEY: API_KEY, MYCELIS_WEB_SESSION_SECRET: SESSION_SECRET, MYCELIS_WEB_IDENTITY_FORWARD_SECRET: FORWARD_SECRET, MYCELIS_LOCAL_ADMIN_PASSWORD: API_KEY });
        const config = getWebAuthConfig();
        expect(config.localPassword).toBe("");
        expect(config.setupIssues.join(" ")).toContain("MYCELIS_LOCAL_ADMIN_PASSWORD");
    });

    it("accepts distinct explicit secrets", () => {
        setEnv({ MYCELIS_API_KEY: API_KEY, MYCELIS_WEB_SESSION_SECRET: SESSION_SECRET, MYCELIS_WEB_IDENTITY_FORWARD_SECRET: FORWARD_SECRET, MYCELIS_LOCAL_ADMIN_PASSWORD: "pw" });
        const config = getWebAuthConfig();
        expect(config.sessionSecret).toBe(SESSION_SECRET);
        expect(config.forwardSecret).toBe(FORWARD_SECRET);
        expect(config.setupIssues).toEqual([]);
    });
});

describe("safeNextPath", () => {
    it.each(["", "/\\evil.example", "/%5Cevil.example", "/%255Cevil.example", "/\t/evil.example", "/%09/evil.example", "/\r\n/x", "//evil.example", "/%2F%2Fevil.example", "/%2f/evil.example", "https://evil.example", "javascript:alert(1)", "dashboard", "/.//evil.example", "/%2e//evil.example", "/a/..//evil.example", "/./%2F/evil.example"])("rejects %j", (value) => {
        expect(safeNextPath(value)).toBe("");
        expect(decodeOAuthStateCookie(encodeOAuthStateCookie("s", value)).next).toBe("/dashboard");
    });

    it("keeps same-origin paths with query and fragment", () => {
        expect(safeNextPath("/dashboard?x=1#y")).toBe("/dashboard?x=1#y");
        expect(safeNextPath(null)).toBe("");
        expect(safeNextPath(undefined)).toBe("");
    });
});
