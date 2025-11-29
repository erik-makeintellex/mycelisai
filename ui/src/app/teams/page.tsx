'use client';

import { useState, useEffect, useRef } from 'react';
import Link from 'next/link';
import { useEventStream } from '@/hooks/useEventStream';
import { API_BASE_URL } from '@/config';

interface Team {
    id: string;
    name: string;
    description: string;
    agents: string[];
    channels: string[];
    inter_comm_channel?: string;
    resource_access: Record<string, string>;
    created_at: string;
}

function TeamLogs({ channel }: { channel: string }) {
    const { events, stats } = useEventStream(channel);
    const scrollRef = useRef<HTMLDivElement>(null);

    useEffect(() => {
        if (scrollRef.current) {
            scrollRef.current.scrollTop = scrollRef.current.scrollHeight;
        }
    }, [events]);

    return (
        <div className="flex flex-col h-full bg-zinc-950 rounded-lg border border-zinc-800 overflow-hidden font-mono text-sm">
            <div className="flex items-center justify-between px-4 py-2 bg-zinc-900 border-b border-zinc-800">
                <div className="flex items-center gap-2">
                    <div className={`w-2 h-2 rounded-full ${stats.isConnected ? 'bg-emerald-500 animate-pulse' : 'bg-red-500'}`} />
                    <span className="text-zinc-400">Live Logs: {channel}</span>
                </div>
                <div className="text-xs text-zinc-500">
                    {stats.eventsPerSecond} eps | {stats.totalEvents} total
                </div>
            </div>
            <div ref={scrollRef} className="flex-1 overflow-y-auto p-4 space-y-2">
                {events.length === 0 && (
                    <div className="text-zinc-600 italic">Waiting for events...</div>
                )}
                {events.map((event, i) => (
                    <div key={i} className="flex gap-2">
                        <span className="text-zinc-500 shrink-0">[{new Date(event.timestamp * 1000).toLocaleTimeString()}]</span>
                        <span className="text-blue-400 shrink-0">{event.source}</span>
                        <span className="text-zinc-300 break-all">
                            {typeof event.payload === 'string' ? event.payload : JSON.stringify(event.payload)}
                        </span>
                    </div>
                ))}
            </div>
        </div>
    );
}

export default function TeamsPage() {
    const [teams, setTeams] = useState<Team[]>([]);
    const [selectedTeam, setSelectedTeam] = useState<Team | null>(null);
    const [loading, setLoading] = useState(true);
    const [showAddAgent, setShowAddAgent] = useState(false);
    const [showEditTeam, setShowEditTeam] = useState(false);
    const [availableAgents, setAvailableAgents] = useState<string[]>([]);
    const [selectedAgentToAdd, setSelectedAgentToAdd] = useState('');
    const [editForm, setEditForm] = useState({
        name: '',
        description: '',
        channels: '',
        resource_access: ''
    });

    const fetchTeams = async () => {
        try {
            const res = await fetch(`${API_BASE_URL}/teams`);
            const data = await res.json();
            setTeams(data);
            if (data.length > 0) {
                // If we have a selected team, update it from the fresh data
                if (selectedTeam) {
                    const updated = data.find((t: Team) => t.id === selectedTeam.id);
                    if (updated) setSelectedTeam(updated);
                } else {
                    setSelectedTeam(data[0]);
                }
            }
        } catch (error) {
            // eslint-disable-next-line no-console
            console.error(error);
        } finally {
            setLoading(false);
        }
    };

    const fetchAgents = async () => {
        try {
            const res = await fetch(`${API_BASE_URL}/agents`);
            const data = await res.json();
            // Filter out agents already in the team
            const allAgents = data.map((a: any) => a.name);
            setAvailableAgents(allAgents);
            if (allAgents.length > 0) setSelectedAgentToAdd(allAgents[0]);
        } catch (error) {
            // eslint-disable-next-line no-console
            console.error(error);
        }
    };

    const handleAddAgent = async () => {
        if (!selectedTeam || !selectedAgentToAdd) return;
        try {
            const res = await fetch(`${API_BASE_URL}/teams/${selectedTeam.id}/agents/${selectedAgentToAdd}`, {
                method: 'POST'
            });
            if (res.ok) {
                setShowAddAgent(false);
                fetchTeams(); // Refresh team data
            }
        } catch (error) {
            // eslint-disable-next-line no-console
            console.error(error);
        }
    };

    const handleRemoveAgent = async (agentName: string) => {
        if (!selectedTeam) return;
        if (!confirm(`Are you sure you want to remove ${agentName} from the team?`)) return;
        try {
            const res = await fetch(`${API_BASE_URL}/teams/${selectedTeam.id}/agents/${agentName}`, {
                method: 'DELETE'
            });
            if (res.ok) {
                fetchTeams();
            }
        } catch (error) {
            // eslint-disable-next-line no-console
            console.error(error);
        }
    };

    const handleEditTeam = async () => {
        if (!selectedTeam) return;
        try {
            const resourceAccessObj: Record<string, string> = {};
            editForm.resource_access.split(',').forEach(pair => {
                const [key, val] = pair.split(':').map(s => s.trim());
                if (key && val) resourceAccessObj[key] = val;
            });

            const body = {
                name: editForm.name,
                description: editForm.description,
                channels: editForm.channels.split(',').map(s => s.trim()).filter(Boolean),
                resource_access: resourceAccessObj
            };

            const res = await fetch(`${API_BASE_URL}/teams/${selectedTeam.id}`, {
                method: 'PUT',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify(body)
            });

            if (res.ok) {
                setShowEditTeam(false);
                fetchTeams();
            }
        } catch (error) {
            // eslint-disable-next-line no-console
            console.error(error);
        }
    };

    const openEditModal = () => {
        if (!selectedTeam) return;
        setEditForm({
            name: selectedTeam.name,
            description: selectedTeam.description,
            channels: selectedTeam.channels.join(', '),
            resource_access: Object.entries(selectedTeam.resource_access).map(([k, v]) => `${k}:${v}`).join(', ')
        });
        setShowEditTeam(true);
    };

    useEffect(() => {
        fetchTeams();
    }, []);

    return (
        <div className="h-[calc(100vh-6rem)] flex flex-col">
            <div className="flex justify-between items-center mb-6">
                <h1 className="text-3xl font-bold text-zinc-100">Teams</h1>
                <Link
                    href="/teams/create"
                    className="bg-zinc-100 text-zinc-900 px-4 py-2 rounded-lg hover:bg-white transition-colors font-semibold shadow-lg shadow-zinc-900/20"
                >
                    + Create Team
                </Link>
            </div>

            <div className="flex-1 flex gap-6 overflow-hidden">
                {/* Sidebar: Team List */}
                <div className="w-1/3 flex flex-col bg-zinc-900 rounded-xl border border-zinc-800 overflow-hidden">
                    <div className="p-4 border-b border-zinc-800 bg-zinc-900/50">
                        <h2 className="font-semibold text-zinc-300">All Teams</h2>
                    </div>
                    <div className="flex-1 overflow-y-auto p-2 space-y-2">
                        {loading ? (
                            <p className="text-zinc-500 p-4">Loading teams...</p>
                        ) : teams.length === 0 ? (
                            <p className="text-zinc-500 p-4">No teams found.</p>
                        ) : (
                            teams.map(team => (
                                <button
                                    key={team.id}
                                    onClick={() => setSelectedTeam(team)}
                                    className={`w-full text-left p-4 rounded-lg transition-all border ${selectedTeam?.id === team.id
                                        ? 'bg-zinc-800 border-zinc-600 shadow-md'
                                        : 'bg-transparent border-transparent hover:bg-zinc-800/50 hover:border-zinc-700'
                                        }`}
                                >
                                    <div className="flex justify-between items-start mb-1">
                                        <span className={`font-medium ${selectedTeam?.id === team.id ? 'text-zinc-100' : 'text-zinc-300'}`}>
                                            {team.name}
                                        </span>
                                        <span className="text-xs bg-zinc-950 text-zinc-500 px-2 py-0.5 rounded-full border border-zinc-800">
                                            {team.agents.length} Agents
                                        </span>
                                    </div>
                                    <p className="text-sm text-zinc-500 line-clamp-2">{team.description}</p>
                                </button>
                            ))
                        )}
                    </div>
                </div>

                {/* Main Area: Team Details */}
                <div className="w-2/3 flex flex-col gap-6 overflow-hidden">
                    {selectedTeam ? (
                        <>
                            {/* Team Header & Stats */}
                            <div className="bg-zinc-900 rounded-xl border border-zinc-800 p-6 shadow-lg">
                                <div className="flex justify-between items-start mb-4">
                                    <div>
                                        <h2 className="text-2xl font-bold text-zinc-100">{selectedTeam.name}</h2>
                                        <p className="text-zinc-400 mt-1">{selectedTeam.description}</p>
                                    </div>
                                    <div className="text-right">
                                        <div className="text-xs text-zinc-500 font-mono">ID: {selectedTeam.id}</div>
                                        <div className="text-xs text-zinc-500 mt-1">Created: {new Date(selectedTeam.created_at).toLocaleDateString()}</div>
                                        <div className="flex gap-2 justify-end mt-2">
                                            <button
                                                onClick={openEditModal}
                                                className="text-xs bg-zinc-800 hover:bg-zinc-700 text-zinc-300 px-3 py-1 rounded border border-zinc-700 transition-colors"
                                            >
                                                Edit Team
                                            </button>
                                            <button
                                                onClick={() => {
                                                    fetchAgents();
                                                    setShowAddAgent(true);
                                                }}
                                                className="text-xs bg-zinc-100 hover:bg-white text-zinc-900 px-3 py-1 rounded border border-zinc-200 transition-colors font-semibold"
                                            >
                                                + Add Agent
                                            </button>
                                        </div>
                                    </div>
                                </div>

                                <div className="grid grid-cols-3 gap-4 mt-6">
                                    <div className="bg-zinc-950/50 p-3 rounded-lg border border-zinc-800">
                                        <div className="text-xs text-zinc-500 uppercase tracking-wider mb-2">Agents</div>
                                        <div className="flex flex-wrap gap-2">
                                            {selectedTeam.agents.map(agent => (
                                                <div key={agent} className="flex items-center gap-1 text-xs px-2 py-1 rounded bg-emerald-900/20 text-emerald-400 border border-emerald-900/30 group">
                                                    <span>{agent}</span>
                                                    <button
                                                        onClick={() => handleRemoveAgent(agent)}
                                                        className="ml-1 text-emerald-600 hover:text-emerald-300 opacity-0 group-hover:opacity-100 transition-opacity"
                                                    >
                                                        ×
                                                    </button>
                                                </div>
                                            ))}
                                            {selectedTeam.agents.length === 0 && <span className="text-zinc-600 text-sm">-</span>}
                                        </div>
                                    </div>
                                    <div className="bg-zinc-950/50 p-3 rounded-lg border border-zinc-800">
                                        <div className="text-xs text-zinc-500 uppercase tracking-wider mb-2">Channels</div>
                                        <div className="flex flex-wrap gap-2">
                                            {selectedTeam.channels.map(channel => (
                                                <span key={channel} className="text-xs px-2 py-1 rounded bg-blue-900/20 text-blue-400 border border-blue-900/30">
                                                    {channel}
                                                </span>
                                            ))}
                                        </div>
                                    </div>
                                    <div className="bg-zinc-950/50 p-3 rounded-lg border border-zinc-800">
                                        <div className="text-xs text-zinc-500 uppercase tracking-wider mb-2">Resources</div>
                                        <div className="space-y-1">
                                            {Object.entries(selectedTeam.resource_access).map(([res, access]) => (
                                                <div key={res} className="flex justify-between text-xs">
                                                    <span className="text-zinc-400">{res}</span>
                                                    <span className="text-zinc-500">{access}</span>
                                                </div>
                                            ))}
                                            {Object.keys(selectedTeam.resource_access).length === 0 && <span className="text-zinc-600 text-sm">-</span>}
                                        </div>
                                    </div>
                                </div>
                            </div>

                            {/* Live Logs */}
                            <div className="flex-1 min-h-0">
                                {selectedTeam.inter_comm_channel ? (
                                    <TeamLogs channel={`team.${selectedTeam.id}`} />
                                    /* Note: Backend stream endpoint appends .>, so we pass team ID or channel base? 
                                       The backend expects 'channel' and subscribes to '{channel}.>'.
                                       Our team channel is usually 'team.{id}.chat'.
                                       If we want to see ALL team events, we should probably subscribe to 'team.{id}'.
                                       Let's pass 'team.{id}' so we get 'team.{id}.>' events.
                                    */
                                ) : (
                                    <div className="h-full bg-zinc-900 rounded-xl border border-zinc-800 flex items-center justify-center text-zinc-500">
                                        No communication channel configured.
                                    </div>
                                )}
                            </div>
                        </>
                    ) : (
                        <div className="flex-1 bg-zinc-900 rounded-xl border border-zinc-800 flex items-center justify-center text-zinc-500">
                            Select a team to view details.
                        </div>
                    )}
                </div>
            </div>

            {/* Add Agent Modal */}
            {showAddAgent && (
                <div className="fixed inset-0 bg-black/50 backdrop-blur-sm flex items-center justify-center z-50">
                    <div className="bg-zinc-900 border border-zinc-700 rounded-xl p-6 w-96 shadow-2xl">
                        <h3 className="text-lg font-bold text-zinc-100 mb-4">Add Agent to Team</h3>
                        <div className="mb-4">
                            <label className="block text-sm font-medium mb-1 text-zinc-400">Select Agent</label>
                            <select
                                value={selectedAgentToAdd}
                                onChange={(e) => setSelectedAgentToAdd(e.target.value)}
                                className="w-full p-2 border border-zinc-700 rounded-lg bg-zinc-950 text-zinc-200 focus:ring-2 focus:ring-zinc-600 outline-none"
                            >
                                {availableAgents.filter(a => !selectedTeam?.agents.includes(a)).map(agent => (
                                    <option key={agent} value={agent}>{agent}</option>
                                ))}
                                {availableAgents.filter(a => !selectedTeam?.agents.includes(a)).length === 0 && (
                                    <option disabled>No available agents</option>
                                )}
                            </select>
                        </div>
                        <div className="flex justify-end gap-2">
                            <button
                                onClick={() => setShowAddAgent(false)}
                                className="px-4 py-2 text-zinc-400 hover:text-zinc-200 transition-colors"
                            >
                                Cancel
                            </button>
                            <button
                                onClick={handleAddAgent}
                                disabled={availableAgents.filter(a => !selectedTeam?.agents.includes(a)).length === 0}
                                className="bg-zinc-100 text-zinc-900 px-4 py-2 rounded-lg hover:bg-white transition-colors font-semibold disabled:opacity-50 disabled:cursor-not-allowed"
                            >
                                Add Agent
                            </button>
                        </div>
                    </div>
                </div>
            )}

            {/* Edit Team Modal */}
            {showEditTeam && (
                <div className="fixed inset-0 bg-black/50 backdrop-blur-sm flex items-center justify-center z-50">
                    <div className="bg-zinc-900 border border-zinc-700 rounded-xl p-6 w-96 shadow-2xl">
                        <h3 className="text-lg font-bold text-zinc-100 mb-4">Edit Team</h3>
                        <div className="space-y-4">
                            <div>
                                <label className="block text-sm font-medium mb-1 text-zinc-400">Name</label>
                                <input
                                    type="text"
                                    value={editForm.name}
                                    onChange={(e) => setEditForm({ ...editForm, name: e.target.value })}
                                    className="w-full p-2 border border-zinc-700 rounded-lg bg-zinc-950 text-zinc-200 focus:ring-2 focus:ring-zinc-600 outline-none"
                                />
                            </div>
                            <div>
                                <label className="block text-sm font-medium mb-1 text-zinc-400">Description</label>
                                <textarea
                                    value={editForm.description}
                                    onChange={(e) => setEditForm({ ...editForm, description: e.target.value })}
                                    className="w-full p-2 border border-zinc-700 rounded-lg bg-zinc-950 text-zinc-200 focus:ring-2 focus:ring-zinc-600 outline-none"
                                    rows={3}
                                />
                            </div>
                            <div>
                                <label className="block text-sm font-medium mb-1 text-zinc-400">Channels (comma separated)</label>
                                <input
                                    type="text"
                                    value={editForm.channels}
                                    onChange={(e) => setEditForm({ ...editForm, channels: e.target.value })}
                                    className="w-full p-2 border border-zinc-700 rounded-lg bg-zinc-950 text-zinc-200 focus:ring-2 focus:ring-zinc-600 outline-none"
                                    placeholder="admin, general"
                                />
                                <p className="text-xs text-zinc-500 mt-1">
                                    NATS subjects for event routing.
                                </p>
                            </div>
                            <div>
                                <label className="block text-sm font-medium mb-1 text-zinc-400">Resource Access (key:value, comma separated)</label>
                                <input
                                    type="text"
                                    value={editForm.resource_access}
                                    onChange={(e) => setEditForm({ ...editForm, resource_access: e.target.value })}
                                    className="w-full p-2 border border-zinc-700 rounded-lg bg-zinc-950 text-zinc-200 focus:ring-2 focus:ring-zinc-600 outline-none"
                                    placeholder="db:read, s3:write"
                                />
                                <p className="text-xs text-zinc-500 mt-1">
                                    Permissions. Format: <code>key:value</code>
                                </p>
                            </div>
                        </div>
                        <div className="flex justify-end gap-2 mt-6">
                            <button
                                onClick={() => setShowEditTeam(false)}
                                className="px-4 py-2 text-zinc-400 hover:text-zinc-200 transition-colors"
                            >
                                Cancel
                            </button>
                            <button
                                onClick={handleEditTeam}
                                className="bg-zinc-100 text-zinc-900 px-4 py-2 rounded-lg hover:bg-white transition-colors font-semibold"
                            >
                                Save Changes
                            </button>
                        </div>
                    </div>
                </div>
            )}
        </div>
    );
}
