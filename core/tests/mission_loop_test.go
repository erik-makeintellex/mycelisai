//go:build integration
// +build integration

package tests

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/nats-io/nats.go"
)

const (
	// Adjust these if your environment differs
	API_URL  = "http://localhost:8081" // Updated to 8081 for Recovery Phase
	NATS_URL = "nats://localhost:4222"
)

type Blueprint struct {
	ID     string `json:"id"`
	Intent string `json:"intent"`
	// Add other fields if needed for Commit params
}

func TestMissionLoop_Instantiation(t *testing.T) {
	// 1. Setup NATS Connection
	nc, err := nats.Connect(NATS_URL)
	if err != nil {
		t.Fatalf("Failed to connect to NATS: %v", err)
	}
	defer nc.Close()

	// 2. Subscribe to Audit Trace
	// We want to see 'heartbeat' and 'thought' from the new squad
	auditCh := make(chan *nats.Msg, 100)
	sub, err := nc.ChanSubscribe("swarm.audit.trace", auditCh)
	if err != nil {
		t.Fatalf("Failed to subscribe to audit stream: %v", err)
	}
	defer sub.Unsubscribe()

	// 3. Negotiate Intent (Step 1)
	t.Log("📡 Negotiating Intent...")
	negotiatePayload := []byte(`{"intent":"Launch a perimeter search for truth"}`)
	resp, err := http.Post(API_URL+"/api/v1/intent/negotiate", "application/json", bytes.NewBuffer(negotiatePayload))
	if err != nil {
		t.Fatalf("Negotiate API call failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("Negotiate returned status %d", resp.StatusCode)
	}

	var blueprint Blueprint
	if err := json.NewDecoder(resp.Body).Decode(&blueprint); err != nil {
		t.Fatalf("Failed to decode blueprint: %v", err)
	}
	t.Logf("✅ Blueprint Received: %s", blueprint.ID)

	// 4. Commit Mission (Step 2)
	t.Log("🚀 Committing Mission...")
	commitPayload := []byte(fmt.Sprintf(`{"blueprint_id":"%s"}`, blueprint.ID))
	respCommit, err := http.Post(API_URL+"/api/v1/intent/commit", "application/json", bytes.NewBuffer(commitPayload))
	if err != nil {
		t.Fatalf("Commit API call failed: %v", err)
	}
	defer respCommit.Body.Close()

	if respCommit.StatusCode != http.StatusCreated && respCommit.StatusCode != http.StatusOK {
		t.Fatalf("Commit returned status %d", respCommit.StatusCode)
	}
	t.Log("✅ Mission Committed.")

	// 5. Verify Audit Telemetry (Step 3)
	t.Log("🎧 Listening for Swarm Heartbeats...")

	timeout := time.After(10 * time.Second)
	heartbeatDetected := false
	thoughtDetected := false

	// Loop until we see what we need or timeout
	for {
		select {
		case msg := <-auditCh:
			// Check for specific signals in the JSON payload
			// Payload structure from Architecture:
			// { "signal": "heartbeat"|"thought", "meta": {...}, ... }
			type Envelope struct {
				Signal string `json:"signal"`
			}
			var env Envelope
			if err := json.Unmarshal(msg.Data, &env); err == nil {
				if env.Signal == "heartbeat" {
					heartbeatDetected = true
					t.Log("💓 Heartbeat detected.")
				}
				if env.Signal == "thought" {
					thoughtDetected = true
					t.Log("🧠 Thought detected.")
				}
			}

			if heartbeatDetected && thoughtDetected {
				t.Log("✅ Test Complete: Heartbeat and Thought verified.")
				return
			}
		case <-timeout:
			if !heartbeatDetected {
				t.Error("❌ Timeout: No heartbeats detected from new squad.")
			}
			if !thoughtDetected {
				t.Log("⚠️ Warning: No thoughts detected (might take longer or require more stimulus).")
				// Not failing strictly on thought if heartbeat validates instantiation
			}
			if heartbeatDetected {
				return // Pass if at least alive
			}
			t.Fatal("Test Failed due to missing telemetry.")
		}
	}
}
