package swarm

import (
	"encoding/json"
	"log"
	"regexp"
	"strings"
)

// truncateLog shortens a string for log output.
func truncateLog(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "..."
}

// responseSuggestsUnexecutedAction detects common "I will/Step 1" patterns where
// the model narrates an action path instead of invoking tools.
func responseSuggestsUnexecutedAction(text string) bool {
	lower := strings.ToLower(strings.TrimSpace(text))
	if lower == "" {
		return false
	}
	patterns := []string{
		"we need to delegate",
		"let's proceed by",
		"step 1",
		"you can use",
		"would you like me to",
		"to have the",
		"example input",
		"this will route your request",
		"this will delegate",
		"this will consult",
		"i'll consult",
		"i will consult",
		"i'll delegate",
		"i will delegate",
		"route your request",
		"i've delegated",
		"i have delegated",
		"task has been delegated",
	}
	for _, p := range patterns {
		if strings.Contains(lower, p) {
			return true
		}
	}
	return false
}

// toolCallPayload represents a tool invocation extracted from LLM output.
type toolCallPayload struct {
	Name      string         `json:"name"`
	Arguments map[string]any `json:"arguments"`
}

// parseToolCall extracts a tool_call JSON block from LLM response text.
// Handles both compact and pretty-printed JSON from LLMs:
//
//	Compact:  {"tool_call": {"name": "read_file", "arguments": {...}}}
//	Pretty:   {\n  "tool_call": {\n    "name": "read_file" ...
//	Fenced:   ```json\n{"tool_call": ...}\n```
//
// Returns nil if no tool call is found.
func parseToolCall(text string) *toolCallPayload {
	keyword := `"tool_call"`
	idx := strings.Index(text, keyword)
	if idx == -1 {
		return parseOperationCall(text)
	}

	// Walk backwards from "tool_call" to find the opening brace.
	// LLMs may emit whitespace/newlines between { and "tool_call".
	start := -1
	for i := idx - 1; i >= 0; i-- {
		ch := text[i]
		if ch == '{' {
			start = i
			break
		}
		if ch != ' ' && ch != '\t' && ch != '\n' && ch != '\r' {
			break
		}
	}
	if start == -1 {
		return nil
	}

	end := scanJSONObject(text, start)
	if end == -1 {
		if loose := parseLooseToolCall(text[start:]); loose != nil {
			return loose
		}
		return nil
	}

	var wrapper struct {
		ToolCall toolCallPayload `json:"tool_call"`
	}
	if err := json.Unmarshal([]byte(text[start:end]), &wrapper); err != nil {
		log.Printf("[parseToolCall] JSON unmarshal failed: %v (excerpt: %s)", err, truncateLog(text[start:end], 200))
		if loose := parseLooseToolCall(text[start:end]); loose != nil {
			return loose
		}
		return nil
	}
	if wrapper.ToolCall.Name == "" {
		return parseOperationCall(text)
	}
	return &wrapper.ToolCall
}

func parseLooseToolCall(text string) *toolCallPayload {
	nameRe := regexp.MustCompile(`"name"\s*:\s*"([^"]+)"`)
	m := nameRe.FindStringSubmatch(text)
	if len(m) < 2 {
		return nil
	}
	name := strings.TrimSpace(m[1])
	if name == "" {
		return nil
	}
	return &toolCallPayload{Name: name, Arguments: map[string]any{}}
}

// parseOperationCall extracts fallback operation-style payloads emitted by some models:
// {"operation":"consult_council","arguments":{...}}
// This lets the runtime execute the intended tool instead of returning instructions.
func parseOperationCall(text string) *toolCallPayload {
	keyword := `"operation"`
	idx := strings.Index(text, keyword)
	if idx == -1 {
		return nil
	}

	start := -1
	for i := idx - 1; i >= 0; i-- {
		ch := text[i]
		if ch == '{' {
			start = i
			break
		}
		if ch != ' ' && ch != '\t' && ch != '\n' && ch != '\r' {
			break
		}
	}
	if start == -1 {
		return nil
	}

	end := scanJSONObject(text, start)
	if end == -1 {
		return nil
	}

	var payload struct {
		Operation string         `json:"operation"`
		Arguments map[string]any `json:"arguments"`
	}
	if err := json.Unmarshal([]byte(text[start:end]), &payload); err != nil {
		return nil
	}
	if strings.TrimSpace(payload.Operation) == "" {
		return nil
	}
	if payload.Arguments == nil {
		payload.Arguments = map[string]any{}
	}
	return &toolCallPayload{Name: payload.Operation, Arguments: payload.Arguments}
}

// autofillToolArguments patches common missing fields for tool calls that are
// clearly actionable but slightly malformed from smaller/local model outputs.
func autofillToolArguments(call *toolCallPayload, latestUserInput string) {
	if call == nil {
		return
	}
	if call.Arguments == nil {
		call.Arguments = map[string]any{}
	}

	switch call.Name {
	case "consult_council":
		member, _ := call.Arguments["member"].(string)
		if strings.TrimSpace(member) == "" {
			member = inferCouncilMemberFromInput(latestUserInput)
		}
		if normalized := normalizeCouncilMember(member); normalized != "" {
			call.Arguments["member"] = normalized
		}
		question, _ := call.Arguments["question"].(string)
		if strings.TrimSpace(question) == "" {
			if q := strings.TrimSpace(latestUserInput); q != "" {
				call.Arguments["question"] = q
			}
		}
	case "read_signals":
		subject, _ := call.Arguments["subject"].(string)
		if strings.TrimSpace(subject) == "" {
			if v, _ := call.Arguments["topic_pattern"].(string); strings.TrimSpace(v) != "" {
				call.Arguments["subject"] = strings.TrimSpace(v)
			} else if v, _ := call.Arguments["topic"].(string); strings.TrimSpace(v) != "" {
				call.Arguments["subject"] = strings.TrimSpace(v)
			} else if v, _ := call.Arguments["channel"].(string); strings.TrimSpace(v) != "" {
				call.Arguments["subject"] = strings.TrimSpace(v)
			} else if v := extractNATSSubject(latestUserInput); v != "" {
				call.Arguments["subject"] = v
			}
		}
	case "delegate_task":
		if _, hasTask := call.Arguments["task"]; !hasTask {
			if teamName, _ := call.Arguments["team_name"].(string); strings.TrimSpace(teamName) != "" {
				call.Name = "create_team"
				call.Arguments["team_id"] = strings.TrimSpace(teamName)
				if role, _ := call.Arguments["agent_type"].(string); strings.TrimSpace(role) != "" {
					call.Arguments["role"] = strings.TrimSpace(role)
				}
			}
		}
	}
}

func scanJSONObject(text string, start int) int {
	depth := 0
	end := -1
	inStr := false
	esc := false
	for i := start; i < len(text); i++ {
		ch := text[i]
		if esc {
			esc = false
			continue
		}
		if ch == '\\' && inStr {
			esc = true
			continue
		}
		if ch == '"' {
			inStr = !inStr
			continue
		}
		if inStr {
			continue
		}
		switch ch {
		case '{':
			depth++
		case '}':
			depth--
			if depth == 0 {
				end = i + 1
			}
		}
		if end != -1 {
			break
		}
	}
	return end
}
