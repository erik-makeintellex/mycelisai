import type { TeamInteraction, TeamWorkItem, TeamWorkItemState } from "@/store/useCortexStore";

const reviewStatePriority: Record<TeamWorkItemState, number> = {
  needs_operator: 0,
  degraded: 1,
  running: 2,
  reviewing: 3,
  queued: 4,
  output_ready: 5,
  paused: 6,
  briefed: 7,
  new: 8,
  archived: 99,
};

export function recoveryReviewQueueItems(items: TeamWorkItem[]) {
  return [...items]
    .filter(isRecoveryReviewQueueItem)
    .sort((left, right) => {
      const leftPriority = recoveryReviewPriority(left);
      const rightPriority = recoveryReviewPriority(right);
      if (leftPriority !== rightPriority) return leftPriority - rightPriority;
      const leftTime = left.updatedAt ? Date.parse(left.updatedAt) : 0;
      const rightTime = right.updatedAt ? Date.parse(right.updatedAt) : 0;
      return rightTime - leftTime;
    });
}

export function recoveryReviewQueueCount(items: TeamWorkItem[]) {
  return items.filter(isRecoveryReviewQueueItem).length;
}

export function isRecoveryReviewQueueItem(item: TeamWorkItem) {
  if (item.state === "archived") return false;
  if (item.needsOperator) return true;
  if (item.recoveryOptions?.length) return true;
  if (hasEnabledAction(item, "recover")) return true;
  return [
    "degraded",
    "needs_operator",
    "output_ready",
    "reviewing",
    "running",
    "queued",
  ].includes(item.state);
}

function recoveryReviewPriority(item: TeamWorkItem) {
  const sourcePenalty = item.source === "projection" ? 20 : 0;
  const operatorBoost = item.needsOperator ? -5 : 0;
  const recoveryBoost = hasEnabledAction(item, "recover") || item.recoveryOptions?.length ? -0.5 : 0;
  return reviewStatePriority[item.state] + sourcePenalty + operatorBoost + recoveryBoost;
}

function hasEnabledAction(item: TeamWorkItem, action: TeamInteraction["action"]) {
  return item.interactions.some((candidate) => candidate.action === action && !candidate.disabled);
}
