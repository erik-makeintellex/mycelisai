"use client";

import { type FormEvent, useState } from "react";
import { Send } from "lucide-react";
import type { TeamWorkItem } from "@/store/useCortexStore";

export function TeamAskForm({
  item,
  onTeamAsk,
}: {
  item: TeamWorkItem;
  onTeamAsk?: (item: TeamWorkItem, message: string) => Promise<void> | void;
}) {
  const [isOpen, setIsOpen] = useState(false);
  const [message, setMessage] = useState("");
  const [isSubmitting, setIsSubmitting] = useState(false);
  if (!onTeamAsk || item.source !== "durable" || item.state === "archived") {
    return null;
  }

  const actionLabel = item.needsOperator ? "Respond" : "Ask team";
  if (!isOpen) {
    return (
      <button
        type="button"
        className="mt-3 inline-flex h-8 items-center justify-center gap-1 rounded-lg border border-cortex-primary/30 px-3 text-xs font-semibold text-cortex-primary hover:bg-cortex-primary/10"
        onClick={() => setIsOpen(true)}
      >
        <Send className="h-3.5 w-3.5" />
        <span>{actionLabel}</span>
      </button>
    );
  }

  const trimmed = message.trim();
  const submit = async (event: FormEvent<HTMLFormElement>) => {
    event.preventDefault();
    if (!trimmed || isSubmitting) return;
    setIsSubmitting(true);
    try {
      await onTeamAsk(item, trimmed);
      setMessage("");
      setIsOpen(false);
    } finally {
      setIsSubmitting(false);
    }
  };

  return (
    <form onSubmit={submit} className="mt-3 flex min-w-0 gap-2">
      <input
        value={message}
        onChange={(event) => setMessage(event.target.value)}
        className="min-w-0 flex-1 rounded-lg border border-cortex-border bg-cortex-surface px-3 py-2 text-sm text-cortex-text-main placeholder:text-cortex-text-muted focus:border-cortex-primary focus:outline-none"
        placeholder={
          item.needsOperator
            ? "Respond with the missing input"
            : "Ask this team for the next bounded output"
        }
        aria-label={`Ask ${item.title}`}
      />
      <button
        type="button"
        className="inline-flex h-10 items-center justify-center rounded-lg border border-cortex-border px-3 text-xs font-semibold text-cortex-text-muted hover:text-cortex-text-main"
        onClick={() => {
          setMessage("");
          setIsOpen(false);
        }}
      >
        Cancel
      </button>
      <button
        type="submit"
        disabled={!trimmed || isSubmitting}
        className="inline-flex h-10 items-center justify-center gap-1 rounded-lg border border-cortex-primary/30 px-3 text-xs font-semibold text-cortex-primary hover:bg-cortex-primary/10 disabled:cursor-not-allowed disabled:opacity-50"
      >
        <Send className="h-3.5 w-3.5" />
        <span>{isSubmitting ? "Queueing" : "Queue ask"}</span>
      </button>
    </form>
  );
}
