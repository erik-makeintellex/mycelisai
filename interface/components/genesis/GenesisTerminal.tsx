"use client";
import React, { useState } from 'react';
import { Network, Search, Compass, MessageSquare, Terminal } from 'lucide-react';

export default function GenesisTerminal() {
    const [state, setState] = useState<'pathfinder' | 'negotiation'>('pathfinder');

    return (
        <div className="flex-1 h-full flex flex-col p-6 bg-white">
            {state === 'pathfinder' ? (
                <PathfinderState onSelect={() => setState('negotiation')} />
            ) : (
                <NegotiationState />
            )}
        </div>
    );
}

// State 1: Role Selection
function PathfinderState({ onSelect }: { onSelect: () => void }) {
    return (
        <div className="flex-1 flex flex-col items-center justify-center animate-in fade-in duration-700">
            <div className="mb-8 text-center">
                <h1 className="text-3xl font-bold text-zinc-900 mb-2 tracking-tight">System Initialization</h1>
                <p className="text-zinc-500 max-w-md mx-auto">
                    The Cortex is untethered. Select a primary directive to begin the bootstrap sequence.
                </p>
            </div>

            <div className="grid grid-cols-1 md:grid-cols-3 gap-6 max-w-4xl w-full">
                <RoleCard
                    icon={Network}
                    title="Architect"
                    desc="Design system topology and infrastructure blueprints."
                    onClick={onSelect}
                />
                <RoleCard
                    icon={Terminal}
                    title="Commander"
                    desc="Execute direct operational control over swarm nodes."
                    onClick={onSelect}
                />
                <RoleCard
                    icon={Compass}
                    title="Explorer"
                    desc="Autonomous discovery and resource mapping."
                    onClick={onSelect}
                />
            </div>
        </div>
    );
}

function RoleCard({ icon: Icon, title, desc, onClick }: any) {
    return (
        <button
            onClick={onClick}
            className="group p-6 rounded-xl border border-zinc-200 bg-white hover:border-indigo-300 hover:shadow-lg hover:shadow-indigo-50/50 transition-all text-left"
        >
            <div className="w-12 h-12 bg-zinc-50 rounded-lg flex items-center justify-center mb-4 group-hover:bg-indigo-50 group-hover:text-indigo-600 transition-colors">
                <Icon className="w-6 h-6 text-zinc-400 group-hover:text-indigo-600" />
            </div>
            <h3 className="tex-lg font-bold text-zinc-900 mb-2">{title}</h3>
            <p className="text-sm text-zinc-500 leading-relaxed group-hover:text-zinc-700">{desc}</p>
        </button>
    );
}

// State 2: Negotiation (Split View)
function NegotiationState() {
    return (
        <div className="flex-1 flex gap-6 animate-in slide-in-from-bottom-8 duration-500">
            {/* Left: Chat Interface */}
            <div className="flex-1 flex flex-col border border-zinc-200 rounded-xl overflow-hidden shadow-sm">
                <div className="bg-zinc-50 p-4 border-b border-zinc-200 flex justify-between items-center">
                    <span className="font-semibold text-zinc-700 flex items-center">
                        <MessageSquare className="w-4 h-4 mr-2" />
                        Consular Link
                    </span>
                    <span className="text-[10px] uppercase font-bold text-emerald-600 bg-emerald-50 px-2 py-1 rounded">Online</span>
                </div>
                <div className="flex-1 bg-white p-4">
                    <div className="p-3 bg-indigo-50 rounded-lg rounded-tl-none max-w-[80%] mb-4">
                        <p className="text-sm text-indigo-900">
                            I am ready to receive your mission parameters. Describe the desired end-state.
                        </p>
                    </div>
                </div>
                <div className="p-4 border-t border-zinc-200 bg-zinc-50">
                    <input
                        type="text"
                        placeholder="Type your directive..."
                        className="w-full px-4 py-2 rounded-lg border border-zinc-300 focus:outline-none focus:ring-2 focus:ring-indigo-500"
                    />
                </div>
            </div>

            {/* Right: Graph Placeholder */}
            <div className="hidden lg:flex w-1/3 flex-col border border-zinc-200 rounded-xl overflow-hidden shadow-sm bg-zinc-900 relative">
                <div className="absolute inset-0 flex items-center justify-center opacity-20">
                    <Network className="w-32 h-32 text-indigo-500" />
                </div>
                <div className="absolute bottom-4 left-4">
                    <p className="text-zinc-400 text-xs font-mono">Topology Visualization</p>
                    <p className="text-zinc-500 text-[10px] font-mono">Waiting for data...</p>
                </div>
            </div>
        </div>
    );
}
