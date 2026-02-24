"use client";

import React, { Suspense, useEffect, useState, useCallback, useMemo } from "react";
import { useSearchParams, useRouter } from "next/navigation";
import ReactMarkdown from "react-markdown";
import remarkGfm from "remark-gfm";
import {
    BookOpen,
    ChevronRight,
    Loader2,
    AlertTriangle,
    ExternalLink,
    Search,
    X,
} from "lucide-react";
import type { DocSection, DocEntry } from "@/lib/docsManifest";

// ── Types ─────────────────────────────────────────────────────

interface ManifestResponse {
    sections: DocSection[];
}

interface DocResponse {
    slug: string;
    label: string;
    content: string;
}

// ── Sidebar ───────────────────────────────────────────────────

function Sidebar({
    sections,
    activeSlug,
    query,
    onSelect,
    onQueryChange,
}: {
    sections: DocSection[];
    activeSlug: string | null;
    query: string;
    onSelect: (entry: DocEntry) => void;
    onQueryChange: (q: string) => void;
}) {
    const filtered: DocSection[] = query.trim()
        ? sections
            .map((s) => ({
                ...s,
                docs: s.docs.filter(
                    (d) =>
                        d.label.toLowerCase().includes(query.toLowerCase()) ||
                        (d.description ?? "").toLowerCase().includes(query.toLowerCase())
                ),
            }))
            .filter((s) => s.docs.length > 0)
        : sections;

    return (
        <div className="w-56 flex-shrink-0 border-r border-cortex-border flex flex-col overflow-hidden">
            {/* Search */}
            <div className="px-3 py-2.5 border-b border-cortex-border">
                <div className="flex items-center gap-2 bg-cortex-bg border border-cortex-border rounded px-2 py-1.5">
                    <Search className="w-3 h-3 text-cortex-text-muted flex-shrink-0" />
                    <input
                        type="text"
                        value={query}
                        onChange={(e) => onQueryChange(e.target.value)}
                        placeholder="Filter docs..."
                        className="flex-1 bg-transparent text-[11px] font-mono text-cortex-text-main placeholder:text-cortex-text-muted/50 outline-none"
                    />
                    {query && (
                        <button onClick={() => onQueryChange("")} className="text-cortex-text-muted hover:text-cortex-text-main">
                            <X className="w-3 h-3" />
                        </button>
                    )}
                </div>
            </div>

            {/* Nav */}
            <div className="flex-1 overflow-y-auto py-2">
                {filtered.map((section) => (
                    <div key={section.section} className="mb-3">
                        <div className="px-3 pb-1 text-[9px] font-mono font-bold uppercase tracking-widest text-cortex-text-muted/60">
                            {section.section}
                        </div>
                        {section.docs.map((doc) => {
                            const isActive = activeSlug === doc.slug;
                            return (
                                <button
                                    key={doc.slug}
                                    onClick={() => onSelect(doc)}
                                    title={doc.description}
                                    className={`w-full text-left flex items-center gap-2 px-3 py-1.5 transition-colors ${
                                        isActive
                                            ? "bg-cortex-primary/10 text-cortex-primary border-r-2 border-cortex-primary"
                                            : "text-cortex-text-muted hover:text-cortex-text-main hover:bg-cortex-bg/60"
                                    }`}
                                >
                                    <ChevronRight
                                        className={`w-3 h-3 flex-shrink-0 transition-transform ${isActive ? "rotate-90 text-cortex-primary" : ""}`}
                                    />
                                    <span className="text-[11px] font-mono truncate">{doc.label}</span>
                                </button>
                            );
                        })}
                    </div>
                ))}

                {filtered.length === 0 && (
                    <p className="px-3 py-4 text-[11px] font-mono text-cortex-text-muted/60 text-center">
                        No docs match "{query}"
                    </p>
                )}
            </div>
        </div>
    );
}

// ── Markdown link resolution ───────────────────────────────────
//
// When a doc contains a relative .md link (e.g. `[API](API_REFERENCE.md)`)
// we resolve it to the matching manifest entry so navigation stays in-app.
// Returns the DocEntry if found, null otherwise.

function resolveDocLink(href: string | undefined, sections: DocSection[]): DocEntry | null {
    if (!href || href.startsWith("http") || href.startsWith("#")) return null;
    // Normalise: strip query/hash, get the bare filename
    const bare = href.split("?")[0].split("#")[0];
    if (!bare.endsWith(".md")) return null;
    const filename = bare.split("/").pop()!.toLowerCase();
    const allDocs = sections.flatMap((s) => s.docs);
    return allDocs.find((d) => d.path.split("/").pop()!.toLowerCase() === filename) ?? null;
}

// ── Markdown Renderer (built inside component — needs loadDoc + sections) ──

function buildMdComponents(
    sections: DocSection[],
    loadDoc: (entry: DocEntry) => void
): React.ComponentProps<typeof ReactMarkdown>["components"] {
    return {
        h1: ({ children }) => (
            <h1 className="text-xl font-bold text-cortex-text-main font-mono mt-6 mb-3 pb-2 border-b border-cortex-border">
                {children}
            </h1>
        ),
        h2: ({ children }) => (
            <h2 className="text-base font-bold text-cortex-text-main font-mono mt-5 mb-2 pb-1 border-b border-cortex-border/50">
                {children}
            </h2>
        ),
        h3: ({ children }) => (
            <h3 className="text-sm font-bold text-cortex-text-main font-mono mt-4 mb-1.5">
                {children}
            </h3>
        ),
        h4: ({ children }) => (
            <h4 className="text-sm font-semibold text-cortex-text-muted font-mono mt-3 mb-1">
                {children}
            </h4>
        ),
        p: ({ children }) => (
            <p className="text-sm text-cortex-text-main font-mono leading-relaxed mb-3">
                {children}
            </p>
        ),
        ul: ({ children }) => (
            <ul className="list-disc list-inside space-y-1 mb-3 pl-2">{children}</ul>
        ),
        ol: ({ children }) => (
            <ol className="list-decimal list-inside space-y-1 mb-3 pl-2">{children}</ol>
        ),
        li: ({ children }) => (
            <li className="text-sm text-cortex-text-main font-mono leading-relaxed">{children}</li>
        ),
        code: ({ children, className }) => {
            const isBlock = className?.startsWith("language-");
            if (isBlock) {
                return (
                    <code className="block text-[11px] font-mono text-cortex-text-main leading-relaxed">
                        {children}
                    </code>
                );
            }
            return (
                <code className="text-[11px] font-mono text-cortex-primary bg-cortex-primary/10 px-1 py-0.5 rounded">
                    {children}
                </code>
            );
        },
        pre: ({ children }) => (
            <pre className="bg-cortex-bg border border-cortex-border rounded p-3 overflow-x-auto mb-3 text-[11px] font-mono text-cortex-text-main leading-relaxed">
                {children}
            </pre>
        ),
        blockquote: ({ children }) => (
            <blockquote className="border-l-4 border-cortex-primary/40 pl-3 my-3 text-cortex-text-muted">
                {children}
            </blockquote>
        ),
        a: ({ href, children }) => {
            const internalEntry = resolveDocLink(href, sections);
            if (internalEntry) {
                return (
                    <button
                        onClick={() => loadDoc(internalEntry)}
                        className="text-cortex-primary hover:underline inline-flex items-center gap-0.5 cursor-pointer"
                    >
                        {children}
                    </button>
                );
            }
            return (
                <a
                    href={href}
                    target="_blank"
                    rel="noopener noreferrer"
                    className="text-cortex-primary hover:underline inline-flex items-center gap-0.5"
                >
                    {children}
                    <ExternalLink className="w-2.5 h-2.5 opacity-60" />
                </a>
            );
        },
        table: ({ children }) => (
            <div className="overflow-x-auto mb-4">
                <table className="w-full text-[11px] font-mono border-collapse border border-cortex-border">
                    {children}
                </table>
            </div>
        ),
        thead: ({ children }) => (
            <thead className="bg-cortex-bg/80">{children}</thead>
        ),
        th: ({ children }) => (
            <th className="border border-cortex-border px-3 py-1.5 text-left text-cortex-text-muted font-bold uppercase tracking-wider text-[9px]">
                {children}
            </th>
        ),
        td: ({ children }) => (
            <td className="border border-cortex-border px-3 py-1.5 text-cortex-text-main">
                {children}
            </td>
        ),
        tr: ({ children }) => (
            <tr className="hover:bg-cortex-bg/40 transition-colors">{children}</tr>
        ),
        strong: ({ children }) => (
            <strong className="font-bold text-cortex-text-main">{children}</strong>
        ),
        em: ({ children }) => (
            <em className="italic text-cortex-text-muted">{children}</em>
        ),
        hr: () => <hr className="border-cortex-border my-4" />,
    };
}

// ── Main Page ─────────────────────────────────────────────────

export default function DocsPage() {
    return (
        <Suspense fallback={<div className="h-full bg-cortex-bg" />}>
            <DocsContent />
        </Suspense>
    );
}

function DocsContent() {
    const searchParams = useSearchParams();
    const router = useRouter();
    const [sections, setSections] = useState<DocSection[]>([]);
    const [activeSlug, setActiveSlug] = useState<string | null>(null);
    const [content, setContent] = useState<string>("");
    const [docLabel, setDocLabel] = useState<string>("");
    const [loadingManifest, setLoadingManifest] = useState(true);
    const [loadingContent, setLoadingContent] = useState(false);
    const [error, setError] = useState<string | null>(null);
    const [query, setQuery] = useState("");

    // Load manifest on mount — honour ?doc= query param for deep-links
    useEffect(() => {
        const requestedSlug = searchParams.get("doc");
        fetch("/docs-api")
            .then((r) => r.json())
            .then((data: ManifestResponse) => {
                setSections(data.sections);
                const allDocs = data.sections.flatMap((s: DocSection) => s.docs);
                // Prefer ?doc= param; fall back to first entry
                const target = requestedSlug
                    ? (allDocs.find((d: DocEntry) => d.slug === requestedSlug) ?? allDocs[0])
                    : allDocs[0];
                if (target) loadDoc(target);
            })
            .catch(() => setError("Failed to load doc manifest"))
            .finally(() => setLoadingManifest(false));
    // eslint-disable-next-line react-hooks/exhaustive-deps
    }, []);

    const loadDoc = useCallback((entry: DocEntry) => {
        setActiveSlug(entry.slug);
        setDocLabel(entry.label);
        // Update URL so the active doc is bookmarkable / shareable
        router.replace(`/docs?doc=${entry.slug}`, { scroll: false });
        setLoadingContent(true);
        setError(null);

        fetch(`/docs-api/${entry.slug}`)
            .then((r) => {
                if (!r.ok) throw new Error(`HTTP ${r.status}`);
                return r.json();
            })
            .then((data: DocResponse) => setContent(data.content))
            .catch((err) => setError(`Failed to load "${entry.label}": ${err.message}`))
            .finally(() => setLoadingContent(false));
    }, []);

    // Rebuild when sections load so the link resolver has the full manifest
    const mdComponents = useMemo(
        () => buildMdComponents(sections, loadDoc),
        [sections, loadDoc]
    );

    return (
        <div className="h-full flex flex-col bg-cortex-bg text-cortex-text-main overflow-hidden">
            {/* Header */}
            <div className="flex items-center gap-3 px-4 py-2.5 border-b border-cortex-border bg-cortex-surface flex-shrink-0">
                <BookOpen className="w-4 h-4 text-cortex-primary flex-shrink-0" />
                <span className="text-[11px] font-mono font-bold uppercase tracking-widest text-cortex-text-muted">
                    Documentation
                </span>
                {docLabel && (
                    <>
                        <span className="text-cortex-border">·</span>
                        <span className="text-[11px] font-mono text-cortex-text-main">{docLabel}</span>
                    </>
                )}
            </div>

            {loadingManifest ? (
                <div className="flex-1 flex items-center justify-center gap-2 text-cortex-text-muted">
                    <Loader2 className="w-4 h-4 animate-spin" />
                    <span className="text-sm font-mono">Loading...</span>
                </div>
            ) : (
                <div className="flex-1 flex overflow-hidden">
                    {/* Sidebar */}
                    <Sidebar
                        sections={sections}
                        activeSlug={activeSlug}
                        query={query}
                        onSelect={loadDoc}
                        onQueryChange={setQuery}
                    />

                    {/* Content */}
                    <div className="flex-1 overflow-y-auto px-8 py-6">
                        {loadingContent ? (
                            <div className="flex items-center gap-2 text-cortex-text-muted py-12 justify-center">
                                <Loader2 className="w-4 h-4 animate-spin" />
                                <span className="text-sm font-mono">Loading doc...</span>
                            </div>
                        ) : error ? (
                            <div className="flex items-start gap-3 text-cortex-danger py-12 justify-center">
                                <AlertTriangle className="w-4 h-4 flex-shrink-0 mt-0.5" />
                                <span className="text-sm font-mono">{error}</span>
                            </div>
                        ) : content ? (
                            <div className="max-w-3xl">
                                <ReactMarkdown
                                    remarkPlugins={[remarkGfm]}
                                    components={mdComponents}
                                >
                                    {content}
                                </ReactMarkdown>
                            </div>
                        ) : (
                            <div className="flex flex-col items-center justify-center py-20 gap-3 text-cortex-text-muted">
                                <BookOpen className="w-10 h-10 opacity-20" />
                                <p className="text-sm font-mono">Select a document from the sidebar</p>
                            </div>
                        )}
                    </div>
                </div>
            )}
        </div>
    );
}
