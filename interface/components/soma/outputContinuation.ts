import { useEffect, type RefObject } from "react";
import type { MissionChatContinuationContext } from "@/store/cortexStoreTypes";

export type OutputContinuationDetail = {
  title: string;
  reference?: string | null;
  proof?: string | null;
  sourceLabel?: string | null;
  teamId?: string | null;
  runId?: string | null;
  workItemId?: string | null;
  outputId?: string | null;
  contentDigest?: string | null;
};

export const OUTPUT_CONTINUATION_EVENT = "mycelis:soma-output-continuation";
const PENDING_OUTPUT_CONTINUATION_KEY = "mycelis:pending-soma-output-continuation";

export function outputContinuationPrompt(detail: OutputContinuationDetail) {
  const title = detail.title.trim() || "this delivered output";
  const sourceLabel = detail.sourceLabel?.trim() || "delivered output";
  const reference = detail.reference?.trim();
  const proof = detail.proof?.trim();
  const parts = [`Use ${sourceLabel} "${title}" as context.`];
  if (reference) parts.push(`Reference: ${reference}.`);
  if (proof) parts.push(`Proof: ${proof}.`);
  parts.push("I want an update, alternate version, or follow-up generation: ");
  return parts.join("\n");
}

export function requestSomaOutputContinuation(
  detail: OutputContinuationDetail,
  options: { persist?: boolean; openSoma?: boolean } = {},
) {
  if (typeof window === "undefined") return;
  if (options.persist) {
    savePendingSomaOutputContinuation(detail);
  }
  window.dispatchEvent(new CustomEvent<OutputContinuationDetail>(OUTPUT_CONTINUATION_EVENT, { detail }));
  if (options.openSoma && window.location.pathname !== "/dashboard") {
    window.location.assign("/dashboard");
  }
}

export function outputContinuationContext(detail: OutputContinuationDetail): MissionChatContinuationContext {
  return {
    kind: "output",
    title: detail.title.trim() || "Delivered output",
    reference: detail.reference?.trim() || undefined,
    proof: detail.proof?.trim() || undefined,
    team_id: detail.teamId?.trim() || undefined,
    run_id: detail.runId?.trim() || undefined,
    work_item_id: detail.workItemId?.trim() || undefined,
    output_id: detail.outputId?.trim() || undefined,
    content_digest: detail.contentDigest?.trim() || undefined,
  };
}

export function savePendingSomaOutputContinuation(detail: OutputContinuationDetail) {
  if (typeof window === "undefined") return;
  window.sessionStorage.setItem(PENDING_OUTPUT_CONTINUATION_KEY, JSON.stringify(detail));
}

export function takePendingSomaOutputContinuation(): OutputContinuationDetail | null {
  if (typeof window === "undefined") return null;
  const raw = window.sessionStorage.getItem(PENDING_OUTPUT_CONTINUATION_KEY);
  if (!raw) return null;
  window.sessionStorage.removeItem(PENDING_OUTPUT_CONTINUATION_KEY);
  try {
    const parsed = JSON.parse(raw) as OutputContinuationDetail;
    if (!parsed || typeof parsed.title !== "string") return null;
    return parsed;
  } catch {
    return null;
  }
}

export function useSomaOutputContinuation({
  disabled,
  inputRef,
  setInput,
  setContinuationContext,
}: {
  disabled: boolean;
  inputRef: RefObject<HTMLTextAreaElement | null>;
  setInput: (value: string) => void;
  setContinuationContext?: (value: MissionChatContinuationContext) => void;
}) {
  useEffect(() => {
    if (typeof window === "undefined") return;
    const applyContinuation = (detail: OutputContinuationDetail) => {
      if (disabled) return;
      if (setContinuationContext) {
        setInput("");
        setContinuationContext(outputContinuationContext(detail));
      } else {
        setInput(outputContinuationPrompt(detail));
      }
      window.setTimeout(() => {
        const input = inputRef.current;
        input?.focus();
        input?.setSelectionRange(input.value.length, input.value.length);
      }, 0);
    };

    const pending = takePendingSomaOutputContinuation();
    if (pending) applyContinuation(pending);

    const handleContinuation = (event: Event) => {
      applyContinuation((event as CustomEvent<OutputContinuationDetail>).detail);
    };
    window.addEventListener(OUTPUT_CONTINUATION_EVENT, handleContinuation);
    return () => window.removeEventListener(OUTPUT_CONTINUATION_EVENT, handleContinuation);
  }, [disabled, inputRef, setContinuationContext, setInput]);
}
