package signal

import (
	"bufio"
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"
)

type fakeEventStore struct {
	mu          sync.Mutex
	replayAfter int64
	replayLimit int
	replayBatch ReplayBatch
	replayErr   error
	appendQueue []StreamEvent
	appendErr   error
}

func (f *fakeEventStore) Append(_ context.Context, payload string) (StreamEvent, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	if f.appendErr != nil {
		return StreamEvent{}, f.appendErr
	}
	if len(f.appendQueue) == 0 {
		return StreamEvent{}, errors.New("no fake event queued")
	}
	event := f.appendQueue[0]
	f.appendQueue = f.appendQueue[1:]
	event.Payload = payload
	return event, nil
}

func (f *fakeEventStore) Replay(_ context.Context, after int64, limit int) (ReplayBatch, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.replayAfter = after
	f.replayLimit = limit
	return f.replayBatch, f.replayErr
}

func TestHandleStreamReplaysAfterLastEventIDInOrder(t *testing.T) {
	store := &fakeEventStore{replayBatch: ReplayBatch{Events: []StreamEvent{
		{Sequence: 3, ID: "3", Payload: `{"type":"status","value":3}`},
		{Sequence: 4, ID: "4", Payload: `{"type":"status","value":4}`},
	}}}
	handler := NewStreamHandler()
	handler.store = store
	response, scanner, cancel := openTestStream(t, handler, "2")
	defer response.Body.Close()
	defer cancel()

	assertFrameDataContains(t, readFrame(t, scanner), `"type":"connected"`)
	first := readFrame(t, scanner)
	second := readFrame(t, scanner)
	if first["id"] != "3" || second["id"] != "4" {
		t.Fatalf("replay order = %q, %q; want 3, 4", first["id"], second["id"])
	}
	if store.replayAfter != 2 || store.replayLimit != defaultReplayLimit {
		t.Fatalf("replay request = after %d limit %d", store.replayAfter, store.replayLimit)
	}
}

func TestHandleStreamSuppressesReplayAndLiveDuplicate(t *testing.T) {
	store := &fakeEventStore{
		replayBatch: ReplayBatch{Events: []StreamEvent{{Sequence: 3, ID: "3", Payload: `{"value":3}`}}},
		appendQueue: []StreamEvent{{Sequence: 3, ID: "3"}, {Sequence: 4, ID: "4"}},
	}
	handler := NewStreamHandler()
	handler.store = store
	response, scanner, cancel := openTestStream(t, handler, "2")
	defer response.Body.Close()
	defer cancel()

	readFrame(t, scanner)
	if frame := readFrame(t, scanner); frame["id"] != "3" {
		t.Fatalf("replayed id = %q; want 3", frame["id"])
	}
	handler.Broadcast(`{"value":"duplicate"}`)
	handler.Broadcast(`{"value":4}`)
	frame := readFrame(t, scanner)
	if frame["id"] != "4" || strings.Contains(frame["data"], "duplicate") {
		t.Fatalf("next frame = %#v; want only event 4", frame)
	}
}

func TestHandleStreamReportsReplayFailureAndContinuesLive(t *testing.T) {
	store := &fakeEventStore{
		replayErr:   errors.New("database offline with private detail"),
		appendQueue: []StreamEvent{{Sequence: 10, ID: "10"}},
	}
	handler := NewStreamHandler()
	handler.store = store
	response, scanner, cancel := openTestStream(t, handler, "9")
	defer response.Body.Close()
	defer cancel()

	readFrame(t, scanner)
	gap := readFrame(t, scanner)
	if gap["event"] != "replay_gap" || !strings.Contains(gap["data"], "replay_unavailable") {
		t.Fatalf("gap frame = %#v", gap)
	}
	if strings.Contains(gap["data"], "private detail") {
		t.Fatal("raw persistence error leaked to SSE client")
	}
	handler.Broadcast(`{"type":"status","value":"live"}`)
	live := readFrame(t, scanner)
	if live["id"] != "10" || !strings.Contains(live["data"], `"value":"live"`) {
		t.Fatalf("live frame = %#v", live)
	}
}

func TestBroadcastWithholdsPayloadWhenPersistenceFails(t *testing.T) {
	handler := NewStreamHandler()
	handler.store = &fakeEventStore{appendErr: errors.New("write failed")}
	response, scanner, cancel := openTestStream(t, handler, "")
	defer response.Body.Close()
	defer cancel()

	readFrame(t, scanner)
	handler.Broadcast(`{"secret":"must-not-deliver"}`)
	gap := readFrame(t, scanner)
	if gap["event"] != "replay_gap" || strings.Contains(gap["data"], "must-not-deliver") {
		t.Fatalf("persistence failure frame = %#v", gap)
	}
}

func TestNilDatabaseStreamsVolatileEventAndReportsReconnectGap(t *testing.T) {
	handler := NewStreamHandler()
	response, scanner, cancel := openTestStream(t, handler, "8")
	defer response.Body.Close()
	defer cancel()

	readFrame(t, scanner)
	gap := readFrame(t, scanner)
	if !strings.Contains(gap["data"], "persistence_unavailable") {
		t.Fatalf("degraded gap = %#v", gap)
	}
	handler.Broadcast(`{"type":"status","value":"degraded-live"}`)
	live := readFrame(t, scanner)
	if !strings.HasPrefix(live["id"], "volatile-") {
		t.Fatalf("volatile id = %q", live["id"])
	}
}

func TestBroadcastDeliversSamePersistedEventToConcurrentClients(t *testing.T) {
	store := &fakeEventStore{appendQueue: []StreamEvent{{Sequence: 12, ID: "12"}}}
	handler := NewStreamHandler()
	handler.store = store
	one, scanOne, cancelOne := openTestStream(t, handler, "")
	defer one.Body.Close()
	defer cancelOne()
	two, scanTwo, cancelTwo := openTestStream(t, handler, "")
	defer two.Body.Close()
	defer cancelTwo()
	readFrame(t, scanOne)
	readFrame(t, scanTwo)

	handler.Broadcast(`{"type":"shared"}`)
	if a, b := readFrame(t, scanOne), readFrame(t, scanTwo); a["id"] != "12" || b["id"] != "12" {
		t.Fatalf("client ids = %q, %q; want 12", a["id"], b["id"])
	}
}

func TestPublishDisconnectsLaggedClientForReplay(t *testing.T) {
	handler := NewStreamHandler()
	client := make(chan streamClientMessage, 1)
	handler.addClient(client)

	handler.publish(streamClientMessage{event: &StreamEvent{ID: "1"}})
	handler.publish(streamClientMessage{event: &StreamEvent{ID: "2"}})

	handler.mu.RLock()
	_, connected := handler.clients[client]
	handler.mu.RUnlock()
	if connected {
		t.Fatal("lagged client remained connected after its replayable buffer filled")
	}
	if message, open := <-client; !open || message.event == nil || message.event.ID != "1" {
		t.Fatalf("buffered message = %#v, open=%v; want event 1 before close", message, open)
	}
	if _, open := <-client; open {
		t.Fatal("lagged client channel remained open")
	}
}

func TestZeroValueHandlerReturnsServiceUnavailable(t *testing.T) {
	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/api/v1/stream", nil)
	var handler StreamHandler

	handler.HandleStream(recorder, request)
	if recorder.Code != http.StatusServiceUnavailable {
		t.Fatalf("status = %d; want %d", recorder.Code, http.StatusServiceUnavailable)
	}
}

func TestStreamHandlerSendsHeartbeatDuringIdleWork(t *testing.T) {
	previous := streamHeartbeatInterval
	streamHeartbeatInterval = 10 * time.Millisecond
	defer func() { streamHeartbeatInterval = previous }()

	server := httptest.NewServer(http.HandlerFunc(NewStreamHandler().HandleStream))
	defer server.Close()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	req, _ := http.NewRequestWithContext(ctx, http.MethodGet, server.URL, nil)
	response, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("open stream: %v", err)
	}
	defer response.Body.Close()
	scanner := bufio.NewScanner(response.Body)
	for scanner.Scan() {
		if strings.TrimSpace(scanner.Text()) == ": keepalive" {
			return
		}
	}
	t.Fatal("stream ended before an idle heartbeat was sent")
}

func openTestStream(t *testing.T, handler *StreamHandler, lastID string) (*http.Response, *bufio.Scanner, context.CancelFunc) {
	t.Helper()
	server := httptest.NewServer(http.HandlerFunc(handler.HandleStream))
	t.Cleanup(server.Close)
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	req, _ := http.NewRequestWithContext(ctx, http.MethodGet, server.URL, nil)
	if lastID != "" {
		req.Header.Set("Last-Event-ID", lastID)
	}
	response, err := http.DefaultClient.Do(req)
	if err != nil {
		cancel()
		t.Fatalf("open stream: %v", err)
	}
	return response, bufio.NewScanner(response.Body), cancel
}

func readFrame(t *testing.T, scanner *bufio.Scanner) map[string]string {
	t.Helper()
	frame := map[string]string{}
	for scanner.Scan() {
		line := scanner.Text()
		if line == "" {
			return frame
		}
		if key, value, ok := strings.Cut(line, ":"); ok {
			frame[key] = strings.TrimSpace(value)
		}
	}
	if err := scanner.Err(); err != nil {
		t.Fatalf("read SSE frame: %v", err)
	}
	t.Fatal("SSE stream ended before frame completed")
	return nil
}

func assertFrameDataContains(t *testing.T, frame map[string]string, want string) {
	t.Helper()
	if !strings.Contains(frame["data"], want) {
		t.Fatalf("frame data = %q; want %q", frame["data"], want)
	}
}
