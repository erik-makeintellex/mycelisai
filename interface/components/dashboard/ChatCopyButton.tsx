"use client";

import { Check, Copy } from "lucide-react";
import { useState } from "react";

export default function ChatCopyButton({ text }: { text: string }) {
    const [copied, setCopied] = useState(false);

    const handleCopy = () => {
        navigator.clipboard.writeText(text);
        setCopied(true);
        setTimeout(() => setCopied(false), 1500);
    };

    return (
        <button
            onClick={handleCopy}
            className="rounded p-1 text-cortex-text-muted transition-colors hover:bg-cortex-border hover:text-cortex-text-main"
            title="Copy"
        >
            {copied ? <Check className="h-3 w-3 text-cortex-success" /> : <Copy className="h-3 w-3" />}
        </button>
    );
}
