import Link from "next/link";
import { AlertTriangle, ExternalLink, Sparkles } from "lucide-react";
import type { ChatMessage } from "@/store/useCortexStore";

type CausalPhase = "idle" | "pending" | "failed" | "answered";

function latestIndex(messages: ChatMessage[], match: (message: ChatMessage) => boolean) {
  for (let i = messages.length - 1; i >= 0; i -= 1) {
    if (match(messages[i])) return i;
  }
  return -1;
}

function isFailure(message: ChatMessage) {
  return message.mode === "blocker" || (message.role === "system" && /error|unreachable|failed/i.test(message.content ?? ""));
}

/** Derive what actually happened on the latest Soma turn — never claim work that did not occur. */
export function deriveSomaCausalPhase(messages: ChatMessage[]): {
  phase: CausalPhase;
  user?: ChatMessage;
  reply?: ChatMessage;
} {
  const userIdx = latestIndex(messages, (m) => m.role === "user");
  const replyIdx = latestIndex(messages, (m) => m.role !== "user" && (m.role !== "system" || isFailure(m)));
  const user = userIdx >= 0 ? messages[userIdx] : undefined;
  const reply = replyIdx >= 0 ? messages[replyIdx] : undefined;
  if (!user && !reply) return { phase: "idle" };
  if (user && replyIdx < userIdx) return { phase: "pending", user };
  if (reply && isFailure(reply)) return { phase: "failed", user, reply };
  return { phase: "answered", user, reply };
}

function deriveLifecycleStageLabel(phase: CausalPhase, reply?: ChatMessage): string {
  if (phase === "pending") return "Stage 5 · Mission Running";
  if (phase === "failed") return "Stage 9 · Recovery & Next Actions";
  if (reply?.mode === "proposal" || reply?.proposal) {
    return reply.proposal_status === "executed" ? "Stage 8 · Proof Verified" : "Stage 3 · Approval Gate";
  }
  if (reply?.delivery_package || reply?.execution_summary || reply?.artifacts?.length) {
    return "Stage 8 · Package & Proof Verified";
  }
  return "Stage 1 · Direct Answer";
}

export function SomaCausalSummary({
  messages,
  teams = ["Soma"],
  updated = ["Soma thread"],
}: {
  messages: ChatMessage[];
  teams?: string[];
  updated?: string[];
}) {
  const { phase, user, reply } = deriveSomaCausalPhase(messages);
  const action = user?.content?.replace(/^\[BROADCAST\]\s*/i, "").trim() || "—";

  if (phase === "idle") {
    return (
      <section data-testid="soma-causal-summary" data-phase="idle" className="rounded-2xl border border-cortex-border bg-cortex-bg p-4">
        <SummaryTitle />
        <p className="mt-2 text-sm text-cortex-text-muted">
          Nothing yet. Soma has not handled a request in this session — tell Soma what you want to accomplish.
        </p>
      </section>
    );
  }

  const failed = phase === "failed";
  const stageLabel = deriveLifecycleStageLabel(phase, reply);
  const runId = reply?.run_id || reply?.delivery_package?.run_id || reply?.execution_summary?.run_id;
  const facts: Array<[string, string]> =
    phase === "pending"
      ? [["Understood", action], ["Status", "Waiting for Soma's response"], ["Produced", "Nothing yet"], ["Next", "Wait for the reply, or check Status if it stalls."]]
      : failed
        ? [
            ["Requested", action],
            ["Result", "Blocked — the request did not complete"],
            ["Produced", "Nothing"],
            ["Updated", "Nothing"],
            ["Next", "Resolve the blocker shown in the thread, then retry."],
          ]
        : [
            ["Understood", action],
            ["Coordinated", teams.join(", ")],
            ["Produced", reply?.artifacts?.length ? reply.artifacts.map((a) => a.title || a.type || "artifact").join(", ") : "Response in thread"],
            ["Updated", updated.join(", ")],
            ["Next", "Review Soma's response or ask for the next action."],
          ];

  return (
    <section
      data-testid="soma-causal-summary"
      data-phase={phase}
      className={`rounded-2xl border p-4 ${failed ? "border-cortex-danger/40 bg-cortex-danger/10" : "border-cortex-primary/25 bg-cortex-primary/10"}`}
    >
      <div className="flex flex-wrap items-center justify-between gap-2">
        <SummaryTitle failed={failed} />
        <div className="flex items-center gap-2">
          <span
            data-testid="soma-causal-stage-badge"
            className={`rounded-full border px-2.5 py-0.5 text-[11px] font-mono ${
              failed
                ? "border-cortex-danger/40 bg-cortex-danger/15 text-cortex-danger"
                : "border-cortex-primary/30 bg-cortex-bg text-cortex-text-main"
            }`}
          >
            {stageLabel}
          </span>
          {runId && (
            <Link
              href={`/runs/${runId}`}
              data-testid="soma-causal-run-link"
              className="inline-flex items-center gap-1 rounded-full border border-cortex-border bg-cortex-bg px-2.5 py-0.5 text-[11px] font-mono text-cortex-primary hover:border-cortex-primary/50"
            >
              Inspect Run {runId}
              <ExternalLink className="h-3 w-3" />
            </Link>
          )}
        </div>
      </div>
      <div className={`mt-3 grid gap-2.5 ${facts.length === 5 ? "lg:grid-cols-5" : "lg:grid-cols-4"}`}>
        {facts.map(([label, value]) => (
          <Fact key={label} label={label} value={value} />
        ))}
      </div>
      {failed && reply?.content && (
        <div
          data-testid="soma-causal-diagnostic"
          className="mt-3 rounded-xl border border-cortex-danger/30 bg-cortex-bg/90 px-3 py-2 text-xs text-cortex-text-muted"
        >
          <span className="font-mono uppercase tracking-wider text-cortex-danger">Diagnostic Review: </span>
          <span className="text-cortex-text-main">{reply.content}</span>
          <span className="ml-2 font-mono text-[11px] text-cortex-text-muted">
            (channels: memory.stream, mission_events)
          </span>
        </div>
      )}
    </section>
  );
}

function SummaryTitle({ failed = false }: { failed?: boolean }) {
  return (
    <div className="flex items-center gap-2">
      {failed ? <AlertTriangle className="h-4 w-4 text-cortex-danger" /> : <Sparkles className="h-4 w-4 text-cortex-primary" />}
      <p className="text-sm font-semibold text-cortex-text-main">
        {failed ? "Soma could not complete this" : "Soma just did this"}
      </p>
    </div>
  );
}

function Fact({ label, value }: { label: string; value: string }) {
  return (
    <div className="rounded-xl border border-cortex-border bg-cortex-bg px-2.5 py-2">
      <p className="text-[10px] font-mono uppercase tracking-[0.16em] text-cortex-text-muted">
        {label}
      </p>
      <p className="mt-1 text-xs leading-5 text-cortex-text-main line-clamp-2 break-words">{value}</p>
    </div>
  );
}
