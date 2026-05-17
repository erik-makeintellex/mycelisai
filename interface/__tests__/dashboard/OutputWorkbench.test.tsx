import { describe, expect, it, vi } from "vitest";
import { fireEvent, render, screen, waitFor } from "@testing-library/react";
import {
  OutputWorkbench,
  outputWorkbenchItems,
  projectPackageOutputs,
} from "@/components/soma/OutputWorkbench";
import type { ExecutionSummaryData } from "@/store/useCortexStore";

describe("OutputWorkbench", () => {
  it("extracts project packages separately from retained outputs and artifacts", () => {
    const summary: ExecutionSummaryData = {
      outputs: [
        {
          kind: "project_package",
          title: "Launch microsite",
          folder: "generated/launch",
          entrypoint: "index.html",
          files: ["index.html", "styles.css"],
        },
        { title: "Launch brief", url: "/runs/run-1" },
      ],
    };

    expect(projectPackageOutputs(summary.outputs)).toHaveLength(1);
    expect(outputWorkbenchItems(summary, [{ type: "document", title: "Launch brief", url: "/runs/run-1" }])).toEqual([
      { text: "Launch brief", url: "/runs/run-1" },
    ]);
    expect(outputWorkbenchItems(summary, [{ type: "document", title: "Operator notes", url: "/notes/1" }])).toEqual([
      { text: "Launch brief", url: "/runs/run-1" },
      { text: "Operator notes", url: "/notes/1" },
    ]);
  });

  it("renders package actions and copyable output quotes", async () => {
    const writeText = vi.fn().mockResolvedValue(undefined);
    Object.defineProperty(navigator, "clipboard", {
      value: { writeText },
      configurable: true,
    });

    render(
      <OutputWorkbench
        outputs={[{ text: "Launch brief", url: "/runs/run-1" }]}
        projectPackages={[
          {
            kind: "project_package",
            title: "Launch microsite",
            summary: "Reviewable output package",
            folder: "generated/launch",
            entrypoint: "index.html",
            files: ["index.html"],
            validation: "Smoke test passed",
          },
        ]}
      />,
    );

    expect(screen.getByTestId("output-workbench")).toBeDefined();
    expect(screen.getByText("Launch microsite")).toBeDefined();
    expect(screen.getByText("Reviewable output package")).toBeDefined();
    expect(screen.getByText("entry: index.html")).toBeDefined();
    expect(screen.getByText("Smoke test passed")).toBeDefined();
    expect(screen.getByRole("button", { name: /Open Launch brief/i })).toBeDefined();
    expect(screen.getByRole("button", { name: /Open local folder for Launch microsite/i })).toBeDefined();

    fireEvent.click(screen.getByRole("button", { name: /Copy output quote for Launch brief/i }));

    await waitFor(() => {
      expect(writeText).toHaveBeenCalledWith("> Launch brief\n/runs/run-1");
      expect(screen.getByRole("button", { name: "Copied output quote" })).toBeDefined();
    });
  });
});
