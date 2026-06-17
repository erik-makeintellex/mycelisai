import { useLayoutEffect } from "react";
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
  inputRef?: React.RefObject<HTMLTextAreaElement | null>;
  onChange: (value: string) => void;
  onSubmit: () => void;
}) {
  useLayoutEffect(() => {
    const input = inputRef?.current;
    if (!input) return;
    input.style.height = "auto";
    const nextHeight = Math.min(input.scrollHeight, 180);
    input.style.height = `${nextHeight}px`;
    input.style.overflowY = input.scrollHeight > nextHeight ? "auto" : "hidden";
  }, [inputRef, value]);

  return (
    <div className="flex items-end gap-2">
      <textarea
        ref={inputRef}
        value={value}
        onChange={(event) => onChange(event.target.value)}
        onKeyDown={(event) => {
          if (event.key === "Enter" && !event.shiftKey) {
            event.preventDefault();
            onSubmit();
          }
        }}
        placeholder={placeholder}
        autoFocus={autoFocus}
        disabled={disabled}
        rows={1}
        className="max-h-[180px] min-h-10 flex-1 resize-none overflow-y-auto rounded-lg border border-cortex-border bg-cortex-bg px-3 py-2 text-sm leading-6 text-cortex-text-main placeholder-cortex-text-muted/60 focus:border-cortex-primary focus:outline-none focus:ring-1 focus:ring-cortex-primary/30 disabled:opacity-50"
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
