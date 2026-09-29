import { beforeEach, describe, expect, it, vi } from "vitest";
import { fireEvent, render, screen } from "@testing-library/react";
import ApprovalsTab from "@/components/automations/ApprovalsTab";
import { useCortexStore } from "@/store/useCortexStore";

vi.mock("reactflow", async () => await import("../mocks/reactflow"));
vi.mock("@/components/workspace/TrustSlider", () => ({ default: () => <div /> }));
vi.mock("@/components/dashboard/ManifestationPanel", () => ({ default: () => <div /> }));

describe("ApprovalsTab audit section load error", () => {
  const fetchAuditLog = vi.fn().mockResolvedValue(undefined);

  beforeEach(() => {
    fetchAuditLog.mockClear();
    useCortexStore.setState({
      pendingApprovals: [],
      isFetchingApprovals: false,
      approvalsError: null,
      resolveApprovalError: null,
      fetchPendingApprovals: vi.fn().mockResolvedValue(undefined),
      resolveApproval: vi.fn().mockResolvedValue(undefined),
      policyConfig: null,
      policyError: null,
      auditLog: [],
      isFetchingAuditLog: false,
      auditLogError: null,
      fetchAuditLog,
    });
  });

  it("shows a blocker with retry, not the empty state, when the audit load failed", () => {
    useCortexStore.setState({ auditLogError: { code: "admin_required", httpStatus: 403 } });
    render(<ApprovalsTab />);
    fireEvent.click(screen.getByRole("button", { name: /Audit/ }));

    expect(screen.getByRole("alert")).toBeDefined();
    expect(screen.queryByText("No recent audit activity")).toBeNull();
    fetchAuditLog.mockClear();
    fireEvent.click(screen.getByRole("button", { name: "Try again" }));
    expect(fetchAuditLog).toHaveBeenCalledTimes(1);
  });

  it("shows the empty state only when there is no error", () => {
    render(<ApprovalsTab />);
    fireEvent.click(screen.getByRole("button", { name: /Audit/ }));

    expect(screen.getByText("No recent audit activity")).toBeDefined();
    expect(screen.queryByRole("alert")).toBeNull();
  });
});
