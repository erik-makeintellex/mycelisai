import type { TeamWorkItem } from "@/store/useCortexStore";

export function prioritizeRequestedWorkItem(items: TeamWorkItem[], requestedId: string | null) {
  if (!requestedId) return items;
  const requested = items.find((item) => item.id === requestedId);
  if (!requested) return items;
  return [requested, ...items.filter((item) => item.id !== requestedId)];
}
