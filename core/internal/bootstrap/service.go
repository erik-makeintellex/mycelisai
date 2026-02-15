package bootstrap

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/nats-io/nats.go"
	"google.golang.org/protobuf/proto"

	pb "github.com/mycelis/core/pkg/pb/swarm"
	"github.com/mycelis/core/pkg/protocol"
)

// Service listens for new hardware announcements
type Service struct {
	db *sql.DB
	nc *nats.Conn
}

func NewService(db *sql.DB, nc *nats.Conn) *Service {
	return &Service{
		db: db,
		nc: nc,
	}
}

// processAnnouncement handles the core logic independent of NATS
func (s *Service) processAnnouncement(data []byte) error {
	var payload struct {
		ID    string          `json:"id"`
		Type  string          `json:"type"`
		Specs json.RawMessage `json:"specs"`
	}
	if err := json.Unmarshal(data, &payload); err != nil {
		log.Printf("Bootstrap: Malformed announcement: %v", err)
		return err
	}

	_, err := s.db.Exec(`
		INSERT INTO nodes (id, type, status, last_seen, specs)
		VALUES ($1, $2, 'pending', NOW(), $3)
		ON CONFLICT (id) DO UPDATE SET
			last_seen = NOW()
	`, payload.ID, payload.Type, payload.Specs)

	if err != nil {
		log.Printf("Bootstrap: DB Error: %v", err)
		return err
	}

	log.Printf("Bootstrap: Detected Node %s", payload.ID)
	return nil
}

// processHeartbeat handles agent pulses via Protobuf
func (s *Service) processHeartbeat(data []byte) error {
	var env pb.MsgEnvelope
	if err := proto.Unmarshal(data, &env); err != nil {
		log.Printf("Bootstrap: Failed to unmarshal Heartbeat: %v", err)
		return err
	}

	agentID := env.SourceAgentId
	teamID := env.TeamId

	// Upsert into Nodes table
	// We treat Agents as Nodes with type 'agent' or 'agent:<team>'
	nodeType := "agent"
	if teamID != "" {
		nodeType = "agent:" + teamID
	}

	// Status from event?
	status := "alive"

	// Specs? Maybe just the swarm config or empty for now.
	specs := []byte("{}")

	_, err := s.db.Exec(`
		INSERT INTO nodes (id, type, status, last_seen, specs)
		VALUES ($1, $2, $3, NOW(), $4)
		ON CONFLICT (id) DO UPDATE SET
			last_seen = NOW(),
			status = $3
	`, agentID, nodeType, status, specs)

	if err != nil {
		log.Printf("Bootstrap: DB Error on Heartbeat: %v", err)
		return err
	}

	return nil
}

func (s *Service) Start() {
	// Announce (Legacy JSON device)
	s.nc.Subscribe(protocol.TopicGlobalAnnounce, func(msg *nats.Msg) {
		s.processAnnouncement(msg.Data)
	})

	// Heartbeat (Protobuf Agent)
	s.nc.Subscribe(protocol.TopicGlobalHeartbeat, func(msg *nats.Msg) {
		s.processHeartbeat(msg.Data)
	})

	log.Println("👂 Bootstrap Listener Active (announce, heartbeat)")
}

// HandlePendingNodes returns list of unassigned nodes
// GET /api/v1/nodes/pending
func (s *Service) HandlePendingNodes(w http.ResponseWriter, r *http.Request) {
	rows, err := s.db.Query("SELECT id, type, status, last_seen, specs FROM nodes WHERE status = 'pending'")
	if err != nil {
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var nodes []map[string]interface{}
	for rows.Next() {
		var id, nType, status string
		var lastSeen time.Time
		var specs json.RawMessage
		if err := rows.Scan(&id, &nType, &status, &lastSeen, &specs); err != nil {
			continue
		}
		nodes = append(nodes, map[string]interface{}{
			"id":        id,
			"type":      nType,
			"status":    status,
			"last_seen": lastSeen,
			"specs":     specs,
		})
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(nodes)
}
