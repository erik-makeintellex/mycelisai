'use client';

interface EditTeamForm {
    name: string;
    description: string;
    channels: string;
    resource_access: string;
}

interface EditTeamModalProps {
    isOpen: boolean;
    onClose: () => void;
    onSave: () => void;
    form: EditTeamForm;
    onChange: (form: EditTeamForm) => void;
}

export default function EditTeamModal({
    isOpen,
    onClose,
    onSave,
    form,
    onChange
}: EditTeamModalProps) {
    if (!isOpen) return null;

    return (
        <div className="fixed inset-0 bg-black/50 backdrop-blur-sm flex items-center justify-center z-50">
            <div className="bg-zinc-900 border border-zinc-700 rounded-xl p-6 w-96 shadow-2xl">
                <h3 className="text-lg font-bold text-zinc-100 mb-4">Edit Team</h3>
                <div className="space-y-4">
                    <div>
                        <label className="block text-sm font-medium mb-1 text-zinc-400">Name</label>
                        <input
                            type="text"
                            value={form.name}
                            onChange={(e) => onChange({ ...form, name: e.target.value })}
                            className="w-full p-2 border border-zinc-700 rounded-lg bg-zinc-950 text-zinc-200 focus:ring-2 focus:ring-zinc-600 outline-none"
                        />
                    </div>
                    <div>
                        <label className="block text-sm font-medium mb-1 text-zinc-400">Description</label>
                        <textarea
                            value={form.description}
                            onChange={(e) => onChange({ ...form, description: e.target.value })}
                            className="w-full p-2 border border-zinc-700 rounded-lg bg-zinc-950 text-zinc-200 focus:ring-2 focus:ring-zinc-600 outline-none"
                            rows={3}
                        />
                    </div>
                    <div>
                        <label className="block text-sm font-medium mb-1 text-zinc-400">Channels (comma separated)</label>
                        <input
                            type="text"
                            value={form.channels}
                            onChange={(e) => onChange({ ...form, channels: e.target.value })}
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
                            value={form.resource_access}
                            onChange={(e) => onChange({ ...form, resource_access: e.target.value })}
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
                        onClick={onClose}
                        className="px-4 py-2 text-zinc-400 hover:text-zinc-200 transition-colors"
                    >
                        Cancel
                    </button>
                    <button
                        onClick={onSave}
                        className="bg-zinc-100 text-zinc-900 px-4 py-2 rounded-lg hover:bg-white transition-colors font-semibold"
                    >
                        Save Changes
                    </button>
                </div>
            </div>
        </div>
    );
}
