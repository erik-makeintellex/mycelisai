'use client';

interface Team {
    id: string;
    name: string;
    description: string;
    agents: string[];
}

interface TeamListProps {
    teams: Team[];
    selectedTeam: Team | null;
    onSelect: (team: Team) => void;
    loading: boolean;
}

export default function TeamList({ teams, selectedTeam, onSelect, loading }: TeamListProps) {
    return (
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
                            onClick={() => onSelect(team)}
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
    );
}
