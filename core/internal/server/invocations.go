package server

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"

	"github.com/mycelis/core/internal/invocation"
	"github.com/mycelis/core/pkg/protocol"
)

// Invocation proposals carry no identity, adapter endpoint or executable grant.
func decodeInvocationRequest(w http.ResponseWriter, r *http.Request, dst any) bool {
	r.Body = http.MaxBytesReader(w, r.Body, 8192)
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(dst); err != nil {
		respondAPIError(w, "Invalid invocation request", http.StatusBadRequest)
		return false
	}
	if decoder.Decode(new(any)) != io.EOF {
		respondAPIError(w, "Expected one request object", http.StatusBadRequest)
		return false
	}
	return true
}

func (s *AdminServer) invocationCaller(w http.ResponseWriter, r *http.Request) (string, bool) {
	identity := IdentityFromContext(r.Context())
	if identity == nil {
		respondAPIError(w, "Authentication required", http.StatusUnauthorized)
		return "", false
	}
	if s.Invocations == nil || s.InvocationAdmissionUnavailable {
		respondAPIError(w, "Invocation persistence unavailable", http.StatusServiceUnavailable)
		return "", false
	}
	return identity.UserID, true
}

func invocationResponse(w http.ResponseWriter, value any, err error) {
	if err == nil {
		respondAPIJSON(w, http.StatusOK, protocol.APIResponse{OK: true, Data: value})
		return
	}
	status, message := http.StatusInternalServerError, "Invocation operation unavailable"
	switch {
	case errors.Is(err, invocation.ErrDenied):
		status, message = http.StatusForbidden, "Invocation authority denied"
	case errors.Is(err, invocation.ErrConflict):
		status, message = http.StatusConflict, "Invocation state or request conflict"
	case errors.Is(err, invocation.ErrInvalid):
		status, message = http.StatusBadRequest, "Invalid invocation request"
	}
	respondAPIError(w, message, status)
}

func (s *AdminServer) HandleInvocationProposal(w http.ResponseWriter, r *http.Request) {
	user, ok := s.invocationCaller(w, r)
	if !ok {
		return
	}
	var proposal invocation.Proposal
	if !decodeInvocationRequest(w, r, &proposal) {
		return
	}
	result, err := s.Invocations.Propose(r.Context(), user, proposal)
	invocationResponse(w, result, err)
}

// A model-generated function call is parsed as a proposal, never dispatched.
func (s *AdminServer) HandleInvocationToolProposal(w http.ResponseWriter, r *http.Request) {
	user, ok := s.invocationCaller(w, r)
	if !ok {
		return
	}
	var call struct {
		Name      string              `json:"name"`
		Arguments invocation.Proposal `json:"arguments"`
	}
	if !decodeInvocationRequest(w, r, &call) {
		return
	}
	if call.Name != "counting.increment" {
		respondAPIError(w, "Unsupported proposal capability", http.StatusBadRequest)
		return
	}
	result, err := s.Invocations.Propose(r.Context(), user, call.Arguments)
	invocationResponse(w, result, err)
}

func (s *AdminServer) HandleInvocation(w http.ResponseWriter, r *http.Request) {
	user, ok := s.invocationCaller(w, r)
	if !ok {
		return
	}
	var request struct {
		GrantID        string           `json:"grant_id"`
		GrantDigest    string           `json:"grant_digest"`
		IdempotencyKey string           `json:"idempotency_key"`
		Input          invocation.Input `json:"input"`
	}
	if !decodeInvocationRequest(w, r, &request) {
		return
	}
	result, _, err := s.Invocations.Admit(r.Context(), user, request.GrantID, request.GrantDigest, request.IdempotencyKey, request.Input)
	if err == nil {
		result, err = s.Invocations.Execute(r.Context(), result.ID)
	}
	invocationResponse(w, result, err)
}

func (s *AdminServer) HandleInvocationRead(w http.ResponseWriter, r *http.Request) {
	user, ok := s.invocationCaller(w, r)
	if !ok {
		return
	}
	result, err := s.Invocations.Get(r.Context(), user, r.PathValue("id"))
	invocationResponse(w, result, err)
}

func (s *AdminServer) HandleInvocationReconcile(w http.ResponseWriter, r *http.Request) {
	user, ok := s.invocationCaller(w, r)
	if !ok {
		return
	}
	var request struct {
		Observation string              `json:"observation"`
		Evidence    invocation.Evidence `json:"evidence"`
	}
	if !decodeInvocationRequest(w, r, &request) {
		return
	}
	result, err := s.Invocations.Reconcile(r.Context(), user, r.PathValue("id"), request.Observation, request.Evidence)
	invocationResponse(w, result, err)
}

func (s *AdminServer) confirmInvocation(w http.ResponseWriter, r *http.Request, token string) bool {
	if s.Invocations == nil {
		return false
	}
	handles, err := s.Invocations.HandlesToken(r.Context(), token)
	if err != nil {
		invocationResponse(w, nil, err)
		return true
	}
	if !handles {
		return false
	}
	user, ok := s.invocationCaller(w, r)
	if !ok {
		return true
	}
	grant, err := s.Invocations.Confirm(r.Context(), user, token)
	invocationResponse(w, grant, err)
	return true
}
