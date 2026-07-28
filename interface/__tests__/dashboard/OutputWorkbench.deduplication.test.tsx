import { render, screen } from "@testing-library/react";
import { describe, expect, it } from "vitest";
import { OutputWorkbench } from "@/components/soma/OutputWorkbench";

describe("OutputWorkbench reference deduplication", () => {
  it("removes exact package duplicates while retaining distinct supporting files", () => {
    render(
      <OutputWorkbench
        outputs={[
          {
            text: "Duplicate package folder",
            url: "/api/v1/workspace/files/view?path=workspace%2Fgenerated%2Flaunch",
            storagePath: "workspace/generated/launch",
          },
          {
            text: "Launch review notes",
            url: "/api/v1/workspace/files/view?path=workspace%2Fgenerated%2Flaunch%2Freview.md",
            storagePath: "workspace/generated/launch/review.md",
          },
        ]}
        projectPackages={[{
          kind: "project_package",
          title: "Launch package",
          folder: "generated/launch",
          entrypoint: "index.html",
        }]}
      />,
    );

    expect(screen.queryByText("Duplicate package folder")).toBeNull();
    expect(screen.getByText("Latest output").closest("article")?.textContent).toContain("Launch review notes");
    expect(screen.getByTestId("latest-output-actions").className).toContain("w-full");
  });
});
