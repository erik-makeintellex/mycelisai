import { describe, expect, it, vi } from "vitest";
import { fireEvent, render, screen, within } from "@testing-library/react";
import { ActiveWorkLane } from "@/components/teams/ActiveWorkLane";
import type { TeamWorkItem } from "@/store/useCortexStore";

const baseItem: TeamWorkItem = {
  id: "work-1",
  title: "Draft launch brief",
  state: "running",
  ownerLabel: "Alpha lead",
  scopeLabel: "Deliverable work",
  teamIds: ["team-alpha"],
  interactions: [],
};

describe("ActiveWorkLane review mode", () => {
  it("summarizes review work and uses concise decision labels", () => {
    render(
      <ActiveWorkLane
        purpose="review"
        items={[
          {
            ...baseItem,
            state: "degraded",
            interactions: [{ action: "archive", label: "Clear from review" }],
          },
          {
            ...baseItem,
            id: "work-2",
            title: "Package ready",
            state: "output_ready",
            interactions: [{ action: "archive", label: "Clear from review" }],
          },
          {
            ...baseItem,
            id: "work-3",
            title: "Deployment proof",
            state: "running",
            interactions: [],
          },
        ]}
        onAction={vi.fn()}
      />,
    );

    const summary = screen.getByLabelText("Review queue summary");
    expect(within(summary).getByLabelText("Needs decision: 1")).toBeDefined();
    expect(within(summary).getByLabelText("Ready output: 1")).toBeDefined();
    expect(within(summary).getByLabelText("Still working: 1")).toBeDefined();
    expect(within(summary).getByLabelText("Can clear: 2")).toBeDefined();
    expect(screen.getByTestId("work-review-inbox")).toBeDefined();
    expect(screen.getByRole("list", { name: "Review work items" })).toBeDefined();
    expect(screen.getByLabelText("Review details for Draft launch brief")).toBeDefined();
    expect(screen.getAllByText("Reason").length).toBeGreaterThan(0);
    expect(screen.getAllByText("Trust").length).toBeGreaterThan(0);
    expect(screen.getAllByText("Move").length).toBeGreaterThan(0);
    expect(screen.getByText("Other available actions")).toBeDefined();
    expect(screen.queryByText("What is safe to rely on")).toBeNull();
  });

  it("keeps recovery items in one failed/trusted/safe-next review grammar", () => {
    render(
      <ActiveWorkLane
        purpose="review"
        items={[
          {
            ...baseItem,
            state: "degraded",
            title: "Recover failed release notes",
            recoveryOptions: ["Retry with retained proof"],
            interactions: [{ action: "recover", label: "Recover" }],
          },
          {
            ...baseItem,
            id: "work-stale",
            state: "degraded",
            title: "Playwright bounded team ask proof",
            recoveryOptions: [
              "Review the failed run, adjust the request or runtime dependency, then retry the proposal.",
            ],
            nextAction: "Clear this stale test item from review or inspect diagnostics before retrying.",
            description: "no approved execution plan was stored for this proposal",
            interactions: [
              { action: "archive", label: "Clear from review" },
              { action: "inspect", label: "Inspect", href: "/runs/run-stale" },
            ],
          },
          {
            ...baseItem,
            id: "work-operator",
            state: "needs_operator",
            title: "Team needs launch direction",
            interactions: [{ action: "steer", label: "Respond" }],
          },
        ]}
        onAction={vi.fn()}
      />,
    );

    const summary = screen.getByLabelText("Review queue summary");
    expect(within(summary).getByLabelText("Needs decision: 3")).toBeDefined();
    expect(screen.getAllByText("Team work needs recovery").length).toBeGreaterThan(0);
    expect(
      screen.getByText("Trusted: retained context, proof refs, and status history. Not trusted: unfinished output."),
    ).toBeDefined();
    expect(screen.getByText(/Recover when the runtime dependency is available/i)).toBeDefined();
    expect(screen.getAllByRole("button", { name: /Recover/i }).length).toBeGreaterThan(0);

    fireEvent.click(screen.getAllByText("Old proposal cannot run")[0]);

    expect(screen.getByText(/Trusted: the failure record and audit trail/i)).toBeDefined();
    expect(screen.getByText(/Not trusted: any implied output from this attempt/i)).toBeDefined();
    expect(screen.getByText(/Clear this from review. Nothing ran/i)).toBeDefined();
    expect(screen.getAllByRole("button", { name: /Clear from review/i }).length).toBeGreaterThan(0);

    fireEvent.click(screen.getAllByText("Team needs your response")[0]);

    expect(screen.getAllByText(/The team is waiting for missing direction/i).length).toBeGreaterThan(0);
    expect(screen.getByText(/Respond or steer the work with the missing decision/i)).toBeDefined();
    expect(screen.getAllByRole("button", { name: /Respond/i }).length).toBeGreaterThan(0);
  });
});
