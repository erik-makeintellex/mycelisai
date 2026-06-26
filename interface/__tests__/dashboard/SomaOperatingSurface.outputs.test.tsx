import React from "react";
import { render, screen, within } from "@testing-library/react";
import { beforeEach, describe, expect, it, vi } from "vitest";
import { SomaOperatingSurface } from "@/components/soma/SomaOperatingSurface";

const mocks = vi.hoisted(() => {
  const useDurableTeamWork = vi.fn();
  const storeState = {
    missionChat: [] as unknown[],
    teamsDetail: [{ id: "team-alpha", name: "Alpha" }],
    durableWorkRefreshVersion: 5,
    selectedTeamId: null as string | null,
    selectTeam: vi.fn(),
    sendMissionChat: vi.fn(),
  };
  return { storeState, useDurableTeamWork };
});

vi.mock("@/store/useCortexStore", () => ({
  useCortexStore: (selector: (state: unknown) => unknown) => selector(mocks.storeState),
}));

vi.mock("@/components/teams/useTeamWorkActionHandler", () => ({
  useTeamWorkActionHandler: () => ({
    activeWorkRefreshVersion: 0,
    activeWorkActionError: null,
    activeWorkActionNotice: null,
    submittedTeamWorkItems: [],
    handleActiveWorkAction: vi.fn(),
    handleTeamAsk: vi.fn(),
  }),
  mergeTeamWorkItems: (
    durableItems: Array<{ id: string }>,
    submittedItems: Array<{ id: string }>,
  ) => [...submittedItems, ...durableItems],
}));

vi.mock("@/components/soma/useDurableTeamWork", () => ({
  useDurableTeamWork: mocks.useDurableTeamWork,
}));

vi.mock("@/components/dashboard/MissionControlChat", () => ({
  default: () => <div data-testid="mission-chat" />,
}));

vi.mock("@/components/soma/SomaWorkspaceFrame", () => ({
  SomaWorkspaceFrame: ({
    expression,
    output,
    primaryPanel,
    showOutputDigest,
  }: {
    expression: React.ReactNode;
    output: React.ReactNode;
    primaryPanel?: string;
    showOutputDigest?: boolean;
  }) => (
    <div
      data-testid="mock-soma-workspace-frame"
      data-primary-panel={primaryPanel ?? ""}
      data-show-output-digest={String(showOutputDigest)}
    >
      {expression}
      {output}
    </div>
  ),
}));

describe("SomaOperatingSurface output projection", () => {
  beforeEach(() => {
    vi.clearAllMocks();
    mocks.storeState.missionChat = [];
    mocks.storeState.selectedTeamId = null;
  });

  it("keeps focused team retained outputs in the workbench instead of stacking a pre-chat dock", () => {
    mocks.useDurableTeamWork.mockReturnValue({
      items: [{
        id: "work-1",
        title: "Comic page generation",
        state: "output_ready",
        teamIds: ["team-alpha"],
        interactions: [],
      }],
      outputRefs: [outputRef("old", "Old comic page", "old-comic-page.png"), outputRef("new", "Comic page", "comic-page.png")],
      emptyMessage: "No active work.",
      status: "ready",
      statusLabel: "Ready",
      degradedMessage: null,
    });

    render(<SomaOperatingSurface focusedTeamId="team-alpha" />);

    const workspace = within(screen.getByTestId("mock-soma-workspace-frame"));
    const workbench = within(workspace.getByTestId("output-workbench"));
    const vault = within(screen.getByTestId("soma-outcome-vault"));
    expect(screen.getByTestId("mock-soma-workspace-frame").getAttribute("data-primary-panel")).toBe("");
    expect(screen.getByTestId("mock-soma-workspace-frame").getAttribute("data-show-output-digest")).toBe("true");
    expect(screen.queryByTestId("focused-team-output-dock")).toBeNull();
    expect(workbench.getByText("Comic page")).toBeDefined();
    expect(vault.getByRole("link", { name: /Open latest deliverable Comic page/i }).getAttribute("data-target-reference")).toContain("groups/team-alpha/media/comic-page.png");
    expect(workspace.getByTestId("output-workbench").textContent?.indexOf("Comic page")).toBeLessThan(
      workspace.getByTestId("output-workbench").textContent?.indexOf("Old comic page") ?? 0,
    );
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
      items: [{ id: "work-1", title: "Focused output", state: "output_ready", teamIds: ["team-alpha"], interactions: [] }],
      outputRefs: [outputRef("focused", "Newest focused brief", "output/newest.md", "file")],
      emptyMessage: "No active work.",
      status: "ready",
      statusLabel: "Ready",
      degradedMessage: null,
    });

    render(<SomaOperatingSurface focusedTeamId="team-alpha" />);

    const workbench = within(screen.getByTestId("mock-soma-workspace-frame")).getByTestId("output-workbench");
    expect(within(workbench).getByText("Latest output").closest("article")?.textContent).toContain("Newest focused brief");
    expect(within(workbench).getByText("More outputs and verification").closest("details")?.textContent).toContain("Older global brief");
  });
});

function outputRef(id: string, label: string, path: string, kind = "media") {
  return {
    output_id: `comic-page-output-${id}`,
    team_id: "team-alpha",
    work_item_id: "work-1",
    kind,
    label,
    storage_ref: `groups/team-alpha/${kind === "media" ? "media" : ""}/${path}`.replace("//", "/"),
    proof_id: `proof-comic-${id}`,
    created_at: id === "old" ? "2026-05-17T18:00:00Z" : "2026-05-17T18:05:00Z",
  };
}
