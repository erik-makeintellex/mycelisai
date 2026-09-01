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
        output={(
          <OutputWorkbench
            outputs={[{
              text: "Output package",
              url: "/api/v1/workspace/files/view?path=generated%2Foutput-package.md",
            }]}
          />
        )}
        trust={<div>Compact trust package</div>}
        context={<div>Context links</div>}
      />,
    );

    const frame = screen.getByTestId("soma-workspace-frame");
    const sideRail = screen.getByTestId("soma-workbench-side-rail");
    const toggle = screen.getByTestId("soma-workbench-panel-toggle");

    expect(toggle.getAttribute("aria-expanded")).toBe("false");
    expect(sideRail.getAttribute("aria-hidden")).toBe("true");
    expect(sideRail.className).toContain("520px");
    expect(sideRail.className).toContain("inset-x-2");
    expect(within(frame).queryByText("Expression")).toBeNull();
    expect(toggle.textContent).toContain("Review output");
    expect(toggle.textContent).toContain("1");
    expect(screen.getByTestId("soma-workbench-output-digest")).toBeDefined();
    expect(within(frame).queryByTestId("output-workbench")).toBeNull();
    expect(within(frame).queryByText("Active lane fallback")).toBeNull();
    expect(within(frame).queryByText("Compact trust package")).toBeNull();
    fireEvent.click(toggle);

    expect(toggle.getAttribute("aria-expanded")).toBe("true");
    expect(sideRail.getAttribute("aria-hidden")).toBe("false");
    expect(sideRail.getAttribute("data-review-surface")).toBe("output");
    expect(sideRail.className).toContain("100vw-20rem");
    expect(screen.getByTestId("soma-workbench-panel-scroll").className).toContain("overflow-x-hidden");
    expect(within(sideRail).getByRole("tab", { name: /Work/i })).toBeDefined();
    expect(within(sideRail).getByRole("tab", { name: /Output/i })).toBeDefined();
    expect(within(sideRail).getByRole("tab", { name: /Trust/i })).toBeDefined();
    expect(within(sideRail).getByRole("tab", { name: /Context/i })).toBeDefined();
    expect(within(frame).getByText("Conversation transcript")).toBeDefined();
    expect(within(sideRail).getByText("Output package")).toBeDefined();
    expect(within(sideRail).queryByText("Active lane fallback")).toBeNull();

    fireEvent.click(within(sideRail).getByRole("tab", { name: /Work/i }));
    expect(within(sideRail).getByText("Current work")).toBeDefined();
    expect(within(frame).getAllByText(/Work that is queued, running, or being checked/i).length).toBeGreaterThan(0);
    expect(within(sideRail).getByText("Active lane fallback")).toBeDefined();
    fireEvent.click(within(sideRail).getByRole("tab", { name: /Trust/i }));
    expect(within(sideRail).getByText("Compact trust package")).toBeDefined();
    fireEvent.click(within(sideRail).getByRole("tab", { name: /Context/i }));
    expect(within(sideRail).getByText("Context links")).toBeDefined();
  });

  it("makes a trusted retained result the main stage while keeping Soma available", () => {
    render(
      <SomaWorkspaceFrame
        expression={<div>Continue with Soma</div>}
        activeWork={<div>Background validation</div>}
        output={(
          <OutputWorkbench
            outputs={[{
              text: "Retained result",
              url: "/api/v1/workspace/files/view?path=generated%2Fresult.html",
            }]}
          />
        )}
        primaryPanel="output"
        resultFirst
        showOutputDigest={false}
        workItemCount={1}
        workStatus="active"
      />,
    );

    const stage = screen.getByTestId("soma-result-first-stage");
    expect(within(stage).getByTestId("output-workbench")).toBeDefined();
    expect(within(stage).getByText("Retained result")).toBeDefined();
    expect(within(screen.getByTestId("soma-continuation-sidecar")).getByText("Continue with Soma")).toBeDefined();
    expect(screen.queryByTestId("soma-workbench-output-digest")).toBeNull();

    const lane = within(screen.getByTestId("soma-current-work-lane"));
    expect(lane.getByText("Work in progress")).toBeDefined();
    expect(lane.queryByText("Work needs review")).toBeNull();
    expect(lane.getByRole("button", { name: /View work/i })).toBeDefined();
    expect(lane.getByLabelText("1 work item")).toBeDefined();

    fireEvent.click(lane.getByRole("button", { name: /View work/i }));
    const sideRail = screen.getByTestId("soma-workbench-side-rail");
    expect(within(sideRail).getAllByText("Current work").length).toBeGreaterThan(0);
    expect(within(sideRail).getByText("Background validation")).toBeDefined();
  });

  it("labels operator-owned work as input needed", () => {
    render(
      <SomaWorkspaceFrame
        expression={<div>Conversation transcript</div>}
        activeWork={<div>Approval question</div>}
        primaryPanel="work"
        reviewCount={1}
        workItemCount={1}
        workStatus="needs_input"
      />,
    );

    const lane = within(screen.getByTestId("soma-current-work-lane"));
    expect(lane.getByText("Input needed")).toBeDefined();
    expect(lane.getByRole("button", { name: /Respond/i })).toBeDefined();
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
    expect(screen.getByTestId("soma-output-digest-layout").className).toContain("flex-col");
    expect(digest.getByText("Latest:")).toBeDefined();
    expect(digest.getByText("Owner note")).toBeDefined();
    expect(digest.queryByText("generated/owner-note.md")).toBeNull();
    expect(digest.getByRole("button", { name: /Open file Owner note/i })).toBeDefined();
    fireEvent.click(digest.getByText("Details and follow-up"));
    expect(digest.getByRole("button", { name: /Open local folder for Owner note/i })).toBeDefined();
    expect(screen.queryByTestId("output-workbench")).toBeNull();

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

  it("contains review focus, dismisses on Escape, and restores the opener", () => {
    render(
      <SomaWorkspaceFrame
        expression={<button type="button">Open Outcome Vault</button>}
        activeWork={<button type="button">Recover work</button>}
        output={(
          <OutputWorkbench
            outputs={[{
              text: "Retained output",
              url: "/api/v1/workspace/files/view?path=generated%2Fretained-output.md",
            }]}
          />
        )}
        trust={<div>Trust details</div>}
        context={<div>Context details</div>}
      />,
    );

    const toggle = screen.getByTestId("soma-workbench-panel-toggle");
    const sideRail = screen.getByTestId("soma-workbench-side-rail");
    toggle.focus();
    fireEvent.click(toggle);

    const closeButton = within(sideRail).getByRole("button", { name: "Close work panel" });
    expect(sideRail.getAttribute("role")).toBe("dialog");
    expect(sideRail.getAttribute("aria-modal")).toBe("true");
    expect(document.activeElement).toBe(closeButton);

    fireEvent.keyDown(document, { key: "Tab", shiftKey: true });
    expect(sideRail.contains(document.activeElement)).toBe(true);
    const lastFocusedElement = document.activeElement;
    fireEvent.keyDown(document, { key: "Tab" });
    expect(document.activeElement).toBe(closeButton);
    expect(lastFocusedElement).not.toBe(closeButton);
    fireEvent.keyDown(document, { key: "Escape" });

    expect(sideRail.getAttribute("aria-hidden")).toBe("true");
    expect(sideRail.getAttribute("role")).toBeNull();
    expect(sideRail.getAttribute("aria-modal")).toBeNull();
    expect(sideRail.className).toContain("invisible");
    expect(sideRail.className).toContain("pointer-events-none");
    expect(document.activeElement).toBe(toggle);

    const vaultButton = screen.getByRole("button", { name: "Open Outcome Vault" });
    vaultButton.focus();
    fireEvent.click(vaultButton);
    expect(document.activeElement).toBe(vaultButton);
  });

  it("opens work first while keeping latest output accessible when active work needs attention", () => {
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
    workItemCount={2}
    workStatus="needs_review"
      />,
    );

    const lane = screen.getByTestId("soma-current-work-lane");
    expect(within(lane).getByText("Work needs review")).toBeDefined();
    expect(within(lane).getByText("Generated page")).toBeDefined();
    fireEvent.click(within(lane).getByText("Details and follow-up"));
    expect(within(lane).getByRole("button", { name: /Open local folder for Generated page/i })).toBeDefined();
    const toggle = screen.getByTestId("soma-workbench-panel-toggle");
    expect(toggle.textContent).toContain("Review work");
    expect(toggle.textContent).toContain("2");

    fireEvent.click(toggle);

    const sideRail = screen.getByTestId("soma-workbench-side-rail");
    expect(within(sideRail).getAllByText("Work to review").length).toBeGreaterThan(0);
    expect(within(sideRail).getByText(/Understand the issue/i)).toBeDefined();
    expect(within(sideRail).queryByRole("tab", { name: /Work/i })).toBeNull();
    expect(within(sideRail).getByRole("link", { name: /Open inbox/i }).getAttribute("href")).toBe("/teams?view=work");
    expect(within(sideRail).getByText("Running team task")).toBeDefined();
    expect(within(sideRail).queryByText("Generated page")).toBeNull();
  });
});
