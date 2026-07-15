import { NextRequest } from "next/server";
import { afterEach, describe, expect, it } from "vitest";
import { GET } from "@/app/api/auth/google/start/route";

const AUTH_ENV = [
  "MYCELIS_WEB_SESSION_SECRET",
  "MYCELIS_API_KEY",
  "MYCELIS_AUTH_GOOGLE_CLIENT_ID",
  "MYCELIS_AUTH_GOOGLE_CLIENT_SECRET",
  "MYCELIS_AUTH_GOOGLE_REDIRECT_URI",
  "MYCELIS_AUTH_GOOGLE_HOSTED_DOMAIN",
  "MYCELIS_AUTH_ALLOWED_DOMAINS",
] as const;

const previousEnv = new Map<string, string | undefined>();

describe("Google auth start route", () => {
  afterEach(() => {
    for (const key of AUTH_ENV) {
      const previous = previousEnv.get(key);
      if (previous === undefined) delete process.env[key];
      else process.env[key] = previous;
    }
    previousEnv.clear();
  });

  it("canonicalizes API re-export requests back to the public SSO route", async () => {
    configureGoogleAuth();

    const response = await GET(new NextRequest("http://[::]:3000/api/auth/google/start?next=%2Fdashboard", {
      headers: { host: "[::]:3000" },
    }));

    expect(response.status).toBe(307);
    expect(response.headers.get("location")).toBe("http://127.0.0.1:3000/auth/google/start?next=%2Fdashboard");
  });

  it("uses the shared Google Workspace policy for the OAuth hosted-domain hint", async () => {
    configureGoogleAuth();

    const response = await GET(new NextRequest("http://127.0.0.1:3000/auth/google/start?next=%2Fdashboard", {
      headers: { host: "127.0.0.1:3000" },
    }));

    const location = response.headers.get("location") ?? "";
    expect(response.status).toBe(307);
    expect(location).toContain("https://accounts.google.com/o/oauth2/v2/auth");
    expect(new URL(location).searchParams.get("hd")).toBe("mycelis.link");
  });
});

function configureGoogleAuth() {
  for (const key of AUTH_ENV) previousEnv.set(key, process.env[key]);
  process.env.MYCELIS_WEB_SESSION_SECRET = "session-secret";
  process.env.MYCELIS_API_KEY = "";
  process.env.MYCELIS_AUTH_GOOGLE_CLIENT_ID = "google-client";
  process.env.MYCELIS_AUTH_GOOGLE_CLIENT_SECRET = "google-secret";
  process.env.MYCELIS_AUTH_GOOGLE_REDIRECT_URI = "http://127.0.0.1:3000/auth/google/callback";
  process.env.MYCELIS_AUTH_GOOGLE_HOSTED_DOMAIN = "mycelis.link";
  process.env.MYCELIS_AUTH_ALLOWED_DOMAINS = "mycelis.link";
}
