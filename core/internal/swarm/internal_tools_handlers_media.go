package swarm

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/mycelis/core/internal/cognitive"
	"github.com/mycelis/core/internal/exchange"
)

type generatedImageResponse struct {
	Data []struct {
		B64JSON       string `json:"b64_json"`
		RevisedPrompt string `json:"revised_prompt"`
	} `json:"data"`
}

func (r *InternalToolRegistry) handleGenerateImage(ctx context.Context, args map[string]any) (string, error) {
	prompt := stringValue(args["prompt"])
	if prompt == "" {
		return "", fmt.Errorf("generate_image requires 'prompt'")
	}
	size := stringValue(args["size"])
	if size == "" {
		size = "1024x1024"
	}
	if r.brain == nil || r.brain.Config == nil || r.brain.Config.Media == nil {
		return "", fmt.Errorf("media provider not configured — set media.provider or media.endpoint/model_id in cognitive.yaml")
	}
	media := r.brain.Config.Media.EffectiveProvider()
	if !media.IsEnabled() || strings.TrimSpace(media.Endpoint) == "" || strings.TrimSpace(media.ModelID) == "" {
		return "", fmt.Errorf("media provider not configured — set media.provider or media.endpoint/model_id in cognitive.yaml")
	}

	if strings.EqualFold(strings.TrimSpace(media.Type), cognitive.MediaProviderTypeForge) {
		return r.generateForgeImage(ctx, media, prompt, size)
	}
	return r.generateOpenAICompatibleImage(ctx, media, prompt, size)
}

func (r *InternalToolRegistry) generateOpenAICompatibleImage(ctx context.Context, media cognitive.MediaProviderConfig, prompt, size string) (string, error) {
	reqBody, _ := json.Marshal(map[string]any{"prompt": prompt, "n": 1, "size": size, "response_format": "b64_json", "model": media.ModelID})
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, media.Endpoint+"/images/generations", bytes.NewReader(reqBody))
	if err != nil {
		return "", fmt.Errorf("failed to create image request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")
	if authKey := strings.TrimSpace(os.Getenv(strings.TrimSpace(media.AuthKeyEnv))); authKey != "" {
		httpReq.Header.Set("Authorization", "Bearer "+authKey)
	}
	resp, err := http.DefaultClient.Do(httpReq)
	if err != nil {
		return "", fmt.Errorf("media engine unreachable: %w", err)
	}
	defer resp.Body.Close()
	respBody, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("media engine error (HTTP %d): %s", resp.StatusCode, string(respBody))
	}

	var imgResp generatedImageResponse
	if err := json.Unmarshal(respBody, &imgResp); err != nil {
		return "", fmt.Errorf("image generated but response metadata could not be parsed: %w", err)
	}
	if len(imgResp.Data) == 0 || strings.TrimSpace(imgResp.Data[0].B64JSON) == "" {
		return "", fmt.Errorf("media engine returned no image data")
	}
	return r.finishGeneratedImage(ctx, prompt, size, imgResp)
}

func (r *InternalToolRegistry) generateForgeImage(ctx context.Context, media cognitive.MediaProviderConfig, prompt, size string) (string, error) {
	width, height := imageDimensions(size)
	reqBody, _ := json.Marshal(map[string]any{
		"prompt": prompt,
		"width":  width,
		"height": height,
		"steps":  24,
	})
	endpoint := strings.TrimRight(strings.TrimSpace(media.Endpoint), "/")
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint+"/sdapi/v1/txt2img", bytes.NewReader(reqBody))
	if err != nil {
		return "", fmt.Errorf("failed to create Forge image request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")
	resp, err := http.DefaultClient.Do(httpReq)
	if err != nil {
		return "", fmt.Errorf("Forge image service is not reachable at %s: %w", endpoint, err)
	}
	defer resp.Body.Close()
	respBody, _ := io.ReadAll(resp.Body)
	if resp.StatusCode == http.StatusNotFound {
		return "", fmt.Errorf("Forge is open at %s, but image API access is off; enable API mode in the Forge/Pinokio launch settings, restart Forge, then ask Soma to try the image again", endpoint)
	}
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("Forge image generation failed (HTTP %d): %s", resp.StatusCode, string(respBody))
	}
	var forgeResponse struct {
		Images []string `json:"images"`
	}
	if err := json.Unmarshal(respBody, &forgeResponse); err != nil {
		return "", fmt.Errorf("Forge generated an image but its response could not be parsed: %w", err)
	}
	if len(forgeResponse.Images) == 0 || strings.TrimSpace(forgeResponse.Images[0]) == "" {
		return "", fmt.Errorf("Forge returned no image data")
	}
	encoded := forgeResponse.Images[0]
	if _, after, ok := strings.Cut(encoded, ","); ok && strings.HasPrefix(encoded, "data:image/") {
		encoded = after
	}
	imgResp := generatedImageResponse{}
	imgResp.Data = append(imgResp.Data, struct {
		B64JSON       string `json:"b64_json"`
		RevisedPrompt string `json:"revised_prompt"`
	}{B64JSON: encoded})
	return r.finishGeneratedImage(ctx, prompt, size, imgResp)
}

func imageDimensions(size string) (int, int) {
	parts := strings.SplitN(strings.ToLower(strings.TrimSpace(size)), "x", 2)
	if len(parts) != 2 {
		return 1024, 1024
	}
	width, widthErr := strconv.Atoi(parts[0])
	height, heightErr := strconv.Atoi(parts[1])
	if widthErr != nil || heightErr != nil || width < 64 || height < 64 || width > 2048 || height > 2048 {
		return 1024, 1024
	}
	return width, height
}

func (r *InternalToolRegistry) finishGeneratedImage(ctx context.Context, prompt, size string, imgResp generatedImageResponse) (string, error) {
	b64Content := ""
	if len(imgResp.Data) > 0 {
		b64Content = imgResp.Data[0].B64JSON
	}
	titleTrunc := prompt
	if len(titleTrunc) > 80 {
		titleTrunc = titleTrunc[:80]
	}
	title := fmt.Sprintf("Generated: %s", titleTrunc)
	expiresAt := time.Now().UTC().Add(60 * time.Minute).Format(time.RFC3339)
	meta := map[string]any{"cache_policy": "ephemeral", "saved": false, "ttl_minutes": 60, "expires_at": expiresAt, "prompt": prompt, "size": size}
	if len(imgResp.Data) > 0 && imgResp.Data[0].RevisedPrompt != "" {
		meta["revised_prompt"] = imgResp.Data[0].RevisedPrompt
	}
	metaJSON, _ := json.Marshal(meta)

	var artifactID string
	if r.db != nil {
		err := r.db.QueryRowContext(ctx, `INSERT INTO artifacts (agent_id, artifact_type, title, content_type, content, metadata, status) VALUES ('internal', 'image', $1, 'image/png', $2, $3, 'completed') RETURNING id`, title, b64Content, metaJSON).Scan(&artifactID)
		if err != nil {
			log.Printf("generate_image: failed to store artifact: %v", err)
		}
	}
	if r.exchange != nil && strings.TrimSpace(artifactID) != "" {
		if parsedID, err := uuid.Parse(artifactID); err == nil {
			runID, runClass, noRunReason := artifactExchangeRuntimeMetadata(ctx)
			_, _ = r.exchange.PublishArtifact(ctx, exchange.ArtifactNormalizationInput{
				ArtifactID:     parsedID,
				ArtifactType:   "image",
				Title:          title,
				AgentID:        "internal",
				RunID:          runID,
				RunClass:       runClass,
				NoRunReason:    noRunReason,
				RetentionClass: "retained",
				Status:         "completed",
				TargetRole:     "soma",
				Tags:           []string{"artifact", "image", "generated"},
			})
		}
	}
	return mustJSON(map[string]any{"message": fmt.Sprintf("Image generated for: \"%s\" (size: %s). Cached for 60 minutes unless saved.", prompt, size), "artifact": map[string]any{"id": artifactID, "type": "image", "title": title, "content_type": "image/png", "content": b64Content, "cached": true, "expires_at": expiresAt}}), nil
}

func (r *InternalToolRegistry) handleSaveCachedImage(ctx context.Context, args map[string]any) (string, error) {
	if r.db == nil {
		return "", fmt.Errorf("database not available — cannot save cached image")
	}
	var id, title, contentType, contentB64 string
	artifactID := stringValue(args["artifact_id"])
	if artifactID != "" {
		err := r.db.QueryRowContext(ctx, `SELECT id::text, title, content_type, content FROM artifacts WHERE id = $1::uuid AND artifact_type = 'image'`, artifactID).Scan(&id, &title, &contentType, &contentB64)
		if err != nil {
			return "", fmt.Errorf("cached image %q not found", artifactID)
		}
	} else {
		err := r.db.QueryRowContext(ctx, `SELECT id::text, title, content_type, content FROM artifacts WHERE artifact_type = 'image' AND COALESCE(metadata->>'cache_policy', '') = 'ephemeral' AND COALESCE(metadata->>'saved', 'false') <> 'true' ORDER BY created_at DESC LIMIT 1`).Scan(&id, &title, &contentType, &contentB64)
		if err != nil {
			return "", fmt.Errorf("no unsaved cached image found")
		}
	}
	data, err := base64.StdEncoding.DecodeString(contentB64)
	if err != nil {
		return "", fmt.Errorf("decode cached image: %w", err)
	}
	rel, filename, err := writeSavedImageFile(title, id, contentType, stringValue(args["folder"]), stringValue(args["filename"]), data)
	if err != nil {
		return "", err
	}
	if _, err := r.db.ExecContext(ctx, `UPDATE artifacts SET file_path = $1::text, file_size_bytes = $2, metadata = COALESCE(metadata, '{}'::jsonb) || jsonb_build_object('saved', true, 'saved_path', $1::text, 'saved_at', NOW()) WHERE id = $3::uuid`, rel, int64(len(data)), id); err != nil {
		return "", fmt.Errorf("update image save metadata: %w", err)
	}
	return mustJSON(map[string]any{"message": fmt.Sprintf("Saved cached image to %s", rel), "artifact": map[string]any{"id": id, "type": "file", "title": filename, "saved_path": rel}}), nil
}

func writeSavedImageFile(title, id, contentType, folder, filename string, data []byte) (string, string, error) {
	if strings.TrimSpace(folder) == "" {
		folder = "saved-media"
	}
	if strings.TrimSpace(filename) == "" {
		filename = sanitizeImageFilename(title)
		if filename == "" {
			filename = fmt.Sprintf("image-%s", id[:8])
		}
	}
	if filepath.Ext(filename) == "" {
		filename += imageFileExt(contentType)
	}
	targetPath, err := validateToolPath(filepath.ToSlash(filepath.Join(folder, filename)))
	if err != nil {
		return "", "", err
	}
	if err := os.MkdirAll(filepath.Dir(targetPath), 0o755); err != nil {
		return "", "", fmt.Errorf("create target directory: %w", err)
	}
	if err := os.WriteFile(targetPath, data, 0o644); err != nil {
		return "", "", fmt.Errorf("write target file: %w", err)
	}
	workspace := os.Getenv("MYCELIS_WORKSPACE")
	if workspace == "" {
		workspace = "./workspace"
	}
	absWorkspace, _ := filepath.Abs(workspace)
	rel, relErr := filepath.Rel(absWorkspace, targetPath)
	if relErr != nil {
		rel = targetPath
	}
	return filepath.ToSlash(rel), filename, nil
}

func sanitizeImageFilename(s string) string {
	s = strings.TrimSpace(strings.ToLower(s))
	if s == "" {
		return ""
	}
	var b strings.Builder
	lastDash := false
	for _, r := range s {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') {
			b.WriteRune(r)
			lastDash = false
			continue
		}
		if !lastDash {
			b.WriteRune('-')
			lastDash = true
		}
	}
	out := strings.Trim(b.String(), "-.")
	if len(out) > 80 {
		out = out[:80]
	}
	return out
}

func imageFileExt(contentType string) string {
	switch strings.ToLower(strings.TrimSpace(contentType)) {
	case "image/jpeg", "image/jpg":
		return ".jpg"
	case "image/webp":
		return ".webp"
	case "image/gif":
		return ".gif"
	default:
		return ".png"
	}
}
