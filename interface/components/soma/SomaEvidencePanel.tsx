import Link from "next/link";
import type { ReactNode } from "react";

export type SomaEvidenceItem = {
  title: string;
  detail: string;
  href?: string;
  icon?: ReactNode;
};

export function SomaEvidencePanel({
  title = "Evidence of Soma's work",
  items,
}: {
  title?: string;
  items: SomaEvidenceItem[];
}) {
  return (
    <section className="rounded-3xl border border-cortex-border bg-cortex-surface p-4">
      <p className="text-[11px] font-mono uppercase tracking-[0.2em] text-cortex-primary">
        Evidence
      </p>
      <h2 className="mt-1 text-base font-semibold text-cortex-text-main">{title}</h2>
      <div className="mt-3 grid gap-2 sm:grid-cols-2 lg:grid-cols-4">
        {items.map((item) => {
          const content = (
            <span className="grid grid-cols-[auto_1fr] gap-2.5 rounded-2xl border border-cortex-border bg-cortex-bg px-3.5 py-2.5 transition-colors hover:border-cortex-primary/30 hover:bg-cortex-surface">
              <span className="mt-0.5 text-cortex-primary">{item.icon}</span>
              <span>
                <span className="block text-xs font-semibold text-cortex-text-main">
                  {item.title}
                </span>
                <span className="mt-0.5 block text-xs leading-4 text-cortex-text-muted line-clamp-2">
                  {item.detail}
                </span>
              </span>
            </span>
          );
          return item.href ? (
            <Link key={item.title} href={item.href}>
              {content}
            </Link>
          ) : (
            <div key={item.title}>{content}</div>
          );
        })}
      </div>
    </section>
  );
}
