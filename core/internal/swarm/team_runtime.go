package swarm

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/mycelis/core/internal/mcp"
	"github.com/mycelis/core/pkg/protocol"
)

// Start activates the Team's subscriptions and member runtime.
func (t *Team) Start() error {
	log.Printf("Team [%s] (%s) Online.", t.Manifest.Name, t.Manifest.Type)
	t.normalizeRuntimeProviderRouting()

	for _, subject := range t.Manifest.Inputs {
		if _, err := t.nc.Subscribe(subject, t.handleTrigger); err != nil {
			log.Printf("Team [%s] Failed to subscribe to input [%s]: %v", t.Manifest.Name, subject, err)
		} else {
			log.Printf("Team [%s] Listening on [%s]", t.Manifest.Name, subject)
		}
	}

	for _, manifest := range t.Manifest.Members {
		member := manifest
		if member.Provider == "" && t.Manifest.Provider != "" {
			member.Provider = t.Manifest.Provider
		}

		if cfg, isSensor := t.sensorConfigs[manifest.ID]; isSensor {
			sensor := NewSensorAgent(t.ctx, member, cfg, t.Manifest.ID, t.nc)
			go sensor.Start()
			continue
		}

		var agentToolExec MCPToolExecutor = t.toolExecutor
		if t.compositeExec != nil {
			mcpRefs := mcp.ExtractMCPRefs(member.Tools)
			agentToolExec = NewScopedToolExecutor(t.compositeExec, mcpRefs, t.mcpServerNames)
		}

		agent := NewAgent(t.ctx, member, t.Manifest.ID, t.nc, t.brain, agentToolExec)
		t.injectAgentToolDescriptions(agent, member.Tools)
		t.injectAgentRuntimeBindings(agent)
		agent.SetTeamTopology(t.Manifest.Inputs, t.Manifest.Deliveries)
		go agent.Start()
	}

	internalResponse := fmt.Sprintf(protocol.TopicTeamInternalRespond, t.Manifest.ID)
	t.nc.Subscribe(internalResponse, t.handleResponse)
	t.startScheduler()
	return nil
}

func (t *Team) injectAgentToolDescriptions(agent *Agent, memberTools []string) {
	if len(memberTools) == 0 || len(t.toolDescs) == 0 {
		return
	}

	agentDescs := make(map[string]string, len(memberTools))
	for _, name := range memberTools {
		if desc, ok := t.toolDescs[name]; ok {
			agentDescs[name] = desc
		}
		if !mcp.IsMCPRef(name) {
			continue
		}
		ref := mcp.ParseToolRef(name)
		if ref == nil {
			continue
		}
		if ref.ToolName != "*" {
			if desc, ok := t.mcpToolDescs[ref.ToolName]; ok {
				agentDescs[ref.ToolName] = desc
			}
			continue
		}
		for toolName, desc := range t.mcpToolDescs {
			agentDescs[toolName] = desc
		}
	}
	agent.SetToolDescriptions(agentDescs)
}

func (t *Team) injectAgentRuntimeBindings(agent *Agent) {
	if t.internalTools != nil {
		agent.SetInternalTools(t.internalTools)
	}
	if t.eventEmitter != nil && t.runID != "" {
		agent.SetEventEmitter(t.eventEmitter, t.runID)
	}
	if t.conversationLogger != nil {
		agent.SetConversationLogger(t.conversationLogger)
	}
}

func (t *Team) startScheduler() {
	if t.Manifest.Schedule == nil || t.Manifest.Schedule.Interval == "" {
		return
	}

	interval, err := time.ParseDuration(t.Manifest.Schedule.Interval)
	if err != nil {
		log.Printf("Team [%s] invalid schedule interval %q: %v", t.Manifest.Name, t.Manifest.Schedule.Interval, err)
		return
	}
	if interval <= 0 {
		return
	}

	const minInterval = 30 * time.Second
	if interval < minInterval {
		log.Printf("WARN: Team [%s] schedule interval %s below minimum, clamping to %s", t.Manifest.Name, interval, minInterval)
		interval = minInterval
	}

	schedCtx, schedCancel := context.WithCancel(t.ctx)
	t.scheduler = &TeamScheduler{
		teamID:   t.Manifest.ID,
		interval: interval,
		nc:       t.nc,
		ctx:      schedCtx,
		cancel:   schedCancel,
	}
	go t.scheduler.Start()
}

func (t *Team) normalizeRuntimeProviderRouting() {
	if t == nil || t.Manifest == nil || t.brain == nil {
		return
	}

	if explicit := strings.TrimSpace(t.Manifest.Provider); explicit != "" {
		if availability := t.brain.ExecutionAvailability("", explicit); availability.Available && availability.FallbackApplied && availability.ProviderID != "" {
			log.Printf("WARN: Team [%s] provider '%s' is not executable; falling back to '%s'", t.Manifest.Name, explicit, availability.ProviderID)
			t.Manifest.Provider = availability.ProviderID
		}
	}

	for idx := range t.Manifest.Members {
		member := &t.Manifest.Members[idx]
		explicit := strings.TrimSpace(member.Provider)
		if explicit == "" && strings.TrimSpace(t.Manifest.Provider) != "" {
			member.Provider = t.Manifest.Provider
			continue
		}
		if explicit == "" {
			continue
		}
		availability := t.brain.ExecutionAvailability("", explicit)
		if availability.Available && availability.FallbackApplied && availability.ProviderID != "" {
			log.Printf("WARN: Team [%s] member [%s] provider '%s' is not executable; falling back to '%s'", t.Manifest.Name, member.ID, explicit, availability.ProviderID)
			member.Provider = availability.ProviderID
		}
	}
}

// Stop shuts down the team and its scheduler (if any).
func (t *Team) Stop() {
	if t.scheduler != nil {
		t.scheduler.Stop()
	}
	t.cancel()
}
