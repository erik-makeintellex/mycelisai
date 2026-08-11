import { describe, expect, it } from "vitest";
import { conversationalProposalReply } from "@/components/dashboard/conversationalProposalReply";

describe("conversationalProposalReply", () => {
  it.each(["approve", "Approve.", "go ahead", "yes proceed"])(
    "recognizes a bounded approval reply: %s",
    (reply) => expect(conversationalProposalReply(reply)).toBe("confirm"),
  );

  it.each(["cancel", "never mind", "do not run"])(
    "recognizes a bounded cancellation reply: %s",
    (reply) => expect(conversationalProposalReply(reply)).toBe("cancel"),
  );

  it.each([
    "approve after changing the title",
    "go ahead and add another team",
    "I am not sure",
    "please revise the plan",
  ])("leaves substantive replies for Soma to interpret: %s", (reply) => {
    expect(conversationalProposalReply(reply)).toBeNull();
  });
});
