import { describe, it, expect, vi } from "vitest";
import { render, screen, waitFor } from "@testing-library/react";
import { mockFetch } from "../setup";
import type { Artifact } from "@/store/cortexStoreTypesPlanning";

const mockFetchArtifacts = vi.fn();

vi.mock("@/store/useCortexStore", () => ({
  useCortexStore: (selector: (state: {
    artifacts: Artifact[];
    isFetchingArtifacts: boolean;
    fetchArtifacts: typeof mockFetchArtifacts;
  }) => unknown) =>
    selector({
      artifacts: [
        {
          id: "artifact-1",
          agent_id: "Soma",
          artifact_type: "document",
          title: "Retained launch note",
          content_type: "text/markdown",
          content: "Launch note",
          metadata: {},
          status: "approved",
          created_at: "2026-01-01T00:00:00.000Z",
        },
      ],
      isFetchingArtifacts: false,
      fetchArtifacts: mockFetchArtifacts,
    }),
}));

import WarmMemoryPanel from "@/components/memory/WarmMemoryPanel";

describe("WarmMemoryPanel", () => {
  it("shows Warm, SitReps, and Artifacts as clickable tabs", async () => {
    mockFetch.mockResolvedValue({
      ok: true,
      json: async () => ({
        sitreps: [
          {
            id: "sitrep-1",
            mission_id: "mission-1",
            summary: "A useful retained work summary.",
            raw_count: 2,
            created_at: "2026-01-01T00:00:00.000Z",
          },
        ],
      }),
    });

    const { fireEvent } = await import("@testing-library/react");
    render(<WarmMemoryPanel />);

    expect(
      screen.getByRole("tab", { name: /Warm/i }).getAttribute("aria-selected"),
    ).toBe("true");
    await waitFor(() =>
      expect(screen.getByText("Latest SitReps")).toBeDefined(),
    );
    expect(screen.getByText("Retained launch note")).toBeDefined();

    fireEvent.click(screen.getByRole("tab", { name: /SitReps/i }));
    await waitFor(() =>
      expect(
        screen.getByRole("tab", { name: /SitReps/i }).getAttribute("aria-selected"),
      ).toBe("true"),
    );

    fireEvent.click(screen.getByRole("tab", { name: /Artifacts/i }));
    await waitFor(() =>
      expect(
        screen.getByRole("tab", { name: /Artifacts/i }).getAttribute("aria-selected"),
      ).toBe("true"),
    );

    fireEvent.click(screen.getByRole("tab", { name: /Warm/i }));
    await waitFor(() =>
      expect(
        screen.getByRole("tab", { name: /Warm/i }).getAttribute("aria-selected"),
      ).toBe("true"),
    );
  });
});
