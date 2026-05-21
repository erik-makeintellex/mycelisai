import { describe, expect, it } from "vitest";
import { createForwardedWebIdentityHeaders, createSessionToken, roleForEmail, sha256Hex, splitList, verifySessionToken, webAuthRedirectURL, type WebSession } from "@/lib/webAuth";

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

    it("keeps auth redirects on the public origin instead of the bind host", () => {
        const previous = process.env.MYCELIS_PUBLIC_ORIGIN;
        process.env.MYCELIS_PUBLIC_ORIGIN = "http://127.0.0.1:3000";
        expect(webAuthRedirectURL("/login", "http://[::]:3000").toString()).toBe("http://127.0.0.1:3000/login");
        if (previous === undefined) delete process.env.MYCELIS_PUBLIC_ORIGIN;
        else process.env.MYCELIS_PUBLIC_ORIGIN = previous;
    });

    it("preserves valid managed-server origins for browser proof", () => {
        const previous = process.env.MYCELIS_PUBLIC_ORIGIN;
        process.env.MYCELIS_PUBLIC_ORIGIN = "http://127.0.0.1:3000";
        expect(webAuthRedirectURL("/dashboard", "http://127.0.0.1:3100").toString()).toBe("http://127.0.0.1:3100/dashboard");
        if (previous === undefined) delete process.env.MYCELIS_PUBLIC_ORIGIN;
        else process.env.MYCELIS_PUBLIC_ORIGIN = previous;
    });
});
