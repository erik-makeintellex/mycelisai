'use client';

import { useState, useEffect } from 'react';
import Link from 'next/link';
import { API_BASE_URL } from '@/config';
import TeamList from '@/components/teams/TeamList';
import TeamDetails from '@/components/teams/TeamDetails';
import AddAgentModal from '@/components/teams/AddAgentModal';
import EditTeamModal from '@/components/teams/EditTeamModal';

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
                if (selectedTeam) {
                    const updated = data.find((t: Team) => t.id === selectedTeam.id);
                    if (updated) setSelectedTeam(updated);
                } else {
                    setSelectedTeam(data[0]);
                }
            }
        } catch (error) {
            console.error(error);
        } finally {
            setLoading(false);
        }
    };

    const fetchAgents = async () => {
        try {
            const res = await fetch(`${API_BASE_URL}/agents`);
            const data = await res.json();
            const allAgents = data.map((a: any) => a.name);
            setAvailableAgents(allAgents);
            if (allAgents.length > 0) setSelectedAgentToAdd(allAgents[0]);
        } catch (error) {
            console.error(error);
        }
    };

    const handleAddAgent = async () => {
        if (!selectedTeam || !selectedAgentToAdd) return;
        try {
            const res = await fetch(`${API_BASE_URL}/teams/${encodeURIComponent(selectedTeam.id)}/agents/${encodeURIComponent(selectedAgentToAdd)}`, {
                method: 'POST'
            });
            if (res.ok) {
                setShowAddAgent(false);
                fetchTeams();
            }
        } catch (error) {
            console.error(error);
        }
    };

    const handleRemoveAgent = async (agentName: string) => {
        if (!selectedTeam) return;
        if (!confirm(`Are you sure you want to remove ${agentName} from the team?`)) return;
        try {
            const res = await fetch(`${API_BASE_URL}/teams/${encodeURIComponent(selectedTeam.id)}/agents/${encodeURIComponent(agentName)}`, {
                method: 'DELETE'
            });
            if (res.ok) {
                fetchTeams();
            }
        } catch (error) {
            console.error(error);
        }
    };

    const handleDeleteTeam = async () => {
        if (!selectedTeam) return;
        if (!confirm(`Are you sure you want to delete team "${selectedTeam.name}"? This cannot be undone.`)) return;
        try {
            const res = await fetch(`${API_BASE_URL}/teams/${encodeURIComponent(selectedTeam.id)}`, {
                method: 'DELETE'
            });
            if (res.ok) {
                setSelectedTeam(null);
                fetchTeams();
            }
        } catch (error) {
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

            const res = await fetch(`${API_BASE_URL}/teams/${encodeURIComponent(selectedTeam.id)}`, {
                method: 'PUT',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify(body)
            });

            if (res.ok) {
                setShowEditTeam(false);
                fetchTeams();
            }
        } catch (error) {
            console.error(error);
        }
    };

    const openEditModal = () => {
        if (!selectedTeam) return;
        setEditForm({
            name: selectedTeam.name,
            description: selectedTeam.description || '',
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
                <TeamList
                    teams={teams}
                    selectedTeam={selectedTeam}
                    onSelect={setSelectedTeam}
                    loading={loading}
                />

                <div className="w-2/3 flex flex-col gap-6 overflow-hidden">
                    <TeamDetails
                        team={selectedTeam}
                        onEdit={openEditModal}
                        onDelete={handleDeleteTeam}
                        onAddAgent={() => {
                            fetchAgents();
                            setShowAddAgent(true);
                        }}
                        onRemoveAgent={handleRemoveAgent}
                    />
                </div>
            </div>

            <AddAgentModal
                isOpen={showAddAgent}
                onClose={() => setShowAddAgent(false)}
                onAdd={handleAddAgent}
                availableAgents={availableAgents}
                selectedAgent={selectedAgentToAdd}
                onSelectAgent={setSelectedAgentToAdd}
                teamAgents={selectedTeam?.agents || []}
            />

            <EditTeamModal
                isOpen={showEditTeam}
                onClose={() => setShowEditTeam(false)}
                onSave={handleEditTeam}
                form={editForm}
                onChange={setEditForm}
            />
        </div>
    );
}
