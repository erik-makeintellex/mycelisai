import { Sparkles } from "lucide-react";
import type { ChatMessage } from "@/store/useCortexStore";

function lastMessage(messages: ChatMessage[], role: ChatMessage["role"]) {
  return [...messages].reverse().find((message) => message.role === role);
}

function lastSomaMessage(messages: ChatMessage[]) {
  return [...messages]
    .reverse()
    .find((message) => message.role !== "user" && message.role !== "system");
}

export function SomaCausalSummary({
  messages,
  fallbackAction = "Ready for your first Soma request",
  teams = ["Soma"],
  outputs = ["Conversation guidance"],
  updated = ["Soma thread"],
}: {
  messages: ChatMessage[];
  fallbackAction?: string;
  teams?: string[];
  outputs?: string[];
  updated?: string[];
}) {
  const latestUser = lastMessage(messages, "user");
  const latestSoma = lastSomaMessage(messages);
  const action = latestUser?.content?.replace(/^\[BROADCAST\]\s*/i, "").trim();
  const produced = latestSoma?.artifacts?.length
    ? latestSoma.artifacts.map((artifact) => artifact.title || artifact.type || "artifact")
    : outputs;
  const next = latestSoma ? "Review Soma's response or ask for the next action." : "Tell Soma what you want to accomplish.";

  return (
    <section className="rounded-2xl border border-cortex-primary/25 bg-cortex-primary/10 p-4">
      <div className="flex items-center gap-2">
        <Sparkles className="h-4 w-4 text-cortex-primary" />
        <p className="text-sm font-semibold text-cortex-text-main">Soma just did this</p>
      </div>
      <div className="mt-4 grid gap-3 lg:grid-cols-5">
        <Fact label="Understood" value={action || fallbackAction} />
        <Fact label="Coordinated" value={teams.join(", ")} />
        <Fact label="Produced" value={produced.join(", ")} />
        <Fact label="Updated" value={updated.join(", ")} />
        <Fact label="Next" value={next} />
      </div>
    </section>
  );
}

function Fact({ label, value }: { label: string; value: string }) {
  return (
    <div className="rounded-xl border border-cortex-border bg-cortex-bg px-3 py-2">
      <p className="text-[10px] font-mono uppercase tracking-[0.16em] text-cortex-text-muted">
        {label}
      </p>
      <p className="mt-2 text-sm leading-5 text-cortex-text-main">{value}</p>
    </div>
  );
}
