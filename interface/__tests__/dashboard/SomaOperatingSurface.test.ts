import { describe, expect, it } from "vitest";
import { prioritizeSomaHomeWorkItems } from "@/components/soma/SomaOperatingSurface";
import { railAlertsFromWorkItems } from "@/components/soma/SomaOperatingSurfaceSupport";
import type { TeamWorkItem } from "@/store/useCortexStore";

function item(
  id: string,
  state: TeamWorkItem["state"],
  updatedAt: string,
  overrides: Partial<TeamWorkItem> = {},
): TeamWorkItem {
  return {
    id,
    title: id,
    state,
    updatedAt,
    ownerLabel: "Soma",
    scopeLabel: "Team work",
    teamIds: ["team-1"],
    interactions: [],
    ...overrides,
  };
}

describe("prioritizeSomaHomeWorkItems", () => {
  it("keeps operator attention and degraded work ahead of newer output history", () => {
    const sorted = prioritizeSomaHomeWorkItems([
      item("newer-output", "output_ready", "2026-05-20T20:00:00Z"),
      item("needs-operator", "queued", "2026-05-20T18:00:00Z", { needsOperator: true }),
      item("degraded", "degraded", "2026-05-20T17:00:00Z"),
      item("running", "running", "2026-05-20T19:00:00Z"),
      item("fallback", "running", "2026-05-20T21:00:00Z", { source: "projection" }),
    ]);

    expect(sorted.map((work) => work.id)).toEqual([
      "needs-operator",
      "degraded",
      "running",
      "newer-output",
      "fallback",
    ]);
  });
});

describe("railAlertsFromWorkItems", () => {
  it("uses API target refs for quiet rail links while preserving work focus", () => {
    const [alert] = railAlertsFromWorkItems([
      item("work-1", "degraded", "2026-05-20T17:00:00Z", {
        targetRef: {
          type: "recovery",
          id: "recovery-target-1",
          team_id: "team-1",
          work_item_id: "work-1",
          label: "Recovery target",
        },
      }),
    ]);

    expect(alert.href).toBe("/teams?view=work&work_item_id=work-1");
    expect(alert.targetReference).toBe("recovery:recovery-target-1");
    expect(alert.target).toMatchObject({
      type: "recovery",
      id: "recovery-target-1",
      label: "Recovery target",
    });
  });

  it("falls back to run targets when no API target ref is present", () => {
    const [alert] = railAlertsFromWorkItems([
      item("work-2", "degraded", "2026-05-20T17:00:00Z", { runId: "run-2" }),
    ]);

    expect(alert.href).toBe("/runs/run-2");
    expect(alert.targetReference).toBe("run:run-2");
  });
});
