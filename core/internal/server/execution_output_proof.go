package server

import (
	"crypto/sha256"
	"encoding/hex"
	"strings"

	"github.com/mycelis/core/pkg/protocol"
)

func attachConfirmActionOutputProofs(outputs []protocol.ExecutionOutput, proofArtifactID, runID, contractID string, results []plannedToolExecutionResult) []protocol.ExecutionOutput {
	for i := range outputs {
		output := &outputs[i]
		output.ProofArtifactID = proofArtifactID
		output.OpenURL = firstNonEmptyString(output.OpenURL, output.Href)
		proof := &protocol.OutputProofEnvelope{
			ProofID:            proofArtifactID,
			OutputRefID:        firstNonEmptyString(output.ID, output.Title),
			ArtifactID:         output.ArtifactID,
			StorageRef:         firstNonEmptyString(output.Href, output.Folder, output.Entrypoint, output.ID),
			SourceRunID:        runID,
			SourceContractID:   contractID,
			ExecutionStatus:    "verified",
			PathBoundaryStatus: "verified",
			ReadbackStatus:     "verified",
		}
		if checksum, bytes, contentType := outputContentProof(output, results); checksum != "" {
			proof.Checksum = checksum
			proof.ChecksumAlgorithm = "sha256"
			proof.Bytes = bytes
			proof.ContentType = contentType
		}
		if proof.StorageRef == "" {
			proof.PathBoundaryStatus = "not_applicable"
		}
		output.Proof = proof
	}
	return outputs
}

func outputContentProof(output *protocol.ExecutionOutput, results []plannedToolExecutionResult) (string, int64, string) {
	if output == nil {
		return "", 0, ""
	}
	outputRef := firstNonEmptyString(output.ID, output.Entrypoint, output.Folder, output.Title)
	for _, result := range results {
		path := firstNonEmptyString(result.Arguments["path"], result.Arguments["file_path"], result.Arguments["target_path"])
		if path != "" && (path == outputRef || path == output.ID || path == output.Entrypoint) {
			content := firstNonEmptyString(result.Arguments["content"], result.Arguments["body"], result.Arguments["text"])
			if content != "" {
				sum := sha256.Sum256([]byte(content))
				return hex.EncodeToString(sum[:]), int64(len([]byte(content))), contentTypeForOutput(output)
			}
		}
		for _, artifact := range result.Artifacts {
			if artifact.Content == "" {
				continue
			}
			if artifact.ID == output.ID || artifact.Title == output.Title || artifact.Entrypoint == output.Entrypoint {
				sum := sha256.Sum256([]byte(artifact.Content))
				return hex.EncodeToString(sum[:]), int64(len([]byte(artifact.Content))), firstNonEmptyString(artifact.ContentType, contentTypeForOutput(output))
			}
		}
	}
	return "", 0, ""
}

func contentTypeForOutput(output *protocol.ExecutionOutput) string {
	if output == nil {
		return ""
	}
	switch strings.ToLower(strings.TrimSpace(output.Kind)) {
	case "code", "file", "document":
		return "text/plain"
	case "project_package":
		return "application/vnd.mycelis.project-package+json"
	default:
		return ""
	}
}
