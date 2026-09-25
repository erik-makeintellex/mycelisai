import { NextRequest } from "next/server";
import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";
import { GET } from "@/app/api/auth/google/callback/route";
import { WEB_SESSION_COOKIE, encodeOAuthStateCookie, verifySessionToken } from "@/lib/webAuth";

const AUTH_ENV = [
  "MYCELIS_PUBLIC_ORIGIN",
  "MYCELIS_WEB_SESSION_SECRET",
  "MYCELIS_WEB_IDENTITY_FORWARD_SECRET",
  "MYCELIS_API_KEY",
  "MYCELIS_AUTH_GOOGLE_CLIENT_ID",
  "MYCELIS_AUTH_GOOGLE_CLIENT_SECRET",
  "MYCELIS_AUTH_GOOGLE_REDIRECT_URI",
  "MYCELIS_AUTH_GOOGLE_HOSTED_DOMAIN",
  "MYCELIS_AUTH_ALLOWED_DOMAINS",
  "MYCELIS_AUTH_ADMIN_EMAILS",
] as const;

const OPEN_REDIRECT_NEXTS = ["/\\evil.example", "/%5Cevil.example", "/\t/evil.example", "//evil.example", "/%2F%2Fevil.example", "https://evil.example", "javascript:alert(1)"];

const previousEnv = new Map<string, string | undefined>();
const fetchMock = vi.fn<typeof fetch>();

describe("Google auth callback route", () => {
  beforeEach(() => {
    for (const key of AUTH_ENV) previousEnv.set(key, process.env[key]);
    process.env.MYCELIS_PUBLIC_ORIGIN = "http://127.0.0.1:3000";
    process.env.MYCELIS_WEB_SESSION_SECRET = "test-session-secret-0123456789abcdef0123";
    process.env.MYCELIS_WEB_IDENTITY_FORWARD_SECRET = "test-forward-secret-0123456789abcdef0123";
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
    await expect(verifySessionToken(sessionCookie?.value, "test-session-secret-0123456789abcdef0123")).resolves.toMatchObject({
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

  it.each(OPEN_REDIRECT_NEXTS)("never redirects off-origin after sign-in for next=%j", async (next) => {
    mockGoogleIdentity({ sub: "google-123", email: "erik@makeintellex.com", email_verified: true, name: "Erik", hd: "makeintellex.com", aud: "test-google-client" });
    const response = await GET(callbackRequest(next));
    const location = new URL(response.headers.get("location") ?? "http://invalid");
    expect(location.origin).toBe("http://127.0.0.1:3000");
    // A value that cannot survive the state cookie round-trip fails closed on state.
    expect(["/dashboard", "/login?error=google_state"]).toContain(redirectTarget(response));
  });

  it("keeps a same-origin next path with query and fragment", async () => {
    mockGoogleIdentity({ sub: "google-123", email: "erik@makeintellex.com", email_verified: true, name: "Erik", hd: "makeintellex.com", aud: "test-google-client" });
    const response = await GET(callbackRequest("/dashboard?x=1#y"));
    expect(response.headers.get("location")).toBe("http://127.0.0.1:3000/dashboard?x=1#y");
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

  it.each(["http://0.0.0.0:3000", "https://attacker.example"])("keeps token failures on the configured public origin instead of %s", async (origin) => {
    vi.spyOn(console, "warn").mockImplementation(() => undefined);
    fetchMock.mockResolvedValueOnce(new Response("", { status: 401 }));
    const request = callbackRequest();
    // Next's test URL normalizer maps loopback names to localhost; model the
    // origin actually observed behind the deployed Compose listener explicitly.
    Object.defineProperty(request, "nextUrl", {
      value: new URL(new URL(request.url).pathname + new URL(request.url).search, origin),
    });
    const response = await GET(request);
    expect(response.headers.get("location")).toBe("http://127.0.0.1:3000/login?error=google_token");
    expect(response.cookies.get(WEB_SESSION_COOKIE)).toBeUndefined();
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
