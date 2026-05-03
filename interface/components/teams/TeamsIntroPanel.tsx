"use client";

import Link from "next/link";
import { Sparkles } from "lucide-react";

export default function TeamsIntroPanel() {
  return (
    <div className="rounded-2xl border border-cortex-border bg-cortex-surface px-4 py-4">
      <div className="flex items-start gap-3">
        <div className="rounded-full border border-cortex-primary/20 bg-cortex-primary/10 p-2 text-cortex-primary">
          <Sparkles className="h-4 w-4" />
        </div>
        <div className="flex-1">
          <div className="flex flex-wrap items-center gap-2">
            <p className="text-sm font-semibold text-cortex-text-main">
              Specialize new teams through Soma
            </p>
            <span className="rounded-full border border-cortex-primary/20 bg-cortex-primary/10 px-2.5 py-1 text-[10px] font-mono uppercase tracking-[0.16em] text-cortex-primary">
              Soma-first
            </span>
          </div>
          <p className="mt-2 text-sm leading-6 text-cortex-text-muted">
            Start new teams through a guided creation flow, then hand the
            finished ask to Soma. Return here after launch to open the team
            lead, review the current roster, and maintain the member templates
            Soma should reuse on future team builds.
          </p>
          <div className="mt-4 flex flex-wrap gap-2">
            <Link href="/teams/create" className="inline-flex items-center justify-center rounded-2xl border border-cortex-primary/30 px-4 py-2 text-sm font-semibold text-cortex-primary hover:bg-cortex-primary/10">
              Open guided team creation
            </Link>
            <Link href="/dashboard" className="inline-flex items-center justify-center rounded-2xl border border-cortex-border px-4 py-2 text-sm font-semibold text-cortex-text-main hover:bg-cortex-border">
              Open Soma workspace
            </Link>
            <Link href="/groups" className="inline-flex items-center justify-center rounded-2xl border border-cortex-border px-4 py-2 text-sm font-semibold text-cortex-text-main hover:bg-cortex-border">
              Open group operations
            </Link>
          </div>
          <div className="mt-4 grid gap-3 md:grid-cols-3">
            <QuickFact
              label="Guide the request"
              value="Use the guided team-creation flow to define the lane, lead posture, and expected outputs before handoff."
            />
            <QuickFact
              label="Work through the lead"
              value="Each team opens around one focused lead counterpart instead of another generic global chat."
            />
            <QuickFact
              label="Templates guide member choice"
              value="Soma can reuse the configured team-member templates below when a lane needs specific agent behavior."
            />
          </div>
        </div>
      </div>
    </div>
  );
}

function QuickFact({ label, value }: { label: string; value: string }) {
  return (
    <div className="rounded-2xl border border-cortex-border bg-cortex-bg px-4 py-3">
      <p className="text-[10px] font-mono uppercase tracking-[0.18em] text-cortex-text-muted">
        {label}
      </p>
      <p className="mt-2 text-sm leading-6 text-cortex-text-main">{value}</p>
    </div>
  );
}
