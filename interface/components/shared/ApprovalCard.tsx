import React from 'react';
import { GovernanceContent } from '@/lib/types/protocol';

export const ApprovalCard: React.FC<{ data: GovernanceContent }> = ({ data }) => {
    return (
        <div className="bg-white border-l-4 border-rose-500 rounded shadow-sm p-4 my-2">
            <div className="flex justify-between items-start mb-2">
                <h3 className="text-sm font-bold text-zinc-900">Permission Requested</h3>
                <span className="text-xs bg-rose-100 text-rose-700 px-2 py-0.5 rounded-full uppercase tracking-wider">{data.status}</span>
            </div>
            <p className="text-sm text-zinc-600 mb-3">{data.description}</p>

            <div className="bg-zinc-50 p-2 rounded border border-zinc-200 mb-4 text-xs font-mono text-zinc-700">
                Operation: {data.action}
            </div>

            <div className="flex space-x-2">
                <button className="flex-1 bg-zinc-900 text-white text-xs font-medium py-2 rounded hover:bg-zinc-800">
                    Approve
                </button>
                <button className="flex-1 bg-white border border-zinc-300 text-zinc-700 text-xs font-medium py-2 rounded hover:bg-zinc-50">
                    Deny
                </button>
            </div>
        </div>
    );
};
