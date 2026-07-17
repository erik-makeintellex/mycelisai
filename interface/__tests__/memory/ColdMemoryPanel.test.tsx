import { describe, expect, it, vi } from "vitest";
import { fireEvent, render, screen, waitFor } from "@testing-library/react";

import ColdMemoryPanel from "@/components/memory/ColdMemoryPanel";

describe("ColdMemoryPanel", () => {
  it("shows a clean degraded state when embeddings are unavailable", async () => {
    vi.spyOn(globalThis, "fetch").mockResolvedValueOnce({
      ok: true,
      json: async () => ({
        results: [],
        count: 0,
        degraded: {
          code: "embedding_unavailable",
          summary:
            "Semantic memory search is unavailable because no embedding provider is available.",
          recommended_action:
            "Configure an embedding-capable AI engine before relying on vector memory recall.",
        },
      }),
    } as Response);

    render(<ColdMemoryPanel />);

    fireEvent.change(screen.getByRole("textbox", { name: /search semantic memory/i }), {
      target: { value: "Soma research" },
    });

    await waitFor(() => {
      expect(screen.getByText(/Memory search needs attention/i)).toBeDefined();
    });
    expect(screen.getByText(/no embedding provider is available/i)).toBeDefined();
    expect(screen.queryByText(/No results found/i)).toBeNull();
  });
});
