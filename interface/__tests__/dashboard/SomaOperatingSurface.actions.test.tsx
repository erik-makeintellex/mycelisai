import React from "react";
import { fireEvent, render, screen, within } from "@testing-library/react";
import { beforeEach, describe, expect, it, vi } from "vitest";
import { SomaOperatingSurface } from "@/components/soma/SomaOperatingSurface";

const mocks = vi.hoisted(() => {
  const selectTeam = vi.fn();
  const handleActiveWorkAction = vi.fn();
  const handleTeamAsk = vi.fn();
  const missionControlChat = vi.fn();
  const useDurableTeamWork = vi.fn();
  const useTeamWorkActionHandler = vi.fn();
  const storeState = {
    missionChat: [] as unknown[],
    teamsDetail: [{ id: "team-alpha", name: "Alpha" }],
    durableWorkRefreshVersion: 5,
    selectTeam,
    selectedTeamId: null as string | null,
  };
  return {
    selectTeam,
    handleActiveWorkAction,
    handleTeamAsk,
    missionControlChat,
    useDurableTeamWork,
    useTeamWorkActionHandler,
    storeState,
  };
});

vi.mock("@/store/useCortexStore", () => ({
  useCortexStore: (selector: (state: unknown) => unknown) => selector(mocks.storeState),
}));

vi.mock("@/components/teams/useTeamWorkActionHandler", () => ({
  useTeamWorkActionHandler: mocks.useTeamWorkActionHandler,
  mergeTeamWorkItems: (
    durableItems: Array<{ id: string }>,
    submittedItems: Array<{ id: string }>,
  ) => [...submittedItems, ...durableItems],
}));

vi.mock("@/components/soma/useDurableTeamWork", () => ({
  useDurableTeamWork: mocks.useDurableTeamWork,
}));

vi.mock("@/components/teams/ActiveWorkLane", () => ({
  ActiveWorkLane: (props: {
    items: Array<{ id: string; title: string }>;
    statusLabel?: string | null;
    degradedMessage?: string | null;
    onAction?: (item: unknown, action: unknown) => void;
    onTeamAsk?: (item: unknown, message: string) => void;
  }) => (
    <section data-testid="active-work-lane">
      <p>{props.statusLabel}</p>
      <p>{props.degradedMessage}</p>
      <button
        type="button"
        onClick={() => props.onAction?.(props.items[0], { action: "pause" })}
      >
        Pause work
      </button>
      <button
        type="button"
        onClick={() => props.onTeamAsk?.(props.items[0], "Continue the proof")}
      >
        Ask team
      </button>
    </section>
  ),
}));

vi.mock("@/components/dashboard/MissionControlChat", () => ({
  default: (props: { focusedTeamId?: string | null }) => {
    mocks.missionControlChat(props);
    return <div data-testid="mission-chat" />;
  },
}));

vi.mock("@/components/soma/SomaHeader", () => ({
  SomaHeader: () => <header data-testid="soma-header" />,
}));

vi.mock("@/components/soma/SomaWorkspaceFrame", () => ({
  SomaWorkspaceFrame: ({
    activeWork,
    context,
    expression,
    output,
    trust,
  }: {
    activeWork: React.ReactNode;
    context: React.ReactNode;
    expression: React.ReactNode;
    output: React.ReactNode;
    trust: React.ReactNode;
  }) => (
    <div>
      {expression}
      {activeWork}
      {trust}
      {output}
      {context}
    </div>
  ),
}));

describe("SomaOperatingSurface active work actions", () => {
  beforeEach(() => {
    vi.clearAllMocks();
    mocks.storeState.missionChat = [];
    mocks.storeState.selectedTeamId = null;
    mocks.useTeamWorkActionHandler.mockReturnValue({
      activeWorkRefreshVersion: 7,
      activeWorkActionError: "Team action needs operator attention.",
      activeWorkActionNotice: "Team ask queued. You can keep working.",
      submittedTeamWorkItems: [],
      handleActiveWorkAction: mocks.handleActiveWorkAction,
      handleTeamAsk: mocks.handleTeamAsk,
    });
    mocks.useDurableTeamWork.mockReturnValue({
      items: [{
        id: "work-1",
        title: "Browser game validation",
        state: "running",
        teamIds: ["team-alpha"],
        interactions: [],
      }],
      outputRefs: [],
      emptyMessage: "No active work.",
      status: "ready",
      statusLabel: "Ready",
      degradedMessage: null,
    });
  });

  it("wires Soma home active work controls to the team action handler", () => {
    render(<SomaOperatingSurface focusedTeamId="team-alpha" />);

    expect(mocks.useTeamWorkActionHandler).toHaveBeenCalledWith(mocks.selectTeam);
    expect(mocks.useDurableTeamWork).toHaveBeenCalledWith(expect.objectContaining({
      focusedTeamId: "team-alpha",
      refreshVersion: 12,
    }));
    expect(mocks.missionControlChat).toHaveBeenCalledWith(expect.objectContaining({
      focusedTeamId: "team-alpha",
    }));
    expect(screen.getByTestId("soma-context-focus-bar").textContent).toContain("Alpha");
    expect(screen.getByTestId("soma-team-context-switcher").textContent).toContain("Work contexts");
    expect(screen.getByRole("button", { name: /Alpha/i })).toBeDefined();
    expect(screen.getByText("Team action needs operator attention.")).toBeDefined();
    expect(screen.getByText("Team ask queued. You can keep working.")).toBeDefined();

    fireEvent.click(screen.getByRole("button", { name: /pause work/i }));
    fireEvent.click(screen.getByRole("button", { name: /ask team/i }));

    expect(mocks.handleActiveWorkAction).toHaveBeenCalledWith(
      expect.objectContaining({ id: "work-1" }),
      expect.objectContaining({ action: "pause" }),
    );
    expect(mocks.handleTeamAsk).toHaveBeenCalledWith(
      expect.objectContaining({ id: "work-1" }),
      "Continue the proof",
    );
  });

  it("lets operators switch the Soma surface into a team work context", () => {
    render(<SomaOperatingSurface />);

    fireEvent.click(screen.getByRole("button", { name: /Alpha/i }));

    expect(mocks.selectTeam).toHaveBeenCalledWith("team-alpha");
  });

  it("does not show standing team topology as work context on a clean root dashboard", () => {
    mocks.useDurableTeamWork.mockReturnValue({
      items: [],
      outputRefs: [],
      emptyMessage: "No active work.",
      status: "ready",
      statusLabel: "Ready",
      degradedMessage: null,
    });

    render(<SomaOperatingSurface />);

    expect(screen.queryByTestId("soma-context-focus-bar")).toBeNull();
    expect(screen.queryByTestId("soma-team-context-switcher")).toBeNull();
  });

  it("keeps work context switching visible when a team is focused", () => {
    mocks.useDurableTeamWork.mockReturnValue({
      items: [],
      outputRefs: [],
      emptyMessage: "No active work.",
      status: "ready",
      statusLabel: "Ready",
      degradedMessage: null,
    });

    render(<SomaOperatingSurface focusedTeamId="team-alpha" />);

    expect(screen.getByTestId("soma-context-focus-bar")).toBeDefined();
    expect(screen.getByTestId("soma-team-context-switcher")).toBeDefined();
  });

  it("shows focused team retained outputs before the work panel is opened", () => {
    mocks.useDurableTeamWork.mockReturnValue({
      items: [{
        id: "work-1",
        title: "Comic page generation",
        state: "output_ready",
        teamIds: ["team-alpha"],
        interactions: [],
      }],
      outputRefs: [{
        output_id: "comic-page-output-old",
        team_id: "team-alpha",
        work_item_id: "work-1",
        kind: "media",
        label: "Old comic page",
        storage_ref: "groups/team-alpha/media/old-comic-page.png",
        proof_id: "proof-comic-1",
        created_at: "2026-05-17T18:00:00Z",
      }, {
        output_id: "comic-page-output-new",
        team_id: "team-alpha",
        work_item_id: "work-1",
        kind: "media",
        label: "Comic page",
        storage_ref: "groups/team-alpha/media/comic-page.png",
        proof_id: "proof-comic-2",
        created_at: "2026-05-17T18:05:00Z",
      }],
      emptyMessage: "No active work.",
      status: "ready",
      statusLabel: "Ready",
      degradedMessage: null,
    });

    render(<SomaOperatingSurface focusedTeamId="team-alpha" />);

    const dock = within(screen.getByTestId("focused-team-output-dock"));
    expect(dock.getByText("Comic page")).toBeDefined();
    expect(screen.getByTestId("focused-team-output-dock").textContent?.indexOf("Comic page")).toBeLessThan(
      screen.getByTestId("focused-team-output-dock").textContent?.indexOf("Old comic page") ?? 0,
    );
    expect(screen.getByRole("link", { name: /Open team/i }).getAttribute("href")).toBe("/teams?team_id=team-alpha");
    expect(dock.getByRole("button", { name: /Open Comic page in a new browser window/i })).toBeDefined();
    expect(dock.getByRole("button", { name: /Open local folder for Comic page/i })).toBeDefined();
  });

  it("prioritizes focused team output over older global Soma chat output", () => {
    mocks.storeState.missionChat = [{
      role: "architect",
      content: "Global package is available.",
      timestamp: "2026-05-17T18:00:00Z",
      execution_summary: {
        outputs: [{ title: "Older global brief", url: "/runs/global-brief" }],
      },
    }];
    mocks.useDurableTeamWork.mockReturnValue({
      items: [{
        id: "work-1",
        title: "Focused output",
        state: "output_ready",
        teamIds: ["team-alpha"],
        interactions: [],
      }],
      outputRefs: [{
        output_id: "focused-output",
        team_id: "team-alpha",
        work_item_id: "work-1",
        kind: "file",
        label: "Newest focused brief",
        storage_ref: "groups/team-alpha/output/newest.md",
        created_at: "2026-05-17T18:05:00Z",
      }],
      emptyMessage: "No active work.",
      status: "ready",
      statusLabel: "Ready",
      degradedMessage: null,
    });

    render(<SomaOperatingSurface focusedTeamId="team-alpha" />);

    const workbench = within(screen.getByTestId("output-workbench"));
    expect(workbench.getByText("Latest output").closest("article")?.textContent).toContain("Newest focused brief");
    expect(workbench.getByText("Output details and proof").closest("details")?.textContent).toContain("Older global brief");
  });
});
