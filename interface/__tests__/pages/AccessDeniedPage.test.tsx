import { describe, expect, it } from "vitest";
import { render, screen } from "@testing-library/react";

import AccessDenied from "@/components/access/AccessDenied";
import AccessDeniedPage from "@/app/(app)/access-denied/page";

describe("AccessDenied", () => {
    it("renders the default access-denied surface with safe navigation", () => {
        render(<AccessDenied />);

        expect(screen.getByRole("heading", { name: "Access denied" })).toBeDefined();
        expect(screen.getByText("Permission required")).toBeDefined();
        expect(screen.getByText("Authorized workspace role")).toBeDefined();
        expect(screen.getByRole("link", { name: "Back to dashboard" }).getAttribute("href")).toBe("/dashboard");
        expect(screen.queryByRole("link", { name: "Request access" })).toBeNull();
    });

    it("supports route-specific copy and request access navigation", () => {
        render(
            <AccessDenied
                title="Admin area unavailable"
                message="Only owners can manage this area."
                requiredAccess="Owner"
                supportHref="/settings?tab=users"
            />,
        );

        expect(screen.getByRole("heading", { name: "Admin area unavailable" })).toBeDefined();
        expect(screen.getByText("Only owners can manage this area.")).toBeDefined();
        expect(screen.getByText("Owner")).toBeDefined();
        expect(screen.getByRole("link", { name: "Request access" }).getAttribute("href")).toBe("/settings?tab=users");
    });
});

describe("AccessDeniedPage", () => {
    it("uses the reusable access denied component for the app route", () => {
        render(<AccessDeniedPage />);

        expect(screen.getByRole("heading", { name: "Access denied" })).toBeDefined();
        expect(screen.getByText("Workspace member, administrator, or owner")).toBeDefined();
        expect(screen.getByRole("link", { name: "Request access" }).getAttribute("href")).toBe("/settings?tab=users");
    });
});
