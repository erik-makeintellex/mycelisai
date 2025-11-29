'use client';

import { useState, useEffect } from 'react';
import { API_BASE_URL } from '@/config';

interface Channel {
    name: string;
    subject: string;
    stream: string;
    description?: string;
    created_at: string;
}

export default function ChannelsPage() {
    const [channels, setChannels] = useState<Channel[]>([]);
    const [loading, setLoading] = useState(true);
    const [showCreateModal, setShowCreateModal] = useState(false);
    const [createForm, setCreateForm] = useState({
        name: '',
        subject: '',
        description: ''
    });

    const fetchChannels = async () => {
        try {
            const res = await fetch(`${API_BASE_URL}/channels`);
            const data = await res.json();
            setChannels(data);
        } catch (error) {
            console.error(error);
        } finally {
            setLoading(false);
        }
    };

    const handleCreateChannel = async () => {
        try {
            const res = await fetch(`${API_BASE_URL}/channels`, {
                method: 'POST',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify(createForm)
            });
            if (res.ok) {
                setShowCreateModal(false);
                setCreateForm({ name: '', subject: '', description: '' });
                fetchChannels();
            }
        } catch (error) {
            console.error(error);
        }
    };

    useEffect(() => {
        fetchChannels();
    }, []);

    return (
        <div className="h-[calc(100vh-6rem)] flex flex-col p-6 max-w-7xl mx-auto w-full">
            <div className="flex justify-between items-center mb-6">
                <div>
                    <h1 className="text-3xl font-bold text-zinc-100">Channels</h1>
                    <p className="text-zinc-400 mt-1">Manage NATS streams and subjects for agent communication.</p>
                </div>
                <button
                    onClick={() => setShowCreateModal(true)}
                    className="bg-zinc-100 text-zinc-900 px-4 py-2 rounded-lg hover:bg-white transition-colors font-semibold shadow-lg shadow-zinc-900/20"
                >
                    + Create Channel
                </button>
            </div>

            <div className="flex-1 overflow-y-auto">
                {loading ? (
                    <p className="text-zinc-500">Loading channels...</p>
                ) : channels.length === 0 ? (
                    <div className="text-center py-20 bg-zinc-900/50 rounded-xl border border-zinc-800 border-dashed">
                        <p className="text-zinc-500">No channels found.</p>
                        <button
                            onClick={() => setShowCreateModal(true)}
                            className="text-emerald-400 hover:text-emerald-300 mt-2 text-sm font-medium"
                        >
                            Create your first channel
                        </button>
                    </div>
                ) : (
                    <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
                        {channels.map((channel, idx) => (
                            <div key={idx} className="bg-zinc-900 rounded-xl border border-zinc-800 p-6 hover:border-zinc-700 transition-colors">
                                <div className="flex justify-between items-start mb-4">
                                    <h3 className="font-bold text-lg text-zinc-100">{channel.name}</h3>
                                    <span className="text-xs bg-blue-900/20 text-blue-400 px-2 py-1 rounded border border-blue-900/30 font-mono">
                                        {channel.stream}
                                    </span>
                                </div>
                                <div className="space-y-3">
                                    <div>
                                        <div className="text-xs text-zinc-500 uppercase tracking-wider mb-1">Subject</div>
                                        <code className="text-sm bg-zinc-950 px-2 py-1 rounded text-zinc-300 block w-full overflow-hidden text-ellipsis">
                                            {channel.subject}
                                        </code>
                                    </div>
                                    {channel.description && (
                                        <div>
                                            <div className="text-xs text-zinc-500 uppercase tracking-wider mb-1">Description</div>
                                            <p className="text-sm text-zinc-400">{channel.description}</p>
                                        </div>
                                    )}
                                    <div className="pt-2 border-t border-zinc-800 flex justify-between items-center text-xs text-zinc-500">
                                        <span>Created: {new Date(channel.created_at).toLocaleDateString()}</span>
                                    </div>
                                </div>
                            </div>
                        ))}
                    </div>
                )}
            </div>

            {/* Create Channel Modal */}
            {showCreateModal && (
                <div className="fixed inset-0 bg-black/50 backdrop-blur-sm flex items-center justify-center z-50">
                    <div className="bg-zinc-900 border border-zinc-700 rounded-xl p-6 w-96 shadow-2xl">
                        <h3 className="text-lg font-bold text-zinc-100 mb-4">Create New Channel</h3>
                        <div className="space-y-4">
                            <div>
                                <label className="block text-sm font-medium mb-1 text-zinc-400">Name</label>
                                <input
                                    type="text"
                                    value={createForm.name}
                                    onChange={(e) => setCreateForm({ ...createForm, name: e.target.value })}
                                    className="w-full p-2 border border-zinc-700 rounded-lg bg-zinc-950 text-zinc-200 focus:ring-2 focus:ring-zinc-600 outline-none"
                                    placeholder="e.g., sensor-data"
                                />
                            </div>
                            <div>
                                <label className="block text-sm font-medium mb-1 text-zinc-400">Subject</label>
                                <input
                                    type="text"
                                    value={createForm.subject}
                                    onChange={(e) => setCreateForm({ ...createForm, subject: e.target.value })}
                                    className="w-full p-2 border border-zinc-700 rounded-lg bg-zinc-950 text-zinc-200 focus:ring-2 focus:ring-zinc-600 outline-none"
                                    placeholder="e.g., sensors.>"
                                />
                                <p className="text-xs text-zinc-500 mt-1">
                                    NATS subject pattern (supports wildcards * and &gt;).
                                </p>
                            </div>
                            <div>
                                <label className="block text-sm font-medium mb-1 text-zinc-400">Description</label>
                                <textarea
                                    value={createForm.description}
                                    onChange={(e) => setCreateForm({ ...createForm, description: e.target.value })}
                                    className="w-full p-2 border border-zinc-700 rounded-lg bg-zinc-950 text-zinc-200 focus:ring-2 focus:ring-zinc-600 outline-none"
                                    rows={3}
                                    placeholder="Optional description..."
                                />
                            </div>
                        </div>
                        <div className="flex justify-end gap-2 mt-6">
                            <button
                                onClick={() => setShowCreateModal(false)}
                                className="px-4 py-2 text-zinc-400 hover:text-zinc-200 transition-colors"
                            >
                                Cancel
                            </button>
                            <button
                                onClick={handleCreateChannel}
                                disabled={!createForm.name || !createForm.subject}
                                className="bg-zinc-100 text-zinc-900 px-4 py-2 rounded-lg hover:bg-white transition-colors font-semibold disabled:opacity-50 disabled:cursor-not-allowed"
                            >
                                Create Channel
                            </button>
                        </div>
                    </div>
                </div>
            )}
        </div>
    );
}
