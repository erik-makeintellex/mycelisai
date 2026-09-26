package swarm

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/mycelis/core/internal/mcp"
	"github.com/mycelis/core/pkg/protocol"
	"github.com/nats-io/nats.go"
)

// Start activates the Team's subscriptions and member runtime. Start holds the
// team lock so a concurrent Stop observes either no runtime or all of it; a
// failed or stopped Start releases every subscription and agent it created.
func (t *Team) Start() error {
	t.mu.Lock()
	err := t.startLocked()
	t.mu.Unlock()
	if err != nil {
		t.Stop()
	}
	return err
}

func (t *Team) startLocked() error {
	if t.stopped {
		return fmt.Errorf("team %s was stopped before start", t.Manifest.ID)
	}
	log.Printf("Team [%s] (%s) Online.", t.Manifest.Name, t.Manifest.Type)
	t.normalizeRuntimeProviderRouting()

	for _, subject := range runtimeTeamInputSubjects(t.Manifest.ID, t.Manifest.Inputs) {
		subscription, err := t.nc.Subscribe(subject, t.handleTrigger)
		if err != nil {
			return fmt.Errorf("team %s subscribe to input %s: %w", t.Manifest.ID, subject, err)
		}
		t.subscriptions = append(t.subscriptions, subscription)
		log.Printf("Team [%s] Listening on [%s]", t.Manifest.Name, subject)
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

		// Every agent gets a declared-tool scope; NewAgent scopes any other executor.
		var agentToolExec MCPToolExecutor = t.toolExecutor
		if t.compositeExec != nil {
			agentToolExec = NewScopedToolExecutor(t.compositeExec, member.Tools, t.mcpServerNames)
		}

		agent := NewAgent(t.ctx, member, t.Manifest.ID, t.nc, t.brain, agentToolExec)
		t.injectAgentToolDescriptions(agent, member.Tools)
		t.injectAgentRuntimeBindings(agent)
		agent.SetTeamTopology(t.Manifest.Inputs, t.Manifest.Deliveries)
		t.agents = append(t.agents, agent)
		agent.Start()
	}

	internalResponse := fmt.Sprintf(protocol.TopicTeamInternalRespond, t.Manifest.ID)
	responseSubscription, err := t.nc.Subscribe(internalResponse, t.handleResponse)
	if err != nil {
		return fmt.Errorf("team %s subscribe to internal responses: %w", t.Manifest.ID, err)
	}
	t.subscriptions = append(t.subscriptions, responseSubscription)
	if err := t.nc.Flush(); err != nil {
		return fmt.Errorf("team %s establish runtime subscriptions: %w", t.Manifest.ID, err)
	}
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

// normalizeRuntimeProviderRouting never rewrites a provider the manifest
// names: an unavailable provider fails closed per request, and substitution
// happens only through a cognitive same-boundary profile_fallbacks entry that
// never reaches the manifest. The only change is that members without a
// provider inherit the team provider.
func (t *Team) normalizeRuntimeProviderRouting() {
	if t == nil || t.Manifest == nil || t.brain == nil {
		return
	}
	teamProvider := strings.TrimSpace(t.Manifest.Provider)
	if teamProvider == "" {
		return
	}
	for idx := range t.Manifest.Members {
		member := &t.Manifest.Members[idx]
		if strings.TrimSpace(member.Provider) == "" {
			member.Provider = t.Manifest.Provider
		}
	}
}

// Stop shuts down the team and its scheduler (if any). It is idempotent and
// permanent: a stopped team never starts again.
func (t *Team) Stop() {
	t.mu.Lock()
	t.stopped = true
	subscriptions := append([]*nats.Subscription(nil), t.subscriptions...)
	agents := append([]*Agent(nil), t.agents...)
	scheduler := t.scheduler
	t.subscriptions = nil
	t.agents = nil
	t.scheduler = nil
	t.mu.Unlock()
	for _, subscription := range subscriptions {
		if subscription != nil {
			_ = subscription.Unsubscribe()
		}
	}
	for _, agent := range agents {
		if agent != nil {
			agent.Stop()
		}
	}
	if scheduler != nil {
		scheduler.Stop()
	}
	if t.cancel != nil {
		t.cancel()
	}
}
