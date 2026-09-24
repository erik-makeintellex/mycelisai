package server

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"

	"github.com/mycelis/core/internal/identity"
	"github.com/mycelis/core/pkg/protocol"
)

type teamOwnershipRequest struct {
	GroupID                string `json:"group_id"`
	ExpectedManifestDigest string `json:"expected_manifest_digest"`
	ExpectedRevision       string `json:"expected_revision"`
	Reason                 string `json:"reason"`
}

func (s *AdminServer) HandleTeamOwnership(w http.ResponseWriter, r *http.Request) {
	caller := IdentityFromContext(r.Context())
	if caller == nil {
		respondAPIError(w, "Authentication required", http.StatusUnauthorized)
		return
	}
	if s.DB == nil {
		respondAPIError(w, "Team ownership persistence unavailable", http.StatusServiceUnavailable)
		return
	}
	teamID := r.PathValue("teamID")
	var groupID, digest, revision, reason string
	if r.Method == http.MethodGet {
		groupID = r.URL.Query().Get("group_id")
		if len(r.URL.Query()) != 1 || len(r.URL.Query()["group_id"]) != 1 {
			respondAPIError(w, "Invalid team ownership request", http.StatusBadRequest)
			return
		}
	} else {
		r.Body = http.MaxBytesReader(w, r.Body, 4096)
		decoder := json.NewDecoder(r.Body)
		decoder.DisallowUnknownFields()
		var input teamOwnershipRequest
		if err := decoder.Decode(&input); err != nil || decoder.Decode(new(any)) != io.EOF {
			respondAPIError(w, "Invalid team ownership request", http.StatusBadRequest)
			return
		}
		groupID, digest, revision, reason = input.GroupID, input.ExpectedManifestDigest, input.ExpectedRevision, input.Reason
	}
	store := identity.NewTeamOwnershipStore(s.DB)
	var value identity.TeamOwnership
	var err error
	switch r.Method {
	case http.MethodGet:
		value, err = store.Inspect(r.Context(), caller.UserID, teamID, groupID)
	case http.MethodPut:
		value, err = store.Bind(r.Context(), caller.UserID, teamID, groupID, digest, revision, reason)
	case http.MethodDelete:
		value, err = store.Revoke(r.Context(), caller.UserID, teamID, groupID, digest, revision, reason)
	default:
		respondAPIError(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	if err != nil {
		status, message := http.StatusInternalServerError, "Team ownership operation unavailable"
		switch {
		case errors.Is(err, identity.ErrTeamOwnershipDenied):
			status, message = http.StatusForbidden, "Team ownership authority denied"
		case errors.Is(err, identity.ErrTeamOwnershipMissing):
			status, message = http.StatusNotFound, "Runtime team manifest not found"
		case errors.Is(err, identity.ErrTeamOwnershipConflict):
			status, message = http.StatusConflict, "Team ownership precondition conflict"
		case errors.Is(err, identity.ErrTeamOwnershipInvalid):
			status, message = http.StatusBadRequest, "Invalid team ownership request"
		}
		respondAPIError(w, message, status)
		return
	}
	respondAPIJSON(w, http.StatusOK, protocol.NewAPISuccess(value))
}
