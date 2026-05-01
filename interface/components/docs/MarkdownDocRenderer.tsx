"use client";

import ReactMarkdown from "react-markdown";
import remarkGfm from "remark-gfm";
import { ExternalLink } from "lucide-react";
import type { DocEntry, DocSection } from "@/lib/docsManifest";

function resolveDocLink(
  href: string | undefined,
  sections: DocSection[],
): DocEntry | null {
  if (!href || href.startsWith("http") || href.startsWith("#")) return null;
  const bare = href.split("?")[0].split("#")[0];
  if (!bare.endsWith(".md")) return null;
  const filename = bare.split("/").pop()!.toLowerCase();
  const docs = sections.flatMap((section) => section.docs);
  return (
    docs.find((doc) => doc.path.split("/").pop()!.toLowerCase() === filename) ??
    null
  );
}

export function MarkdownDocRenderer({
  content,
  sections,
  onSelectDoc,
}: {
  content: string;
  sections: DocSection[];
  onSelectDoc: (entry: DocEntry) => void;
}) {
  return (
    <ReactMarkdown
      remarkPlugins={[remarkGfm]}
      components={{
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
          <ul className="list-disc list-inside space-y-1 mb-3 pl-2">
            {children}
          </ul>
        ),
        ol: ({ children }) => (
          <ol className="list-decimal list-inside space-y-1 mb-3 pl-2">
            {children}
          </ol>
        ),
        li: ({ children }) => (
          <li className="text-sm text-cortex-text-main font-mono leading-relaxed">
            {children}
          </li>
        ),
        code: ({ children, className }) =>
          className?.startsWith("language-") ? (
            <code className="block text-[11px] font-mono text-cortex-text-main leading-relaxed">
              {children}
            </code>
          ) : (
            <code className="text-[11px] font-mono text-cortex-primary bg-cortex-primary/10 px-1 py-0.5 rounded">
              {children}
            </code>
          ),
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
                type="button"
                onClick={() => onSelectDoc(internalEntry)}
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
        thead: ({ children }) => <thead className="bg-cortex-bg/80">{children}</thead>,
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
        em: ({ children }) => <em className="italic text-cortex-text-muted">{children}</em>,
        hr: () => <hr className="border-cortex-border my-4" />,
      }}
    >
      {content}
    </ReactMarkdown>
  );
}
