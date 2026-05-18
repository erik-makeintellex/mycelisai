package server

import (
	"database/sql"
	"net/http"
	"strings"

	"github.com/mycelis/core/internal/trust"
	"github.com/mycelis/core/pkg/protocol"
)

func (s *AdminServer) HandleListExecutionContracts(w http.ResponseWriter, r *http.Request) {
	if !validateTrustReadFilters(w, r, false) {
		return
	}
	store := trust.NewStore(s.getDB())
	items, err := store.ListExecutionContracts(r.Context(), trustListOptions(r, ""))
	if err != nil {
		respondAPIError(w, "Failed to list execution contracts: "+err.Error(), http.StatusServiceUnavailable)
		return
	}
	respondAPIJSON(w, http.StatusOK, protocol.NewAPISuccess(items))
}

func (s *AdminServer) HandleGetExecutionContract(w http.ResponseWriter, r *http.Request) {
	id := strings.TrimSpace(r.PathValue("id"))
	if err := validateOptionalUUID("id", id); err != nil || id == "" {
		respondAPIError(w, "id must be a UUID", http.StatusBadRequest)
		return
	}
	item, err := trust.NewStore(s.getDB()).GetExecutionContract(r.Context(), id)
	if err != nil {
		respondTrustReadError(w, "execution contract", err)
		return
	}
	respondAPIJSON(w, http.StatusOK, protocol.NewAPISuccess(item))
}

func (s *AdminServer) HandleListProofArtifacts(w http.ResponseWriter, r *http.Request) {
	if !validateTrustReadFilters(w, r, true) {
		return
	}
	store := trust.NewStore(s.getDB())
	items, err := store.ListProofArtifacts(r.Context(), trustListOptions(r, r.URL.Query().Get("contract_id")))
	if err != nil {
		respondAPIError(w, "Failed to list proof artifacts: "+err.Error(), http.StatusServiceUnavailable)
		return
	}
	respondAPIJSON(w, http.StatusOK, protocol.NewAPISuccess(items))
}

func (s *AdminServer) HandleGetProofArtifact(w http.ResponseWriter, r *http.Request) {
	id := strings.TrimSpace(r.PathValue("id"))
	if err := validateOptionalUUID("id", id); err != nil || id == "" {
		respondAPIError(w, "id must be a UUID", http.StatusBadRequest)
		return
	}
	item, err := trust.NewStore(s.getDB()).GetProofArtifact(r.Context(), id)
	if err != nil {
		respondTrustReadError(w, "proof artifact", err)
		return
	}
	respondAPIJSON(w, http.StatusOK, protocol.NewAPISuccess(item))
}

func trustListOptions(r *http.Request, contractID string) trust.ListOptions {
	q := r.URL.Query()
	return trust.ListOptions{
		Limit:         parseLimit(q.Get("limit"), 20),
		RunID:         q.Get("run_id"),
		IntentProofID: q.Get("intent_proof_id"),
		ContractID:    contractID,
		Status:        q.Get("status"),
	}
}

func validateTrustReadFilters(w http.ResponseWriter, r *http.Request, includeContract bool) bool {
	q := r.URL.Query()
	for _, key := range []string{"run_id", "intent_proof_id"} {
		if err := validateOptionalUUID(key, q.Get(key)); err != nil {
			respondAPIError(w, err.Error(), http.StatusBadRequest)
			return false
		}
	}
	if includeContract {
		if err := validateOptionalUUID("contract_id", q.Get("contract_id")); err != nil {
			respondAPIError(w, err.Error(), http.StatusBadRequest)
			return false
		}
	}
	return true
}

func respondTrustReadError(w http.ResponseWriter, label string, err error) {
	if err == sql.ErrNoRows {
		respondAPIError(w, label+" not found", http.StatusNotFound)
		return
	}
	respondAPIError(w, "Failed to read "+label+": "+err.Error(), http.StatusServiceUnavailable)
}
