package main

import (
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/nats-io/nats.go"
)

func main() {
	natsURL := os.Getenv("NATS_URL")
	if natsURL == "" {
		natsURL = nats.DefaultURL
	}

	nc, err := nats.Connect(natsURL)
	if err != nil {
		log.Fatal(err)
	}
	defer nc.Close()

	// 1. Spawn Team Alpha
	// We retry a few times in case backend is waking up
	spawnURL := "http://localhost:8081/api/swarm/teams"
	spawnBody := `{"id":"team-alpha","name":"Alpha Unit","type":"action"}`

	log.Println("Attempting to spawn Team Alpha...")
	resp, err := http.Post(spawnURL, "application/json", strings.NewReader(spawnBody))
	if err != nil {
		log.Printf("Failed to spawn team (might already exist or backend down): %v", err)
	} else {
		log.Printf("Spawn Team Status: %s", resp.Status)
		resp.Body.Close()
	}

	// 2. Send Trigger
	topic := "swarm.team.team-alpha.internal.trigger"
	log.Println("Injecting Trigger to Team Alpha...")

	// Triggers for the Agent to process via Ollama
	prompts := []string{
		"System check. Report status.",
		"What is your primary directive as an Action Unit?",
	}

	for _, p := range prompts {
		log.Printf("Triggering [%s]: %s", topic, p)
		if err := nc.Publish(topic, []byte(p)); err != nil {
			log.Println("Error publishing:", err)
		}
		// Give Ollama time to think and respond
		time.Sleep(10 * time.Second)
	}

	log.Println("Signal Injection Complete.")
}
