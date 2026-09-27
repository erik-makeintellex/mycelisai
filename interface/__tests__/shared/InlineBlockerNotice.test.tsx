import { describe, expect, it, vi } from "vitest";
import { render, screen } from "@testing-library/react";
import InlineBlockerNotice from "@/components/shared/InlineBlockerNotice";

describe("InlineBlockerNotice", () => {
    it("renders the non-admin copy when viewerIsAdmin is false", () => {
        render(<InlineBlockerNotice code="governance_policy_unavailable" viewerIsAdmin={false} />);
        expect(screen.getByText("Safety lock is on")).toBeDefined();
        expect(screen.queryByText("Safety rules didn't load")).toBeNull();
    });

    it("renders the admin variant when viewerIsAdmin is true", () => {
        render(<InlineBlockerNotice code="governance_policy_unavailable" viewerIsAdmin />);
        expect(screen.getByText("Safety rules didn't load")).toBeDefined();
        expect(screen.queryByText("Safety lock is on")).toBeNull();
    });

    it("falls back to the base copy for a code with no admin variant, regardless of viewerIsAdmin", () => {
        render(<InlineBlockerNotice code="token_already_used" viewerIsAdmin />);
        expect(screen.getByText("Already approved")).toBeDefined();
    });

    it("renders role=alert and an onRetry button labeled from the resolved copy", () => {
        const onRetry = vi.fn();
        render(<InlineBlockerNotice code="request_failed" viewerIsAdmin={false} onRetry={onRetry} />);
        expect(screen.getByRole("alert")).toBeDefined();
        expect(screen.getByText("Try again")).toBeDefined();
    });
});
