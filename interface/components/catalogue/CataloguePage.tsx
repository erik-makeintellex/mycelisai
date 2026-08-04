"use client";

import React, { useCallback, useEffect, useState } from "react";
import { BookOpen, Plus, Search } from "lucide-react";
import { useCortexStore, type CatalogueAgent } from "@/store/useCortexStore";
import AgentCard from "./AgentCard";
import AgentEditorDrawer from "./AgentEditorDrawer";

type LibraryView = "built_in" | "user";

export default function CataloguePage() {
  const catalogueAgents = useCortexStore((state) => state.catalogueAgents);
  const isFetchingCatalogue = useCortexStore((state) => state.isFetchingCatalogue);
  const selectedCatalogueAgent = useCortexStore((state) => state.selectedCatalogueAgent);
  const fetchCatalogue = useCortexStore((state) => state.fetchCatalogue);
  const createCatalogueAgent = useCortexStore((state) => state.createCatalogueAgent);
  const updateCatalogueAgent = useCortexStore((state) => state.updateCatalogueAgent);
  const deleteCatalogueAgent = useCortexStore((state) => state.deleteCatalogueAgent);
  const selectCatalogueAgent = useCortexStore((state) => state.selectCatalogueAgent);
  const [isCreating, setIsCreating] = useState(false);
  const [libraryView, setLibraryView] = useState<LibraryView>("built_in");
  const [query, setQuery] = useState("");

  useEffect(() => { void fetchCatalogue(); }, [fetchCatalogue]);

  const isBuiltIn = (agent: CatalogueAgent) => agent.source === "built_in" || Boolean(agent.locked);
  const builtInCount = catalogueAgents.filter(isBuiltIn).length;
  const userCount = catalogueAgents.length - builtInCount;
  const filteredAgents = catalogueAgents.filter((agent) => {
    const source: LibraryView = isBuiltIn(agent) ? "built_in" : "user";
    const searchText = `${agent.name} ${agent.role} ${agent.description ?? ""}`.toLowerCase();
    return source === libraryView && searchText.includes(query.trim().toLowerCase());
  });

  const closeEditor = useCallback(() => {
    setIsCreating(false);
    selectCatalogueAgent(null);
  }, [selectCatalogueAgent]);

  const saveProfile = useCallback((data: Partial<CatalogueAgent>) => {
    if (selectedCatalogueAgent && !isCreating) {
      void updateCatalogueAgent(selectedCatalogueAgent.id, data);
    } else {
      void createCatalogueAgent(data);
    }
    closeEditor();
  }, [selectedCatalogueAgent, isCreating, updateCatalogueAgent, createCatalogueAgent, closeEditor]);

  const duplicateProfile = useCallback((agent: CatalogueAgent) => {
    const suffix = globalThis.crypto?.randomUUID?.().replaceAll("-", "").slice(0, 5) ?? Date.now().toString(16).slice(-5);
    void createCatalogueAgent({
      ...agent,
      id: undefined,
      profile_key: undefined,
      name: `${agent.name} custom-${suffix}`,
      source: "user",
      locked: false,
    });
    closeEditor();
    setLibraryView("user");
  }, [createCatalogueAgent, closeEditor]);

  const emptyMessage = query
    ? "No matching profiles."
    : libraryView === "user" ? "No custom profiles yet." : "Default profiles are not loaded.";

  return (
    <div className="relative flex h-full flex-col bg-cortex-bg">
      <header className="flex flex-wrap items-center justify-between gap-3 border-b border-cortex-border bg-cortex-surface/50 px-5 py-4">
        <div className="flex min-w-0 items-center gap-3">
          <BookOpen className="h-5 w-5 text-cortex-success" />
          <div>
            <h2 className="text-base font-semibold text-cortex-text-main">Worker profiles</h2>
            <p className="text-xs text-cortex-text-muted">Reusable teammates Soma can assign to governed work.</p>
          </div>
          {isFetchingCatalogue && <span className="text-xs text-cortex-text-muted animate-pulse">Loading</span>}
        </div>
        <button
          type="button"
          onClick={() => { selectCatalogueAgent(null); setIsCreating(true); setLibraryView("user"); }}
          className="flex items-center gap-1.5 rounded-md border border-cortex-success/40 bg-cortex-success/10 px-3 py-2 text-xs font-semibold text-cortex-success hover:bg-cortex-success/20"
        >
          <Plus className="h-4 w-4" /> New profile
        </button>
      </header>

      <div className="flex flex-wrap items-center justify-between gap-3 border-b border-cortex-border px-5 py-3">
        <div className="inline-flex rounded-md border border-cortex-border bg-cortex-surface p-1" role="tablist" aria-label="Profile library">
          {([['built_in', 'Ready-made', builtInCount], ['user', 'My profiles', userCount]] as const).map(([value, label, count]) => (
            <button
              key={value}
              type="button"
              role="tab"
              aria-selected={libraryView === value}
              onClick={() => setLibraryView(value)}
              className={`rounded px-3 py-1.5 text-xs font-semibold ${libraryView === value ? "bg-cortex-primary/15 text-cortex-primary" : "text-cortex-text-muted hover:text-cortex-text-main"}`}
            >
              {label} <span className="ml-1 opacity-70">{count}</span>
            </button>
          ))}
        </div>
        <label className="flex min-w-56 items-center gap-2 rounded-md border border-cortex-border bg-cortex-surface px-3 py-2">
          <Search className="h-4 w-4 text-cortex-text-muted" />
          <span className="sr-only">Search profiles</span>
          <input
            value={query}
            onChange={(event) => setQuery(event.target.value)}
            placeholder="Search profiles"
            className="w-full bg-transparent text-xs text-cortex-text-main outline-none placeholder:text-cortex-text-muted"
          />
        </label>
      </div>

      <main className="flex-1 overflow-y-auto">
        {filteredAgents.length === 0 && !isFetchingCatalogue ? (
          <div className="flex h-full flex-col items-center justify-center text-cortex-text-muted">
            <BookOpen className="mb-3 h-10 w-10 opacity-20" />
            <p className="text-sm">{emptyMessage}</p>
          </div>
        ) : (
          <div className="grid grid-cols-1 gap-4 p-5 md:grid-cols-2 xl:grid-cols-3">
            {filteredAgents.map((agent) => (
              <AgentCard
                key={agent.id}
                agent={agent}
                onSelect={(selected) => { setIsCreating(false); selectCatalogueAgent(selected); }}
                onDelete={(id) => void deleteCatalogueAgent(id)}
              />
            ))}
          </div>
        )}
      </main>

      {(isCreating || selectedCatalogueAgent) && (
        <AgentEditorDrawer
          agent={isCreating ? null : selectedCatalogueAgent}
          onClose={closeEditor}
          onSave={saveProfile}
          onDuplicate={duplicateProfile}
        />
      )}
    </div>
  );
}
