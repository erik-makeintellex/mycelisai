import { describe, it, expect, vi } from "vitest";
import { render, screen, fireEvent } from "@testing-library/react";
import FocusModeToggle from "@/components/dashboard/FocusModeToggle";

describe("FocusModeToggle", () => {
    it("renders focus-off state and triggers toggle", () => {
        const onToggle = vi.fn();
        render(<FocusModeToggle focused={false} onToggle={onToggle} />);

        const button = screen.getByRole("button", { name: "Focus Off" });
        expect(button).toBeDefined();
        expect(button.getAttribute("title")).toContain("Toggle Focus Mode");

        fireEvent.click(button);
        expect(onToggle).toHaveBeenCalledTimes(1);
    });

    it("renders focus-on state", () => {
        render(<FocusModeToggle focused onToggle={() => {}} />);
        expect(screen.getByRole("button", { name: "Focus On" })).toBeDefined();
        expect(screen.queryByRole("button", { name: "Focus Off" })).toBeNull();
    });
});
