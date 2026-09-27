import { beforeEach, describe, expect, it, vi } from "vitest";
import { fireEvent, render, screen, waitFor } from "@testing-library/react";
import { mockFetch } from "../setup";

const mockAdvancedMode = vi.fn(() => false);
const mockToggleAdvancedMode = vi.fn();
const mockSearchParams = new URLSearchParams();

vi.mock("next/navigation", () => ({
  usePathname: () => "/advanced-area",
  useSearchParams: () => mockSearchParams,
}));

vi.mock("@/store/useCortexStore", () => ({
  useCortexStore: (selector: (state: {
    advancedMode: boolean;
    toggleAdvancedMode: typeof mockToggleAdvancedMode;
  }) => unknown) => selector({
    advancedMode: mockAdvancedMode(),
    toggleAdvancedMode: mockToggleAdvancedMode,
  }),
}));

import AdvancedModeRoute from "@/components/shared/AdvancedModeRoute";

function mockSessionRole(role: "admin" | "standard") {
  mockFetch.mockResolvedValue({ ok: true, json: async () => ({ data: { user: { role } } }) });
}

describe("AdvancedModeRoute", () => {
  beforeEach(() => {
    mockSearchParams.delete("advanced");
    mockToggleAdvancedMode.mockClear();
    window.history.pushState({}, "", "/");
  });

  it("shows a clear advanced gate to an admin when advanced mode is off", async () => {
    mockAdvancedMode.mockReturnValue(false);
    mockSessionRole("admin");

    render(
      <AdvancedModeRoute
        title="Advanced area"
        summary="This area is for deeper inspection."
      >
        <div>Hidden advanced content</div>
      </AdvancedModeRoute>,
    );

    expect(await screen.findByText("Advanced area")).toBeDefined();
    expect(screen.getByText("This area is for deeper inspection.")).toBeDefined();
    expect(screen.queryByText("Hidden advanced content")).toBeNull();
  });

  it("renders advanced content to an admin when advanced mode is on", async () => {
    mockAdvancedMode.mockReturnValue(true);
    mockSessionRole("admin");

    render(
      <AdvancedModeRoute
        title="Advanced area"
        summary="This area is for deeper inspection."
      >
        <div>Visible advanced content</div>
      </AdvancedModeRoute>,
    );

    expect(await screen.findByText("Visible advanced content")).toBeDefined();
    expect(screen.queryByText("Advanced area")).toBeNull();
  });

  it("lets an admin open advanced mode from the gate", async () => {
    mockAdvancedMode.mockReturnValue(false);
    mockToggleAdvancedMode.mockClear();
    mockSessionRole("admin");

    render(
      <AdvancedModeRoute
        title="Advanced area"
        summary="This area is for deeper inspection."
      >
        <div>Hidden advanced content</div>
      </AdvancedModeRoute>,
    );

    const link = await screen.findByRole("link", { name: "Open admin tools" });
    expect(link.getAttribute("href")).toBe("/advanced-area?advanced=1");
    fireEvent.click(link);

    expect(mockToggleAdvancedMode).toHaveBeenCalledTimes(1);
  });

  it("renders advanced content to an admin when the URL asks to open advanced mode", async () => {
    mockAdvancedMode.mockReturnValue(false);
    mockSessionRole("admin");
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

  describe("standard users (never told to Turn on Admin tools)", () => {
    it("shows an admin_required blocker instead of the advanced gate", async () => {
      mockAdvancedMode.mockReturnValue(false);
      mockSessionRole("standard");

      render(
        <AdvancedModeRoute
          title="Advanced area"
          summary="This area is for deeper inspection."
        >
          <div>Hidden advanced content</div>
        </AdvancedModeRoute>,
      );

      expect(await screen.findByRole("alert")).toBeDefined();
      expect(screen.queryByText(/turn on admin tools/i)).toBeNull();
      expect(screen.queryByText("Advanced area")).toBeNull();
      expect(screen.queryByText("Hidden advanced content")).toBeNull();
    });

    it("still shows the blocker even with ?advanced=1 in the URL", async () => {
      mockAdvancedMode.mockReturnValue(false);
      mockSessionRole("standard");
      window.history.pushState({}, "", "/advanced-area?advanced=1");

      render(
        <AdvancedModeRoute
          title="Advanced area"
          summary="This area is for deeper inspection."
        >
          <div>Hidden advanced content</div>
        </AdvancedModeRoute>,
      );

      expect(await screen.findByRole("alert")).toBeDefined();
      expect(screen.queryByText("Hidden advanced content")).toBeNull();
    });
  });
});
