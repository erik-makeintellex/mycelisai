import { render, screen, fireEvent, waitFor } from "@testing-library/react";
import { describe, it, expect, beforeEach } from "vitest";
import GroupManagementPanel from "@/components/teams/GroupManagementPanel";
import { mockFetch } from "../setup";
import {
  documentArtifact,
  installGroupsFetch,
  tempGroup,
} from "./GroupManagementPanel.testSupport";

describe("GroupManagementPanel output classes", () => {
  beforeEach(() => {
    mockFetch.mockReset();
    window.localStorage.clear();
  });

  it("does not treat planning-only records as delivered outputs", async () => {
    const group = tempGroup({
      group_id: "group-planned",
      name: "Mixed Output Team",
      team_ids: ["mixed-output-team"],
    });
    installGroupsFetch({
      groups: [group],
      outputs: {
        "group-planned": [
          documentArtifact({
            id: "planning-artifact",
            title: "groups/mixed-output-team/planning/TEAM_EVOCATION.md",
            file_path: "groups/mixed-output-team/planning/TEAM_EVOCATION.md",
            metadata: { output_class: "planning" },
          }),
        ],
      },
      lifecycle: plannedOnlyLifecycle(group.expiry as string | null),
    });

    render(<GroupManagementPanel initialSelectedGroupId="group-planned" />);

    await waitFor(() =>
      expect(screen.getAllByText("Planned only").length).toBeGreaterThan(0),
    );
    fireEvent.click(screen.getByRole("tab", { name: /Outputs/i }));
    expect(screen.getByTestId("groups-output-summary").textContent).toContain(
      "0 delivered",
    );
    expect(screen.getByText(/No delivered outputs yet/i)).toBeDefined();
    expect(screen.queryByText(/TEAM_EVOCATION.md/)).toBeNull();

    fireEvent.click(
      screen.getByRole("checkbox", {
        name: /Show planning, proof, and team source records/i,
      }),
    );

    await waitFor(() =>
      expect(screen.getByText(/TEAM_EVOCATION.md/)).toBeDefined(),
    );
    expect(screen.getByTestId("groups-output-summary").textContent).toContain(
      "1 internal",
    );
  });
});

function plannedOnlyLifecycle(expiry?: string | null) {
  return {
    generated_at: new Date().toISOString(),
    summary: {
      total_groups: 1,
      active_groups: 1,
      expired_active_groups: 0,
      standing_no_expiry_groups: 0,
      stale_standing_groups: 0,
      review_needed_groups: 1,
      output_ready_idle_groups: 0,
      team_work_needing_attention: 0,
    },
    items: [
      {
        group_id: "group-planned",
        name: "Mixed Output Team",
        status: "active",
        work_mode: "propose_only",
        kind: "temporary",
        recommendation: "review_work",
        reason:
          "Linked team work has retained internal material but no user-facing deliverable yet.",
        expiry,
        expired: false,
        age_hours: 1,
        team_count: 1,
        output_count: 0,
        team_work_count: 1,
        active_or_blocked_work_count: 0,
        output_ready_work_count: 1,
        archived_work_count: 0,
      },
    ],
  };
}
