import { describe, it, expect, vi, beforeEach } from "vitest";
import { render, screen, fireEvent, waitFor } from "@testing-library/react";
import { mockFetch } from "../setup";
import UsersPage from "@/components/settings/UsersPage";

function mockMeResponse({
    role = "owner",
    accessManagementTier = "release",
    productEdition = "self_hosted_release",
    identityMode = "local_only",
    sharedAgentSpecificityOwner = "root_admin",
    principalType = "local_admin",
    authSource = "local_api_key",
    effectiveRole = "owner",
    breakGlass = false,
}: {
    role?: string;
    accessManagementTier?: "release" | "enterprise";
    productEdition?: "self_hosted_release" | "self_hosted_enterprise" | "hosted_control_plane";
    identityMode?: "local_only" | "hybrid" | "federated";
    sharedAgentSpecificityOwner?: "root_admin" | "delegated_owner";
    principalType?: "local_admin" | "break_glass_admin" | "federated_user" | "service_principal" | "user";
    authSource?: string;
    effectiveRole?: string;
    breakGlass?: boolean;
} = {}) {
    mockFetch.mockResolvedValue({
        ok: true,
        json: async () => ({
            ok: true,
            data: {
                id: "me-1",
                email: "me@local",
                role,
                effective_role: effectiveRole,
                name: "Current User",
                principal_type: principalType,
                auth_source: authSource,
                break_glass: breakGlass,
                settings: {
                    access_management_tier: accessManagementTier,
                    product_edition: productEdition,
                    identity_mode: identityMode,
                    shared_agent_specificity_owner: sharedAgentSpecificityOwner,
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

        expect(screen.getByTestId("deployment-access-model")).toBeDefined();
        expect(screen.getByText("Self-hosted release")).toBeDefined();
        expect(screen.getByText(/local_admin/i)).toBeDefined();
        expect(screen.getByText(/local_api_key/i)).toBeDefined();
        expect(screen.getByText("Base release layer")).toBeDefined();
        expect(screen.getByTestId("enterprise-directory-locked")).toBeDefined();
        expect(screen.queryByTestId("users-management-panel")).toBeNull();
        expect(screen.getByTestId("users-groups-section")).toBeDefined();
        expect(screen.getByRole("link", { name: /Open groups workspace/i }).getAttribute("href")).toBe("/groups");
    });

    it("keeps deployment-owned access model fields read-only even for owners", async () => {
        mockMeResponse({ role: "admin", accessManagementTier: "release" });

        render(<UsersPage />);

        await waitFor(() => {
            expect(screen.getByTestId("deployment-access-model")).toBeDefined();
        });

        expect(screen.getByText("Hosted control plane").closest("button")).toHaveProperty("disabled", true);
        expect(screen.getByLabelText("Identity Mode")).toHaveProperty("disabled", true);
        expect(screen.getByLabelText("Shared Agent Specificity Owner")).toHaveProperty("disabled", true);
        expect(screen.getByTestId("save-access-model")).toHaveProperty("disabled", true);
        expect(mockFetch).toHaveBeenCalledTimes(1);
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
        expect(screen.getByLabelText("Identity Mode")).toHaveProperty("disabled", true);
        expect(screen.getByTestId("save-access-model")).toHaveProperty("disabled", true);
    });

    it("shows break-glass principal posture when hybrid recovery is active", async () => {
        mockMeResponse({
            role: "admin",
            accessManagementTier: "enterprise",
            identityMode: "hybrid",
            principalType: "break_glass_admin",
            authSource: "local_break_glass",
            breakGlass: true,
        });

        render(<UsersPage />);

        await waitFor(() => {
            expect(screen.getByText(/break_glass_admin/i)).toBeDefined();
        });

        expect(screen.getByText(/local_break_glass/i)).toBeDefined();
        expect(screen.getByText(/break-glass recovery active/i)).toBeDefined();
    });
});
