import { describe, expect, it } from "vitest";
import { fireEvent, render, screen } from "@testing-library/react";
import { OutputCanvas } from "@/components/soma/OutputCanvas";

describe("OutputCanvas", () => {
  it("gives retained work the main surface and returns to the exact Soma context", () => {
    render(
      <OutputCanvas
        label="Playable game"
        source="/api/v1/workspace/files/view?path=groups%2Fgame-team%2Fgenerated%2Fgame%2Findex.html"
        storagePath="groups/game-team/generated/game/index.html"
        returnTo="/dashboard?team_id=game-team&outcome_id=outcome-7#latest"
        proofArtifactId="proof-7"
      />,
    );

    const backLinks = screen.getAllByRole("link", { name: "Back to Soma" });
    expect(backLinks).toHaveLength(1);
    expect(backLinks[0].getAttribute("href")).toBe("/dashboard?team_id=game-team&outcome_id=outcome-7#latest");
    expect(screen.getByTitle("Playable game").getAttribute("src")).toBe(
      "/api/v1/workspace/files/view?path=groups%2Fgame-team%2Fgenerated%2Fgame%2Findex.html",
    );
    expect(screen.getByRole("button", { name: "Open separately Playable game in a new browser window" })).toBeDefined();

    fireEvent.click(screen.getByText("File details"));
    expect(screen.getByText("groups/game-team/generated/game/index.html")).toBeDefined();
    expect(screen.getByText("proof proof-7")).toBeDefined();
    expect(screen.getByRole("button", { name: /Open local folder for Playable game/i })).toBeDefined();
  });

  it("keeps unavailable output neutral and does not present backend noise as degradation", () => {
    render(
      <OutputCanvas
        label="External result"
        source="https://example.com/result.html"
        returnTo="https://example.com/dashboard"
      />,
    );

    expect(screen.getByText("This output cannot be displayed")).toBeDefined();
    expect(screen.getByRole("link", { name: "Back to Soma" }).getAttribute("href")).toBe("/dashboard");
    expect(screen.queryByText(/degraded/i)).toBeNull();
    expect(screen.queryByText(/backend/i)).toBeNull();
    expect(screen.queryByTitle("External result")).toBeNull();
  });
});
