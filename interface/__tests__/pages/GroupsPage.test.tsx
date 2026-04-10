import { describe, it, expect, vi } from "vitest";
import { render, screen } from "@testing-library/react";

const mocks = vi.hoisted(() => ({
    groupPanelMock: vi.fn(() => <div data-testid="group-management-panel">Groups Panel</div>),
}));

vi.mock("@/components/teams/GroupManagementPanel", () => ({
    __esModule: true,
    default: mocks.groupPanelMock,
}));

import GroupsPage from "@/app/(app)/groups/page";

describe("GroupsPage", () => {
    it("renders the dedicated groups workspace surface", async () => {
        render(await GroupsPage({}));
        expect(screen.getByTestId("group-management-panel")).toBeDefined();
    });

    it("passes the requested group id into the groups workspace", async () => {
        render(await GroupsPage({ searchParams: Promise.resolve({ group_id: "group-temp" }) }));
        expect(mocks.groupPanelMock).toHaveBeenCalledWith(
            expect.objectContaining({ initialSelectedGroupId: "group-temp" }),
            undefined,
        );
    });
});
