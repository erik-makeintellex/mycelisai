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
                kind: "project_package",
                label: "Launch brief",
                storage_ref: "generated/launch/brief.md",
                entrypoint: "generated/launch/index.html",
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
    expect(screen.getByText("Running, output may still change")).toBeDefined();
    expect(screen.getByText("1 package retained")).toBeDefined();
    expect(screen.getByText("Proof available")).toBeDefined();
  });

  it("opens retained project-package entrypoints and proof carried on output refs", () => {
    render(
      <ActiveWorkLane
        items={[
          {
            ...baseItem,
            outputRefs: [
              {
                output_id: "out-package",
                team_id: "team-alpha",
                work_item_id: "work-1",
                kind: "project_package",
                label: "Playable package",
                storage_ref: "workspace/generated/playable",
                entrypoint: "index.html",
                proof_id: "proof-package-1",
              },
            ],
            interactions: [],
          },
        ]}
      />,
    );

    expect(
      screen.getByRole("link", { name: /Playable package/i }).getAttribute("href"),
    ).toBe(
      "/api/v1/workspace/files/view?path=workspace%2Fgenerated%2Fplayable%2Findex.html",
    );
    expect(screen.getByText("1 package retained")).toBeDefined();
    expect(screen.getByText("Proof available")).toBeDefined();
    expect(screen.getByText("Proof proof-pa...")).toBeDefined();
  });

  it("joins retained package folders with relative nested entrypoints", () => {
    render(
      <ActiveWorkLane
        items={[
          {
            ...baseItem,
            outputRefs: [
              {
                output_id: "out-nested-package",
                team_id: "team-alpha",
                work_item_id: "work-1",
                kind: "project_package",
                label: "Nested playable package",
                storage_ref: "workspace/generated/nested",
                entrypoint: "dist/index.html",
                proof_ref: "proof-nested-1",
              },
            ],
            interactions: [],
          },
        ]}
      />,
    );

    expect(
      screen
        .getByRole("link", { name: /Nested playable package/i })
        .getAttribute("href"),
    ).toBe(
      "/api/v1/workspace/files/view?path=workspace%2Fgenerated%2Fnested%2Fdist%2Findex.html",
    );
    expect(screen.getByText("Proof available")).toBeDefined();
    expect(screen.getByText("Proof proof-ne...")).toBeDefined();
  });

  it("summarizes degraded work with recovery and proof gaps", () => {
    render(
      <ActiveWorkLane
        items={[
          {
            ...baseItem,
            state: "degraded",
            fallbackReason: "Team response timed out before output was retained.",
            recoveryOptions: ["Retry with retained context"],
            interactions: [{ action: "recover", label: "Recover" }],
          },
        ]}
        onAction={vi.fn()}
      />,
    );

    expect(screen.getByText("Needs recovery")).toBeDefined();
    expect(screen.getByText("No retained output yet")).toBeDefined();
    expect(screen.getByText("Proof pending")).toBeDefined();
    expect(screen.getByText("Recovery: Retry with retained context")).toBeDefined();
  });

  it("keeps the default lane compact when a visible item cap is provided", () => {
    const items = Array.from({ length: 5 }, (_, index) => ({
      ...baseItem,
      id: `work-${index + 1}`,
      title: `Work item ${index + 1}`,
    }));

    render(
      <ActiveWorkLane
        items={items}
        maxVisibleItems={2}
        totalItemCount={5}
        moreItemsHref="/teams"
      />,
    );

    expect(screen.getByText("5 items")).toBeDefined();
    expect(screen.getByText("Work item 1")).toBeDefined();
    expect(screen.getByText("Work item 2")).toBeDefined();
    expect(screen.queryByText("Work item 3")).toBeNull();
    expect(screen.getByRole("link", { name: /3 more work items in Teams/i }).getAttribute("href")).toBe("/teams");
  });

  it("makes hidden evidence count visible when refs are capped", () => {
    render(
      <ActiveWorkLane
        items={[{
          ...baseItem,
          outputRefs: Array.from({ length: 4 }, (_, index) => ({
            output_id: `out-${index + 1}`,
            team_id: "team-alpha",
            work_item_id: "work-1",
            kind: "file",
            label: `Output ${index + 1}`,
            storage_ref: `generated/output-${index + 1}.md`,
          })),
          proofRefs: ["proof-1", "proof-2", "proof-3", "proof-4"],
          auditRefs: ["audit-1", "audit-2", "audit-3"],
        }]}
      />,
    );

    expect(screen.getByText("+3 more in inspect")).toBeDefined();
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
    fireEvent.click(screen.getByRole("button", { name: /queue ask/i }));

    await waitFor(() => {
      expect(onTeamAsk).toHaveBeenCalledWith(
        expect.objectContaining({ id: "work-1" }),
        "Create the next validation note",
      );
    });
    expect(screen.queryByLabelText(/Ask Draft launch brief/i)).toBeNull();
  });
});
