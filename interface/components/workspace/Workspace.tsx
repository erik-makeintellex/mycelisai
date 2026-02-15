"use client";

import React from 'react';
import ArchitectChat from './ArchitectChat';
import BlueprintDrawer from './BlueprintDrawer';
import CircuitBoard from './CircuitBoard';
import SquadRoom from './SquadRoom';
import DeliverablesTray from './DeliverablesTray';
import GovernanceModal from '../shell/GovernanceModal';
import { NatsWaterfall } from '../stream/NatsWaterfall';
import { useCortexStore } from '@/store/useCortexStore';

export default function Workspace() {
    const activeSquadRoomId = useCortexStore((s) => s.activeSquadRoomId);

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
                </div>
            </div>

            {/* Bottom: Spectrum Console */}
            <NatsWaterfall />

            {/* Governance Modal — z-50 overlay */}
            <GovernanceModal />
        </div>
    );
}
