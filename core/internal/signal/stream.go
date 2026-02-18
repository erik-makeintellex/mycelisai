package signal

import (
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/mycelis/core/internal/memory"
)

// StreamHandler manages SSE connections and broadcasts events
type StreamHandler struct {
	clients map[chan string]bool
	mu      sync.RWMutex
}

func NewStreamHandler() *StreamHandler {
	return &StreamHandler{
		clients: make(map[chan string]bool),
	}
}

// HandleStream handles the SSE connection
func (s *StreamHandler) HandleStream(w http.ResponseWriter, r *http.Request) {
	// Verify the ResponseWriter supports streaming
	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, `{"error":"streaming not supported"}`, http.StatusInternalServerError)
		return
	}

	// Guard: if clients map is nil (zero-value struct), reject gracefully
	s.mu.Lock()
	if s.clients == nil {
		s.mu.Unlock()
		w.Header().Set("Content-Type", "application/json")
		http.Error(w, `{"error":"stream handler not initialized"}`, http.StatusServiceUnavailable)
		return
	}
	s.mu.Unlock()

	// Set SSE headers
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	// Create channel for this client
	clientChan := make(chan string, 8)

	// Register client
	s.mu.Lock()
	s.clients[clientChan] = true
	s.mu.Unlock()

	// Cleanup on disconnect
	defer func() {
		s.mu.Lock()
		delete(s.clients, clientChan)
		close(clientChan)
		s.mu.Unlock()
		log.Println("SSE Client Disconnected")
	}()

	log.Println("SSE Client Connected")

	// Send connected event
	fmt.Fprintf(w, "data: %s\n\n", `{"type": "connected", "timestamp": "`+time.Now().Format(time.RFC3339)+`"}`)
	flusher.Flush()

	// Stream loop — respects client disconnect via request context
	ctx := r.Context()
	for {
		select {
		case <-ctx.Done():
			return
		case msg, open := <-clientChan:
			if !open {
				return
			}
			fmt.Fprintf(w, "data: %s\n\n", msg)
			flusher.Flush()
		}
	}
}

// Broadcast sends a message to all connected clients
func (s *StreamHandler) Broadcast(msg string) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	for clientChan := range s.clients {
		select {
		case clientChan <- msg:
		default:
			// fast fail
		}
	}
}

// BroadcastLogEntry formats a LogEntry for the stream
func (s *StreamHandler) BroadcastLogEntry(entry *memory.LogEntry) {
	jsonMsg := fmt.Sprintf(`{"type": "log", "source": "%s", "level": "%s", "message": "%s", "timestamp": "%s"}`,
		entry.Source, entry.Level, entry.Message, entry.Timestamp.Format(time.RFC3339))
	s.Broadcast(jsonMsg)
}
