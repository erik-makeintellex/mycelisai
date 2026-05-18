import { Loader2, Send } from "lucide-react";
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
}: {
  value: string;
  disabled?: boolean;
  loading?: boolean;
  placeholder: string;
  autoFocus?: boolean;
  inputRef?: React.RefObject<HTMLInputElement | null>;
  onChange: (value: string) => void;
  onSubmit: () => void;
}) {
  return (
    <div className="flex gap-2">
      <input
        ref={inputRef}
        type="text"
        value={value}
        onChange={(event) => onChange(event.target.value)}
        onKeyDown={(event) => event.key === "Enter" && onSubmit()}
        placeholder={placeholder}
        autoFocus={autoFocus}
        disabled={disabled}
        className="flex-1 rounded-lg border border-cortex-border bg-cortex-bg px-3 py-2 text-sm text-cortex-text-main placeholder-cortex-text-muted/60 focus:border-cortex-primary focus:outline-none focus:ring-1 focus:ring-cortex-primary/30 disabled:opacity-50"
      />
      <button
        type="button"
        aria-label={loading ? "Submitting Soma request" : "Submit Soma request"}
        onClick={onSubmit}
        disabled={disabled || !value.trim()}
        className="flex h-10 min-w-10 items-center justify-center rounded-lg bg-cortex-primary px-3 text-white transition-colors hover:bg-cortex-primary/80 disabled:bg-cortex-border disabled:text-cortex-text-muted"
      >
        {loading ? (
          <Loader2 className="h-3.5 w-3.5 animate-spin" />
        ) : (
          <Send className="h-3.5 w-3.5" />
        )}
      </button>
    </div>
  );
}
