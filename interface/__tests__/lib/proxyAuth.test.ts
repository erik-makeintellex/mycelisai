import { NextRequest } from "next/server";
import { afterEach, beforeEach, describe, expect, it } from "vitest";
import { createSessionToken } from "@/lib/webAuth";
import { proxy, shouldRedirectUnauthenticatedApiRequest } from "@/proxy";

describe("proxy unauthenticated API handling", () => {
  it("redirects browser navigations to login instead of showing raw API auth JSON", () => {
    expect(shouldRedirectUnauthenticatedApiRequest(new Headers({
      accept: "text/html,application/xhtml+xml",
      "sec-fetch-mode": "navigate",
      "sec-fetch-dest": "document",
    }))).toBe(true);
  });

  it("keeps programmatic fetches on structured 401 JSON", () => {
    expect(shouldRedirectUnauthenticatedApiRequest(new Headers({
      accept: "application/json",
      "sec-fetch-mode": "cors",
      "sec-fetch-dest": "empty",
    }))).toBe(false);
  });
});

const SESSION_SECRET = "session-secret-0123456789abcdef01234567";
const FORWARD_SECRET = "forward-secret-0123456789abcdef01234567";
const AUTH_ENV = ["MYCELIS_API_KEY", "MYCELIS_BREAK_GLASS_API_KEY", "MYCELIS_WEB_SESSION_SECRET", "MYCELIS_WEB_IDENTITY_FORWARD_SECRET"];
const FORGED = {
  authorization: "Bearer browser-supplied",
  "x-mycelis-web-identity": "forged-payload",
  "x-mycelis-web-identity-signature": "forged-signature",
  "x-mycelis-qa-fixture-scope": "11111111-1111-1111-1111-111111111111",
};

async function sessionCookie(role: "admin" | "standard" = "standard") {
  const now = Math.floor(Date.now() / 1000);
  const token = await createSessionToken({ sub: "u", email: "u@example.com", name: "U", role, provider: "local", iat: now, exp: now + 600 }, SESSION_SECRET);
  return `mycelis_web_session=${token}`;
}

function forwarded(response: Response, name: string): string | null {
  return response.headers.get(`x-middleware-request-${name}`);
}

describe("proxy header hygiene", () => {
  const saved = new Map<string, string | undefined>();
  beforeEach(() => {
    for (const key of AUTH_ENV) saved.set(key, process.env[key]);
    process.env.MYCELIS_API_KEY = "api-key-0123456789abcdef0123456789abcdef";
    process.env.MYCELIS_BREAK_GLASS_API_KEY = "";
    process.env.MYCELIS_WEB_SESSION_SECRET = SESSION_SECRET;
    process.env.MYCELIS_WEB_IDENTITY_FORWARD_SECRET = FORWARD_SECRET;
  });
  afterEach(() => {
    for (const key of AUTH_ENV) {
      const previous = saved.get(key);
      if (previous === undefined) delete process.env[key];
      else process.env[key] = previous;
    }
  });

  it("returns 401 for browser-forged identity headers without a session", async () => {
    const response = await proxy(new NextRequest("http://127.0.0.1:3000/api/v1/audit", { headers: FORGED }));
    expect(response.status).toBe(401);
  });

  it("replaces browser identity headers with its own and drops the QA fixture header", async () => {
    const response = await proxy(new NextRequest("http://127.0.0.1:3000/api/v1/audit", {
      headers: { ...FORGED, cookie: await sessionCookie() },
    }));
    expect(forwarded(response, "authorization")).toBe("Bearer api-key-0123456789abcdef0123456789abcdef");
    expect(forwarded(response, "x-mycelis-web-identity")).not.toBe("forged-payload");
    expect(forwarded(response, "x-mycelis-web-identity")).toMatch(/^[A-Za-z0-9_-]+$/);
    expect(forwarded(response, "x-mycelis-web-identity-signature")).not.toBe("forged-signature");
    expect(forwarded(response, "x-mycelis-qa-fixture-scope")).toBeNull();
  });

  it("strips browser authority headers even when no API key is configured", async () => {
    process.env.MYCELIS_API_KEY = "";
    const response = await proxy(new NextRequest("http://127.0.0.1:3000/api/v1/audit", {
      headers: { ...FORGED, cookie: await sessionCookie() },
    }));
    for (const name of Object.keys(FORGED)) {
      expect(forwarded(response, name)).toBeNull();
    }
    const override = response.headers.get("x-middleware-override-headers");
    expect(override).not.toBeNull();
    for (const name of Object.keys(FORGED)) {
      expect((override ?? "").split(",")).not.toContain(name);
    }
  });
});
