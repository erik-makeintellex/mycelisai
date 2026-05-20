import { describe, expect, it, vi } from "vitest";
import { fireEvent, render, screen, waitFor } from "@testing-library/react";
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

  it("shows durable run proof and retained output links", () => {
    render(
      <ActiveWorkLane
        items={[
          {
            ...baseItem,
            runId: "run-1",
            outputRefs: [
              {
                output_id: "out-1",
                team_id: "team-alpha",
                work_item_id: "work-1",
                kind: "file",
                label: "Launch brief",
                storage_ref: "generated/launch/brief.md",
              },
            ],
            proofRefs: ["proof-1"],
            auditRefs: ["audit-1"],
            interactions: [],
          },
        ]}
      />,
    );

    expect(screen.getByRole("link", { name: /Run proof/i }).getAttribute("href")).toBe("/runs/run-1");
    expect(screen.getByRole("link", { name: /Launch brief/i }).getAttribute("href")).toBe("/api/v1/workspace/files/view?path=generated%2Flaunch%2Fbrief.md");
    expect(screen.getByRole("link", { name: /Proof proof-1/i }).getAttribute("href")).toBe("/runs/run-1");
    expect(screen.getByText(/Audit audit-1/i)).toBeDefined();
  });

  it("submits a bounded ask for durable active work", async () => {
    const onTeamAsk = vi.fn().mockResolvedValue(undefined);
    render(
      <ActiveWorkLane
        items={[
          {
            ...baseItem,
            source: "durable",
            interactions: [],
          },
        ]}
        onTeamAsk={onTeamAsk}
      />,
    );

    fireEvent.click(screen.getByRole("button", { name: /ask team/i }));
    const input = screen.getByLabelText(/Ask Draft launch brief/i) as HTMLInputElement;
    fireEvent.change(input, {
      target: { value: "Create the next validation note" },
    });
    fireEvent.click(screen.getByRole("button", { name: /send/i }));

    await waitFor(() => {
      expect(onTeamAsk).toHaveBeenCalledWith(
        expect.objectContaining({ id: "work-1" }),
        "Create the next validation note",
      );
    });
    expect(screen.queryByLabelText(/Ask Draft launch brief/i)).toBeNull();
  });
});
