package server

import (
	"context"
	"database/sql/driver"
	"encoding/json"
	"strings"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/mycelis/core/internal/cognitive"
)

// expectAtomicDeploymentSave expects the M1 save: artifact and chunk rows in
// one transaction, the chunk embedding, the artifact status refresh, and the
// opportunistic backfill that follows a successful embed.
func expectAtomicDeploymentSave(mock sqlmock.Sqlmock, agentID, title, content string, artifactMeta sqlmock.Argument, artifactID string) {
	mock.ExpectBegin()
	mock.ExpectQuery("INSERT INTO artifacts").
		WithArgs(agentID, title, "text/markdown", content, artifactMeta).
		WillReturnRows(sqlmock.NewRows([]string{"id", "created_at"}).AddRow(artifactID, time.Now()))
	mock.ExpectQuery("INSERT INTO context_vectors").
		WithArgs(content, metadataContains{"embedding_status": "pending", "artifact_id": artifactID}).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow("11111111-1111-1111-1111-111111111111"))
	mock.ExpectCommit()
	mock.ExpectExec("UPDATE context_vectors SET embedding").WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectBegin()
	mock.ExpectExec("SELECT id FROM artifacts").WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectQuery("UPDATE artifacts a SET metadata").
		WillReturnRows(sqlmock.NewRows([]string{"id", "status", "embedded"}).AddRow(artifactID, "embedded", 1))
	mock.ExpectCommit()
	mock.ExpectBegin()
	mock.ExpectQuery("SELECT id::text, content, metadata FROM context_vectors").
		WillReturnRows(sqlmock.NewRows([]string{"id", "content", "metadata"}))
	mock.ExpectCommit()
	mock.ExpectQuery(`SELECT count\(\*\) FROM context_vectors`).WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(0))
}

type metadataContains map[string]any

func (m metadataContains) Match(v driver.Value) bool {
	var raw []byte
	switch value := v.(type) {
	case []byte:
		raw = value
	case string:
		raw = []byte(value)
	default:
		return false
	}

	var got map[string]any
	if err := json.Unmarshal(raw, &got); err != nil {
		return false
	}
	for key, want := range m {
		if !metadataValueContains(got[key], want) {
			return false
		}
	}
	return true
}

func metadataValueContains(got any, want any) bool {
	switch want := want.(type) {
	case string:
		return strings.TrimSpace(strings.ToLower(stringValue(got))) == strings.TrimSpace(strings.ToLower(want))
	case []string:
		values := normalizeStringSliceValue(got)
		if len(values) == 0 {
			return false
		}
		have := map[string]struct{}{}
		for _, value := range values {
			have[strings.ToLower(strings.TrimSpace(value))] = struct{}{}
		}
		for _, item := range want {
			if _, ok := have[strings.ToLower(strings.TrimSpace(item))]; !ok {
				return false
			}
		}
		return true
	case []any:
		expected := make([]string, 0, len(want))
		for _, item := range want {
			if text, ok := item.(string); ok && strings.TrimSpace(text) != "" {
				expected = append(expected, text)
			}
		}
		return metadataValueContains(got, expected)
	default:
		return false
	}
}

func normalizeStringSliceValue(value any) []string {
	switch typed := value.(type) {
	case []string:
		return typed
	case []any:
		out := make([]string, 0, len(typed))
		for _, item := range typed {
			if text, ok := item.(string); ok && strings.TrimSpace(text) != "" {
				out = append(out, text)
			}
		}
		return out
	default:
		return nil
	}
}

func stringValue(value any) string {
	if text, ok := value.(string); ok {
		return text
	}
	return ""
}

type fakeDeploymentContextProvider struct{ vec []float64 }

func (fakeDeploymentContextProvider) Infer(_ context.Context, _ string, _ cognitive.InferOptions) (*cognitive.InferResponse, error) {
	return &cognitive.InferResponse{Text: "ok", ModelUsed: "stub", Provider: "stub"}, nil
}

func (fakeDeploymentContextProvider) Probe(_ context.Context) (bool, error) {
	return true, nil
}

func (f fakeDeploymentContextProvider) Embed(_ context.Context, _ string, _ string) ([]float64, error) {
	return f.vec, nil
}

// newDeploymentContextBrain embeds with the 768-dim store width.
func newDeploymentContextBrain() *cognitive.Router {
	vec := make([]float64, 768)
	vec[0], vec[1] = 0.11, 0.22
	return newDeploymentContextBrainWithVector(vec)
}

func newDeploymentContextBrainWithVector(vec []float64) *cognitive.Router {
	return &cognitive.Router{
		Config: &cognitive.BrainConfig{
			Providers: map[string]cognitive.ProviderConfig{
				"stub": {Enabled: true, ModelID: "stub-model"},
			},
			Profiles: map[string]string{
				"chat":  "stub",
				"embed": "stub",
			},
		},
		Adapters: map[string]cognitive.LLMProvider{
			"stub": fakeDeploymentContextProvider{vec: vec},
		},
	}
}
