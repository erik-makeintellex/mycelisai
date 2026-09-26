package server

import (
	"database/sql"
	"errors"
	"net/http"
	"regexp"
	"strings"

	"github.com/google/uuid"
	"github.com/mycelis/core/internal/configdocuments"
	"github.com/mycelis/core/pkg/protocol"
)

var configDocumentStoredDigestPattern = regexp.MustCompile(`^sha256:[0-9a-f]{64}$`)

const configDocumentExportRedactionNotice = "Some values were hidden. This content is not the stored revision and is not a re-importable copy; hidden values stay in .env or the configured secret backend."

type configDocumentExportResponse struct {
	RecordID     string `json:"record_id"`
	StoredDigest string `json:"stored_digest,omitempty"`
	protocol.ConfigDocumentExport
	ContentMatchesStoredDigest bool   `json:"content_matches_stored_digest"`
	Notice                     string `json:"notice,omitempty"`
}

// HandleExportConfigDocument renders one stored revision as redacted YAML or
// JSON for read-only viewing. It performs no writes, activation, publish, or
// audit mutation, and never echoes document fragments on error.
func (s *AdminServer) HandleExportConfigDocument(w http.ResponseWriter, r *http.Request) {
	if _, ok := requireRootAdminScope(w, r, "config_documents:read"); !ok {
		return
	}
	query := r.URL.Query()
	if values, present := query["format"]; present && len(values) != 1 {
		respondAPIError(w, "Unsupported export format; use yaml or json", http.StatusBadRequest)
		return
	}
	format, err := protocol.ParseConfigDocumentExportFormat(query.Get("format"))
	if err != nil || (query.Has("format") && query.Get("format") == "") {
		respondAPIError(w, "Unsupported export format; use yaml or json", http.StatusBadRequest)
		return
	}
	recordID, ok := parseConfigDocumentRecordID(r.PathValue("recordId"))
	if !ok {
		respondAPIError(w, "Invalid config document record id", http.StatusBadRequest)
		return
	}
	store, ok := s.configDocumentStore(w)
	if !ok {
		return
	}
	record, err := store.GetRevision(r.Context(), "default", recordID)
	if err != nil {
		respondConfigDocumentExportError(w, err)
		return
	}
	if !exportableConfigDocumentKind(record.Document.Kind) {
		respondAPIError(w, "Config document kind is not exportable", http.StatusBadRequest)
		return
	}
	export, err := protocol.RenderConfigDocumentExport(record.Document, format)
	if err != nil {
		respondAPIError(w, "Config document export failed", http.StatusInternalServerError)
		return
	}
	response := configDocumentExportResponse{RecordID: recordID, ConfigDocumentExport: export}
	if export.RedactionApplied {
		// Lead decision T3: a redacted export is never presented beside the
		// stored digest, so stored_digest is omitted and the match stays false.
		response.Notice = configDocumentExportRedactionNotice
	} else if configDocumentStoredDigestPattern.MatchString(record.Digest) {
		response.StoredDigest = record.Digest
		response.ContentMatchesStoredDigest = storedDigestMatchesExport(format, export.Content, record.Digest)
	}
	respondAPIJSON(w, http.StatusOK, protocol.NewAPISuccess(response))
}

func parseConfigDocumentRecordID(raw string) (string, bool) {
	raw = strings.TrimSpace(raw)
	if len(raw) != 36 {
		return "", false
	}
	parsed, err := uuid.Parse(raw)
	if err != nil {
		return "", false
	}
	return parsed.String(), true
}

func exportableConfigDocumentKind(kind protocol.ConfigDocumentKind) bool {
	switch kind {
	case protocol.ConfigDocumentKindOutcomeTemplate, protocol.ConfigDocumentKindWorkerProfile, protocol.ConfigDocumentKindCodeContextSource:
		return true
	}
	return false
}

// storedDigestMatchesExport re-parses the unredacted content through the
// governed parser and recomputes the canonical digest. Any failure is false.
func storedDigestMatchesExport(format protocol.ConfigDocumentExportFormat, content, storedDigest string) bool {
	parsed, err := configdocuments.ParseDocument([]byte(content), string(format))
	if err != nil {
		return false
	}
	digest, err := protocol.CanonicalConfigDocumentDigest(parsed)
	return err == nil && digest == storedDigest
}

// respondConfigDocumentExportError maps store failures to fixed messages so
// driver or decode errors can never carry document fragments to the client.
func respondConfigDocumentExportError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, configdocuments.ErrRevisionNotFound), errors.Is(err, sql.ErrNoRows):
		respondAPIError(w, "Config document not found", http.StatusNotFound)
	case strings.Contains(err.Error(), "database not available"):
		respondAPIError(w, "Configuration database unavailable", http.StatusServiceUnavailable)
	default:
		respondAPIError(w, "Config document export failed", http.StatusInternalServerError)
	}
}
