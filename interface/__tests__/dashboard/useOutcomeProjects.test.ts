import { describe, expect, it } from "vitest";
import { outcomeProjectSummaryFromAPI, selectOutcomeProject } from "@/components/soma/useOutcomeProjects";
import type { TeamDetailEntry } from "@/store/useCortexStore";

const teams: TeamDetailEntry[] = [{
  id: "media-team",
  name: "Media Team",
  role: "coordinator",
  type: "standing",
  mission_id: null,
  mission_intent: null,
  inputs: [],
  deliveries: [],
  agents: [],
}];

describe("outcomeProjectSummaryFromAPI", () => {
  it("maps durable outcome project state into a user-facing Vault card", () => {
    const summary = outcomeProjectSummaryFromAPI({
      project_id: "project-1",
      outcome_id: "outcome-1",
      title: "Weekly Media Pack",
      status: "output_ready",
      run_id: "run-123",
      workspace_folder: "groups/media/generated",
      work_item_refs: ["work-1"],
      team_registry_refs: ["registry-1"],
      output_refs: [{
        output_id: "output-1",
        team_id: "media-team",
        work_item_id: "work-1",
        kind: "project_package",
        label: "Media pack",
        storage_ref: "groups/media/generated/index.html",
      }],
    }, teams);

    expect(summary.title).toBe("Weekly Media Pack");
    expect(summary.detail).toContain("durable project");
    expect(summary.ownerLabel).toBe("Soma");
    expect(summary.leadLabel).toBe("Media Team, coordinator");
    expect(summary.registryOwnerLabel).toBe("Media Team lead");
    expect(summary.teamCount).toBe(1);
    expect(summary.workCount).toBe(1);
    expect(summary.outputCount).toBe(1);
    expect(summary.recoveryCount).toBe(0);
    expect(summary.href).toBe("/runs/run-123");
    expect(summary.hrefLabel).toBe("Open run receipt");
    expect(summary.outputHref).toBe("/resources?tab=workspace&path=workspace%2Fgroups%2Fmedia%2Fgenerated");
    expect(summary.outputLabel).toBe("Open outputs");
  });

  it("marks durable projects with recovery refs as needing trust review", () => {
    const summary = outcomeProjectSummaryFromAPI({
      project_id: "project-2",
      outcome_id: "outcome-2",
      title: "Audit Review",
      status: "needs_attention",
      recovery_refs: ["recover-1"],
      team_registry_refs: ["registry-1"],
    }, []);

    expect(summary.detail).toContain("recovery work");
    expect(summary.registryOwnerLabel).toBe("Registered team lead");
    expect(summary.recoveryCount).toBe(1);
    expect(summary.href).toBe("/teams?view=work");
  });

  it("uses target refs for the primary outcome link when provided", () => {
    const summary = outcomeProjectSummaryFromAPI({
      project_id: "project-targeted",
      outcome_id: "outcome-targeted",
      title: "Targeted Review",
      status: "needs_attention",
      run_id: "run-fallback",
      target_ref: {
        type: "recovery",
        id: "recovery-target",
        work_item_id: "work-targeted",
        label: "Recovery item",
      },
    }, []);

    expect(summary.href).toBe("/teams?view=work&work_item_id=work-targeted");
    expect(summary.hrefLabel).toBe("Open work");
    expect(summary.targetReference).toBe("recovery:recovery-target");
  });

  it("does not let an unrelated latest project override focused team context", () => {
    const projects = [{
      project_id: "project-3",
      outcome_id: "outcome-3",
      title: "Unrelated latest project",
      status: "output_ready" as const,
      output_refs: [{
        output_id: "output-3",
        team_id: "other-team",
        work_item_id: "work-3",
        kind: "project_package",
        label: "Other output",
      }],
    }];

    expect(selectOutcomeProject(projects, "media-team")).toBeNull();
    expect(selectOutcomeProject(projects, null)?.title).toBe("Unrelated latest project");
  });
});
