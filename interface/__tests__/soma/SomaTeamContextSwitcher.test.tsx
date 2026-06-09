import { fireEvent, render, screen, within } from "@testing-library/react";
import { describe, expect, it, vi } from "vitest";
import { SomaTeamContextSwitcher } from "@/components/soma/SomaTeamContextSwitcher";
import type { TeamDetailEntry, TeamWorkItem } from "@/store/useCortexStore";

function makeTeam(index: number): TeamDetailEntry {
  return {
    id: `team-${index}`,
    name: `Workflow ${index}`,
  } as TeamDetailEntry;
}

function makeWorkItem(index: number): TeamWorkItem {
  return {
    id: `work-${index}`,
    title: `Work ${index}`,
    state: index === 2 ? "degraded" : "output_ready",
    teamIds: [`team-${index}`],
    outputCount: index % 2 === 0 ? 1 : 0,
    needsOperator: index === 2,
    ownerLabel: `Workflow ${index}`,
    scopeLabel: `Workflow ${index}`,
    interactions: [],
  } as TeamWorkItem;
}

describe("SomaTeamContextSwitcher", () => {
  it("uses a bounded workflow picker instead of horizontally growing tabs", () => {
    const onRootSelect = vi.fn();
    const onTeamSelect = vi.fn();
    const teams = Array.from({ length: 12 }, (_, index) => makeTeam(index + 1));
    const workItems = Array.from({ length: 12 }, (_, index) => makeWorkItem(index + 1));

    render(
      <SomaTeamContextSwitcher
        teams={teams}
        workItems={workItems}
        focusedTeamId={null}
        onRootSelect={onRootSelect}
        onTeamSelect={onTeamSelect}
      />,
    );

    expect(screen.queryByRole("tablist")).toBeNull();
    expect(screen.getByRole("button", { name: /Soma root/i }).getAttribute("aria-expanded")).toBe("false");

    fireEvent.click(screen.getByRole("button", { name: /Soma root/i }));

    const listbox = screen.getByRole("listbox", { name: /choose current workflow/i });
    expect(screen.getByRole("button", { name: /Soma root/i }).getAttribute("aria-expanded")).toBe("true");
    expect(within(listbox).getAllByRole("option").length).toBe(13);
    expect(within(listbox).getByRole("option", { name: /Workflow 12/i })).toBeDefined();

    fireEvent.click(within(listbox).getByRole("option", { name: /Workflow 2/i }));

    expect(onTeamSelect).toHaveBeenCalledWith("team-2");
    expect(onRootSelect).not.toHaveBeenCalled();
  });

  it("keeps a focused team recoverable even before team metadata is hydrated", () => {
    const onRootSelect = vi.fn();

    render(
      <SomaTeamContextSwitcher
        teams={[]}
        workItems={[]}
        focusedTeamId="missing-team"
        onRootSelect={onRootSelect}
        onTeamSelect={vi.fn()}
      />,
    );

    expect(screen.getByTestId("soma-team-context-switcher").textContent).toContain("Focused team");

    fireEvent.click(screen.getByRole("button", { name: /Focused team/i }));
    fireEvent.click(screen.getByRole("option", { name: /Soma root/i }));

    expect(onRootSelect).toHaveBeenCalledTimes(1);
  });

  it("hydrates from no visible teams into the workflow picker without changing hook order", () => {
    const onTeamSelect = vi.fn();
    const { rerender } = render(
      <SomaTeamContextSwitcher
        teams={[]}
        workItems={[]}
        focusedTeamId={null}
        onRootSelect={vi.fn()}
        onTeamSelect={onTeamSelect}
      />,
    );

    expect(screen.queryByTestId("soma-team-context-switcher")).toBeNull();

    rerender(
      <SomaTeamContextSwitcher
        teams={[makeTeam(1)]}
        workItems={[makeWorkItem(1)]}
        focusedTeamId={null}
        onRootSelect={vi.fn()}
        onTeamSelect={onTeamSelect}
      />,
    );

    fireEvent.click(screen.getByRole("button", { name: /Soma root/i }));
    fireEvent.click(screen.getByRole("option", { name: /Workflow 1/i }));

    expect(onTeamSelect).toHaveBeenCalledWith("team-1");
  });
});
