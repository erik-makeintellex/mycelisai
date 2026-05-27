import { ShieldCheck, Sparkles } from "lucide-react";

export function SomaHeader({
  organizationName,
  status = "Ready",
  activeMode,
  governancePosture = "Governed execution enabled",
}: {
  organizationName?: string | null;
  status?: string;
  activeMode?: string | null;
  governancePosture?: string;
}) {
  const scope = organizationName
    ? `Ready to coordinate work for ${organizationName}`
    : "Ready to help create or resume an AI Organization";

  return (
    <header className="flex flex-col gap-3 border-b border-cortex-border px-4 py-3 lg:flex-row lg:items-center lg:justify-between">
      <div className="min-w-0">
        <div className="inline-flex items-center gap-2 rounded-full border border-cortex-primary/25 bg-cortex-primary/10 px-3 py-1 text-[11px] font-mono uppercase tracking-[0.2em] text-cortex-primary">
          <Sparkles className="h-3.5 w-3.5" />
          Soma
        </div>
        <h1 className="mt-2 text-2xl font-semibold tracking-tight text-cortex-text-main">
          What do you want Soma to do?
        </h1>
        <p className="mt-1 truncate text-sm text-cortex-text-muted">
          {scope}
        </p>
      </div>
      <div className="flex flex-wrap items-center gap-2 rounded-xl border border-cortex-border bg-cortex-bg px-3 py-2 text-xs text-cortex-text-muted">
        <span className="font-medium text-cortex-text-main">{status}</span>
        {activeMode ? <span>Mode: {activeMode}</span> : null}
        <span className="flex items-center gap-2">
          <ShieldCheck className="h-4 w-4 text-cortex-primary" />
          {governancePosture}
        </span>
      </div>
    </header>
  );
}
