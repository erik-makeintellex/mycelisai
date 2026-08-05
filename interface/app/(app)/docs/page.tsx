"use client";

import { Suspense, useCallback, useEffect, useRef, useState } from "react";
import { useSearchParams, useRouter } from "next/navigation";
import { AlertTriangle, ArrowLeft, BookOpen, Loader2 } from "lucide-react";
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
  const [mobileView, setMobileView] = useState<"list" | "article">(
    searchParams?.get("doc") ? "article" : "list",
  );
  const docRequestRef = useRef(0);

  const loadDoc = useCallback(
    (entry: DocEntry, openMobile = true) => {
      const requestId = docRequestRef.current + 1;
      docRequestRef.current = requestId;
      setActiveSlug(entry.slug);
      setDocLabel(entry.label);
      if (openMobile) setMobileView("article");
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
      .then((response) => {
        if (!response.ok) throw new Error(`HTTP ${response.status}`);
        return response.json();
      })
      .then((data: ManifestResponse) => {
        if (!Array.isArray(data.sections)) {
          throw new Error("manifest response did not include sections");
        }
        setSections(data.sections);
        const allDocs = data.sections.flatMap((section) => section.docs);
        const target = requestedSlug
          ? allDocs.find((doc) => doc.slug === requestedSlug) ?? allDocs[0]
          : allDocs[0];
        if (target) loadDoc(target, Boolean(requestedSlug));
      })
      .catch((err) => setError(`Failed to load doc manifest: ${err instanceof Error ? err.message : String(err)}`))
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
        <div className="flex min-h-0 min-w-0 flex-1 overflow-hidden">
          <DocsSidebar
            sections={sections}
            activeSlug={activeSlug}
            query={query}
            onSelect={loadDoc}
            onQueryChange={setQuery}
            mobileHidden={mobileView === "article"}
          />
          <DocContent
            content={content}
            error={error}
            loading={loadingContent}
            sections={sections}
            onSelectDoc={loadDoc}
            onShowDocs={() => setMobileView("list")}
            mobileHidden={mobileView === "list"}
          />
        </div>
      )}
    </div>
  );
}

function DocsHeader({ docLabel }: { docLabel: string }) {
  return (
    <div className="flex min-w-0 flex-shrink-0 items-center gap-3 border-b border-cortex-border bg-cortex-surface px-4 py-2.5">
      <BookOpen className="w-4 h-4 text-cortex-primary flex-shrink-0" />
      <h1 className="text-[11px] font-mono font-bold uppercase tracking-widest text-cortex-text-muted">
        Documentation and guidance
      </h1>
      {docLabel ? (
        <>
          <span className="text-cortex-border">.</span>
          <span className="min-w-0 truncate text-[11px] font-mono text-cortex-text-main">
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
  onShowDocs,
  mobileHidden,
}: {
  content: string;
  error: string | null;
  loading: boolean;
  sections: DocSection[];
  onSelectDoc: (entry: DocEntry) => void;
  onShowDocs: () => void;
  mobileHidden: boolean;
}) {
  return (
    <main
      data-testid="docs-article-pane"
      className={`${mobileHidden ? "hidden" : "flex"} min-h-0 min-w-0 flex-1 flex-col overflow-x-hidden overflow-y-auto px-4 py-4 md:flex md:px-8 md:py-6`}
    >
      <button
        type="button"
        onClick={onShowDocs}
        className="mb-2 inline-flex min-h-10 w-fit items-center gap-2 text-sm font-semibold text-cortex-primary md:hidden"
      >
        <ArrowLeft className="h-4 w-4" />
        All docs
      </button>
      {loading ? (
        <LoadingState label="Loading doc..." />
      ) : error ? (
        <div className="flex items-start gap-3 text-cortex-danger py-12 justify-center">
          <AlertTriangle className="w-4 h-4 flex-shrink-0 mt-0.5" />
          <span className="text-sm font-mono">{error}</span>
        </div>
      ) : content ? (
        <div className="w-full min-w-0 max-w-3xl">
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
    </main>
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
