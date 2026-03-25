"use client";

import Link from "next/link";
import { ArrowLeft, Eye } from "lucide-react";

export default function AdvancedModeGate({
    title,
    summary,
    returnHref = "/dashboard",
    returnLabel = "Return to AI Organization",
}: {
    title: string;
    summary: string;
    returnHref?: string;
    returnLabel?: string;
}) {
    return (
        <div className="flex h-full items-center justify-center bg-cortex-bg px-6 py-10">
            <div className="max-w-xl rounded-3xl border border-cortex-border bg-cortex-surface p-6 shadow-[0_18px_40px_rgba(148,163,184,0.16)]">
                <div className="inline-flex items-center gap-2 rounded-full border border-cortex-primary/25 bg-cortex-primary/10 px-3 py-1 text-[11px] font-mono uppercase tracking-[0.22em] text-cortex-primary">
                    <Eye className="h-3.5 w-3.5" />
                    Advanced Mode
                </div>
                <h1 className="mt-4 text-2xl font-semibold tracking-tight text-cortex-text-main">{title}</h1>
                <p className="mt-3 text-sm leading-7 text-cortex-text-muted">{summary}</p>
                <p className="mt-3 text-sm leading-7 text-cortex-text-muted">
                    Turn on Advanced mode from the left rail when you want to inspect deeper tools, system details, or configuration surfaces.
                </p>
                <div className="mt-5">
                    <Link
                        href={returnHref}
                        className="inline-flex items-center gap-2 rounded-xl border border-cortex-border bg-cortex-bg px-3 py-2 text-sm font-medium text-cortex-text-main transition-colors hover:border-cortex-primary/20"
                    >
                        <ArrowLeft className="h-4 w-4" />
                        {returnLabel}
                    </Link>
                </div>
            </div>
        </div>
    );
}
