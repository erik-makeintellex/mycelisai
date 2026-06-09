import { describe, it, expect, vi, beforeEach, afterEach } from "vitest";
import { render, screen, waitFor } from "@testing-library/react";
import { mockFetch } from "../setup";

vi.mock("@/components/teams/TeamDetailDrawer", () => ({
  __esModule: true,
  default: ({ team, onClose }: any) => (
    <div data-testid="team-detail-drawer">
      <span>Drawer: {team.name}</span>
      <button onClick={onClose}>Close</button>
    </div>
  ),
}));

vi.mock("@/components/catalogue/AgentEditorDrawer", () => ({
  __esModule: true,
  default: ({ agent, onClose }: any) => (
    <div data-testid="agent-editor-drawer">
      <span>{agent ? `Editing: ${agent.name}` : "Creating new template"}</span>
      <button onClick={onClose}>Close template drawer</button>
    </div>
  ),
}));

import TeamsPage from "@/components/teams/TeamsPage";
import { useCortexStore } from "@/store/useCortexStore";
import { mockTeamWorkFetch, mockTeams, mockTemplates } from "./TeamsPage.fixtures";

describe("TeamsPage review route", () => {
  const originalSetInterval = global.setInterval;
  const originalClearInterval = global.clearInterval;

  beforeEach(() => {
    window.history.pushState({}, "", "/teams?view=work");
    global.setInterval = vi.fn(() => 1) as any;
    global.clearInterval = vi.fn() as any;
    useCortexStore.setState({
      teamsDetail: [],
      isFetchingTeamsDetail: false,
      catalogueAgents: [],
      isFetchingCatalogue: false,
      selectedTeamId: null,
      isTeamDrawerOpen: false,
      teamsFilter: "all",
      fetchTeamsDetail: vi.fn(),
      fetchCatalogue: vi.fn(),
      createCatalogueAgent: vi.fn(),
      updateCatalogueAgent: vi.fn(),
    });
    mockTeamWorkFetch(mockFetch);
  });

  afterEach(() => {
    global.setInterval = originalSetInterval;
    global.clearInterval = originalClearInterval;
  });

  it("opens a review-focused page from the Soma review rail", async () => {
    useCortexStore.setState({
      teamsDetail: mockTeams,
      catalogueAgents: mockTemplates,
    });

    render(<TeamsPage />);

    expect(screen.getByText("Work to Review")).toBeDefined();
    expect(screen.getByText("Decide what happens to this work")).toBeDefined();
    await waitFor(() => {
      expect(screen.getByText("Draft launch package")).toBeDefined();
    });
    expect(screen.getByLabelText("Review queue summary")).toBeDefined();
    expect(screen.getAllByText("Reason").length).toBeGreaterThan(0);
    expect(screen.getAllByText("Decision actions").length).toBeGreaterThan(0);
    expect(screen.getByText(/Clear this from review. Nothing ran/i)).toBeDefined();
    expect(screen.queryByText("Archived stale proof")).toBeNull();
    const pageText = document.body.textContent ?? "";
    expect(pageText.indexOf("Active work lane")).toBe(-1);
    expect(pageText.indexOf("Work to review")).toBeLessThan(
      pageText.indexOf("Team context"),
    );
    expect(screen.getByRole("link", { name: /Open all teams/i }).getAttribute("href")).toBe("/teams");
  });
});
