"use client";

import { Check, Quote } from "lucide-react";
import type React from "react";

export type FactModel = { label: string; value: string; icon: React.ReactNode; quoteValue?: string };

function QuoteFactButton({ fact, copied, onQuote }: { fact: FactModel; copied: boolean; onQuote: () => void }) {
  if (!fact.quoteValue) return null;
  return (
    <button
      type="button"
      onClick={onQuote}
      className="inline-flex h-5 w-5 shrink-0 items-center justify-center rounded border border-cortex-border/70 text-cortex-text-muted transition-colors hover:border-cortex-info/40 hover:bg-cortex-info/10 hover:text-cortex-info"
      title={copied ? "Copied output quote" : "Copy output quote"}
      aria-label={copied ? "Copied output quote" : `Copy output quote for ${fact.value}`}
    >
      {copied ? <Check className="h-3 w-3" /> : <Quote className="h-3 w-3" />}
    </button>
  );
}

export function Fact({
  fact,
  copied,
  onQuote,
  compact = false,
}: {
  fact: FactModel;
  copied: boolean;
  onQuote: () => void;
  compact?: boolean;
}) {
  const shellClass = compact
    ? "min-w-0 rounded-lg border border-cortex-border bg-cortex-bg/80 px-3 py-2"
    : "min-h-[82px] min-w-0 overflow-hidden rounded-lg border border-cortex-border bg-cortex-bg px-3 py-2";
  const textClass = compact
    ? "mt-1 overflow-hidden text-xs leading-4 text-cortex-text-main [display:-webkit-box] [-webkit-box-orient:vertical] [-webkit-line-clamp:1]"
    : "mt-2 overflow-hidden text-sm leading-5 text-cortex-text-main [display:-webkit-box] [-webkit-box-orient:vertical] [-webkit-line-clamp:2]";

  return (
    <div className={shellClass}>
      <div className="flex items-center justify-between gap-2 text-cortex-primary">
        <div className="flex min-w-0 items-center gap-1.5">
          {fact.icon}
          <p className="text-[10px] font-mono uppercase tracking-[0.16em] text-cortex-text-muted">{fact.label}</p>
        </div>
        <QuoteFactButton fact={fact} copied={copied} onQuote={onQuote} />
      </div>
      <p className={textClass} title={fact.value}>
        {fact.value}
      </p>
    </div>
  );
}

export function CompactFact({ fact, copied, onQuote }: { fact: FactModel; copied: boolean; onQuote: () => void }) {
  return (
    <div className="min-w-0 rounded border border-cortex-border bg-cortex-bg/70 px-3 py-2">
      <div className="flex items-center justify-between gap-2 text-cortex-primary">
        <div className="flex min-w-0 items-center gap-1.5">
          {fact.icon}
          <p className="text-[10px] font-mono uppercase tracking-[0.14em] text-cortex-text-muted">{fact.label}</p>
        </div>
        <QuoteFactButton fact={fact} copied={copied} onQuote={onQuote} />
      </div>
      <p
        className="mt-1 overflow-hidden text-xs leading-4 text-cortex-text-main [display:-webkit-box] [-webkit-box-orient:vertical] [-webkit-line-clamp:2]"
        title={fact.value}
      >
        {fact.value}
      </p>
    </div>
  );
}
