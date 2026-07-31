package signal

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"sync"
	"sync/atomic"
	"time"

	"github.com/google/uuid"
	"github.com/mycelis/core/internal/memory"
)

var streamHeartbeatInterval = 15 * time.Second

const defaultReplayLimit = 200

type streamClientMessage struct {
	event *StreamEvent
	gap   *ReplayGap
}

// StreamHandler manages durable SSE replay and live client delivery.
type StreamHandler struct {
	clients       map[chan streamClientMessage]struct{}
	store         eventPersistence
	replayLimit   int
	instanceID    string
	volatileCount atomic.Uint64
	broadcastMu   sync.Mutex
	mu            sync.RWMutex
}

// NewStreamHandler preserves nil-database operation for tests and degraded startup.
func NewStreamHandler(databases ...*sql.DB) *StreamHandler {
	var store eventPersistence
	if len(databases) > 0 && databases[0] != nil {
		store = NewEventStore(databases[0])
	}
	return &StreamHandler{
		clients:     make(map[chan streamClientMessage]struct{}),
		store:       store,
		replayLimit: defaultReplayLimit,
		instanceID:  uuid.NewString(),
	}
}

// HandleStream replays events after Last-Event-ID, then follows live events.
func (s *StreamHandler) HandleStream(w http.ResponseWriter, r *http.Request) {
	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, `{"error":"streaming not supported"}`, http.StatusInternalServerError)
		return
	}
	if !s.initialized() {
		w.Header().Set("Content-Type", "application/json")
		http.Error(w, `{"error":"stream handler not initialized"}`, http.StatusServiceUnavailable)
		return
	}

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")

	client := make(chan streamClientMessage, 64)
	s.addClient(client)
	defer s.removeClient(client)

	writeSSEData(w, "", "", fmt.Sprintf(
		`{"type":"connected","timestamp":%q}`, time.Now().UTC().Format(time.RFC3339),
	))
	flusher.Flush()

	highWater := s.replay(w, r)
	flusher.Flush()
	s.follow(w, r, flusher, client, highWater)
}

// Broadcast persists an operator event before making it visible to clients.
func (s *StreamHandler) Broadcast(payload string) {
	if !s.initialized() {
		return
	}
	payload = sanitizeOperatorPayload(payload)
	// Serialize append and fanout so concurrent callers cannot publish sequence N+1
	// before sequence N and cause reconnect cursors to skip an older event.
	s.broadcastMu.Lock()
	defer s.broadcastMu.Unlock()
	if s.store == nil {
		count := s.volatileCount.Add(1)
		event := StreamEvent{
			ID:        fmt.Sprintf("volatile-%s-%d", s.instanceID, count),
			Payload:   payload,
			CreatedAt: time.Now().UTC(),
		}
		s.publish(streamClientMessage{event: &event})
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	event, err := s.store.Append(ctx, payload)
	if err != nil {
		log.Printf("WARN: SSE event persistence failed; operator event withheld: %v", err)
		gap := ReplayGap{Reason: "persistence_unavailable", Replayable: false}
		s.publish(streamClientMessage{gap: &gap})
		return
	}
	s.publish(streamClientMessage{event: &event})
}

// BroadcastLogEntry formats a LogEntry for the stream.
func (s *StreamHandler) BroadcastLogEntry(entry *memory.LogEntry) {
	ctxJSON, _ := json.Marshal(entry.Context)
	jsonMsg := fmt.Sprintf(`{"type":"log","source":%q,"level":%q,"message":%q,"timestamp":%q,"context":%s}`,
		entry.Source, entry.Level, entry.Message, entry.Timestamp.Format(time.RFC3339), string(ctxJSON))
	s.Broadcast(jsonMsg)
}

func (s *StreamHandler) initialized() bool {
	if s == nil {
		return false
	}
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.clients != nil
}

func (s *StreamHandler) addClient(client chan streamClientMessage) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.clients[client] = struct{}{}
}

func (s *StreamHandler) removeClient(client chan streamClientMessage) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, exists := s.clients[client]; !exists {
		return
	}
	delete(s.clients, client)
	close(client)
	log.Println("SSE Client Disconnected")
}

func (s *StreamHandler) publish(message streamClientMessage) {
	s.mu.RLock()
	lagged := make([]chan streamClientMessage, 0)
	for client := range s.clients {
		select {
		case client <- message:
		default:
			lagged = append(lagged, client)
		}
	}
	s.mu.RUnlock()
	for _, client := range lagged {
		log.Println("WARN: SSE client buffer full; reconnecting client for durable replay")
		s.removeClient(client)
	}
}

func parseLastEventID(value string) (int64, error) {
	if value == "" {
		return 0, nil
	}
	id, err := strconv.ParseInt(value, 10, 64)
	if err != nil || id < 0 {
		return 0, fmt.Errorf("invalid SSE cursor")
	}
	return id, nil
}
