package main

import (
	"context"
	"log"
	"net/http"
	"os"
	os_signal "os/signal"
	"time"

	coreServer "github.com/mycelis/core/internal/server"
	"github.com/mycelis/core/internal/swarm"
	"github.com/nats-io/nats.go"
)

func main() {
	// Action CLI mode: use the same binary to call Mycelis API endpoints.
	// Example: server action GET /api/v1/services/status
	if len(os.Args) > 1 && os.Args[1] == "action" {
		if err := runActionCLI(os.Args[2:]); err != nil {
			log.Fatal(err)
		}
		return
	}

	log.Println("Starting Mycelis Core [System]...")

	ctx, stop := os_signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	apiKey := os.Getenv("MYCELIS_API_KEY")
	if apiKey == "" {
		log.Fatal("FATAL: MYCELIS_API_KEY not set. Server refuses to start without authentication.")
	}

	natsURL := os.Getenv("NATS_URL")
	if natsURL == "" {
		natsURL = nats.DefaultURL // nats://localhost:4222
	}

	core := startCoreRuntime(ctx, natsURL)
	defer core.DrainNATS()

	mux := http.NewServeMux()
	product := startProductRuntime(ctx, mux, core)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	corsOrigin := os.Getenv("CORS_ORIGIN")
	if corsOrigin == "" {
		corsOrigin = "http://localhost:3000"
	}

	srv := newHTTPServer(port, apiKey, corsOrigin, mux)
	startGracefulShutdown(ctx, srv, product)

	log.Printf("HTTP Server listening on :%s", port)
	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("HTTP Server failed: %v", err)
	}

	log.Println("Mycelis Core shutdown complete.")
}

func newHTTPServer(port, apiKey, corsOrigin string, mux *http.ServeMux) *http.Server {
	authedMux := coreServer.AuthMiddleware(apiKey, mux)
	corsMux := http.HandlerFunc(func(w http.ResponseWriter, hr *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", corsOrigin)
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS, PUT, DELETE")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		if hr.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		authedMux.ServeHTTP(w, hr)
	})

	return &http.Server{
		Addr:    ":" + port,
		Handler: corsMux,
	}
}

func startGracefulShutdown(ctx context.Context, srv *http.Server, runtime *productRuntime) {
	go func() {
		<-ctx.Done()
		log.Println("Shutdown signal received. Draining...")

		shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		if runtime != nil && runtime.Soma != nil {
			runtime.Soma.Shutdown()
		}

		if runtime != nil && runtime.Admin != nil {
			if runtime.Admin.TriggerEngine != nil {
				runtime.Admin.TriggerEngine.Stop()
			}
			runtime.Admin.StopLoopScheduler()
		}

		if runtime != nil && runtime.MCPPool != nil {
			runtime.MCPPool.ShutdownAll()
		}

		if err := srv.Shutdown(shutdownCtx); err != nil {
			log.Printf("HTTP shutdown error: %v", err)
		}
	}()
}

type productRuntime struct {
	Admin   *coreServer.AdminServer
	Soma    *swarm.Soma
	MCPPool mcpPoolShutdowner
}

type mcpPoolShutdowner interface {
	ShutdownAll()
}
