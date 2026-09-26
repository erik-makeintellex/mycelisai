import { describe, it, expect, vi, beforeEach, afterEach } from "vitest";
import { fireEvent, render, screen, waitFor } from "@testing-library/react";
import ConfigDocumentViewDialog from "@/components/resources/ConfigDocumentViewDialog";

function exportResponse(status: number, data: unknown = null) {
  return {
    ok: status >= 200 && status < 300,
    status,
    json: async () => ({
      ok: status < 300,
      data,
      error: status >= 300 ? "raw backend detail should never render" : undefined,
    }),
  };
}

describe("ConfigDocumentViewDialog", () => {
  const fetchMock = vi.fn();
  const writeText = vi.fn().mockResolvedValue(undefined);

  beforeEach(() => {
    vi.stubGlobal("fetch", fetchMock);
    Object.assign(navigator, { clipboard: { writeText } });
  });

  afterEach(() => {
    vi.unstubAllGlobals();
    fetchMock.mockReset();
    writeText.mockClear();
  });

  it("renders exported content as plain text, never as HTML", async () => {
    fetchMock.mockResolvedValueOnce(
      exportResponse(200, { content: "<script>window.x=1</script>", redaction_applied: false }),
    );

    const { container } = render(<ConfigDocumentViewDialog recordId="rec-1" onClose={() => {}} />);

    await waitFor(() => {
      expect(screen.getByText("<script>window.x=1</script>")).toBeDefined();
    });
    expect(container.querySelector("script")).toBeNull();
  });

  it("shows the redaction notice when redaction_applied is true and never claims it is re-importable", async () => {
    fetchMock.mockResolvedValueOnce(
      exportResponse(200, { content: "spec:\n  api_key: '[redacted]'", redaction_applied: true }),
    );

    render(<ConfigDocumentViewDialog recordId="rec-1" onClose={() => {}} />);

    await waitFor(() => {
      expect(screen.getByText("Some values were hidden.")).toBeDefined();
    });
    expect(screen.queryByText(/re-importable/i)).toBeNull();
  });

  it("copies content to the clipboard", async () => {
    fetchMock.mockResolvedValueOnce(exportResponse(200, { content: "apiVersion: mycelis.ai/v1", redaction_applied: false }));

    render(<ConfigDocumentViewDialog recordId="rec-1" onClose={() => {}} />);

    const copyButton = await screen.findByRole("button", { name: /copy/i });
    fireEvent.click(copyButton);

    await waitFor(() => expect(writeText).toHaveBeenCalledWith("apiVersion: mycelis.ai/v1"));
    expect(await screen.findByText("Copied")).toBeDefined();
  });

  it("shows a normalized alert on 403 and never renders the raw backend error body", async () => {
    fetchMock.mockResolvedValueOnce(exportResponse(403));

    render(<ConfigDocumentViewDialog recordId="rec-1" onClose={() => {}} />);

    await waitFor(() => {
      expect(screen.getByText(/do not have access/i)).toBeDefined();
    });
    expect(screen.queryByText(/raw backend detail/i)).toBeNull();
  });

  it("shows a normalized alert on 503 and never renders the raw backend error body", async () => {
    fetchMock.mockResolvedValueOnce(exportResponse(503));

    render(<ConfigDocumentViewDialog recordId="rec-1" onClose={() => {}} />);

    await waitFor(() => {
      expect(screen.getByText(/unavailable right now/i)).toBeDefined();
    });
    expect(screen.queryByText(/raw backend detail/i)).toBeNull();
  });

  it("shows a normalized alert on 404 without leaking backend detail", async () => {
    fetchMock.mockResolvedValueOnce(exportResponse(404));

    render(<ConfigDocumentViewDialog recordId="rec-1" onClose={() => {}} />);

    await waitFor(() => {
      expect(screen.getByText(/could not be found/i)).toBeDefined();
    });
    expect(screen.queryByText(/raw backend detail/i)).toBeNull();
  });
});
