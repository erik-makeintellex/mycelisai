import { describe, it, expect, vi, beforeEach } from "vitest";
import { render, screen, fireEvent, waitFor } from "@testing-library/react";
import { mockFetch } from "../setup";
import UsersPage from "@/components/settings/UsersPage";

vi.mock("@/components/teams/GroupManagementPanel", () => ({
    __esModule: true,
    default: () => <div data-testid="group-management-panel">GroupManagementPanel</div>,
}));

function mockMeResponse({
    role = "owner",
    accessManagementTier = "release",
}: {
    role?: string;
    accessManagementTier?: "release" | "enterprise";
} = {}) {
    mockFetch.mockResolvedValue({
        ok: true,
        json: async () => ({
            ok: true,
            data: {
                id: "me-1",
                email: "me@local",
                role,
                name: "Current User",
                settings: {
                    access_management_tier: accessManagementTier,
                },
            },
        }),
    });
}

describe("UsersPage", () => {
    beforeEach(() => {
        mockFetch.mockReset();
        mockMeResponse();
    });

    it("keeps base release focused on organization access and groups", async () => {
        render(<UsersPage />);

        await waitFor(() => {
            expect(screen.getByTestId("organization-access-layer")).toBeDefined();
        });

        expect(screen.getByText("Base release layer")).toBeDefined();
        expect(screen.getByTestId("enterprise-directory-locked")).toBeDefined();
        expect(screen.queryByTestId("users-management-panel")).toBeNull();
        expect(screen.getByTestId("users-groups-section")).toBeDefined();
        expect(screen.getByTestId("group-management-panel")).toBeDefined();
    });

    it("maps backend admin identity to owner access", async () => {
        mockMeResponse({ role: "admin", accessManagementTier: "release" });

        render(<UsersPage />);

        await waitFor(() => {
            expect(screen.getByText(/Current User is currently mapped as owner/i)).toBeDefined();
        });
    });

    it("shows enterprise user management for an owner and allows adding a local user", async () => {
        mockMeResponse({ role: "admin", accessManagementTier: "enterprise" });

        render(<UsersPage />);

        await waitFor(() => {
            expect(screen.getByTestId("users-management-panel")).toBeDefined();
        });

        fireEvent.change(screen.getByLabelText("Name"), { target: { value: "Alex" } });
        fireEvent.change(screen.getByLabelText("Email"), { target: { value: "alex@example.com" } });
        fireEvent.click(screen.getByTestId("users-add-button"));

        await waitFor(() => {
            expect(screen.getByText("alex@example.com")).toBeDefined();
        });
    });

    it("keeps enterprise directory read-only for non-owner roles", async () => {
        mockMeResponse({ role: "operator", accessManagementTier: "enterprise" });

        render(<UsersPage />);

        await waitFor(() => {
            expect(screen.getByTestId("enterprise-directory-readonly")).toBeDefined();
        });

        expect(screen.queryByTestId("users-management-panel")).toBeNull();
        expect(screen.queryByTestId("users-add-button")).toBeNull();
    });
});
