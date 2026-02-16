"use client";

import React, { useEffect, useState, useCallback } from 'react';
import { BookOpen, Plus, Filter } from 'lucide-react';
import { useCortexStore, type CatalogueAgent } from '@/store/useCortexStore';
import AgentCard from './AgentCard';
import AgentEditorDrawer from './AgentEditorDrawer';

const ROLE_FILTERS = ['all', 'cognitive', 'sensory', 'actuation', 'ledger'] as const;

export default function CataloguePage() {
    const catalogueAgents = useCortexStore((s) => s.catalogueAgents);
    const isFetchingCatalogue = useCortexStore((s) => s.isFetchingCatalogue);
    const selectedCatalogueAgent = useCortexStore((s) => s.selectedCatalogueAgent);
    const fetchCatalogue = useCortexStore((s) => s.fetchCatalogue);
    const createCatalogueAgent = useCortexStore((s) => s.createCatalogueAgent);
    const updateCatalogueAgent = useCortexStore((s) => s.updateCatalogueAgent);
    const deleteCatalogueAgent = useCortexStore((s) => s.deleteCatalogueAgent);
    const selectCatalogueAgent = useCortexStore((s) => s.selectCatalogueAgent);

    const [isCreating, setIsCreating] = useState(false);
    const [roleFilter, setRoleFilter] = useState<string>('all');

    useEffect(() => {
        fetchCatalogue();
    }, [fetchCatalogue]);

    // Filtered agents
    const filteredAgents =
        roleFilter === 'all'
            ? catalogueAgents
            : catalogueAgents.filter((a) => a.role === roleFilter);

    // Drawer visibility
    const drawerOpen = isCreating || selectedCatalogueAgent !== null;

    // Drawer agent (null for create, agent for edit)
    const drawerAgent = isCreating ? null : selectedCatalogueAgent;

    const handleNewAgent = useCallback(() => {
        selectCatalogueAgent(null);
        setIsCreating(true);
    }, [selectCatalogueAgent]);

    const handleSelectAgent = useCallback(
        (agent: CatalogueAgent) => {
            setIsCreating(false);
            selectCatalogueAgent(agent);
        },
        [selectCatalogueAgent],
    );

    const handleDeleteAgent = useCallback(
        (id: string) => {
            deleteCatalogueAgent(id);
        },
        [deleteCatalogueAgent],
    );

    const handleDrawerClose = useCallback(() => {
        setIsCreating(false);
        selectCatalogueAgent(null);
    }, [selectCatalogueAgent]);

    const handleDrawerSave = useCallback(
        (data: Partial<CatalogueAgent>) => {
            if (selectedCatalogueAgent && !isCreating) {
                updateCatalogueAgent(selectedCatalogueAgent.id, data);
            } else {
                createCatalogueAgent(data);
            }
            setIsCreating(false);
            selectCatalogueAgent(null);
        },
        [selectedCatalogueAgent, isCreating, updateCatalogueAgent, createCatalogueAgent, selectCatalogueAgent],
    );

    return (
        <div className="flex flex-col h-full relative bg-cortex-bg">
            {/* Header bar */}
            <div className="h-12 border-b border-cortex-border bg-cortex-surface/50 backdrop-blur-sm flex items-center justify-between px-4 flex-shrink-0">
                <div className="flex items-center gap-2">
                    <BookOpen className="w-4 h-4 text-cortex-success" />
                    <span className="text-xs font-mono font-bold text-cortex-text-main uppercase tracking-wide">
                        Agent Catalogue
                    </span>
                    {isFetchingCatalogue && (
                        <span className="text-[9px] font-mono text-cortex-text-muted animate-pulse">
                            loading...
                        </span>
                    )}
                </div>

                <div className="flex items-center gap-2">
                    {/* Role filter dropdown */}
                    <div className="relative flex items-center gap-1">
                        <Filter className="w-3 h-3 text-cortex-text-muted" />
                        <select
                            value={roleFilter}
                            onChange={(e) => setRoleFilter(e.target.value)}
                            className="bg-cortex-bg border border-cortex-border rounded px-2 py-1 text-[10px] font-mono text-cortex-text-main focus:outline-none focus:border-cortex-primary transition-colors appearance-none pr-5"
                        >
                            {ROLE_FILTERS.map((f) => (
                                <option key={f} value={f}>
                                    {f === 'all' ? 'All Roles' : f}
                                </option>
                            ))}
                        </select>
                    </div>

                    {/* New Agent button */}
                    <button
                        onClick={handleNewAgent}
                        className="flex items-center gap-1.5 px-3 py-1.5 rounded text-[10px] font-mono font-bold uppercase bg-cortex-success/15 border border-cortex-success/30 text-cortex-success hover:bg-cortex-success/25 transition-colors"
                    >
                        <Plus className="w-3.5 h-3.5" />
                        New Agent
                    </button>
                </div>
            </div>

            {/* Body: Agent grid or empty state */}
            <div className="flex-1 overflow-y-auto">
                {filteredAgents.length === 0 && !isFetchingCatalogue ? (
                    /* Empty state */
                    <div className="flex flex-col items-center justify-center h-full text-cortex-text-muted">
                        <BookOpen className="w-16 h-16 mb-4 opacity-20" />
                        <p className="text-sm font-mono">
                            No agents configured.
                        </p>
                        <p className="text-xs font-mono mt-1 text-cortex-text-muted/60">
                            Create your first agent.
                        </p>
                    </div>
                ) : (
                    <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4 gap-4 p-6">
                        {filteredAgents.map((agent) => (
                            <AgentCard
                                key={agent.id}
                                agent={agent}
                                onSelect={handleSelectAgent}
                                onDelete={handleDeleteAgent}
                            />
                        ))}
                    </div>
                )}
            </div>

            {/* Editor Drawer */}
            {drawerOpen && (
                <AgentEditorDrawer
                    agent={drawerAgent}
                    onClose={handleDrawerClose}
                    onSave={handleDrawerSave}
                />
            )}
        </div>
    );
}
