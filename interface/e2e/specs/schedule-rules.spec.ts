import { expect, test } from "@playwright/test";

test.describe("Schedule Rules", () => {
  test("renders propose-only cadence proof and recovery fields", async ({ page }) => {
    await page.route("**/api/v1/triggers", async (route) => {
      await route.fulfill({
        status: 200,
        contentType: "application/json",
        body: JSON.stringify({
          ok: true,
          data: [{
            id: "schedule-rule-1",
            tenant_id: "default",
            name: "Weekly evidence review",
            trigger_kind: "schedule",
            event_pattern: "scheduler.due",
            condition: {},
            target_mission_id: "mission-review",
            mode: "propose",
            cooldown_seconds: 3600,
            schedule_interval_seconds: 3600,
            next_run_at: new Date(Date.now() + 3600_000).toISOString(),
            proof_expectations: "Operator-visible result, audit event, and retained proof.",
            recovery_behavior: "Pause the schedule and review the last proposed run.",
            max_depth: 5,
            max_active_runs: 1,
            is_active: true,
            created_at: new Date().toISOString(),
            updated_at: new Date().toISOString(),
          }],
        }),
      });
    });

    await page.goto("/automations?tab=schedules");
    await expect(page.getByRole("button", { name: "Schedule Rules" })).toBeVisible();
    await expect(page.getByText("Weekly evidence review")).toBeVisible();
    await expect(page.getByText("propose only")).toBeVisible();
    await expect(page.getByText("1h cadence")).toBeVisible();
    await expect(page.getByText("Operator-visible result, audit event, and retained proof.")).toBeVisible();
    await expect(page.getByText("Pause the schedule and review the last proposed run.")).toBeVisible();
  });

  test("keeps long disabled schedule names from overlapping controls on narrow viewports", async ({ page }) => {
    await page.setViewportSize({ width: 390, height: 760 });
    await page.route("**/api/v1/triggers", async (route) => {
      await route.fulfill({
        status: 200,
        contentType: "application/json",
        body: JSON.stringify({
          ok: true,
          data: [{
            id: "schedule-rule-long",
            tenant_id: "default",
            name: "Extremely long weekly governed evidence review cadence for release readiness proof and trust package inspection",
            trigger_kind: "schedule",
            event_pattern: "scheduler.due",
            condition: {},
            target_mission_id: "mission-review",
            mode: "propose",
            cooldown_seconds: 3600,
            schedule_interval_seconds: 3600,
            next_run_at: new Date(Date.now() + 3600_000).toISOString(),
            proof_expectations: "Operator-visible result, audit event, and retained proof.",
            recovery_behavior: "Pause the schedule and review the last proposed run.",
            max_depth: 5,
            max_active_runs: 1,
            is_active: false,
            created_at: new Date().toISOString(),
            updated_at: new Date().toISOString(),
          }],
        }),
      });
    });

    await page.goto("/automations?tab=schedules");

    const name = page.getByTestId("schedule-rule-name-schedule-rule-long");
    const proposeBadge = page.getByTestId("schedule-rule-badge-propose-schedule-rule-long");
    const disabledBadge = page.getByTestId("schedule-rule-badge-disabled-schedule-rule-long");
    const resume = page.getByRole("button", { name: "Resume" });
    await expect(name).toBeVisible();
    await expect(proposeBadge).toBeVisible();
    await expect(disabledBadge).toBeVisible();
    await expect(resume).toBeVisible();

    const boxes = await Promise.all([
      name.boundingBox(),
      proposeBadge.boundingBox(),
      disabledBadge.boundingBox(),
      resume.boundingBox(),
    ]);
    for (const box of boxes) expect(box).not.toBeNull();

    const [nameBox, proposeBox, disabledBox, resumeBox] = boxes as NonNullable<typeof boxes[number]>[];
    const overlaps = (a: typeof nameBox, b: typeof nameBox) =>
      a.x < b.x + b.width && a.x + a.width > b.x && a.y < b.y + b.height && a.y + a.height > b.y;
    expect(overlaps(nameBox, proposeBox)).toBe(false);
    expect(overlaps(nameBox, disabledBox)).toBe(false);
    expect(overlaps(nameBox, resumeBox)).toBe(false);
  });
});
