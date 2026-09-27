import { describe, expect, it } from "vitest";
import { render, screen } from "@testing-library/react";
import { OutputProofBadges } from "@/components/soma/OutputWorkbenchProofDetails";
import type { OutputProofEnvelope } from "@/store/useCortexStore";

// PH-B: readback_status must never render green for a non-verified state,
// and the label text itself (not just color) must say so.
describe("OutputProofBadges", () => {
    it("shows a green success badge for a verified readback", () => {
        const proof = { readback_status: "verified" } as OutputProofEnvelope;
        const { container } = render(<OutputProofBadges proof={proof} />);
        expect(screen.getByText("readback verified")).toBeDefined();
        expect(container.querySelector(".text-cortex-success")).not.toBeNull();
    });

    it("shows a neutral 'Not verified yet' badge for an unverified readback, not green", () => {
        const proof = { readback_status: "unverified" } as OutputProofEnvelope;
        const { container } = render(<OutputProofBadges proof={proof} />);
        expect(screen.getByText("Not verified yet")).toBeDefined();
        expect(screen.queryByText(/readback unverified/i)).toBeNull();
        expect(container.querySelector(".text-cortex-success")).toBeNull();
    });

    it("shows a warning 'Couldn't confirm the file' badge for a failed readback, not green", () => {
        const proof = { readback_status: "failed" } as OutputProofEnvelope;
        const { container } = render(<OutputProofBadges proof={proof} />);
        expect(screen.getByText("Couldn't confirm the file")).toBeDefined();
        expect(container.querySelector(".text-cortex-success")).toBeNull();
        expect(container.querySelector(".text-cortex-warning")).not.toBeNull();
    });

    it("shows the same warning badge for any output_* readback status", () => {
        const proof = { readback_status: "output_readback_missing" } as OutputProofEnvelope;
        const { container } = render(<OutputProofBadges proof={proof} />);
        expect(screen.getByText("Couldn't confirm the file")).toBeDefined();
        expect(container.querySelector(".text-cortex-success")).toBeNull();
    });
});
