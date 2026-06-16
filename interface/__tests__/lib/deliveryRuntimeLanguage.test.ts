import { describe, expect, it } from "vitest";
import {
  DEGRADED_TEAM_WORK_REVIEW_COPY,
  OUTPUT_PACKAGE_ACTION_LABELS,
  RUN_RECEIPT_SECTION_LABELS,
  STALE_FAILED_PLAN_REVIEW_COPY,
  localMediaDependencyRecovery,
  outputFolderButtonLabel,
  outputPackageKindLabel,
  recoveryTrustLines,
  teamWorkStateGroup,
  teamWorkStateLabel,
} from "@/lib/deliveryRuntimeLanguage";
import type { TeamWorkItemState } from "@/store/useCortexStore";
import { buildMediaDegradation, buildTeamOutputRef, buildTeamWorkItem } from "../support/deliveryRuntimeFixtures";

describe("deliveryRuntimeLanguage", () => {
  it("maps every durable team work state into stable operator labels and groups", () => {
    const states: TeamWorkItemState[] = [
      "new",
      "briefed",
      "queued",
      "running",
      "reviewing",
      "paused",
      "output_ready",
      "degraded",
      "needs_operator",
      "archived",
    ];

    expect(states.map((state) => [state, teamWorkStateLabel(state), teamWorkStateGroup(state)])).toEqual([
      ["new", "Ready to brief", "not_started"],
      ["briefed", "Ready to start", "not_started"],
      ["queued", "Queued", "running"],
      ["running", "In progress", "running"],
      ["reviewing", "In review", "needs_review"],
      ["paused", "Paused", "running"],
      ["output_ready", "Output ready", "output_ready"],
      ["degraded", "Needs recovery", "needs_recovery"],
      ["needs_operator", "Needs response", "needs_review"],
      ["archived", "Archived", "archived"],
    ]);
  });

  it("keeps output package and receipt action labels in one contract", () => {
    expect(OUTPUT_PACKAGE_ACTION_LABELS).toMatchObject({
      openFile: "Open file",
      openFolder: "Open folder",
      openInResources: "Open in Resources",
      viewProof: "View proof",
    });
    expect(RUN_RECEIPT_SECTION_LABELS).toMatchObject({
      outcome: "Outcome",
      output: "Output",
      trust: "Trust",
      proof: "Proof",
      recovery: "Recovery",
    });
    const packageOutput = buildTeamOutputRef();
    expect(outputPackageKindLabel(packageOutput.kind, packageOutput.entrypoint)).toBe("Project package");
    expect(outputPackageKindLabel("image")).toBe("Media output");
    expect(outputPackageKindLabel("document")).toBe("Document");
    expect(outputPackageKindLabel("code")).toBe("File");
  });

  it("keeps output package folder actions readable across transient states", () => {
    expect(outputFolderButtonLabel("idle", "Open folder")).toBe("Open folder");
    expect(outputFolderButtonLabel("opening", "Open folder")).toBe("Opening...");
    expect(outputFolderButtonLabel("opened", "Open folder")).toBe("Folder opened");
    expect(outputFolderButtonLabel("failed", "Open folder")).toBe("Open failed");
  });

  it("turns local media dependency failures into the standard trust package", () => {
    const recovery = localMediaDependencyRecovery({
      ...buildMediaDegradation(),
    });

    expect(recovery).toEqual({
      failed: "Local media generation is not reachable, so Soma could not create the image output.",
      trusted: "The approval, request, failed run record, and audit trail remain available for review.",
      invalid: "No completed image output or execution proof should be trusted for this attempt.",
      recovery: "Start or reconnect the configured ComfyUI upstream, then retry. If you only need text/files, ask Soma to rerun without image generation.",
    });
    expect(recovery ? recoveryTrustLines(recovery) : []).toEqual([
      "Local media generation is not reachable, so Soma could not create the image output.",
      "Still available: The approval, request, failed run record, and audit trail remain available for review.",
      "Not reliable: No completed image output or execution proof should be trusted for this attempt.",
      "Safe next: Start or reconnect the configured ComfyUI upstream, then retry. If you only need text/files, ask Soma to rerun without image generation.",
    ]);
  });

  it("exposes compact review copy for failed and degraded team work", () => {
    expect(STALE_FAILED_PLAN_REVIEW_COPY.title).toBe("Old proposal cannot run");
    expect(STALE_FAILED_PLAN_REVIEW_COPY.nextAction).toContain("Start a new Soma ask");
    expect(DEGRADED_TEAM_WORK_REVIEW_COPY.title).toBe("Team work needs recovery");
    expect(DEGRADED_TEAM_WORK_REVIEW_COPY.trustedState).toContain("Not trusted: unfinished output");
    expect(buildTeamWorkItem({ state: "degraded" }).state).toBe("degraded");
  });
});
