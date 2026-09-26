import { render, screen } from "@testing-library/react";
import { beforeEach, describe, expect, it, vi } from "vitest";
import LoginPage from "@/app/login/page";

const { authConfig } = vi.hoisted(() => ({
    authConfig: {
        sessionSecret: "session-secret",
        forwardSecret: "forward-secret",
        localUsername: "admin",
        localPassword: "",
        localPasswordSha256: "",
        googleClientId: "",
        googleClientSecret: "",
        googleRedirectUri: "",
        googleHostedDomain: "",
        allowedDomains: [] as string[],
        adminEmails: [] as string[],
        setupIssues: [] as string[],
    },
}));

vi.mock("next/headers", () => ({
    cookies: vi.fn(async () => ({ get: vi.fn(() => undefined) })),
}));

vi.mock("next/navigation", () => ({
    redirect: vi.fn(),
}));

vi.mock("@/components/shell/ThemeSync", () => ({
    ThemeSync: () => null,
}));

vi.mock("@/lib/webAuth", async (importOriginal) => ({
    safeNextPath: (await importOriginal<typeof import("@/lib/webAuth")>()).safeNextPath,
    WEB_SESSION_COOKIE: "mycelis_web_session",
    getWebAuthConfig: () => authConfig,
    googleConfigured: () => Boolean(authConfig.googleClientId && authConfig.googleClientSecret && authConfig.googleRedirectUri),
    googleWorkspacePolicy: () => ({
        hostedDomain: authConfig.googleHostedDomain,
        allowedDomains: authConfig.allowedDomains,
        displayDomains: authConfig.allowedDomains.length
            ? authConfig.allowedDomains
            : (authConfig.googleHostedDomain ? [authConfig.googleHostedDomain] : []),
        domainLabel: authConfig.allowedDomains.join(", ") || authConfig.googleHostedDomain,
    }),
    localLoginConfigured: () => Boolean(authConfig.sessionSecret && (authConfig.localPassword || authConfig.localPasswordSha256)),
    verifySessionToken: vi.fn(async () => null),
}));

describe("LoginPage", () => {
    beforeEach(() => {
        authConfig.googleClientId = "";
        authConfig.googleClientSecret = "";
        authConfig.googleRedirectUri = "";
        authConfig.googleHostedDomain = "";
        authConfig.allowedDomains = [];
        authConfig.setupIssues = [];
    });

    it("names the variables to set when local sign-in is not configured", async () => {
        authConfig.setupIssues = ["Set MYCELIS_LOCAL_ADMIN_PASSWORD_SHA256 (preferred) or MYCELIS_LOCAL_ADMIN_PASSWORD to enable local sign-in."];

        render(await LoginPage({ searchParams: Promise.resolve({ error: "config" }) }));

        expect(screen.getByText(/Sign-in setup is incomplete/)).toBeTruthy();
        expect(screen.getByText(/MYCELIS_LOCAL_ADMIN_PASSWORD_SHA256/)).toBeTruthy();
        expect(screen.getByRole("button", { name: /Sign in as local admin/ }).hasAttribute("disabled")).toBe(true);
    });

    it("shows provider-neutral guidance when enterprise SSO is not configured", async () => {
        render(await LoginPage({ searchParams: Promise.resolve({}) }));

        expect(screen.getByText("Enterprise SSO is not configured for this deployment.")).toBeTruthy();
        expect(screen.queryByRole("link", { name: "Sign in with Google Workspace" })).toBeNull();
        expect(screen.queryByText(/Accepted Google account domain/)).toBeNull();
        expect(screen.queryByText(/Personal Gmail accounts are rejected/)).toBeNull();
    });

    it("shows Google sign-in and configured domain guidance when Google is enabled", async () => {
        authConfig.googleClientId = "google-client";
        authConfig.googleClientSecret = "google-secret";
        authConfig.googleRedirectUri = "http://127.0.0.1:3000/auth/google/callback";
        authConfig.googleHostedDomain = "makeintellex.com";
        authConfig.allowedDomains = ["makeintellex.com"];

        render(await LoginPage({ searchParams: Promise.resolve({ next: "/dashboard" }) }));

        expect(screen.getByRole("link", { name: "Sign in with Google Workspace" }).getAttribute("href"))
            .toBe("/auth/google/start?next=%2Fdashboard");
        expect(screen.getByText("makeintellex.com")).toBeTruthy();
        expect(screen.getByText(/Personal Gmail accounts are rejected for this deployment/)).toBeTruthy();
        expect(screen.queryByText("Enterprise SSO is not configured for this deployment.")).toBeNull();
    });
});
