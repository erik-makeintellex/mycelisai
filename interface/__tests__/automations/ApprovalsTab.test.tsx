import { beforeEach, describe, expect, it, vi } from "vitest";
import { fireEvent, render, screen, waitFor } from "@testing-library/react";

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
  default: () => (
    <div data-testid="manifestation-panel">ManifestationPanel</div>
  ),
}));

import ApprovalsTab from "@/components/automations/ApprovalsTab";
import { useCortexStore } from "@/store/useCortexStore";

describe("ApprovalsTab", () => {
  beforeEach(() => {
    useCortexStore.setState({
      pendingApprovals: [],
      approvalsLoadError: null,
      isFetchingApprovals: false,
      fetchPendingApprovals: vi.fn().mockResolvedValue(undefined),
      resolveApproval: vi.fn().mockResolvedValue(undefined),
      policyConfig: null,
      policyLoadError: null,
      isFetchingPolicy: false,
      fetchPolicy: vi.fn().mockResolvedValue(undefined),
      updatePolicy: vi.fn().mockResolvedValue(undefined),
      auditLogLoadError: null,
      auditLog: [
        {
          id: "audit-1",
          actor: "Soma",
          user: "local-user",
          action: "proposal_generated",
          timestamp: "2026-03-26T12:00:00Z",
          capability_used: "planning",
          result_status: "pending",
          approval_status: "approval_required",
          approval_reason: "capability_risk",
          intent_proof_id: "proof-123",
          details: {
            operator_summary: "Researcher ask: Map the governed release proof.",
            team_id: "research-team",
            ask_kind: "research",
            lane_role: "researcher",
          },
        },
      ],
      isFetchingAuditLog: false,
      fetchAuditLog: vi.fn().mockResolvedValue(undefined),
    });
  });

  it("renders the inspect-only audit activity view", async () => {
    render(<ApprovalsTab />);

    fireEvent.click(screen.getByRole("button", { name: "Audit" }));

    await waitFor(() => {
      expect(useCortexStore.getState().fetchAuditLog).toHaveBeenCalledTimes(1);
    });
    expect(screen.getByText("Activity Log")).toBeDefined();
    expect(screen.getByText(/proposal generated/i)).toBeDefined();
    expect(
      screen.getByText(/researcher ask: map the governed release proof\./i),
    ).toBeDefined();
    expect(screen.getByText(/capability: planning/i)).toBeDefined();
    expect(screen.getByText(/reason: capability risk/i)).toBeDefined();
    expect(screen.getByText(/team: research-team/i)).toBeDefined();
    expect(screen.getByText(/ask: research/i)).toBeDefined();
    expect(screen.getByText(/role: researcher/i)).toBeDefined();
  }, 15000);

  it("shows a load failure with Retry for the approvals queue instead of 'All Clear'", async () => {
    const fetchPendingApprovals = vi.fn().mockResolvedValue(undefined);
    useCortexStore.setState({
      pendingApprovals: [],
      approvalsLoadError: 'Could not load approvals (HTTP 500).',
      fetchPendingApprovals,
    });

    render(<ApprovalsTab />);

    expect(screen.queryByText('All Clear')).toBeNull();
    const failure = screen.getByTestId('approvals-load-failure');
    expect(failure.textContent).toMatch(/Could not load approvals/);

    fireEvent.click(screen.getByRole('button', { name: /Retry/i }));
    await waitFor(() => {
      expect(fetchPendingApprovals).toHaveBeenCalled();
    });
  });

  it("shows 'All Clear' (not a failure) after a successful empty approvals fetch", () => {
    useCortexStore.setState({ pendingApprovals: [], approvalsLoadError: null });

    render(<ApprovalsTab />);

    expect(screen.getByText('All Clear')).toBeDefined();
    expect(screen.queryByTestId('approvals-load-failure')).toBeNull();
  });

  it('shows a load failure with Retry for the audit log instead of the empty state', async () => {
    const fetchAuditLog = vi.fn().mockResolvedValue(undefined);
    useCortexStore.setState({
      auditLog: [],
      auditLogLoadError: 'Could not load the audit log (HTTP 500).',
      fetchAuditLog,
    });

    render(<ApprovalsTab />);
    fireEvent.click(screen.getByRole('button', { name: 'Audit' }));

    const failure = await screen.findByTestId('audit-log-load-failure');
    expect(failure.textContent).toMatch(/Could not load the audit log/);
    expect(screen.queryByText('No recent audit activity')).toBeNull();

    fireEvent.click(screen.getByRole('button', { name: /Retry/i }));
    await waitFor(() => {
      expect(fetchAuditLog).toHaveBeenCalled();
    });
  });

  it('shows a load failure with Retry for policy instead of the group editor', async () => {
    const fetchPolicy = vi.fn().mockResolvedValue(undefined);
    useCortexStore.setState({
      policyConfig: null,
      policyLoadError: 'Could not load policy (HTTP 500).',
      fetchPolicy,
    });

    render(<ApprovalsTab />);
    fireEvent.click(screen.getByRole('button', { name: 'Policy' }));

    const failure = await screen.findByTestId('policy-load-failure');
    expect(failure.textContent).toMatch(/Could not load policy/);

    fireEvent.click(screen.getByRole('button', { name: /Retry/i }));
    await waitFor(() => {
      expect(fetchPolicy).toHaveBeenCalled();
    });
  });
});
