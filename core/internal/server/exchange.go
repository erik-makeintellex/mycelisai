package server

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/google/uuid"
	"github.com/mycelis/core/internal/exchange"
)

func (s *AdminServer) handleListExchangeFields(w http.ResponseWriter, r *http.Request) {
	if s.Exchange == nil {
		respondError(w, "exchange service not initialized", http.StatusServiceUnavailable)
		return
	}
	fields, err := s.Exchange.ListFields(r.Context())
	if err != nil {
		respondError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	respondJSON(w, fields)
}

func (s *AdminServer) handleListExchangeSchemas(w http.ResponseWriter, r *http.Request) {
	if s.Exchange == nil {
		respondError(w, "exchange service not initialized", http.StatusServiceUnavailable)
		return
	}
	schemas, err := s.Exchange.ListSchemas(r.Context())
	if err != nil {
		respondError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	respondJSON(w, schemas)
}

func (s *AdminServer) handleListExchangeChannels(w http.ResponseWriter, r *http.Request) {
	if s.Exchange == nil {
		respondError(w, "exchange service not initialized", http.StatusServiceUnavailable)
		return
	}
	channels, err := s.Exchange.ListChannels(r.Context())
	if err != nil {
		respondError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	respondJSON(w, channels)
}

func (s *AdminServer) handleListExchangeThreads(w http.ResponseWriter, r *http.Request) {
	if s.Exchange == nil {
		respondError(w, "exchange service not initialized", http.StatusServiceUnavailable)
		return
	}
	limit := parsePositiveInt(r.URL.Query().Get("limit"), 50)
	threads, err := s.Exchange.ListThreads(r.Context(), r.URL.Query().Get("channel"), r.URL.Query().Get("status"), limit)
	if err != nil {
		respondError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	respondJSON(w, threads)
}

func (s *AdminServer) handleCreateExchangeThread(w http.ResponseWriter, r *http.Request) {
	if s.Exchange == nil {
		respondError(w, "exchange service not initialized", http.StatusServiceUnavailable)
		return
	}
	var input exchange.CreateThreadInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		respondError(w, "invalid JSON body", http.StatusBadRequest)
		return
	}
	thread, err := s.Exchange.CreateThread(r.Context(), input)
	if err != nil {
		respondError(w, err.Error(), http.StatusBadRequest)
		return
	}
	w.WriteHeader(http.StatusCreated)
	respondJSON(w, thread)
}

func (s *AdminServer) handleListExchangeItems(w http.ResponseWriter, r *http.Request) {
	if s.Exchange == nil {
		respondError(w, "exchange service not initialized", http.StatusServiceUnavailable)
		return
	}
	limit := parsePositiveInt(r.URL.Query().Get("limit"), 50)
	var threadID *uuid.UUID
	if raw := r.URL.Query().Get("thread_id"); raw != "" {
		id, err := uuid.Parse(raw)
		if err != nil {
			respondError(w, "invalid thread_id", http.StatusBadRequest)
			return
		}
		threadID = &id
	}
	items, err := s.Exchange.ListItems(r.Context(), r.URL.Query().Get("channel"), threadID, limit)
	if err != nil {
		respondError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	respondJSON(w, items)
}

func (s *AdminServer) handleCreateExchangeItem(w http.ResponseWriter, r *http.Request) {
	if s.Exchange == nil {
		respondError(w, "exchange service not initialized", http.StatusServiceUnavailable)
		return
	}
	var input exchange.PublishInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		respondError(w, "invalid JSON body", http.StatusBadRequest)
		return
	}
	item, err := s.Exchange.Publish(r.Context(), input)
	if err != nil {
		respondError(w, err.Error(), http.StatusBadRequest)
		return
	}
	w.WriteHeader(http.StatusCreated)
	respondJSON(w, item)
}

func (s *AdminServer) handleSearchExchangeItems(w http.ResponseWriter, r *http.Request) {
	if s.Exchange == nil {
		respondError(w, "exchange service not initialized", http.StatusServiceUnavailable)
		return
	}
	query := r.URL.Query().Get("q")
	if query == "" {
		respondError(w, "query parameter 'q' is required", http.StatusBadRequest)
		return
	}
	limit := parsePositiveInt(r.URL.Query().Get("limit"), 5)
	results, err := s.Exchange.Search(r.Context(), query, limit)
	if err != nil {
		respondError(w, err.Error(), http.StatusBadGateway)
		return
	}
	respondJSON(w, results)
}

func parsePositiveInt(raw string, fallback int) int {
	if raw == "" {
		return fallback
	}
	parsed, err := strconv.Atoi(raw)
	if err != nil || parsed <= 0 {
		return fallback
	}
	return parsed
}
