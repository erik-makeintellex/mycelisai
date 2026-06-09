import { describe, expect, it, vi } from "vitest";
import { fireEvent, render, screen } from "@testing-library/react";

const mockAdvancedMode = vi.fn(() => false);
const mockToggleAdvancedMode = vi.fn();

vi.mock("@/store/useCortexStore", () => ({
  useCortexStore: (selector: any) => selector({
    advancedMode: mockAdvancedMode(),
    toggleAdvancedMode: mockToggleAdvancedMode,
  }),
}));

import AdvancedModeRoute from "@/components/shared/AdvancedModeRoute";

describe("AdvancedModeRoute", () => {
  it("shows a clear advanced gate when advanced mode is off", () => {
    mockAdvancedMode.mockReturnValue(false);

    render(
      <AdvancedModeRoute
        title="Advanced area"
        summary="This area is for deeper inspection."
      >
        <div>Hidden advanced content</div>
      </AdvancedModeRoute>,
    );

    expect(screen.getByText("Advanced area")).toBeDefined();
    expect(screen.getByText("This area is for deeper inspection.")).toBeDefined();
    expect(screen.queryByText("Hidden advanced content")).toBeNull();
  });

  it("renders advanced content when advanced mode is on", () => {
    mockAdvancedMode.mockReturnValue(true);

    render(
      <AdvancedModeRoute
        title="Advanced area"
        summary="This area is for deeper inspection."
      >
        <div>Visible advanced content</div>
      </AdvancedModeRoute>,
    );

    expect(screen.getByText("Visible advanced content")).toBeDefined();
    expect(screen.queryByText("Advanced area")).toBeNull();
  });

  it("lets the operator open advanced mode from the gate", () => {
    mockAdvancedMode.mockReturnValue(false);
    mockToggleAdvancedMode.mockClear();

    render(
      <AdvancedModeRoute
        title="Advanced area"
        summary="This area is for deeper inspection."
      >
        <div>Hidden advanced content</div>
      </AdvancedModeRoute>,
    );

    fireEvent.click(screen.getByRole("button", { name: "Open Advanced mode" }));

    expect(mockToggleAdvancedMode).toHaveBeenCalledTimes(1);
  });
});
