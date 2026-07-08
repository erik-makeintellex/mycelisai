package protocol

// ChatArtifactRef is an inline artifact reference embedded in a chat response.
// It is the operator-facing output contract for rich results produced by Soma
// or consulted specialists. Small content is inline; large/binary content uses
// ID or URL references that can be fetched separately.
type ChatArtifactRef struct {
	ID          string      `json:"id,omitempty"` // artifact table ID (for stored artifacts)
	Type        string      `json:"type"`         // code | document | image | audio | data | chart | file
	OutputClass OutputClass `json:"output_class,omitempty"`
	Title       string      `json:"title"`
	ContentType string      `json:"content_type,omitempty"` // MIME type
	Content     string      `json:"content,omitempty"`      // inline content (text, JSON, base64 for images)
	URL         string      `json:"url,omitempty"`          // external URL (for links, images)
	// Optional image-cache lifecycle metadata.
	Cached     bool     `json:"cached,omitempty"`
	ExpiresAt  string   `json:"expires_at,omitempty"`
	SavedPath  string   `json:"saved_path,omitempty"`
	Entrypoint string   `json:"entrypoint,omitempty"`
	Folder     string   `json:"folder,omitempty"`
	Files      []string `json:"files,omitempty"`
	Validation string   `json:"validation,omitempty"`
}
