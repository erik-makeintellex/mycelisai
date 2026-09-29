import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";
import { render, screen, waitFor } from "@testing-library/react";

vi.mock("reactflow", async () => await import("../mocks/reactflow"));
vi.mock("@/components/dashboard/ManifestationPanel", () => ({
  __esModule: true,
  default: () => <div data-testid="manifestation-panel">ManifestationPanel</div>,
}));

import ApprovalsTab from "@/components/automations/ApprovalsTab";

// C2-RETIRE: the in-memory approval queue is gone. The Approvals tab opens on
// Proposals (the durable approval path) and never polls the retired queue.
describe("ApprovalsTab without the retired approval queue", () => {
  const fetchMock = vi.fn(async () => new Response("[]", { status: 200, headers: { "Content-Type": "application/json" } }));

  beforeEach(() => {
    vi.useFakeTimers({ shouldAdvanceTime: true });
    fetchMock.mockClear();
    vi.stubGlobal("fetch", fetchMock);
  });

  afterEach(() => {
    vi.useRealTimers();
    vi.unstubAllGlobals();
  });

  it("does not fetch /api/v1/governance/pending, even after the old 5 s poll", async () => {
    render(<ApprovalsTab />);
    await vi.advanceTimersByTimeAsync(11_000);
    const urls = fetchMock.mock.calls.map((call) => String((call as unknown[])[0]));
    expect(urls.filter((url) => url.includes("/governance/pending"))).toEqual([]);
    expect(urls.filter((url) => url.includes("/governance/resolve"))).toEqual([]);
  });

  it("has no Queue sub-tab and opens on Proposals", async () => {
    render(<ApprovalsTab />);
    expect(screen.queryByRole("button", { name: "Queue" })).toBeNull();
    await waitFor(() => expect(screen.getByTestId("manifestation-panel")).toBeDefined());
  });
});
