import { mockFetch } from "../setup";

export type TestGroup = Record<string, unknown> & {
  group_id: string;
  name: string;
  status: string;
};

export function jsonResponse(
  payload: unknown,
  ok = true,
  status = 200,
): Response {
  return { ok, status, json: async () => payload } as Response;
}

export function standingGroup(): TestGroup {
  return {
    group_id: "group-standing",
    name: "Standing Ops",
    goal_statement: "Run durable operations",
    work_mode: "propose_only",
    member_user_ids: [],
    team_ids: ["team-ops"],
    coordinator_profile: "ops-lead",
    approval_policy_ref: "",
    status: "active",
    created_by: "admin",
    created_at: new Date().toISOString(),
  };
}

export function tempGroup(overrides: Partial<TestGroup> = {}): TestGroup {
  return {
    group_id: "group-temp",
    name: "Temp Campaign",
    goal_statement: "Produce one campaign package",
    work_mode: "propose_only",
    member_user_ids: [],
    team_ids: ["team-marketing"],
    coordinator_profile: "marketing-lead",
    approval_policy_ref: "",
    status: "active",
    expiry: new Date(Date.now() + 60_000).toISOString(),
    created_by: "admin",
    created_at: new Date().toISOString(),
    ...overrides,
  };
}

export function documentArtifact(overrides: Record<string, unknown> = {}) {
  return {
    id: "artifact-1",
    agent_id: "marketing-lead",
    artifact_type: "document",
    title: "Launch brief",
    content_type: "text/markdown",
    content: "Campaign summary",
    metadata: {},
    status: "approved",
    created_at: new Date().toISOString(),
    ...overrides,
  };
}

export function installApprovalCreateFetch(
  postBodies: Array<Record<string, unknown>> = [],
) {
  mockFetch.mockImplementation(
    async (input: RequestInfo | URL, init?: RequestInit) => {
      const url = urlFromInput(input);
      if (url === "/api/v1/groups" && (!init?.method || init.method === "GET"))
        return jsonResponse({ ok: true, data: [] });
      if (url === "/api/v1/groups/monitor")
        return jsonResponse({
          ok: true,
          data: { status: "online", published_count: 0 },
        });
      if (url === "/api/v1/groups" && init?.method === "POST") {
        postBodies.push(JSON.parse(String(init.body)));
        if (postBodies.length === 1)
          return jsonResponse(
            {
              ok: true,
              data: {
                requires_approval: true,
                confirm_token: { token: "tok-123" },
                intent_proof: { id: "proof-1" },
              },
            },
            true,
            202,
          );
        return jsonResponse(
          { ok: true, data: { group_id: "group-1" } },
          true,
          201,
        );
      }
      return jsonResponse({ error: "not found" }, false, 404);
    },
  );
}

export function installGroupsFetch({
  groups,
  outputs = {},
  monitor = { status: "online", published_count: 0 },
}: {
  groups: TestGroup[];
  outputs?: Record<string, unknown[]>;
  monitor?: Record<string, unknown>;
}) {
  mockFetch.mockImplementation(
    async (input: RequestInfo | URL, init?: RequestInit) => {
      const url = urlFromInput(input);
      if (url === "/api/v1/groups" && (!init?.method || init.method === "GET"))
        return jsonResponse({
          ok: true,
          data: groups.map((group) => ({ ...group })),
        });
      if (url === "/api/v1/groups/monitor")
        return jsonResponse({ ok: true, data: monitor });
      if (
        url === "/api/v1/groups/group-temp/status" &&
        init?.method === "PATCH"
      ) {
        const index = groups.findIndex(
          (group) => group.group_id === "group-temp",
        );
        groups[index] = { ...groups[index], status: "archived" };
        return jsonResponse({ ok: true, data: groups[index] });
      }
      const match = url.match(/^\/api\/v1\/groups\/([^/]+)\/outputs\?limit=8$/);
      if (match)
        return jsonResponse({
          ok: true,
          data: outputs[decodeURIComponent(match[1])] ?? [],
        });
      return jsonResponse({ error: "not found" }, false, 404);
    },
  );
}

function urlFromInput(input: RequestInfo | URL): string {
  if (typeof input === "string") return input;
  if (input instanceof URL) return input.pathname + input.search;
  const requestLike = input as Request;
  if (typeof requestLike.url === "string") {
    try {
      const parsed = new URL(requestLike.url);
      return parsed.pathname + parsed.search;
    } catch {
      return requestLike.url;
    }
  }
  return String(input);
}
