'use client';
import { useEffect } from 'react';

interface AddAgentModalProps {
    isOpen: boolean;
    onClose: () => void;
    onAdd: () => void;
    availableAgents: string[];
    selectedAgent: string;
    onSelectAgent: (agent: string) => void;
    teamAgents: string[];
}

export default function AddAgentModal({
    isOpen,
    onClose,
    onAdd,
    availableAgents,
    selectedAgent,
    onSelectAgent,
    teamAgents
}: AddAgentModalProps) {
    if (!isOpen) return null;

    const filteredAgents = availableAgents.filter(a => !teamAgents.includes(a));

    // Auto-select first valid agent if current selection is invalid
    useEffect(() => {
        if (filteredAgents.length > 0 && !filteredAgents.includes(selectedAgent)) {
            onSelectAgent(filteredAgents[0]);
        }
    }, [filteredAgents, selectedAgent, onSelectAgent]);

    return (
        <div className="fixed inset-0 bg-black/50 backdrop-blur-sm flex items-center justify-center z-50">
            <div className="bg-zinc-900 border border-zinc-700 rounded-xl p-6 w-96 shadow-2xl">
                <h3 className="text-lg font-bold text-zinc-100 mb-4">Add Agent to Team</h3>
                <div className="mb-4">
                    <label className="block text-sm font-medium mb-1 text-zinc-400">Select Agent</label>
                    <select
                        value={selectedAgent}
                        onChange={(e) => onSelectAgent(e.target.value)}
                        className="w-full p-2 border border-zinc-700 rounded-lg bg-zinc-950 text-zinc-200 focus:ring-2 focus:ring-zinc-600 outline-none"
                    >
                        {filteredAgents.map(agent => (
                            <option key={agent} value={agent}>{agent}</option>
                        ))}
                        {filteredAgents.length === 0 && (
                            <option disabled>No available agents</option>
                        )}
                    </select>
                </div>
                <div className="flex justify-end gap-2">
                    <button
                        onClick={onClose}
                        className="px-4 py-2 text-zinc-400 hover:text-zinc-200 transition-colors"
                    >
                        Cancel
                    </button>
                    <button
                        onClick={onAdd}
                        disabled={filteredAgents.length === 0}
                        className="bg-zinc-100 text-zinc-900 px-4 py-2 rounded-lg hover:bg-white transition-colors font-semibold disabled:opacity-50 disabled:cursor-not-allowed"
                    >
                        Add Agent
                    </button>
                </div>
            </div>
        </div>
    );
}
