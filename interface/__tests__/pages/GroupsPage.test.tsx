import { describe, it, expect, vi, beforeEach } from "vitest";
import { render, screen } from "@testing-library/react";

const mocks = vi.hoisted(() => ({
    groupPanelMock: vi.fn(() => <div data-testid="group-management-panel">Groups Panel</div>),
    searchParams: new URLSearchParams(),
}));

vi.mock("@/components/teams/GroupManagementPanel", () => ({
    __esModule: true,
    default: mocks.groupPanelMock,
}));

vi.mock("next/navigation", () => ({
    usePathname: () => "/groups",
    useSearchParams: () => mocks.searchParams,
}));

import GroupsPage from "@/app/(app)/groups/page";

describe("GroupsPage", () => {
    beforeEach(() => {
        mocks.groupPanelMock.mockClear();
        mocks.searchParams.delete("advanced");
    });

    it("renders directly as a main operator workspace", async () => {
        render(await GroupsPage({}));

        expect(screen.queryByText(/Advanced coordination view/i)).toBeNull();
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
