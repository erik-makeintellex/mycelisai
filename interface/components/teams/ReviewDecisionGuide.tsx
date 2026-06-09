"use client";

import type { TeamWorkItem } from "@/store/useCortexStore";
import {
  recommendedReviewChoice,
  reviewReason,
  trustedState,
} from "./activeWorkCompact";

export function ReviewDecisionGuide({
  item,
  concise = false,
}: {
  item: TeamWorkItem;
  concise?: boolean;
}) {
  const labels = concise
    ? ["Reason", "Trust", "Move"]
    : ["Why review", "What is safe to rely on", "Best next move"];
  return (
    <div className="mt-3 grid gap-2 rounded-lg border border-cortex-border bg-cortex-surface px-3 py-3 text-sm leading-6 md:grid-cols-3">
      <div>
        <p className="font-mono text-[10px] uppercase tracking-[0.14em] text-cortex-primary">
          {labels[0]}
        </p>
        <p className="mt-1 text-cortex-text-muted">{reviewReason(item)}</p>
      </div>
      <div>
        <p className="font-mono text-[10px] uppercase tracking-[0.14em] text-cortex-primary">
          {labels[1]}
        </p>
        <p className="mt-1 text-cortex-text-muted">{trustedState(item)}</p>
      </div>
      <div>
        <p className="font-mono text-[10px] uppercase tracking-[0.14em] text-cortex-primary">
          {labels[2]}
        </p>
        <p className="mt-1 text-cortex-text-main">
          {recommendedReviewChoice(item)}
        </p>
      </div>
    </div>
  );
}
