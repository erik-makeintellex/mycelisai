import { API_BASE_URL } from '@/config';

export default function ConfigPage() {
    return (
        <div className="space-y-8">
            <h1 className="text-3xl font-bold text-zinc-100">System Configuration</h1>
            <div className="p-6 border border-zinc-700 rounded-xl bg-zinc-800 shadow-lg">
                <p className="text-zinc-400">Configuration settings will be implemented here.</p>
                <div className="mt-4 space-y-2">
                    <p className="text-sm text-zinc-500">Current Environment: <span className="text-zinc-300">Development</span></p>
                    <p className="text-sm text-zinc-500">API Endpoint: <span className="text-zinc-300">{API_BASE_URL}</span></p>
                </div>
            </div>
        </div>
    );
}
