package server

import (
	"strings"

	"github.com/mycelis/core/pkg/protocol"
)

func configDocumentRequestBoundary(organizationID, teamID, operatorID string) *protocol.ConfigDocumentRequestBoundary {
	organizationID = strings.TrimSpace(organizationID)
	teamID = strings.TrimSpace(teamID)
	return &protocol.ConfigDocumentRequestBoundary{
		OrganizationID: organizationID,
		WorkspaceID:    firstNonEmptyString(organizationID, teamID),
		TeamID:         teamID,
		OperatorID:     strings.TrimSpace(operatorID),
	}
}
