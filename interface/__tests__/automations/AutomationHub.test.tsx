import { describe, it, expect, vi } from "vitest";
import { render, screen, fireEvent } from "@testing-library/react";

vi.mock("@/components/automations/TeamInstantiationWizard", () => ({
    __esModule: true,
    default: () => <div data-testid="team-instantiation-wizard">Team Instantiation Wizard</div>,
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

    it("toggles team instantiation wizard visibility", () => {
        render(<AutomationHub openTab={() => {}} advancedMode={false} />);

        expect(screen.queryByTestId("team-instantiation-wizard")).toBeNull();
        fireEvent.click(screen.getByTestId("open-instantiation-wizard"));
        expect(screen.getByTestId("team-instantiation-wizard")).toBeDefined();

        fireEvent.click(screen.getByTestId("open-instantiation-wizard"));
        expect(screen.queryByTestId("team-instantiation-wizard")).toBeNull();
    });

    it("only shows neural wiring card in advanced mode", () => {
        const openTab = vi.fn();
        const { rerender } = render(<AutomationHub openTab={openTab} advancedMode={false} />);
        expect(screen.queryByText("Neural Wiring")).toBeNull();

        rerender(<AutomationHub openTab={openTab} advancedMode />);
        expect(screen.getByText("Neural Wiring")).toBeDefined();

        const createButtons = screen.getAllByRole("button", { name: "Create" });
        const create = createButtons[createButtons.length - 1];
        fireEvent.click(create);
        expect(openTab).toHaveBeenCalledWith("wiring");
    });
});
