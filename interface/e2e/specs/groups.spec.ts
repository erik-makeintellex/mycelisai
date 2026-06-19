import { expect, test } from "@playwright/test";
import { mockGroupsWorkspace, openGroups } from "../support/groups-workspace";

test.describe("Groups workspace (/groups)", () => {
  test("shows active and archived groups with tabbed output review", async ({ page }) => {
    await mockGroupsWorkspace(page);
    await openGroups(page);

    await expect(
      page.getByRole("heading", { name: "Manage focused collaboration lanes." }),
    ).toBeVisible();
    await expect(page.getByRole("heading", { name: "Standing groups" })).toBeVisible();
    await expect(page.getByRole("heading", { name: "Temporary groups" })).toBeVisible();
    await expect(page.getByRole("heading", { name: "Completed records" })).toBeVisible();
    await expect(page.getByRole("tab", { name: /Overview/i })).toBeVisible();
    await expect(page.getByRole("tab", { name: /Workflow Log/i })).toBeVisible();
    await expect(page.getByRole("tab", { name: /Outputs/i })).toBeVisible();
    await expect(page.getByRole("tab", { name: /Message/i })).toBeVisible();
    await expect(page.getByRole("tab", { name: /Settings/i })).toBeVisible();
    await expect(page.getByRole("tab", { name: /Create/i })).toBeVisible();

    await page.getByTestId("groups-list-item-group-temp-launch").click();
    await expect(page.getByText("Temporary group", { exact: true })).toBeVisible();
    await expect(page.getByTestId("groups-output-summary")).toContainText("1 output");
    await expect(page.getByTestId("groups-output-summary")).toContainText(
      "1 contributing lead",
    );
    await expect(page.getByRole("link", { name: "Open Soma", exact: true })).toHaveCount(2);
    await expect(
      page.getByLabel("OverviewScope and links").getByRole("link", { name: "Open Soma" }),
    ).toHaveAttribute("href", "/dashboard");
    await expect(page.getByTitle("Open launch-lead lead")).toHaveAttribute(
      "href",
      "/dashboard?team_id=launch-lead",
    );
    await page.getByRole("tab", { name: /Workflow Log/i }).click();
    await expect(page.getByTestId("groups-workflow-log")).toContainText(
      "Prepare launch brief and asset bundle",
    );
    await expect(page.getByTestId("groups-workflow-log")).toContainText("Launch Brief");
    await expect(page.getByTestId("groups-workflow-log")).toContainText(
      "run run-launch-proof",
    );
    await page.getByRole("tab", { name: /Outputs/i }).click();
    await expect(page.getByText("Launch Brief", { exact: true })).toBeVisible();
    await expect(page.getByRole("link", { name: "Download" }).first()).toHaveAttribute(
      "href",
      "/api/v1/artifacts/artifact-brief/download",
    );

    await page.getByTestId("groups-list-item-group-temp-archived").click();
    await expect(
      page.getByText("Archived temporary group", { exact: true }),
    ).toBeVisible();
    await page.getByRole("tab", { name: /Message/i }).click();
    await expect(page.getByRole("link", { name: "Open bus diagnostics" })).toHaveAttribute(
      "href",
      "/system?tab=nats&advanced=1",
    );
    await expect(page.getByTestId("groups-archived-readonly-note")).toContainText(
      "retained output review",
    );
    await page.getByRole("tab", { name: /Outputs/i }).click();
    await expect(page.getByTestId("groups-retained-outputs-note")).toContainText(
      "Downloads remain available",
    );
    await expect(page.getByText("Retrospective Summary")).toBeVisible();
    await expect(page.getByText("Approved outputs")).toBeVisible();
    await expect(page.getByRole("button", { name: "Broadcast to group" })).toHaveCount(0);
  });

  test("archives a temporary group and keeps retained outputs visible", async ({ page }) => {
    const harness = await mockGroupsWorkspace(page);
    await openGroups(page);

    await page.getByTestId("groups-list-item-group-temp-launch").click();
    await page.getByRole("button", { name: "Archive temporary group" }).click();

    await expect(page.getByTestId("groups-notice")).toContainText(
      "Temporary group archived.",
    );
    await expect.poll(() => harness.readStatusBodies().length).toBe(1);
    await expect.poll(() => String(harness.readStatusBodies()[0]?.status ?? "")).toBe(
      "archived",
    );
    await expect(
      page.getByText("Archived temporary group", { exact: true }),
    ).toBeVisible();
    await page.getByRole("tab", { name: /Message/i }).click();
    await expect(page.getByTestId("groups-archived-readonly-note")).toContainText(
      "retained output review",
    );
    await page.getByRole("tab", { name: /Outputs/i }).click();
    await expect(page.getByTestId("groups-retained-outputs-note")).toContainText(
      "Downloads remain available",
    );
    await expect(page.getByText("Launch Brief", { exact: true })).toBeVisible();
    await expect(page.getByRole("button", { name: "Broadcast to group" })).toHaveCount(0);
  });

  test("opens a selected group directly into a requested panel", async ({ page }) => {
    await mockGroupsWorkspace(page);
    await page.goto("/groups?group_id=group-temp-launch&panel=workflow", {
      waitUntil: "domcontentloaded",
    });

    await expect(page.getByRole("heading", { name: /Manage focused collaboration lanes/i })).toBeVisible();
    await expect(page.getByTestId("groups-list-item-group-temp-launch")).toHaveAttribute(
      "aria-current",
      "true",
    );
    await expect(page.getByRole("tab", { name: /Workflow Log/i })).toHaveAttribute(
      "aria-selected",
      "true",
    );
    await expect(page.getByTestId("groups-workflow-log")).toContainText(
      "Prepare launch brief and asset bundle",
    );
  });

  test("creates a temporary group and supports broadcast workflow", async ({ page }) => {
    const harness = await mockGroupsWorkspace(page);
    await openGroups(page);

    await page.getByRole("button", { name: "Create group" }).click();
    await page.getByLabel("Name").fill("Regional Expansion Sprint");
    await page
      .getByLabel("Goal Statement")
      .fill("Prepare outreach, pricing notes, and launch assets for a new region.");
    await page.getByLabel("Work Mode").selectOption("execute_with_approval");
    await page.getByLabel("Expiry").fill("2026-04-15T09:30");
    await page.getByLabel("Team IDs").fill("launch-lead, design-lead");
    await page.getByLabel("Member IDs").fill("owner, marketing-lead");
    await page.getByLabel("Coordinator Profile").fill("Regional expansion lead");
    await page.getByLabel("Allowed Capabilities").fill("write_file, publish_signal");
    await page.getByLabel("Approval Policy Ref").fill("regional-expansion");
    await page.getByTestId("groups-create-button").click();
    await expect(page.getByTestId("groups-notice")).toContainText(
      "Group created successfully.",
    );

    await page.getByTestId("groups-list-item-group-temp-launch").click();
    await page.getByRole("tab", { name: /Message/i }).click();
    await page
      .getByLabel("Broadcast message")
      .fill("Generate a launch brief, pricing checklist, and asset package.");
    await page.getByRole("button", { name: "Broadcast to group" }).click();

    await expect(page.getByTestId("groups-notice")).toContainText(
      "Broadcast queued for the selected group.",
    );
    await page.getByRole("tab", { name: /Overview/i }).click();
    await expect(page.getByTestId("groups-output-summary")).toContainText("1 output");
    await expect.poll(() => harness.readBroadcastBodies().length).toBe(1);
    await expect
      .poll(() => String(harness.readBroadcastBodies()[0]?.message ?? ""))
      .toContain("pricing checklist");
  });
});
