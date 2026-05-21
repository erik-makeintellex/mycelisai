import { describe, expect, it } from "vitest";
import { prioritizeSomaHomeWorkItems } from "@/components/soma/SomaOperatingSurface";
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
