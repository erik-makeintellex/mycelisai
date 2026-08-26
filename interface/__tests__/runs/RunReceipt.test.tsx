import { describe, expect, it } from "vitest";
import { fireEvent, render, screen } from "@testing-library/react";
import type { MissionEvent } from "@/store/useCortexStore";
import RunReceipt, { buildRunReceipt } from "@/components/runs/RunReceipt";

const now = new Date().toISOString();

function event(event_type: string, payload?: Record<string, unknown>): MissionEvent {
  return {
    id: `${event_type}-${Math.random()}`,
    run_id: "run-abc",
    tenant_id: "default",
    event_type,
    severity: event_type.includes("failed") ? "error" : "info",
    emitted_at: now,
    payload,
  };
}

describe("RunReceipt model", () => {
  it("summarizes completed runs with output and proof references", () => {
    const receipt = buildRunReceipt(
      [
        event("mission.started", { mission_id: "mission-1" }),
        event("artifact.created", { path: "workspace/generated/report.md", artifact_id: "artifact-1" }),
        event("proof.created", { proof_id: "proof-1", audit_event_id: "audit-1" }),
        event("mission.completed", { operator_summary: "Run completed with retained output." }),
      ],
      "run-abc",
    );

    expect(receipt.status).toBe("completed");
    expect(receipt.headline).toBe("Run completed");
    expect(receipt.result).toBe("Run completed with retained output.");
    expect(receipt.outputRefs).toContain("workspace/generated/report.md");
    expect(receipt.proofRefs).toContain("proof-1");
  });

  it("keeps failure evidence trusted while marking completed output proof unreliable", () => {
    const receipt = buildRunReceipt(
      [
        event("mission.started"),
        event("tool.failed", { error: "Planner validation provider timed out." }),
        event("mission.failed", { error: "Mission stopped after retry budget was exhausted.", audit_event_id: "audit-1" }),
      ],
      "run-abc",
    );

    expect(receipt.status).toBe("failed");
    expect(receipt.headline).toBe("Run needs recovery");
    expect(receipt.failure).toBe("Planner validation provider timed out.");
    expect(receipt.trust).toMatch(/failure evidence remain trusted/i);
    expect(receipt.next).toMatch(/retry from Soma/i);
    expect(receipt.proofRefs).toContain("audit-1");
  });

  it("reconstructs approved execution, output, and lifecycle contracts", () => {
    const receipt = buildRunReceipt(
      [
        event("team_work.status", {
          execution_mode: "team_async",
          work_intent: {
            kind: "project",
            objective: "Deliver a reviewable application.",
            output_contract: {
              shape: "app_package",
              primary_deliverable: "generated/app/index.html",
              retention: "user_deliverable",
              launch_hint: "Open index.html in a browser.",
              validation: ["launches", "supports keyboard input"],
            },
            lifecycle: {
              stop_action: "pause_project",
              retry_action: "resume_project",
              recovery_action: "restore_checkpoint",
              control_summary: "Pause, resume, or restore retained work.",
            },
          },
          output_refs: [{ storage_ref: "workspace/generated/app", entrypoint: "index.html" }],
        }),
        event("mission.completed", { operator_summary: "Application package retained.", execution_mode: "team_async" }),
      ],
      "run-abc",
    );

    expect(receipt.status).toBe("completed");
    expect(receipt.outputRefs).toContain("workspace/generated/app");
    expect(receipt.approvedWork).toMatchObject({
      executionMode: "team_async",
      kind: "project",
      outputShape: "app_package",
      primaryDeliverable: "generated/app/index.html",
      retention: "user_deliverable",
      stopAction: "pause_project",
      retryAction: "resume_project",
      recoveryAction: "restore_checkpoint",
    });
  });

  it("degrades completed work when its approved deliverable was not retained", () => {
    const receipt = buildRunReceipt(
      [
        event("mission.started", {
          execution_mode: "confirm_then_execute",
          work_intent: {
            kind: "one_shot",
            output_contract: {
              shape: "document",
              primary_deliverable: "review.md",
              retention: "user_deliverable",
            },
          },
        }),
        event("mission.completed", { operator_summary: "Work finished." }),
      ],
      "run-abc",
    );

    expect(receipt.status).toBe("degraded");
    expect(receipt.headline).toBe("Run needs output recovery");
    expect(receipt.result).toMatch(/approved deliverable was not retained/i);
    expect(receipt.trust).toMatch(/not ready to rely on/i);
    expect(receipt.next).toMatch(/retain the approved output/i);
  });

  it("opens a retained candidate directly while clearly marking failed validation", () => {
    const events = [event("team_work.status", {
      state: "degraded",
      headline: "Deliverable needs repair",
      next_action: "Ask Soma to repair the primary interaction.",
      output_refs: [{
        label: "Generated project package",
        storage_ref: "groups/delivery-team/generated/package",
        entrypoint: "index.html",
      }],
    })];
    const receipt = buildRunReceipt(events, "run-abc");
    expect(receipt.status).toBe("degraded");
    expect(receipt.outputLinks).toEqual([{
      label: "Generated project package",
      href: "/api/v1/workspace/files/view?path=groups%2Fdelivery-team%2Fgenerated%2Fpackage%2Findex.html",
    }]);

    render(<RunReceipt runId="run-abc" events={events} />);
    expect(screen.getByText("Output needs repair")).toBeDefined();
    expect(screen.getByText(/validation has not passed/i)).toBeDefined();
    expect(screen.getByRole("link", { name: /Open unverified generated project package/i }).getAttribute("href")).toBe(
      "/api/v1/workspace/files/view?path=groups%2Fdelivery-team%2Fgenerated%2Fpackage%2Findex.html",
    );
  });

  it("renders failed runs as recoverable work with trust and inspect evidence", () => {
    render(
      <RunReceipt
        runId="run-abc"
        events={[
          event("mission.started"),
          event("tool.failed", { error: "Planner validation provider timed out." }),
          event("mission.failed", { error: "Mission stopped after retry budget was exhausted.", audit_event_id: "audit-1" }),
        ]}
      />,
    );

    expect(screen.getByText("Run needs recovery")).toBeDefined();
    expect(screen.getByLabelText("Outcome health: Blocked")).toBeDefined();
    expect(screen.getByText("What happened")).toBeDefined();
    expect(screen.getByText("What to trust")).toBeDefined();
    expect(screen.getByText("Next step")).toBeDefined();
    expect(screen.getByText(/failure evidence remain trusted/i)).toBeDefined();
    expect(screen.getByText(/Completed output proof is not reliable/i)).toBeDefined();
    expect(screen.getByText(/retry from Soma or the owning workflow/i)).toBeDefined();
    expect(screen.getByText("No retained output yet")).toBeDefined();
    expect(screen.getByText("1 proof ref")).toBeDefined();
    expect(screen.getByText("Recovery needed")).toBeDefined();

    fireEvent.click(screen.getByRole("button", { name: /Inspect receipt evidence/i }));

    expect(screen.getByText("run-abc")).toBeDefined();
    expect(screen.getByText("audit-1")).toBeDefined();
  });

  it("renders reconstructed approved work only after the operator opens Inspect", () => {
    render(
      <RunReceipt
        runId="run-abc"
        events={[
          event("team_work.status", {
            execution_mode: "team_async",
            work_intent: {
              kind: "project",
              output_contract: { primary_deliverable: "generated/app/index.html", retention: "user_deliverable" },
              lifecycle: { control_summary: "Pause, resume, or restore retained work." },
            },
            output_refs: [{ storage_ref: "workspace/generated/app" }],
          }),
          event("mission.completed"),
        ]}
      />,
    );

    expect(screen.queryByText("Approved work")).toBeNull();
    fireEvent.click(screen.getByRole("button", { name: /Inspect receipt evidence/i }));
    expect(screen.getByText("Approved work")).toBeDefined();
    expect(screen.getByText("project · team_async")).toBeDefined();
    expect(screen.getByText("generated/app/index.html")).toBeDefined();
    expect(screen.getByText("Pause, resume, or restore retained work.")).toBeDefined();
  });
});
