package httpapi

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/mycelis/framework-runs/internal/protocol"
)

func (server *Server) events(response http.ResponseWriter, request *http.Request) {
	runID := request.PathValue("run_id")
	if protocol.ValidateExternalID("run_id", runID) != nil {
		writeError(response, http.StatusNotFound, "run_not_found", "Run was not found.", false)
		return
	}
	cursor, err := decimalCursor(request.Header)
	if err != nil {
		writeError(response, http.StatusUnprocessableEntity, "invalid_cursor", "Last-Event-ID must be a decimal sequence.", false)
		return
	}
	initial, err := server.service.Events(request.Context(), runID, cursor)
	if err != nil {
		writeServiceError(response, err)
		return
	}
	flusher, ok := response.(http.Flusher)
	if !ok {
		writeError(response, http.StatusServiceUnavailable, "unready", "Event streaming is unavailable.", true)
		return
	}
	response.Header().Set("Content-Type", "text/event-stream")
	response.Header().Set("Cache-Control", "no-cache")
	response.Header().Set("X-Accel-Buffering", "no")
	response.WriteHeader(http.StatusOK)

	lastStatus := protocol.Status("")
	writeEvents := func(events []protocol.Event) bool {
		for _, event := range events {
			raw, marshalErr := json.Marshal(event)
			if marshalErr != nil {
				return false
			}
			_, _ = fmt.Fprintf(response, "id: %d\ndata: %s\n\n", event.Sequence, raw)
			cursor = event.Sequence
			lastStatus = event.Status
		}
		flusher.Flush()
		return true
	}
	if !writeEvents(initial) || terminal(lastStatus) {
		return
	}
	if snapshot, getErr := server.service.Get(request.Context(), runID); getErr != nil || terminal(snapshot.Status) {
		return
	}

	poll := time.NewTicker(50 * time.Millisecond)
	heartbeat := time.NewTicker(15 * time.Second)
	defer poll.Stop()
	defer heartbeat.Stop()
	for {
		select {
		case <-request.Context().Done():
			return
		case <-heartbeat.C:
			_, _ = response.Write([]byte(": keepalive\n\n"))
			flusher.Flush()
		case <-poll.C:
			next, eventsErr := server.service.Events(request.Context(), runID, cursor)
			if eventsErr != nil || !writeEvents(next) || terminal(lastStatus) {
				return
			}
		}
	}
}
