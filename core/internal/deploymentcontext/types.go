package deploymentcontext

import "time"

const (
	defaultChunkSize               = 1200
	defaultChunkOverlap            = 160
	KnowledgeClassCustomerContext  = "customer_context"
	KnowledgeClassCompanyKnowledge = "company_knowledge"
	KnowledgeClassSomaOperating    = "soma_operating_context"
	KnowledgeClassUserPrivate      = "user_private_context"
	KnowledgeClassReflection       = "reflection_synthesis"
)

type IngestRequest struct {
	KnowledgeClass    string
	Title             string
	Content           string
	ContentType       string
	SourceLabel       string
	SourceKind        string
	Visibility        string
	SensitivityClass  string
	TrustClass        string
	Tags              []string
	AgentID           string
	TeamID            string
	UserLabel         string
	SomaContextKind   string
	OutputSpecificity string
	ContentDomain     string
	TargetGoalSets    []string
	ExtraMetadata     map[string]any
}

type PromoteRequest struct {
	SourceArtifactID string
	Title            string
	Content          string
	ContentType      string
	SourceLabel      string
	SourceKind       string
	Visibility       string
	SensitivityClass string
	TrustClass       string
	Tags             []string
	AgentID          string
	TeamID           string
	UserLabel        string
}

type IngestResult struct {
	ArtifactID       string    `json:"artifact_id"`
	KnowledgeClass   string    `json:"knowledge_class"`
	Title            string    `json:"title"`
	SourceLabel      string    `json:"source_label"`
	SourceKind       string    `json:"source_kind"`
	Visibility       string    `json:"visibility"`
	SensitivityClass string    `json:"sensitivity_class"`
	TrustClass       string    `json:"trust_class"`
	ChunkCount       int       `json:"chunk_count"`
	VectorCount      int       `json:"vector_count"`
	ContentPreview   string    `json:"content_preview"`
	ContentLength    int       `json:"content_length"`
	ContentDomain    string    `json:"content_domain,omitempty"`
	TargetGoalSets   []string  `json:"target_goal_sets,omitempty"`
	CreatedAt        time.Time `json:"created_at"`
}

type Entry struct {
	ArtifactID       string    `json:"artifact_id"`
	KnowledgeClass   string    `json:"knowledge_class"`
	Title            string    `json:"title"`
	SourceLabel      string    `json:"source_label"`
	SourceKind       string    `json:"source_kind"`
	Visibility       string    `json:"visibility"`
	SensitivityClass string    `json:"sensitivity_class"`
	TrustClass       string    `json:"trust_class"`
	ChunkCount       int       `json:"chunk_count"`
	VectorCount      int       `json:"vector_count"`
	ContentPreview   string    `json:"content_preview"`
	ContentLength    int       `json:"content_length"`
	ContentDomain    string    `json:"content_domain,omitempty"`
	TargetGoalSets   []string  `json:"target_goal_sets,omitempty"`
	CreatedAt        time.Time `json:"created_at"`
}
