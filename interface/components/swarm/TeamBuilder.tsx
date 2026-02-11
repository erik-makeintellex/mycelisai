"use client";

import React, { useState } from 'react';
import { Users, Cpu, Activity, Database, Save } from 'lucide-react';

type TeamType = 'action' | 'expression';

export default function TeamBuilder() {
    const [teamName, setTeamName] = useState('');
    const [teamType, setTeamType] = useState<TeamType>('action');

    return (
        <div className="h-full grid grid-cols-12 gap-0">
            {/* Left Panel: Configuration */}
            <div className="col-span-3 border-r border-zinc-800 p-6 bg-zinc-900/50">
                <h2 className="text-lg font-semibold mb-6 flex items-center gap-2">
                    <Cpu className="w-5 h-5 text-indigo-400" />
                    Core Configuration
                </h2>

                <div className="space-y-6">
                    <div>
                        <label className="block text-xs font-medium text-zinc-400 uppercase tracking-wider mb-2">Team Name</label>
                        <input
                            type="text"
                            className="w-full bg-zinc-950 border border-zinc-800 rounded-md p-2 text-sm focus:ring-1 focus:ring-indigo-500 outline-none"
                            placeholder="e.g. Research Core"
                            value={teamName}
                            onChange={(e) => setTeamName(e.target.value)}
                        />
                    </div>

                    <div>
                        <label className="block text-xs font-medium text-zinc-400 uppercase tracking-wider mb-2">Cluster Type</label>
                        <div className="grid grid-cols-2 gap-2">
                            <button
                                onClick={() => setTeamType('action')}
                                className={`p-3 rounded-md border text-sm flex flex-col items-center gap-2 transition-colors ${teamType === 'action'
                                        ? 'bg-indigo-900/30 border-indigo-500 text-indigo-200'
                                        : 'bg-zinc-950 border-zinc-800 text-zinc-500 hover:border-zinc-700'
                                    }`}
                            >
                                <Cpu className="w-5 h-5" />
                                <span>Action</span>
                            </button>
                            <button
                                onClick={() => setTeamType('expression')}
                                className={`p-3 rounded-md border text-sm flex flex-col items-center gap-2 transition-colors ${teamType === 'expression'
                                        ? 'bg-emerald-900/30 border-emerald-500 text-emerald-200'
                                        : 'bg-zinc-950 border-zinc-800 text-zinc-500 hover:border-zinc-700'
                                    }`}
                            >
                                <Activity className="w-5 h-5" />
                                <span>Expression</span>
                            </button>
                        </div>
                        <p className="text-xs text-zinc-500 mt-2">
                            {teamType === 'action' ? 'Focuses on logic, computation, and content generation.' : 'Focuses on output, visualization, and hardware actuation.'}
                        </p>
                    </div>

                    <div>
                        <label className="block text-xs font-medium text-zinc-400 uppercase tracking-wider mb-2">Knowledge Binding</label>
                        <div className="p-4 rounded-md border border-zinc-800 bg-zinc-950 border-dashed flex flex-col items-center justify-center text-zinc-500 hover:border-zinc-600 transition-colors cursor-pointer">
                            <Database className="w-6 h-6 mb-2" />
                            <span className="text-xs">Drag Sources Here</span>
                        </div>
                    </div>
                </div>

                <div className="mt-8 pt-6 border-t border-zinc-800">
                    <button className="w-full bg-zinc-100 text-zinc-900 hover:bg-white font-medium py-2 px-4 rounded-md flex items-center justify-center gap-2 shadow-lg shadow-indigo-500/10">
                        <Save className="w-4 h-4" />
                        Deploy Core
                    </button>
                </div>
            </div>

            {/* Right Panel: Composition Canvas */}
            <div className="col-span-9 bg-[url('/grid.svg')] bg-zinc-950 relative">
                <div className="absolute inset-0 bg-gradient-to-b from-zinc-950/50 to-zinc-950/80 pointer-events-none" />

                <div className="relative z-10 p-8 h-full flex items-center justify-center">
                    <div className="text-center">
                        <div className="w-24 h-24 rounded-full bg-zinc-900 border-2 border-dashed border-zinc-800 flex items-center justify-center mx-auto mb-4">
                            <Users className="w-8 h-8 text-zinc-600" />
                        </div>
                        <h3 className="text-lg font-medium text-zinc-400">Agent Composition Canvas</h3>
                        <p className="text-zinc-600">Drag agents from the registry to add them to this core.</p>
                    </div>
                </div>
            </div>
        </div>
    );
}
