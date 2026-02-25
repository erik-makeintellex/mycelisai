// Agent Conversation Log — TypeScript types for turn-by-turn conversation data.
// Mirrors core conversation_turns table.
// Used by ConversationLog.tsx, TurnCard.tsx components.

export type TurnRole =
  | 'system'
  | 'user'
  | 'assistant'
  | 'tool_call'
  | 'tool_result'
  | 'interjection';

export interface ConversationTurn {
    id: string;
    run_id?: string;
    session_id: string;
    agent_id: string;
    team_id?: string;
    turn_index: number;
    role: TurnRole;
    content: string;
    provider_id?: string;
    model_used?: string;
    tool_name?: string;
    tool_args?: Record<string, unknown>;
    parent_turn_id?: string;
    consultation_of?: string;
    created_at: string;
}

// Role-based border colors for TurnCard rendering
export const TURN_ROLE_STYLES: Record<TurnRole, { border: string; label: string; labelColor: string }> = {
    system:       { border: 'border-l-zinc-500',   label: 'SYSTEM',                labelColor: 'text-zinc-400' },
    user:         { border: 'border-l-cyan-500',    label: 'USER',                  labelColor: 'text-cyan-400' },
    assistant:    { border: 'border-l-emerald-500', label: 'ASSISTANT',             labelColor: 'text-emerald-400' },
    tool_call:    { border: 'border-l-violet-500',  label: 'TOOL CALL',            labelColor: 'text-violet-400' },
    tool_result:  { border: 'border-l-amber-500',   label: 'TOOL RESULT',          labelColor: 'text-amber-400' },
    interjection: { border: 'border-l-red-500',     label: 'OPERATOR INTERJECTION', labelColor: 'text-red-400' },
};
