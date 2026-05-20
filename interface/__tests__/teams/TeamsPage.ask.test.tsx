import { describe, it, expect, vi, beforeEach, afterEach } from "vitest";
import { fireEvent, render, screen, waitFor, within } from "@testing-library/react";
import { mockFetch } from "../setup";

vi.mock("@/components/teams/TeamDetailDrawer", () => ({
  __esModule: true,
  default: () => <div data-testid="team-detail-drawer" />,
}));

vi.mock("@/components/catalogue/AgentEditorDrawer", () => ({
  __esModule: true,
  default: () => <div data-testid="agent-editor-drawer" />,
}));

import TeamsPage from "@/components/teams/TeamsPage";
import { useCortexStore } from "@/store/useCortexStore";
import { mockTeamWorkFetch, mockTeams, mockTemplates } from "./TeamsPage.fixtures";

describe("TeamsPage bounded ask", () => {
  const originalSetInterval = global.setInterval;
  const originalClearInterval = global.clearInterval;

  beforeEach(() => {
    global.setInterval = vi.fn(() => 1) as any;
    global.clearInterval = vi.fn() as any;
    useCortexStore.setState({
      teamsDetail: mockTeams,
      isFetchingTeamsDetail: false,
      catalogueAgents: mockTemplates,
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

  it("posts bounded team asks and refreshes active work", async () => {
    render(<TeamsPage />);

    await waitFor(() => {
      expect(screen.getByText("Draft launch package")).toBeDefined();
    });
    const targetRow = screen
      .getByText("Draft launch package")
      .closest("article") as HTMLElement;
    fireEvent.click(
      within(targetRow).getByRole("button", { name: /ask team/i }),
    );
    const input = await waitFor(() =>
      within(targetRow).getByLabelText(/Ask Draft launch package/i),
    );
    fireEvent.change(input, {
      target: { value: "Create the next validation note" },
    });
    fireEvent.submit(input.closest("form") as HTMLFormElement);

    await waitFor(() => {
      expect(mockFetch).toHaveBeenCalledWith(
        "/api/v1/teams/team-alpha/work/ask",
        expect.objectContaining({
          method: "POST",
          body: expect.stringContaining("Create the next validation note"),
        }),
      );
    });
    expect(
      mockFetch.mock.calls.filter(([url]) =>
        String(url).includes("/api/v1/teams/team-alpha/work?limit=8"),
      ).length,
    ).toBeGreaterThan(1);
  });
});
