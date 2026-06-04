import { expect, test, type Route } from "@playwright/test";
import {
  confirmProposal,
  createOrganization,
  fulfillJSON,
  liveAPIGet,
  liveTimeoutMs,
  openLiveWorkspace,
  submitLiveWorkspaceChat,
} from "../support/finalization-browser-package";
import {
  mockOrganizationWorkspace,
  openOrganization,
  sendWorkspaceMessage,
  type ChatRequestBody,
  type RouteResponse,
} from "../support/soma-ui-testing";

type MediaOutput = {
  kind?: string;
  title?: string;
  id?: string;
  href?: string;
  retained?: boolean;
  proof_artifact_id?: string;
  proof?: {
    proof_id?: string;
    source_run_id?: string;
    source_contract_id?: string;
    path_boundary_status?: string;
    readback_status?: string;
  };
};

type ConfirmActionData = {
  run_id?: string;
  verified?: boolean;
  execution_state?: string;
  audit_event_id?: string;
  proof_artifact_id?: string;
  intent_proof_id?: string;
  contract_id?: string;
  execution_summary?: {
    outputs?: MediaOutput[];
    proof?: {
      run_id?: string;
      proof_id?: string;
      audit_event_id?: string;
      intent_proof_id?: string;
      contract_id?: string;
      proof_class?: string;
      run_class?: string;
      verified?: boolean;
    };
  };
};

type ProofRecord = {
  id?: string;
  status?: string;
  proof_class?: string;
  validation_source?: string;
  evidence_strength?: string;
  proof_quality?: string;
  audit_refs?: Array<Record<string, string>>;
  output_refs?: Array<Record<string, unknown>>;
};

test.skip(({ browserName }) => browserName !== "chromium", "ComfyUI media proof is stabilized in Chromium.");

function escaped(value: string) {
  return value.replace(/[.*+?^${}()|[\]\\]/g, "\\$&");
}

function mediaTeamProposal(): RouteResponse {
  return {
    status: 200,
    body: {
      ok: true,
      data: {
        meta: { source_node: "admin", timestamp: "2026-06-01T12:00:00Z" },
        signal_type: "chat.reply",
        trust_score: 0.91,
        template_id: "chat-to-proposal",
        mode: "proposal",
        payload: {
          text: "I can create a local/private media team output and retain it for review.",
          tools_used: ["create_team", "generate_image", "save_cached_image"],
          consultations: [],
          artifacts: [],
          proposal: {
            intent: "create_local_private_media_team_output",
            operator_summary: "create a media team and generate one image through local/private ComfyUI.",
            expected_result: "A retained media output will be available with proof after approval.",
            affected_resources: ["workspace/saved-media/comfyui-journey/storyboard-frame.png"],
            teams: 1,
            agents: 2,
            tools: ["create_team", "generate_image", "save_cached_image"],
            risk_level: "medium",
            confirm_token: "confirm-media-comfyui-journey",
            intent_proof_id: "proof-media-comfyui-journey",
            approval_required: false,
            approval_reason: "capability_risk",
            approval_mode: "optional",
            capability_risk: "medium",
            capability_ids: ["create_team", "generate_image", "save_cached_image"],
            external_data_use: false,
            estimated_cost: 0,
          },
        },
      },
    },
  };
}

async function mockUnavailableComfyUIExecution(pageRoute: Route) {
  await fulfillJSON(pageRoute, 500, {
    ok: false,
    error: "local/private ComfyUI gateway unavailable",
    data: {
      run_id: "run-media-comfyui-unavailable",
      execution_summary: {
        execution: {
          shape: "directed_execution",
          status: "failed",
          summary: "Approved media execution could not reach the local/private ComfyUI gateway.",
        },
        capability_use: [
          { id: "generate_image", label: "generate_image", status: "failed" },
          { id: "save_cached_image", label: "save_cached_image", status: "not_started" },
        ],
        outputs: [],
        proof: {
          run_id: "run-media-comfyui-unavailable",
          proof_class: "failed_run",
          verified: false,
        },
        audit_recovery: {
          approval_status: "approved",
          recovery_state: "blocked",
          blocker: "local/private ComfyUI gateway unavailable",
          retryable: true,
          degradation: {
            code: "media_provider_unavailable",
            what_failed: "local/private ComfyUI gateway unavailable",
            trusted_state: "The approval, intent proof, failed run record, and audit event remain trusted.",
            invalidated_proof: "No generated media output or retained image should be trusted for this attempt.",
            safe_continuation: "Restore the local/private media gateway, confirm ComfyUI health, then retry the media proposal.",
            requires_attention: true,
          },
        },
      },
    },
  });
}

test.describe("Soma ComfyUI media journey", () => {
  test("mocked unavailable local media provider shows trusted degradation guidance", async ({ page }) => {
    await mockOrganizationWorkspace(page, (_requestBody: ChatRequestBody) => mediaTeamProposal());
    await page.route("**/api/v1/intent/confirm-action", mockUnavailableComfyUIExecution);

    await openOrganization(page);
    await sendWorkspaceMessage(page, "Create a local/private ComfyUI media team output.");

    await expect(page.getByText("PROPOSED ACTION").last()).toBeVisible({ timeout: 20_000 });
    await page.getByRole("button", { name: /Approve & Execute|Execute|Run/i }).last().click();

    const failureCard = page.getByTestId("execution-summary-card").last();
    await expect(failureCard.getByText("Needs review").first()).toBeVisible({ timeout: 20_000 });
    await expect(failureCard.getByText("Review request, proof, and recovery")).toBeVisible();
    await failureCard.getByText("Review request, proof, and recovery").click();
    await expect(failureCard.getByText("Failed: local/private ComfyUI gateway unavailable")).toBeVisible();
    await expect(
      failureCard.getByText("Still available: The approval, intent proof, failed run record, and audit event remain trusted."),
    ).toBeVisible();
    await expect(
      failureCard.getByText("Not reliable: No generated media output or retained image should be trusted for this attempt."),
    ).toBeVisible();
    await expect(
      failureCard.getByText("Safe next: Restore the local/private media gateway, confirm ComfyUI health, then retry the media proposal."),
    ).toBeVisible();
    await expect(page.getByText("Run proof + retained output")).toHaveCount(0);
  });

  test("live Soma media run proves ComfyUI identity, retained output, and proof reload", async ({ page }) => {
    test.skip(
      !process.env.PLAYWRIGHT_LIVE_MEDIA_RETAINED_OUTPUT,
      "requires live Core, media gateway, and ComfyUI; set PLAYWRIGHT_LIVE_MEDIA_RETAINED_OUTPUT=1",
    );
    test.setTimeout(liveTimeoutMs);
    test.slow();

    const healthResponse = await page.request.get("http://127.0.0.1:8001/health");
    expect(healthResponse.ok(), await healthResponse.text()).toBeTruthy();
    const health = await healthResponse.json();
    expect(health.backend).toBe("comfyui");
    expect(health.upstream).toContain("127.0.0.1:8188");

    const stamp = Date.now();
    const organizationId = await createOrganization(page, `QA ComfyUI Media Journey ${stamp}`);
    const teamID = `qa-comfyui-media-team-${stamp}`;

    await openLiveWorkspace(page, organizationId);
    const proposal = await submitLiveWorkspaceChat(
      page,
      [
        `Create a compact local/private media team with team_id ${teamID}.`,
        "Generate one polished comic storyboard image with the configured local/private ComfyUI provider.",
        "Save the retained output under saved-media and return the saved workspace file as a retained output.",
        "Expose the saved media output, run proof, and local folder access for operator review.",
      ].join(" "),
    );
    expect(proposal.response.ok(), proposal.body ? JSON.stringify(proposal.body) : proposal.raw).toBeTruthy();
    expect(proposal.body?.data?.mode).toBe("proposal");
    await expect(page.getByText("PROPOSED ACTION").last()).toBeVisible({ timeout: 30_000 });

    const confirmed = await confirmProposal(page);
    expect(confirmed.response.ok(), confirmed.body ? JSON.stringify(confirmed.body) : confirmed.raw).toBeTruthy();
    const data = confirmed.body?.data as ConfirmActionData | undefined;
    expect(data?.verified).toBeTruthy();
    expect(data?.execution_state).toBe("verified");
    expect(data?.run_id).toBeTruthy();
    expect(data?.audit_event_id).toBeTruthy();
    expect(data?.proof_artifact_id).toBeTruthy();
    expect(data?.contract_id).toBeTruthy();

    const proof = data?.execution_summary?.proof;
    expect(proof?.run_id).toBe(data?.run_id);
    expect(proof?.proof_id).toBe(data?.proof_artifact_id);
    expect(proof?.audit_event_id).toBe(data?.audit_event_id);
    expect(proof?.intent_proof_id).toBe(data?.intent_proof_id);
    expect(proof?.contract_id).toBe(data?.contract_id);
    expect(proof?.proof_class).toBe("run_and_audit");
    expect(proof?.run_class).toBe("run_linked");
    expect(proof?.verified).toBeTruthy();

    const outputs = data?.execution_summary?.outputs ?? [];
    const mediaOutput = outputs.find((output) => (
      output.retained
      && (output.kind === "file" || output.kind === "image")
      && typeof output.href === "string"
      && output.href.includes("/api/v1/workspace/files/view")
    ));
    expect(mediaOutput, JSON.stringify(outputs)).toBeTruthy();
    expect(mediaOutput?.proof_artifact_id).toBe(data?.proof_artifact_id);
    expect(mediaOutput?.proof?.proof_id).toBe(data?.proof_artifact_id);
    expect(mediaOutput?.proof?.source_run_id).toBe(data?.run_id);
    expect(mediaOutput?.proof?.source_contract_id).toBe(data?.contract_id);
    expect(mediaOutput?.proof?.path_boundary_status).toBe("verified");
    expect(mediaOutput?.proof?.readback_status).toBe("verified");

    const outputLabel = mediaOutput!.title || mediaOutput!.id || "Team output";
    await expect(page.getByText("Run proof + retained output").last()).toBeVisible({ timeout: 30_000 });
    await expect(page.getByText(outputLabel).last()).toBeVisible();
    await expect(
      page.getByRole("button", { name: new RegExp(`Open .*${escaped(outputLabel)}.*browser window`) }).last(),
    ).toBeVisible();
    await expect(page.getByRole("button", { name: `Open local folder for ${outputLabel}` }).last()).toBeVisible();

    const firstRead = await liveAPIGet(page, mediaOutput!.href!);
    expect(firstRead.ok(), await firstRead.text()).toBeTruthy();
    expect((await firstRead.body()).length).toBeGreaterThan(0);
    const reloadRead = await liveAPIGet(page, mediaOutput!.href!);
    expect(reloadRead.ok(), await reloadRead.text()).toBeTruthy();
    expect((await reloadRead.body()).length).toBeGreaterThan(0);

    const trustResponse = await liveAPIGet(
      page,
      `/api/v1/trust/proof-artifacts?run_id=${encodeURIComponent(data!.run_id!)}&limit=5`,
    );
    expect(trustResponse.ok(), await trustResponse.text()).toBeTruthy();
    const records = ((await trustResponse.json()) as { data?: ProofRecord[] }).data ?? [];
    const record = records.find((item) => item.id === data?.proof_artifact_id);
    expect(record, JSON.stringify(records)).toBeTruthy();
    expect(record?.status).toBe("success");
    expect(record?.proof_class).toBe("run_and_audit");
    expect(record?.validation_source).toBe("confirm_action");
    expect(record?.evidence_strength).toBe("run_audit");
    expect(record?.proof_quality).toBe("verified");
    expect(record?.audit_refs).toEqual(
      expect.arrayContaining([expect.objectContaining({ audit_event_id: data?.audit_event_id })]),
    );
    expect(record?.output_refs).toEqual(
      expect.arrayContaining([expect.objectContaining({ proof_artifact_id: data?.proof_artifact_id })]),
    );
  });
});
