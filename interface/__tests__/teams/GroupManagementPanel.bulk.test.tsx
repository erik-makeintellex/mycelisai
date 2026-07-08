import { describe, it, expect, beforeEach } from "vitest";
import { render, screen, fireEvent, waitFor } from "@testing-library/react";
import GroupManagementPanel from "@/components/teams/GroupManagementPanel";
import { mockFetch } from "../setup";
import {
  installGroupsFetch,
  standingGroup,
  tempGroup,
} from "./GroupManagementPanel.testSupport";

describe("GroupManagementPanel bulk actions", () => {
  beforeEach(() => {
    mockFetch.mockReset();
    window.localStorage.clear();
  });

  it("bulk clears selected active groups from the records rail", async () => {
    const groups = [
      standingGroup(),
      tempGroup({ work_mode: "execute_with_approval" }),
    ];
    installGroupsFetch({ groups });

    render(<GroupManagementPanel />);

    await waitFor(() =>
      expect(
        screen.getByRole("button", { name: /Temp Campaign/i }),
      ).toBeDefined(),
    );

    fireEvent.click(screen.getByRole("button", { name: "Select" }));
    await waitFor(() =>
      expect(screen.getByTestId("groups-bulk-actions")).toBeDefined(),
    );
    fireEvent.click(
      screen.getByRole("button", { name: "Select visible active" }),
    );
    expect(screen.getByTestId("groups-bulk-actions").textContent).toContain(
      "2 selected",
    );
    fireEvent.click(
      screen.getByRole("button", { name: "Clear selected groups" }),
    );

    await waitFor(() =>
      expect(screen.getByTestId("groups-notice").textContent).toContain(
        "2 selected groups cleared from active lanes",
      ),
    );
    const clearRequests = mockFetch.mock.calls.filter(([input, init]) => {
      const url = typeof input === "string" ? input : String(input);
      return url.endsWith("/clear") && init?.method === "POST";
    });
    expect(clearRequests.map(([input]) => input)).toEqual([
      "/api/v1/groups/group-standing/clear",
      "/api/v1/groups/group-temp/clear",
    ]);
    clearRequests.forEach(([, init]) =>
      expect(JSON.parse(String(init?.body))).toEqual({
        include_outputs: false,
      }),
    );
    expect(screen.queryByTestId("groups-bulk-actions")).toBeNull();
  });
});
