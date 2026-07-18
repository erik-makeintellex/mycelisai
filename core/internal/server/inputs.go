package server

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"github.com/mycelis/core/internal/inputs"
	"github.com/mycelis/core/pkg/protocol"
)

func (s *AdminServer) inputService() *inputs.Service {
	if s.Inputs != nil {
		return s.Inputs
	}
	s.Inputs = inputs.NewService()
	return s.Inputs
}

func (s *AdminServer) HandleInputSources(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		respondAPIJSON(w, http.StatusOK, protocol.NewAPISuccess(s.inputService().List()))
	case http.MethodPost:
		var req inputs.SourceInput
		decoder := json.NewDecoder(r.Body)
		decoder.DisallowUnknownFields()
		if err := decoder.Decode(&req); err != nil {
			respondAPIError(w, "invalid input source request", http.StatusBadRequest)
			return
		}
		source, err := s.inputService().Add(r.Context(), req)
		if err != nil {
			respondAPIError(w, err.Error(), inputs.ErrorStatus(err))
			return
		}
		respondAPIJSON(w, http.StatusCreated, protocol.NewAPISuccess(source))
	default:
		respondAPIError(w, "method not allowed", http.StatusMethodNotAllowed)
	}
}

func (s *AdminServer) HandleInputSource(w http.ResponseWriter, r *http.Request) {
	id := strings.TrimSpace(r.PathValue("id"))
	switch r.Method {
	case http.MethodGet:
		source, err := s.inputService().Get(r.Context(), id)
		if err != nil {
			respondAPIError(w, err.Error(), inputs.ErrorStatus(err))
			return
		}
		respondAPIJSON(w, http.StatusOK, protocol.NewAPISuccess(source))
	case http.MethodPatch:
		var req inputs.SourceInput
		decoder := json.NewDecoder(r.Body)
		decoder.DisallowUnknownFields()
		if err := decoder.Decode(&req); err != nil {
			respondAPIError(w, "invalid input source request", http.StatusBadRequest)
			return
		}
		source, err := s.inputService().Update(r.Context(), id, req)
		if err != nil {
			respondAPIError(w, err.Error(), inputs.ErrorStatus(err))
			return
		}
		respondAPIJSON(w, http.StatusOK, protocol.NewAPISuccess(source))
	case http.MethodDelete:
		if err := s.inputService().Delete(r.Context(), id); err != nil {
			respondAPIError(w, err.Error(), inputs.ErrorStatus(err))
			return
		}
		respondAPIJSON(w, http.StatusOK, protocol.NewAPISuccess(map[string]any{"id": id, "deleted": true}))
	default:
		respondAPIError(w, "method not allowed", http.StatusMethodNotAllowed)
	}
}

func (s *AdminServer) HandleInputSourceBuffer(w http.ResponseWriter, r *http.Request) {
	id := strings.TrimSpace(r.PathValue("id"))
	limit := 50
	if raw := strings.TrimSpace(r.URL.Query().Get("limit")); raw != "" {
		parsed, err := strconv.Atoi(raw)
		if err != nil {
			respondAPIError(w, "limit must be a number", http.StatusBadRequest)
			return
		}
		limit = parsed
	}
	view, err := s.inputService().Buffer(
		r.Context(),
		id,
		strings.TrimSpace(r.URL.Query().Get("mode")),
		strings.TrimSpace(r.URL.Query().Get("channel_key")),
		limit,
	)
	if err != nil {
		respondAPIError(w, err.Error(), inputs.ErrorStatus(err))
		return
	}
	respondAPIJSON(w, http.StatusOK, protocol.NewAPISuccess(view))
}
