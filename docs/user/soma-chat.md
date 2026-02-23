# Using Soma Chat

> The primary way to interact with Mycelis Cortex — type natural language, Soma orchestrates everything.

---

## Overview

**Workspace** (the first item in the nav rail) is your command interface. Type anything — Soma receives it, reasons about it, consults specialists as needed, and responds.

```
You type → Soma receives → Soma thinks (ReAct loop, up to 5 steps)
         → Soma may consult council (Architect, Coder, Creative, Sentry)
         → Response arrives with optional DelegationTrace + Proposal
```

---

## Sending a Message

1. Navigate to **Workspace** (home icon or `/dashboard`)
2. Type your intent in the input at the bottom of the chat panel
3. Press **Enter** or click the send button

While Soma is processing, a live activity indicator shows what's happening:

| Display | Meaning |
|---------|---------|
| "Soma is thinking" | Initial reasoning |
| "Consulting Architect..." | Soma delegated to the Architect |
| "Generating mission blueprint..." | Blueprint generation tool running |
| "Searching memory..." | Semantic recall in progress |
| "Writing file.py..." | File write tool executing |

---

## Reading a Response

Soma's reply appears as a **council message bubble** in the chat. It contains:

### 1. The Answer
Plain text or markdown — rendered with syntax highlighting for code blocks, tables, links.

### 2. Delegation Trace (when present)
If Soma consulted council members, a compact trace appears beneath the text:

```
Soma consulted
┌──────────────┐  ┌──────────────┐
│  Architect   │  │    Coder     │
│ Designed the │  │ Wrote the    │
│ module layout│  │ initial impl │
└──────────────┘  └──────────────┘
```

Each card shows the member's name and a brief excerpt of their contribution.

### 3. Proposed Action Block (when present)
If your message triggered a system mutation (create files, spawn teams, schedule tasks), a **ProposedActionBlock** appears with:
- Action type and description
- Estimated impact
- **Confirm & Execute** and **Dismiss** buttons

> **Nothing executes until you confirm.** Soma proposes; you decide.

---

## Confirming a Proposal

1. Review the ProposedActionBlock
2. Click **Confirm & Execute** (or **Launch Crew** for crew missions)
3. The system shows a **"Mission activated"** pill in the chat:

   ```
   ⚡ Mission activated — abc1234... →
   ```

4. Click the pill to open the **Run Timeline** at `/runs/{run_id}`

---

## Direct Council Access

To bypass Soma's synthesis and talk directly to a specialist:

1. Click the **⚡ Direct** button in the chat header
2. Select a council member: Architect, Coder, Creative, or Sentry
3. The header shows `→ Architect` in amber — messages now route directly
4. Click **← Soma** to return to Soma-first mode

Direct mode is useful when you want raw specialist output without orchestration — for example, asking the Coder to write a specific function without Soma's architectural framing.

---

## LaunchCrew Modal

For structured multi-step missions, use the **LaunchCrew** button (rocket icon):

1. **Step 1 — Intent**: Describe the mission goal in detail
2. **Step 2 — Wait**: Soma generates and returns a blueprint proposal
3. **Step 3 — Confirm**: Review and click **Launch Crew**
4. A "Mission activated" link appears in chat → click to open the Run Timeline

---

## Tips

- **Be specific**: "Write a Python CSV parser that handles quoted fields and emits a dict per row" gets better results than "write a parser"
- **Reference context**: "Based on the plan you just outlined, implement step 2" — Soma has full chat history
- **Check the trace**: The DelegationTrace tells you which specialist contributed what — useful for understanding how Soma decomposed your request
- **Mutations are safe**: Until you click Confirm, nothing in the system changes. Soma proposes all structural changes explicitly.
