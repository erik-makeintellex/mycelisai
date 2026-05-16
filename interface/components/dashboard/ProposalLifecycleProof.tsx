"use client";

import Link from "next/link";
import { AlertTriangle, CheckCircle2, Clock3, ExternalLink, ShieldOff, XCircle } from "lucide-react";
import type { ProposalLifecycleStatus } from "@/store/useCortexStore";

type RenderedProposalLifecycle = ProposalLifecycleStatus | "confirmed_pending_execution";

export default function ProposalLifecycleProof({
    lifecycle,
    runId,
}: {
    lifecycle: RenderedProposalLifecycle;
    runId?: string;
}) {
    if (lifecycle === "active") return null;

    const proof = proofFor(lifecycle, runId);
    const Icon = proof.icon;

    return (
        <div className={`rounded border px-3 py-2 text-xs leading-5 ${proof.className}`}>
            <div className="flex flex-wrap items-center gap-x-2 gap-y-1">
                <Icon className="h-3.5 w-3.5 shrink-0" />
                <span className="font-mono uppercase tracking-[0.14em]">{proof.label}</span>
                <span className="text-current/85">{proof.detail}</span>
                {runId && lifecycle === "executed" ? (
                    <Link href={`/runs/${runId}`} className="inline-flex items-center gap-1 font-mono underline underline-offset-2">
                        Open run proof
                        <ExternalLink className="h-3 w-3" />
                    </Link>
                ) : null}
            </div>
        </div>
    );
}

function proofFor(lifecycle: RenderedProposalLifecycle, runId?: string) {
    if (lifecycle === "executed" && runId) {
        return {
            icon: CheckCircle2,
            label: "Proof created",
            detail: "The approved action produced a durable run proof. The result can be revisited and audited.",
            className: "border-cortex-success/30 bg-cortex-success/10 text-cortex-success",
        };
    }
    if (lifecycle === "confirmed_pending_execution") {
        return {
            icon: Clock3,
            label: "Approved, not yet trusted",
            detail: "Approval was recorded, but no durable run proof has arrived yet. Wait for proof before relying on changes.",
            className: "border-amber-400/25 bg-amber-400/10 text-amber-300",
        };
    }
    if (lifecycle === "failed") {
        return {
            icon: AlertTriangle,
            label: "No trusted execution",
            detail: "Confirmation did not execute. No run proof was created, so requested changes should not be trusted.",
            className: "border-red-400/30 bg-red-400/10 text-red-300",
        };
    }
    return {
        icon: lifecycle === "cancelled" ? XCircle : ShieldOff,
        label: "No action executed",
        detail: "The proposal did not run, so it changed nothing and created no execution proof.",
        className: "border-cortex-border bg-cortex-bg/60 text-cortex-text-muted",
    };
}
