"use client";

import type { OutputProofEnvelope } from "@/store/useCortexStore";

function shortHash(value?: string) {
  return value && value.length > 12 ? value.slice(0, 12) : value;
}

function proofStatusLabel(value?: string) {
  if (!value) return null;
  return value.replace(/[_-]+/g, " ");
}

export function OutputProofBadges({ proof, proofArtifactId }: { proof?: OutputProofEnvelope; proofArtifactId?: string }) {
  const pathStatus = proofStatusLabel(proof?.path_boundary_status);
  const readbackStatus = proofStatusLabel(proof?.readback_status);
  const checksum = shortHash(proof?.checksum);
  if (!proof && !proofArtifactId) return null;
  return (
    <span className="inline-flex max-w-full flex-wrap items-center gap-1 text-[10px] leading-4 text-cortex-text-muted">
      {pathStatus ? <span className="rounded border border-cortex-success/40 px-1 text-cortex-success">path {pathStatus}</span> : null}
      {readbackStatus ? <span className="rounded border border-cortex-success/40 px-1 text-cortex-success">readback {readbackStatus}</span> : null}
      {checksum ? <span className="rounded border border-cortex-border/70 px-1 font-mono">sha256 {checksum}</span> : null}
      {proofArtifactId && !checksum ? <span className="rounded border border-cortex-border/70 px-1 font-mono">proof {shortHash(proofArtifactId)}</span> : null}
    </span>
  );
}

export function OutputProofDetails({ proof, proofArtifactId }: { proof?: OutputProofEnvelope; proofArtifactId?: string }) {
  if (!proof && !proofArtifactId) return null;
  return (
    <details className="mt-2 text-xs text-cortex-text-muted">
      <summary className="cursor-pointer text-[10px] font-mono uppercase tracking-[0.16em]">Verification details</summary>
      <OutputProofBadges proof={proof} proofArtifactId={proofArtifactId} />
    </details>
  );
}
