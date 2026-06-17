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
    expect(screen.getAllByText("failed").length).toBeGreaterThan(0);
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
});
