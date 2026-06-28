import { useEffect, type RefObject } from "react";

export type OutputContinuationDetail = {
  title: string;
  reference?: string | null;
  proof?: string | null;
};

export const OUTPUT_CONTINUATION_EVENT = "mycelis:soma-output-continuation";

export function outputContinuationPrompt(detail: OutputContinuationDetail) {
  const title = detail.title.trim() || "this delivered output";
  const reference = detail.reference?.trim();
  const proof = detail.proof?.trim();
  const parts = [`Use delivered output "${title}" as context.`];
  if (reference) parts.push(`Reference: ${reference}.`);
  if (proof) parts.push(`Proof: ${proof}.`);
  parts.push("I want an update, alternate version, or follow-up generation: ");
  return parts.join("\n");
}

export function requestSomaOutputContinuation(detail: OutputContinuationDetail) {
  if (typeof window === "undefined") return;
  window.dispatchEvent(new CustomEvent<OutputContinuationDetail>(OUTPUT_CONTINUATION_EVENT, { detail }));
}

export function useSomaOutputContinuation({
  disabled,
  inputRef,
  setInput,
}: {
  disabled: boolean;
  inputRef: RefObject<HTMLTextAreaElement | null>;
  setInput: (value: string) => void;
}) {
  useEffect(() => {
    if (typeof window === "undefined") return;
    const handleContinuation = (event: Event) => {
      if (disabled) return;
      const detail = (event as CustomEvent<OutputContinuationDetail>).detail;
      setInput(outputContinuationPrompt(detail));
      window.setTimeout(() => {
        const input = inputRef.current;
        input?.focus();
        input?.setSelectionRange(input.value.length, input.value.length);
      }, 0);
    };
    window.addEventListener(OUTPUT_CONTINUATION_EVENT, handleContinuation);
    return () => window.removeEventListener(OUTPUT_CONTINUATION_EVENT, handleContinuation);
  }, [disabled, inputRef, setInput]);
}
