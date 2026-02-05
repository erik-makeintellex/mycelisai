package bootstrap

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/nats-io/nats.go"
)

type Service struct {
	db *sql.DB
	nc *nats.Conn
}

type Node struct {
	ID           string    `json:"id"`
	Status       string    `json:"status"`
	Capabilities []string  `json:"capabilities"`
	LastSeen     time.Time `json:"last_seen"`
}

func NewService(db *sql.DB, nc *nats.Conn) *Service {
	return &Service{
		db: db,
		nc: nc,
	}
}

func (s *Service) Start() {
	// Subscribe to Announcement
	_, err := s.nc.Subscribe("swarm.bootstrap.announce", func(msg *nats.Msg) {
		var payload struct {
			ID           string   `json:"id"`
			Capabilities []string `json:"capabilities"`
		}
		if err := json.Unmarshal(msg.Data, &payload); err != nil {
			log.Printf("Bootstrap: Malformed announcement: %v", err)
			return
		}

		log.Printf("Bootstrap: Detected Node %s", payload.ID)

		// Check DB
		var exists string
		err := s.db.QueryRow("SELECT id FROM nodes WHERE id = $1", payload.ID).Scan(&exists)
		if err == sql.ErrNoRows {
			// Insert New
			capsJSON, _ := json.Marshal(payload.Capabilities)
			_, err := s.db.Exec("INSERT INTO nodes (id, status, capabilities) VALUES ($1, 'pending', $2)", payload.ID, capsJSON)
			if err != nil {
				log.Printf("Bootstrap: Failed to insert node: %v", err)
				return
			}

			// Publish Event for UI
			event := map[string]interface{}{
				"type":    "system",
				"message": "New Hardware Detected: " + payload.ID,
			}
			eventBytes, _ := json.Marshal(event)
			s.nc.Publish("swarm.system.events", eventBytes)

		} else if err != nil {
			log.Printf("Bootstrap: DB Error: %v", err)
		} else {
			// Update Last Seen
			s.db.Exec("UPDATE nodes SET last_seen = NOW() WHERE id = $1", payload.ID)
		}
	})

	if err != nil {
		log.Printf("Bootstrap: Failed to subscribe to NATS: %v", err)
	}
}

func (s *Service) HandlePendingNodes(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	rows, err := s.db.Query("SELECT id, status, capabilities, last_seen FROM nodes WHERE status = 'pending'")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	nodes := []Node{}
	for rows.Next() {
		var n Node
		var capsJSON []byte
		if err := rows.Scan(&n.ID, &n.Status, &capsJSON, &n.LastSeen); err != nil {
			continue
		}
		json.Unmarshal(capsJSON, &n.Capabilities)
		nodes = append(nodes, n)
	}

	json.NewEncoder(w).Encode(nodes)
}
