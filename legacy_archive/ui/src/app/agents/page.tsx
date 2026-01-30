'use client';

import { useState, useEffect } from 'react';
import AgentForm from '@/components/AgentForm';
import { API_BASE_URL } from '@/config';
import Link from 'next/link';
import { Trash2 } from 'lucide-react';

interface Agent {
    name: string;
    languages: string[];
    capabilities: string[];
    backend: string;
}

export default function AgentsPage() {
    const [agents, setAgents] = useState<Agent[]>([]);
    const [editingAgent, setEditingAgent] = useState<Agent | null>(null);
    const [error, setError] = useState<string | null>(null);

    const fetchAgents = async () => {
        try {
            const res = await fetch(`${API_BASE_URL}/agents`);
            const data = await res.json();
            setAgents(data);
        } catch (error: any) {
            // eslint-disable-next-line no-console
            console.error(error);
            setError(`Agent Fetch Error: ${error.message}`);
        }
    };

    const handleDelete = async (name: string) => {
        if (!confirm(`Are you sure you want to delete agent ${name}?`)) return;
        try {
            const res = await fetch(`${API_BASE_URL}/agents/${name}`, {
                method: 'DELETE'
            });
            if (res.ok) {
                fetchAgents();
                if (editingAgent?.name === name) setEditingAgent(null);
            }
        } catch (error) {
            // eslint-disable-next-line no-console
            console.error(error);
        }
    };

    useEffect(() => {
        fetchAgents();
    }, []);

    return (
        <div className="space-y-8">
            <div>
                <h1 className="text-3xl font-bold text-zinc-100">Agent Management</h1>
                <p className="text-zinc-400 mt-2">
                    Agents are autonomous workers powered by LLMs. Define their capabilities, backend models, and communication channels here.
                </p>
            </div>

            {error && (
                <div className="p-4 text-sm text-red-400 bg-red-900/20 border border-red-900 rounded-lg">
                    {error}
                </div>
            )}

            <div className="grid grid-cols-1 md:grid-cols-2 gap-8">
                <div>
                    <h2 className="text-xl font-semibold mb-4 text-zinc-300">
                        {editingAgent ? `Edit Agent: ${editingAgent.name}` : 'Create New Agent'}
                    </h2>
                    {editingAgent && (
                        <button
                            onClick={() => setEditingAgent(null)}
                            className="text-xs text-zinc-400 hover:text-zinc-200 underline mb-2"
                        >
                            Cancel Edit
                        </button>
                    )}
                    <AgentForm
                        initialData={editingAgent}
                        onSuccess={() => {
                            fetchAgents();
                            setEditingAgent(null);
                        }}
                    />
                </div>

                <div>
                    <div className="flex justify-between items-center mb-4">
                        <h2 className="text-xl font-semibold text-zinc-300">Active Agents</h2>
                        <button
                            onClick={fetchAgents}
                            className="text-xs text-zinc-400 hover:text-zinc-200 underline"
                        >
                            Refresh
                        </button>
                    </div>
                    <div className="space-y-4">
                        {agents.length === 0 ? (
                            <div className="p-4 border border-zinc-700 rounded-xl bg-zinc-800 shadow-lg">
                                <p className="text-zinc-500 italic">No agents active yet.</p>
                            </div>
                        ) : (
                            agents.map(agent => (
                                <div key={agent.name} className="p-4 border border-zinc-700 rounded-xl bg-zinc-800 shadow-lg">
                                    <div className="flex justify-between items-start">
                                        <h3 className="font-bold text-zinc-200">{agent.name}</h3>
                                        <div className="flex flex-col items-end gap-2">
                                            <div className="flex gap-2">
                                                <span className="text-xs px-2 py-1 rounded bg-zinc-900 text-zinc-400 border border-zinc-700">
                                                    {agent.backend}
                                                </span>
                                                <button
                                                    onClick={() => setEditingAgent(agent)}
                                                    className="text-xs text-blue-400 hover:text-blue-300"
                                                >
                                                    Edit
                                                </button>
                                            </div>
                                            <div className="flex gap-2">
                                                <Link
                                                    href={`/agents/${agent.name}/chat`}
                                                    className="px-3 py-1 bg-emerald-600 hover:bg-emerald-500 text-white rounded text-xs font-medium transition-colors"
                                                >
                                                    Chat
                                                </Link>
                                                <button
                                                    onClick={() => handleDelete(agent.name)}
                                                    className="text-xs text-red-400 hover:text-red-300"
                                                    title="Delete Agent"
                                                >
                                                    <Trash2 size={14} />
                                                </button>
                                            </div>
                                        </div>
                                    </div>
                                    <div className="mt-2 flex flex-wrap gap-2">
                                        {agent.capabilities.map(cap => (
                                            <span key={cap} className="text-xs px-2 py-1 rounded bg-zinc-700 text-zinc-300">
                                                {cap}
                                            </span>
                                        ))}
                                    </div>
                                    <div className="mt-2 text-xs text-zinc-500">
                                        Languages: {agent.languages.join(', ')}
                                    </div>
                                </div>
                            ))
                        )}
                    </div>
                </div>
            </div>
        </div>
    );
}
