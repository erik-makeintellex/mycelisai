package server

import (
	"context"
	"database/sql/driver"
	"encoding/json"
	"strings"

	"github.com/mycelis/core/internal/cognitive"
)

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

type fakeDeploymentContextProvider struct{}

func (fakeDeploymentContextProvider) Infer(_ context.Context, _ string, _ cognitive.InferOptions) (*cognitive.InferResponse, error) {
	return &cognitive.InferResponse{Text: "ok", ModelUsed: "stub", Provider: "stub"}, nil
}

func (fakeDeploymentContextProvider) Probe(_ context.Context) (bool, error) {
	return true, nil
}

func (fakeDeploymentContextProvider) Embed(_ context.Context, _ string, _ string) ([]float64, error) {
	return []float64{0.11, 0.22}, nil
}

func newDeploymentContextBrain() *cognitive.Router {
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
			"stub": fakeDeploymentContextProvider{},
		},
	}
}
