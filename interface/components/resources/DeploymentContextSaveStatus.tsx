import { AlertTriangle, ShieldAlert, ShieldCheck } from "lucide-react";

export type SaveState =
    | { kind: "idle" }
    | { kind: "saving" }
    | { kind: "saved"; message: string; semantic: boolean }
    | { kind: "error"; message: string };

const KEYWORD_ONLY = "Saved. Soma can recall this by keywords; semantic search needs an embedding engine.";

/** Builds the saved state from the server response; never claims more than the server reported. */
export function savedStateFrom(payload: { status_message?: unknown; retrieval_modes?: unknown }): SaveState {
    const modes = Array.isArray(payload.retrieval_modes) ? payload.retrieval_modes : [];
    const message = typeof payload.status_message === "string" && payload.status_message.trim() ? payload.status_message : KEYWORD_ONLY;
    return { kind: "saved", message, semantic: modes.includes("semantic") };
}

/** Inline save status shown directly above the Save button. */
export default function DeploymentContextSaveStatus({ state }: { state: SaveState }) {
    if (state.kind === "error") {
        return (
            <div role="alert" className="flex items-start gap-2 rounded-lg border border-cortex-danger/40 bg-cortex-danger/10 px-3 py-2">
                <AlertTriangle className="mt-0.5 h-4 w-4 flex-shrink-0 text-cortex-danger" />
                <div className="min-w-0">
                    <p className="text-sm font-semibold text-cortex-danger">Not saved</p>
                    <p className="mt-0.5 break-words text-xs text-cortex-text-main">{state.message}</p>
                </div>
            </div>
        );
    }
    if (state.kind === "saved") {
        const Icon = state.semantic ? ShieldCheck : ShieldAlert;
        return (
            <div role="status" className="flex items-start gap-2 text-cortex-text-main">
                <Icon className={`mt-0.5 h-4 w-4 flex-shrink-0 ${state.semantic ? "text-cortex-success" : "text-cortex-warning"}`} />
                <p className="text-xs">{state.message}</p>
            </div>
        );
    }
    return (
        <p role="status" className="text-[11px] text-cortex-text-muted">
            {state.kind === "saving" ? "Saving..." : "Not saved yet. Soma uses saved sources in later chats."}
        </p>
    );
}
