import React from 'react';
import { ArtifactContent } from '@/lib/types/protocol';
import { FileCode, FileText, FileImage, ExternalLink } from 'lucide-react';

export function ArtifactCard({ content }: { content: ArtifactContent }) {
    const getIcon = (mime: string) => {
        if (mime.includes('image')) return FileImage;
        if (mime.includes('json') || mime.includes('javascript')) return FileCode;
        return FileText;
    };

    const Icon = getIcon(content.mime_type);

    return (
        <div className="group border border-zinc-200 rounded-lg p-4 bg-white hover:border-zinc-300 hover:shadow-sm transition-all max-w-sm">
            <div className="flex items-start justify-between mb-3">
                <div className="p-2 bg-zinc-100 rounded-lg group-hover:bg-zinc-200 transition-colors">
                    <Icon className="w-5 h-5 text-zinc-600" />
                </div>
                <span className="text-[10px] font-mono text-zinc-400 bg-zinc-50 px-2 py-1 rounded">
                    {content.mime_type}
                </span>
            </div>

            <h4 className="text-sm font-bold text-zinc-900 mb-1 truncate" title={content.title}>
                {content.title}
            </h4>

            <p className="text-xs text-zinc-500 mb-3 line-clamp-2 min-h-[2.5em]">
                {content.preview}
            </p>

            <button className="w-full text-xs font-medium text-zinc-600 flex items-center justify-center py-2 bg-zinc-50 hover:bg-zinc-100 rounded border border-zinc-100 transition-colors">
                <ExternalLink size={12} className="mr-1.5" />
                Open Artifact
            </button>
        </div>
    );
}
