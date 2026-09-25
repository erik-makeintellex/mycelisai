import { describe, it, expect, beforeEach, vi } from "vitest";
import { render, screen, fireEvent, waitFor } from "@testing-library/react";
import { ConfigViewer } from "@/components/deployments/ConfigViewer";
import { DeploymentDetailPane } from "@/components/deployments/DeploymentDetailPane";
import { mockFetch } from "../setup";
import type { DeploymentConfigRef, DeploymentRecord } from "@/components/deployments/deploymentTypes";

function jsonResponse(status: number, body: unknown) {
    return { ok: status >= 200 && status < 300, status, json: async () => body };
}

const configRef: DeploymentConfigRef = {
    kind: "team",
    id: "team-1",
    path: "/api/v1/config/teams/team-1",
};

const configDoc = {
    kind: "team",
    id: "team-1",
    format: "yaml",
    document: "id: team-1\nname: Delivery Team\n",
    source: "runtime" as const,
    redactions: ["members[0].provider.api_key"],
    generated_at: "2026-09-25T17:58:09Z",
};

function record(overrides: Partial<DeploymentRecord> = {}): DeploymentRecord {
    return {
        id: "rec-1",
        kind: "standing_team",
        name: "Standing Team One",
        what_runs: "3 agents answering product requests",
        scope: { kind: "global" },
        origin: "bundle",
        lifecycle_state: "running",
        members: [],
        tools: [],
        config_ref: configRef,
        ...overrides,
    };
}

describe("ConfigViewer", () => {
    beforeEach(() => {
        mockFetch.mockReset();
    });

    it("renders the YAML document with source and redactions", async () => {
        mockFetch.mockResolvedValue(jsonResponse(200, { ok: true, data: configDoc }));

        render(<ConfigViewer configRef={configRef} onClose={() => {}} />);

        const doc = await screen.findByTestId("config-viewer-document");
        expect(doc.textContent).toMatch(/name: Delivery Team/);
        expect(mockFetch).toHaveBeenCalledWith("/api/v1/config/teams/team-1?format=yaml", expect.any(Object));
        expect(screen.getByTestId("config-viewer-source").textContent).toMatch(/Runtime/);
        expect(screen.getByTestId("config-viewer-redaction-item").textContent).toMatch(
            "members[0].provider.api_key",
        );
    });

    it("refetches with ?format=json when the JSON toggle is clicked", async () => {
        mockFetch.mockResolvedValue(jsonResponse(200, { ok: true, data: configDoc }));
        render(<ConfigViewer configRef={configRef} onClose={() => {}} />);
        await screen.findByTestId("config-viewer-document");

        mockFetch.mockResolvedValue(
            jsonResponse(200, { ok: true, data: { ...configDoc, format: "json", document: '{"id":"team-1"}' } }),
        );
        fireEvent.click(screen.getByTestId("config-viewer-format-json"));

        await waitFor(() =>
            expect(mockFetch).toHaveBeenLastCalledWith("/api/v1/config/teams/team-1?format=json", expect.any(Object)),
        );
        const doc = await screen.findByTestId("config-viewer-document");
        await waitFor(() => expect(doc.textContent).toMatch(/"id":"team-1"/));
    });

    it("copies the document to the clipboard", async () => {
        const writeText = vi.fn();
        Object.defineProperty(navigator, "clipboard", { value: { writeText }, configurable: true });
        mockFetch.mockResolvedValue(jsonResponse(200, { ok: true, data: configDoc }));

        render(<ConfigViewer configRef={configRef} onClose={() => {}} />);
        await screen.findByTestId("config-viewer-document");
        fireEvent.click(screen.getByTestId("config-viewer-copy"));

        expect(writeText).toHaveBeenCalledWith(configDoc.document);
    });

    it("shows a restricted notice on 403, not a generic error", async () => {
        mockFetch.mockResolvedValue(jsonResponse(403, {}));
        render(<ConfigViewer configRef={configRef} onClose={() => {}} />);

        const restricted = await screen.findByTestId("config-viewer-restricted");
        expect(restricted.textContent).toMatch(/requires admin access/i);
        expect(screen.queryByTestId("config-viewer-load-failure")).toBeNull();
    });

    it("shows a not-found notice on 404", async () => {
        mockFetch.mockResolvedValue(jsonResponse(404, {}));
        render(<ConfigViewer configRef={configRef} onClose={() => {}} />);

        const notFound = await screen.findByTestId("config-viewer-not-found");
        expect(notFound.textContent).toMatch(/config not found/i);
    });

    it("shows LoadFailureState with Retry on 503, and Retry refetches", async () => {
        mockFetch.mockResolvedValueOnce(jsonResponse(503, {}));
        render(<ConfigViewer configRef={configRef} onClose={() => {}} />);

        expect(await screen.findByTestId("config-viewer-load-failure")).toBeTruthy();

        mockFetch.mockResolvedValueOnce(jsonResponse(200, { ok: true, data: configDoc }));
        fireEvent.click(screen.getByRole("button", { name: /Retry/i }));

        const doc = await screen.findByTestId("config-viewer-document");
        expect(doc.textContent).toMatch(/name: Delivery Team/);
    });
});

describe("DeploymentDetailPane View config affordance", () => {
    beforeEach(() => {
        mockFetch.mockReset();
    });

    it("hides View config when Advanced mode is off", () => {
        render(<DeploymentDetailPane record={record()} advancedMode={false} />);
        expect(screen.queryByTestId("deployments-view-config")).toBeNull();
    });

    it("opens the ConfigViewer when View config is clicked in Advanced mode", async () => {
        mockFetch.mockResolvedValue(jsonResponse(200, { ok: true, data: configDoc }));
        render(<DeploymentDetailPane record={record()} advancedMode={true} />);

        expect(screen.queryByTestId("config-viewer")).toBeNull();
        fireEvent.click(screen.getByTestId("deployments-view-config"));

        expect(await screen.findByTestId("config-viewer")).toBeTruthy();
        const doc = await screen.findByTestId("config-viewer-document");
        expect(doc.textContent).toMatch(/name: Delivery Team/);
    });
});
