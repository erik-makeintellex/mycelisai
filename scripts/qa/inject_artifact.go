package main

import (
	"encoding/json"
	"log"

	"github.com/nats-io/nats.go"
)

const (
	NATS_URL = "nats://127.0.0.1:4222"
	TOPIC    = "swarm.audit.trace"
)

func main() {
	nc, err := nats.Connect(NATS_URL)
	if err != nil {
		log.Fatalf("❌ NATS Connect Failed: %v", err)
	}
	defer nc.Close()

	payload := map[string]interface{}{
		"meta": map[string]interface{}{
			"source_node": "writer-agent",
			"team_id":     "test-uuid",
		},
		"signal": "artifact",
		"payload": map[string]interface{}{
			"content":     "# Market Analysis\n\nPricing is optimal.",
			"trust_score": 0.99,
		},
	}

	data, _ := json.Marshal(payload)
	if err := nc.Publish(TOPIC, data); err != nil {
		log.Fatalf("❌ Publish Failed: %v", err)
	}
	log.Printf("✅ Injected Artifact: %s", string(data))

	// Flush to ensure it's sent before exit
	nc.Flush()
}
