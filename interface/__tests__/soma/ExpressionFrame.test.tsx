import { fireEvent, render, screen } from "@testing-library/react";
import { describe, expect, it, vi } from "vitest";
import { ExpressionFrame, type ExpressionFrameKind } from "@/components/soma/ExpressionFrame";

const frameKinds: ExpressionFrameKind[] = [
  "direct_answer",
  "proposal",
  "active_work",
  "output_ready",
  "proof",
  "degraded",
  "blocked",
  "recovery",
];

describe("ExpressionFrame", () => {
  it("renders every canonical ExpressionFrame kind", () => {
    for (const kind of frameKinds) {
      const { unmount } = render(
        <ExpressionFrame kind={kind} intent={`Intent for ${kind}`} />,
      );

      expect(screen.getByText(`Intent for ${kind}`)).toBeDefined();
      expect(screen.getAllByText(kind.replace(/_/g, " "), { exact: false }).length).toBeGreaterThan(0);
      unmount();
    }
  });

  it("renders intent, state, next action, approval, outputs, proof, recovery, and inspect details", () => {
    render(
      <ExpressionFrame
        kind="proposal"
        title="Governed media proposal"
        intent="Create a retained launch reel."
        state="Waiting for operator approval before execution."
        nextAction="Approve or revise the proposal."
        risk="Uses external media model quota."
        approval="Approval required"
        outputs={[{ label: "Storyboard", href: "/api/v1/workspace/files/view?path=storyboard.md" }]}
        proof={[{ label: "Intent proof", href: "/runs/run-1" }]}
        recovery="Revise scope or retry with a smaller output."
        inspect={[
          { label: "Run", value: "run-1" },
          { label: "Team", value: "media-team" },
        ]}
      />,
    );

    expect(screen.getByText("Governed media proposal")).toBeDefined();
    expect(screen.getByText("Create a retained launch reel.")).toBeDefined();
    expect(screen.getByText("Waiting for operator approval before execution.")).toBeDefined();
    expect(screen.getByText("Approve or revise the proposal.")).toBeDefined();
    expect(screen.getByText("Uses external media model quota.")).toBeDefined();
    expect(screen.getByText("Approval required")).toBeDefined();
    expect(screen.getByRole("link", { name: /Storyboard/i }).getAttribute("href")).toBe("/api/v1/workspace/files/view?path=storyboard.md");
    expect(screen.getByRole("link", { name: /Intent proof/i }).getAttribute("href")).toBe("/runs/run-1");
    expect(screen.getByText("Revise scope or retry with a smaller output.")).toBeDefined();

    fireEvent.click(screen.getByText("More details"));
    expect(screen.getByText("run-1")).toBeDefined();
    expect(screen.getByText("media-team")).toBeDefined();
  });

  it("supports linked, button, and disabled actions", () => {
    const onClick = vi.fn();
    render(
      <ExpressionFrame
        kind="active_work"
        intent="Continue the running task."
        primaryAction={{ label: "Inspect run", href: "/runs/run-2" }}
        secondaryAction={{ label: "Recover", onClick }}
      />,
    );

    expect(screen.getByRole("link", { name: /Inspect run/i }).getAttribute("href")).toBe("/runs/run-2");
    fireEvent.click(screen.getByRole("button", { name: /Recover/i }));
    expect(onClick).toHaveBeenCalledTimes(1);
  });
});
