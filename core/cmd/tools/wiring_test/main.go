package main

import (
	"log"
	"os"
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
		log.Fatalf("Failed to connect to NATS: %v", err)
	}
	defer nc.Close()

	inputSubject := "test.input.alpha"
	outputSubject := "test.output.omega"

	// 1. Subscribe to Output
	sub, err := nc.SubscribeSync(outputSubject)
	if err != nil {
		log.Fatalf("Failed to subscribe to output: %v", err)
	}
	log.Printf("🎧 Listening on [%s]...", outputSubject)

	// 2. Publish Input
	payload := []byte("Hello Neural Wiring")
	log.Printf("📢 Sending Trigger to [%s]: %s", inputSubject, string(payload))
	if err := nc.Publish(inputSubject, payload); err != nil {
		log.Fatalf("Failed to publish: %v", err)
	}

	// 3. Wait for Response
	msg, err := sub.NextMsg(10 * time.Second)
	if err != nil {
		log.Fatalf("❌ Verification Failed: No response received within 10s (%v)", err)
	}

	log.Printf("✅ Verified: Received response on [%s]: %s", msg.Subject, string(msg.Data))
}
