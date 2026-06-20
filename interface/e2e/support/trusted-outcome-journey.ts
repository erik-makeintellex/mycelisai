import type { Page, Route } from "@playwright/test";
import { fulfillJSON, type ArtifactRecord, type GroupRecord } from "./finalization-browser-package";
import { mockOrganizationWorkspace, type RouteResponse } from "./soma-ui-testing";

type BrowserFetchOptions = { method?: string; headers?: Record<string, string>; body?: string };

export const trustedJourney = {
  runId: "run-trusted-outcome-journey",
  proofArtifactId: "proof-artifact-trusted-outcome",
  executionContractId: "execution-contract-trusted-outcome",
  teamId: "trusted-outcome-team",
  groupId: "trusted-outcome-group",
  workItemId: "trusted-outcome-work",
  degradedWorkItemId: "trusted-outcome-recovery",
  folder: "workspace/generated/trusted-outcome-kit",
  entrypoint: "workspace/generated/trusted-outcome-kit/index.html",
  packageTitle: "Trusted Outcome Kit",
};

export async function apiFetch<T>(page: Page, path: string, options?: BrowserFetchOptions) {
  return page.evaluate(
    async ({ url, init }) => {
      const response = await fetch(url, init);
      const body = await response.json();
      return { ok: response.ok, status: response.status, body };
    },
    { url: path, init: options },
  ) as Promise<{ ok: boolean; status: number; body: T }>;
}

export function trustedOutput() {
  const j = trustedJourney;
  return {
    kind: "project_package",
    title: j.packageTitle,
    id: "output-trusted-outcome",
    href: `/api/v1/workspace/files/view?path=${encodeURIComponent(j.entrypoint)}`,
    retained: true,
    entrypoint: j.entrypoint,
    folder: j.folder,
    files: ["index.html", "README.md", "PROOF.md"],
    validation: "Opened in browser and verified with retained proof.",
  };
}

function proposalEnvelope(): RouteResponse {
  const j = trustedJourney;
  return {
    status: 200,
    body: {
      ok: true,
      data: {
        meta: { source_node: "admin", timestamp: "2026-06-19T12:00:00Z" },
        signal_type: "chat.reply",
        trust_score: 0.91,
        template_id: "chat-to-proposal",
        mode: "proposal",
        payload: {
          text: "I can create a retained journey kit after approval.",
          tools_used: ["create_team", "write_file"],
          proposal: {
            intent: "create_trusted_outcome_kit",
            operator_summary: "Create a browser-openable output package with proof and recovery notes.",
            expected_result: "A retained Trusted Outcome Kit will open from Soma, Resources, Groups, and the run receipt.",
            affected_resources: [j.entrypoint, `${j.folder}/PROOF.md`],
            teams: 1,
            agents: 1,
            tools: ["create_team", "write_file"],
            risk_level: "medium",
            confirm_token: "confirm-trusted-outcome",
            intent_proof_id: "intent-proof-trusted-outcome",
            approval: {
              approval_required: false,
              approval_reason: "capability_risk",
              approval_mode: "optional",
              capability_risk: "medium",
              capability_ids: ["create_team", "write_file"],
              external_data_use: false,
              estimated_cost: 0,
            },
          },
        },
      },
    },
  };
}

function workItem(status: "completed" | "needs_recovery") {
  const j = trustedJourney;
  const degraded = status === "needs_recovery";
  return {
    work_item_id: degraded ? j.degradedWorkItemId : j.workItemId,
    team_id: j.teamId,
    run_id: j.runId,
    objective: degraded ? "Recovery sample for output dependency" : "Create trusted outcome kit",
    owner: "Soma",
    execution_shape: "deliverable",
    state: degraded ? "degraded" : "output_ready",
    last_event: {
      headline: degraded ? "Preview dependency failed; retained output still exists." : "Trusted Outcome Kit produced with retained proof.",
      details: degraded ? "Output and approval remain trusted, but preview proof needs recovery." : "Output package, run receipt, and proof artifact agree.",
      next_action: degraded ? "Repair preview dependency, then retry proof capture." : "Open the output or proof receipt.",
    },
    needs_operator: degraded,
    degradation_state: degraded ? "Preview dependency unavailable." : undefined,
    recovery_options: degraded ? ["Repair dependency or retry proof capture."] : [],
    expected_outputs: [j.entrypoint],
    expected_proof: [j.proofArtifactId],
    proof_refs: degraded ? [] : [j.proofArtifactId],
    audit_refs: [`audit-${degraded ? "recovery" : "output"}-trusted-outcome`],
    id: degraded ? j.degradedWorkItemId : j.workItemId,
    title: degraded ? "Recovery sample for output dependency" : "Create trusted outcome kit",
    status,
    source_agent: "Soma",
    assigned_team_id: j.teamId,
    created_at: "2026-06-19T12:01:00Z",
    updated_at: "2026-06-19T12:03:00Z",
    intent: degraded ? "recover missing preview dependency" : "create retained output package",
    summary: degraded
      ? "Preview dependency failed; output and approval remain trusted, preview proof needs recovery."
      : "Trusted Outcome Kit produced with retained proof.",
    next_step: degraded ? "Repair preview dependency, then retry proof capture." : "Open the output or proof receipt.",
    outputs: degraded ? [] : [trustedOutput()],
    proof: degraded ? { status: "degraded", recovery: "Preview proof invalid until dependency is repaired." } : { status: "verified", run_id: j.runId, artifact_id: j.proofArtifactId },
    trust: degraded ? "Output request remains trusted; preview proof is not reliable yet." : "Output package, run receipt, and proof artifact agree.",
    recovery: degraded
      ? { failed: "Preview dependency unavailable.", still_trusted: "Approval, request intent, and retained output reference.", not_trusted: "Preview screenshot proof for this attempt.", safe_next: "Repair dependency or retry proof capture." }
      : undefined,
  };
}

function groupRecord(): GroupRecord {
  return {
    group_id: trustedJourney.groupId,
    name: "Trusted Outcome Delivery Lane",
    work_mode: "propose_only",
    status: "ready",
    team_ids: [trustedJourney.teamId],
  };
}

function teamDetailRecord() {
  const j = trustedJourney;
  return {
    id: j.teamId, name: "Trusted Outcome Team", role: "delivery", type: "mission", mission_id: "mission-trusted-outcome",
    mission_intent: "Create retained outputs with proof and recovery ownership.",
    inputs: ["operator request", "execution contract"],
    deliveries: [j.entrypoint, `${j.folder}/PROOF.md`],
    agents: [
      { id: "trusted-outcome-lead", role: "lead", status: 2, last_heartbeat: "2026-06-19T12:03:00Z", tools: ["write_file", "store_artifact"], model: "balanced" },
    ],
  };
}

function artifactRecord(): ArtifactRecord {
  const j = trustedJourney;
  return {
    id: "artifact-trusted-outcome-kit",
    title: j.packageTitle,
    artifact_type: "project_package",
    team_id: j.teamId,
    agent_id: "soma",
    content_type: "text/html",
    file_path: j.entrypoint,
    status: "ready",
    created_at: "2026-06-19T12:03:00Z",
    metadata: trustedOutput(),
  };
}

export async function installTrustedOutcomeJourneyMocks(page: Page) {
  const j = trustedJourney;
  await mockOrganizationWorkspace(page, () => proposalEnvelope());
  await page.route("**/api/v1/services/status", ok({ data: [{ name: "core", status: "online" }] }));
  await page.route("**/api/v1/intent/confirm-action", ok({
    data: {
      run_id: j.runId,
      verified: true,
      execution_state: "completed",
      execution_summary: { outputs: [trustedOutput()], proof_artifact_id: j.proofArtifactId },
    },
  }));

  await page.route(/\/api\/v1\/groups(?:\?.*)?$/, ok({ data: [groupRecord()] }));
  await page.route(`**/api/v1/groups/${j.groupId}`, ok({ data: groupRecord() }));
  await page.route(`**/api/v1/groups/${j.groupId}/outputs**`, ok({ data: [artifactRecord()] }));
  await page.route(`**/api/v1/groups/${j.groupId}/workflow-log**`, ok({
    data: [
      { id: "log-1", role: "team_lead", message: "Soma delegated the retained package work.", created_at: "2026-06-19T12:01:00Z" },
      { id: "log-2", role: "coder", message: "Created browser output and proof note.", created_at: "2026-06-19T12:02:00Z" },
    ],
  }));

  await page.route("**/api/v1/teams/detail", async (route) => fulfillJSON(route, 200, [teamDetailRecord()]));
  await page.route(`**/api/v1/teams/${j.teamId}/work/ask`, ok({ data: workItem("completed") }));
  await page.route(/\/api\/v1\/teams\/trusted-outcome-team\/work(?:\?.*)?$/, ok({ data: [workItem("completed"), workItem("needs_recovery")] }));
  await page.route(`**/api/v1/teams/${j.teamId}/status-events**`, ok({
    data: [
      { id: "status-1", work_item_id: j.workItemId, status: "running", message: "Creating output package." },
      { id: "status-2", work_item_id: j.workItemId, status: "completed", message: "Output package verified." },
    ],
  }));
  await page.route(`**/api/v1/teams/${j.teamId}/interactions**`, ok({
    data: [
      { id: "interaction-1", role: "team_lead", content: "Plan confirmed.", created_at: "2026-06-19T12:01:00Z" },
      { id: "interaction-2", role: "qa", content: "Proof artifact linked.", created_at: "2026-06-19T12:03:00Z" },
    ],
  }));
  await page.route(`**/api/v1/teams/${j.teamId}/work/${j.degradedWorkItemId}/actions`, ok({ data: { status: "queued", action: "retry" } }));

  await page.route(`**/api/v1/trust/proof-artifacts/${j.proofArtifactId}`, ok({
    data: {
      id: j.proofArtifactId,
      run_id: j.runId,
      status: "verified",
      confidence: "high",
      summary: "Output package, run events, and group log reconstruct the same result.",
    },
  }));
  await page.route(`**/api/v1/trust/execution-contracts/${j.executionContractId}`, ok({
    data: { id: j.executionContractId, run_id: j.runId, approved: true, capability_ids: ["write_file"], output_refs: [j.entrypoint] },
  }));

  await page.route(/\/api\/v1\/runs(?:\?.*)?$/, ok({ data: [{ id: j.runId, status: "completed", title: "Trusted Outcome Kit run" }] }));
  await page.route(`**/api/v1/runs/${j.runId}`, ok({
    data: { id: j.runId, status: "completed", title: "Trusted Outcome Kit run", outputs: [trustedOutput()] },
  }));
  await page.route(`**/api/v1/runs/${j.runId}/events**`, ok({
    data: [
      {
        id: "event-start",
        event_type: "mission.started",
        timestamp: "2026-06-19T12:01:00Z",
        payload: { summary: "Operator asked Soma for a retained deliverable." },
      },
      {
        id: "event-output",
        event_type: "artifact.created",
        timestamp: "2026-06-19T12:02:00Z",
        payload: { title: j.packageTitle, path: j.entrypoint, proof_artifact_id: j.proofArtifactId },
      },
      {
        id: "event-completed",
        event_type: "mission.completed",
        timestamp: "2026-06-19T12:03:00Z",
        payload: { summary: "Trusted Outcome Kit delivered.", proof_artifact_id: j.proofArtifactId },
      },
    ],
  }));
  await page.route(`**/api/v1/runs/${j.runId}/conversation**`, ok({ data: [] }));

  await page.route("**/api/v1/mcp/servers", async (route) => {
    await fulfillJSON(route, 200, [{
      id: "filesystem-server",
      name: "filesystem",
      status: "connected",
      transport: "stdio",
      tools: [
        { id: "list-directory", name: "list_directory", description: "List workspace files" },
        { id: "read-text-file", name: "read_text_file", description: "Read workspace files" },
      ],
    }]);
  });
  await page.route("**/api/v1/mcp/servers/filesystem-server/tools/*/call", async (route) => {
    const tool = route.request().url().match(/\/tools\/([^/]+)\/call$/)?.[1] ?? "";
    const body = route.request().postDataJSON() as { arguments?: { path?: string } };
    const requestedPath = body.arguments?.path ?? "";
    if (tool === "list_directory" && requestedPath === j.folder) {
      await fulfillJSON(route, 200, { content: [{ type: "text", text: "[FILE] index.html\n[FILE] README.md\n[FILE] PROOF.md" }] });
      return;
    }
    if (tool === "list_directory") {
      await fulfillJSON(route, 200, { content: [{ type: "text", text: "[DIR] generated\n[DIR] groups\n[DIR] logs" }] });
      return;
    }
    if (tool === "read_text_file") {
      await fulfillJSON(route, 200, { content: [{ type: "text", text: "# Trusted Outcome Kit\nProof and recovery notes." }] });
      return;
    }
    await fulfillJSON(route, 404, { error: `unexpected filesystem tool call: ${tool}` });
  });
  await page.context().route("**/api/v1/workspace/files/view**", async (route) => {
    await route.fulfill({
      status: 200,
      contentType: "text/html",
      body: "<!doctype html><title>Trusted Outcome Kit</title><main><h1>Trusted Outcome Kit</h1><p>Recover, trust, and revisit this output.</p></main>",
    });
  });
  await page.route("**/api/v1/workspace/files/reveal**", ok({ data: { opened: true, path: j.folder } }));
}

function ok(body: Record<string, unknown>) {
  return async (route: Route) => fulfillJSON(route, 200, { ok: true, ...body });
}
