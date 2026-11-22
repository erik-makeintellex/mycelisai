import { Layers, Send } from "lucide-react";
import type React from "react";

export function SomaIntentInput({
  value,
  disabled,
  loading,
  placeholder,
  autoFocus,
  inputRef,
  onChange,
  onSubmit,
  onQueue,
}: {
  value: string;
  disabled?: boolean;
  loading?: boolean;
  placeholder: string;
  autoFocus?: boolean;
  inputRef?: React.RefObject<HTMLInputElement | null>;
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
        placeholder={loading ? "Type another work prompt to stack into Soma's queue..." : placeholder}
        autoFocus={autoFocus}
        disabled={disabled}
        className="flex-1 rounded-lg border border-cortex-border bg-cortex-bg px-3 py-2 text-sm text-cortex-text-main placeholder-cortex-text-muted/60 focus:border-cortex-primary focus:outline-none focus:ring-1 focus:ring-cortex-primary/30 disabled:opacity-50"
      />
      {!loading && onQueue && (
        <button
          type="button"
          data-testid="soma-queue-prompt-btn"
          onClick={onQueue}
          disabled={disabled || !value.trim()}
          title="Stack work prompt into Soma's queue"
          className="flex h-10 items-center justify-center gap-1 rounded-lg border border-cortex-primary/30 bg-cortex-primary/10 px-2.5 text-xs font-semibold text-cortex-primary transition-colors hover:bg-cortex-primary/20 disabled:border-cortex-border disabled:bg-cortex-bg disabled:text-cortex-text-muted"
        >
          <Layers className="h-3.5 w-3.5" />
          <span>+ Queue</span>
        </button>
      )}
      <button
        type="button"
        onClick={onSubmit}
        disabled={disabled || !value.trim()}
        title={loading ? "Queue work prompt for Soma to review next" : "Send to Soma"}
        className="flex h-10 min-w-10 items-center justify-center gap-1.5 rounded-lg bg-cortex-primary px-3 text-xs font-semibold text-white transition-colors hover:bg-cortex-primary/80 disabled:bg-cortex-border disabled:text-cortex-text-muted"
      >
        {loading ? (
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
