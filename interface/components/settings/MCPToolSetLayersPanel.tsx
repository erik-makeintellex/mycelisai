"use client";

import type { FormEvent } from "react";
import { useMemo, useState } from "react";
import { Layers3, RefreshCw, Save, ShieldCheck } from "lucide-react";
import type { MCPToolSet, MCPToolSetCreate, MCPToolSetScopeKind } from "@/store/useCortexStore";
import { useCortexStore } from "@/store/useCortexStore";
import { MCPToolSetCommonChoices } from "./MCPToolSetCommonChoices";

type SaveState = "idle" | "saving" | "saved" | "error";

type Props = {
    toolSets: MCPToolSet[];
    isLoading: boolean;
    error: string | null;
    onRefresh: () => void;
    onCreate: (input: MCPToolSetCreate) => Promise<boolean>;
};
const scopes: Array<{ kind: MCPToolSetScopeKind; label: string; help: string; example: string }> = [
    { kind: "all", label: "Everyone", help: "Default tools Soma may use across this workspace.", example: "Workspace files" },
    { kind: "group", label: "Group", help: "Tools for one Outcome or collaboration lane.", example: "Marketing deliverables" },
    { kind: "host", label: "Host", help: "Tools limited to one runtime machine or service host.", example: "Local media node" },
];

export function MCPToolSetLayersPanel({ toolSets, isLoading, error, onRefresh, onCreate }: Props) {
    const [name, setName] = useState("");
    const [description, setDescription] = useState("");
    const [scopeKind, setScopeKind] = useState<MCPToolSetScopeKind>("all");
    const [scopeRef, setScopeRef] = useState("");
    const [toolRefsText, setToolRefsText] = useState("");
    const [formError, setFormError] = useState<string | null>(null);
    const [saveState, setSaveState] = useState<SaveState>("idle");
    const counts = useMemo(() => {
        const result = { all: 0, group: 0, host: 0 };
        for (const toolSet of toolSets) {
            if (toolSet.scope_kind === "all" || toolSet.scope_kind === "group" || toolSet.scope_kind === "host") {
                result[toolSet.scope_kind] += 1;
            }
        }
        return result;
    }, [toolSets]);
    async function handleSubmit(event: FormEvent<HTMLFormElement>) {
        event.preventDefault();
        const parsedRefs = parseToolRefs(toolRefsText);
        const nextName = name.trim();
        const target = scopeRef.trim();
        if (!nextName) {
            setFormError("Name this permission group before saving.");
            return;
        }
        if (scopeKind !== "all" && !target) {
            setFormError(`${scopeLabel(scopeKind)} permissions need a target.`);
            return;
        }
        if (parsedRefs.length === 0) {
            setFormError("Add at least one capability reference, such as mcp:filesystem/*.");
            return;
        }
        setFormError(null);
        setSaveState("saving");
        const ok = await onCreate({
            name: nextName,
            description: description.trim() || undefined,
            tool_refs: parsedRefs,
            scope_kind: scopeKind,
            scope_ref: scopeKind === "all" ? undefined : target,
        });
        setSaveState(ok ? "saved" : "error");
        if (ok) {
            setName("");
            setDescription("");
            setScopeRef("");
            setToolRefsText("");
        }
    }
    const applyChoice = (refs: string[], recommendedScope: MCPToolSetScopeKind) => {
        const existingRefs = parseToolRefs(toolRefsText);
        const nextRefs = [...existingRefs];
        for (const ref of refs) {
            if (!nextRefs.includes(ref)) nextRefs.push(ref);
        }
        setToolRefsText(nextRefs.join("\n"));
        if (!scopeRef.trim() && scopeKind === "all" && recommendedScope !== "all") {
            setScopeKind(recommendedScope);
        }
    };
    return (
        <section className="rounded-xl border border-cortex-border bg-cortex-surface px-4 py-4">
            <div className="flex flex-col gap-3 md:flex-row md:items-start md:justify-between">
                <div className="flex items-start gap-3">
                    <div className="rounded-lg border border-cortex-primary/25 bg-cortex-primary/10 p-2">
                        <Layers3 className="h-4 w-4 text-cortex-primary" />
                    </div>
                    <div>
                        <p className="text-[10px] font-mono font-bold uppercase tracking-wider text-cortex-text-muted">
                            Capability permissions
                        </p>
                        <p className="mt-1 text-sm font-semibold text-cortex-text-main">
                            Choose where Soma can use connected tools
                        </p>
                        <p className="mt-1 text-xs leading-5 text-cortex-text-muted">
                            Start with safe workspace defaults, then narrow sensitive tools to one group or runtime host when a tighter boundary is needed.
                        </p>
                    </div>
                </div>
                <button
                    type="button"
                    onClick={onRefresh}
                    className="inline-flex items-center justify-center gap-2 rounded-lg border border-cortex-border bg-cortex-bg px-3 py-2 text-[10px] font-mono font-bold uppercase text-cortex-text-muted hover:text-cortex-text-main"
                >
                    <RefreshCw className="h-3.5 w-3.5" />
                    Refresh
                </button>
            </div>

            <div className="mt-4 grid gap-2 md:grid-cols-3">
                {scopes.map((scope) => (
                    <div key={scope.kind} className="rounded-lg border border-cortex-border bg-cortex-bg/60 px-3 py-3">
                        <div className="flex items-center justify-between gap-2">
                            <p className="text-xs font-semibold text-cortex-text-main">{scope.label}</p>
                            <span className="rounded-full border border-cortex-border bg-cortex-surface px-2 py-0.5 text-[10px] font-mono text-cortex-text-muted">
                                {counts[scope.kind]}
                            </span>
                        </div>
                        <p className="mt-1 text-[11px] leading-4 text-cortex-text-muted">{scope.help}</p>
                        <p className="mt-2 text-[10px] font-mono uppercase tracking-wider text-cortex-primary/80">{scope.example}</p>
                    </div>
                ))}
            </div>

            {error && <Notice tone="warning" message={error} />}
            {formError && <Notice tone="danger" message={formError} />}
            {saveState === "saved" && <Notice tone="success" message="Permission group saved and refreshed." />}
            {saveState === "error" && <Notice tone="danger" message="Permission group could not be saved." />}

            <div className="mt-4 grid gap-3 lg:grid-cols-[minmax(0,1fr)_minmax(280px,360px)]">
                <div className="min-h-[160px] rounded-lg border border-cortex-border bg-cortex-bg/50 p-3">
                    <p className="text-[10px] font-mono font-bold uppercase tracking-wider text-cortex-text-muted">
                        Current permission groups
                    </p>
                    {isLoading && toolSets.length === 0 ? (
                        <p className="mt-3 text-xs text-cortex-text-muted">Loading capability permissions...</p>
                    ) : toolSets.length === 0 ? (
                        <p className="mt-3 text-xs leading-5 text-cortex-text-muted">
                            No reusable permission groups are visible yet. Save an Everyone group first, then add group or host limits for sensitive tools.
                        </p>
                    ) : (
                        <div className="mt-3 max-h-72 overflow-y-auto pr-1">
                            {toolSets.map((toolSet) => (
                                <ToolSetRow key={toolSet.id} toolSet={toolSet} />
                            ))}
                        </div>
                    )}
                </div>

                <form onSubmit={handleSubmit} className="rounded-lg border border-cortex-border bg-cortex-bg/50 p-3">
                    <p className="text-[10px] font-mono font-bold uppercase tracking-wider text-cortex-text-muted">
                        Add permission group
                    </p>
                    <label className="mt-3 block text-[10px] font-mono uppercase tracking-wider text-cortex-text-muted">
                        Name
                        <input
                            value={name}
                            onChange={(event) => setName(event.target.value)}
                            className="mt-1 w-full rounded-lg border border-cortex-border bg-cortex-surface px-3 py-2 text-sm normal-case tracking-normal text-cortex-text-main outline-none focus:border-cortex-primary"
                            placeholder="Workspace file access"
                        />
                    </label>
                    <div className="mt-3 grid grid-cols-3 gap-1 rounded-lg border border-cortex-border bg-cortex-surface p-1">
                        {scopes.map((scope) => (
                            <button
                                key={scope.kind}
                                type="button"
                                onClick={() => setScopeKind(scope.kind)}
                                className={`rounded-md px-2 py-1.5 text-[10px] font-mono font-bold uppercase ${
                                    scopeKind === scope.kind
                                        ? "bg-cortex-primary/20 text-cortex-primary"
                                        : "text-cortex-text-muted hover:text-cortex-text-main"
                                }`}
                            >
                                {scope.label}
                            </button>
                        ))}
                    </div>
                    {scopeKind !== "all" && (
                        <label className="mt-3 block text-[10px] font-mono uppercase tracking-wider text-cortex-text-muted">
                            Target {scopeLabel(scopeKind)}
                            <input
                                value={scopeRef}
                                onChange={(event) => setScopeRef(event.target.value)}
                                className="mt-1 w-full rounded-lg border border-cortex-border bg-cortex-surface px-3 py-2 text-sm normal-case tracking-normal text-cortex-text-main outline-none focus:border-cortex-primary"
                                placeholder={scopeKind === "group" ? "marketing-outcome" : "local-media-host"}
                            />
                        </label>
                    )}
                    <label className="mt-3 block text-[10px] font-mono uppercase tracking-wider text-cortex-text-muted">
                        Capability refs
                        <textarea
                            value={toolRefsText}
                            onChange={(event) => setToolRefsText(event.target.value)}
                            className="mt-1 min-h-20 w-full rounded-lg border border-cortex-border bg-cortex-surface px-3 py-2 text-sm normal-case tracking-normal text-cortex-text-main outline-none focus:border-cortex-primary"
                            placeholder={"mcp:filesystem/*\ntoolset:research"}
                        />
                    </label>
                    <MCPToolSetCommonChoices onChoose={applyChoice} />
                    <label className="mt-3 block text-[10px] font-mono uppercase tracking-wider text-cortex-text-muted">
                        Description
                        <input
                            value={description}
                            onChange={(event) => setDescription(event.target.value)}
                            className="mt-1 w-full rounded-lg border border-cortex-border bg-cortex-surface px-3 py-2 text-sm normal-case tracking-normal text-cortex-text-main outline-none focus:border-cortex-primary"
                            placeholder="When Soma should be allowed to use these tools"
                        />
                    </label>
                    <button
                        type="submit"
                        disabled={saveState === "saving"}
                        className="mt-4 inline-flex w-full items-center justify-center gap-2 rounded-lg border border-cortex-primary/30 bg-cortex-primary/10 px-3 py-2 text-[10px] font-mono font-bold uppercase text-cortex-primary hover:bg-cortex-primary/20 disabled:opacity-50"
                    >
                        <Save className="h-3.5 w-3.5" />
                        {saveState === "saving" ? "Saving" : "Save permissions"}
                    </button>
                </form>
            </div>
        </section>
    );
}

export function MCPToolSetLayersStorePanel() {
    const toolSets = useCortexStore((state) => state.mcpToolSets);
    const isLoading = useCortexStore((state) => state.isFetchingMCPToolSets);
    const error = useCortexStore((state) => state.mcpToolSetsError);
    const onRefresh = useCortexStore((state) => state.fetchMCPToolSets);
    const onCreate = useCortexStore((state) => state.createMCPToolSet);

    return (
        <MCPToolSetLayersPanel
            toolSets={toolSets}
            isLoading={isLoading}
            error={error}
            onRefresh={onRefresh}
            onCreate={onCreate}
        />
    );
}

function ToolSetRow({ toolSet }: { toolSet: MCPToolSet }) {
    const scope = scopeLabel(toolSet.scope_kind);
    return (
        <div className="mb-2 rounded-lg border border-cortex-border bg-cortex-surface px-3 py-2">
            <div className="flex flex-wrap items-center justify-between gap-2">
                <p className="text-sm font-semibold text-cortex-text-main">{toolSet.name}</p>
                <span className="rounded border border-cortex-border bg-cortex-bg px-2 py-0.5 text-[10px] font-mono uppercase text-cortex-text-muted">
                    {scope}{toolSet.scope_ref ? `: ${toolSet.scope_ref}` : ""}
                </span>
            </div>
            {toolSet.description && <p className="mt-1 text-[11px] leading-4 text-cortex-text-muted">{toolSet.description}</p>}
            <p className="mt-2 flex items-start gap-1.5 text-[11px] leading-4 text-cortex-text-muted">
                <ShieldCheck className="mt-0.5 h-3 w-3 flex-none text-cortex-primary" />
                <span>{permissionSummary(toolSet)}</span>
            </p>
            <p className="mt-2 break-words text-[11px] font-mono leading-4 text-cortex-primary">
                {toolSet.tool_refs.length > 0 ? toolSet.tool_refs.join(", ") : "No tool refs"}
            </p>
        </div>
    );
}

function Notice({ message, tone }: { message: string; tone: "success" | "warning" | "danger" }) {
    const className = tone === "success"
        ? "border-cortex-success/25 bg-cortex-success/10 text-cortex-success"
        : tone === "warning"
        ? "border-cortex-warning/25 bg-cortex-warning/10 text-cortex-warning"
        : "border-cortex-danger/25 bg-cortex-danger/10 text-cortex-danger";
    return <p className={`mt-3 rounded-lg border px-3 py-2 text-[11px] font-mono ${className}`}>{message}</p>;
}

function parseToolRefs(value: string): string[] {
    return value.split(/[,\n]/).map((item) => item.trim()).filter(Boolean);
}

function scopeLabel(scope: string): string {
    if (scope === "all") return "Everyone";
    if (scope === "group") return "Group";
    if (scope === "host") return "Host";
    return scope;
}

function permissionSummary(toolSet: MCPToolSet): string {
    if (toolSet.scope_kind === "all") {
        return "Soma can use these capabilities as the workspace default when work is approved.";
    }
    if (toolSet.scope_kind === "group") {
        return "Soma can use these capabilities only inside the selected group or Outcome lane.";
    }
    if (toolSet.scope_kind === "host") {
        return "Soma can use these capabilities only on the selected runtime host.";
    }
    return "Soma can use these capabilities inside this configured boundary.";
}
