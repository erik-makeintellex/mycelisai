package signal

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"
)

func (s *StreamHandler) replay(w http.ResponseWriter, r *http.Request) int64 {
	rawCursor := strings.TrimSpace(r.Header.Get("Last-Event-ID"))
	if rawCursor == "" {
		return 0
	}
	cursor, err := parseLastEventID(rawCursor)
	if err != nil {
		writeReplayGap(w, ReplayGap{Reason: "invalid_cursor", Replayable: s.store != nil})
		return 0
	}
	if s.store == nil {
		writeReplayGap(w, ReplayGap{Reason: "persistence_unavailable", RequestedAfter: cursor, Replayable: false})
		return cursor
	}

	batch, err := s.store.Replay(r.Context(), cursor, s.replayLimit)
	if err != nil {
		writeReplayGap(w, ReplayGap{Reason: "replay_unavailable", RequestedAfter: cursor, Replayable: true})
		return cursor
	}
	if batch.Gap != nil {
		writeReplayGap(w, *batch.Gap)
	}
	highWater := cursor
	for _, event := range batch.Events {
		writeSSEData(w, event.ID, "", event.Payload)
		if event.Sequence > highWater {
			highWater = event.Sequence
		}
	}
	return highWater
}

func (s *StreamHandler) follow(
	w http.ResponseWriter,
	r *http.Request,
	flusher http.Flusher,
	client <-chan streamClientMessage,
	highWater int64,
) {
	heartbeat := time.NewTicker(streamHeartbeatInterval)
	defer heartbeat.Stop()
	for {
		select {
		case <-r.Context().Done():
			return
		case <-heartbeat.C:
			if _, err := fmt.Fprint(w, ": keepalive\n\n"); err != nil {
				return
			}
			flusher.Flush()
		case message, open := <-client:
			if !open {
				return
			}
			if message.gap != nil {
				writeReplayGap(w, *message.gap)
				flusher.Flush()
				continue
			}
			if message.event == nil || (message.event.Sequence > 0 && message.event.Sequence <= highWater) {
				continue
			}
			writeSSEData(w, message.event.ID, "", message.event.Payload)
			if message.event.Sequence > highWater {
				highWater = message.event.Sequence
			}
			flusher.Flush()
		}
	}
}

func writeReplayGap(w http.ResponseWriter, gap ReplayGap) {
	raw, _ := json.Marshal(map[string]any{
		"type":            "replay_gap",
		"reason":          gap.Reason,
		"requested_after": gap.RequestedAfter,
		"first_replayed":  gap.FirstReplayed,
		"omitted":         gap.Omitted,
		"replayable":      gap.Replayable,
	})
	writeSSEData(w, "", "replay_gap", string(raw))
}

func writeSSEData(w http.ResponseWriter, id, eventName, payload string) {
	if id != "" {
		fmt.Fprintf(w, "id: %s\n", id)
	}
	if eventName != "" {
		fmt.Fprintf(w, "event: %s\n", eventName)
	}
	for _, line := range strings.Split(payload, "\n") {
		fmt.Fprintf(w, "data: %s\n", line)
	}
	fmt.Fprint(w, "\n")
}
