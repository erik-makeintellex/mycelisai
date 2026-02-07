"use client";
import React, { useState } from 'react';
import { AlertCircle, Check, X } from 'lucide-react';

export function DecisionOverlay() {
    // Mock State - Real version connects to Governance Store
    const [activeRequest, setActiveRequest] = useState<null | { id: string, desc: string }>(null);

    if (!activeRequest) return null;

    return (
        <div className="fixed inset-0 bg-black/50 backdrop-blur-sm z-[9999] flex items-center justify-center p-4">
            <div className="bg-white rounded-xl shadow-2xl max-w-md w-full overflow-hidden border border-red-100">
                {/* Header */}
                <div className="bg-red-50 p-4 flex items-center border-b border-red-100">
                    <div className="w-10 h-10 bg-red-100 rounded-full flex items-center justify-center mr-3">
                        <AlertCircle className="w-5 h-5 text-red-600" />
                    </div>
                    <div>
                        <h3 className="text-sm font-bold text-red-900 uppercase">Governance Intercept</h3>
                        <p className="text-xs text-red-700">Human Approval Required</p>
                    </div>
                </div>

                {/* Body */}
                <div className="p-6">
                    <p className="text-zinc-700 font-medium mb-2">Request ID: {activeRequest.id}</p>
                    <p className="text-sm text-zinc-500 mb-6 bg-zinc-50 p-3 rounded border border-zinc-100 font-mono">
                        {activeRequest.desc}
                    </p>

                    <div className="flex gap-3">
                        <button
                            onClick={() => setActiveRequest(null)}
                            className="flex-1 flex items-center justify-center py-2.5 bg-white border border-zinc-300 text-zinc-700 rounded-lg hover:bg-zinc-50 font-medium transition-colors"
                        >
                            <X className="w-4 h-4 mr-2" />
                            Deny
                        </button>
                        <button
                            onClick={() => setActiveRequest(null)}
                            className="flex-1 flex items-center justify-center py-2.5 bg-indigo-600 text-white rounded-lg hover:bg-indigo-700 font-medium transition-colors shadow-sm"
                        >
                            <Check className="w-4 h-4 mr-2" />
                            Approve
                        </button>
                    </div>
                </div>
            </div>
        </div>
    );
}
