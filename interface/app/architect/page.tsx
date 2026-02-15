"use client";

import React from 'react';
import TeamBuilder from '../../components/swarm/TeamBuilder';

export default function ArchitectPage() {
    return (
        <div className="h-full flex flex-col bg-zinc-950 text-zinc-100">
            <header className="p-6 border-b border-zinc-800 flex justify-between items-center">
                <div>
                    <h1 className="text-2xl font-bold bg-gradient-to-r from-indigo-400 to-cyan-400 bg-clip-text text-transparent">
                        Swarm Architect
                    </h1>
                    <p className="text-zinc-500 text-sm mt-1">
                        Design and configure Soma's Neural Clusters.
                    </p>
                </div>
            </header>

            <main className="flex-1 overflow-hidden">
                <TeamBuilder />
            </main>
        </div>
    );
}
