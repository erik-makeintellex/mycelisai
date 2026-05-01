import { describe, it, expect, vi, beforeEach, afterEach } from "vitest";
import { fireEvent, render, screen } from "@testing-library/react";

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

const mockTeams = [
  {
    id: "team-alpha",
    name: "Alpha Squad",
    role: "action",
    type: "standing" as const,
    mission_id: null,
    mission_intent: null,
    inputs: ["nats.input.alpha"],
    deliveries: ["nats.output.alpha"],
    agents: [
      {
        id: "agent-1",
        role: "cognitive",
        status: 1,
        last_heartbeat: new Date().toISOString(),
        tools: [],
        model: "qwen",
      },
      {
        id: "agent-2",
        role: "sensory",
        status: 0,
        last_heartbeat: new Date().toISOString(),
        tools: [],
        model: "qwen",
      },
    ],
  },
  {
    id: "team-bravo",
    name: "Bravo Ops",
    role: "expression",
    type: "mission" as const,
    mission_id: "mission-001",
    mission_intent: "Deploy sentinel network",
    inputs: [],
    deliveries: [],
    agents: [
      {
        id: "agent-3",
        role: "actuation",
        status: 2,
        last_heartbeat: new Date().toISOString(),
        tools: ["exec"],
        model: "llama",
      },
    ],
  },
];

const mockTemplates = [
  {
    id: "template-marketing-writer",
    name: "Marketing Writer",
    role: "cognitive",
    system_prompt: "Write and refine launch copy.",
    model: "qwen3:8b",
    tools: ["recall"],
    inputs: ["briefs"],
    outputs: ["campaign copy", "launch messaging"],
    verification_strategy: "semantic",
    verification_rubric: ["clear", "on-brand"],
    validation_command: "",
    created_at: new Date("2026-04-07T10:00:00Z").toISOString(),
    updated_at: new Date("2026-04-07T12:00:00Z").toISOString(),
  },
  {
    id: "template-researcher",
    name: "Audience Researcher",
    role: "cognitive",
    system_prompt: "Research campaigns and audience insight.",
    model: "llama3.1:8b",
    tools: ["fetch"],
    inputs: ["requests"],
    outputs: ["research briefs"],
    verification_strategy: "semantic",
    verification_rubric: ["grounded"],
    validation_command: "",
    created_at: new Date("2026-04-07T09:00:00Z").toISOString(),
    updated_at: new Date("2026-04-07T11:00:00Z").toISOString(),
  },
];

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
  });

  afterEach(() => {
    global.setInterval = originalSetInterval;
    global.clearInterval = originalClearInterval;
  });

  it("renders the team roster plus Soma team-specialization controls", () => {
    useCortexStore.setState({
      teamsDetail: mockTeams,
      catalogueAgents: mockTemplates,
    });

    render(<TeamsPage />);

    expect(screen.getByText("Alpha Squad")).toBeDefined();
    expect(screen.getByText("Bravo Ops")).toBeDefined();
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
    expect(
      screen.getByText(/Groups have their own workspace now/i),
    ).toBeDefined();
    expect(
      screen
        .getAllByRole("link", { name: /Open groups workspace/i })[0]
        .getAttribute("href"),
    ).toBe("/groups");
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

  it("filter dropdown filters teams by type", () => {
    useCortexStore.setState({
      teamsDetail: mockTeams,
      catalogueAgents: mockTemplates,
    });

    render(<TeamsPage />);

    expect(screen.getByText("Alpha Squad")).toBeDefined();
    expect(screen.getByText("Bravo Ops")).toBeDefined();

    fireEvent.change(screen.getByDisplayValue("All Teams"), {
      target: { value: "standing" },
    });

    expect(screen.getByText("Alpha Squad")).toBeDefined();
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
    expect(screen.getByTestId("team-team-alpha-view-wiring")).toBeDefined();
    expect(screen.getByTestId("team-team-alpha-view-logs")).toBeDefined();
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
