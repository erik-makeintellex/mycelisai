import React from "react";
import { fireEvent, render, screen, within } from "@testing-library/react";
import { beforeEach, describe, expect, it, vi } from "vitest";
import { SomaOperatingSurface } from "@/components/soma/SomaOperatingSurface";

const mocks = vi.hoisted(() => {
  const selectTeam = vi.fn();
  const handleActiveWorkAction = vi.fn();
  const handleTeamAsk = vi.fn();
  const missionControlChat = vi.fn();
  const sendMissionChat = vi.fn();
  const useDurableTeamWork = vi.fn();
  const useTeamWorkActionHandler = vi.fn();
  const storeState = {
    missionChat: [] as unknown[],
    teamsDetail: [{ id: "team-alpha", name: "Alpha" }],
    durableWorkRefreshVersion: 5,
    selectTeam,
    selectedTeamId: null as string | null,
    sendMissionChat,
  };
  return {
    selectTeam,
    handleActiveWorkAction,
    handleTeamAsk,
    missionControlChat,
    sendMissionChat,
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

vi.mock("@/components/soma/SomaWorkspaceFrame", () => ({
  SomaWorkspaceFrame: ({
    activeWork,
    context,
    expression,
    output,
    primaryPanel,
    reviewCount,
    showOutputDigest,
    trust,
  }: { activeWork: React.ReactNode; context: React.ReactNode; expression: React.ReactNode; output: React.ReactNode; primaryPanel?: string; reviewCount?: number; showOutputDigest?: boolean; trust: React.ReactNode }) => (
    <div
      data-testid="mock-soma-workspace-frame"
      data-primary-panel={primaryPanel ?? ""}
      data-review-count={reviewCount ?? ""}
      data-show-output-digest={String(showOutputDigest)}
    >
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
    expect(mocks.useDurableTeamWork).toHaveBeenCalledWith(expect.objectContaining({ focusedTeamId: "team-alpha", refreshVersion: 12 }));
    expect(mocks.missionControlChat).toHaveBeenCalledWith(expect.objectContaining({ focusedTeamId: "team-alpha" }));
    expect(screen.getByTestId("soma-team-context-switcher").textContent).toContain("Working in");
    expect(screen.getByTestId("soma-team-context-switcher").textContent).toContain("Alpha");
    expect(screen.getByTestId("soma-team-context-switcher").textContent).toContain("Team chat, work, outputs, and proof");
    expect(screen.queryByRole("tab", { name: /Alpha/i })).toBeNull();
    expect(screen.getByRole("button", { name: /Alpha/i }).getAttribute("aria-expanded")).toBe("false");
    expect(screen.getByTestId("mock-soma-workspace-frame").getAttribute("data-primary-panel")).toBe("work");
    expect(screen.getByTestId("mock-soma-workspace-frame").getAttribute("data-show-output-digest")).toBe("true");
    expect(screen.getByTestId("soma-action-shelf")).toBeDefined();
    expect(screen.getAllByText("Soma").length).toBeGreaterThan(0);
    expect(screen.getAllByText("Team action needs operator attention.").length).toBeGreaterThan(0);
    expect(screen.getAllByText("Team ask queued. You can keep working.").length).toBeGreaterThan(0);
    fireEvent.click(screen.getAllByRole("button", { name: /Open Outcome Vault/i })[0]);

    expect(screen.getByTestId("soma-outcome-vault-overlay")).toBeDefined();
    expect(screen.getByText("Outcome ready to revisit")).toBeDefined();
    expect(screen.getByText("Alpha outcome workspace")).toBeDefined();
    expect(screen.queryByText("OutcomeProject owner:")).toBeNull();
    expect(screen.queryByText("TeamRegistry owner:")).toBeNull();
    expect(screen.getByRole("link", { name: /Review next step/i }).getAttribute("href")).toBe("/teams?view=work");
    const typedRailAlert = screen.getByRole("link", { name: /Review background work: Work in progress/i });
    expect(typedRailAlert.getAttribute("href")).toBe("/teams?view=work&work_item_id=work-1");
    expect(typedRailAlert.getAttribute("data-target-reference")).toBe("work:work-1");
    expect(typedRailAlert.getAttribute("data-target-type")).toBe("work");
    expect(typedRailAlert.getAttribute("data-target-id")).toBe("work-1");

    const workspace = within(screen.getByTestId("mock-soma-workspace-frame"));
    fireEvent.click(screen.getByRole("button", { name: /Run Expense Audit/i }));
    fireEvent.click(workspace.getByRole("button", { name: /pause work/i }));
    fireEvent.click(workspace.getByRole("button", { name: /ask team/i }));

    expect(mocks.sendMissionChat).toHaveBeenCalledWith(expect.stringContaining("expense audit"));
    expect(mocks.handleActiveWorkAction).toHaveBeenCalledWith(expect.objectContaining({ id: "work-1" }), expect.objectContaining({ action: "pause" }));
    expect(mocks.handleTeamAsk).toHaveBeenCalledWith(expect.objectContaining({ id: "work-1" }), "Continue the proof");
  });

  it("opens Outcome Vault as an overlay instead of a default side rail", () => {
    render(<SomaOperatingSurface focusedTeamId="team-alpha" />);

    expect(screen.queryByTestId("soma-outcome-vault")).toBeNull();
    expect(screen.queryByText("Outcome Vault")).toBeNull();
    expect(screen.getByTestId("mission-chat")).toBeDefined();

    fireEvent.click(screen.getAllByRole("button", { name: /Open Outcome Vault/i })[0]);

    expect(screen.getByTestId("soma-outcome-vault-overlay")).toBeDefined();
    expect(screen.getByTestId("soma-outcome-vault").getAttribute("data-state")).toBe("expanded");
    expect(screen.getByText("Outcome Vault")).toBeDefined();

    fireEvent.click(screen.getByRole("button", { name: /Close Outcome Vault/i }));

    expect(screen.queryByTestId("soma-outcome-vault")).toBeNull();
    expect(screen.getByTestId("mission-chat")).toBeDefined();
  });

  it("uses API target refs for quiet right-rail alert links", () => {
    mocks.useDurableTeamWork.mockReturnValue({
      items: [{
        id: "work-1",
        title: "Recover browser game validation",
        state: "running",
        teamIds: ["team-alpha"],
        interactions: [],
        targetRef: {
          type: "recovery",
          id: "work-1",
          work_item_id: "work-1",
          team_id: "team-alpha",
          label: "Recovery item",
        },
      }],
      outputRefs: [],
      emptyMessage: "No active work.",
      status: "ready",
      statusLabel: "Ready",
      degradedMessage: null,
    });

    render(<SomaOperatingSurface focusedTeamId="team-alpha" />);

    fireEvent.click(screen.getAllByRole("button", { name: /Open Outcome Vault/i })[0]);

    const typedRailAlert = screen.getByRole("link", { name: /Review background work: Work in progress/i });
    expect(typedRailAlert.getAttribute("href")).toBe("/teams?view=work&work_item_id=work-1");
    expect(typedRailAlert.getAttribute("data-target-reference")).toBe("recovery:work-1");
    expect(typedRailAlert.getAttribute("data-target-type")).toBe("recovery");
    expect(typedRailAlert.getAttribute("data-target-id")).toBe("work-1");
  });

  it("lets operators switch the Soma surface into a team work context", () => {
    render(<SomaOperatingSurface />);

    fireEvent.click(screen.getByRole("button", { name: /Soma root/i }));
    fireEvent.click(screen.getByRole("option", { name: /Alpha/i }));

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

    expect(screen.queryByTestId("soma-context-focus-bar")).toBeNull();
    expect(screen.getByTestId("soma-team-context-switcher")).toBeDefined();
  });

});
