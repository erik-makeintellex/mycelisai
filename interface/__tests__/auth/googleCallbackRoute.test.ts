import { NextRequest } from "next/server";
import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";
import { GET } from "@/app/api/auth/google/callback/route";
import { WEB_SESSION_COOKIE, encodeOAuthStateCookie, verifySessionToken } from "@/lib/webAuth";

const AUTH_ENV = [
  "MYCELIS_WEB_SESSION_SECRET",
  "MYCELIS_API_KEY",
  "MYCELIS_AUTH_GOOGLE_CLIENT_ID",
  "MYCELIS_AUTH_GOOGLE_CLIENT_SECRET",
  "MYCELIS_AUTH_GOOGLE_REDIRECT_URI",
  "MYCELIS_AUTH_GOOGLE_HOSTED_DOMAIN",
  "MYCELIS_AUTH_ALLOWED_DOMAINS",
  "MYCELIS_AUTH_ADMIN_EMAILS",
] as const;

const previousEnv = new Map<string, string | undefined>();
const fetchMock = vi.fn<typeof fetch>();

describe("Google auth callback route", () => {
  beforeEach(() => {
    for (const key of AUTH_ENV) previousEnv.set(key, process.env[key]);
    process.env.MYCELIS_WEB_SESSION_SECRET = "test-session-secret";
    process.env.MYCELIS_API_KEY = "";
    process.env.MYCELIS_AUTH_GOOGLE_CLIENT_ID = "test-google-client";
    process.env.MYCELIS_AUTH_GOOGLE_CLIENT_SECRET = "test-google-secret";
    process.env.MYCELIS_AUTH_GOOGLE_REDIRECT_URI = "http://127.0.0.1:3000/auth/google/callback";
    process.env.MYCELIS_AUTH_GOOGLE_HOSTED_DOMAIN = "makeintellex.com";
    process.env.MYCELIS_AUTH_ALLOWED_DOMAINS = "makeintellex.com";
    process.env.MYCELIS_AUTH_ADMIN_EMAILS = "erik@makeintellex.com";
    vi.stubGlobal("fetch", fetchMock);
  });

  afterEach(() => {
    vi.restoreAllMocks();
    vi.unstubAllGlobals();
    fetchMock.mockReset();
    for (const key of AUTH_ENV) {
      const previous = previousEnv.get(key);
      if (previous === undefined) delete process.env[key];
      else process.env[key] = previous;
    }
    previousEnv.clear();
  });

  it("accepts the allowed Workspace domain and creates a signed admin session", async () => {
    mockGoogleIdentity({
      sub: "google-123",
      email: "erik@makeintellex.com",
      email_verified: true,
      name: "Erik",
      hd: "makeintellex.com",
      aud: "test-google-client",
    });

    const response = await GET(callbackRequest("/dashboard"));

    expect(response.status).toBe(307);
    expect(redirectTarget(response)).toBe("/dashboard");
    const sessionCookie = response.cookies.get(WEB_SESSION_COOKIE);
    expect(sessionCookie?.httpOnly).toBe(true);
    expect(sessionCookie?.sameSite).toBe("lax");
    await expect(verifySessionToken(sessionCookie?.value, "test-session-secret")).resolves.toMatchObject({
      sub: "google-123",
      email: "erik@makeintellex.com",
      role: "admin",
      provider: "google",
      hd: "makeintellex.com",
    });
    expect(response.cookies.get("mycelis_google_state")?.value).toBe("");
  });

  it.each([
    ["personal Gmail", { email: "person@gmail.com", hd: undefined }],
    ["another hosted domain", { email: "person@example.com", hd: "example.com" }],
  ])("rejects %s identities", async (_label, identity) => {
    mockGoogleIdentity({ ...identity, email_verified: true, aud: "test-google-client" });

    const response = await GET(callbackRequest());

    expect(redirectTarget(response)).toBe("/login?error=domain");
    expect(response.cookies.get(WEB_SESSION_COOKIE)).toBeUndefined();
    expect(response.cookies.get("mycelis_google_state")?.value).toBe("");
  });

  it("rejects an identity token issued for another OAuth client", async () => {
    mockGoogleIdentity({ email: "erik@makeintellex.com", email_verified: true, hd: "makeintellex.com", aud: "other-client" });
    const response = await GET(callbackRequest());
    expect(redirectTarget(response)).toBe("/login?error=google_identity");
  });

  it.each([false, "false", undefined])("rejects an identity whose verified-email claim is %s", async (emailVerified) => {
    mockGoogleIdentity({ email: "erik@makeintellex.com", email_verified: emailVerified, hd: "makeintellex.com", aud: "test-google-client" });
    const response = await GET(callbackRequest());
    expect(redirectTarget(response)).toBe("/login?error=google_identity");
  });

  it("rejects mismatched OAuth state, clears state, and never contacts Google", async () => {
    const response = await GET(callbackRequest("/dashboard", "wrong-state"));
    expect(redirectTarget(response)).toBe("/login?error=google_state");
    expect(response.cookies.get("mycelis_google_state")?.value).toBe("");
    expect(fetchMock).not.toHaveBeenCalled();
  });

  it("returns a normalized provider failure and logs only phase and status", async () => {
    const warn = vi.spyOn(console, "warn").mockImplementation(() => undefined);
    fetchMock.mockResolvedValueOnce(new Response(JSON.stringify({
      error: "invalid_client",
      error_description: "provider detail containing test-google-secret",
    }), { status: 401, headers: { "content-type": "application/json" } }));

    const response = await GET(callbackRequest());

    expect(redirectTarget(response)).toBe("/login?error=google_token");
    expect(warn).toHaveBeenCalledWith("[auth/google] callback failed", { phase: "token_exchange", status: 401 });
    expect(JSON.stringify(warn.mock.calls)).not.toContain("invalid_client");
    expect(JSON.stringify(warn.mock.calls)).not.toContain("test-google-secret");
  });

  it("does not log exception messages that may contain request credentials", async () => {
    const warn = vi.spyOn(console, "warn").mockImplementation(() => undefined);
    fetchMock.mockRejectedValueOnce(new TypeError("request contained test-google-secret"));

    const response = await GET(callbackRequest());

    expect(redirectTarget(response)).toBe("/login?error=google_exception");
    expect(warn).toHaveBeenCalledWith("[auth/google] callback failed", { phase: "exception", errorType: "TypeError" });
    expect(JSON.stringify(warn.mock.calls)).not.toContain("test-google-secret");
  });
});

function callbackRequest(next = "/dashboard", state = "state-123") {
  const saved = encodeOAuthStateCookie("state-123", next);
  return new NextRequest(`http://127.0.0.1:3000/auth/google/callback?code=auth-code&state=${encodeURIComponent(state)}`, {
    headers: { cookie: `mycelis_google_state=${saved}` },
  });
}

function redirectTarget(response: Response) {
  const location = new URL(response.headers.get("location") ?? "http://invalid");
  return `${location.pathname}${location.search}`;
}

function mockGoogleIdentity(identity: Record<string, unknown>) {
  fetchMock
    .mockResolvedValueOnce(new Response(JSON.stringify({ id_token: "test-id-token" }), {
      status: 200,
      headers: { "content-type": "application/json" },
    }))
    .mockResolvedValueOnce(new Response(JSON.stringify(identity), {
      status: 200,
      headers: { "content-type": "application/json" },
    }));
}
