"use client";

import React, { useState } from 'react';
import PriorityStream from './PriorityStream';
import ManifestationPanel from './ManifestationPanel';
import TeamExplorer from './TeamExplorer';

type Tab = 'priority' | 'teams' | 'manifest';

const TABS: { key: Tab; label: string }[] = [
    { key: 'priority', label: 'PRIORITY' },
    { key: 'teams', label: 'TEAMS' },
    { key: 'manifest', label: 'MANIFEST' },
];

export default function CenterTabs() {
    const [active, setActive] = useState<Tab>('priority');

    return (
        <div className="h-full flex flex-col min-w-0" data-testid="center-tabs">
            {/* Tab bar */}
            <div className="h-8 flex items-stretch border-b border-cortex-border flex-shrink-0">
                {TABS.map((tab) => (
                    <button
                        key={tab.key}
                        onClick={() => setActive(tab.key)}
                        className={`px-4 text-[9px] font-mono font-bold uppercase tracking-widest transition-colors ${
                            active === tab.key
                                ? 'text-cortex-primary border-b-2 border-cortex-primary'
                                : 'text-cortex-text-muted hover:text-cortex-text-main'
                        }`}
                        data-testid={`tab-${tab.key}`}
                    >
                        {tab.label}
                    </button>
                ))}
            </div>

            {/* Tab content */}
            <div className="flex-1 min-h-0 overflow-hidden">
                {active === 'priority' && <PriorityStream />}
                {active === 'teams' && <TeamExplorer />}
                {active === 'manifest' && <ManifestationPanel />}
            </div>
        </div>
    );
}
