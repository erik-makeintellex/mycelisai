import { beforeEach, describe, expect, it, vi } from "vitest";
import { fireEvent, render, screen, waitFor } from "@testing-library/react";

const mockAdvancedMode = vi.fn(() => false);
const mockToggleAdvancedMode = vi.fn();
const mockSearchParams = new URLSearchParams();

vi.mock("next/navigation", () => ({
  usePathname: () => "/advanced-area",
  useSearchParams: () => mockSearchParams,
}));

vi.mock("@/store/useCortexStore", () => ({
  useCortexStore: (selector: any) => selector({
    advancedMode: mockAdvancedMode(),
    toggleAdvancedMode: mockToggleAdvancedMode,
  }),
}));

import AdvancedModeRoute from "@/components/shared/AdvancedModeRoute";

describe("AdvancedModeRoute", () => {
  beforeEach(() => {
    mockSearchParams.delete("advanced");
    mockToggleAdvancedMode.mockClear();
    window.history.pushState({}, "", "/");
  });

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

    const link = screen.getByRole("link", { name: "Open admin tools" });
    expect(link.getAttribute("href")).toBe("/advanced-area?advanced=1");
    fireEvent.click(link);

    expect(mockToggleAdvancedMode).toHaveBeenCalledTimes(1);
  });

  it("renders advanced content when the URL asks to open advanced mode", async () => {
    mockAdvancedMode.mockReturnValue(false);
    window.history.pushState({}, "", "/advanced-area?advanced=1");

    render(
      <AdvancedModeRoute
        title="Advanced area"
        summary="This area is for deeper inspection."
      >
        <div>Visible from query</div>
      </AdvancedModeRoute>,
    );

    await waitFor(() => {
      expect(screen.getByText("Visible from query")).toBeDefined();
    });
    expect(mockToggleAdvancedMode).toHaveBeenCalled();
  });
});
