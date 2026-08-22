import type { Page } from "@playwright/test";
import type { APIEnvelope, GroupRecord } from "./finalization-browser-package";
import { liveAPIHeaders, liveAPIURL } from "./live-api-auth";

type FixtureResourceKind =
  | "organization"
  | "group"
  | "team"
  | "run"
  | "outcome"
  | "artifact"
  | "workspace_path";

export type QAFixtureScope = {
  id: string;
  ownerRef: string;
  executionRef: string;
};

export type QAFixtureResource = {
  kind: FixtureResourceKind;
  ref: string;
};

export async function createQAFixtureScope(
  page: Page,
  executionRef: string,
): Promise<QAFixtureScope> {
  const ownerRef = "playwright";
  const response = await page.request.post(liveAPIURL("/api/v1/testing/fixture-scopes"), {
    headers: jsonHeaders(),
    data: { owner_ref: ownerRef, execution_ref: executionRef, ttl_seconds: 86_400 },
  });
  if (!response.ok()) {
    throw new Error(`Create QA fixture scope failed (${response.status()}): ${await response.text()}`);
  }
  const body = await response.json() as APIEnvelope<{ id: string }>;
  if (!body.data?.id) throw new Error("Create QA fixture scope returned no scope ID");
	await page.setExtraHTTPHeaders({ "X-Mycelis-QA-Fixture-Scope": body.data.id });
  return { id: body.data.id, ownerRef, executionRef };
}

export async function registerQAFixtureResources(
  page: Page,
  scope: QAFixtureScope,
  resources: QAFixtureResource[],
) {
  if (resources.length === 0) return;
  const response = await page.request.post(
    liveAPIURL(`/api/v1/testing/fixture-scopes/${encodeURIComponent(scope.id)}/resources`),
    {
      headers: jsonHeaders(),
      data: {
        owner_ref: scope.ownerRef,
        execution_ref: scope.executionRef,
        resources,
      },
    },
  );
  if (!response.ok()) {
    throw new Error(`Register QA fixture resources failed (${response.status()}): ${await response.text()}`);
  }
}

export async function purgeDeliveryFixture(
  page: Page,
  scope: QAFixtureScope,
  options: { teamID?: string; organizationID?: string; runID?: string } = {},
) {
  const resources: QAFixtureResource[] = [];
  const runIDs = new Set<string>();
  let discoveryError: unknown;
  if (options.runID) runIDs.add(options.runID);
  if (options.organizationID) {
    resources.push({ kind: "organization", ref: options.organizationID });
  }
  if (options.teamID) {
    try {
      const teamID = options.teamID;
      let teamExists = false;
      const response = await page.request.get(liveAPIURL("/api/v1/groups"), { headers: liveAPIHeaders() });
      if (!response.ok()) {
        throw new Error(`Discover QA fixture groups failed (${response.status()}): ${await response.text()}`);
      }
      const groups = ((await response.json()) as APIEnvelope<GroupRecord[]>).data ?? [];
      const group = groups.find((candidate) => candidate.team_ids?.includes(teamID));
      if (group) {
        teamExists = true;
        resources.push({ kind: "group", ref: group.group_id });
        if (group.workspace_folder) {
          resources.push({ kind: "workspace_path", ref: group.workspace_folder });
        }
      }
      const workResponse = await page.request.get(
        liveAPIURL(`/api/v1/teams/${encodeURIComponent(teamID)}/work?limit=100`),
        { headers: liveAPIHeaders() },
      );
      if (!workResponse.ok()) {
        throw new Error(`Discover QA fixture team work failed (${workResponse.status()}): ${await workResponse.text()}`);
      }
      const work = ((await workResponse.json()) as APIEnvelope<Array<{ run_id?: string }>>).data ?? [];
      const discoveredRunIDs = work.map((item) => item.run_id).filter((runID): runID is string => Boolean(runID));
      for (const runID of discoveredRunIDs) runIDs.add(runID);
      for (const runID of runIDs) resources.push({ kind: "run", ref: runID });
      if (teamExists || discoveredRunIDs.length > 0) resources.push({ kind: "team", ref: teamID });
    } catch (error) {
      discoveryError = error;
    }
  } else {
    for (const runID of runIDs) resources.push({ kind: "run", ref: runID });
  }
  let cleanupError: unknown;
  try {
    if (resources.length > 0) {
      await registerQAFixtureResources(page, scope, resources);
    }
  } catch (error) {
    cleanupError = error;
  }
  try {
    const response = await page.request.post(
      liveAPIURL(`/api/v1/testing/fixture-scopes/${encodeURIComponent(scope.id)}/purge`),
      {
        headers: jsonHeaders(),
        data: {
          owner_ref: scope.ownerRef,
          execution_ref: scope.executionRef,
          confirm: true,
        },
      },
    );
    if (!response.ok()) {
      throw new Error(`Purge QA fixture scope failed (${response.status()}): ${await response.text()}`);
    }
  } catch (error) {
    cleanupError = cleanupError ?? error;
  } finally {
    await page.setExtraHTTPHeaders({});
  }
  if (cleanupError) throw cleanupError;
  if (discoveryError) throw discoveryError;
}

function jsonHeaders() {
  return { ...(liveAPIHeaders() ?? {}), "Content-Type": "application/json" };
}
