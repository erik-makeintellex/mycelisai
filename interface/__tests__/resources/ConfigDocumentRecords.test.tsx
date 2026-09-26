import { describe, it, expect, vi, beforeEach, afterEach } from "vitest";
import { fireEvent, render, screen, waitFor } from "@testing-library/react";
import ConfigDocumentRecords from "@/components/resources/ConfigDocumentRecords";

function sessionResponse(role: "admin" | "standard") {
  return {
    ok: true,
    status: 200,
    json: async () => ({ ok: true, data: { user: { role } } }),
  };
}

function listResponse(status: number, data: unknown = []) {
  return {
    ok: status >= 200 && status < 300,
    status,
    json: async () => ({ ok: status < 300, data, error: status >= 300 ? "backend detail should never render" : undefined }),
  };
}

const sampleRecord = {
  record_id: "7d9f4c1e-2a3b-4c5d-8e9f-0a1b2c3d4e5f",
  document: {
    kind: "WorkerProfile",
    metadata: { id: "seeded-profile", version: "1", scope: { kind: "workspace", ref: "primary" }, enabled: true },
  },
};

describe("ConfigDocumentRecords", () => {
  const fetchMock = vi.fn();

  beforeEach(() => {
    vi.stubGlobal("fetch", fetchMock);
  });

  afterEach(() => {
    vi.unstubAllGlobals();
    fetchMock.mockReset();
  });

  it("renders nothing for a non-admin session and never fetches the list", async () => {
    fetchMock.mockResolvedValueOnce(sessionResponse("standard"));

    const { container } = render(<ConfigDocumentRecords />);

    await waitFor(() => expect(fetchMock).toHaveBeenCalledTimes(1));
    expect(container.firstChild).toBeNull();
    expect(fetchMock).not.toHaveBeenCalledWith(expect.stringContaining("/api/v1/config-documents"), expect.anything());
  });

  it("lists exportable records with kind, id, version, scope, and active state for an admin", async () => {
    fetchMock
      .mockResolvedValueOnce(sessionResponse("admin"))
      .mockResolvedValueOnce(listResponse(200, [sampleRecord]));

    render(<ConfigDocumentRecords />);

    await waitFor(() => {
      expect(screen.getByText(/WorkerProfile.*seeded-profile/)).toBeDefined();
    });
    expect(screen.getByText(/v1/)).toBeDefined();
    expect(screen.getByText(/workspace\/primary/)).toBeDefined();
    expect(screen.getByText(/Active/)).toBeDefined();
    expect(screen.getByRole("button", { name: /View config/i })).toBeDefined();
  });

  it("opens the viewer dialog when View config is clicked", async () => {
    fetchMock
      .mockResolvedValueOnce(sessionResponse("admin"))
      .mockResolvedValueOnce(listResponse(200, [sampleRecord]))
      .mockResolvedValueOnce({
        ok: true,
        status: 200,
        json: async () => ({ ok: true, data: { content: "apiVersion: mycelis.ai/v1", redaction_applied: false } }),
      });

    render(<ConfigDocumentRecords />);

    await waitFor(() => screen.getByRole("button", { name: /View config/i }));
    fireEvent.click(screen.getByRole("button", { name: /View config/i }));

    expect(screen.getByRole("dialog", { name: "View config" })).toBeDefined();
    await waitFor(() => {
      expect(screen.getByText("apiVersion: mycelis.ai/v1")).toBeDefined();
    });
  });

  it("shows a normalized alert on 403 and never renders the raw backend error body", async () => {
    fetchMock.mockResolvedValueOnce(sessionResponse("admin")).mockResolvedValueOnce(listResponse(403));

    render(<ConfigDocumentRecords />);

    await waitFor(() => {
      expect(screen.getByText(/do not have access/i)).toBeDefined();
    });
    expect(screen.queryByText(/backend detail should never render/i)).toBeNull();
  });

  it("shows a normalized alert on 503 and never renders the raw backend error body", async () => {
    fetchMock.mockResolvedValueOnce(sessionResponse("admin")).mockResolvedValueOnce(listResponse(503));

    render(<ConfigDocumentRecords />);

    await waitFor(() => {
      expect(screen.getByText(/unavailable right now/i)).toBeDefined();
    });
    expect(screen.queryByText(/backend detail should never render/i)).toBeNull();
  });
});
