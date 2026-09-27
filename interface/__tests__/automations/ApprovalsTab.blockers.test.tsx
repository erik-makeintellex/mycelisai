import { beforeEach, describe, expect, it, vi } from "vitest";
import { fireEvent, render, screen } from "@testing-library/react";

vi.mock("reactflow", async () => {
  const mock = await import("../mocks/reactflow");
  return mock;
});

vi.mock("@/components/approvals/DecisionCard", () => ({
  DecisionCard: () => <div data-testid="decision-card">DecisionCard</div>,
}));

vi.mock("@/components/workspace/TrustSlider", () => ({
  __esModule: true,
  default: () => <div data-testid="trust-slider">TrustSlider</div>,
}));

vi.mock("@/components/dashboard/ManifestationPanel", () => ({
  __esModule: true,
  default: () => <div data-testid="manifestation-panel">ManifestationPanel</div>,
}));

import ApprovalsTab from "@/components/automations/ApprovalsTab";
import { useCortexStore } from "@/store/useCortexStore";

describe("ApprovalsTab blockers (no fake success)", () => {
  beforeEach(() => {
    useCortexStore.setState({
      pendingApprovals: [],
      isFetchingApprovals: false,
      approvalsError: null,
      resolveApprovalError: null,
      fetchPendingApprovals: vi.fn().mockResolvedValue(undefined),
      resolveApproval: vi.fn().mockResolvedValue(undefined),
      policyConfig: null,
      policyError: null,
      isFetchingPolicy: false,
      fetchPolicy: vi.fn().mockResolvedValue(undefined),
      updatePolicy: vi.fn().mockResolvedValue(undefined),
      auditLog: [],
      isFetchingAuditLog: false,
      fetchAuditLog: vi.fn().mockResolvedValue(undefined),
    });
  });

  it("does not render All Clear when the queue 403s; shows a blocker instead", () => {
    useCortexStore.setState({ approvalsError: { code: "admin_required", httpStatus: 403 } });
    render(<ApprovalsTab />);

    expect(screen.queryByText("All Clear")).toBeNull();
    expect(screen.getByRole("alert")).toBeDefined();
  });

  it("hides the pending-count line on a 403/503 instead of showing '0 pending requests'", () => {
    useCortexStore.setState({ approvalsError: { code: "admin_required", httpStatus: 403 } });
    render(<ApprovalsTab />);

    expect(screen.queryByText(/pending request/)).toBeNull();
  });

  it("shows the empty state only when there is truly no error and no pending items", () => {
    render(<ApprovalsTab />);
    expect(screen.getByText("Nothing waiting for approval")).toBeDefined();
    expect(screen.queryByRole("alert")).toBeNull();
  });

  it("surfaces a resolve failure instead of the card silently staying put", () => {
    useCortexStore.setState({ resolveApprovalError: { code: "request_failed", httpStatus: 503 } });
    render(<ApprovalsTab />);
    expect(screen.getByRole("alert")).toBeDefined();
  });

  it("does not render a fake empty DENY policy editor when the Policy tab 403s/503s", () => {
    useCortexStore.setState({ policyError: { code: "policy_unavailable", httpStatus: 503 } });
    render(<ApprovalsTab />);
    fireEvent.click(screen.getByRole("button", { name: "Policy" }));

    expect(screen.queryByText("Add Group")).toBeNull();
    expect(screen.queryByText("Save Policy")).toBeNull();
    expect(screen.getByRole("alert")).toBeDefined();
  });

  it("renders the real policy editor once policyConfig loads successfully", () => {
    useCortexStore.setState({
      policyConfig: { groups: [], defaults: { default_action: "ALLOW" } },
      policyError: null,
    });
    render(<ApprovalsTab />);
    fireEvent.click(screen.getByRole("button", { name: "Policy" }));

    expect(screen.getByText("Add Group")).toBeDefined();
    expect(screen.getByText("Save Policy")).toBeDefined();
  });
});
