package swarm

import (
	"fmt"
	"log"
	"regexp"
	"strings"

	"github.com/mycelis/core/pkg/protocol"
)

// toolPermittedForPlanning reports whether a proposal-planning capture may
// record this tool. Agents without a scoped executor fall back to their
// manifest's declared list, so capture never widens authority.
func (a *Agent) toolPermittedForPlanning(name string) bool {
	if scoped, ok := a.toolExecutor.(*ScopedToolExecutor); ok && scoped != nil {
		return scoped.PermitsTool(a.ctx, name)
	}
	name = strings.TrimSpace(name)
	return name != "" && newAgentToolScope(a.Manifest.Tools, nil).allowsName(a.ctx, name)
}

// recordToolDenied logs a declared-tool denial and writes a tool.denied
// mission event when the agent is bound to a run.
func (a *Agent) recordToolDenied(tool, phase string, runtimeOwned bool) {
	tool = boundedToolName(tool)
	log.Printf("Agent [%s] tool denied: %q is not declared (phase=%s runtime_owned=%t)", a.Manifest.ID, tool, phase, runtimeOwned)
	if a.eventEmitter == nil || a.runID == "" {
		return
	}
	payload := map[string]interface{}{"tool": tool, "phase": phase, "reason": "not_declared"}
	if runtimeOwned {
		payload["runtime_owned"] = true
	}
	go a.eventEmitter.Emit(a.ctx, a.runID, protocol.EventToolDenied, protocol.SeverityWarn, a.Manifest.ID, a.TeamID, payload) //nolint:errcheck
}

// toolDeniedFeedback is the normalized model-facing denial message.
func toolDeniedFeedback(tool string) string {
	return fmt.Sprintf("Tool '%s' is not permitted for this agent and was not executed. Use only your declared tools, or answer directly.", boundedToolName(tool))
}

var backtickedToolName = regexp.MustCompile("`([a-z_][a-z0-9_.]*)`")

// promptToolNames returns the registered internal tool names cited in
// backticks in text.
func (r *InternalToolRegistry) promptToolNames(text string) []string {
	var names []string
	for _, match := range backtickedToolName.FindAllStringSubmatch(text, -1) {
		if r.Has(match[1]) {
			names = append(names, match[1])
		}
	}
	return names
}

// withoutUndeclaredToolLines drops runtime-context lines that tell the agent
// to call a registered tool it has not declared, so the shared lead protocol
// never steers an agent into a denied call.
func (r *InternalToolRegistry) withoutUndeclaredToolLines(text string, declared []string) string {
	allowed := make(map[string]bool, len(declared))
	for _, tool := range declared {
		allowed[strings.TrimSpace(tool)] = true
	}
	lines := strings.SplitAfter(text, "\n")
	kept := make([]string, 0, len(lines))
	for _, line := range lines {
		keep := true
		for _, name := range r.promptToolNames(line) {
			keep = keep && allowed[name]
		}
		if keep {
			kept = append(kept, line)
		}
	}
	return strings.Join(kept, "")
}
