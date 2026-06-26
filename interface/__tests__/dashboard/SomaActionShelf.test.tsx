import { fireEvent, render, screen } from "@testing-library/react";
import { beforeEach, describe, expect, it, vi } from "vitest";
import { SomaActionShelf } from "@/components/soma/SomaActionShelf";

describe("SomaActionShelf", () => {
  beforeEach(() => {
    window.localStorage.clear();
  });

  it("saves a reusable quick action without sending a chat prompt immediately", () => {
    const onRunAction = vi.fn();
    render(<SomaActionShelf onRunAction={onRunAction} />);

    fireEvent.click(screen.getByRole("button", { name: /Create new quick action/i }));
    fireEvent.change(screen.getByLabelText("Button label"), { target: { value: "Client risk brief" } });
    fireEvent.change(screen.getByLabelText("Outcome"), { target: { value: "Create a retained brief with risks and next steps" } });
    fireEvent.change(screen.getByLabelText("Output format"), { target: { value: "Markdown" } });
    fireEvent.click(screen.getByRole("button", { name: /Save action/i }));

    expect(onRunAction).not.toHaveBeenCalled();
    expect(screen.getByRole("button", { name: /Client risk brief/i })).toBeDefined();
    expect(window.localStorage.getItem("mycelis-soma-saved-actions")).toContain("Client risk brief");
  });

  it("runs saved actions as natural Soma requests", () => {
    window.localStorage.setItem("mycelis-soma-saved-actions", JSON.stringify([
      {
        label: "Client risk brief",
        prompt: "Quick action: Client risk brief. Outcome: Create a retained brief.",
        userSaved: true,
      },
    ]));
    const onRunAction = vi.fn();

    render(<SomaActionShelf onRunAction={onRunAction} />);
    fireEvent.click(screen.getByRole("button", { name: /Client risk brief/i }));

    expect(onRunAction).toHaveBeenCalledWith(expect.stringContaining("Quick action: Client risk brief"));
  });
});
