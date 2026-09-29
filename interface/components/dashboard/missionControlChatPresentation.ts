import type { ChatMessage } from "@/store/useCortexStore";

function operationalKey(message: ChatMessage) {
    // E2E-T1: a still-pending proposal card is the only place
    // "Confirmed, waiting for result" / "Approved, still running" is
    // rendered. Once confirmProposal() records a run_id on this message
    // (cortexStoreProposalExecutionSlice.ts), it would otherwise share an
    // operational key with the terminal system message appended for that
    // same run and get silently collapsed away, leaving no
    // confirmed-pending-execution trace at all. Exempt it from this dedup
    // while it is still pending. Once the proposal reaches its terminal
    // "executed" lifecycle, its own completed-state render duplicates the
    // appended terminal card, so let the normal dedup collapse it there.
    if (message.proposal && message.proposal_status !== "executed") return null;
    const event = message.thread_event ?? message.thread_events?.at(-1);
    if (!event) return null;
    return message.run_id ?? event.run_id ?? event.team_id ?? null;
}

/**
 * Keep durable machine history in state while showing only the latest
 * thread-state update per work item (run_id/team_id), across the whole
 * conversation.
 *
 * Live test L1: this used to compact only within one contiguous run of
 * "system"-role messages, so a "council" role card, or a user message
 * inserted between an earlier "running" update and a later "completed" one
 * for the same run, kept the stale running card on screen alongside the
 * completed result. Deduplicating globally by key collapses the running
 * card into the completed state regardless of what role or message sits
 * between them.
 */
export function presentMissionChat(messages: ChatMessage[]) {
    const lastIndexByKey = new Map<string, number>();
    messages.forEach((message, index) => {
        const key = operationalKey(message);
        if (key) lastIndexByKey.set(key, index);
    });

    return messages.filter((message, index) => {
        const key = operationalKey(message);
        return !key || lastIndexByKey.get(key) === index;
    });
}
