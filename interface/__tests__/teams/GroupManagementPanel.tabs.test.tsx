import { describe, it, expect, beforeEach } from "vitest";
import { render, screen, fireEvent, waitFor, act } from "@testing-library/react";
import GroupManagementPanel from "@/components/teams/GroupManagementPanel";
import { mockFetch } from "../setup";
import {
  installGroupsFetch,
  tempGroup,
} from "./GroupManagementPanel.testSupport";

describe("GroupManagementPanel workspace tabs", () => {
  beforeEach(() => {
    mockFetch.mockReset();
    window.localStorage.clear();
  });

  it("supports keyboard movement across deep workspace sections", async () => {
    installGroupsFetch({ groups: [tempGroup()] });
    render(<GroupManagementPanel initialSelectedGroupId="group-temp" />);

    await waitFor(() =>
      expect(screen.getByRole("tab", { name: /Overview/i })).toHaveProperty(
        "id",
        "groups-overview-tab",
      ),
    );

    const tablist = screen.getByRole("tablist", {
      name: "Group workspace sections",
    });
    await act(async () => {
      fireEvent.keyDown(tablist, { key: "ArrowRight" });
    });
    expect(
      screen.getByRole("tab", { name: /Outputs/i }).getAttribute("aria-selected"),
    ).toBe("true");

    await act(async () => {
      fireEvent.keyDown(tablist, { key: "End" });
    });
    expect(
      screen.getByRole("tab", { name: /Create/i }).getAttribute("aria-selected"),
    ).toBe("true");

    await act(async () => {
      fireEvent.keyDown(tablist, { key: "Home" });
    });
    expect(
      screen.getByRole("tab", { name: /Groups/i }).getAttribute("aria-selected"),
    ).toBe("true");
  });
});
