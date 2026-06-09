import type { TeamWorkItem } from "@/store/useCortexStore";

export function ReviewQueueSummary({ items }: { items: TeamWorkItem[] }) {
  const needsDecision = items.filter((item) =>
    item.state === "degraded" || item.state === "needs_operator",
  ).length;
  const outputReady = items.filter((item) => item.state === "output_ready").length;
  const stillWorking = items.filter((item) =>
    item.state === "new" ||
    item.state === "briefed" ||
    item.state === "queued" ||
    item.state === "running" ||
    item.state === "reviewing" ||
    item.state === "paused",
  ).length;
  const canClear = items.filter((item) =>
    item.interactions.some(
      (action) => action.action === "archive" && !action.disabled,
    ),
  ).length;
  const stats = [
    { label: "Needs decision", value: needsDecision, tone: "text-amber-300" },
    { label: "Ready output", value: outputReady, tone: "text-cortex-primary" },
    { label: "Still working", value: stillWorking, tone: "text-cortex-success" },
    { label: "Can clear", value: canClear, tone: "text-cortex-text-main" },
  ];

  return (
    <div
      className="mt-3 grid gap-2 sm:grid-cols-2 lg:grid-cols-4"
      aria-label="Review queue summary"
    >
      {stats.map((stat) => (
        <div
          key={stat.label}
          className="rounded-lg border border-cortex-border bg-cortex-bg px-3 py-2"
          aria-label={`${stat.label}: ${stat.value}`}
        >
          <p className={`font-mono text-base font-semibold ${stat.tone}`}>
            {stat.value}
          </p>
          <p className="mt-0.5 font-mono text-[10px] uppercase tracking-[0.14em] text-cortex-text-muted">
            {stat.label}
          </p>
        </div>
      ))}
    </div>
  );
}
