"use client";

import { Suspense, useCallback, useEffect, useRef, useState } from "react";
import { useSearchParams, useRouter } from "next/navigation";
import { AlertTriangle, BookOpen, Loader2 } from "lucide-react";
import { DocsSidebar } from "@/components/docs/DocsSidebar";
import { MarkdownDocRenderer } from "@/components/docs/MarkdownDocRenderer";
import type { DocEntry, DocSection } from "@/lib/docsManifest";

type ManifestResponse = {
  sections: DocSection[];
};

type DocResponse = {
  slug: string;
  label: string;
  content: string;
};

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
  const [content, setContent] = useState("");
  const [docLabel, setDocLabel] = useState("");
  const [loadingManifest, setLoadingManifest] = useState(true);
  const [loadingContent, setLoadingContent] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [query, setQuery] = useState("");
  const docRequestRef = useRef(0);

  const loadDoc = useCallback(
    (entry: DocEntry) => {
      const requestId = docRequestRef.current + 1;
      docRequestRef.current = requestId;
      setActiveSlug(entry.slug);
      setDocLabel(entry.label);
      router.replace(`/docs?doc=${entry.slug}`, { scroll: false });
      setLoadingContent(true);
      setError(null);

      fetch(`/docs-api/${entry.slug}`)
        .then((response) => {
          if (!response.ok) throw new Error(`HTTP ${response.status}`);
          return response.json();
        })
        .then((data: DocResponse) => {
          if (docRequestRef.current === requestId) setContent(data.content);
        })
        .catch((err) => {
          if (docRequestRef.current === requestId) {
            setError(`Failed to load "${entry.label}": ${err.message}`);
          }
        })
        .finally(() => {
          if (docRequestRef.current === requestId) setLoadingContent(false);
        });
    },
    [router],
  );

  useEffect(() => {
    const requestedSlug = searchParams?.get("doc") ?? null;
    fetch("/docs-api")
      .then((response) => response.json())
      .then((data: ManifestResponse) => {
        setSections(data.sections);
        const allDocs = data.sections.flatMap((section) => section.docs);
        const target = requestedSlug
          ? allDocs.find((doc) => doc.slug === requestedSlug) ?? allDocs[0]
          : allDocs[0];
        if (target) loadDoc(target);
      })
      .catch(() => setError("Failed to load doc manifest"))
      .finally(() => setLoadingManifest(false));
    // Load the initial manifest once; document clicks call loadDoc directly.
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, []);

  return (
    <div className="h-full flex flex-col bg-cortex-bg text-cortex-text-main overflow-hidden">
      <DocsHeader docLabel={docLabel} />
      {loadingManifest ? (
        <LoadingState label="Loading..." />
      ) : (
        <div className="flex-1 flex overflow-hidden">
          <DocsSidebar
            sections={sections}
            activeSlug={activeSlug}
            query={query}
            onSelect={loadDoc}
            onQueryChange={setQuery}
          />
          <DocContent
            content={content}
            error={error}
            loading={loadingContent}
            sections={sections}
            onSelectDoc={loadDoc}
          />
        </div>
      )}
    </div>
  );
}

function DocsHeader({ docLabel }: { docLabel: string }) {
  return (
    <div className="flex items-center gap-3 px-4 py-2.5 border-b border-cortex-border bg-cortex-surface flex-shrink-0">
      <BookOpen className="w-4 h-4 text-cortex-primary flex-shrink-0" />
      <span className="text-[11px] font-mono font-bold uppercase tracking-widest text-cortex-text-muted">
        Documentation
      </span>
      {docLabel ? (
        <>
          <span className="text-cortex-border">.</span>
          <span className="text-[11px] font-mono text-cortex-text-main">
            {docLabel}
          </span>
        </>
      ) : null}
    </div>
  );
}

function DocContent({
  content,
  error,
  loading,
  sections,
  onSelectDoc,
}: {
  content: string;
  error: string | null;
  loading: boolean;
  sections: DocSection[];
  onSelectDoc: (entry: DocEntry) => void;
}) {
  return (
    <div className="flex-1 overflow-y-auto px-8 py-6">
      {loading ? (
        <LoadingState label="Loading doc..." />
      ) : error ? (
        <div className="flex items-start gap-3 text-cortex-danger py-12 justify-center">
          <AlertTriangle className="w-4 h-4 flex-shrink-0 mt-0.5" />
          <span className="text-sm font-mono">{error}</span>
        </div>
      ) : content ? (
        <div className="max-w-3xl">
          <MarkdownDocRenderer
            content={content}
            sections={sections}
            onSelectDoc={onSelectDoc}
          />
        </div>
      ) : (
        <div className="flex flex-col items-center justify-center py-20 gap-3 text-cortex-text-muted">
          <BookOpen className="w-10 h-10 opacity-20" />
          <p className="text-sm font-mono">Select a document from the sidebar</p>
        </div>
      )}
    </div>
  );
}

function LoadingState({ label }: { label: string }) {
  return (
    <div className="flex-1 flex items-center justify-center gap-2 text-cortex-text-muted">
      <Loader2 className="w-4 h-4 animate-spin" />
      <span className="text-sm font-mono">{label}</span>
    </div>
  );
}
