import { expect, test } from "@playwright/test";
import {
  createLiveGroup,
  createLiveTeamIDs,
  createOrganization,
  expectExecutionContractOutputs,
  openTeamCreation,
  openWorkspace,
  storeLiveArtifact,
  submitTeamDesign,
  submitWorkspaceChat,
} from "../support/workflow-variants-live";
import { reviewArchivedOutputs } from "../support/workflow-variants-groups";

test.describe("Workflow variants live backend contract", () => {
  test.skip(
    !process.env.PLAYWRIGHT_LIVE_BACKEND,
    "requires a live Core backend",
  );
  test.setTimeout(180_000);

  test("proves live direct answer plus compact and multi-team retained-output review", async ({
    page,
  }) => {
    test.slow();
    const stamp = Date.now();
    const organization = await createOrganization(
      page,
      `QA Workflow Variants ${stamp}`,
    );

    await openWorkspace(page, organization.id);

    const direct = await submitWorkspaceChat(
      page,
      "Summarize the current Workspace V8 design objectives.",
    );
    expect(
      direct.response.ok(),
      direct.body ? JSON.stringify(direct.body) : direct.raw,
    ).toBeTruthy();
    expect(direct.body?.data?.mode).toBe("answer");
    expect(direct.body?.data?.payload?.ask_class).toBe("direct_answer");
    expect(
      (direct.body?.data?.payload?.text ?? "").trim().length,
    ).toBeGreaterThan(0);
    await expect(
      page.getByText(/could not produce a readable reply/i),
    ).toHaveCount(0);

    await openTeamCreation(page, organization.id);
    await expect(page.getByText("Current organization")).toBeVisible();
    await expect(
      page.getByText(organization.name, { exact: true }).last(),
    ).toBeVisible();

    const compactPrompt =
      "Create a temporary marketing launch team for a new product rollout.";
    const compactGuidance = await submitTeamDesign(
      page,
      organization.id,
      compactPrompt,
    );
    const compactContract = compactGuidance.execution_contract!;
    expect(compactContract.execution_mode).toBe("native_team");
    expect(compactContract.coordination_model).toBe("compact_team");
    expect(compactContract.team_name).toBeTruthy();
    expect(compactContract.target_outputs?.length).toBe(3);
    expect(compactContract.workflow_group?.work_mode).toBe("propose_only");
    expect((compactGuidance.headline ?? "").trim().length).toBeGreaterThan(0);
    await expect(
      page.getByText(compactContract.team_name!, { exact: true }),
    ).toBeVisible();
    await expectExecutionContractOutputs(
      page,
      compactContract.target_outputs ?? [],
    );

    const compactTeamIDs = await createLiveTeamIDs(
      page,
      compactContract.target_outputs?.length ?? 3,
    );
    const compactGroup = await createLiveGroup(
      page,
      compactContract.workflow_group!,
      compactTeamIDs,
    );
    for (const [index, output] of (
      compactContract.target_outputs ?? []
    ).entries()) {
      await storeLiveArtifact(
        page,
        compactTeamIDs[index]!,
        output,
        `compact-lead-${index + 1}`,
        `# ${output}\n\n- Compact workflow proof\n- Live retained output`,
      );
    }
    await reviewArchivedOutputs(
      page,
      compactGroup,
      compactContract.target_outputs ?? [],
      compactTeamIDs.length,
    );

    await openTeamCreation(page, organization.id);
    const multiPrompt =
      "Create a company-wide product launch program across marketing, sales, support, docs, and engineering so the organization can coordinate several workstreams at once.";
    const multiGuidance = await submitTeamDesign(
      page,
      organization.id,
      multiPrompt,
    );
    const multiContract = multiGuidance.execution_contract!;
    expect(multiContract.execution_mode).toBe("native_team");
    expect(multiContract.coordination_model).toBe("multi_team_orchestration");
    expect(multiContract.recommended_team_count).toBe(3);
    expect(multiContract.initial_member_count).toBe(1);
    expect(multiContract.recommended_team_member_limit).toBe(3);
    expect(multiContract.workflow_group?.initial_member_count).toBe(1);
    expect(multiContract.target_outputs?.length).toBeGreaterThan(0);
    expect(multiContract.workflow_group?.work_mode).toBe("propose_only");
    expect((multiGuidance.headline ?? "").trim().length).toBeGreaterThan(0);
    if (multiContract.team_name) {
      await expect(
        page.getByText(multiContract.team_name, { exact: true }),
      ).toBeVisible();
    }
    await expectExecutionContractOutputs(
      page,
      multiContract.target_outputs ?? [],
    );

    const multiOutputs = (multiContract.target_outputs ?? []).slice(0, 3);
    const multiTeamIDs = await createLiveTeamIDs(
      page,
      multiOutputs.length || 3,
    );
    const multiGroup = await createLiveGroup(
      page,
      multiContract.workflow_group!,
      multiTeamIDs,
    );
    for (const [index, output] of multiOutputs.entries()) {
      await storeLiveArtifact(
        page,
        multiTeamIDs[index]!,
        output,
        `lane-lead-${index + 1}`,
        `# ${output}\n\n- Multi-team workflow proof\n- Live retained output`,
      );
    }
    await reviewArchivedOutputs(
      page,
      multiGroup,
      multiOutputs,
      multiTeamIDs.length,
    );
  });
});
