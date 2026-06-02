import { describe, expect, it } from "vitest";
import { fireEvent, render, screen, within } from "@testing-library/react";
import { SomaWorkspaceFrame } from "@/components/soma/SomaWorkspaceFrame";

describe("SomaWorkspaceFrame", () => {
  it("renders the bounded Soma workspace slots without requiring runtime APIs", () => {
    render(
      <SomaWorkspaceFrame
        expression={<div>Conversation transcript</div>}
        activeWork={<div>Active lane fallback</div>}
        output={<div>Output package</div>}
        trust={<div>Compact trust package</div>}
        context={<div>Context links</div>}
      />,
    );

    const frame = screen.getByTestId("soma-workspace-frame");
    const sideRail = screen.getByTestId("soma-workbench-side-rail");
    const toggle = screen.getByTestId("soma-workbench-panel-toggle");

    expect(toggle.getAttribute("aria-expanded")).toBe("false");
    expect(sideRail.getAttribute("aria-hidden")).toBe("true");
    expect(within(frame).queryByText("Expression")).toBeNull();
    fireEvent.click(toggle);

    expect(toggle.getAttribute("aria-expanded")).toBe("true");
    expect(sideRail.getAttribute("aria-hidden")).toBe("false");
    expect(within(sideRail).getByRole("tab", { name: /Work/i })).toBeDefined();
    expect(within(sideRail).getByRole("tab", { name: /Output/i })).toBeDefined();
    expect(within(sideRail).getByRole("tab", { name: /Trust/i })).toBeDefined();
    expect(within(sideRail).getByRole("tab", { name: /Context/i })).toBeDefined();
    expect(within(sideRail).getByText("Active work")).toBeDefined();
    expect(within(frame).getAllByText(/Current work that needs review or follow-up/i).length).toBeGreaterThan(0);
    expect(within(frame).getByText("Conversation transcript")).toBeDefined();
    expect(within(sideRail).getByText("Active lane fallback")).toBeDefined();
    expect(within(sideRail).queryByText("Output package")).toBeNull();

    fireEvent.click(within(sideRail).getByRole("tab", { name: /Output/i }));
    expect(within(sideRail).getByText("Output package")).toBeDefined();
    fireEvent.click(within(sideRail).getByRole("tab", { name: /Trust/i }));
    expect(within(sideRail).getByText("Compact trust package")).toBeDefined();
    fireEvent.click(within(sideRail).getByRole("tab", { name: /Context/i }));
    expect(within(sideRail).getByText("Context links")).toBeDefined();
  });

  it("keeps the first-run Soma surface focused when there is nothing to review yet", () => {
    render(<SomaWorkspaceFrame expression={<div>Ask Soma anything</div>} />);

    expect(screen.getByText("Ask Soma anything")).toBeDefined();
    expect(screen.queryByTestId("soma-workbench-panel-toggle")).toBeNull();
    expect(screen.queryByTestId("soma-workbench-side-rail")).toBeNull();
    expect(screen.queryByText("Expression")).toBeNull();
  });
});
