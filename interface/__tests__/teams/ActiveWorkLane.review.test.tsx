import { describe, expect, it, vi } from "vitest";
import { render, screen, within } from "@testing-library/react";
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
});
