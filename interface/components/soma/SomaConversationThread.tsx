import type React from "react";

export function SomaConversationThread({
  children,
  scrollRef,
}: {
  children: React.ReactNode;
  scrollRef: React.RefObject<HTMLDivElement | null>;
}) {
  return (
    <div
      ref={scrollRef}
      data-testid="soma-conversation-thread"
      className="min-h-0 flex-1 space-y-2.5 overflow-y-auto px-3 py-3 scrollbar-thin scrollbar-thumb-cortex-border"
    >
      {children}
    </div>
  );
}
