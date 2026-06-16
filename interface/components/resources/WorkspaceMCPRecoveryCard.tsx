"use client";

import { AlertTriangle, Database } from "lucide-react";
import type { ReactNode } from "react";

export default function WorkspaceMCPRecoveryCard({
    title,
    detail,
    onOpenToolsTab,
    onRefresh,
}: {
    title: string;
    detail: ReactNode;
    onOpenToolsTab: () => void;
    onRefresh: () => void;
}) {
    return (
        <div className="h-full p-6">
            <div className="max-w-3xl rounded-xl border border-cortex-warning/30 bg-cortex-warning/10 p-5">
                <div className="mb-2 flex items-center gap-2 text-cortex-warning">
                    <AlertTriangle className="h-4 w-4" />
                    <h3 className="text-sm font-semibold">{title}</h3>
                </div>
                <p className="mb-3 text-xs leading-5 text-cortex-text-muted">{detail}</p>
                <div className="flex flex-wrap gap-2">
                    <button
                        type="button"
                        onClick={onOpenToolsTab}
                        className="rounded border border-cortex-primary/30 px-3 py-1.5 text-xs font-mono text-cortex-primary hover:bg-cortex-primary/10"
                    >
                        Open Capabilities
                    </button>
                    <a
                        href="/system?tab=deployments"
                        className="inline-flex items-center gap-1 rounded border border-cortex-border px-3 py-1.5 text-xs font-mono text-cortex-text-main hover:bg-cortex-border"
                    >
                        <Database className="h-3.5 w-3.5" />
                        View storage roots
                    </a>
                    <button
                        type="button"
                        onClick={onRefresh}
                        className="rounded border border-cortex-border px-3 py-1.5 text-xs font-mono text-cortex-text-main hover:bg-cortex-border"
                    >
                        Refresh
                    </button>
                </div>
            </div>
        </div>
    );
}
