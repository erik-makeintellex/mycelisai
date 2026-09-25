"use client";

import { useCallback, useState } from "react";
import {
  previewOrConfirmApply,
  type ApplyConfigResult,
  type GovernedPreview,
} from "./configGovernance";

export type ApplyDialogPhase =
  | { status: "editing" }
  | { status: "loading" }
  | { status: "previewed"; preview: GovernedPreview }
  | { status: "success"; result: ApplyConfigResult }
  | { status: "conflict" }
  | { status: "violations"; violations: string[] }
  | { status: "denied" }
  | { status: "unavailable"; detail: string }
  | { status: "failed"; detail: string };

/** Drives "Edit and apply": preview (202) on demand, approve with the bound token. */
export function useApplyConfigDialog(kind: "templates" | "groups", id: string, initialDocument: string) {
  const [document, setDocument] = useState(initialDocument);
  const [format] = useState<"yaml" | "json">("yaml");
  const [phase, setPhase] = useState<ApplyDialogPhase>({ status: "editing" });

  const setDocumentAndReset = useCallback((next: string) => {
    setDocument(next);
    // Editing the document after a preview invalidates the hash-bound token.
    setPhase((prev) => (prev.status === "previewed" ? { status: "editing" } : prev));
  }, []);

  const preview = useCallback(async () => {
    setPhase({ status: "loading" });
    const outcome = await previewOrConfirmApply(kind, id, { document, format });
    applyOutcome(outcome, setPhase);
  }, [document, format, kind, id]);

  const approve = useCallback(async () => {
    if (phase.status !== "previewed" || phase.preview.blocked || !phase.preview.confirm_token) return;
    setPhase({ status: "loading" });
    const outcome = await previewOrConfirmApply(kind, id, {
      document,
      format,
      confirm_token: phase.preview.confirm_token.token,
    });
    if (outcome.kind === "conflict") {
      // The stored document changed since preview; re-preview automatically.
      const retried = await previewOrConfirmApply(kind, id, { document, format });
      applyOutcome(retried, setPhase);
      return;
    }
    applyOutcome(outcome, setPhase);
  }, [document, format, kind, id, phase]);

  const reset = useCallback(() => {
    setDocument(initialDocument);
    setPhase({ status: "editing" });
  }, [initialDocument]);

  return { document, setDocument: setDocumentAndReset, format, phase, preview, approve, reset };
}

function applyOutcome(
  outcome: Awaited<ReturnType<typeof previewOrConfirmApply>>,
  setPhase: (phase: ApplyDialogPhase) => void,
) {
  switch (outcome.kind) {
    case "preview":
      setPhase({ status: "previewed", preview: outcome.preview });
      return;
    case "success":
      setPhase({ status: "success", result: outcome.data });
      return;
    case "conflict":
      setPhase({ status: "conflict" });
      return;
    case "write_back_violations":
      setPhase({ status: "violations", violations: outcome.violations });
      return;
    case "forbidden":
      setPhase({ status: "denied" });
      return;
    case "unavailable":
      setPhase({ status: "unavailable", detail: outcome.detail });
      return;
    case "bad_request":
      setPhase({ status: "failed", detail: outcome.detail });
      return;
    case "not_found":
      setPhase({ status: "failed", detail: "Config target was not found." });
      return;
    case "error":
      setPhase({ status: "failed", detail: outcome.detail });
      return;
  }
}
