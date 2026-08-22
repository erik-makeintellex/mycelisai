"use client";

import React, { useMemo, useState } from "react";
import { CheckCircle2, FileCheck2, Loader2, MessageSquareText, ShieldAlert } from "lucide-react";
import {
  activateWorkerProfileDraftPrompt,
  requestSomaPromptHandoff,
  saveWorkerProfileDraftPrompt,
} from "@/components/soma/somaPromptHandoff";

type ConfigIssue = {
  code?: string;
  field?: string;
  message?: string;
};

type DryRunEffect = {
  action?: string;
  document_id?: string;
  document_version?: string;
  scope?: { kind?: string; ref?: string };
  risk_level?: string;
  approval_posture?: string;
};

type DryRunResult = {
  valid?: boolean;
  digest?: string;
  effect?: DryRunEffect;
  issues?: ConfigIssue[];
};

type PreviewEnvelope = {
  ok?: boolean;
  data?: {
    dry_run?: DryRunResult;
  };
  error?: string;
};

const starterProfile = `apiVersion: mycelis.ai/v1
kind: WorkerProfile
metadata:
  id: research-specialist
  name: Research Specialist
  version: 0.1.0
  owner_id: operator
  enabled: true
  scope:
    kind: workspace
    ref: soma-root
  source:
    kind: soma
    ref: resources-worker-profile-draft
  governance:
    risk_level: medium
    approval_posture: confirm
spec:
  description: Researches public web, approved local sources, and mounted data when the work contract allows it.
  role: research
  system_prompt: Summarize findings clearly, name source boundaries, and flag uncertainty before teams rely on the result.
  capability_refs:
    - web.research
    - local_sources.search
  context_bindings:
    - kind: workspace
      ref: approved-sources
      access: search
  usage_policy:
    selection: soma_or_manual
    scope: team
  outputs:
    - research_summary
    - source_list
  verification_strategy: semantic
  verification_rubric:
    - Names whether sources are public web, local, mounted, or private API.
    - Separates evidence from recommendation.`;

function trimDigest(digest?: string) {
  if (!digest) return "";
  return digest.length > 14 ? `${digest.slice(0, 10)}...${digest.slice(-4)}` : digest;
}

function effectSummary(effect?: DryRunEffect) {
  if (!effect) return "No activation effect available yet.";
  const scope = effect.scope?.kind
    ? `${effect.scope.kind}${effect.scope.ref ? `/${effect.scope.ref}` : ""}`
    : "unspecified scope";
  return `${effect.document_id ?? "profile"} ${effect.document_version ?? ""} would ${effect.action ?? "preview"} for ${scope}.`;
}

export default function WorkerProfileAuthoringPanel() {
  const [content, setContent] = useState(starterProfile);
  const [preview, setPreview] = useState<DryRunResult | null>(null);
  const [error, setError] = useState("");
  const [isPreviewing, setIsPreviewing] = useState(false);

  const hasPreviewed = preview !== null || Boolean(error);
  const canHandOff = content.trim().length > 0;
  const previewStatus = useMemo(() => {
    if (isPreviewing) return "Checking";
    if (preview?.valid) return "Valid preview";
    if (preview && !preview.valid) return "Needs edits";
    if (error) return "Preview blocked";
    return "Not previewed";
  }, [error, isPreviewing, preview]);

  async function previewDraft() {
    setIsPreviewing(true);
    setError("");
    setPreview(null);
    try {
      const response = await fetch("/api/v1/config-documents/dry-run", {
        method: "POST",
        headers: { "content-type": "application/json" },
        body: JSON.stringify({ content, format: "yaml" }),
      });
      const envelope = (await response.json()) as PreviewEnvelope;
      const dryRun = envelope.data?.dry_run;
      if (!response.ok || !envelope.ok || !dryRun) {
        setError(envelope.error || "Soma could not preview this profile right now.");
        return;
      }
      setPreview(dryRun);
    } catch {
      setError("Configuration preview is unreachable. Check Core service health and try again.");
    } finally {
      setIsPreviewing(false);
    }
  }

  return (
    <section className="border-b border-cortex-border bg-cortex-surface/30 px-5 py-4">
      <div className="rounded-lg border border-cortex-border bg-cortex-panel p-4">
        <div className="flex flex-wrap items-start justify-between gap-3">
          <div className="max-w-2xl">
            <p className="text-[11px] font-bold uppercase tracking-[0.18em] text-cortex-success">Draft profile</p>
            <h3 className="mt-1 text-lg font-semibold text-cortex-text-main">Preview a teammate template before Soma saves it.</h3>
            <p className="mt-1 text-sm text-cortex-text-muted">
              Profiles are inert until a governed team uses them. Preview checks the exact YAML; Soma handles save and activation through chat.
            </p>
          </div>
          <span className="rounded-full border border-cortex-border bg-cortex-bg px-3 py-1 text-xs font-semibold text-cortex-text-muted">
            {previewStatus}
          </span>
        </div>

        <div className="mt-4 grid gap-4 xl:grid-cols-[minmax(0,1.35fr)_minmax(280px,0.65fr)]">
          <label className="block">
            <span className="sr-only">Worker Profile YAML</span>
            <textarea
              value={content}
              onChange={(event) => setContent(event.target.value)}
              spellCheck={false}
              className="min-h-72 w-full resize-y rounded-md border border-cortex-border bg-cortex-bg p-3 font-mono text-xs leading-5 text-cortex-text-main outline-none focus:border-cortex-success"
            />
          </label>

          <div className="flex flex-col gap-3">
            <div className="rounded-md border border-cortex-border bg-cortex-bg p-3">
              <div className="flex items-center gap-2 text-sm font-semibold text-cortex-text-main">
                {preview?.valid ? <CheckCircle2 className="h-4 w-4 text-cortex-success" /> : <FileCheck2 className="h-4 w-4 text-cortex-success" />}
                Preview result
              </div>
              {!hasPreviewed ? (
                <p className="mt-2 text-sm text-cortex-text-muted">Run preview to check identity, scope, risk, and field issues without saving anything.</p>
              ) : error ? (
                <p className="mt-2 text-sm text-cortex-danger">{error}</p>
              ) : preview?.valid ? (
                <div className="mt-2 space-y-2 text-sm text-cortex-text-muted">
                  <p className="font-medium text-cortex-success">{effectSummary(preview.effect)}</p>
                  <p>Risk: {preview.effect?.risk_level ?? "not declared"} · Approval: {preview.effect?.approval_posture ?? "not declared"}</p>
                  <details className="rounded-md border border-cortex-border px-3 py-2">
                    <summary className="cursor-pointer text-xs font-semibold uppercase tracking-[0.14em]">Digest</summary>
                    <p className="mt-2 break-all font-mono text-xs">{trimDigest(preview.digest)}</p>
                  </details>
                </div>
              ) : (
                <div className="mt-2 space-y-2">
                  {(preview?.issues ?? []).slice(0, 4).map((issue, index) => (
                    <p key={`${issue.code ?? "issue"}-${issue.field ?? index}`} className="text-sm text-cortex-warning">
                      {issue.field ? `${issue.field}: ` : ""}{issue.message ?? issue.code ?? "Needs correction"}
                    </p>
                  ))}
                  {(preview?.issues?.length ?? 0) === 0 && <p className="text-sm text-cortex-warning">The profile needs correction before it can be saved.</p>}
                </div>
              )}
            </div>

            <div className="grid gap-2">
              <button
                type="button"
                onClick={previewDraft}
                disabled={isPreviewing || !canHandOff}
                className="flex items-center justify-center gap-2 rounded-md border border-cortex-success/40 bg-cortex-success/10 px-3 py-2 text-sm font-semibold text-cortex-success hover:bg-cortex-success/20 disabled:cursor-not-allowed disabled:opacity-50"
              >
                {isPreviewing ? <Loader2 className="h-4 w-4 animate-spin" /> : <FileCheck2 className="h-4 w-4" />}
                Preview only
              </button>
              <button
                type="button"
                onClick={() => requestSomaPromptHandoff(saveWorkerProfileDraftPrompt(content))}
                disabled={!canHandOff}
                className="flex items-center justify-center gap-2 rounded-md border border-cortex-border bg-cortex-bg px-3 py-2 text-sm font-semibold text-cortex-text-main hover:border-cortex-success/50 disabled:cursor-not-allowed disabled:opacity-50"
              >
                <MessageSquareText className="h-4 w-4" />
                Ask Soma to save
              </button>
              <button
                type="button"
                onClick={() => requestSomaPromptHandoff(activateWorkerProfileDraftPrompt(content))}
                disabled={!canHandOff}
                className="flex items-center justify-center gap-2 rounded-md border border-cortex-warning/40 bg-cortex-warning/10 px-3 py-2 text-sm font-semibold text-cortex-warning hover:bg-cortex-warning/20 disabled:cursor-not-allowed disabled:opacity-50"
              >
                <ShieldAlert className="h-4 w-4" />
                Ask Soma to activate
              </button>
            </div>
          </div>
        </div>
      </div>
    </section>
  );
}
