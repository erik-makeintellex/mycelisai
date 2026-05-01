"use client";

import React, { useState } from "react";
import { X } from "lucide-react";

export function TagInput({
  label,
  value,
  onChange,
}: {
  label: string;
  value: string[];
  onChange: (tags: string[]) => void;
}) {
  const [input, setInput] = useState("");
  const handleKeyDown = (e: React.KeyboardEvent<HTMLInputElement>) => {
    if (e.key !== "Enter") return;
    e.preventDefault();
    const newTags = input
      .split(",")
      .map((t) => t.trim())
      .filter((t) => t.length > 0 && !value.includes(t));
    if (newTags.length > 0) onChange([...value, ...newTags]);
    setInput("");
  };
  const removeTag = (tag: string) => onChange(value.filter((t) => t !== tag));
  return (
    <div>
      <label className="block text-[10px] font-mono uppercase text-cortex-text-muted mb-1">{label}</label>
      {value.length > 0 && (
        <div className="flex flex-wrap gap-1 mb-1.5">
          {value.map((tag) => (
            <span key={tag} className="inline-flex items-center gap-1 text-[10px] font-mono px-1.5 py-0.5 rounded bg-cortex-primary/15 text-cortex-primary border border-cortex-primary/30">
              {tag}
              <button type="button" onClick={() => removeTag(tag)} className="hover:text-cortex-danger transition-colors">
                <X className="w-2.5 h-2.5" />
              </button>
            </span>
          ))}
        </div>
      )}
      <input
        type="text"
        value={input}
        onChange={(e) => setInput(e.target.value)}
        onKeyDown={handleKeyDown}
        placeholder="Type + Enter (comma-separated)"
        className="w-full bg-cortex-bg border border-cortex-border rounded px-2.5 py-1.5 text-xs font-mono text-cortex-text-main placeholder:text-cortex-text-muted/50 focus:outline-none focus:border-cortex-primary transition-colors"
      />
    </div>
  );
}
