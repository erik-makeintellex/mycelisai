import { describe, it, expect, beforeEach, vi } from "vitest";
import { render, screen, fireEvent, waitFor } from "@testing-library/react";
import GroupManagementPanel from "@/components/teams/GroupManagementPanel";
import { mockFetch } from "../setup";
import {
  documentArtifact,
  installGroupsFetch,
  tempGroup,
} from "./GroupManagementPanel.testSupport";

describe("GroupManagementPanel project package outputs", () => {
  beforeEach(() => {
    mockFetch.mockReset();
    window.localStorage.clear();
  });

  it("renders retained package outputs with open and storage controls", async () => {
    const openWindow = vi.spyOn(window, "open").mockImplementation(() => null);
    installGroupsFetch({
      groups: [tempGroup()],
      outputs: {
        "group-temp": [
          documentArtifact({
            id: "artifact-game",
            artifact_type: "project_package",
            title: "Coin Runner Game",
            content_type: "application/vnd.mycelis.project+json",
            metadata: {
              entrypoint: "dist/index.html",
              folder: "workspace/generated/coin-runner",
              files: ["index.html", "game.js", "styles.css"],
              validation: "Opened in browser and score increased after click.",
            },
          }),
        ],
      },
    });

    render(<GroupManagementPanel initialSelectedGroupId="group-temp" />);

    fireEvent.click(screen.getByRole("tab", { name: /Outputs/i }));
    await waitFor(() =>
      expect(screen.getByText("Coin Runner Game")).toBeDefined(),
    );
    expect(screen.getByText("Project package")).toBeDefined();
    expect(
      screen.getByText("dist/index.html"),
    ).toBeDefined();
    expect(screen.getByText("workspace/generated/coin-runner")).toBeDefined();
    expect(screen.getByText("game.js")).toBeDefined();
    expect(
      screen.getByText("Opened in browser and score increased after click."),
    ).toBeDefined();

    fireEvent.click(
      screen.getByRole("button", {
        name: /Open file Coin Runner Game in a new browser window/i,
      }),
    );
    expect(openWindow).toHaveBeenCalledWith(
      "/api/v1/workspace/files/view?path=workspace%2Fgenerated%2Fcoin-runner%2Fdist%2Findex.html",
      "_blank",
      "noopener,noreferrer",
    );
    expect(
      screen.getByRole("link", {
        name: /Open Coin Runner Game in Resources/i,
      }).getAttribute("href"),
    ).toBe("/resources?tab=workspace&path=workspace%2Fgenerated%2Fcoin-runner");

    fireEvent.click(
      screen.getByRole("button", {
        name: /Open local folder for Coin Runner Game/i,
      }),
    );
    await waitFor(() =>
      expect(mockFetch).toHaveBeenCalledWith(
        "/api/v1/workspace/files/reveal?path=workspace%2Fgenerated%2Fcoin-runner",
        { method: "POST" },
      ),
    );
    openWindow.mockRestore();
  });

  it("opens sparse package outputs from the saved file path", async () => {
    const openWindow = vi.spyOn(window, "open").mockImplementation(() => null);
    installGroupsFetch({
      groups: [tempGroup()],
      outputs: {
        "group-temp": [
          documentArtifact({
            id: "artifact-sparse-game",
            artifact_type: "project_package",
            title: "Sparse Game Package",
            content_type: "application/vnd.mycelis.project+json",
            file_path: "workspace/generated/sparse-game/index.html",
            metadata: {},
          }),
        ],
      },
    });

    render(<GroupManagementPanel initialSelectedGroupId="group-temp" />);

    fireEvent.click(screen.getByRole("tab", { name: /Outputs/i }));
    await waitFor(() =>
      expect(screen.getByText("Sparse Game Package")).toBeDefined(),
    );

    fireEvent.click(
      screen.getByRole("button", {
        name: /Open file Sparse Game Package in a new browser window/i,
      }),
    );
    expect(openWindow).toHaveBeenCalledWith(
      "/api/v1/workspace/files/view?path=workspace%2Fgenerated%2Fsparse-game%2Findex.html",
      "_blank",
      "noopener,noreferrer",
    );
    expect(
      screen.getByRole("link", {
        name: /Open Sparse Game Package in Resources/i,
      }).getAttribute("href"),
    ).toBe("/resources?tab=workspace&path=workspace%2Fgenerated%2Fsparse-game");
    openWindow.mockRestore();
  });

  it("opens saved HTML outputs instead of presenting raw HTML as the result", async () => {
    const openWindow = vi.spyOn(window, "open").mockImplementation(() => null);
    installGroupsFetch({
      groups: [tempGroup()],
      outputs: {
        "group-temp": [
          documentArtifact({
            id: "artifact-html",
            artifact_type: "code",
            title: "workspace/logs/qa_orbit_dash.html",
            content_type: "text/html",
            content:
              "<!doctype html><title>Dot Dodge</title><style>body{background:#111}</style>",
          }),
        ],
      },
    });

    render(<GroupManagementPanel initialSelectedGroupId="group-temp" />);

    fireEvent.click(screen.getByRole("tab", { name: /Outputs/i }));
    await waitFor(() =>
      expect(screen.getByText("workspace/logs/qa_orbit_dash.html")).toBeDefined(),
    );
    expect(
      screen.getByText(/HTML output is saved as a browser-viewable file/i),
    ).toBeDefined();
    expect(screen.queryByText(/<!doctype html><title>Dot Dodge/i)).toBeNull();

    fireEvent.click(
      screen.getByRole("button", {
        name: /Open file workspace\/logs\/qa_orbit_dash.html in a new browser window/i,
      }),
    );
    expect(openWindow).toHaveBeenCalledWith(
      "/api/v1/workspace/files/view?path=workspace%2Flogs%2Fqa_orbit_dash.html",
      "_blank",
      "noopener,noreferrer",
    );

    openWindow.mockRestore();
  });
});
