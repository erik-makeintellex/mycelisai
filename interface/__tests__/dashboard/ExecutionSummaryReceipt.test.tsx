import { fireEvent, render, screen } from "@testing-library/react";
import { describe, expect, it, vi } from "vitest";
import ExecutionSummaryReceipt from "@/components/soma/ExecutionSummaryReceipt";
import { OUTPUT_CONTINUATION_EVENT } from "@/components/soma/outputContinuation";
import type { ExecutionSummaryData } from "@/store/useCortexStore";

describe("ExecutionSummaryReceipt", () => {
  it("keeps an approval run pending until the team result is explicitly verified", () => {
    const summary: ExecutionSummaryData = {
      execution_mode: "team_async",
      execution: {
        status: "completed",
        summary: "The approved work was handed to the team.",
      },
      proof: [{ run_id: "run-pending-1" }],
    };

    render(<ExecutionSummaryReceipt summary={summary} runId="run-pending-1" />);

    expect(screen.getByText("Work started")).toBeDefined();
    expect(screen.getByText(/after the team returns a verified result/i)).toBeDefined();
    expect(screen.queryByText("Result verified")).toBeNull();
    expect(screen.queryByText("Result saved")).toBeNull();
  });

  it("reveals generated app packages with launch, resources, and reply paths", () => {
    const continuation = vi.fn();
    window.addEventListener(OUTPUT_CONTINUATION_EVENT, continuation);
    const summary: ExecutionSummaryData = {
      execution: {
        shape: "team_execution",
        status: "verified",
        summary: "Generated a playable browser app package.",
      },
      outputs: [{
        kind: "project_package",
        title: "Coin Runner Game",
        folder: "workspace/generated/coin-runner",
        entrypoint: "index.html",
        validation: "Browser opened and score increased after click.",
        proof_artifact_id: "proof-game-1",
      }],
      proof: [{ run_id: "run-game-1" }],
    };

    render(<ExecutionSummaryReceipt summary={summary} runId="run-game-1" />);

    expect(screen.getByText(/App\/package output is ready/i)).toBeDefined();
    expect(screen.getByText("App/package:")).toBeDefined();
    expect(screen.getByText("Coin Runner Game")).toBeDefined();
    expect(screen.getByText("Browser opened and score increased after click.")).toBeDefined();
    expect(screen.getByRole("link", { name: /Open Coin Runner Game in Resources/i }).getAttribute("href"))
      .toBe("/resources?tab=workspace&path=workspace%2Fgenerated%2Fcoin-runner");
    expect(screen.getByRole("button", { name: /Open app Coin Runner Game/i })).toBeDefined();

    fireEvent.click(screen.getByRole("button", { name: /Reply to Coin Runner Game in Soma/i }));

    expect(continuation).toHaveBeenCalled();
    expect(continuation.mock.calls[0][0].detail).toMatchObject({
      title: "Coin Runner Game",
      reference: "workspace/generated/coin-runner",
      proof: "proof-game-1",
    });
    window.removeEventListener(OUTPUT_CONTINUATION_EVENT, continuation);
  });

  it("lets file deliverables continue back into Soma from the compact receipt", () => {
    const continuation = vi.fn();
    window.addEventListener(OUTPUT_CONTINUATION_EVENT, continuation);
    const summary: ExecutionSummaryData = {
      execution: {
        shape: "tool_execution",
        status: "verified",
        summary: "Created a retained launch brief.",
      },
      outputs: [{
        kind: "file",
        title: "Launch brief",
        url: "/api/v1/workspace/files/view?path=workspace%2Fgenerated%2Flaunch%2Fbrief.md",
        path: "workspace/generated/launch/brief.md",
        proof_artifact_id: "proof-brief-1",
      }],
      proof: [{ run_id: "run-brief-1" }],
    };

    render(<ExecutionSummaryReceipt summary={summary} runId="run-brief-1" />);

    fireEvent.click(screen.getByRole("button", { name: /Reply to Launch brief in Soma/i }));

    expect(continuation).toHaveBeenCalled();
    expect(continuation.mock.calls[0][0].detail).toMatchObject({
      title: "Launch brief",
      reference: "workspace/generated/launch/brief.md",
      proof: "proof-brief-1",
    });
    window.removeEventListener(OUTPUT_CONTINUATION_EVENT, continuation);
  });

  it("keeps technical run inspection secondary when no deliverable exists yet", () => {
    const continuation = vi.fn();
    window.addEventListener(OUTPUT_CONTINUATION_EVENT, continuation);
    const summary: ExecutionSummaryData = {
      execution: {
        shape: "team_execution",
        status: "verified",
        summary: "Team work started.",
      },
      proof: [{ run_id: "run-work-1" }],
    };

    render(<ExecutionSummaryReceipt summary={summary} runId="run-work-1" />);

    expect(screen.queryByRole("link", { name: /^Run$/i })).toBeNull();
    fireEvent.click(screen.getByRole("button", { name: /Continue with Soma/i }));
    expect(continuation).toHaveBeenCalled();
    expect(continuation.mock.calls[0][0].detail).toMatchObject({
      reference: "run:run-work-1",
      proof: "run-work-1",
    });

    fireEvent.click(screen.getByText("Proof and execution details"));
    expect(screen.getByRole("link", { name: /Inspect run receipt/i }).getAttribute("href"))
      .toBe("/runs/run-work-1");
    window.removeEventListener(OUTPUT_CONTINUATION_EVENT, continuation);
  });

  it("does not present team planning files as completed user output", () => {
    const summary: ExecutionSummaryData = {
      execution: {
        shape: "team_execution",
        status: "verified",
        summary: "Created the team and queued delivery work.",
      },
      outputs: [{
        kind: "file",
        output_class: "planning",
        title: "Team evocation brief",
        path: "groups/game-team/planning/TEAM_EVOCATION.md",
        url: "/api/v1/workspace/files/view?path=groups%2Fgame-team%2Fplanning%2FTEAM_EVOCATION.md",
        retained: true,
      }],
      proof: [{ run_id: "run-planning-1" }],
    };

    render(<ExecutionSummaryReceipt
      summary={summary}
      runId="run-planning-1"
      artifacts={[{
        id: "planning-artifact",
        type: "file",
        title: "Team evocation brief",
        output_class: "planning",
        saved_path: "groups/game-team/planning/TEAM_EVOCATION.md",
      }]}
    />);

    expect(screen.getByText("Work started")).toBeDefined();
    expect(screen.getByText(/deliverable will appear here only after/i)).toBeDefined();
    expect(screen.queryByText("Team evocation brief")).toBeNull();
  });
});
