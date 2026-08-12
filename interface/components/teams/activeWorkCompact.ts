import type { TeamInteraction, TeamWorkItem } from "@/store/useCortexStore";
import {
  DEGRADED_TEAM_WORK_REVIEW_COPY,
  NEEDS_OPERATOR_REVIEW_COPY,
  OUTPUT_READY_REVIEW_COPY,
  QUEUED_RECOVERY_REVIEW_COPY,
  STALE_FAILED_PLAN_REVIEW_COPY,
} from "@/lib/deliveryRuntimeLanguage";

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
  if (isStaleFailedPlanItem(item)) return STALE_FAILED_PLAN_REVIEW_COPY.title;
  if (needsExternalMutationVerification(item)) return "External result needs verification";
  if (item.state === "degraded") return DEGRADED_TEAM_WORK_REVIEW_COPY.title;
  if (item.state === "needs_operator") return NEEDS_OPERATOR_REVIEW_COPY.title;
  if (item.state === "queued" && isRecoveryRequest(item)) return QUEUED_RECOVERY_REVIEW_COPY.title;
  return item.title;
}

export function compactDescription(item: TeamWorkItem) {
  if (isStaleFailedPlanItem(item)) {
    return STALE_FAILED_PLAN_REVIEW_COPY.description;
  }
  if (needsExternalMutationVerification(item)) {
    return "The external action may have completed, so retrying now could apply it twice.";
  }
  if (item.state === "degraded") {
    return DEGRADED_TEAM_WORK_REVIEW_COPY.description;
  }
  if (item.state === "needs_operator") {
    return NEEDS_OPERATOR_REVIEW_COPY.description;
  }
  if (item.state === "queued" && isRecoveryRequest(item)) {
    return QUEUED_RECOVERY_REVIEW_COPY.description;
  }
  return item.description;
}

export function compactNextAction(item: TeamWorkItem) {
  if (isStaleFailedPlanItem(item)) {
    return STALE_FAILED_PLAN_REVIEW_COPY.nextAction;
  }
  if (needsExternalMutationVerification(item)) {
    return "Tell Soma what you observed so it can verify the external result before considering a retry.";
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
    return STALE_FAILED_PLAN_REVIEW_COPY.reason;
  }
  if (needsExternalMutationVerification(item)) {
    return "The request reached an external system, but Mycelis could not confirm whether the change completed.";
  }
  if (item.state === "degraded") {
    return DEGRADED_TEAM_WORK_REVIEW_COPY.reason;
  }
  if (item.state === "needs_operator") {
    return NEEDS_OPERATOR_REVIEW_COPY.reason;
  }
  if (item.state === "output_ready") {
    return OUTPUT_READY_REVIEW_COPY.reason;
  }
  return "This work item is still active or retained and may need a decision before it leaves review.";
}

export function trustedState(item: TeamWorkItem) {
  if (isStaleFailedPlanItem(item)) {
    return STALE_FAILED_PLAN_REVIEW_COPY.trustedState;
  }
  if (needsExternalMutationVerification(item)) {
    return "Trusted: the request, approval, and attempt record. Not trusted yet: whether the external change completed.";
  }
  if (item.state === "degraded" || item.state === "needs_operator") {
    return item.state === "degraded" ? DEGRADED_TEAM_WORK_REVIEW_COPY.trustedState : NEEDS_OPERATOR_REVIEW_COPY.trustedState;
  }
  if (item.state === "output_ready") {
    return OUTPUT_READY_REVIEW_COPY.trustedState;
  }
  return "Trusted so far: durable work state. Output may change until the item finishes.";
}

export function recommendedReviewChoice(item: TeamWorkItem) {
  if (isStaleFailedPlanItem(item)) {
    return STALE_FAILED_PLAN_REVIEW_COPY.recommendedChoice;
  }
  if (needsExternalMutationVerification(item)) {
    return "Tell Soma what you can observe in the external system. Do not retry until the result is verified.";
  }
  if (item.state === "degraded") {
    return DEGRADED_TEAM_WORK_REVIEW_COPY.recommendedChoice;
  }
  if (item.state === "needs_operator") {
    return NEEDS_OPERATOR_REVIEW_COPY.recommendedChoice;
  }
  if (item.state === "output_ready") {
    return OUTPUT_READY_REVIEW_COPY.recommendedChoice;
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

function needsExternalMutationVerification(item: TeamWorkItem) {
  return item.degradationState === "external_mutation_outcome_unknown";
}
