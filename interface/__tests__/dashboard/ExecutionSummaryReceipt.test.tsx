import { fireEvent, render, screen } from "@testing-library/react";
import { describe, expect, it, vi } from "vitest";
import ExecutionSummaryReceipt from "@/components/soma/ExecutionSummaryReceipt";
import { OUTPUT_CONTINUATION_EVENT } from "@/components/soma/outputContinuation";
import type { ExecutionSummaryData } from "@/store/useCortexStore";

describe("ExecutionSummaryReceipt", () => {
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
});
