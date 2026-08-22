"use client";

import React, { useCallback, useEffect, useState } from "react";
import { BookOpen, MessageSquareText, Search } from "lucide-react";
import { useCortexStore, type CatalogueAgent } from "@/store/useCortexStore";
import { customizeWorkerProfilePrompt, newWorkerProfilePrompt, requestSomaPromptHandoff } from "@/components/soma/somaPromptHandoff";
import WorkerProfileAuthoringPanel from "@/components/resources/WorkerProfileAuthoringPanel";
import AgentCard from "./AgentCard";
import AgentEditorDrawer from "./AgentEditorDrawer";

export default function CataloguePage() {
  const catalogueAgents = useCortexStore((state) => state.catalogueAgents);
  const isFetchingCatalogue = useCortexStore((state) => state.isFetchingCatalogue);
  const selectedCatalogueAgent = useCortexStore((state) => state.selectedCatalogueAgent);
  const fetchCatalogue = useCortexStore((state) => state.fetchCatalogue);
  const selectCatalogueAgent = useCortexStore((state) => state.selectCatalogueAgent);
  const [query, setQuery] = useState("");

  useEffect(() => { void fetchCatalogue(); }, [fetchCatalogue]);

  const isBuiltIn = (agent: CatalogueAgent) => agent.source === "built_in" || Boolean(agent.locked);
  const builtInAgents = catalogueAgents.filter(isBuiltIn);
  const filteredAgents = builtInAgents.filter((agent) => {
    const searchText = `${agent.name} ${agent.role} ${agent.description ?? ""}`.toLowerCase();
    return searchText.includes(query.trim().toLowerCase());
  });

  const closeEditor = useCallback(() => {
    selectCatalogueAgent(null);
  }, [selectCatalogueAgent]);

  const emptyMessage = query ? "No matching profiles." : "Default profiles are not loaded.";

  return (
    <div className="relative flex h-full flex-col bg-cortex-bg">
      <header className="flex flex-wrap items-center justify-between gap-3 border-b border-cortex-border bg-cortex-surface/50 px-5 py-4">
        <div className="flex min-w-0 items-center gap-3">
          <BookOpen className="h-5 w-5 text-cortex-success" />
          <div>
            <h2 className="text-base font-semibold text-cortex-text-main">Worker profiles</h2>
            <p className="text-xs text-cortex-text-muted">Inspect ready-made teammates, or ask Soma to create an activated custom profile.</p>
          </div>
          {isFetchingCatalogue && <span className="text-xs text-cortex-text-muted animate-pulse">Loading</span>}
        </div>
        <button
          type="button"
          onClick={() => requestSomaPromptHandoff(newWorkerProfilePrompt())}
          className="flex items-center gap-1.5 rounded-md border border-cortex-success/40 bg-cortex-success/10 px-3 py-2 text-xs font-semibold text-cortex-success hover:bg-cortex-success/20"
        >
          <MessageSquareText className="h-4 w-4" /> Create with Soma
        </button>
      </header>

      <div className="flex flex-wrap items-center justify-between gap-3 border-b border-cortex-border px-5 py-3">
        <p className="text-xs font-semibold text-cortex-text-muted">{builtInAgents.length} ready-made profiles</p>
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
        <WorkerProfileAuthoringPanel />
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
                onSelect={selectCatalogueAgent}
              />
            ))}
          </div>
        )}
      </main>

      {selectedCatalogueAgent && (
        <AgentEditorDrawer
          agent={selectedCatalogueAgent}
          onClose={closeEditor}
          onSave={() => undefined}
          onCustomize={(agent) => requestSomaPromptHandoff(customizeWorkerProfilePrompt(agent.name))}
        />
      )}
    </div>
  );
}
