import { Wrench } from "lucide-react";

export function MCPRegistryEmptyBanner() {
    return (
        <div className="rounded-xl border border-cortex-warning/25 bg-cortex-warning/10 px-4 py-3">
            <p className="text-[10px] font-mono font-bold uppercase tracking-wider text-cortex-warning">
                No MCP servers installed yet
            </p>
            <p className="mt-1 text-xs font-mono leading-5 text-cortex-text-main">
                Default bootstrap is disabled in the compose home runtime, so Add MCP Server is the first activation step for connected capabilities.
            </p>
            <p className="mt-2 text-[10px] font-mono leading-5 text-cortex-text-muted">
                Install a curated server such as filesystem or fetch, then return here to confirm the server card and recent activity appear.
            </p>
        </div>
    );
}

export function MCPRegistryErrorBanner({ error }: { error: string | null }) {
    return (
        <div className="rounded-xl border border-cortex-danger/25 bg-cortex-danger/10 px-4 py-3">
            <p className="text-[10px] font-mono font-bold uppercase tracking-wider text-cortex-danger">
                MCP registry unreachable
            </p>
            <p className="mt-1 text-xs font-mono leading-5 text-cortex-text-main">
                Capabilities could not confirm installed MCP servers, so this is not treated as an empty registry.
            </p>
            <p className="mt-2 text-[10px] font-mono leading-5 text-cortex-text-muted">{error}</p>
        </div>
    );
}

export function MCPInstallNotice({ message }: { message: string }) {
    return (
        <div className="rounded-xl border border-cortex-success/25 bg-cortex-success/10 px-4 py-3">
            <p className="text-xs font-mono leading-5 text-cortex-text-main">{message}</p>
        </div>
    );
}

export function MCPRegistryEmptyHero({ onRequest }: { onRequest: () => void }) {
    return (
        <div className="flex flex-col items-center justify-center py-24 text-cortex-text-muted">
            <Wrench className="w-12 h-12 mb-3 opacity-20" />
            <p className="text-sm font-mono">No MCP servers installed.</p>
            <p className="text-[10px] font-mono mt-1 opacity-50">
                Use Add MCP Server to install the first approved tool server for this group.
            </p>
            <button
                onClick={onRequest}
                className="mt-4 rounded-lg border border-cortex-primary/30 bg-cortex-primary/10 px-3 py-1.5 text-[10px] font-mono font-bold text-cortex-primary transition-colors hover:bg-cortex-primary/20"
            >
                REQUEST FIRST CAPABILITY
            </button>
        </div>
    );
}
