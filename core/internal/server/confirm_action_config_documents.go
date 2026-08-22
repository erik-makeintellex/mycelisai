package server

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/mycelis/core/internal/configdocuments"
)

func (s *AdminServer) claimConfirmedConfigDocuments(
	ctx context.Context,
	tx *sql.Tx,
	fixtureScopeID string,
	results []plannedToolExecutionResult,
) error {
	if strings.TrimSpace(fixtureScopeID) == "" {
		return nil
	}
	resources, err := configDocumentFixtureResources(results)
	if err != nil || len(resources) == 0 {
		return err
	}
	for _, resource := range resources {
		if err := claimQAFixtureResourceTx(ctx, tx, fixtureScopeID, resource); err != nil {
			return err
		}
	}
	return nil
}

func configDocumentFixtureResources(results []plannedToolExecutionResult) ([]qaFixtureResource, error) {
	resources := make([]qaFixtureResource, 0)
	for _, result := range results {
		switch {
		case strings.EqualFold(strings.TrimSpace(result.Name), "store_config_document"):
			receipt, ok := configDocumentReceiptFromResult(result)
			if !ok {
				return nil, fmt.Errorf("stored ConfigDocument result did not include retained revision identity")
			}
			resources = append(resources, qaFixtureResource{Kind: "config_document", Ref: receipt.RecordID})
		case strings.EqualFold(strings.TrimSpace(result.Name), "activate_config_document"):
			var payload struct {
				Activation configdocuments.ActivationResult `json:"activation"`
			}
			if err := json.Unmarshal([]byte(result.Output), &payload); err != nil || payload.Activation.HistoryID == "" {
				return nil, fmt.Errorf("activated ConfigDocument result did not include exact activation identity")
			}
			resources = append(resources, qaFixtureResource{
				Kind: "config_document", Ref: configDocumentActivationFixtureRef(payload.Activation.HistoryID),
			})
		}
	}
	if len(resources) == 0 {
		return nil, nil
	}
	return resources, nil
}
