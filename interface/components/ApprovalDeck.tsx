"use client";

import { useEffect, useState } from "react";
import axios from "axios";
import { Shield, Check, X, RefreshCw, Clock } from "lucide-react";

interface ApprovalRequest {
    id: string;
    source: string;
    rule_id: string;
    intent: string;
    payload_summary: string;
    timestamp: string;
}

export default function ApprovalDeck() {
    const [requests, setRequests] = useState<ApprovalRequest[]>([]);
    const [loading, setLoading] = useState(false);

    const fetchApprovals = async () => {
        setLoading(true);
        try {
            const res = await axios.get("/api/admin/approvals");
            // If empty AND we are in dev/demo mode, maybe show mocks? 
            // For now, adhere to Strict Real Data, but keep mock logic commented out or available if user asks.
            setRequests(res.data || []);
        } catch (e) {
            console.error("Failed to fetch approvals", e);
        } finally {
            setLoading(false);
        }
    };

    const handleAction = async (id: string, action: "APPROVE" | "DENY") => {
        try {
            await axios.post(`/api/admin/approvals/${id}`, { action });
            fetchApprovals();
        } catch (e) {
            alert("Action failed");
        }
    };

    useEffect(() => {
        fetchApprovals();
        const interval = setInterval(fetchApprovals, 3000);
        return () => clearInterval(interval);
    }, []);

    if (requests.length === 0) {
        return (
            <div className="h-full w-full flex flex-col items-center justify-center border-2 border-dashed border-slate-800 rounded-xl bg-slate-900/50 text-slate-600">
                <Shield size={48} className="opacity-20 mb-4" />
                <p className="text-sm">No Active Governance Requests</p>
                <button onClick={fetchApprovals} className="mt-4 text-xs text-blue-500 hover:text-blue-400 flex items-center gap-2">
                    <RefreshCw size={12} className={loading ? "animate-spin" : ""} /> Refresh
                </button>
            </div>
        )
    }

    return (
        <div className="h-full overflow-y-auto grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4 gap-4 pr-2">
            {requests.map((req) => (
                <div key={req.id} className="bg-slate-800 border border-slate-700 p-5 rounded-xl shadow-lg hover:border-blue-500/50 hover:shadow-blue-900/20 transition-all group flex flex-col">

                    <div className="flex justify-between items-start mb-4">
                        <div className="flex flex-col">
                            <span className="text-[10px] font-bold text-blue-400 uppercase tracking-wider mb-1">{req.source}</span>
                            <span className="text-sm font-bold text-white leading-tight">{req.intent}</span>
                        </div>
                        <span className="text-[10px] text-slate-500 font-mono bg-slate-900 px-2 py-1 rounded border border-slate-800">
                            {req.timestamp}
                        </span>
                    </div>

                    <div className="flex-1 bg-slate-900/50 rounded border border-slate-800/50 p-3 mb-4 font-mono text-xs text-slate-300 break-all">
                        {req.payload_summary}
                    </div>

                    <div className="grid grid-cols-2 gap-3 mt-auto">
                        <button
                            onClick={() => handleAction(req.id, "DENY")}
                            className="bg-slate-700 hover:bg-red-900/30 hover:border-red-800/50 text-slate-300 hover:text-red-400 border border-transparent text-xs font-bold py-2.5 rounded-lg transition-all"
                        >
                            DENY
                        </button>
                        <button
                            onClick={() => handleAction(req.id, "APPROVE")}
                            className="bg-blue-600 hover:bg-blue-500 text-white shadow-lg shadow-blue-900/50 text-xs font-bold py-2.5 rounded-lg transition-all"
                        >
                            APPROVE
                        </button>
                    </div>

                </div>
            ))}
        </div>
    );
}
