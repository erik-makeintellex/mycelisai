package server

func executionSummaryHasVerifiedOutputProof(outputs []any, kind, id, proofID, checksum string) bool {
	for _, raw := range outputs {
		output, ok := raw.(map[string]any)
		if !ok {
			continue
		}
		if output["kind"] != kind || output["id"] != id {
			continue
		}
		if output["proof_artifact_id"] != proofID || output["open_url"] != "/api/v1/workspace/files/view?path=output%2Fconfirmed.txt" {
			return false
		}
		proof, ok := output["proof"].(map[string]any)
		if !ok {
			return false
		}
		return proof["proof_id"] == proofID &&
			proof["source_run_id"] != "" &&
			proof["source_contract_id"] == "33333333-3333-3333-3333-333333333333" &&
			proof["path_boundary_status"] == "verified" &&
			proof["readback_status"] == "verified" &&
			proof["checksum_algorithm"] == "sha256" &&
			proof["checksum"] == checksum
	}
	return false
}
