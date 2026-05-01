import { describe, it, expect, beforeEach } from "vitest";
import { render, screen, fireEvent, waitFor } from "@testing-library/react";
import GroupManagementPanel from "@/components/teams/GroupManagementPanel";
import { mockFetch } from "../setup";
import {
  documentArtifact,
  installApprovalCreateFetch,
  installGroupsFetch,
  standingGroup,
  tempGroup,
} from "./GroupManagementPanel.testSupport";

function fillRequiredCreateFields() {
  fireEvent.change(screen.getByLabelText("Name"), {
    target: { value: "Ops Group" },
  });
  fireEvent.change(screen.getByLabelText("Goal Statement"), {
    target: { value: "Coordinate operational runs" },
  });
}

describe("GroupManagementPanel", () => {
  beforeEach(() => {
    mockFetch.mockReset();
    window.localStorage.clear();
  });

  it("shows approval-required state with confirm token", async () => {
    installApprovalCreateFetch();

    render(<GroupManagementPanel />);
    fillRequiredCreateFields();
    fireEvent.click(screen.getByTestId("groups-create-button"));

    await waitFor(() =>
      expect(screen.getByTestId("groups-approval-card")).toBeDefined(),
    );
    expect(screen.getByTestId("groups-confirm-token-input")).toHaveProperty(
      "value",
      "tok-123",
    );
    expect(screen.getByTestId("groups-notice").textContent).toContain(
      "Approval required",
    );
  });

  it("resubmits create request with confirm_token after approval prompt", async () => {
    const postBodies: Array<Record<string, unknown>> = [];
    installApprovalCreateFetch(postBodies);

    render(<GroupManagementPanel />);
    fillRequiredCreateFields();
    fireEvent.click(screen.getByTestId("groups-create-button"));

    await waitFor(() =>
      expect(screen.getByTestId("groups-approval-card")).toBeDefined(),
    );
    fireEvent.click(screen.getByTestId("groups-create-button"));

    await waitFor(() =>
      expect(screen.getByTestId("groups-notice").textContent).toContain(
        "Group created successfully",
      ),
    );
    expect(postBodies).toHaveLength(2);
    expect(postBodies[1].confirm_token).toBe("tok-123");
  });

  it("archives temporary groups and keeps retained outputs reviewable", async () => {
    const groups = [
      standingGroup(),
      tempGroup({ work_mode: "execute_with_approval" }),
    ];
    installGroupsFetch({
      groups,
      monitor: {
        status: "online",
        published_count: 2,
        last_group_id: "group-temp",
      },
      outputs: { "group-temp": [documentArtifact()] },
    });

    render(<GroupManagementPanel />);
    await waitFor(() =>
      expect(
        screen.getByRole("button", { name: /Temp Campaign/i }),
      ).toBeDefined(),
    );
    fireEvent.click(screen.getByRole("button", { name: /Temp Campaign/i }));

    await waitFor(() => expect(screen.getByText("Launch brief")).toBeDefined());
    fireEvent.click(
      screen.getByRole("button", { name: "Archive temporary group" }),
    );

    await waitFor(() =>
      expect(screen.getByTestId("groups-notice").textContent).toContain(
        "Temporary group archived",
      ),
    );
    fireEvent.click(screen.getByTestId("groups-list-item-group-temp"));
    await waitFor(() =>
      expect(
        screen.getByTestId("groups-archived-readonly-note").textContent,
      ).toContain("retained output review"),
    );
    expect(
      screen.getByTestId("groups-retained-outputs-note").textContent,
    ).toContain("Downloads remain available");
    expect(screen.getByText("Campaign summary")).toBeDefined();
    expect(screen.getByTestId("groups-output-summary").textContent).toContain(
      "1 output",
    );
    expect(
      screen.getByRole("link", { name: /Download/i }).getAttribute("href"),
    ).toBe("/api/v1/artifacts/artifact-1/download");
    expect(
      screen
        .getByRole("link", { name: "Open lead workspace" })
        .getAttribute("href"),
    ).toBe("/dashboard?team_id=team-marketing");
    await waitFor(() =>
      expect(
        screen.queryByRole("button", { name: "Broadcast to group" }),
      ).toBeNull(),
    );
  });

  it("honors an initially selected group id from the route", async () => {
    installGroupsFetch({ groups: [standingGroup(), tempGroup()] });

    render(<GroupManagementPanel initialSelectedGroupId="group-temp" />);

    await waitFor(() =>
      expect(
        screen.getByRole("heading", { name: "Temp Campaign" }),
      ).toBeDefined(),
    );
  });

  it("keeps a route-selected group visible when saved filters hide it", async () => {
    window.localStorage.setItem(
      "mycelis.groups.recordFilters",
      JSON.stringify({ query: "", kind: "standing", state: "all", retentionDays: 30 }),
    );
    installGroupsFetch({ groups: [standingGroup(), tempGroup()] });

    render(<GroupManagementPanel initialSelectedGroupId="group-temp" />);

    await waitFor(() =>
      expect(
        screen.getByRole("heading", { name: "Temp Campaign" }),
      ).toBeDefined(),
    );
    expect(screen.getByText("Selected outside filters")).toBeDefined();
    expect(screen.getByTestId("groups-list").textContent).toContain(
      "Standing Ops",
    );
    expect(screen.getByTestId("groups-list").textContent).not.toContain(
      "Temp Campaign",
    );
  });

  it("prioritizes setup and review surfaces before coordination logging", async () => {
    installGroupsFetch({
      groups: [
        tempGroup({
          team_ids: ["team-marketing", "team-design"],
          allowed_capabilities: ["runs.read", "runs.propose"],
        }),
      ],
      outputs: {
        "group-temp": [
          documentArtifact({ title: "Launch Brief" }),
          documentArtifact({
            id: "artifact-2",
            agent_id: "design-lead",
            artifact_type: "file",
            title: "Asset Bundle",
            file_path: "workspace/asset-bundle.zip",
          }),
        ],
      },
    });

    render(<GroupManagementPanel initialSelectedGroupId="group-temp" />);

    await waitFor(() => expect(screen.getByText("Launch Brief")).toBeDefined());
    expect(screen.getByText("Agent backend model")).toBeDefined();
    expect(screen.getByText("Inherits organization AI Engine")).toBeDefined();
    expect(screen.getByText("runs.read, runs.propose")).toBeDefined();
    expect(
      screen
        .getByText("Define group action lane")
        .compareDocumentPosition(screen.getByText("Group records")) &
        Node.DOCUMENT_POSITION_FOLLOWING,
    ).toBeTruthy();
    expect(
      screen
        .getByText("Group records")
        .compareDocumentPosition(screen.getByText("Coordination activity")) &
        Node.DOCUMENT_POSITION_FOLLOWING,
    ).toBeTruthy();
    expect(screen.getByTestId("groups-output-summary").textContent).toContain(
      "2 outputs",
    );
    expect(screen.getByTestId("groups-output-summary").textContent).toContain(
      "2 contributing leads",
    );
  });

  it("filters group records by type, state, and completed record retention", async () => {
    installGroupsFetch({
      groups: [
        standingGroup(),
        tempGroup(),
        tempGroup({
          group_id: "group-complete",
          name: "Recent Complete",
          status: "archived",
          expiry: new Date(Date.now() - 2 * 86_400_000).toISOString(),
        }),
        tempGroup({
          group_id: "group-old-complete",
          name: "Old Complete",
          status: "archived",
          expiry: new Date(Date.now() - 45 * 86_400_000).toISOString(),
        }),
      ],
    });

    render(<GroupManagementPanel />);

    await waitFor(() =>
      expect(screen.getByTestId("groups-list").textContent).toContain(
        "Standing Ops",
      ),
    );
    expect(screen.getByTestId("groups-list").textContent).toContain(
      "Temp Campaign",
    );
    expect(screen.getByTestId("groups-list").textContent).toContain(
      "Recent Complete",
    );
    expect(screen.getByTestId("groups-list").textContent).not.toContain(
      "Old Complete",
    );

    fireEvent.click(screen.getByRole("button", { name: "Temp" }));
    expect(screen.getByTestId("groups-list").textContent).not.toContain(
      "Standing Ops",
    );

    fireEvent.change(screen.getByLabelText("Search group records"), {
      target: { value: "recent" },
    });
    expect(screen.getByTestId("groups-list").textContent).toContain(
      "Recent Complete",
    );
    expect(screen.getByTestId("groups-list").textContent).not.toContain(
      "Temp Campaign",
    );
    fireEvent.change(screen.getByLabelText("Search group records"), {
      target: { value: "" },
    });

    fireEvent.click(screen.getByRole("button", { name: "Complete" }));
    expect(screen.getByTestId("groups-list").textContent).toContain(
      "Recent Complete",
    );
    expect(screen.getByTestId("groups-list").textContent).not.toContain(
      "Temp Campaign",
    );

    fireEvent.change(screen.getByLabelText("Completed record retention days"), {
      target: { value: "60" },
    });
    await waitFor(() =>
      expect(screen.getByTestId("groups-list").textContent).toContain(
        "Old Complete",
      ),
    );
  });
});
