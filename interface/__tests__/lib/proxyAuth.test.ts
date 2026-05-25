import { describe, expect, it } from "vitest";
import { shouldRedirectUnauthenticatedApiRequest } from "@/proxy";

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
