"use client";

import React, { useState } from 'react';
import { Users, Cpu, Activity, Database, Save, Plus, Trash, User } from 'lucide-react';

type TeamType = 'action' | 'expression';

interface TeamMember {
    id: string;
    role: string;
    system_prompt?: string;
    model?: string;
}

export default function TeamBuilder() {
    const [teamName, setTeamName] = useState('');
    const [teamType, setTeamType] = useState<TeamType>('action');
    const [members, setMembers] = useState<TeamMember[]>([{ id: 'architect', role: 'lead' }]);
    const [newMemberId, setNewMemberId] = useState('');
    const [newMemberRole, setNewMemberRole] = useState('');
    // New State
    const [newSystemPrompt, setNewSystemPrompt] = useState('');
    const [newModel, setNewModel] = useState('');

    const [inputs, setInputs] = useState<string[]>([]);
    const [deliveries, setDeliveries] = useState<string[]>([]);
    const [newInput, setNewInput] = useState('');
    const [newDelivery, setNewDelivery] = useState('');

    const addMember = () => {
        if (newMemberId && newMemberRole) {
            setMembers([...members, {
                id: newMemberId,
                role: newMemberRole,
                system_prompt: newSystemPrompt,
                model: newModel
            }]);
            setNewMemberId('');
            setNewMemberRole('');
            setNewSystemPrompt('');
            setNewModel('');
        }
    };

    const removeMember = (index: number) => {
        const newMembers = [...members];
        newMembers.splice(index, 1);
        setMembers(newMembers);
    };

    const addInput = () => {
        if (newInput) {
            setInputs([...inputs, newInput]);
            setNewInput('');
        }
    };

    const addDelivery = () => {
        if (newDelivery) {
            setDeliveries([...deliveries, newDelivery]);
            setNewDelivery('');
        }
    };

    const handleDeploy = async () => {
        if (!teamName) return;
        if (members.length === 0) {
            alert('At least one agent is required.');
            return;
        }

        const id = teamName.toLowerCase().replace(/\s+/g, '-');

        // Default input if none specified
        const finalInputs = inputs.length > 0 ? inputs : [`swarm.team.${id}.input`];

        const manifest = {
            id,
            name: teamName,
            type: teamType,
            members: members,
            inputs: finalInputs,
            deliveries: deliveries
        };

        try {
            const res = await fetch('/api/swarm/teams', {
                method: 'POST',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify(manifest)
            });

            if (res.ok) {
                alert(`Team ${teamName} Spawned!`);
                setTeamName('');
                setMembers([{ id: 'architect', role: 'lead' }]);
                setInputs([]);
                setDeliveries([]);
            } else {
                const err = await res.text();
                alert(`Failed: ${err}`);
            }
        } catch (e) {
            console.error(e);
            alert('Deploy Error');
        }
    };

    return (
        <div className="h-full grid grid-cols-12 gap-0">
            {/* Left Panel: Configuration */}
            <div className="col-span-3 border-r border-zinc-800 p-6 bg-zinc-900/50 flex flex-col h-full">
                <h2 className="text-lg font-semibold mb-6 flex items-center gap-2">
                    <Cpu className="w-5 h-5 text-indigo-400" />
                    Core Configuration
                </h2>

                <div className="space-y-6 flex-1 overflow-y-auto pr-2">
                    <div>
                        <label className="block text-xs font-medium text-zinc-400 uppercase tracking-wider mb-2">Team Name</label>
                        <input
                            type="text"
                            className="w-full bg-zinc-950 border border-zinc-800 rounded-md p-2 text-sm focus:ring-1 focus:ring-indigo-500 outline-none text-zinc-200 placeholder-zinc-600"
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
                    </div>

                    {/* Communication Wiring */}
                    <div>
                        <label className="block text-xs font-medium text-zinc-400 uppercase tracking-wider mb-2">Neural Wiring</label>
                        <div className="space-y-4 bg-zinc-950 border border-zinc-800 rounded-md p-3">
                            {/* Inputs */}
                            <div>
                                <div className="text-[10px] text-zinc-500 uppercase mb-1 flex justify-between">
                                    <span>Inputs (Triggers)</span>
                                </div>
                                <div className="flex gap-2 mb-2">
                                    <input
                                        className="flex-1 bg-zinc-900 border border-zinc-800 rounded px-2 py-1 text-xs outline-none focus:border-indigo-500"
                                        placeholder="swarm.topic.>"
                                        value={newInput}
                                        onChange={(e) => setNewInput(e.target.value)}
                                    />
                                    <button onClick={addInput} disabled={!newInput} className="bg-zinc-800 px-2 rounded disabled:opacity-50 hover:bg-zinc-700">+</button>
                                </div>
                                <div className="space-y-1">
                                    {inputs.map((item, i) => (
                                        <div key={i} className="text-xs bg-indigo-500/10 text-indigo-300 px-2 py-1 rounded flex justify-between items-center">
                                            <span className="truncate">{item}</span>
                                            <button onClick={() => setInputs(inputs.filter((_, idx) => idx !== i))} className="hover:text-red-400">×</button>
                                        </div>
                                    ))}
                                    {inputs.length === 0 && <span className="text-[10px] text-zinc-600 italic">Defaults to swarm.team.&#123;id&#125;.input</span>}
                                </div>
                            </div>

                            <div className="h-px bg-zinc-800/50" />

                            {/* Outputs */}
                            <div>
                                <div className="text-[10px] text-zinc-500 uppercase mb-1">Outputs (Deliveries)</div>
                                <div className="flex gap-2 mb-2">
                                    <input
                                        className="flex-1 bg-zinc-900 border border-zinc-800 rounded px-2 py-1 text-xs outline-none focus:border-emerald-500"
                                        placeholder="Target Topic / Team"
                                        value={newDelivery}
                                        onChange={(e) => setNewDelivery(e.target.value)}
                                    />
                                    <button onClick={addDelivery} disabled={!newDelivery} className="bg-zinc-800 px-2 rounded disabled:opacity-50 hover:bg-zinc-700">+</button>
                                </div>
                                <div className="space-y-1">
                                    {deliveries.map((item, i) => (
                                        <div key={i} className="text-xs bg-emerald-500/10 text-emerald-300 px-2 py-1 rounded flex justify-between items-center">
                                            <span className="truncate">{item}</span>
                                            <button onClick={() => setDeliveries(deliveries.filter((_, idx) => idx !== i))} className="hover:text-red-400">×</button>
                                        </div>
                                    ))}
                                </div>
                            </div>
                        </div>
                    </div>

                    {/* Agent Composition */}
                    <div>
                        <label className="block text-xs font-medium text-zinc-400 uppercase tracking-wider mb-2">Agent Roster</label>
                        <div className="bg-zinc-950 border border-zinc-800 rounded-md p-3 space-y-3">
                            {/* Add New Agent */}
                            <div className="space-y-3 p-3 bg-zinc-900/30 rounded border border-zinc-900">
                                <div className="grid grid-cols-2 gap-2">
                                    <input
                                        className="bg-zinc-900 border border-zinc-800 rounded px-2 py-1.5 text-xs outline-none focus:border-indigo-500"
                                        placeholder="Agent ID (e.g. poet)"
                                        value={newMemberId}
                                        onChange={(e) => setNewMemberId(e.target.value)}
                                    />
                                    <input
                                        className="bg-zinc-900 border border-zinc-800 rounded px-2 py-1.5 text-xs outline-none focus:border-indigo-500"
                                        placeholder="Role (e.g. creative writer)"
                                        value={newMemberRole}
                                        onChange={(e) => setNewMemberRole(e.target.value)}
                                    />
                                </div>

                                {/* Advanced Config */}
                                <div className="grid grid-cols-2 gap-2">
                                    <input
                                        className="bg-zinc-900 border border-zinc-800 rounded px-2 py-1.5 text-xs outline-none focus:border-indigo-500 font-mono"
                                        placeholder="System Prompt (Optional)"
                                        value={newSystemPrompt}
                                        onChange={(e) => setNewSystemPrompt(e.target.value)}
                                    />
                                    <input
                                        className="bg-zinc-900 border border-zinc-800 rounded px-2 py-1.5 text-xs outline-none focus:border-indigo-500 font-mono"
                                        placeholder="Model Profile (e.g. chat)"
                                        value={newModel}
                                        onChange={(e) => setNewModel(e.target.value)}
                                    />
                                </div>

                                <button
                                    onClick={addMember}
                                    disabled={!newMemberId || !newMemberRole}
                                    className="w-full bg-indigo-900/30 hover:bg-indigo-900/50 text-indigo-300 border border-indigo-500/30 rounded py-1.5 flex items-center justify-center text-xs font-medium disabled:opacity-50 transition-colors"
                                >
                                    <Plus className="w-3 h-3 mr-1" /> Add Agent
                                </button>
                            </div>

                            {/* List */}
                            <div className="space-y-2 mt-2 max-h-48 overflow-y-auto">
                                {members.map((m, i) => (
                                    <div key={i} className="flex flex-col bg-zinc-900/50 p-2 rounded border border-zinc-800/50 group hover:border-zinc-700 transition-colors">
                                        <div className="flex justify-between items-start">
                                            <div className="flex items-center gap-2">
                                                <div className="w-5 h-5 rounded-full bg-indigo-500/10 flex items-center justify-center flex-shrink-0">
                                                    <User className="w-3 h-3 text-indigo-400" />
                                                </div>
                                                <div>
                                                    <div className="flex items-center gap-2">
                                                        <span className="text-xs font-bold text-zinc-300">{m.id}</span>
                                                        <span className="text-[10px] text-zinc-500 uppercase tracking-wide bg-zinc-950 px-1 rounded">{m.role}</span>
                                                    </div>
                                                </div>
                                            </div>
                                            <button
                                                onClick={() => removeMember(i)}
                                                className="text-zinc-600 hover:text-red-400 p-1 opacity-0 group-hover:opacity-100 transition-opacity"
                                            >
                                                <Trash className="w-3 h-3" />
                                            </button>
                                        </div>
                                        {/* Meta details */}
                                        {(m.system_prompt || m.model) && (
                                            <div className="mt-1.5 ml-7 space-y-0.5">
                                                {m.model && (
                                                    <div className="text-[10px] text-zinc-500 font-mono flex items-center gap-1">
                                                        <Cpu className="w-3 h-3" /> {m.model}
                                                    </div>
                                                )}
                                                {m.system_prompt && (
                                                    <div className="text-[10px] text-zinc-600 font-mono truncate max-w-[180px]" title={m.system_prompt}>
                                                        "{m.system_prompt}"
                                                    </div>
                                                )}
                                            </div>
                                        )}
                                    </div>
                                ))}
                            </div>
                        </div>
                    </div>

                </div>

                <div className="mt-4 pt-4 border-t border-zinc-800">
                    <button
                        onClick={handleDeploy}
                        className="w-full bg-zinc-100 text-zinc-900 hover:bg-white font-medium py-2 px-4 rounded-md flex items-center justify-center gap-2 shadow-lg shadow-indigo-500/10 active:scale-95 transition-transform"
                    >
                        <Save className="w-4 h-4" />
                        Deploy Core
                    </button>
                </div>
            </div>

            {/* Right Panel: Composition Canvas */}
            <div className="col-span-9 bg-[url('/grid.svg')] bg-zinc-950 relative">
                <div className="absolute inset-0 bg-gradient-to-b from-zinc-950/50 to-zinc-950/80 pointer-events-none" />

                <div className="relative z-10 p-8 h-full flex items-center justify-center">
                    {members.length > 0 ? (
                        <div className="grid grid-cols-2 md:grid-cols-3 gap-6 animate-fade-in">
                            {members.map((m, i) => (
                                <div key={i} className="w-40 h-40 bg-zinc-900/80 backdrop-blur border border-zinc-700/50 rounded-lg flex flex-col items-center justify-center p-4 shadow-xl hover:scale-105 transition-transform">
                                    <div className="w-16 h-16 rounded-full bg-indigo-500/20 flex items-center justify-center mb-3">
                                        <Users className="w-8 h-8 text-indigo-400" />
                                    </div>
                                    <h4 className="text-sm font-medium text-zinc-200">{m.id}</h4>
                                    <span className="text-xs text-zinc-500 uppercase tracking-wide mt-1">{m.role}</span>
                                </div>
                            ))}
                        </div>
                    ) : (
                        <div className="text-center opacity-50">
                            <div className="w-24 h-24 rounded-full bg-zinc-900 border-2 border-dashed border-zinc-800 flex items-center justify-center mx-auto mb-4">
                                <Users className="w-8 h-8 text-zinc-600" />
                            </div>
                            <h3 className="text-lg font-medium text-zinc-400">Agent Composition Canvas</h3>
                            <p className="text-zinc-600">Add agents from the configuration panel to visualize the team.</p>
                        </div>
                    )}
                </div>
            </div>
        </div>
    );
}
