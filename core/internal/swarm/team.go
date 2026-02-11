package swarm

import (
	"context"
	"log"
	"sync"

	"github.com/mycelis/core/internal/cognitive"
	"github.com/nats-io/nats.go"
)

type TeamType string

const (
	TeamTypeAction     TeamType = "action"
	TeamTypeExpression TeamType = "expression"
)

// TeamMember defines an agent's participation in a team.
type TeamMember struct {
	ID   string `yaml:"id"`
	Role string `yaml:"role"`
}

// TeamManifest defines the configuration for a Swarm Team.
type TeamManifest struct {
	ID          string   `yaml:"id"`
	Name        string   `yaml:"name"`
	Type        TeamType `yaml:"type"`
	Description string   `yaml:"description"`
	// Members are the Agents (and their roles) that form this team
	Members []TeamMember `yaml:"members"`
	// Inputs are the NATS subjects this team listens to (Triggers)
	Inputs []string `yaml:"inputs"`
	// Deliveries are the output channels
	Deliveries []string `yaml:"deliveries"`
}

// Team represents a running instance of a TeamManifest.
// It acts as a "Group Chat" container using an internal NATS subject space.
type Team struct {
	Manifest *TeamManifest
	nc       *nats.Conn
	brain    *cognitive.Router
	ctx      context.Context
	cancel   context.CancelFunc
	mu       sync.Mutex
}

// NewTeam creates a new Team instance.
func NewTeam(manifest *TeamManifest, nc *nats.Conn, brain *cognitive.Router) *Team {
	ctx, cancel := context.WithCancel(context.Background())
	return &Team{
		Manifest: manifest,
		nc:       nc,
		brain:    brain,
		ctx:      ctx,
		cancel:   cancel,
	}
}

// Start activates the Team's subscriptions.
func (t *Team) Start() error {
	log.Printf("🐝 Team [%s] (%s) Online.", t.Manifest.Name, t.Manifest.Type)

	// 1. Subscribe to defined Inputs (Triggers)
	for _, subject := range t.Manifest.Inputs {
		t.nc.Subscribe(subject, t.handletrigger)
	}

	// 2. Spawn Agents
	for _, member := range t.Manifest.Members {
		agent := NewAgent(member.ID, member.Role, t.Manifest.ID, t.nc, t.brain)
		go agent.Start()
	}

	return nil
}

// handletrigger receives an external signal and broadens it to the internal team bus.
func (t *Team) handletrigger(msg *nats.Msg) {
	log.Printf("🐝 Team [%s] Triggered by [%s]", t.Manifest.Name, msg.Subject)

	// Forward to Internal Bus for Agents to react
	// Subject: swarm.team.<id>.internal.trigger
	internalSubject := "swarm.team." + t.Manifest.ID + ".internal.trigger"
	t.nc.Publish(internalSubject, msg.Data)
}

// Stop shuts down the team.
func (t *Team) Stop() {
	t.cancel()
}
