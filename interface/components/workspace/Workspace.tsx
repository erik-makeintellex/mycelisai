"use client";

import React from 'react';
import { Wrench } from 'lucide-react';
import ArchitectChat from './ArchitectChat';
import BlueprintDrawer from './BlueprintDrawer';
import ToolsPalette from './ToolsPalette';
import CircuitBoard from './CircuitBoard';
import SquadRoom from './SquadRoom';
import DeliverablesTray from './DeliverablesTray';
import GovernanceModal from '../shell/GovernanceModal';
import { NatsWaterfall } from '../stream/NatsWaterfall';
import { useCortexStore } from '@/store/useCortexStore';

export default function Workspace() {
    const activeSquadRoomId = useCortexStore((s) => s.activeSquadRoomId);
    const toggleToolsPalette = useCortexStore((s) => s.toggleToolsPalette);
    const isToolsPaletteOpen = useCortexStore((s) => s.isToolsPaletteOpen);

    return (
        <div className="h-full flex flex-col bg-cortex-bg">
            {/* Top: Chat + Canvas */}
            <div className="flex-1 grid grid-cols-[360px_1fr] gap-6 p-6 min-h-0">
                {/* Left: Negotiation Chat + Blueprint Drawer */}
                <div className="bg-cortex-surface rounded-xl shadow-lg overflow-hidden relative">
                    <ArchitectChat />
                    <BlueprintDrawer />
                </div>

                {/* Right: Circuit Board or Squad Room (fractal drill-down) */}
                <div className="bg-cortex-surface rounded-xl shadow-lg overflow-hidden relative">
                    {activeSquadRoomId ? (
                        <SquadRoom teamId={activeSquadRoomId} />
                    ) : (
                        <CircuitBoard />
                    )}

                    {/* Deliverables Tray — docked above canvas */}
                    <DeliverablesTray />

                    {/* Tools Palette — slides from left edge of canvas */}
                    <ToolsPalette />

                    {/* Tools toggle button */}
                    {!isToolsPaletteOpen && (
                        <button
                            onClick={toggleToolsPalette}
                            className="absolute top-2 left-2 z-30 p-1.5 rounded-lg bg-cortex-surface/80 border border-cortex-border hover:bg-cortex-primary/10 hover:border-cortex-primary/30 text-cortex-text-muted hover:text-cortex-primary transition-all backdrop-blur-sm"
                            title="Tool Registry"
                        >
                            <Wrench className="w-4 h-4" />
                        </button>
                    )}
                </div>
            </div>

            {/* Bottom: Spectrum Console */}
            <NatsWaterfall />

            {/* Governance Modal — z-50 overlay */}
            <GovernanceModal />
        </div>
    );
}
