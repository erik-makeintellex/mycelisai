import type { ChatMessage, UIResponseStateProjection } from "@/store/useCortexStore";

export function defaultResponseState(message?: ChatMessage): UIResponseStateProjection {
  if (!message) return { kind: "direct_answer", label: "First request", tone: "neutral" };
  if (message.ui_response_state) return message.ui_response_state;
  if (message.execution_summary?.ui_response_state) return message.execution_summary.ui_response_state;
  if (message.mode === "blocker") return { kind: "blocker", label: "Blocked", tone: "danger" };
  if (message.proposal_status === "confirmed_pending_execution") {
    return { kind: "awaiting_approval", label: "Awaiting approval", tone: "warning" };
  }
  if (message.proposal) return { kind: "proposal", label: "Proposal", tone: "warning" };
  if (message.mode === "execution_result" || message.run_id) {
    return { kind: "execution_result", label: "Execution result", tone: "success" };
  }
  if (message.execution_summary?.audit_recovery && typeof message.execution_summary.audit_recovery !== "string") {
    const degradation = message.execution_summary.audit_recovery.degradation;
    if (degradation?.requires_attention) {
      return { kind: "degraded_execution", label: "Degraded execution", tone: "danger" };
    }
    if (message.execution_summary.audit_recovery.retryable) {
      return { kind: "retry_required", label: "Retry required", tone: "warning" };
    }
    if (message.execution_summary.audit_recovery.recovery_state) {
      return { kind: "recovery_state", label: "Recovery state", tone: "warning" };
    }
  }
  return { kind: "direct_answer", label: "Direct answer", tone: "info" };
}

export function responseStateToneClass(tone: UIResponseStateProjection["tone"]) {
  if (tone === "success") return "border-cortex-success/25 bg-cortex-success/10 text-cortex-success";
  if (tone === "warning") return "border-amber-400/25 bg-amber-400/10 text-amber-300";
  if (tone === "danger") return "border-red-400/30 bg-red-400/10 text-red-300";
  if (tone === "info") return "border-cortex-info/25 bg-cortex-info/10 text-cortex-info";
  return "border-cortex-border bg-cortex-surface/70 text-cortex-text-muted";
}
