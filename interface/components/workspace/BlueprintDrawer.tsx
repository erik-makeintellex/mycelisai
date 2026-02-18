"use client";

import React, { useRef } from 'react';
import { X, FileJson, Upload, Download } from 'lucide-react';
import { useCortexStore } from '@/store/useCortexStore';
import type { MissionBlueprint } from '@/store/useCortexStore';

export default function BlueprintDrawer() {
    const isOpen = useCortexStore((s) => s.isBlueprintDrawerOpen);
    const toggle = useCortexStore((s) => s.toggleBlueprintDrawer);
    const savedBlueprints = useCortexStore((s) => s.savedBlueprints);
    const loadBlueprint = useCortexStore((s) => s.loadBlueprint);
    const saveBlueprint = useCortexStore((s) => s.saveBlueprint);
    const currentBlueprint = useCortexStore((s) => s.blueprint);
    const fileInputRef = useRef<HTMLInputElement>(null);

    if (!isOpen) return null;

    const handleImport = (e: React.ChangeEvent<HTMLInputElement>) => {
        const file = e.target.files?.[0];
        if (!file) return;

        const reader = new FileReader();
        reader.onload = (ev) => {
            try {
                const bp = JSON.parse(ev.target?.result as string) as MissionBlueprint;
                if (bp.mission_id && bp.teams) {
                    saveBlueprint(bp);
                }
            } catch {
                console.error('[BLUEPRINT] Invalid JSON file');
            }
        };
        reader.readAsText(file);
        e.target.value = '';
    };

    const handleExport = (bp: MissionBlueprint) => {
        const blob = new Blob([JSON.stringify(bp, null, 2)], { type: 'application/json' });
        const url = URL.createObjectURL(blob);
        const a = document.createElement('a');
        a.href = url;
        a.download = `${bp.mission_id}.json`;
        a.click();
        URL.revokeObjectURL(url);
    };

    return (
        <div className="absolute right-0 top-0 bottom-0 w-72 z-40 bg-cortex-surface border-l border-cortex-border shadow-2xl flex flex-col">
            {/* Header */}
            <div className="flex items-center justify-between px-3 py-2.5 border-b border-cortex-border">
                <div className="flex items-center gap-2">
                    <FileJson className="w-4 h-4 text-cortex-primary" />
                    <span className="text-xs font-mono font-bold text-cortex-text-main uppercase">
                        Blueprint Library
                    </span>
                </div>
                <button onClick={toggle} className="p-1 rounded hover:bg-cortex-border text-cortex-text-muted hover:text-cortex-text-main transition-colors">
                    <X className="w-4 h-4" />
                </button>
            </div>

            {/* Actions */}
            <div className="px-3 py-2 border-b border-cortex-border flex gap-2">
                <button
                    onClick={() => fileInputRef.current?.click()}
                    className="flex-1 flex items-center justify-center gap-1.5 px-2 py-1.5 rounded text-[10px] font-mono bg-cortex-primary/10 border border-cortex-primary/30 text-cortex-primary hover:bg-cortex-primary/20 transition-colors"
                >
                    <Upload className="w-3 h-3" />
                    IMPORT JSON
                </button>
                {currentBlueprint && (
                    <button
                        onClick={() => saveBlueprint(currentBlueprint)}
                        className="flex-1 flex items-center justify-center gap-1.5 px-2 py-1.5 rounded text-[10px] font-mono bg-cortex-success/10 border border-cortex-success/30 text-cortex-success hover:bg-cortex-success/20 transition-colors"
                    >
                        SAVE CURRENT
                    </button>
                )}
            </div>
            <input ref={fileInputRef} type="file" accept=".json" className="hidden" onChange={handleImport} />

            {/* Blueprint list */}
            <div className="flex-1 overflow-y-auto p-2 space-y-2">
                {savedBlueprints.length === 0 ? (
                    <div className="flex flex-col items-center justify-center h-32 text-cortex-text-muted">
                        <FileJson className="w-8 h-8 mb-2 opacity-30" />
                        <p className="text-[10px] font-mono">No saved blueprints</p>
                        <p className="text-[9px] font-mono mt-0.5">Import JSON or save a draft</p>
                    </div>
                ) : (
                    savedBlueprints.map((bp) => {
                        const agentCount = bp.teams.reduce((sum, t) => sum + t.agents.length, 0);
                        return (
                            <div
                                key={bp.mission_id}
                                className="p-2.5 rounded-lg border border-cortex-border bg-cortex-bg hover:border-cortex-text-muted transition-colors cursor-pointer group"
                                onClick={() => loadBlueprint(bp)}
                            >
                                <div className="flex items-start justify-between">
                                    <div className="flex-1 min-w-0">
                                        <p className="text-[11px] font-mono font-bold text-cortex-text-main truncate">
                                            {bp.mission_id}
                                        </p>
                                        <p className="text-[9px] text-cortex-text-muted mt-0.5 line-clamp-2">
                                            {bp.intent}
                                        </p>
                                        <div className="flex gap-2 mt-1.5">
                                            <span className="text-[8px] font-mono text-cortex-info">
                                                {bp.teams.length} teams
                                            </span>
                                            <span className="text-[8px] font-mono text-cortex-success">
                                                {agentCount} agents
                                            </span>
                                        </div>
                                    </div>
                                    <button
                                        onClick={(e) => { e.stopPropagation(); handleExport(bp); }}
                                        className="p-1 rounded opacity-0 group-hover:opacity-100 hover:bg-cortex-border text-cortex-text-muted transition-all"
                                    >
                                        <Download className="w-3 h-3" />
                                    </button>
                                </div>
                            </div>
                        );
                    })
                )}
            </div>
        </div>
    );
}
