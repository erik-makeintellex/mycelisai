"use client";

import { useCallback, useState } from "react";
import {
  previewOrConfirmClone,
  type CloneTemplateResult,
  type GovernedPreview,
} from "./configGovernance";

export type CloneDialogPhase =
  | { status: "form" }
  | { status: "loading" }
  | { status: "previewed"; preview: GovernedPreview }
  | { status: "success"; result: CloneTemplateResult }
  | { status: "conflict" }
  | { status: "violations"; violations: string[] }
  | { status: "denied" }
  | { status: "unavailable"; detail: string }
  | { status: "failed"; detail: string };

/** Drives "Clone as custom template": preview (202), then approve with the token. */
export function useCloneTemplateDialog(sourceTemplateId: string) {
  const [name, setName] = useState("");
  const [phase, setPhase] = useState<CloneDialogPhase>({ status: "form" });

  const setNameAndReset = useCallback((next: string) => {
    setName(next);
    // Editing the name after a preview invalidates the bound token.
    setPhase((prev) => (prev.status === "previewed" ? { status: "form" } : prev));
  }, []);

  const preview = useCallback(async () => {
    if (!name.trim()) return;
    setPhase({ status: "loading" });
    const outcome = await previewOrConfirmClone(sourceTemplateId, { name: name.trim() });
    applyOutcome(outcome, setPhase);
  }, [name, sourceTemplateId]);

  const approve = useCallback(async () => {
    if (phase.status !== "previewed" || phase.preview.blocked || !phase.preview.confirm_token) return;
    setPhase({ status: "loading" });
    const outcome = await previewOrConfirmClone(sourceTemplateId, {
      name: name.trim(),
      confirm_token: phase.preview.confirm_token.token,
    });
    if (outcome.kind === "conflict") {
      // The document changed since preview; re-preview automatically.
      const retried = await previewOrConfirmClone(sourceTemplateId, { name: name.trim() });
      applyOutcome(retried, setPhase);
      return;
    }
    applyOutcome(outcome, setPhase);
  }, [name, phase, sourceTemplateId]);

  const reset = useCallback(() => {
    setName("");
    setPhase({ status: "form" });
  }, []);

  return { name, setName: setNameAndReset, phase, preview, approve, reset };
}

function applyOutcome(
  outcome: Awaited<ReturnType<typeof previewOrConfirmClone>>,
  setPhase: (phase: CloneDialogPhase) => void,
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
      setPhase({ status: "failed", detail: "Source template was not found." });
      return;
    case "error":
      setPhase({ status: "failed", detail: outcome.detail });
      return;
  }
}
