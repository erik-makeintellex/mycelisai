package server

import (
	"encoding/json"
	"net/http"

	"github.com/mycelis/core/pkg/protocol"
)

// respondAPIJSON writes a protocol.APIResponse as JSON with an explicit status code.
func respondAPIJSON(w http.ResponseWriter, status int, resp protocol.APIResponse) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(resp)
}

// respondAPIError writes a structured APIResponse error.
func respondAPIError(w http.ResponseWriter, msg string, status int) {
	respondAPIJSON(w, status, protocol.NewAPIError(msg))
}
