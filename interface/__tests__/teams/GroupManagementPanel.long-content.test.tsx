import { describe, it, expect, beforeEach } from "vitest";
import { render, screen, fireEvent, waitFor } from "@testing-library/react";
import GroupManagementPanel from "@/components/teams/GroupManagementPanel";
import { mockFetch } from "../setup";
import {
  documentArtifact,
  installGroupsFetch,
  tempGroup,
} from "./GroupManagementPanel.testSupport";

describe("GroupManagementPanel long generated content", () => {
  beforeEach(() => {
    mockFetch.mockReset();
    window.localStorage.clear();
  });

  it("bounds long generated prompts and moves deep content behind tabs", async () => {
    const inlineCode =
      "<style>body{background:#111;color:white;font-family:sans-serif}</style>";
    const longPrompt = `Create reviewable browser output with inline code ${inlineCode.repeat(12)}`;

    installGroupsFetch({
      groups: [
        tempGroup({
          name: "QA Generated Game Studio",
          goal_statement: longPrompt,
        }),
      ],
      outputs: { "group-temp": [documentArtifact({ title: "Dot Dodge" })] },
    });

    render(<GroupManagementPanel initialSelectedGroupId="group-temp" />);

    await waitFor(() =>
      expect(
        screen.getByRole("heading", { name: "QA Generated Game Studio" }),
      ).toBeDefined(),
    );
    const goalSummary = screen.getByTestId("groups-goal-summary");
    expect(goalSummary.textContent).toContain("<style>body");
    expect(goalSummary.className).toContain("max-h-32");
    expect(goalSummary.className).toContain("overflow-y-auto");
    expect(screen.queryByText("Dot Dodge")).toBeNull();

    fireEvent.click(screen.getByRole("tab", { name: /Outputs/i }));
    expect(screen.getByText("Dot Dodge")).toBeDefined();
  });
});
