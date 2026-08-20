import { fireEvent, render, screen, waitFor } from "@testing-library/react";
import { describe, expect, it, vi } from "vitest";
import { ExternalOutcomeVerificationForm } from "@/components/teams/ExternalOutcomeVerificationForm";
import type { TeamWorkItem } from "@/store/useCortexStore";

const item: TeamWorkItem = {
  id: "work-external",
  title: "Update external account",
  state: "degraded",
  degradationState: "external_mutation_outcome_unknown",
  ownerLabel: "Soma",
  scopeLabel: "Delegated work",
  teamIds: ["team-external"],
  interactions: [],
};

describe("ExternalOutcomeVerificationForm", () => {
  it("requires an observation and submits one normalized verification", async () => {
    const onVerify = vi.fn().mockResolvedValue(undefined);
    render(<ExternalOutcomeVerificationForm item={item} onVerify={onVerify} />);

    const submit = screen.getByRole("button", { name: "Submit verification" });
    expect((submit as HTMLButtonElement).disabled).toBe(true);
    expect(screen.getByText(/does not retry the action/i)).toBeDefined();

    fireEvent.click(screen.getByRole("radio", { name: "Confirmed not applied" }));
    expect(screen.getByText(/receipt or audit confirms no change was applied/i)).toBeDefined();
    expect((submit as HTMLButtonElement).disabled).toBe(true);

    fireEvent.change(screen.getByRole("textbox", { name: /What did you observe/i }), {
      target: { value: "The provider receipt confirms the update was rejected before commit." },
    });
    fireEvent.change(screen.getByRole("textbox", { name: /Evidence references/i }), {
      target: { value: "receipt-42, https://example.test/audit/42, receipt-42" },
    });
    expect((submit as HTMLButtonElement).disabled).toBe(false);
    fireEvent.click(submit);

    await waitFor(() => {
      expect(onVerify).toHaveBeenCalledWith(item, {
        result: "not_committed",
        summary: "The provider receipt confirms the update was rejected before commit.",
        evidenceRefs: ["receipt-42", "https://example.test/audit/42"],
      });
    });
    expect(screen.getByRole("status").textContent).toMatch(/no retry was requested/i);
    expect((screen.getByRole("button", { name: "Submit verification" }) as HTMLButtonElement).disabled).toBe(true);
  });

  it("explains that an unclear result keeps retry unavailable", () => {
    render(<ExternalOutcomeVerificationForm item={item} onVerify={vi.fn()} compact />);

    fireEvent.click(screen.getByRole("radio", { name: "Still unclear" }));

    expect(screen.getByText("Keeps the outcome unknown and retry unavailable.")).toBeDefined();
  });
});
