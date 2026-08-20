"use client";

import { useState, type FormEvent } from "react";
import { ShieldCheck } from "lucide-react";
import type { TeamWorkItem } from "@/store/useCortexStore";
import type {
  ExternalOutcomeResult,
  ExternalOutcomeVerification,
} from "./teamWorkActions";

const outcomeOptions: Array<{
  value: ExternalOutcomeResult;
  label: string;
  description: string;
}> = [
  {
    value: "committed",
    label: "Confirmed applied",
    description: "Use when the external system confirms the change completed. This does not run it again.",
  },
  {
    value: "not_committed",
    label: "Confirmed not applied",
    description: "Use only when a receipt or audit confirms no change was applied. A new attempt needs a new proposal.",
  },
  {
    value: "still_unknown",
    label: "Still unclear",
    description: "Keeps the outcome unknown and retry unavailable.",
  },
];

export function ExternalOutcomeVerificationForm({
  item,
  compact = false,
  onVerify,
}: {
  item: TeamWorkItem;
  compact?: boolean;
  onVerify?: (
    item: TeamWorkItem,
    verification: ExternalOutcomeVerification,
  ) => Promise<void> | void;
}) {
  const [result, setResult] = useState<ExternalOutcomeResult | "">("");
  const [summary, setSummary] = useState("");
  const [evidenceRefs, setEvidenceRefs] = useState("");
  const [submitting, setSubmitting] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [success, setSuccess] = useState<string | null>(null);
  const canSubmit = Boolean(result && summary.trim() && onVerify && !submitting);
  const selectedOutcome = outcomeOptions.find((option) => option.value === result);

  const submit = async (event: FormEvent<HTMLFormElement>) => {
    event.preventDefault();
    if (!result || !summary.trim() || !onVerify) return;
    setSubmitting(true);
    setError(null);
    setSuccess(null);
    try {
      await onVerify(item, {
        result,
        summary: summary.trim(),
        evidenceRefs: parseEvidenceRefs(evidenceRefs),
      });
      setResult("");
      setSummary("");
      setEvidenceRefs("");
      setSuccess("Verification recorded. Soma is updating the trusted work state; no retry was requested.");
    } catch (submissionError) {
      setError(submissionError instanceof Error
        ? submissionError.message
        : "Could not submit verification.");
    } finally {
      setSubmitting(false);
    }
  };

  return (
    <form
      className={compact
        ? "mt-4 border-t border-cortex-border pt-4"
        : "mt-5 border-t border-cortex-border pt-4"}
      aria-label={`Verify external result for ${item.title}`}
      onSubmit={submit}
    >
      <div className="flex flex-wrap items-baseline justify-between gap-x-3 gap-y-1">
        <h4 className="text-sm font-semibold text-cortex-text-main">Verify external result</h4>
        <p className="text-xs leading-5 text-cortex-text-muted">This records evidence only. It does not retry the action.</p>
      </div>

      <fieldset className="mt-3">
        <legend className="text-xs font-semibold text-cortex-text-main">Observed result</legend>
        <div className="mt-1.5 grid grid-cols-1 gap-1.5 sm:grid-cols-3">
          {outcomeOptions.map((option) => (
            <label
              key={option.value}
              className={`flex min-h-9 min-w-0 cursor-pointer items-center justify-center gap-2 rounded-lg border px-2.5 py-1.5 text-center transition-colors ${
                result === option.value
                  ? "border-cortex-primary/60 bg-cortex-primary/10"
                  : "border-cortex-border bg-cortex-surface hover:border-cortex-primary/35"
              }`}
            >
              <input
                type="radio"
                name={`external-outcome-${item.id}`}
                value={option.value}
                checked={result === option.value}
                onChange={() => {
                  setResult(option.value);
                  setSuccess(null);
                }}
                className="h-4 w-4 shrink-0 accent-cortex-primary"
              />
              <span className="min-w-0 text-xs font-semibold text-cortex-text-main">{option.label}</span>
            </label>
          ))}
        </div>
        <p className="mt-1.5 min-h-5 text-[11px] leading-5 text-cortex-text-muted" aria-live="polite">
          {selectedOutcome?.description ?? "Choose the result that matches what the external system shows."}
        </p>
      </fieldset>

      <div className="mt-2 grid items-end gap-2 lg:grid-cols-[minmax(0,1.4fr)_minmax(12rem,0.8fr)_auto]">
        <label className="block text-xs font-semibold text-cortex-text-main">
          What did you observe? <span className="text-amber-300">Required</span>
          <textarea
            value={summary}
            onChange={(event) => {
              setSummary(event.target.value);
              setSuccess(null);
            }}
            required
            rows={2}
            className="mt-1.5 w-full resize-y rounded-lg border border-cortex-border bg-cortex-surface px-3 py-2 text-sm font-normal leading-5 text-cortex-text-main outline-none focus:border-cortex-primary/60"
          />
        </label>
        <label className="block text-xs font-semibold text-cortex-text-main">
          Evidence references <span className="font-normal text-cortex-text-muted">Optional</span>
          <input
            type="text"
            value={evidenceRefs}
            onChange={(event) => {
              setEvidenceRefs(event.target.value);
              setSuccess(null);
            }}
            placeholder="Receipt URL or evidence ID"
            className="mt-1.5 h-[3.875rem] w-full rounded-lg border border-cortex-border bg-cortex-surface px-3 py-2 text-sm font-normal leading-5 text-cortex-text-main outline-none focus:border-cortex-primary/60"
          />
        </label>
        <button
          type="submit"
          disabled={!canSubmit}
          className="inline-flex h-[3.875rem] items-center justify-center gap-2 rounded-lg border border-cortex-primary/40 bg-cortex-primary/15 px-3 py-2 text-sm font-semibold text-cortex-primary hover:border-cortex-primary/70 disabled:cursor-not-allowed disabled:opacity-50"
        >
          <ShieldCheck className="h-4 w-4" />
          {submitting ? "Submitting..." : "Submit verification"}
        </button>
      </div>

      {error ? (
        <p role="alert" className="mt-3 text-xs leading-5 text-rose-300">{error}</p>
      ) : null}
      {success ? (
        <p role="status" className="mt-3 text-xs leading-5 text-cortex-primary">{success}</p>
      ) : null}

    </form>
  );
}

export function parseEvidenceRefs(value: string) {
  return Array.from(new Set(
    value
      .split(/[\n,]/)
      .map((reference) => reference.trim())
      .filter(Boolean),
  ));
}
