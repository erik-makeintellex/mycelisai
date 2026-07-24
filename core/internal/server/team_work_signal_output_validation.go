package server

import "github.com/mycelis/core/pkg/protocol"

func deliverableResultMissingOutputs(item protocol.TeamWorkItem, payloadKind protocol.SignalPayloadKind, outputRefs []protocol.TeamOutputRef) bool {
	return payloadKind == protocol.PayloadKindResult &&
		(item.ExecutionShape == protocol.TeamExecutionShapeDeliverable ||
			item.ExecutionShape == protocol.TeamExecutionShapeDelegatedWork) &&
		len(item.ExpectedOutputs) > 0 &&
		len(outputRefs) == 0
}
