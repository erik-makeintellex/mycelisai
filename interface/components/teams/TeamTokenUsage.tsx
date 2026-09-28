"use client";

import { useEffect, useState } from "react";
import { formatTokenCount, getTokenBudgetUsage, type TokenBudgetUsage } from "@/lib/tokenBudgets";

// B1 D10 UI summary line: "This team used 12k of 64k tokens this run · 180k
// of 2M today." Not mounted anywhere yet — its host, TeamDetailDrawer.tsx, is
// owned by the U1 lane (see the mount spec in the B1-W2 close-out report).
// Renders nothing while loading and on any fetch failure (including the
// non-member 403 `token_budget_usage_forbidden`): never a fake zero.

export interface TeamTokenUsageProps {
    teamId: string;
    /** This run's tokens used/limit so far, from the caller's own execution
     * summary (the run scope has no separate ownership check to add here). */
    runUsed: number;
    runLimit: number;
}

export default function TeamTokenUsage({ teamId, runUsed, runLimit }: TeamTokenUsageProps) {
    const [dayUsage, setDayUsage] = useState<TokenBudgetUsage | null>(null);

    useEffect(() => {
        let cancelled = false;
        setDayUsage(null);
        (async () => {
            const result = await getTokenBudgetUsage({ team_id: teamId });
            if (!cancelled && result.ok) {
                setDayUsage(result.data);
            }
        })();
        return () => {
            cancelled = true;
        };
    }, [teamId]);

    if (!dayUsage) return null;

    const dayUsed = `${dayUsage.usage_reported ? "" : "at least "}${formatTokenCount(dayUsage.used)}`;
    const dayLabel = dayUsage.period === "since_restart" ? "since restart" : "today";
    // B1R-A: `reserved` (the output allowance held by in-flight calls) is
    // optional on this branch's base contract; only append the clause when
    // present and non-zero, never claim a reservation that wasn't reported.
    const reservedClause = dayUsage.reserved ? ` · ${formatTokenCount(dayUsage.reserved)} reserved` : "";

    return (
        <p className="text-[11px] text-cortex-text-muted">
            {`This team used ${formatTokenCount(runUsed)} of ${formatTokenCount(runLimit)} tokens this run · ${dayUsed} of ${formatTokenCount(dayUsage.limit)} ${dayLabel}${reservedClause}.`}
        </p>
    );
}
