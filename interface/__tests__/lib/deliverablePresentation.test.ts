import { describe, expect, it } from "vitest";
import { deliverablePresentation } from "@/lib/deliverablePresentation";

describe("deliverablePresentation", () => {
  it.each([
    [{ outputContract: { primary_deliverable: "Playable browser game", shape: "app_package" }, kind: "project_package" }, "game", "Play game"],
    [{ outputContract: { shape: "app_package" }, kind: "project_package" }, "app", "Open app"],
    [{ kind: "project_package", title: "Playable browser game" }, "game", "Play game"],
    [{ kind: "project_package", title: "Retained package" }, "app", "Open app"],
    [{ kind: "report" }, "report", "Read report"],
    [{ type: "document" }, "document", "Review document"],
    [{ outputContract: { primary_deliverable: "CSV report", shape: "table" } }, "data", "View data"],
    [{ contentType: "video/mp4" }, "media", "Play media"],
    [{ kind: "image" }, "image", "View image"],
  ] as const)("uses explicit metadata for %s", (input, kind, actionLabel) => {
    expect(deliverablePresentation(input)).toEqual({ kind, actionLabel });
  });

  it.each([
    ["generated/app/index.html", "app", "Open app"],
    ["reports/review.md", "document", "Review document"],
    ["exports/results.parquet", "data", "View data"],
    ["media/preview.png", "image", "View image"],
    ["media/walkthrough.webm", "media", "Play media"],
    ["generated/result.bin", "result", "Open result"],
  ] as const)("falls back to an unambiguous path when metadata is absent", (path, kind, actionLabel) => {
    expect(deliverablePresentation({ path })).toEqual({ kind, actionLabel });
  });

  it("does not override unknown explicit metadata with a path guess", () => {
    expect(deliverablePresentation({ kind: "file", path: "generated/app/index.html" })).toEqual({
      kind: "result",
      actionLabel: "Open result",
    });
  });

  it("keeps an unknown retained result generic", () => {
    expect(deliverablePresentation({ kind: "unknown", title: "Retained outcome" })).toEqual({
      kind: "result",
      actionLabel: "Open result",
    });
  });
});
