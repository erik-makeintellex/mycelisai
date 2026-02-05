package bootstrap

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/nats-io/nats.go"
)

// BootstrapService listens for new hardware announcements
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

func (s *Service) Start() {
	s.nc.Subscribe("swarm.bootstrap.announce", func(msg *nats.Msg) {
		s.processAnnouncement(msg.Data)
	})
	log.Println("👂 Bootstrap Listener Active (swarm.bootstrap.announce)")
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
