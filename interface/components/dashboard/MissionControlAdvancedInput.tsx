import { Layers, Send } from "lucide-react";
import type React from "react";

export function MissionControlAdvancedInput({
  value,
  autoFocus,
  broadcastMode,
  isLoading,
  placeholder,
  inputRef,
  onChange,
  onSubmit,
  onQueue,
}: {
  value: string;
  autoFocus?: boolean;
  broadcastMode: boolean;
  isLoading: boolean;
  placeholder: string;
  inputRef: React.RefObject<HTMLInputElement | null>;
  onChange: (value: string) => void;
  onSubmit: () => void;
  onQueue?: () => void;
}) {
  return (
    <div className="flex gap-2">
      <input
        ref={inputRef}
        type="text"
        value={value}
        onChange={(event) => onChange(event.target.value)}
        onKeyDown={(event) => event.key === "Enter" && onSubmit()}
        autoFocus={autoFocus}
        placeholder={isLoading ? "Stack another work prompt into Soma's queue..." : placeholder}
        className={`flex-1 rounded-lg border bg-cortex-bg px-3 py-2 text-sm font-mono text-cortex-text-main placeholder-cortex-text-muted/60 focus:outline-none focus:ring-1 ${
          broadcastMode
            ? "border-cortex-warning/40 focus:border-cortex-warning focus:ring-cortex-warning/30"
            : "border-cortex-border focus:border-cortex-primary focus:ring-cortex-primary/30"
        }`}
      />
      {!isLoading && onQueue && (
        <button
          type="button"
          data-testid="soma-queue-prompt-btn"
          onClick={onQueue}
          disabled={!value.trim()}
          title="Stack work prompt into Soma's queue"
          className="flex h-10 items-center justify-center gap-1 rounded-lg border border-cortex-primary/30 bg-cortex-primary/10 px-2.5 text-xs font-mono text-cortex-primary transition-colors hover:bg-cortex-primary/20 disabled:border-cortex-border disabled:bg-cortex-bg disabled:text-cortex-text-muted"
        >
          <Layers className="h-3.5 w-3.5" />
          <span>+ Queue</span>
        </button>
      )}
      <button
        type="button"
        onClick={onSubmit}
        disabled={!value.trim()}
        className={`flex h-10 min-w-10 items-center justify-center gap-1.5 rounded-lg px-3 text-xs font-mono text-white transition-colors disabled:bg-cortex-border disabled:text-cortex-text-muted ${
          broadcastMode
            ? "bg-cortex-warning hover:bg-cortex-warning/80"
            : "bg-cortex-primary hover:bg-cortex-primary/80"
        }`}
      >
        {isLoading ? (
          <>
            <Layers className="h-3.5 w-3.5" />
            <span>Queue</span>
          </>
        ) : (
          <Send className="h-3.5 w-3.5" />
        )}
      </button>
    </div>
  );
}
