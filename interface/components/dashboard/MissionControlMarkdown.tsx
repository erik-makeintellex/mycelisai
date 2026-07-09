"use client";

import ReactMarkdown from "react-markdown";
import remarkGfm from "remark-gfm";
import { ExternalLink } from "lucide-react";

export default function MissionControlMarkdown({ content }: { content: string }) {
    return (
        <ReactMarkdown
            remarkPlugins={[remarkGfm]}
            components={{
                a: ({ href, children }) => (
                    <a
                        href={href}
                        target="_blank"
                        rel="noopener noreferrer"
                        className="text-cortex-primary hover:text-cortex-primary/80 underline underline-offset-2 decoration-cortex-primary/30 hover:decoration-cortex-primary/60 transition-colors inline-flex items-center gap-0.5"
                    >
                        {children}
                        <ExternalLink className="w-2.5 h-2.5 inline-block flex-shrink-0" />
                    </a>
                ),
                pre: ({ children }) => (
                    <div className="relative group my-2">
                        <pre className="bg-cortex-bg border border-cortex-border rounded-lg p-3 overflow-x-auto text-[11px] leading-relaxed max-h-64 overflow-y-auto">
                            {children}
                        </pre>
                    </div>
                ),
                code: ({ children, className }) => {
                    const isBlock = className?.includes("language-");
                    if (isBlock) {
                        return <code className="font-mono">{children}</code>;
                    }
                    return (
                        <code className="font-mono text-[11px] px-1 py-0.5 rounded bg-cortex-bg border border-cortex-border text-cortex-primary">
                            {children}
                        </code>
                    );
                },
                h1: ({ children }) => (
                    <h1 className="text-sm font-bold text-cortex-text-main mt-3 mb-1 first:mt-0">{children}</h1>
                ),
                h2: ({ children }) => (
                    <h2 className="text-xs font-bold text-cortex-text-main mt-2.5 mb-1 first:mt-0">{children}</h2>
                ),
                h3: ({ children }) => (
                    <h3 className="text-xs font-bold text-cortex-text-muted mt-2 mb-0.5 first:mt-0">{children}</h3>
                ),
                p: ({ children }) => <p className="mb-1.5 last:mb-0">{children}</p>,
                ul: ({ children }) => <ul className="list-disc list-inside space-y-0.5 mb-1.5 ml-1">{children}</ul>,
                ol: ({ children }) => <ol className="list-decimal list-inside space-y-0.5 mb-1.5 ml-1">{children}</ol>,
                li: ({ children }) => <li className="text-cortex-text-main">{children}</li>,
                table: ({ children }) => (
                    <div className="my-2 overflow-x-auto rounded-lg border border-cortex-border bg-cortex-bg/70 shadow-sm">
                        <table className="w-full min-w-[360px] border-collapse text-left text-[12px] leading-5">
                            {children}
                        </table>
                    </div>
                ),
                thead: ({ children }) => (
                    <thead className="border-b border-cortex-border bg-cortex-surface/80">{children}</thead>
                ),
                th: ({ children }) => (
                    <th className="whitespace-nowrap px-3 py-2 text-left text-[10px] font-bold uppercase tracking-[0.14em] text-cortex-text-muted">{children}</th>
                ),
                td: ({ children }) => (
                    <td className="border-t border-cortex-border/50 px-3 py-2 align-top text-cortex-text-main">{children}</td>
                ),
                blockquote: ({ children }) => (
                    <blockquote className="border-l-2 border-cortex-primary/40 pl-3 my-1.5 text-cortex-text-muted italic">
                        {children}
                    </blockquote>
                ),
                hr: () => <hr className="border-cortex-border my-2" />,
                strong: ({ children }) => <strong className="font-bold text-cortex-text-main">{children}</strong>,
                em: ({ children }) => <em className="italic text-cortex-text-main/80">{children}</em>,
                img: ({ src, alt }) => (
                    <span className="block my-2">
                        {/* eslint-disable-next-line @next/next/no-img-element */}
                        <img
                            src={src}
                            alt={alt || ""}
                            className="max-w-full max-h-80 rounded-lg border border-cortex-border object-contain"
                        />
                    </span>
                ),
            }}
        >
            {content}
        </ReactMarkdown>
    );
}
