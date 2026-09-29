import { beforeEach, describe, expect, it, vi } from "vitest";
import { fireEvent, render, screen } from "@testing-library/react";

vi.mock("reactflow", async () => {
  const mock = await import("../mocks/reactflow");
  return mock;
});

vi.mock("@/components/dashboard/ManifestationPanel", () => ({
  __esModule: true,
  default: () => <div data-testid="manifestation-panel">ManifestationPanel</div>,
}));

import ApprovalsTab from "@/components/automations/ApprovalsTab";
import { useCortexStore } from "@/store/useCortexStore";

describe("ApprovalsTab blockers (no fake success)", () => {
  beforeEach(() => {
    useCortexStore.setState({
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
