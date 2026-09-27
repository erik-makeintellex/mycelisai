package protocol

// ContextSourceRef names one governed context source that was injected into
// the model's context for a reply. Used is true only when the reply shares
// wording with the source that the request did not already contain; other
// refs were consulted, not used.
type ContextSourceRef struct {
	ArtifactID     string `json:"artifact_id"`
	Title          string `json:"title"`
	KnowledgeClass string `json:"knowledge_class"`
	RetrievalMode  string `json:"retrieval_mode"`
	Used           bool   `json:"used"`
}
