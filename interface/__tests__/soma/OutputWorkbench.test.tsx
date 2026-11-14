import { describe, it, expect } from "vitest";
import { render, screen, within } from "@testing-library/react";
import OutputWorkbench from "@/components/soma/OutputWorkbench";
import type { DeliveryPackageSummary, OutputRef } from "@/store/useCortexStore";

describe("OutputWorkbench", () => {
    const sampleRefs: OutputRef[] = [
        {
            id: "out-1",
            kind: "code",
            label: "Handler Patch",
            storage_ref: "core/internal/server/patch.go",
            role: "primary",
            status: "verified",
            content_type: "text/x-go",
            bytes: 512,
            sha256: "1234567890abcdef1234567890abcdef",
            open_url: "/api/v1/artifacts/out-1/download",
            proof: {
                proof_id: "proof-1",
                storage_ref: "core/internal/server/patch.go",
                checksum: "1234567890abcdef1234567890abcdef",
                checksum_algorithm: "sha256",
                bytes: 512,
                readback_status: "verified",
            },
        },
        {
            id: "out-2",
            kind: "document",
            label: "Package Proof",
            storage_ref: "workspace/deliverables/proof.md",
            role: "proof",
            status: "verified",
            content_type: "text/markdown",
            bytes: 256,
            sha256: "abcdef0123456789abcdef0123456789",
            open_url: "/api/v1/artifacts/out-2/download",
        },
    ];

    const samplePackage: DeliveryPackageSummary = {
        package_id: "implementation_bundle:team-alpha",
        package_kind: "implementation_bundle",
        status: "verified",
        run_id: "run-alpha-1",
        items: sampleRefs,
    };

    it("renders output-workbench, stage strip, 5 stage pills, proof sha256, and file list items", () => {
        render(<OutputWorkbench deliveryPackage={samplePackage} outputRefs={sampleRefs} />);

        const workbench = screen.getByTestId("output-workbench");
        expect(workbench).toBeDefined();

        const stageStrip = screen.getByTestId("output-workbench-stage-strip");
        expect(stageStrip).toBeDefined();

        for (const label of ["Intent", "Approval", "Execution", "Package", "Proof"]) {
            const pill = within(stageStrip).getByText(label);
            expect(pill).toBeDefined();
            expect(pill.getAttribute("data-stage-state")).toBe("completed");
        }

        expect(screen.getByText("Handler Patch")).toBeDefined();
        expect(screen.getByText("core/internal/server/patch.go")).toBeDefined();
        expect(screen.getByText("Package Proof")).toBeDefined();
        expect(screen.getByText("workspace/deliverables/proof.md")).toBeDefined();
        expect(screen.getByText("abcdef0123456789")).toBeDefined();
    });

    it("reflects pending_review and failed proof stage states in the stage strip", () => {
        const { rerender } = render(
            <OutputWorkbench
                deliveryPackage={{ ...samplePackage, status: "pending_review" }}
                outputRefs={sampleRefs}
            />,
        );

        let stageStrip = screen.getByTestId("output-workbench-stage-strip");
        expect(within(stageStrip).getByText("Proof").getAttribute("data-stage-state")).toBe("active");

        rerender(
            <OutputWorkbench
                deliveryPackage={{ ...samplePackage, status: "failed" }}
                outputRefs={sampleRefs}
            />,
        );

        stageStrip = screen.getByTestId("output-workbench-stage-strip");
        expect(within(stageStrip).getByText("Proof").getAttribute("data-stage-state")).toBe("failed");
    });
});
