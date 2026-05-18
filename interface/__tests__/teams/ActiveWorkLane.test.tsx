import { describe, expect, it, vi } from "vitest";
import { fireEvent, render, screen } from "@testing-library/react";
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

describe("ActiveWorkLane", () => {
  it("disables non-link controls when no action handler is wired", () => {
    render(
      <ActiveWorkLane
        items={[
          {
            ...baseItem,
            interactions: [
              { action: "inspect", label: "Inspect", href: "/runs/run-1" },
              { action: "pause", label: "Pause" },
            ],
          },
        ]}
      />,
    );

    expect(screen.getByRole("link", { name: /inspect/i }).getAttribute("href")).toBe("/runs/run-1");
    const pause = screen.getByRole("button", { name: /pause/i }) as HTMLButtonElement;
    expect(pause.disabled).toBe(true);
    expect(pause.getAttribute("title")).toContain("action API is not connected yet");
  });

  it("keeps non-link controls executable when a handler is provided", () => {
    const onAction = vi.fn();
    render(
      <ActiveWorkLane
        items={[
          {
            ...baseItem,
            interactions: [{ action: "pause", label: "Pause" }],
          },
        ]}
        onAction={onAction}
      />,
    );

    const pause = screen.getByRole("button", { name: /pause/i }) as HTMLButtonElement;
    expect(pause.disabled).toBe(false);
    fireEvent.click(pause);
    expect(onAction).toHaveBeenCalledWith(
      expect.objectContaining({ id: "work-1" }),
      expect.objectContaining({ action: "pause" }),
    );
  });
});
