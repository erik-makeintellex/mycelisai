import { describe, expect, it } from "vitest";
import { render, screen, within } from "@testing-library/react";
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
    expect(within(frame).getByText("Expression")).toBeDefined();
    expect(within(frame).getByText("Active work")).toBeDefined();
    expect(within(frame).getByText("Output")).toBeDefined();
    expect(within(frame).getByText("Trust")).toBeDefined();
    expect(within(frame).getByText("Context")).toBeDefined();
    expect(within(frame).getByText("Conversation transcript")).toBeDefined();
    expect(within(frame).getByText("Compact trust package")).toBeDefined();
  });
});
