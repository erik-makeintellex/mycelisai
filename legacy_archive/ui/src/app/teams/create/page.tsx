'use client';

import { useState, useEffect } from 'react';
import { useRouter } from 'next/navigation';
import { API_BASE_URL } from '@/config';

interface Agent {
    name: string;
    capabilities: string[];
}

export default function CreateTeamPage() {
    const router = useRouter();
    const [agents, setAgents] = useState<Agent[]>([]);
    const [teamName, setTeamName] = useState('');
    const [description, setDescription] = useState('');
    const [channels, setChannels] = useState('');
    const [interCommChannel, setInterCommChannel] = useState('');
    const [resourceAccess, setResourceAccess] = useState('');
    const [selectedAgents, setSelectedAgents] = useState<string[]>([]);

    useEffect(() => {
        const fetchAgents = async () => {
            try {
                const res = await fetch(`${API_BASE_URL}/agents`);
                const data = await res.json();
                setAgents(data);
            } catch (error) {
                // eslint-disable-next-line no-console
                console.error(error);
            }
        };
        fetchAgents();
    }, []);

    const handleSubmit = async (e: React.FormEvent) => {
        e.preventDefault();
        const team = {
            id: teamName.toLowerCase().replace(/\s+/g, '-'),
            name: teamName,
            description,
            agents: selectedAgents,
            channels: channels.split(',').map(c => c.trim()).filter(Boolean),
            inter_comm_channel: interCommChannel || null,
            resource_access: resourceAccess ? JSON.parse(resourceAccess) : {},
            shared_context: {}
        };

        try {
            const res = await fetch(`${API_BASE_URL}/teams`, {
                method: 'POST',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify(team)
            });
            if (res.ok) {
                router.push('/'); // Redirect to dashboard
            }
        } catch (error) {
            // eslint-disable-next-line no-console
            console.error(error);
        }
    };

    const toggleAgent = (agentName: string) => {
        if (selectedAgents.includes(agentName)) {
            setSelectedAgents(selectedAgents.filter(a => a !== agentName));
        } else {
            setSelectedAgents([...selectedAgents, agentName]);
        }
    };

    return (
        <div className="max-w-2xl mx-auto space-y-8">
            <h1 className="text-3xl font-bold text-zinc-100">Create New Team</h1>

            <form onSubmit={handleSubmit} className="space-y-6 p-6 border border-zinc-700 rounded-xl bg-zinc-800 shadow-lg">
                <div>
                    <label className="block text-sm font-medium mb-1 text-zinc-400">Team Name</label>
                    <input
                        type="text"
                        value={teamName}
                        onChange={(e) => setTeamName(e.target.value)}
                        className="w-full p-2 border border-zinc-700 rounded-lg bg-zinc-950 text-zinc-200 focus:ring-2 focus:ring-zinc-600 outline-none"
                        required
                        placeholder="e.g., Research Squad"
                    />
                </div>

                <div>
                    <label className="block text-sm font-medium mb-1 text-zinc-400">Description</label>
                    <textarea
                        value={description}
                        onChange={(e) => setDescription(e.target.value)}
                        className="w-full p-2 border border-zinc-700 rounded-lg bg-zinc-950 text-zinc-200 focus:ring-2 focus:ring-zinc-600 outline-none h-24"
                        placeholder="Describe the team's purpose..."
                    />
                </div>

                <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
                    <div>
                        <label className="block text-sm font-medium mb-1 text-zinc-400">Channels (comma separated)</label>
                        <input
                            type="text"
                            value={channels}
                            onChange={(e) => setChannels(e.target.value)}
                            className="w-full p-2 border border-zinc-700 rounded-lg bg-zinc-800 text-zinc-200 focus:ring-2 focus:ring-zinc-600 outline-none"
                            placeholder="e.g., sensors, processing, alerts"
                        />
                        <p className="text-xs text-zinc-500 mt-1">
                            NATS subjects this team will interact with. Used for event routing.
                        </p>
                    </div>
                    <div>
                        <label className="block text-sm font-medium mb-1 text-zinc-400">Inter-Comm Channel (Auto-generated)</label>
                        <input
                            type="text"
                            value={interCommChannel}
                            readOnly
                            className="w-full p-2 border border-zinc-700 rounded-lg bg-zinc-900 text-zinc-500 cursor-not-allowed outline-none"
                            placeholder="Will be team.{id}.chat"
                        />
                        <p className="text-xs text-zinc-500 mt-1">Automatically assigned upon creation.</p>
                    </div>
                </div>

                <div>
                    <label className="block text-sm font-medium mb-1 text-zinc-400">Resource Access (JSON)</label>
                    <textarea
                        value={resourceAccess}
                        onChange={(e) => setResourceAccess(e.target.value)}
                        className="w-full p-2 border border-zinc-700 rounded-lg bg-zinc-800 text-zinc-200 focus:ring-2 focus:ring-zinc-600 outline-none h-24 font-mono text-sm"
                        placeholder='{"db-1": "read", "s3-bucket": "write"}'
                    />
                    <p className="text-xs text-zinc-500 mt-1">
                        Define permissions for external resources. Format: <code>{"{ \"resource_id\": \"access_level\" }"}</code>
                    </p>
                </div>

                <div>
                    <label className="block text-sm font-medium mb-3 text-zinc-400">Select Agents</label>
                    <div className="grid grid-cols-1 gap-3">
                        {agents.map(agent => (
                            <div
                                key={agent.name}
                                onClick={() => toggleAgent(agent.name)}
                                className={`p-4 border rounded-lg cursor-pointer transition-all ${selectedAgents.includes(agent.name)
                                    ? 'border-emerald-500 bg-emerald-900/20'
                                    : 'border-zinc-700 bg-zinc-950 hover:border-zinc-600'
                                    }`}
                            >
                                <div className="flex justify-between items-center">
                                    <span className="font-medium text-zinc-200">{agent.name}</span>
                                    {selectedAgents.includes(agent.name) && (
                                        <span className="text-emerald-500 text-sm">Selected</span>
                                    )}
                                </div>
                                <div className="mt-1 flex gap-2">
                                    {agent.capabilities.map(cap => (
                                        <span key={cap} className="text-xs px-2 py-1 rounded bg-zinc-800 text-zinc-400">
                                            {cap}
                                        </span>
                                    ))}
                                </div>
                            </div>
                        ))}
                        {agents.length === 0 && (
                            <p className="text-zinc-500 italic text-center py-4">No agents available. Create an agent first.</p>
                        )}
                    </div>
                </div>

                <button
                    type="submit"
                    className="w-full bg-zinc-100 text-zinc-900 p-3 rounded-lg hover:bg-white transition-colors font-semibold shadow-lg shadow-zinc-900/20"
                >
                    Create Team
                </button>
            </form>
        </div>
    );
}
