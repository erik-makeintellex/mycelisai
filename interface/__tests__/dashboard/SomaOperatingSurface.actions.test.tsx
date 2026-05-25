import React from "react";
import { fireEvent, render, screen } from "@testing-library/react";
import { beforeEach, describe, expect, it, vi } from "vitest";
import { SomaOperatingSurface } from "@/components/soma/SomaOperatingSurface";

const mocks = vi.hoisted(() => {
  const selectTeam = vi.fn();
  const handleActiveWorkAction = vi.fn();
  const handleTeamAsk = vi.fn();
  const useDurableTeamWork = vi.fn();
  const useTeamWorkActionHandler = vi.fn();
  return {
    selectTeam,
    handleActiveWorkAction,
    handleTeamAsk,
    useDurableTeamWork,
    useTeamWorkActionHandler,
  };
});

vi.mock("@/store/useCortexStore", () => ({
  useCortexStore: (selector: (state: unknown) => unknown) => selector({
    missionChat: [],
    teamsDetail: [{ id: "team-alpha", name: "Alpha" }],
    durableWorkRefreshVersion: 5,
    selectTeam: mocks.selectTeam,
  }),
}));

vi.mock("@/components/teams/useTeamWorkActionHandler", () => ({
  useTeamWorkActionHandler: mocks.useTeamWorkActionHandler,
}));

vi.mock("@/components/soma/useDurableTeamWork", () => ({
  useDurableTeamWork: mocks.useDurableTeamWork,
}));

vi.mock("@/components/teams/ActiveWorkLane", () => ({
  ActiveWorkLane: (props: {
    items: Array<{ id: string; title: string }>;
    degradedMessage?: string | null;
    onAction?: (item: unknown, action: unknown) => void;
    onTeamAsk?: (item: unknown, message: string) => void;
  }) => (
    <section data-testid="active-work-lane">
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
  default: () => <div data-testid="mission-chat" />,
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
    mocks.useTeamWorkActionHandler.mockReturnValue({
      activeWorkRefreshVersion: 7,
      activeWorkActionError: "Team action needs operator attention.",
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
    expect(screen.getByText("Team action needs operator attention.")).toBeDefined();

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
});
