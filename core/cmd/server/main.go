package main

import (
	"log"
	"net/http"
	"os"
	"time"

	"github.com/nats-io/nats.go"

	"github.com/mycelis/core/internal/governance"
	"github.com/mycelis/core/internal/router"
	"github.com/mycelis/core/internal/server"
	mycelis_nats "github.com/mycelis/core/internal/transport/nats"
)

func main() {
	log.Println("Starting Mycelis Core [Brain]...")

	// 1. Config
	natsURL := os.Getenv("NATS_URL")
	if natsURL == "" {
		natsURL = nats.DefaultURL // nats://localhost:4222
	}

	// 1b. Load Governance Policy
	policyPath := "core/policy/policy.yaml" // Assuming we run from root or handle path
	// Fix path for container vs local? For now, relative.

	gk, err := governance.NewGatekeeper(policyPath)
	if err != nil {
		log.Printf("⚠️ Governance Policy not loaded: %v. Allowing all.", err)
		gk = nil
	} else {
		log.Println("🛡️ Gatekeeper Active.")
	}

	// 2. Connect to NATS
	var ncWrapper *mycelis_nats.Client
	var connErr error

	// Simple retry loop
	for i := 0; i < 10; i++ {
		log.Printf("Connecting to NATS at %s (Attempt %d/10)...", natsURL, i+1)
		ncWrapper, connErr = mycelis_nats.Connect(natsURL)
		if connErr == nil {
			break
		}
		log.Printf("NATS connection failed: %v. Retrying in 2s...", connErr)
		time.Sleep(2 * time.Second)
	}

	if connErr != nil {
		log.Fatalf("Failed to connect to NATS after retries: %v", connErr)
	}
	defer ncWrapper.Drain()

	nc := ncWrapper.Conn

	// 3. Start Router
	r := router.NewRouter(nc, gk)
	if err := r.Start(); err != nil {
		log.Fatalf("Failed to start Router: %v", err)
	}

	// 5. Http Server
	mux := http.NewServeMux()

	// Create Admin Server
	adminSrv := server.NewAdminServer(r, gk)
	adminSrv.RegisterRoutes(mux)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("HTTP Server listening on :%s", port)
	if err := http.ListenAndServe(":"+port, mux); err != nil {
		log.Fatalf("HTTP Server failed: %v", err)
	}
}
