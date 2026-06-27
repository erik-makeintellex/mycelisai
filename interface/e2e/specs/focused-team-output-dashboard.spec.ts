import { expect, test, type Page, type Route } from "@playwright/test";

const focusedTeamId = "focused-proof-team";
const focusedTeamName = "Focused Browser Proof Team";
const focusedOutputLabel = "Focused team newest output";
const focusedOutputPath = "generated/focused-team/newest-output.md";
const olderSomaOutputLabel = "Older Soma output";
const olderSomaOutputPath = "generated/soma-root/older-soma-output.md";
const focusedChatScope = `root::team::${focusedTeamId}`;

async function fulfillJSON(route: Route, status: number, body: unknown) {
  await route.fulfill({
    status,
    contentType: "application/json",
    body: JSON.stringify(body),
  });
}

async function installFocusedTeamOutputMocks(page: Page) {
  await page.addInitScript(
    ({ chatScope, olderLabel, olderPath }) => {
      const chatKeys = [];
      for (let index = 0; index < window.localStorage.length; index += 1) {
        const key = window.localStorage.key(index);
        if (key?.startsWith("mycelis-workspace-chat")) chatKeys.push(key);
      }
      chatKeys.forEach((key) => window.localStorage.removeItem(key));
      window.localStorage.setItem(
        `mycelis-workspace-chat:${chatScope}`,
        JSON.stringify([
          {
            role: "council",
            content: "Soma kept the earlier output for comparison, but the focused team has a newer lane result.",
            source_node: "admin",
            mode: "answer",
            timestamp: "2026-05-17T12:01:00Z",
            execution_summary: {
              execution_status: "completed",
              outputs: [
                {
                  kind: "file",
                  title: olderLabel,
                  path: olderPath,
                  retained: true,
                  proof_artifact_id: "proof-older-soma",
                },
              ],
              proof: [{ label: "Older Soma proof", run_id: "run-older-soma", verified: true }],
            },
          },
        ]),
      );
    },
    {
      chatScope: focusedChatScope,
      olderLabel: olderSomaOutputLabel,
      olderPath: olderSomaOutputPath,
    },
  );

  await page.route("**/auth/session", async (route) => {
    await fulfillJSON(route, 200, {
      ok: true,
      data: {
        user: {
          email: "operator@example.test",
          name: "QA Operator",
          role: "admin",
          provider: "local",
        },
      },
    });
  });

  await page.route("**/api/v1/user/me", async (route) => {
    await fulfillJSON(route, 200, {
      ok: true,
      data: { id: "operator-qa", name: "QA Operator", email: "operator@example.test" },
    });
  });

  await page.route("**/api/v1/services/status", async (route) => {
    await fulfillJSON(route, 200, {
      ok: true,
      data: [
        { name: "nats", status: "online", detail: "JetStream accepting team signals." },
        { name: "postgres", status: "online", detail: "Persistence responding." },
      ],
    });
  });

  await page.route("**/api/v1/teams/detail**", async (route) => {
    await fulfillJSON(route, 200, [
      {
        id: focusedTeamId,
        name: focusedTeamName,
        role: "builder",
        type: "mission",
        mission_id: "mission-focused-proof",
        mission_intent: "Keep focused-team output visible from the Soma dashboard.",
        inputs: ["operator request", "retained Soma context"],
        deliveries: [focusedOutputPath],
        agents: [
          {
            id: "focused-proof-lead",
            role: "lead",
            status: 2,
            last_heartbeat: "2026-05-17T12:10:00Z",
            tools: ["write_file", "store_artifact"],
            model: "balanced",
          },
        ],
      },
    ]);
  });

  await page.route("**/api/v1/teams/**", async (route) => {
    const url = new URL(route.request().url());
    if (url.pathname === "/api/v1/teams/detail") {
      await fulfillJSON(route, 200, [
        {
          id: focusedTeamId,
          name: focusedTeamName,
          role: "builder",
          type: "mission",
          mission_id: "mission-focused-proof",
          mission_intent: "Keep focused-team output visible from the Soma dashboard.",
          inputs: ["operator request", "retained Soma context"],
          deliveries: [focusedOutputPath],
          agents: [
            {
              id: "focused-proof-lead",
              role: "lead",
              status: 2,
              last_heartbeat: "2026-05-17T12:10:00Z",
              tools: ["write_file", "store_artifact"],
              model: "balanced",
            },
          ],
        },
      ]);
      return;
    }

    if (url.pathname === `/api/v1/teams/${focusedTeamId}/work`) {
      await fulfillJSON(route, 200, {
        ok: true,
        data: [
          {
            work_item_id: "work-focused-newest",
            team_id: focusedTeamId,
            run_id: "run-focused-newest",
            objective: "Focused dashboard proof package",
            execution_shape: "deliverable",
            state: "output_ready",
            last_event: {
              headline: "Focused output retained",
              details: "The newest team output is attached to this focused lane.",
              next_action: "Review output from Soma dashboard.",
            },
            output_refs: [
              {
                output_id: "out-focused-newest",
                team_id: focusedTeamId,
                work_item_id: "work-focused-newest",
                run_id: "run-focused-newest",
                kind: "file",
                label: focusedOutputLabel,
                storage_ref: focusedOutputPath,
                proof_id: "proof-focused-newest",
                created_at: "2026-05-17T12:10:00Z",
              },
            ],
            proof_refs: ["proof-focused-newest"],
            audit_refs: ["audit-focused-newest"],
            updated_at: "2026-05-17T12:10:00Z",
          },
        ],
      });
      return;
    }

    await fulfillJSON(route, 200, { ok: true, data: [] });
  });

  await page.route("**/api/v1/workspace/files/reveal?**", async (route) => {
    await fulfillJSON(route, 200, { ok: true, data: { opened: true } });
  });

  await page.route("**/api/v1/workspace/files/view?**", async (route) => {
    const url = new URL(route.request().url());
    const path = url.searchParams.get("path") ?? "unknown";
    await route.fulfill({
      status: 200,
      contentType: "text/html",
      body: `<!doctype html><title>${path}</title><main>${path}</main>`,
    });
  });
}

async function expectFocusedDashboardLane(page: Page) {
  await expect(page.getByTestId("soma-operating-surface")).toBeVisible({ timeout: 20_000 });
  await expect(page.getByText(`Mode: ${focusedTeamName}`)).toBeVisible();

  await expect(page.getByTestId("soma-context-focus-bar")).toHaveCount(0);
  await expect(page.getByTestId("focused-team-output-dock")).toHaveCount(0);
  const switcher = page.getByTestId("soma-team-context-switcher");
  await expect(switcher).toBeVisible();
  await expect(switcher).toContainText("Working in");
  await expect(switcher).toContainText(focusedTeamName);
  await expect(switcher).toContainText("Team chat, work, outputs, and proof");

  await page.getByRole("button", { name: /Open Outcome Vault/i }).click();
  const vault = page.getByTestId("soma-outcome-vault");
  await expect(vault).toContainText("Outcome ready to revisit");
  await expect(vault).toContainText(`${focusedTeamName} outcome workspace`);
  await expect(vault).not.toContainText("OutcomeProject owner:");
  await expect(vault).not.toContainText("TeamRegistry owner:");
  await expect(vault.getByRole("link", { name: "Revisit work", exact: true })).toHaveAttribute("href", "/teams?view=work");
  await expect(vault.getByRole("link", { name: "Open saved outcomes", exact: true })).toHaveAttribute("href", "/resources?tab=workspace");
  await vault.getByRole("button", { name: /Close Outcome Vault/i }).click();
  await expect(page.getByTestId("soma-outcome-vault")).toHaveCount(0);

  await switcher.getByRole("button", { name: /Focused Browser Proof Team/i }).click();
  await expect(page.getByRole("listbox", { name: "Choose current workflow" })).toBeVisible();
  await expect(page.getByRole("option", { name: /Soma root/i })).toBeVisible();
  await expect(page.getByRole("option", { name: /Focused Browser Proof Team/i })).toBeVisible();
  await page.keyboard.press("Escape");
  await expect(page.getByRole("link", { name: /Manage teams/i })).toHaveAttribute("href", "/teams");

  const digest = page.getByTestId("soma-workbench-output-digest");
  await expect(digest).toBeVisible();
  await expect(digest).toContainText("Latest output");
  await expect(digest).toContainText(focusedOutputLabel);
  await expect(digest).toContainText(focusedOutputPath);
  await expect(digest).not.toContainText(olderSomaOutputLabel);
}

test.describe("Dashboard focused-team output proof", () => {
  test("keeps URL-focused team output ahead of older Soma output through controls and reload", async ({ page }) => {
    await installFocusedTeamOutputMocks(page);

    await page.goto(`/dashboard?team_id=${focusedTeamId}`, { waitUntil: "domcontentloaded" });
    await expectFocusedDashboardLane(page);

    const digestFolderButton = page
      .getByTestId("soma-workbench-output-digest")
      .getByRole("button", { name: /Open local folder/i });
    await digestFolderButton.click();
    await expect(digestFolderButton).toContainText("Folder opened");

    const panelToggle = page.getByTestId("soma-workbench-panel-toggle");
    await expect(panelToggle).toContainText("Review output");
    await expect(panelToggle).toContainText("2");
    await expect(panelToggle).toHaveAttribute("aria-expanded", "false");
    await panelToggle.click();
    await expect(panelToggle).toHaveAttribute("aria-expanded", "true");

    const rail = page.getByTestId("soma-workbench-side-rail");
    await expect(rail).toHaveAttribute("aria-hidden", "false");
    await expect(rail.getByRole("tab", { name: /Output/i })).toHaveAttribute("aria-selected", "true");
    await expect(rail.getByText(focusedOutputLabel).first()).toBeVisible();

    await rail.getByText("More outputs and verification").click();
    await expect(rail.getByText(olderSomaOutputLabel).first()).toBeVisible();

    const railText = await rail.textContent();
    expect(railText?.indexOf(focusedOutputLabel)).toBeGreaterThanOrEqual(0);
    expect(railText?.indexOf(olderSomaOutputLabel)).toBeGreaterThanOrEqual(0);
    expect(railText!.indexOf(focusedOutputLabel)).toBeLessThan(railText!.indexOf(olderSomaOutputLabel));
    await expect(
      rail.getByRole("button", { name: `Open file ${focusedOutputLabel} in a new browser window` }),
    ).toBeVisible();
    await expect(
      rail.getByRole("button", { name: new RegExp(`Open local folder for ${focusedOutputLabel}`) }),
    ).toBeVisible();

    await page.reload({ waitUntil: "domcontentloaded" });
    await expect(page).toHaveURL(new RegExp(`/dashboard\\?team_id=${focusedTeamId}$`));
    await expectFocusedDashboardLane(page);
  });
});
