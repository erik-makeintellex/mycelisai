"use client";

import { ArrowRight } from "lucide-react";

export function ContextMetric({ label, value }: { label: string; value: string }) {
    return (
        <div className="rounded-2xl border border-cortex-border bg-cortex-surface px-4 py-3">
            <p className="text-[10px] font-mono uppercase tracking-[0.18em] text-cortex-text-muted">{label}</p>
            <p className="mt-2 text-sm font-medium text-cortex-text-main">{value}</p>
        </div>
    );
}

export function NextStepCard({ title, detail }: { title: string; detail: string }) {
    return (
        <div className="rounded-2xl border border-cortex-border bg-cortex-surface px-4 py-3">
            <div className="flex items-start justify-between gap-3">
                <div>
                    <p className="text-sm font-semibold text-cortex-text-main">{title}</p>
                    <p className="mt-2 text-sm leading-6 text-cortex-text-muted">{detail}</p>
                </div>
                <ArrowRight className="mt-1 h-4 w-4 text-cortex-text-muted" />
            </div>
        </div>
    );
}
