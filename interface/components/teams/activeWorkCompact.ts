import type { TeamInteraction, TeamWorkItem } from "@/store/useCortexStore";

export function compactActions(item: TeamWorkItem) {
  const actions = item.interactions;
  const actionByKind = new Map(actions.map((action) => [action.action, action]));
  const order = isStaleFailedPlanItem(item)
    ? ["archive", "inspect", "recover"]
    : ["recover", "archive", "inspect"];
  return order
    .map((kind) => actionByKind.get(kind as TeamInteraction["action"]))
    .filter((action): action is TeamInteraction => Boolean(action && !action.disabled));
}

export function compactTitle(item: TeamWorkItem) {
  if (isStaleFailedPlanItem(item)) return "Old proposal cannot run";
  if (item.state === "degraded") return "Team work needs recovery";
  if (item.state === "needs_operator") return "Team needs your response";
  if (item.state === "queued" && isRecoveryRequest(item)) return "Recovery request queued";
  return item.title;
}

export function compactDescription(item: TeamWorkItem) {
  if (isStaleFailedPlanItem(item)) {
    return "Soma could not find an approved execution plan for this older proposal. Nothing changed, and there is no completed output to trust.";
  }
  if (item.state === "degraded") {
    return "The team did not finish this work. The saved work item is still available, but no output should be trusted yet.";
  }
  if (item.state === "needs_operator") {
    return "The team is waiting for missing direction before it can continue.";
  }
  if (item.state === "queued" && isRecoveryRequest(item)) {
    return "Soma has queued a recovery attempt for this work item. Wait for a new output or proof before trusting the result.";
  }
  return item.description;
}

export function compactNextAction(item: TeamWorkItem) {
  if (isStaleFailedPlanItem(item)) {
    return "Clear this from review. Nothing ran, and there is no output to trust. Start a new Soma ask if you still want this work.";
  }
  if (item.state === "degraded") {
    return item.nextAction ?? "Recover this work item when the team runtime is available, or archive it if this was only test data.";
  }
  if (item.state === "needs_operator") {
    return item.nextAction ?? "Respond with the missing direction, then let the team continue.";
  }
  return item.nextAction;
}

export function reviewReason(item: TeamWorkItem) {
  if (isStaleFailedPlanItem(item)) {
    return "The original proposal is not executable anymore because no approved execution plan was retained for it.";
  }
  if (item.state === "degraded") {
    return "The team did not finish cleanly, so the result needs recovery or cleanup before you rely on it.";
  }
  if (item.state === "needs_operator") {
    return "The team needs a decision or missing direction before it can safely continue.";
  }
  if (item.state === "output_ready") {
    return "The team produced retained output. Review it, then continue or archive the work item.";
  }
  return "This work item is still active or retained and may need a decision before it leaves review.";
}

export function trustedState(item: TeamWorkItem) {
  if (isStaleFailedPlanItem(item)) {
    return "Trusted: the failure record and audit trail. Not trusted: any implied output from this attempt.";
  }
  if (item.state === "degraded" || item.state === "needs_operator") {
    return "Trusted: retained context, proof refs, and status history. Not trusted: unfinished output.";
  }
  if (item.state === "output_ready") {
    return "Trusted after review: retained outputs, proof refs, and run history shown on this item.";
  }
  return "Trusted so far: durable work state. Output may change until the item finishes.";
}

export function recommendedReviewChoice(item: TeamWorkItem) {
  if (isStaleFailedPlanItem(item)) {
    return "Clear from review if this is old test data. Inspect only if you need the failure details.";
  }
  if (item.state === "degraded") {
    return "Recover when the runtime dependency is available, or archive if this attempt is no longer useful.";
  }
  if (item.state === "needs_operator") {
    return "Respond or steer the work with the missing decision.";
  }
  if (item.state === "output_ready") {
    return "Open the output, verify the proof, then archive or ask for follow-up.";
  }
  return "Inspect for context, pause if it should stop, or wait for the next retained event.";
}

export function isStaleFailedPlanItem(item: TeamWorkItem) {
  const text = `${item.title} ${item.description ?? ""} ${item.nextAction ?? ""} ${item.recoveryOptions?.join(" ") ?? ""}`;
  return /no approved execution plan was stored/i.test(text);
}

function isRecoveryRequest(item: TeamWorkItem) {
  return /recovery requested/i.test(`${item.title} ${item.description ?? ""} ${item.nextAction ?? ""}`);
}
