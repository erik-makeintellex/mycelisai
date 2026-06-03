import { describe, expect, it } from "vitest";
import { fireEvent, render, screen, within } from "@testing-library/react";
import { SomaWorkspaceFrame } from "@/components/soma/SomaWorkspaceFrame";
import { OutputWorkbench } from "@/components/soma/OutputWorkbench";

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
    expect(toggle.textContent).toContain("Review output");
    expect(toggle.textContent).toContain("4");
    fireEvent.click(toggle);

    expect(toggle.getAttribute("aria-expanded")).toBe("true");
    expect(sideRail.getAttribute("aria-hidden")).toBe("false");
    expect(within(sideRail).getByRole("tab", { name: /Work/i })).toBeDefined();
    expect(within(sideRail).getByRole("tab", { name: /Output/i })).toBeDefined();
    expect(within(sideRail).getByRole("tab", { name: /Trust/i })).toBeDefined();
    expect(within(sideRail).getByRole("tab", { name: /Context/i })).toBeDefined();
    expect(within(frame).getByText("Conversation transcript")).toBeDefined();
    expect(within(sideRail).getByText("Output package")).toBeDefined();
    expect(within(sideRail).queryByText("Active lane fallback")).toBeNull();

    fireEvent.click(within(sideRail).getByRole("tab", { name: /Work/i }));
    expect(within(sideRail).getByText("Work to review")).toBeDefined();
    expect(within(frame).getAllByText(/Work that needs a decision, check, or follow-up/i).length).toBeGreaterThan(0);
    expect(within(sideRail).getByText("Active lane fallback")).toBeDefined();
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

  it("surfaces the latest output digest before opening the review rail", () => {
    render(
      <SomaWorkspaceFrame
        expression={<div>Conversation transcript</div>}
        activeWork={<div>Active lane fallback</div>}
        output={(
          <OutputWorkbench
            outputs={[{
              text: "Owner note",
              url: "/api/v1/workspace/files/view?path=generated%2Fowner-note.md",
            }]}
          />
        )}
      />,
    );

    const digest = within(screen.getByTestId("soma-workbench-output-digest"));
    expect(digest.getByText("Latest output")).toBeDefined();
    expect(digest.getByText("Owner note")).toBeDefined();
    expect(digest.getByText("generated/owner-note.md")).toBeDefined();
    expect(digest.getByRole("button", { name: /Open file Owner note/i })).toBeDefined();
    expect(digest.getByRole("button", { name: /Open local folder for Owner note/i })).toBeDefined();

    const toggle = screen.getByTestId("soma-workbench-panel-toggle");
    expect(toggle.textContent).toContain("Review output");
    expect(toggle.textContent).toContain("1");

    fireEvent.click(toggle);

    const sideRail = screen.getByTestId("soma-workbench-side-rail");
    expect(within(sideRail).getByRole("tab", { name: /Output/i }).getAttribute("aria-selected")).toBe("true");
    expect(within(sideRail).getByTestId("output-workbench")).toBeDefined();
    expect(within(sideRail).getByText("Owner note")).toBeDefined();
    expect(within(sideRail).queryByText("Active lane fallback")).toBeNull();
  });

  it("opens work first and hides the compact output digest when active work needs attention", () => {
    render(
      <SomaWorkspaceFrame
        expression={<div>Conversation transcript</div>}
        activeWork={<div>Running team task</div>}
        output={(
          <OutputWorkbench
            outputs={[{
              text: "Generated page",
              url: "/api/v1/workspace/files/view?path=generated%2Fpage.html",
            }]}
          />
        )}
        primaryPanel="work"
        reviewCount={2}
        showOutputDigest={false}
      />,
    );

    expect(screen.queryByTestId("soma-workbench-output-digest")).toBeNull();
    const toggle = screen.getByTestId("soma-workbench-panel-toggle");
    expect(toggle.textContent).toContain("Review work");
    expect(toggle.textContent).toContain("2");

    fireEvent.click(toggle);

    const sideRail = screen.getByTestId("soma-workbench-side-rail");
    expect(within(sideRail).getByRole("tab", { name: /Work/i }).getAttribute("aria-selected")).toBe("true");
    expect(within(sideRail).getByText("Running team task")).toBeDefined();
    expect(within(sideRail).queryByText("Generated page")).toBeNull();
  });
});
