'use client';

import Link from 'next/link';
import TeamLogs from './TeamLogs';

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

interface TeamDetailsProps {
    team: Team | null;
    onEdit: () => void;
    onDelete: () => void;
    onAddAgent: () => void;
    onRemoveAgent: (agent: string) => void;
}

export default function TeamDetails({ team, onEdit, onDelete, onAddAgent, onRemoveAgent }: TeamDetailsProps) {
    if (!team) {
        return (
            <div className="flex-1 bg-zinc-900 rounded-xl border border-zinc-800 flex items-center justify-center text-zinc-500">
                Select a team to view details.
            </div>
        );
    }

    return (
        <>
            {/* Team Header & Stats */}
            <div className="bg-zinc-900 rounded-xl border border-zinc-800 p-6 shadow-lg">
                <div className="flex justify-between items-start mb-4">
                    <div>
                        <h2 className="text-2xl font-bold text-zinc-100">{team.name}</h2>
                        <p className="text-zinc-400 mt-1">{team.description}</p>
                    </div>
                    <div className="text-right">
                        <div className="text-xs text-zinc-500 font-mono">ID: {team.id}</div>
                        <div className="text-xs text-zinc-500 mt-1">Created: {new Date(team.created_at).toLocaleDateString()}</div>
                        <div className="flex gap-2 justify-end mt-2">
                            <button
                                onClick={onEdit}
                                className="text-xs bg-zinc-800 hover:bg-zinc-700 text-zinc-300 px-3 py-1 rounded border border-zinc-700 transition-colors"
                            >
                                Edit Team
                            </button>
                            <button
                                onClick={onDelete}
                                className="text-xs bg-red-900/30 hover:bg-red-900/50 text-red-400 px-3 py-1 rounded border border-red-900/50 transition-colors"
                            >
                                Delete
                            </button>
                            <Link
                                href={`/teams/${team.id}/chat`}
                                className="text-xs bg-blue-600 hover:bg-blue-500 text-white px-3 py-1 rounded border border-blue-500 transition-colors font-semibold flex items-center gap-1"
                            >
                                <span>💬</span> Chat
                            </Link>
                            <button
                                onClick={onAddAgent}
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
                            {team.agents.map(agent => (
                                <div key={agent} className="flex items-center gap-1 text-xs px-2 py-1 rounded bg-emerald-900/20 text-emerald-400 border border-emerald-900/30 group">
                                    <span>{agent}</span>
                                    <button
                                        onClick={() => onRemoveAgent(agent)}
                                        className="ml-1 text-emerald-600 hover:text-emerald-300 opacity-0 group-hover:opacity-100 transition-opacity"
                                    >
                                        ×
                                    </button>
                                </div>
                            ))}
                            {team.agents.length === 0 && <span className="text-zinc-600 text-sm">-</span>}
                        </div>
                    </div>
                    <div className="bg-zinc-950/50 p-3 rounded-lg border border-zinc-800">
                        <div className="text-xs text-zinc-500 uppercase tracking-wider mb-2">Channels</div>
                        <div className="flex flex-wrap gap-2">
                            {team.channels.map(channel => (
                                <span key={channel} className="text-xs px-2 py-1 rounded bg-blue-900/20 text-blue-400 border border-blue-900/30">
                                    {channel}
                                </span>
                            ))}
                        </div>
                    </div>
                    <div className="bg-zinc-950/50 p-3 rounded-lg border border-zinc-800">
                        <div className="text-xs text-zinc-500 uppercase tracking-wider mb-2">Resources</div>
                        <div className="space-y-1">
                            {Object.entries(team.resource_access).map(([res, access]) => (
                                <div key={res} className="flex justify-between text-xs">
                                    <span className="text-zinc-400">{res}</span>
                                    <span className="text-zinc-500">{access}</span>
                                </div>
                            ))}
                            {Object.keys(team.resource_access).length === 0 && <span className="text-zinc-600 text-sm">-</span>}
                        </div>
                    </div>
                </div>
            </div>

            {/* Live Logs */}
            <div className="flex-1 min-h-0">
                {team.inter_comm_channel ? (
                    <TeamLogs channel={`team.${team.id}`} />
                ) : (
                    <div className="h-full bg-zinc-900 rounded-xl border border-zinc-800 flex items-center justify-center text-zinc-500">
                        No communication channel configured.
                    </div>
                )}
            </div>
        </>
    );
}
