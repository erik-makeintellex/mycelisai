import { useLayoutEffect } from "react";
import { Loader2, Send } from "lucide-react";
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
}: {
  value: string;
  autoFocus?: boolean;
  broadcastMode: boolean;
  isLoading: boolean;
  placeholder: string;
  inputRef: React.RefObject<HTMLTextAreaElement | null>;
  onChange: (value: string) => void;
  onSubmit: () => void;
}) {
  useLayoutEffect(() => {
    const input = inputRef.current;
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
        autoFocus={autoFocus}
        placeholder={placeholder}
        disabled={isLoading}
        rows={1}
        className={`max-h-[180px] min-h-10 flex-1 resize-none overflow-y-auto rounded-lg border bg-cortex-bg px-3 py-2 text-sm font-mono leading-6 text-cortex-text-main placeholder-cortex-text-muted/60 focus:outline-none focus:ring-1 disabled:opacity-50 ${
          broadcastMode
            ? "border-cortex-warning/40 focus:border-cortex-warning focus:ring-cortex-warning/30"
            : "border-cortex-border focus:border-cortex-primary focus:ring-cortex-primary/30"
        }`}
      />
      <button
        type="button"
        aria-label={isLoading ? "Submitting Soma message" : "Submit Soma message"}
        onClick={onSubmit}
        disabled={isLoading || !value.trim()}
        className={`flex h-10 min-w-10 items-center justify-center rounded-lg px-3 text-white transition-colors disabled:bg-cortex-border disabled:text-cortex-text-muted ${
          broadcastMode
            ? "bg-cortex-warning hover:bg-cortex-warning/80"
            : "bg-cortex-primary hover:bg-cortex-primary/80"
        }`}
      >
        {isLoading ? (
          <Loader2 className="h-3.5 w-3.5 animate-spin" />
        ) : (
          <Send className="h-3.5 w-3.5" />
        )}
      </button>
    </div>
  );
}
