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
    <section className="rounded-3xl border border-cortex-border bg-cortex-surface p-5">
      <p className="text-[11px] font-mono uppercase tracking-[0.2em] text-cortex-primary">
        Evidence
      </p>
      <h2 className="mt-2 text-base font-semibold text-cortex-text-main">{title}</h2>
      <div className="mt-4 divide-y divide-cortex-border overflow-hidden rounded-2xl border border-cortex-border bg-cortex-bg">
        {items.map((item) => {
          const content = (
            <span className="grid grid-cols-[auto_1fr] gap-3 px-4 py-3 transition-colors hover:bg-cortex-surface">
              <span className="mt-0.5 text-cortex-primary">{item.icon}</span>
              <span>
                <span className="block text-sm font-semibold text-cortex-text-main">
                  {item.title}
                </span>
                <span className="mt-1 block text-sm leading-6 text-cortex-text-muted">
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
