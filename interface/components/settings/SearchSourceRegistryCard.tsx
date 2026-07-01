"use client";

import React from "react";
import { Plus, Search } from "lucide-react";
import type { SearchCapabilitySource } from "@/store/useCortexStore";
import { SearchSourceList } from "./SearchSourceList";
import { SearchSourceAddForm, emptySourceLabel, sourceToDraft, type SearchSourceDraft } from "./SearchSourceForm";

export function SearchSourceRegistryCard({
    sources,
    isLoading,
    addSupported,
    error,
    addNotice,
    isAdding,
    onAddSearchSource,
    onUpdateSearchSource,
    onDeleteSearchSource,
}: {
    sources: SearchCapabilitySource[];
    isLoading: boolean;
    addSupported: boolean;
    error: string | null;
    addNotice: string | null;
    isAdding: boolean;
    onAddSearchSource: (input: SearchSourceDraft) => Promise<boolean>;
    onUpdateSearchSource: (sourceId: string, input: SearchSourceDraft) => Promise<boolean>;
    onDeleteSearchSource: (sourceId: string, sourceName: string) => Promise<boolean>;
}) {
    const [isFormOpen, setIsFormOpen] = React.useState(false);
    const [editingSource, setEditingSource] = React.useState<SearchCapabilitySource | null>(null);

    return (
        <section className="rounded-xl border border-cortex-border bg-cortex-surface px-4 py-4">
            <div className="flex flex-col gap-3 sm:flex-row sm:items-start sm:justify-between">
                <div className="flex items-start gap-3">
                    <div className="mt-0.5 rounded-lg border border-cortex-info/25 bg-cortex-info/10 p-2">
                        <Search className="h-4 w-4 text-cortex-info" />
                    </div>
                    <div>
                        <p className="text-[10px] font-mono font-bold uppercase tracking-wider text-cortex-text-muted">
                            Search sources
                        </p>
                        <p className="mt-1 text-sm font-semibold text-cortex-text-main">
                            {sources.length > 0 ? `${sources.length} configured source${sources.length === 1 ? "" : "s"}` : emptySourceLabel(addSupported)}
                        </p>
                        <p className="mt-1 text-xs leading-5 text-cortex-text-muted">
                            Where Soma may search for governed research, with scope and auth boundaries visible.
                        </p>
                    </div>
                </div>
                {addSupported && (
                    <button
                        type="button"
                        onClick={() => setIsFormOpen((open) => !open)}
                        className="inline-flex shrink-0 items-center justify-center gap-2 rounded-lg border border-cortex-primary/30 bg-cortex-primary/10 px-3 py-2 text-xs font-semibold text-cortex-primary transition hover:bg-cortex-primary/20"
                    >
                        <Plus className="h-3.5 w-3.5" />
                        Add search source
                    </button>
                )}
            </div>

            {isLoading && sources.length === 0 && (
                <p className="mt-3 text-xs text-cortex-text-muted">Checking configured search sources...</p>
            )}
            {error && (
                <div className="mt-3 rounded-lg border border-cortex-warning/25 bg-cortex-warning/10 px-3 py-2">
                    <p className="text-[10px] font-mono text-cortex-warning">{error}</p>
                </div>
            )}
            {addNotice && (
                <div className="mt-3 rounded-lg border border-cortex-success/25 bg-cortex-success/10 px-3 py-2">
                    <p className="text-[10px] font-mono text-cortex-success">{addNotice}</p>
                </div>
            )}
            {(isFormOpen || editingSource) && (
                <SearchSourceAddForm
                    key={editingSource?.id ?? "new-search-source"}
                    isAdding={isAdding}
                    initialDraft={editingSource ? sourceToDraft(editingSource) : undefined}
                    submitLabel={editingSource ? "Update search source" : "Add search source"}
                    onCancel={() => {
                        setIsFormOpen(false);
                        setEditingSource(null);
                    }}
                    onSubmit={async (input) => {
                        const ok = editingSource
                            ? await onUpdateSearchSource(editingSource.id, input)
                            : await onAddSearchSource(input);
                        if (ok) {
                            setIsFormOpen(false);
                            setEditingSource(null);
                        }
                        return ok;
                    }}
                />
            )}
            <SearchSourceList
                sources={sources}
                title="Configured source details"
                onDeleteSource={onDeleteSearchSource}
                onEditSource={(source) => {
                    setIsFormOpen(false);
                    setEditingSource(source);
                }}
            />
        </section>
    );
}
