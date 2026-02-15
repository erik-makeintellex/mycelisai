package swarm

import (
	"context"
	"fmt"
	"log"

	"github.com/mycelis/core/internal/cognitive"
	"github.com/nats-io/nats.go"
)

// Agent represents a single node in a Swarm Team.
type Agent struct {
	ID     string
	Role   string
	TeamID string
	nc     *nats.Conn
	brain  *cognitive.Router
}

// NewAgent creates a new Agent instance.
func NewAgent(id, role, teamID string, nc *nats.Conn, brain *cognitive.Router) *Agent {
	return &Agent{
		ID:     id,
		Role:   role,
		TeamID: teamID,
		nc:     nc,
		brain:  brain,
	}
}

// Start brings the Agent online to listen to its team's internal chatter.
func (a *Agent) Start() {
	// Listen to triggers on the internal team bus
	subject := fmt.Sprintf("swarm.team.%s.internal.trigger", a.TeamID)
	a.nc.Subscribe(subject, a.handleTrigger)
	log.Printf("🤖 Agent [%s] (%s) joined Team [%s]", a.ID, a.Role, a.TeamID)
}

func (a *Agent) handleTrigger(msg *nats.Msg) {
	log.Printf("🤖 Agent [%s] thinking about: %s", a.ID, string(msg.Data))

	if a.brain == nil {
		log.Printf("⚠️ Agent [%s] has no brain. Skipping inference.", a.ID)
		return
	}

	// 1. Construct Prompt from Role + Input
	prompt := fmt.Sprintf("You are a %s in the %s team. Input: %s", a.Role, a.TeamID, string(msg.Data))

	// 2. Infer (Default to 'chat' profile or 'local_ollama' directly?)
	// Let's use 'chat' profile for now.
	req := cognitive.InferRequest{
		Profile: "chat",
		Prompt:  prompt,
	}

	resp, err := a.brain.InferWithContract(context.Background(), req)
	if err != nil {
		log.Printf("❌ Agent [%s] brain freeze: %v", a.ID, err)
		return
	}

	// 3. Reply
	replySubject := fmt.Sprintf("swarm.team.%s.internal.response", a.TeamID)
	a.nc.Publish(replySubject, []byte(resp.Text))
	log.Printf("🤖 Agent [%s] replied.", a.ID)
}
