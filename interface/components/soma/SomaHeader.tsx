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
    <header className="flex flex-col gap-4 border-b border-cortex-border p-5 lg:flex-row lg:items-start lg:justify-between">
      <div>
        <div className="inline-flex items-center gap-2 rounded-full border border-cortex-primary/25 bg-cortex-primary/10 px-3 py-1 text-[11px] font-mono uppercase tracking-[0.2em] text-cortex-primary">
          <Sparkles className="h-3.5 w-3.5" />
          Soma
        </div>
        <h1 className="mt-3 text-3xl font-semibold tracking-tight text-cortex-text-main">
          What do you want Soma to do?
        </h1>
        <p className="mt-2 max-w-3xl text-sm leading-7 text-cortex-text-muted">
          {scope}
        </p>
      </div>
      <div className="rounded-2xl border border-cortex-border bg-cortex-bg px-4 py-3 text-sm text-cortex-text-muted lg:max-w-sm">
        <p className="font-medium text-cortex-text-main">{status}</p>
        {activeMode ? <p className="mt-1">Mode: {activeMode}</p> : null}
        <p className="mt-1 flex items-center gap-2">
          <ShieldCheck className="h-4 w-4 text-cortex-primary" />
          {governancePosture}
        </p>
      </div>
    </header>
  );
}
