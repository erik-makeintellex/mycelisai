import { fireEvent, render, screen } from "@testing-library/react";
import { describe, expect, it, vi } from "vitest";

import { GroupsHeader } from "@/components/teams/GroupsHeader";

describe("GroupsHeader", () => {
  it("leads with work and results while keeping group setup advanced", () => {
    const onRefresh = vi.fn();

    render(
      <GroupsHeader
        monitor={{ status: "online" }}
        lifecycleReport={null}
        refreshing={false}
        archivingExpired={false}
        onArchiveExpired={vi.fn()}
        onRefresh={onRefresh}
      />,
    );

    expect(screen.getByText("Work")).toBeDefined();
    expect(screen.getByText("Review active work and delivered results.")).toBeDefined();
    expect(screen.getByText(/Group and team details stay available here for advanced review/i)).toBeDefined();
    expect(screen.getByRole("link", { name: /Open Soma/i }).getAttribute("href")).toBe("/dashboard");
    expect(screen.getByRole("link", { name: /Advanced group setup/i }).getAttribute("href")).toBe("/groups?panel=create");
    expect(screen.getByText("Advanced group monitor is online.")).toBeDefined();

    fireEvent.click(screen.getByRole("button", { name: "Refresh" }));
    expect(onRefresh).toHaveBeenCalledOnce();
  });
});
