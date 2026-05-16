"use client";

export type ExecutionSummaryOutputPreview = {
    text: string;
    url: string | null;
};

function mediaPreviewKind(url: string | null): "image" | "audio" | "video" | null {
    if (!url) return null;
    try {
        const parsed = new URL(url, "http://mycelis.local");
        const workspacePath = parsed.searchParams.get("path");
        const candidate = (workspacePath || parsed.pathname).toLowerCase();
        if (/\.(png|jpe?g|gif|webp|avif|svg)$/.test(candidate)) return "image";
        if (/\.(mp3|wav|ogg|m4a|flac)$/.test(candidate)) return "audio";
        if (/\.(mp4|webm|mov|m4v)$/.test(candidate)) return "video";
    } catch {
        const lower = url.toLowerCase();
        if (/\.(png|jpe?g|gif|webp|avif|svg)(\?|#|$)/.test(lower)) return "image";
        if (/\.(mp3|wav|ogg|m4a|flac)(\?|#|$)/.test(lower)) return "audio";
        if (/\.(mp4|webm|mov|m4v)(\?|#|$)/.test(lower)) return "video";
    }
    return null;
}

export default function ExecutionSummaryMediaPreview({ outputs }: { outputs: ExecutionSummaryOutputPreview[] }) {
    const previews = outputs
        .map((output, index) => ({ output, index, kind: mediaPreviewKind(output.url) }))
        .filter((item): item is { output: ExecutionSummaryOutputPreview; index: number; kind: "image" | "audio" | "video" } => Boolean(item.kind && item.output.url));

    if (!previews.length) return null;

    return (
        <div className="grid gap-2 sm:grid-cols-2">
            {previews.map(({ output, index, kind }) => {
                const key = `${output.text}-${output.url}-${kind}-${index}`;
                return (
                    <div key={key} className="overflow-hidden rounded border border-cortex-border/70 bg-cortex-bg">
                        {kind === "image" ? (
                            <img src={output.url!} alt={output.text} className="max-h-52 w-full object-contain p-1" />
                        ) : kind === "audio" ? (
                            <audio controls className="w-full px-2 py-2" src={output.url!}>Your browser does not support audio playback.</audio>
                        ) : (
                            <video controls className="max-h-52 w-full bg-black" src={output.url!}>Your browser does not support video playback.</video>
                        )}
                    </div>
                );
            })}
        </div>
    );
}
