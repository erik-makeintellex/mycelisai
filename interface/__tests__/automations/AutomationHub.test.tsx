import { describe, it, expect, vi } from "vitest";
import { render, screen, fireEvent } from "@testing-library/react";

vi.mock("@/components/automations/MissionProfileWizard", () => ({
    __esModule: true,
    default: () => <div data-testid="mission-profile-wizard">Mission Profile Wizard</div>,
}));

import AutomationHub from "@/components/automations/AutomationHub";

describe("AutomationHub", () => {
    it("renders baseline CTA and routes primary setup action", () => {
        const openTab = vi.fn();
        render(<AutomationHub openTab={openTab} advancedMode={false} />);

        expect(screen.getByTestId("automations-hub-baseline")).toBeDefined();
        fireEvent.click(screen.getByRole("button", { name: /Set Up Your First Automation Chain/i }));
        expect(openTab).toHaveBeenCalledWith("triggers");
    });

    it("toggles mission profile wizard visibility", () => {
        render(<AutomationHub openTab={() => {}} advancedMode={false} />);

        expect(screen.queryByTestId("mission-profile-wizard")).toBeNull();
        fireEvent.click(screen.getByTestId("open-mission-profile-wizard"));
        expect(screen.getByTestId("mission-profile-wizard")).toBeDefined();

        fireEvent.click(screen.getByTestId("open-mission-profile-wizard"));
        expect(screen.queryByTestId("mission-profile-wizard")).toBeNull();
    });

    it("only shows advanced workflow cards in advanced mode", () => {
        const openTab = vi.fn();
        const { rerender } = render(<AutomationHub openTab={openTab} advancedMode={false} />);
        expect(screen.queryByText("Workflow Builder")).toBeNull();
        expect(screen.queryByText("Shared Teams")).toBeNull();
        expect(screen.queryByText("Coming Soon")).toBeNull();
        expect(screen.getByText("Team workstreams")).toBeDefined();
        expect(screen.getByText("Review loop visibility")).toBeDefined();

        rerender(<AutomationHub openTab={openTab} advancedMode />);
        expect(screen.getByText("Workflow Builder")).toBeDefined();
        expect(screen.queryByText("Shared Teams")).toBeNull();

        const createButtons = screen.getAllByRole("button", { name: "Create" });
        const create = createButtons[createButtons.length - 1];
        fireEvent.click(create);
        expect(openTab).toHaveBeenCalledWith("wiring");
    });
});
