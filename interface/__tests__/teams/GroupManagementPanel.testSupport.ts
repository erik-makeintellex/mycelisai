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
      if (url === "/api/v1/groups/lifecycle")
        return jsonResponse({ ok: true, data: emptyLifecycleReport([]) });
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
  teamWork = {},
  workflowLogs = {},
  monitor = { status: "online", published_count: 0 },
  lifecycle,
}: {
  groups: TestGroup[];
  outputs?: Record<string, unknown[]>;
  teamWork?: Record<string, unknown[]>;
  workflowLogs?: Record<string, unknown>;
  monitor?: Record<string, unknown>;
  lifecycle?: Record<string, unknown>;
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
      if (url === "/api/v1/groups/lifecycle")
        return jsonResponse({
          ok: true,
          data: lifecycle ?? emptyLifecycleReport(groups),
        });
      if (
        url === "/api/v1/groups/lifecycle/archive-expired" &&
        init?.method === "POST"
      )
        return jsonResponse({
          ok: true,
          data: {
            archived_count: 0,
            archived_group_ids: [],
            report: emptyLifecycleReport(groups),
          },
        });
      if (
        url.startsWith("/api/v1/workspace/files/reveal") &&
        init?.method === "POST"
      )
        return jsonResponse({
          ok: true,
          data: { workspace_path: "workspace/generated/coin-runner" },
        });
      const clearMatch = url.match(/^\/api\/v1\/groups\/([^/]+)\/clear$/);
      if (clearMatch && init?.method === "POST") {
        const groupId = decodeURIComponent(clearMatch[1]);
        const index = groups.findIndex(
          (group) => group.group_id === groupId,
        );
        if (index < 0) return jsonResponse({ error: "not found" }, false, 404);
        groups[index] = { ...groups[index], status: "archived" };
        const body = init.body ? JSON.parse(String(init.body)) : {};
        const includeOutputs = body.include_outputs === true;
        return jsonResponse({
          ok: true,
          data: {
            group: groups[index],
            outputs_cleared: includeOutputs,
            workspace_removed: includeOutputs,
            artifacts_archived: includeOutputs ? 1 : 0,
            operator_description: includeOutputs
              ? "Group cleared from active lanes and retained output files were removed from the group workspace."
              : "Group cleared from active lanes. Message-bus handoff data is transient; retained output files were kept.",
          },
        });
      }
      const match = url.match(
        /^\/api\/v1\/groups\/([^/]+)\/outputs\?limit=8(?:&include_internal=true)?$/,
      );
      if (match) {
        const groupOutputs = outputs[decodeURIComponent(match[1])] ?? [];
        const includeInternal = url.includes("include_internal=true");
        return jsonResponse({
          ok: true,
          data: includeInternal
            ? groupOutputs
            : groupOutputs.filter(isUserDeliverableTestArtifact),
        });
      }
      const workflowLogMatch = url.match(
        /^\/api\/v1\/groups\/([^/]+)\/workflow-log\?limit=50&include_outputs=true&include_audit=true$/,
      );
      if (workflowLogMatch) {
        const groupId = decodeURIComponent(workflowLogMatch[1]);
        const workflowLog = workflowLogs[groupId];
        if (workflowLog)
          return jsonResponse({
            ok: true,
            data: workflowLog,
          });
        return jsonResponse({ error: "not found" }, false, 404);
      }
      const teamWorkMatch = url.match(
        /^\/api\/v1\/teams\/([^/]+)\/work\?limit=8&include_archived=false$/,
      );
      if (teamWorkMatch)
        return jsonResponse({
          ok: true,
          data: teamWork[decodeURIComponent(teamWorkMatch[1])] ?? [],
        });
      return jsonResponse({ error: "not found" }, false, 404);
    },
  );
}

function isUserDeliverableTestArtifact(item: unknown) {
  if (!item || typeof item !== "object") return true;
  const artifact = item as Record<string, unknown>;
  const metadata =
    artifact.metadata && typeof artifact.metadata === "object"
      ? (artifact.metadata as Record<string, unknown>)
      : {};
  const explicit = String(
    metadata.output_class ?? metadata.delivery_status ?? metadata.visibility ?? "",
  ).toLowerCase();
  if (["planning", "proof", "internal", "internal_handoff"].includes(explicit))
    return false;
  const path = String(artifact.file_path ?? artifact.title ?? "")
    .replaceAll("\\", "/")
    .toLowerCase();
  return !(
    path.includes("/planning/") ||
    path.includes("/proof/") ||
    path.includes("/watch/") ||
    path.endsWith("team_evocation.md") ||
    path.endsWith("research_council_handoff.md")
  );
}

function emptyLifecycleReport(groups: TestGroup[]) {
  return {
    generated_at: new Date().toISOString(),
    summary: {
      total_groups: groups.length,
      active_groups: groups.filter((group) => group.status === "active").length,
      expired_active_groups: 0,
      standing_no_expiry_groups: groups.filter((group) => !group.expiry).length,
      stale_standing_groups: 0,
      review_needed_groups: 0,
      output_ready_idle_groups: 0,
      team_work_needing_attention: 0,
    },
    items: groups.map((group) => ({
      group_id: group.group_id,
      name: group.name,
      status: group.status,
      work_mode: group.work_mode ?? "propose_only",
      kind: group.expiry ? "temporary" : "standing",
      recommendation: "keep_active",
      reason: "Test group is active.",
      expiry: group.expiry ?? null,
      expired: false,
      age_hours: 1,
      team_count: Array.isArray(group.team_ids) ? group.team_ids.length : 0,
      output_count: 0,
      team_work_count: 0,
      active_or_blocked_work_count: 0,
      output_ready_work_count: 0,
      archived_work_count: 0,
    })),
  };
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
