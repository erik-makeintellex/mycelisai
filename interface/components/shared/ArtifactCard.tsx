import React from 'react';
import { ArtifactContent } from '@/lib/types/protocol';

export const ArtifactCard: React.FC<{ data: ArtifactContent }> = ({ data }) => {
    return (
        <div className="bg-white border border-zinc-200 rounded p-3 my-2 hover:shadow-sm transition-shadow">
            <div className="flex items-center space-x-2 mb-2">
                <div className="w-8 h-8 bg-zinc-100 rounded flex items-center justify-center">
                    <span className="text-xs font-mono">DOC</span>
                </div>
                <div className="overflow-hidden">
                    <h4 className="text-sm font-medium truncate" title={data.title}>{data.title}</h4>
                    <p className="text-[10px] text-zinc-400 font-mono">{data.mime_type}</p>
                </div>
            </div>
            <div className="bg-zinc-50 p-2 rounded text-xs font-mono text-zinc-600 mb-2 truncate">
                {data.preview}
            </div>
            <a href={data.uri} className="block text-center text-xs bg-zinc-900 text-white py-1 rounded hover:bg-zinc-700 transition-colors">
                Download
            </a>
        </div>
    );
};
