import { render, screen } from "@testing-library/react";
import { describe, expect, it, vi } from "vitest";
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

describe("ActiveWorkLane compact review", () => {
  it("keeps the dashboard review version readable for degraded work", () => {
    render(
      <ActiveWorkLane
        frame={false}
        purpose="review"
        statusLabel="Durable team-work state loaded."
        items={[{
          ...baseItem,
          state: "degraded",
          source: "durable",
          sourceLabel: "Durable team work",
          scopeLabel: "Delegated work",
          fallbackReason: "Team ask degraded: context deadline exceeded.",
          recoveryOptions: ["Retry with retained context"],
          interactions: [
            { action: "inspect", label: "Open run", href: "/runs/run-1" },
            { action: "steer", label: "Respond" },
            { action: "start_work", label: "Start task", disabled: true },
            { action: "pause", label: "Pause", disabled: true },
            { action: "recover", label: "Retry recovery" },
            { action: "archive", label: "Clear from review" },
          ],
        }]}
        onAction={vi.fn()}
        onTeamAsk={vi.fn()}
      />,
    );

    expect(screen.getByText("Team work needs recovery")).toBeDefined();
    expect(screen.getByText(/The team did not finish this work/i)).toBeDefined();
    expect(screen.getByLabelText("Review queue summary")).toBeDefined();
    expect(screen.getByLabelText("Needs decision: 1")).toBeDefined();
    expect(screen.getByText(/No retained output yet/)).toBeDefined();
    expect(screen.queryByText("Durable team-work state loaded.")).toBeNull();
    expect(screen.queryByText("Durable team work")).toBeNull();
    expect(screen.queryByRole("button", { name: /Start task/i })).toBeNull();
    expect(screen.queryByRole("button", { name: /Pause/i })).toBeNull();
    expect(screen.queryByRole("button", { name: /^Respond$/i })).toBeNull();
    expect(screen.queryByRole("button", { name: /Ask team/i })).toBeNull();
    expect(screen.getByRole("button", { name: /Retry recovery/i })).toBeDefined();
    expect(screen.getByRole("link", { name: /Open run/i })).toBeDefined();
    expect(screen.getByRole("button", { name: /Clear from review/i })).toBeDefined();
  });

  it("summarizes queued recovery plainly in the dashboard review version", () => {
    render(
      <ActiveWorkLane
        frame={false}
        purpose="review"
        items={[{
          ...baseItem,
          state: "queued",
          title: "Playwright bounded team ask proof 9f4773a9-bbe0-4d63-b403-32728c8da8ad",
          description: "Recovery requested. Next: Watch for new status, retained output, or proof.",
          nextAction: "Watch for new status, retained output, or proof.",
          interactions: [
            { action: "inspect", label: "Open run", href: "/runs/run-1" },
            { action: "archive", label: "Clear from review" },
          ],
        }]}
        onAction={vi.fn()}
        onTeamAsk={vi.fn()}
      />,
    );

    expect(screen.getByText("Recovery request queued")).toBeDefined();
    expect(screen.getByText(/Soma has queued a recovery attempt/i)).toBeDefined();
    expect(screen.getAllByText(/Watch for new status, retained output, or proof/i)).toHaveLength(1);
    expect(screen.queryByText(/Playwright bounded team ask proof/i)).toBeNull();
    expect(screen.queryByRole("button", { name: /Ask team/i })).toBeNull();
  });
});
