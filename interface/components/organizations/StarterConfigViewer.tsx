"use client";

// RC-004G-2: lets an operator open the governed ConfigViewer for an
// organization starter template directly from the setup starter cards.
// Advanced-mode only, and never selects the card underneath it.

import { useState } from "react";
import { useCortexStore } from "@/store/useCortexStore";
import { ConfigViewer } from "@/components/deployments/ConfigViewer";

export function StarterConfigViewerAction({ templateId }: { templateId: string }) {
    const advancedMode = useCortexStore((s) => s.advancedMode);
    const [open, setOpen] = useState(false);

    if (!advancedMode) {
        return null;
    }

    return (
        <>
            <button
                type="button"
                data-testid={`starter-card-${templateId}-view-config`}
                onClick={(event) => {
                    event.stopPropagation();
                    setOpen(true);
                }}
                className="rounded-full border border-cortex-border bg-cortex-surface px-2 py-0.5 text-[10px] font-mono uppercase tracking-[0.14em] text-cortex-text-muted transition-colors hover:border-cortex-primary/30 hover:text-cortex-primary"
            >
                View config
            </button>
            {open ? (
                <div
                    data-testid={`starter-card-${templateId}-config-overlay`}
                    className="fixed inset-0 z-50 bg-cortex-bg/40"
                    onClick={(event) => {
                        // ConfigViewer fills this box via absolute positioning; only
                        // a click on the backdrop itself (not a descendant) closes it.
                        if (event.target === event.currentTarget) {
                            setOpen(false);
                        }
                    }}
                >
                    <ConfigViewer
                        configRef={{
                            kind: "template",
                            id: templateId,
                            path: `/api/v1/config/templates/${templateId}`,
                        }}
                        onClose={() => setOpen(false)}
                    />
                </div>
            ) : null}
        </>
    );
}
