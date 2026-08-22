import type { ChatMessage } from "@/store/useCortexStore";

function operationalKey(message: ChatMessage) {
    const event = message.thread_event ?? message.thread_events?.at(-1);
    if (!event) return null;
    return message.run_id ?? event.run_id ?? event.team_id ?? null;
}

/** Keep durable machine history in state while showing only the latest update per work item. */
export function presentMissionChat(messages: ChatMessage[]) {
    const hiddenIndexes = new Set<number>();
    let blockStart = 0;

    const compactBlock = (start: number, end: number) => {
        const latestByWork = new Map<string, number>();
        for (let index = start; index < end; index += 1) {
            const key = operationalKey(messages[index]);
            if (!key) continue;
            const previous = latestByWork.get(key);
            if (previous !== undefined) hiddenIndexes.add(previous);
            latestByWork.set(key, index);
        }
    };

    for (let index = 0; index <= messages.length; index += 1) {
        if (index < messages.length && messages[index].role === "system") continue;
        compactBlock(blockStart, index);
        blockStart = index + 1;
    }

    return messages.filter((_, index) => !hiddenIndexes.has(index));
}
