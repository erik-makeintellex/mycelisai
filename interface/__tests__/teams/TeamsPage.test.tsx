import { describe, it, expect, vi, beforeEach, afterEach } from "vitest";
import { fireEvent, render, screen, waitFor } from "@testing-library/react";
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

describe("TeamsPage", () => {
  const originalSetInterval = global.setInterval;
  const originalClearInterval = global.clearInterval;

  beforeEach(() => {
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

  it("renders the team roster plus durable active-work controls", async () => {
    useCortexStore.setState({
      teamsDetail: mockTeams,
      catalogueAgents: mockTemplates,
    });

    render(<TeamsPage />);

    expect(screen.getAllByText("Alpha Squad").length).toBeGreaterThan(0);
    expect(screen.getAllByText("Bravo Ops").length).toBeGreaterThan(0);
    expect(screen.getByText("Team Lead Workspaces")).toBeDefined();
    expect(screen.getByText(/Review live teams here/i)).toBeDefined();
    expect(
      screen.getByText(/Specialize new teams through Soma/i),
    ).toBeDefined();
    expect(screen.getByText(/Soma team-member templates/i)).toBeDefined();
    expect(
      screen
        .getByRole("link", { name: /Open guided team creation/i })
        .getAttribute("href"),
    ).toBe("/teams/create");
    expect(screen.getAllByText("Marketing Writer").length).toBeGreaterThan(0);
    expect(screen.getAllByText("Audience Researcher").length).toBeGreaterThan(
      0,
    );
    expect(screen.getAllByText(/campaign copy/i).length).toBeGreaterThan(0);
    expect(
      screen.getByText(/Agent type, model, and MCP access/i),
    ).toBeDefined();
    expect(
      screen
        .getByRole("link", { name: /Manage MCP tools/i })
        .getAttribute("href"),
    ).toBe("/resources?tab=tools");
    expect(screen.getByText(/Outputs and active collaboration/i)).toBeDefined();
    expect(screen.getByTestId("active-work-lane")).toBeDefined();
    expect(screen.getByText("Active work lane")).toBeDefined();
    await waitFor(() => {
      expect(screen.getByText("Draft launch package")).toBeDefined();
    });
    expect(screen.getByText("Durable team-work state loaded.")).toBeDefined();
    expect(screen.getAllByText("output ready").length).toBeGreaterThan(0);
    expect(screen.getAllByText("Durable team work").length).toBeGreaterThan(0);
    expect(screen.queryByText("Projection fallback")).toBeNull();
    expect(screen.queryByText(/Durable TeamWorkItem records were unavailable/i)).toBeNull();
    expect(screen.getByRole("link", { name: /Run proof/i }).getAttribute("href")).toBe("/runs/run-alpha");
    expect(screen.getByRole("link", { name: /Launch brief/i }).getAttribute("href")).toBe("/api/v1/workspace/files/view?path=generated%2Falpha%2Fbrief.md");
    expect(
      screen
        .getAllByRole("link", { name: /Review outputs|Review group outputs/i })[0]
        .getAttribute("href"),
    ).toBe("/groups");
    expect(
      screen
        .getByRole("link", { name: /Configure event rules/i })
        .getAttribute("href"),
    ).toBe("/automations?tab=triggers");
    expect(
      screen
        .getByRole("link", { name: /Open Soma workspace/i })
        .getAttribute("href"),
    ).toBe("/dashboard");
    expect(
      screen
        .getByRole("link", { name: /Open full role library/i })
        .getAttribute("href"),
    ).toBe("/resources?tab=roles");
    expect(screen.getByText(/2 teams/)).toBeDefined();
    expect(screen.getByText("2/3 agents online")).toBeDefined();
  });

  it("posts durable active-work actions and refreshes the lane", async () => {
    useCortexStore.setState({
      teamsDetail: mockTeams,
      catalogueAgents: mockTemplates,
    });

    render(<TeamsPage />);

    await waitFor(() => {
      expect(screen.getByText("Review deployment proof")).toBeDefined();
    });
    const pause = screen
      .getAllByRole("button", { name: /pause/i })
      .find((button) => !(button as HTMLButtonElement).disabled);
    expect(pause).toBeDefined();

    fireEvent.click(pause as HTMLElement);

    await waitFor(() => {
      expect(mockFetch).toHaveBeenCalledWith(
        "/api/v1/teams/team-bravo/work/work-bravo/actions",
        expect.objectContaining({
          method: "POST",
          body: expect.stringContaining('"action":"pause"'),
        }),
      );
    });
    expect(
      mockFetch.mock.calls.filter(([url]) =>
        String(url).includes("/api/v1/teams/team-bravo/work?limit=8"),
      ).length,
    ).toBeGreaterThan(1);
  });

  it("posts recover and steer actions as durable team-work evidence", async () => {
    useCortexStore.setState({
      teamsDetail: mockTeams,
      catalogueAgents: mockTemplates,
    });

    render(<TeamsPage />);

    await waitFor(() => {
      expect(screen.getByText("Recover failed release notes")).toBeDefined();
    });
    const recover = screen
      .getAllByRole("button", { name: /recover/i })
      .find((button) => !(button as HTMLButtonElement).disabled);
    expect(recover).toBeDefined();
    fireEvent.click(recover as HTMLElement);

    await waitFor(() => {
      expect(mockFetch).toHaveBeenCalledWith(
        "/api/v1/teams/team-bravo/work/work-bravo-recover/actions",
        expect.objectContaining({
          method: "POST",
          body: expect.stringContaining('"action":"recover"'),
        }),
      );
    });
    fireEvent.click(screen.getAllByRole("button", { name: /steer/i })[0]);
    await waitFor(() => {
      expect(mockFetch).toHaveBeenCalledWith(
        expect.stringContaining("/actions"),
        expect.objectContaining({
          method: "POST",
          body: expect.stringContaining('"action":"steer"'),
        }),
      );
    });
  });

  it("filter dropdown filters teams by type", () => {
    useCortexStore.setState({
      teamsDetail: mockTeams,
      catalogueAgents: mockTemplates,
    });

    render(<TeamsPage />);

    expect(screen.getAllByText("Alpha Squad").length).toBeGreaterThan(0);
    expect(screen.getAllByText("Bravo Ops").length).toBeGreaterThan(0);

    fireEvent.change(screen.getByDisplayValue("All Teams"), {
      target: { value: "standing" },
    });

    expect(screen.getAllByText("Alpha Squad").length).toBeGreaterThan(0);
    expect(screen.queryByText("Bravo Ops")).toBeNull();
  });

  it("clicking a team card opens the detail drawer", () => {
    useCortexStore.setState({
      teamsDetail: mockTeams,
      catalogueAgents: mockTemplates,
      selectedTeamId: null,
      isTeamDrawerOpen: false,
    });

    render(<TeamsPage />);

    expect(screen.queryByTestId("team-detail-drawer")).toBeNull();
    fireEvent.click(screen.getByRole("button", { name: /Alpha Squad/i }));
    expect(screen.getByTestId("team-detail-drawer")).toBeDefined();
    expect(screen.getByText("Drawer: Alpha Squad")).toBeDefined();
  });

  it("renders team quick action links", () => {
    useCortexStore.setState({
      teamsDetail: mockTeams,
      catalogueAgents: mockTemplates,
    });

    render(<TeamsPage />);

    expect(screen.getAllByText("Open lead workspace").length).toBeGreaterThan(
      0,
    );
    expect(
      screen.getByTestId("team-team-alpha-open-chat").getAttribute("href"),
    ).toBe("/dashboard?team_id=team-alpha");
    expect(screen.getByTestId("team-team-alpha-view-runs")).toBeDefined();
    expect(screen.queryByTestId("team-team-alpha-view-wiring")).toBeNull();
    expect(screen.queryByTestId("team-team-alpha-view-logs")).toBeNull();
    expect(screen.queryByText("nats.output.alpha")).toBeNull();
  });

  it("opens the team-member template drawer from the teams page", () => {
    useCortexStore.setState({
      teamsDetail: mockTeams,
      catalogueAgents: mockTemplates,
    });

    render(<TeamsPage />);

    fireEvent.click(screen.getByRole("button", { name: /new template/i }));
    expect(screen.getByTestId("agent-editor-drawer")).toBeDefined();
    expect(screen.getByText("Creating new template")).toBeDefined();

    fireEvent.click(screen.getByRole("button", { name: /Marketing Writer/i }));
    expect(screen.getByText("Editing: Marketing Writer")).toBeDefined();
  });
});
